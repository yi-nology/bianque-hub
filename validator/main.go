// bianque-hub 结构契约校验器（独立于扁鹊平台，CI 与本地同款）。
//
// 校验分六层：pack.yaml 契约门、专家/技能/链结构规则、跨包唯一性与工具面覆盖、
// 装包闸对账（变更面授权知情同意〔批四四三/四四四〕与 depends_on 环境前提〔批四四二〕——
// 平台在装包/升级时判定，不满足即拒；hub 在 CI 提前拦）、
// 版本同步（CHANGELOG 头部条目与 README 包清单表）、脱敏扫描。
// 规则与扁鹊 internal/agents 装载校验（packs.go）同源，但独立实现——平台私仓
// 不外泄，规则漂移以本文件注释为准绳同步。
//
// # 文件布局（批607 拆分；判据逐字未动，validator 输出与拆分前逐字节一致）
//
//	main.go        编排器（mode/validator 核心结构 + main/parseArgs/runWithMode
//	               + checkPack 薄编排 + checkPackManifestGate 契约门）
//	manifest 相关常量（api_version/插件名法度/semver/正则）也在本文件头部——
//	它们是「契约常量」，与主仓 pack-lint（对账型判据）共享的事实：改这里须同步
//	主仓 internal/tooling/packlint 与 PACK_GUIDE。
//	experts.go     专家目录判据 + provides/experts 装配核对 + 独立安装性
//	skills.go      技能 frontmatter 判据 + frontmatter/skillRef 工具
//	changes.go     changes/*.yaml 契约门 + changeTainted 模板快检
//	chains.go      chain.yaml（v2 steps / v3 graph 双形态）判据 + 节点视图
//	depends.go     depends_on 环境前提 + 版本区间判据
//	userconfig.go  user_config 声明面校验（镜像主仓批598E 同判据）
//	crossrefs.go   跨包核对（技能钉扎工具面覆盖 / 变更块 slug 唯一）
//	repoassets.go  仓级资产（README 包清单表 / index.json 对账）
//	scans.go       脱敏扫描 + 危险命令扫描
//	assets.go      包资产域（render/charts/挂载核对 + 平台耦合文件 + mcp 契约集）
//	keywords.go    路由关键词跨包唯一性登记
//
// # 与主仓 pack-lint 的判据分工（两份判据同管包内容，分工面必须清楚）
//
//   - 本校验器 = **内容契约**（独立可跑）：包内容 vs PACK_GUIDE 契约。
//   - 主仓 internal/tooling/packlint = **登记面对账**（只能在主仓跑）：
//     工具面 registry 对账 / 破坏性工具单源镜像对账 / skills-lock 核对——
//     它判的是「内容 vs 平台登记面」，没有平台登记面就判不了。
//   - 共享契约常量（api_version=1、插件名法度、SemVer 形态）两仓各有一份：
//     漂移形态 = 同一份 pack.yaml 两边裁决不一致。改任一侧须对表另一侧
//     （本文件头部常量 ↔ 主仓 internal/agents packs.go / tooling/packlint）。
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
	pluginNameRe      = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{0,128}$`)
	renderPathRe      = regexp.MustCompile(`^\$(\.[A-Za-z_][A-Za-z0-9_]*|\[\d+\]|\[\*\])*$`)
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

// destructiveTools 平台侧变更面（破坏性）工具单源镜像（bianque
// internal/service/common/lint.go DestructiveTools——批次四百四十三包级工具授权面
// ToolSurface 与批次四百四十四装包知情同意闸的判定面）。对齐语义：agent.yaml tools 的
// allow 空 = 该 server 全量授予（registry 契约），聚合面按最坏记 server/*；显式 allow
// 命中清单内工具名按具体破坏面记。任一命中 ⇒ 该包装/升级被平台闸拦（管理员须显式带
// allow_destructive_tools 知情同意）。hub 侧除登记豁免的设计内执行档外一律 ERROR——
// 装包被拦是发布后事故（bq-markettool API 通道不带 consent 参数，装机必 400），CI 先拦。
var destructiveTools = map[string]map[string]bool{
	"ask-ops": {"run_approved_command": true, "run_approved_commands": true},
	"terminal-assistant": {
		"shell_execute": true, "ssh_execute": true, "terminal_session": true,
		"service_manager": true, "file_manager": true, "config_manager": true,
		"package_manager": true, "system_manager": true,
	},
	"browser-ops": {"browser_click": true, "browser_type": true, "browser_eval": true},
}

// destructiveExemptions 设计内执行档豁免登记（键 = 包名/专家 slug，值 = 理由，进 CI
// 输出供复核）。豁免只放行 hub CI，不改变平台行为——该包装/升级时站点仍须勾选知情
// 同意。新增破坏面授权必须在此登记理由，否则 CI 拦（防静默扩权）。
var destructiveExemptions = map[string]string{
	"web-ops/imports/web-inspector": "browser-ops 执行档三工具（click/type/eval）是技能钉扎的设计能力" +
		"（spa-render-diagnosis 的 eval 只读三连等），prompt 自带执行档纪律；browser-ops server " +
		"治理缺省硬关（装包≠开洞）——批次二百六十四既有决策，安装时须站点知情同意",
}

// ---- 结构 ----

// ---- 结构 ----

type packManifest struct {
	APIVersion  int                    `yaml:"api_version"`
	Name        string                 `yaml:"name"`
	Version     string                 `yaml:"version"`
	Description string                 `yaml:"description"`
	Changelog   string                 `yaml:"changelog"`
	DependsOn   *packDepends           `yaml:"depends_on"`
	UserConfig  []packUserConfigOption `yaml:"user_config"`
	// RequiresCredentials 凭证类型前提（警告级声明，与主仓 PackManifest 同名同构；
	// 批607 披露面普及：消费专用凭证面的包必须如实声明）。
	RequiresCredentials []string `yaml:"requires_credentials"`
	Provides            struct {
		Experts []string `yaml:"experts"`
		Skills  []string `yaml:"skills"`
	} `yaml:"provides"`
}

// packUserConfigOption pack.yaml user_config[] 项（批601，形态与 bianque 批598E
// PackUserConfigOption 同构；key/label/type/required/sensitive/default）。

// UnmarshalYAML 兼容裸串形态（平台 skill.Decl 同款双形态）。
func (s *packDepSkill) UnmarshalYAML(unmarshal func(any) error) error {
	var bare string
	if err := unmarshal(&bare); err == nil {
		s.Name = bare
		return nil
	}
	type raw packDepSkill
	var r raw
	if err := unmarshal(&r); err != nil {
		return err
	}
	*s = packDepSkill(r)
	return nil
}

type finding struct {
	level string // ERROR | WARN
	pack  string
	msg   string
}

// kwFirst 路由词首登记录（平台 packs.go registerRouteKeyword 同构镜像）。

// mode 校验模式：判据按「被校验对象是什么」分档。
//
// **hub 仓根**与**内嵌快照**（扁鹊主仓里由 `bq-markettool seed-sync` 同步来的包副本）
// 是两种对象，仓级资产判据只对前者成立：
//
//   - `provenance.json` 的「必须有」——它是**市场包**的来源标记，内嵌快照按内容资产
//     单源纪律本就不该有（它是 hub `packs/` 的副本，不是市场包）；
//   - `index.json`——hub 仓根的**检索索引**（`bq-markettool search/info` 的面），
//     内嵌快照不做检索；
//   - README 包清单表——hub 仓级的版本纪律表，内嵌快照没有那张表。
//
// 混用判据的后果是**恒红**：同一份完全合法的内容换个位置校验就报 error，
// 于是「用 validator 校验主仓 seed」这条路走不通——而那恰恰是最有价值的用法
// （seed 是同步产物，「同步有没有把内容弄坏」正是需要机械校验的地方）。
// 恒红会教会人忽略 validator，与「假红是门禁的头号死法」同一条道理。
//
// 故加模式判据，**不做自动判定**：自动判定在「两仓并排维护者布局」下容易误判
// （hub 仓里也可能有一份 seed 副本），显式 flag 不会猜错。
type mode string

const (
	// modeHub hub 仓根（默认）：全量校验，含仓级资产判据。
	modeHub mode = "hub"
	// modeEmbedded 内嵌快照：跳过仓级资产判据；**包内容判据一条不减**
	// （内容判据才是「同步有没有把内容弄坏」的答案）。
	modeEmbedded mode = "embedded"
)

// validator 全仓校验状态：跨包唯一性登记 + 全包扫描后才能判的延迟核对项。

// validator 全仓校验状态：跨包唯一性登记 + 全包扫描后才能判的延迟核对项。
type validator struct {
	mode         mode
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
	// 装包闸对账输入（批次四四二/四四四）：技能名 → 本仓现值版本；包名 → depends_on 声明
	skillVersions map[string]string
	packDependsOn map[string]*packDepends
}

// expertRef 专家闭包引用快照（checkExpertInstall 的核对输入）。

func (v *validator) add(level, pack, format string, a ...any) {
	v.findings = append(v.findings, finding{level, pack, fmt.Sprintf(format, a...)})
}

func main() {
	m, dirs, err := parseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "validator: %v\n", err)
		os.Exit(2) // 用法错与「校验不过」分档：2 = 用法错，1 = 校验不过
	}
	if len(dirs) == 0 {
		dirs = []string{"./packs"}
	}
	findings, errN, warnN := runWithMode(dirs, m)
	for _, f := range findings {
		fmt.Printf("%-5s [%s] %s\n", f.level, f.pack, f.msg)
	}
	fmt.Printf("\n校验完成：%d error / %d warning\n", errN, warnN)
	if errN > 0 {
		os.Exit(1)
	}
}

// parseArgs 解析命令行：`--mode=hub|embedded`（默认 hub）+ 位置参数（校验目录）。
//
// 不用 flag 包：flag.Parse 遇首个位置参数即停，
// `validator ./packs --mode=embedded` 里的模式开关会被当成目录参数吞掉。

// parseArgs 解析命令行：`--mode=hub|embedded`（默认 hub）+ 位置参数（校验目录）。
//
// 不用 flag 包：flag.Parse 遇首个位置参数即停，
// `validator ./packs --mode=embedded` 里的模式开关会被当成目录参数吞掉。
func parseArgs(args []string) (mode, []string, error) {
	m := modeHub
	var dirs []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		var raw string
		switch {
		case a == "--mode" || a == "-mode":
			if i+1 >= len(args) {
				return "", nil, fmt.Errorf("%s 缺值（可选 hub|embedded）", a)
			}
			i++
			raw = args[i]
		case strings.HasPrefix(a, "--mode="):
			raw = strings.TrimPrefix(a, "--mode=")
		case strings.HasPrefix(a, "-mode="):
			raw = strings.TrimPrefix(a, "-mode=")
		case strings.HasPrefix(a, "-"):
			return "", nil, fmt.Errorf("未知参数 %q（可用：--mode=hub|embedded）", a)
		default:
			dirs = append(dirs, a)
			continue
		}
		switch raw {
		case string(modeHub):
			m = modeHub
		case string(modeEmbedded):
			m = modeEmbedded
		default:
			return "", nil, fmt.Errorf("未知模式 %q（可选 hub|embedded）", raw)
		}
	}
	return m, dirs, nil
}

// run 执行全部校验层（hub 模式；独立于 main 以便测试）。

// run 执行全部校验层（hub 模式；独立于 main 以便测试）。
func run(dirs []string) ([]finding, int, int) {
	return runWithMode(dirs, modeHub)
}

// runWithMode 按模式执行全部校验层：包发现 → 包级校验 → 跨包核对 →
// 版本同步核对 → 脱敏扫描，返回 findings 与 error/warning 计数。

// runWithMode 按模式执行全部校验层：包发现 → 包级校验 → 跨包核对 →
// 版本同步核对 → 脱敏扫描，返回 findings 与 error/warning 计数。
func runWithMode(dirs []string, m mode) ([]finding, int, int) {
	v := &validator{
		mode:          m,
		skillOwner:    map[string]string{},
		skillNeeds:    map[string][]string{},
		changeOwner:   map[string]string{},
		skillChanges:  map[string][]string{},
		expertOwner:   map[string]string{},
		expertTools:   map[string][]string{},
		chainOwner:    map[string]string{},
		kwOwner:       map[string]kwFirst{},
		packVersions:  map[string]string{},
		packMCP:       map[string]map[string]bool{},
		expertRefs:    map[string]expertRef{},
		skillVersions: map[string]string{},
		packDependsOn: map[string]*packDepends{},
	}
	if m == modeEmbedded {
		// 模式提示放在**最前**：跑的人第一眼就该知道这轮判据面比 hub 模式窄，
		// 免得把「没报」当成「都过了」。
		v.add("WARN", "mode", "内嵌快照模式：已跳过仓级资产判据"+
			"（provenance.json 的「必须有」要求 / index.json 对账 / README 包清单表）"+
			"——它们是 hub 仓级资产，不是包内容契约；"+
			"包内容判据（契约门/版本同步/跨包唯一性/工具面覆盖/脱敏/危险命令）一条未减")
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
		} else if v.mode == modeHub {
			v.checkIndex(root)
		}
	}
	v.checkCrossRefs()
	v.checkExpertInstall()
	v.checkChangeRefs()
	v.checkDepends()
	if v.mode == modeHub {
		v.checkReadme()
	}
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

	m := v.checkPackManifestGate(packDir, packName, add)
	expertSlugs, skillRefs, servers := v.checkPackExperts(packDir, packName, add)
	v.checkPackCredentialDisclosure(packName, m.RequiresCredentials, servers, add)
	present := v.checkPackSkills(packDir, packName, m, add)
	v.checkPackProvides(packDir, packName, m, expertSlugs, skillRefs, present, add)
	v.checkPackChanges(packDir, packName, add)
	v.checkPackCoupledFiles(packDir, expertSlugs, add)
	v.checkPackChain(packDir, packName, expertSlugs, present, add)
	v.checkPackAssets(packDir, packName, add)
}

// checkCrossRefs 全包扫描后核对工具面覆盖：技能 requires_mcp 声明的 server 必须
// 出现在挂载它的专家（或链步骤目标专家）的 tools 授权里——「技能钉了面、专家
// 没授权」是发布后不可达事故（obs-ops 0.3.0 loki-triage 即此形态），CI 必拦。
// 目标专家或技能不在本仓（跨包/平台内置）时无法判定，跳过（存在性另有 WARN）。

func (v *validator) checkPackManifestGate(packDir, packName string, add func(level, format string, a ...any)) packManifest {
	// pack.yaml 契约门
	raw, err := os.ReadFile(filepath.Join(packDir, "pack.yaml"))
	if err != nil {
		add("ERROR", "读 pack.yaml: %v", err)
		return packManifest{}
	}
	var m packManifest
	if err := yaml.Unmarshal(raw, &m); err != nil {
		add("ERROR", "pack.yaml 解析: %v", err)
		return packManifest{}
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
	// 平台插件名法度（bianque 批次四百四十三装包面 plugin_svc 对齐）：不合即装包拒收。
	if m.Name != "" && !pluginNameRe.MatchString(m.Name) {
		add("ERROR", "pack.yaml name=%q 不合平台插件名法度（^[a-z0-9][a-z0-9._-]{0,128}$，装包即拒）", m.Name)
	}
	// depends_on 环境前提声明（批次四四二装包闸）：登记后跨包核对（checkDepends）。
	if m.DependsOn != nil {
		v.packDependsOn[packName] = m.DependsOn
	}
	// user_config 声明面校验（批601，镜像 bianque 批598E validatePackUserConfig 同判据：
	// 源头挡非法声明——hub CI 拒收后包进不了市场，平台装载期是第二道同判据闸）。
	checkPackUserConfig(add, m.UserConfig)
	// provenance.json 是**市场包**的来源标记：hub 仓根的每个包都要有（装进扁鹊
	// 靠它识别为市场版），内嵌快照按内容资产单源纪律本就不该有（它是 hub packs/
	// 的副本）。故「必须有」这条要求只在 hub 模式成立；在场就照常校验内容
	// （内嵌副本若带了 provenance.json，坏了照样要报）。
	provPath := filepath.Join(packDir, "provenance.json")
	if _, err := os.Stat(provPath); err != nil {
		if v.mode == modeHub {
			add("ERROR", "缺 provenance.json（hub 模式要求：市场包来源标记，装进扁鹊靠它识别为市场版；校验内嵌快照请用 --mode=embedded）")
		}
	} else if pj, err := os.ReadFile(provPath); err == nil {
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

	return m
}
