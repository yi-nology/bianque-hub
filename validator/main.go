// bianque-hub 结构契约校验器（独立于扁鹊平台，CI 与本地同款）。
//
// 校验分五层：pack.yaml 契约门、专家/技能/链结构规则、跨包唯一性与工具面覆盖、
// 版本同步（CHANGELOG 头部条目与 README 包清单表）、脱敏扫描。
// 规则与扁鹊 internal/agents 装载校验（packs.go）同源，但独立实现——平台私仓
// 不外泄，规则漂移以本文件注释为准绳同步。
//
// 用法（约定在仓库根运行，脱敏扫描自动覆盖当前目录，无需显式传入）：
//
//	go run ./validator              # 等价于 ./packs
//	go run ./validator ./packs      # 容器目录：校验其下每个子目录中的包
//	go run ./validator packs/foo    # 单包目录：参数自身含 pack.yaml 即按单包校验
//
// 退出码：有 error 非零；warning 不影响退出码（打印供人复核）。
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// ---- 契约常量（与平台侧同步；漂移=装载失败，改这里须同步扁que PACK_GUIDE）----

const wantAPIVersion = 1 // 扁鹊 PackAPIVersion

// routeGroups 平台 _shared/agents.yaml route_groups 白名单（未知组=平台装载失败）。
var routeGroups = map[string]bool{
	"system": true, "storage": true, "network": true, "middleware": true,
	"infra": true, "security": true, "offline": true, "tooling": true,
	"knowledge": true, "status_query": true, "monitoring": true, "kubernetes": true,
}

var (
	kinds             = map[string]bool{"engine": true, "llm": true, "stub": true}
	maturities        = map[string]bool{"experimental": true, "stable": true, "frozen": true, "deprecated": true}
	skillModes        = map[string]bool{"static": true, "on_demand": true}
	prioRe            = regexp.MustCompile(`^P[0-6]$`)
	semverRe          = regexp.MustCompile(`^\d+\.\d+\.\d+(-[0-9A-Za-z.-]+)?$`)
	changelogHeadRe   = regexp.MustCompile(`(?m)^##\s+(\d+\.\d+\.\d+(?:-[0-9A-Za-z.-]+)?)`)
	readmeRowRe       = regexp.MustCompile(`(?m)^\|\s*\[[^\]]+\]\(packs/([^/)]+)/\)\s*\|\s*([^\s|]+)`)
	toolServerAllowRe = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)
	// 脱敏（error 级）：内网域名/内部工具名/密钥形态/私网 IP（127.0.0.1 与 0.0.0.0 豁免）。
	sensitivePatterns = []struct {
		re  *regexp.Regexp
		why string
	}{
		{regexp.MustCompile(`git\.enjoye\.top`), "内网模块域名"},
		{regexp.MustCompile(`kyaiops`), "内部工具名（kyaiops）"},
		{regexp.MustCompile(`-----BEGIN [A-Z ]*PRIVATE KEY`), "私钥形态"},
		{regexp.MustCompile(`AKIA[0-9A-Z]{16}`), "AWS AccessKey 形态"},
		{regexp.MustCompile(`\b(10\.\d+\.\d+\.\d+|172\.(1[6-9]|2\d|3[01])\.\d+\.\d+|192\.168\.\d+\.\d+)\b`), "私网 IP"},
	}
	// scanExts 脱敏扫描纳入的文件类型：文档/配置文本 + 无扩展名文件
	//（凭证可能藏在任意文本类型里；二进制按字节匹配无害）。
	scanExts = map[string]bool{
		"": true, ".md": true, ".yaml": true, ".yml": true, ".json": true,
		".txt": true, ".sh": true, ".cfg": true, ".conf": true, ".ini": true, ".toml": true,
	}
)

// hostSideMCP 平台宿主侧工具面白名单：包内无契约文件、实例装载即存在的 server。
// 专家级子集安装（spec 2026-10-05-per-expert-install）据此裁决 tools server
// 可携带性——白名单外且无包内 mcp/ 契约 = 该专家装不出去。新增宿主面 server
// 与平台侧（conf 挂载面/experts/_shared）同步。
var hostSideMCP = map[string]bool{
	"ask-ops":            true, // 宿主侧受审只读命令面
	"security-assistant": true, // 安全巡检采集面（内置）
	"datasources":        true, // 平台数据源查询面（n8n/loki 等）
	"platform":           true, // 平台内置工具保留授予名（非 MCP server）：search_knowledge/memory_search/http_query 等（bianque-hub-inner om-gitlab-ci 首个包内授予方）
}

// ---- 结构 ----

type packManifest struct {
	APIVersion  int    `yaml:"api_version"`
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Description string `yaml:"description"`
	Changelog   string `yaml:"changelog"`
	Provides    struct {
		Experts []string `yaml:"experts"`
		Skills  []string `yaml:"skills"`
	} `yaml:"provides"`
}

// mcpServer 工具面声明（agent.yaml 的 tools 与 SKILL.md 的 requires_mcp 同构）。
type mcpServer struct {
	Server string   `yaml:"server"`
	Tools  []string `yaml:"tools"`
	Allow  []string `yaml:"allow"`
}

type agentDef struct {
	Slug           string      `yaml:"slug"`
	Name           string      `yaml:"name"`
	Kind           string      `yaml:"kind"`
	PromptFile     string      `yaml:"prompt_file"`
	PromptIncludes []string    `yaml:"prompt_includes"`
	Skills         []any       `yaml:"skills"`
	RoutePriority  string      `yaml:"route_priority"`
	RouteGroup     string      `yaml:"route_group"`
	RouteDesc      string      `yaml:"route_desc"`
	RouteKeywords  []string    `yaml:"route_keywords"`
	Symptoms       []string    `yaml:"symptoms"`
	Tools          []mcpServer `yaml:"tools"`
}

type chainStep struct {
	Type        string `yaml:"type"`
	Agent       string `yaml:"agent"`
	Instruction string `yaml:"instruction"`
	Skill       any    `yaml:"skill"`
}

type chainDef struct {
	Slug          string      `yaml:"slug"`
	Name          string      `yaml:"name"`
	RoutePriority string      `yaml:"route_priority"`
	RouteGroup    string      `yaml:"route_group"`
	RouteDesc     string      `yaml:"route_desc"`
	RouteKeywords []string    `yaml:"route_keywords"`
	Symptoms      []string    `yaml:"symptoms"`
	Steps         []chainStep `yaml:"steps"`
}

type skillFront struct {
	Name        string      `yaml:"name"`
	Description string      `yaml:"description"`
	Mode        string      `yaml:"mode"`
	Version     string      `yaml:"version"`
	Maturity    string      `yaml:"maturity"`
	RequiresMCP []mcpServer `yaml:"requires_mcp"`
	// ProvidesChanges 编目变更块引用（批次九十四「受审执行」）：本技能方法论
	// 覆盖的处方编目 slug 清单，须与包内（或跨包）changes/*.yaml 对上——
	// 开放键，其他运行时可忽略（requires_mcp 同例）。
	ProvidesChanges []string `yaml:"provides_changes"`
}

// changeDef 编目变更块（<包>/changes/<slug>.yaml；平台 agents.ChangeDef 同构镜像，
// 校验子集）。治理红线：包侧只声明编目，审批/执行/验证归平台。
type changeDef struct {
	APIVersion  int      `yaml:"api_version"`
	Slug        string   `yaml:"slug"`
	Title       string   `yaml:"title"`
	RequestType string   `yaml:"request_type"`
	Risk        int      `yaml:"risk"`
	Rollback    string   `yaml:"rollback"`
	Commands    []string `yaml:"commands"`
}

type finding struct {
	level string // ERROR | WARN
	pack  string
	msg   string
}

// kwFirst 路由词首登记录（平台 packs.go registerRouteKeyword 同构镜像）。
type kwFirst struct{ prio, pack string }

// registerKeyword 镜像平台跨包冲突规则：同包 OK；跨包同优先级 = 真歧义（ERROR，
// 平台装载会失败——PR 必须在 CI 拦下）；跨层 = 警告（高优先级胜出）。
// 局限：hub 只见本仓包，与平台内置域词（_shared/os-basics 内置版）的冲突在此
// 查不到，靠 CONTRIBUTING 路由词纪律人工把关。
func registerKeyword(kwOwner map[string]kwFirst, kw, pack, prio string) (level, msg string, dup bool) {
	first, hit := kwOwner[kw]
	if !hit {
		kwOwner[kw] = kwFirst{prio: prio, pack: pack}
		return "", "", false
	}
	if first.pack == pack {
		return "", "", false
	}
	if first.prio == prio {
		return "ERROR", fmt.Sprintf("关键词/症状 %q 在 %s 与 %s 同优先级 %s 重复（真歧义，平台装载会失败）", kw, first.pack, pack, prio), true
	}
	return "WARN", fmt.Sprintf("关键词/症状 %q 跨层重复（%s@%s 与 %s@%s），高优先级胜出", kw, first.pack, first.prio, pack, prio), true
}

// validator 全仓校验状态：跨包唯一性登记 + 全包扫描后才能判的延迟核对项。
type validator struct {
	findings     []finding
	skillOwner   map[string]string   // 技能名 → 包（跨包唯一性）
	skillNeeds   map[string][]string // 技能名 → requires_mcp server 清单（工具面覆盖核对）
	changeOwner  map[string]string   // 变更块 slug → 包（跨包唯一性，平台装载同规则）
	skillChanges map[string][]string // 技能名 → provides_changes 引用（延迟核对存在性）
	expertOwner  map[string]string   // 专家 slug → 包
	expertTools  map[string][]string // 专家 slug → tools 授权的 server 清单
	chainOwner   map[string]string   // 链 slug → 包
	kwOwner      map[string]kwFirst  // 路由词/症状 → 首登（prio, pack）
	mountChecks  []mountCheck        // 专家挂载/链钉扎（跨包解析后核对工具面覆盖）
	packVersions map[string]string   // 包名 → version（README 包清单核对）
	// 专家独立安装性核对输入（spec 2026-10-05-per-expert-install §6）
	packMCP    map[string]map[string]bool // 包名 → 包内 mcp/ 契约 server 集
	expertRefs map[string]expertRef       // 专家 slug → 闭包引用快照
}

// expertRef 专家闭包引用快照（checkExpertInstall 的核对输入）。
type expertRef struct {
	pack   string
	skills []string
	tools  []string
}

// mountCheck 记一处「专家↔技能」绑定：chainSlug 空 = 专家挂载，非空 = 链步骤钉扎。
type mountCheck struct {
	pack, expert, skill, chainSlug string
}

func (v *validator) add(level, pack, format string, a ...any) {
	v.findings = append(v.findings, finding{level, pack, fmt.Sprintf(format, a...)})
}

func main() {
	dirs := os.Args[1:]
	if len(dirs) == 0 {
		dirs = []string{"./packs"}
	}
	findings, errN, warnN := run(dirs)
	for _, f := range findings {
		fmt.Printf("%-5s [%s] %s\n", f.level, f.pack, f.msg)
	}
	fmt.Printf("\n校验完成：%d error / %d warning\n", errN, warnN)
	if errN > 0 {
		os.Exit(1)
	}
}

// run 执行全部校验层（独立于 main 以便测试）：包发现 → 包级校验 → 跨包核对 →
// 版本同步核对 → 脱敏扫描，返回 findings 与 error/warning 计数。
func run(dirs []string) ([]finding, int, int) {
	v := &validator{
		skillOwner:   map[string]string{},
		skillNeeds:   map[string][]string{},
		changeOwner:  map[string]string{},
		skillChanges: map[string][]string{},
		expertOwner:  map[string]string{},
		expertTools:  map[string][]string{},
		chainOwner:   map[string]string{},
		kwOwner:      map[string]kwFirst{},
		packVersions: map[string]string{},
		packMCP:      map[string]map[string]bool{},
		expertRefs:   map[string]expertRef{},
	}
	for _, root := range dirs {
		// 参数自身含 pack.yaml：单包模式（模板包自检命令即此形态——曾因只认
		// 容器目录而静默校验 0 项假绿，见 chain-starter README）。
		if _, err := os.Stat(filepath.Join(root, "pack.yaml")); err == nil {
			v.checkPack(root, filepath.Base(root))
			continue
		}
		entries, err := os.ReadDir(root)
		if err != nil {
			v.add("ERROR", filepath.Base(root), "读取目录 %s: %v", root, err)
			continue
		}
		found := false
		for _, e := range entries {
			if !e.IsDir() || strings.HasPrefix(e.Name(), ".") || strings.HasPrefix(e.Name(), "_") {
				continue
			}
			packDir := filepath.Join(root, e.Name())
			if _, err := os.Stat(filepath.Join(packDir, "pack.yaml")); err != nil {
				continue // 非包目录（无 pack.yaml）不校验
			}
			v.checkPack(packDir, e.Name())
			found = true
		}
		if !found {
			v.add("ERROR", filepath.Base(root), "目录 %s 下未发现任何包（校验单个包请指向含 pack.yaml 的目录）", root)
		} else {
			v.checkIndex(root)
		}
	}
	v.checkCrossRefs()
	v.checkExpertInstall()
	v.checkChangeRefs()
	v.checkReadme()
	// 脱敏扫描：显式参数 + 当前目录（默认仓库根——根 README/CONTRIBUTING 也必须
	// 在覆盖内），文件级去重防双报。
	roots := slices.Clone(dirs)
	roots = append(roots, ".")
	v.findings = append(v.findings, scanSensitive(roots...)...)
	v.findings = append(v.findings, scanDangerous(roots...)...)

	errN, warnN := 0, 0
	for _, f := range v.findings {
		if f.level == "ERROR" {
			errN++
		} else {
			warnN++
		}
	}
	return v.findings, errN, warnN
}

func (v *validator) checkPack(packDir, packName string) {
	add := func(level, format string, a ...any) {
		v.add(level, packName, format, a...)
	}

	// pack.yaml 契约门
	raw, err := os.ReadFile(filepath.Join(packDir, "pack.yaml"))
	if err != nil {
		add("ERROR", "读 pack.yaml: %v", err)
		return
	}
	var m packManifest
	if err := yaml.Unmarshal(raw, &m); err != nil {
		add("ERROR", "pack.yaml 解析: %v", err)
		return
	}
	if m.APIVersion != wantAPIVersion {
		add("ERROR", "pack.yaml api_version=%d 与契约 %d 不兼容", m.APIVersion, wantAPIVersion)
	}
	if m.Name == "" {
		add("ERROR", "pack.yaml 缺 name")
	} else if m.Name != packName {
		add("ERROR", "pack.yaml name=%q 与目录名 %q 不一致（安装目录以 name 为准）", m.Name, packName)
	}
	if !semverRe.MatchString(m.Version) {
		add("ERROR", "pack.yaml version=%q 非 SemVer", m.Version)
	}
	if strings.TrimSpace(m.Description) == "" {
		add("ERROR", "pack.yaml 缺 description")
	}
	if _, err := os.Stat(filepath.Join(packDir, "provenance.json")); err != nil {
		add("ERROR", "缺 provenance.json（市场包来源标记，装进扁鹊靠它识别为市场版）")
	} else if pj, err := os.ReadFile(filepath.Join(packDir, "provenance.json")); err == nil {
		var prov struct {
			BaseURL string `json:"base_url"`
		}
		if json.Unmarshal(pj, &prov) != nil || prov.BaseURL == "" {
			add("ERROR", "provenance.json 缺 base_url")
		}
	}
	// 版本同步：CHANGELOG 头部条目必须等于 pack.yaml version（版本纪律的机械门）。
	cl := m.Changelog
	if cl == "" {
		cl = "CHANGELOG.md"
	}
	if craw, err := os.ReadFile(filepath.Join(packDir, cl)); err != nil {
		add("ERROR", "缺 %s（任何内容变更须升版本并留痕）", cl)
	} else if head := changelogHeadRe.FindStringSubmatch(string(craw)); head == nil {
		add("ERROR", "%s 缺版本条目（## <version>）", cl)
	} else if head[1] != m.Version {
		add("ERROR", "%s 头部版本 %s 与 pack.yaml version %s 不一致", cl, head[1], m.Version)
	}
	v.packVersions[packName] = m.Version

	// 专家目录
	expertSlugs := map[string]string{} // slug → kind
	skillRefs := map[string][]string{} // slug → 挂载技能名（存在性核对用）
	expertDirs, _ := os.ReadDir(packDir)
	for _, e := range expertDirs {
		if !e.IsDir() || e.Name() == "skills" || e.Name() == "prompts" || e.Name() == "mcp" || e.Name() == "credentials" {
			continue
		}
		agentPath := filepath.Join(packDir, e.Name(), "agent.yaml")
		araw, err := os.ReadFile(agentPath)
		if err != nil {
			continue // 无 agent.yaml 的目录（如 _trash 类）跳过
		}
		var d agentDef
		if err := yaml.Unmarshal(araw, &d); err != nil {
			add("ERROR", "%s/agent.yaml 解析: %v", e.Name(), err)
			continue
		}
		if d.Slug == "" {
			add("ERROR", "%s/agent.yaml 缺 slug", e.Name())
			continue
		}
		if prev, dup := v.expertOwner[d.Slug]; dup {
			add("ERROR", "专家 slug %q 跨包重复（%s 已占用）", d.Slug, prev)
		} else {
			v.expertOwner[d.Slug] = packName
		}
		expertSlugs[d.Slug] = d.Kind
		if !kinds[d.Kind] {
			add("ERROR", "%s kind=%q 非法（engine|llm|stub）", d.Slug, d.Kind)
		}
		if !prioRe.MatchString(d.RoutePriority) {
			add("ERROR", "%s route_priority=%q 非法（P0-P6）", d.Slug, d.RoutePriority)
		}
		if d.RouteGroup != "" && !routeGroups[d.RouteGroup] {
			add("ERROR", "%s route_group=%q 不在平台 route_groups 白名单", d.Slug, d.RouteGroup)
		}
		// 包侧新契约（imports/ 前缀专家）：P0 红线禁用、P2 起步、route_desc 必备。
		// 等位接管包（specialists/、workflow/ 等平台前缀）镜像内置优先级与形态，不受此限。
		if strings.HasPrefix(d.Slug, "imports/") {
			if d.RoutePriority == "P0" {
				add("ERROR", "%s 包侧专家禁用 P0（P0 红线是平台专家域）", d.Slug)
			} else if d.RoutePriority == "P1" {
				add("WARN", "%s 包侧专家应从 P2 起步（P1 需与内置入口错位的充分理由）", d.Slug)
			}
			if d.RouteDesc == "" {
				add("WARN", "%s 缺 route_desc（LLM 语义面，防路由错位的第一防线）", d.Slug)
			}
		}
		if len(d.RouteKeywords) == 0 {
			add("WARN", "%s 无 route_keywords（路由不可达，孤立专家）", d.Slug)
		}
		for _, kw := range append(slices.Clone(d.RouteKeywords), d.Symptoms...) {
			if level, msg, dup := registerKeyword(v.kwOwner, kw, packName, d.RoutePriority); dup {
				add(level, "%s: %s", d.Slug, msg)
			}
		}
		if d.Kind == "llm" {
			if d.PromptFile == "" {
				add("ERROR", "%s kind=llm 必填 prompt_file", d.Slug)
			} else if _, err := os.Stat(filepath.Join(packDir, e.Name(), d.PromptFile)); err != nil {
				add("ERROR", "%s prompt_file %q 不存在", d.Slug, d.PromptFile)
			}
		}
		for _, inc := range d.PromptIncludes {
			if strings.HasPrefix(inc, "_shared/") {
				continue // 平台共享层（根相对路径，任意扁鹊实例必有）；非扁鹊读者见仓内 knowledge/ 副本
			}
			if _, err := os.Stat(filepath.Join(packDir, e.Name(), inc)); err != nil {
				add("ERROR", "%s prompt_includes %q 不存在（专家目录内相对路径）", d.Slug, inc)
			}
		}
		var toolServers []string
		for _, t := range d.Tools {
			if t.Server == "" {
				add("ERROR", "%s tools 条目缺 server", d.Slug)
				continue
			}
			if !toolServerAllowRe.MatchString(t.Server) {
				add("ERROR", "%s tools server=%q 命名非法（小写中划线）", d.Slug, t.Server)
			}
			toolServers = append(toolServers, t.Server)
		}
		v.expertTools[d.Slug] = toolServers
		for _, s := range d.Skills {
			name := skillRefName(s)
			if name == "" {
				continue
			}
			skillRefs[d.Slug] = append(skillRefs[d.Slug], name)
			v.mountChecks = append(v.mountChecks, mountCheck{pack: packName, expert: d.Slug, skill: name})
		}
		v.expertRefs[d.Slug] = expertRef{pack: packName, skills: skillRefs[d.Slug], tools: toolServers}
	}

	// 技能
	skillDirs, _ := os.ReadDir(filepath.Join(packDir, "skills"))
	present := map[string]bool{}
	if skillDirs != nil {
		for _, e := range skillDirs {
			if !e.IsDir() {
				continue
			}
			name := e.Name()
			present[name] = true
			p := filepath.Join(packDir, "skills", name, "SKILL.md")
			front, body, err := readFrontmatter(p)
			if err != nil {
				add("ERROR", "skills/%s/SKILL.md: %v", name, err)
				continue
			}
			var sf skillFront
			if err := yaml.Unmarshal([]byte(front), &sf); err != nil {
				add("ERROR", "skills/%s frontmatter 非 YAML: %v", name, err)
				continue
			}
			if sf.Name == "" || sf.Description == "" {
				add("ERROR", "skills/%s frontmatter 缺 name/description（开放技能标准必填）", name)
			}
			if sf.Name != name {
				add("ERROR", "skills/%s frontmatter name=%q 与目录名不一致", name, sf.Name)
			}
			if sf.Mode != "" && !skillModes[sf.Mode] {
				add("ERROR", "skills/%s mode=%q 非法（static|on_demand）", name, sf.Mode)
			}
			if sf.Version == "" {
				add("ERROR", "skills/%s frontmatter 缺 version（技能内容变更须同步升版）", name)
			} else if !semverRe.MatchString(sf.Version) {
				add("ERROR", "skills/%s version=%q 非 SemVer", name, sf.Version)
			}
			if sf.Maturity != "" && !maturities[sf.Maturity] {
				add("ERROR", "skills/%s maturity=%q 非法", name, sf.Maturity)
			}
			for _, mc := range sf.RequiresMCP {
				if mc.Server == "" {
					add("ERROR", "skills/%s requires_mcp 条目缺 server", name)
					continue
				}
				if !toolServerAllowRe.MatchString(mc.Server) {
					add("ERROR", "skills/%s requires_mcp server=%q 命名非法（小写中划线）", name, mc.Server)
				}
				for _, t := range mc.Tools {
					if strings.TrimSpace(t) == "" {
						add("ERROR", "skills/%s requires_mcp server %q 的 tools 含空工具名", name, mc.Server)
					}
				}
				v.skillNeeds[sf.Name] = append(v.skillNeeds[sf.Name], mc.Server)
			}
			for _, cs := range sf.ProvidesChanges {
				if strings.TrimSpace(cs) == "" {
					add("ERROR", "skills/%s provides_changes 含空 slug", name)
					continue
				}
				v.skillChanges[sf.Name] = append(v.skillChanges[sf.Name], cs)
			}
			if strings.TrimSpace(body) == "" {
				add("ERROR", "skills/%s 正文为空", name)
			}
			if prev, dup := v.skillOwner[sf.Name]; dup {
				add("ERROR", "技能名 %q 跨包重复（%s 已占用；技能名是全局唯一空间）", sf.Name, prev)
			} else if sf.Name != "" {
				v.skillOwner[sf.Name] = packName
			}
		}
	}
	for _, s := range m.Provides.Skills {
		if !present[s] {
			add("ERROR", "provides.skills 声明 %q 但 skills/%s/SKILL.md 不存在", s, s)
		}
	}
	for name := range present {
		if !slices.Contains(m.Provides.Skills, name) {
			add("WARN", "skills/%s 存在但未登记 provides.skills（平台按 provides 装配，漏登记=装不进）", name)
		}
	}
	// 包内 MCP 契约集（专家独立安装性核对：tools server 须有包内契约或属宿主面）
	if mcpDirs, _ := os.ReadDir(filepath.Join(packDir, "mcp")); mcpDirs != nil {
		set := map[string]bool{}
		for _, e := range mcpDirs {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".yaml") {
				set[strings.TrimSuffix(e.Name(), ".yaml")] = true
			}
		}
		v.packMCP[packName] = set
	}
	for _, s := range m.Provides.Experts {
		if _, ok := expertSlugs[s]; !ok {
			add("ERROR", "provides.experts 声明 %q 但包内无此 slug", s)
		}
	}
	for slug := range expertSlugs {
		if !slices.Contains(m.Provides.Experts, slug) {
			add("WARN", "专家 %s 存在但未登记 provides.experts（平台按 provides 装配，漏登记=装不进）", slug)
		}
	}
	// 专家技能引用存在性（包内或声明 requires 跨包时降级 warn）
	for slug, names := range skillRefs {
		for _, n := range names {
			if !present[n] {
				add("WARN", "%s 引用技能 %q 不在本包（跨包引用合法，请确认目标包已安装）", slug, n)
			}
		}
	}

	// 编目变更块（批次九十四「受审执行」）：changes/*.yaml 契约门 + 跨包 slug 唯一
	// + 模板结构层快检（命令替换/控制字符拒——平台装载同规则，CI 先拦）。
	if changeDirs, _ := os.ReadDir(filepath.Join(packDir, "changes")); changeDirs != nil {
		for _, e := range changeDirs {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
				continue
			}
			craw, cerr := os.ReadFile(filepath.Join(packDir, "changes", e.Name()))
			if cerr != nil {
				add("ERROR", "changes/%s: %v", e.Name(), cerr)
				continue
			}
			var cd changeDef
			if yerr := yaml.Unmarshal(craw, &cd); yerr != nil {
				add("ERROR", "changes/%s 解析: %v", e.Name(), yerr)
				continue
			}
			if cd.APIVersion != wantAPIVersion {
				add("ERROR", "changes/%s api_version=%d 与契约 %d 不兼容", e.Name(), cd.APIVersion, wantAPIVersion)
			}
			if cd.Slug == "" || cd.Title == "" || cd.RequestType == "" || len(cd.Commands) == 0 {
				add("ERROR", "changes/%s 缺 slug/title/request_type/commands（变更块最小契约）", e.Name())
				continue
			}
			if cd.Risk < 1 || cd.Risk > 4 {
				add("ERROR", "changes/%s risk=%d 越界（1-4）", e.Name(), cd.Risk)
			}
			if strings.TrimSpace(cd.Rollback) == "" {
				add("WARN", "changes/%s 缺 rollback（审批卡回退说明，人读必需）", e.Name())
			}
			for _, tmpl := range cd.Commands {
				if msg := changeTainted(tmpl); msg != "" {
					add("ERROR", "changes/%s 命令模板非法: %s", e.Name(), msg)
				}
			}
			if prev, dup := v.changeOwner[cd.Slug]; dup {
				add("ERROR", "变更块 slug %q 跨包重复（%s 与 %s；平台装载同规则 fail）", cd.Slug, prev, packName)
			} else {
				v.changeOwner[cd.Slug] = packName
			}
		}
	}

	// 平台耦合文件（等位接管包携带）：语法门 + 映射目标核对
	if araw, err := os.ReadFile(filepath.Join(packDir, "analyzers.yaml")); err == nil {
		var an struct {
			Analyzers []struct {
				Slug string `yaml:"slug"`
			} `yaml:"analyzers"`
		}
		if err := yaml.Unmarshal(araw, &an); err != nil {
			add("ERROR", "analyzers.yaml 解析: %v", err)
		} else {
			for _, a := range an.Analyzers {
				if _, ok := expertSlugs[a.Slug]; !ok {
					add("ERROR", "analyzers.yaml 映射未知专家 %q（装载期会整体回落成检查盲区）", a.Slug)
				}
			}
		}
	}
	if draw, err := os.ReadFile(filepath.Join(packDir, "disambiguation.yaml")); err == nil {
		var x any // 消歧目标可指向平台内置 slug，只做语法门不做存在性核对
		if yaml.Unmarshal(draw, &x) != nil {
			add("ERROR", "disambiguation.yaml 解析失败")
		}
	}

	// chain.yaml（可选）
	craw, err := os.ReadFile(filepath.Join(packDir, "chain.yaml"))
	if err == nil {
		var c chainDef
		if err := yaml.Unmarshal(craw, &c); err != nil {
			add("ERROR", "chain.yaml 解析: %v", err)
		} else {
			if !strings.HasPrefix(c.Slug, "workflow/") {
				add("ERROR", "chain slug %q 必须 workflow/ 前缀", c.Slug)
			}
			if prev, dup := v.chainOwner[c.Slug]; dup {
				add("ERROR", "chain slug %q 跨包重复（%s 已占用）", c.Slug, prev)
			} else if c.Slug != "" {
				v.chainOwner[c.Slug] = packName
			}
			if len(c.Steps) == 0 {
				add("ERROR", "chain %s 无步骤", c.Slug)
			}
			if c.RoutePriority == "" {
				add("ERROR", "chain %s 缺 route_priority", c.Slug)
			} else if c.RoutePriority == "P0" {
				add("ERROR", "chain %s 禁用 P0（P0 红线是专家域）", c.Slug)
			} else if !prioRe.MatchString(c.RoutePriority) {
				add("ERROR", "chain %s route_priority=%q 非法", c.Slug, c.RoutePriority)
			}
			if c.RouteGroup != "" && !routeGroups[c.RouteGroup] {
				add("ERROR", "chain %s route_group=%q 不在白名单", c.Slug, c.RouteGroup)
			}
			if c.RouteDesc == "" {
				add("WARN", "chain %s 缺 route_desc（链入口的 LLM 语义面）", c.Slug)
			}
			for _, kw := range append(slices.Clone(c.RouteKeywords), c.Symptoms...) {
				if level, msg, dup := registerKeyword(v.kwOwner, kw, packName, c.RoutePriority); dup {
					add(level, "chain %s: %s", c.Slug, msg)
				}
			}
			for i, st := range c.Steps {
				kind := st.Type
				if kind == "" {
					kind = "agent"
				}
				if kind == "agent" {
					if strings.HasPrefix(st.Agent, "workflow/") {
						add("ERROR", "chain %s 步骤%d 引用 %q 非法：链步骤只能是叶子专家（治理闭环由平台衔接）", c.Slug, i+1, st.Agent)
					}
					if k, ok := expertSlugs[st.Agent]; ok && k == "engine" {
						add("ERROR", "chain %s 步骤%d 引用 engine 专家 %q（链步骤禁 engine）", c.Slug, i+1, st.Agent)
					} else if !ok {
						add("WARN", "chain %s 步骤%d 引用 %q 不在本包（跨包引用合法，请确认目标包已安装）", c.Slug, i+1, st.Agent)
					}
					if pin := skillRefName(st.Skill); pin != "" {
						if !present[pin] {
							add("WARN", "chain %s 步骤%d 钉扎技能 %q 不在本包（跨包引用合法，请确认目标包已安装且技能名无误）", c.Slug, i+1, pin)
						}
						v.mountChecks = append(v.mountChecks, mountCheck{pack: packName, expert: st.Agent, skill: pin, chainSlug: c.Slug})
					}
					if st.Instruction == "" && st.Skill == nil {
						add("WARN", "chain %s 步骤%d 无 instruction 且未钉扎技能", c.Slug, i+1)
					}
				}
			}
		}
	}
}

// checkCrossRefs 全包扫描后核对工具面覆盖：技能 requires_mcp 声明的 server 必须
// 出现在挂载它的专家（或链步骤目标专家）的 tools 授权里——「技能钉了面、专家
// 没授权」是发布后不可达事故（obs-ops 0.3.0 loki-triage 即此形态），CI 必拦。
// 目标专家或技能不在本仓（跨包/平台内置）时无法判定，跳过（存在性另有 WARN）。
func (v *validator) checkCrossRefs() {
	for _, mc := range v.mountChecks {
		if _, ok := v.skillOwner[mc.skill]; !ok {
			continue
		}
		granted, known := v.expertTools[mc.expert]
		if !known {
			continue
		}
		for _, srv := range v.skillNeeds[mc.skill] {
			if slices.Contains(granted, srv) {
				continue
			}
			if mc.chainSlug == "" {
				v.add("ERROR", mc.pack, "专家 %s 挂载技能 %s 声明 requires_mcp server %q，但 agent.yaml tools 未授权该 server", mc.expert, mc.skill, srv)
			} else {
				v.add("ERROR", mc.pack, "链 %s 步骤钉扎技能 %s 声明 requires_mcp server %q，但目标专家 %s 的 tools 未授权", mc.chainSlug, mc.skill, srv, mc.expert)
			}
		}
	}
}

// checkExpertInstall 专家可独立安装性（spec 2026-10-05-per-expert-install §6）：
// 子集安装按闭包合成迷你包，闭包只含本包资产——跨包技能引用、无契约又非宿主面的
// tools server 都让该专家装不出去（安装器会拒），CI 提前拦。
func (v *validator) checkExpertInstall() {
	for slug, ref := range v.expertRefs {
		for _, s := range ref.skills {
			if owner, ok := v.skillOwner[s]; ok && owner != ref.pack {
				v.add("ERROR", ref.pack, "专家 %s 引用跨包技能 %s（属 %s）——专家级子集安装无法携带，须整包安装；如确需跨包挂载请在 provenance 说明并走整包设计", slug, s, owner)
			}
		}
		for _, srv := range ref.tools {
			if v.packMCP[ref.pack][srv] || hostSideMCP[srv] {
				continue
			}
			// 契约清单是工具面事实记录（inventory），平台对缺席容忍（os-basics
			// k8s-health 引 k8sgpt、契约在 k8s-ops 即现状）——降 WARN：迷你包
			// 不带该契约，站点须确认 server 已在平台挂载。
			v.add("WARN", ref.pack, "专家 %s tools server %q 无包内契约（mcp/%s.yaml）且不属宿主面白名单——专家级安装不带该契约清单，站点须确认平台已挂载", slug, srv, srv)
		}
	}
}

// checkChangeRefs 全包扫描后核对 provides_changes 引用闭包：技能声明的编目 slug
// 必须真实存在（本仓任一包）；跨包引用合法但 WARN 提示依赖目标包安装。
// 注意不做工具面交叉核对：受审执行走引擎直取（不经专家 allowlist），与
// requires_mcp 的工具面覆盖语义不同（那是技能采集面的授权）。
func (v *validator) checkChangeRefs() {
	for skill, refs := range v.skillChanges {
		for _, slug := range refs {
			owner, ok := v.changeOwner[slug]
			if !ok {
				v.add("ERROR", v.skillOwner[skill], "技能 %s provides_changes 引用 %q 不存在于本仓任何包（编目引用闭包）", skill, slug)
				continue
			}
			if owner != v.skillOwner[skill] {
				v.add("WARN", v.skillOwner[skill], "技能 %s provides_changes 引用 %q 属于包 %s（跨包引用合法，站点须同时安装）", skill, slug, owner)
			}
		}
	}
}

// changeTainted 变更模板结构层快检（平台 agents.changeTainted 同构镜像：命令替换
// /控制字符/未闭合引号拒；管道/重定向属审批文本合法组成）。
func changeTainted(s string) string {
	inSingle := false
	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if inSingle {
			if r == '\'' {
				inSingle = false
			}
			continue
		}
		if r < 0x20 || r == 0x7f {
			return "control character / newline in command"
		}
		switch r {
		case '\'':
			inSingle = true
		case '`':
			return "backtick (command substitution)"
		case '$':
			if i+1 < len(runes) && runes[i+1] == '(' {
				return "$( (command substitution)"
			}
		}
	}
	if inSingle {
		return "unterminated quote"
	}
	return ""
}

// checkReadme 核对 README 包清单版本表与各 pack.yaml 同步（手工表格是版本纪律
// 最常漂移之处）。约定在仓库根运行；找不到 README.md 时降级 WARN 跳过。
func (v *validator) checkReadme() {
	if len(v.packVersions) == 0 {
		return
	}
	raw, err := os.ReadFile("README.md")
	if err != nil {
		v.add("WARN", "README", "未在当前目录找到 README.md，跳过包清单版本核对（应在仓库根运行）")
		return
	}
	table := map[string]string{}
	for _, m := range readmeRowRe.FindAllStringSubmatch(string(raw), -1) {
		table[m[1]] = m[2]
	}
	if len(table) == 0 {
		v.add("WARN", "README", "README 未识别到包清单表行（| [包名](packs/包名/) | 版本 |），跳过版本核对")
		return
	}
	names := make([]string, 0, len(v.packVersions))
	for name := range v.packVersions {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		ver := v.packVersions[name]
		rv, ok := table[name]
		if !ok {
			v.add("WARN", name, "README 包清单缺 %s 行", name)
			continue
		}
		if rv != ver {
			v.add("ERROR", name, "README 包清单版本 %s 与 pack.yaml %s 不一致（发版须同步 README 表）", rv, ver)
		}
	}
}

// checkIndex index.json 对账（批次一百三十九）：索引是 bq-markettool search/info 的
// 机器可读事实源（Claude Code marketplace.json 同构物）。缺失=WARN（渐进采用，检索
// 侧自动降级现场扫描）；在而漂移=ERROR——多录/漏包/版本不符逼 CI 重新
// `bq-markettool index-gen` 后入库，安装正确性不依赖索引（install 仍现扫目录）。
// 包清单自本容器目录现扫（不依赖跨根累积的 packVersions——多根调用互不污染）。
func (v *validator) checkIndex(packsRoot string) {
	raw, err := os.ReadFile(filepath.Join(filepath.Dir(packsRoot), "index.json"))
	if err != nil {
		if os.IsNotExist(err) {
			v.add("WARN", "index", "index.json 缺失（仓根）——跑 `bq-markettool index-gen -hub <本仓>` 生成并提交（检索面走索引，缺省降级现场扫描）")
			return
		}
		v.add("ERROR", "index", "读取 index.json: %v", err)
		return
	}
	var idx struct {
		Packs []struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"packs"`
	}
	if err := json.Unmarshal(raw, &idx); err != nil {
		v.add("ERROR", "index", "index.json 解析失败: %v", err)
		return
	}
	onDisk := map[string]string{}
	entries, err := os.ReadDir(packsRoot)
	if err != nil {
		return // run 的容器发现已报，此处不双报
	}
	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") || strings.HasPrefix(e.Name(), "_") {
			continue
		}
		rawManifest, err := os.ReadFile(filepath.Join(packsRoot, e.Name(), "pack.yaml"))
		if err != nil {
			continue // 非包目录不校验（与 run 同口径）
		}
		var m packManifest
		if yaml.Unmarshal(rawManifest, &m) != nil || m.Name == "" {
			continue
		}
		onDisk[m.Name] = m.Version
	}
	seen := map[string]bool{}
	for _, p := range idx.Packs {
		if seen[p.Name] {
			v.add("ERROR", "index", "index.json 重复条目 %s", p.Name)
			continue
		}
		seen[p.Name] = true
		ver, ok := onDisk[p.Name]
		if !ok {
			v.add("ERROR", "index", "index.json 多录 %s（packs/ 无此包——重跑 index-gen 刷新）", p.Name)
			continue
		}
		if ver != "" && ver != p.Version {
			v.add("ERROR", p.Name, "index.json 版本漂移：索引 %s / pack.yaml %s（重跑 index-gen 刷新）", p.Version, ver)
		}
	}
	for name := range onDisk {
		if !seen[name] {
			v.add("ERROR", "index", "index.json 漏包 %s（重跑 index-gen 刷新）", name)
		}
	}
}

// dangerousPatterns 发布侧内容扫描（批次一百四十，G4b；ClawHub「SKILL.md 变安装器」
// 投毒案对标）：技能/专家正文是喂给 LLM 的执行面指令。分档依据实弹校准——
// `curl|sh` 在运维指引正文有合法形态（helm/tidb 等工具安装，oo-devops 现存 4 处）
// → 正文 WARN；**changes/ 变更块是受审执行通道的实际执行面 → 任何命中升 ERROR**
// （声明层零容忍）。启发式有边界：只拦明文形态，变形/编码载荷靠平台执行面治理
// （受审命令+审批门）兜底——扫描是内容审的第一道网，不是替代执行面治理。
var dangerousPatterns = []struct {
	re    *regexp.Regexp
	level string
	why   string
}{
	{regexp.MustCompile(`(?i)(curl|wget)\b[^|\n]{0,200}\|\s*(sudo\s+)?(ba|z|da|k)?sh\b`), "WARN", "pipe-to-shell 安装形态（curl|sh——投毒标准载荷；正文指引供复核）"},
	{regexp.MustCompile(`(?im)^\s*rm\s+-rf\s+/(?:\s|$|\*)`), "WARN", "根目录递归删除形态（rm -rf /；示例/反例文档供复核）"},
	{regexp.MustCompile(`(?i)(ignore|disregard|override)\s+(all\s+|any\s+)?(previous|prior|above)\s+(instructions|rules|prompts?)`), "WARN", "英文提示注入短语（ignore previous instructions 形态）"},
	{regexp.MustCompile(`忽略(之前|上面|以上|先前)(的)?(所有)?(系统|历史)?指令|无视(之前|系统|上面)(的)?指令`), "WARN", "中文提示注入短语（忽略之前指令形态）"},
	{regexp.MustCompile(`(?i)(cat|less|more|head|tail)\s+[^|\n;]*(id_rsa|\.aws/credentials|\.kube/config|/etc/shadow|\.ssh/)`), "WARN", "凭证文件直读建议（应走平台凭证面 requires_credentials）"},
}

// scanDangerous 对 packs 文本做危险内容扫描（walk 与 scanSensitive 同口径：跳过
// ./_ 前缀目录、scanExts 文件类型、跨根去重）。changes/ 目录命中一律升 ERROR。
func scanDangerous(roots ...string) []finding {
	var out []finding
	seen := map[string]bool{}
	for _, root := range roots {
		abs, err := filepath.Abs(root)
		if err != nil {
			continue
		}
		filepath.WalkDir(abs, func(p string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() {
				if p != abs && (strings.HasPrefix(d.Name(), ".") || strings.HasPrefix(d.Name(), "_")) {
					return filepath.SkipDir
				}
				return nil
			}
			if seen[p] {
				return nil
			}
			seen[p] = true
			if !scanExts[strings.ToLower(filepath.Ext(p))] {
				return nil
			}
			raw, _ := os.ReadFile(p)
			rel, _ := filepath.Rel(abs, p)
			rel = filepath.ToSlash(rel)
			executable := strings.HasPrefix(rel, "changes/")
			for _, pat := range dangerousPatterns {
				if pat.re.Match(raw) {
					level := pat.level
					if executable {
						level = "ERROR" // 受审执行通道的实际执行面，声明层零容忍
					}
					out = append(out, finding{level, rel, "内容扫描：" + pat.why})
				}
			}
			return nil
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].pack < out[j].pack })
	return out
}

func skillRefName(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case map[string]any:
		if n, ok := t["name"].(string); ok {
			return n
		}
	}
	return ""
}

// readFrontmatter 拆 SKILL.md 的 YAML 围栏（--- ... ---），返回 frontmatter 与正文。
// 行尾统一 LF、去 BOM——Windows 编辑器产出的 CRLF/BOM 文件不误报「缺围栏」。
func readFrontmatter(p string) (front, body string, err error) {
	raw, err := os.ReadFile(p)
	if err != nil {
		return "", "", err
	}
	s := strings.TrimPrefix(strings.ReplaceAll(string(raw), "\r\n", "\n"), "\ufeff")
	if !strings.HasPrefix(s, "---\n") {
		return "", "", fmt.Errorf("缺 frontmatter 起始围栏")
	}
	end := strings.Index(s[4:], "\n---")
	if end < 0 {
		return "", "", fmt.Errorf("缺 frontmatter 结束围栏")
	}
	return s[4 : 4+end], s[4+end+4:], nil
}

// scanSensitive 脱敏扫描（error 级）：多个根去重遍历；隐藏/下划线目录（.git、
// .v2c、_trash）不入；文本类文件全扫（含无扩展名），防凭证藏在白名单外类型里。
func scanSensitive(roots ...string) []finding {
	var out []finding
	seen := map[string]bool{}
	for _, root := range roots {
		abs, err := filepath.Abs(root)
		if err != nil {
			continue
		}
		filepath.WalkDir(abs, func(p string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() {
				if p != abs && (strings.HasPrefix(d.Name(), ".") || strings.HasPrefix(d.Name(), "_")) {
					return filepath.SkipDir
				}
				return nil
			}
			if seen[p] {
				return nil
			}
			seen[p] = true
			if !scanExts[strings.ToLower(filepath.Ext(p))] {
				return nil
			}
			raw, _ := os.ReadFile(p)
			for _, pat := range sensitivePatterns {
				if loc := pat.re.Find(raw); loc != nil {
					rel, _ := filepath.Rel(abs, p)
					out = append(out, finding{"ERROR", rel, "脱敏拦截：" + pat.why})
					break
				}
			}
			return nil
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].pack < out[j].pack })
	return out
}
