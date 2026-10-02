// bianque-hub 结构契约校验器（独立于扁鹊平台，CI 与本地同款）。
//
// 校验分四层：pack.yaml 契约门、专家/技能/链结构规则、跨包唯一性、脱敏扫描。
// 规则与扁鹊 internal/agents 装载校验（packs.go）同源，但独立实现——平台私仓
// 不外泄，规则漂移以本文件注释为准绳同步。
//
// 用法：go run ./validator ./packs [更多目录...]
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
	kinds     = map[string]bool{"engine": true, "llm": true, "stub": true}
	maturities = map[string]bool{"experimental": true, "stable": true, "frozen": true, "deprecated": true}
	skillModes = map[string]bool{"static": true, "on_demand": true}
	prioRe     = regexp.MustCompile(`^P[0-6]$`)
	semverRe   = regexp.MustCompile(`^\d+\.\d+\.\d+(-[0-9A-Za-z.-]+)?$`)
	// 脱敏（error 级）：内网域名/内部工具名/密钥形态/私网 IP（127.0.0.1 与 0.0.0.0 豁免）。
	sensitivePatterns = []struct {
		re   *regexp.Regexp
		why  string
	}{
		{regexp.MustCompile(`git\.enjoye\.top`), "内网模块域名"},
		{regexp.MustCompile(`kyaiops`), "内部工具名（kyaiops）"},
		{regexp.MustCompile(`-----BEGIN [A-Z ]*PRIVATE KEY`), "私钥形态"},
		{regexp.MustCompile(`AKIA[0-9A-Z]{16}`), "AWS AccessKey 形态"},
		{regexp.MustCompile(`\b(10\.\d+\.\d+\.\d+|172\.(1[6-9]|2\d|3[01])\.\d+\.\d+|192\.168\.\d+\.\d+)\b`), "私网 IP"},
	}
)

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

type agentDef struct {
	Slug           string   `yaml:"slug"`
	Name           string   `yaml:"name"`
	Kind           string   `yaml:"kind"`
	PromptFile     string   `yaml:"prompt_file"`
	PromptIncludes []string `yaml:"prompt_includes"`
	Skills         []any    `yaml:"skills"`
	RoutePriority  string   `yaml:"route_priority"`
	RouteGroup     string   `yaml:"route_group"`
	RouteKeywords  []string `yaml:"route_keywords"`
	Symptoms       []string `yaml:"symptoms"`
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
	RouteKeywords []string    `yaml:"route_keywords"`
	Symptoms      []string    `yaml:"symptoms"`
	Steps         []chainStep `yaml:"steps"`
}

type skillFront struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Mode        string `yaml:"mode"`
	Version     string `yaml:"version"`
	Maturity    string `yaml:"maturity"`
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
	return "WARN", fmt.Sprintf("关键词/症状 %q 跨层重复（%s@%s 与 %s@%s），高优先级胜出", kw, first.pack, first.prio, pack, first.prio), true
}

func main() {
	dirs := os.Args[1:]
	if len(dirs) == 0 {
		dirs = []string{"./packs"}
	}
	var findings []finding
	skillOwner := map[string]string{} // 技能名 → 包（跨包唯一性）
	expertOwner := map[string]string{} // 专家 slug → 包
	chainOwner := map[string]string{}  // 链 slug → 包
	kwOwner := map[string]kwFirst{}    // 路由词/症状 → 首登（ prio, pack）

	for _, root := range dirs {
		entries, err := os.ReadDir(root)
		if err != nil {
			fmt.Printf("ERROR 读取目录 %s: %v\n", root, err)
			os.Exit(1)
		}
		for _, e := range entries {
			if !e.IsDir() || strings.HasPrefix(e.Name(), ".") || strings.HasPrefix(e.Name(), "_") {
				continue
			}
			packDir := filepath.Join(root, e.Name())
			if _, err := os.Stat(filepath.Join(packDir, "pack.yaml")); err != nil {
				continue // 非包目录（无 pack.yaml）不校验
			}
			findings = append(findings, checkPack(packDir, e.Name(), skillOwner, expertOwner, chainOwner, kwOwner)...)
		}
	}
	findings = append(findings, scanSensitive(dirs...)...)

	errN, warnN := 0, 0
	for _, f := range findings {
		fmt.Printf("%-5s [%s] %s\n", f.level, f.pack, f.msg)
		if f.level == "ERROR" {
			errN++
		} else {
			warnN++
		}
	}
	fmt.Printf("\n校验完成：%d error / %d warning\n", errN, warnN)
	if errN > 0 {
		os.Exit(1)
	}
}

func checkPack(packDir, packName string, skillOwner, expertOwner, chainOwner map[string]string, kwOwner map[string]kwFirst) []finding {
	var f []finding
	add := func(level, format string, a ...any) {
		f = append(f, finding{level, packName, fmt.Sprintf(format, a...)})
	}

	// pack.yaml 契约门
	raw, err := os.ReadFile(filepath.Join(packDir, "pack.yaml"))
	if err != nil {
		return []finding{{"ERROR", packName, "读 pack.yaml: " + err.Error()}}
	}
	var m packManifest
	if err := yaml.Unmarshal(raw, &m); err != nil {
		return []finding{{"ERROR", packName, "pack.yaml 解析: " + err.Error()}}
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

	// 专家目录
	expertSlugs := map[string]string{} // slug → kind
	packSkillNames := map[string]bool{}
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
		if prev, dup := expertOwner[d.Slug]; dup {
			add("ERROR", "专家 slug %q 跨包重复（%s 已占用）", d.Slug, prev)
		} else {
			expertOwner[d.Slug] = packName
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
		if len(d.RouteKeywords) == 0 {
			add("WARN", "%s 无 route_keywords（路由不可达，孤立专家）", d.Slug)
		}
		for _, kw := range append(slices.Clone(d.RouteKeywords), d.Symptoms...) {
			if level, msg, dup := registerKeyword(kwOwner, kw, packName, d.RoutePriority); dup {
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
		for _, s := range d.Skills {
			name := skillRefName(s)
			if name == "" {
				continue
			}
			packSkillNames[name] = true // 引用存在性在技能扫描后统一判定
		}
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
			if sf.Version != "" && !semverRe.MatchString(sf.Version) {
				add("ERROR", "skills/%s version=%q 非 SemVer", name, sf.Version)
			}
			if sf.Maturity != "" && !maturities[sf.Maturity] {
				add("ERROR", "skills/%s maturity=%q 非法", name, sf.Maturity)
			}
			if strings.TrimSpace(body) == "" {
				add("ERROR", "skills/%s 正文为空", name)
			}
			if prev, dup := skillOwner[sf.Name]; dup {
				add("ERROR", "技能名 %q 跨包重复（%s 已占用；技能名是全局唯一空间）", sf.Name, prev)
			} else if sf.Name != "" {
				skillOwner[sf.Name] = packName
			}
		}
	}
	for _, s := range m.Provides.Skills {
		if !present[s] {
			add("ERROR", "provides.skills 声明 %q 但 skills/%s/SKILL.md 不存在", s, s)
		}
	}
	for _, s := range m.Provides.Experts {
		if _, ok := expertSlugs[s]; !ok {
			add("ERROR", "provides.experts 声明 %q 但包内无此 slug", s)
		}
	}
	// 专家技能引用存在性（包内或声明 requires 跨包时降级 warn）
	for slug, names := range collectSkillRefs(packDir) {
		for _, n := range names {
			if !present[n] {
				add("WARN", "%s 引用技能 %q 不在本包（跨包引用合法，请确认目标包已安装）", slug, n)
			}
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
			if prev, dup := chainOwner[c.Slug]; dup {
				add("ERROR", "chain slug %q 跨包重复（%s 已占用）", c.Slug, prev)
			} else if c.Slug != "" {
				chainOwner[c.Slug] = packName
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
			for _, kw := range append(slices.Clone(c.RouteKeywords), c.Symptoms...) {
				if level, msg, dup := registerKeyword(kwOwner, kw, packName, c.RoutePriority); dup {
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
					if pin := skillRefName(st.Skill); pin != "" && !present[pin] {
						add("WARN", "chain %s 步骤%d 钉扎技能 %q 不在本包（跨包引用合法，请确认目标包已安装且技能名无误）", c.Slug, i+1, pin)
					}
					if st.Instruction == "" && st.Skill == nil {
						add("WARN", "chain %s 步骤%d 无 instruction 且未钉扎技能", c.Slug, i+1)
					}
				}
			}
		}
	}
	return f
}

// collectSkillRefs 重读专家 yaml 收集 slug→技能名（SkillRef 裸串/对象双形态）。
func collectSkillRefs(packDir string) map[string][]string {
	out := map[string][]string{}
	entries, _ := os.ReadDir(packDir)
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(packDir, e.Name(), "agent.yaml"))
		if err != nil {
			continue
		}
		var d agentDef
		if yaml.Unmarshal(raw, &d) != nil {
			continue
		}
		for _, s := range d.Skills {
			if n := skillRefName(s); n != "" {
				out[d.Slug] = append(out[d.Slug], n)
			}
		}
	}
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
func readFrontmatter(p string) (front, body string, err error) {
	raw, err := os.ReadFile(p)
	if err != nil {
		return "", "", err
	}
	s := string(raw)
	if !strings.HasPrefix(s, "---\n") {
		return "", "", fmt.Errorf("缺 frontmatter 起始围栏")
	}
	end := strings.Index(s[4:], "\n---")
	if end < 0 {
		return "", "", fmt.Errorf("缺 frontmatter 结束围栏")
	}
	return s[4 : 4+end], s[4+end+4:], nil
}

// scanSensitive 全仓脱敏扫描（error 级）。
func scanSensitive(roots ...string) []finding {
	var out []finding
	for _, root := range roots {
		filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			if ext := filepath.Ext(p); ext != ".md" && ext != ".yaml" && ext != ".yml" && ext != ".json" {
				return nil
			}
			raw, _ := os.ReadFile(p)
			for _, pat := range sensitivePatterns {
				if loc := pat.re.Find(raw); loc != nil {
					rel, _ := filepath.Rel(root, p)
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
