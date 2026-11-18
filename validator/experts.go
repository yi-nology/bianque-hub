// 本文件由批607 validator 重构拆出（纯代码搬移，判据逐字未动）。
package main

import (
	"os"
	"path/filepath"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

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

// expertRef 专家闭包引用快照（checkExpertInstall 的核对输入）。
type expertRef struct {
	pack   string
	skills []string
	tools  []string
}

// mountCheck 记一处「专家↔技能」绑定：chainSlug 空 = 专家挂载，非空 = 链步骤钉扎。

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

func (v *validator) checkPackExperts(packDir, packName string, add func(level, format string, a ...any)) (map[string]string, map[string][]string, map[string]bool) {
	// 专家目录
	expertSlugs := map[string]string{}   // slug → kind
	skillRefs := map[string][]string{}   // slug → 挂载技能名（存在性核对用）
	consumedServers := map[string]bool{} // 包内专家 tools 引用过的 server 集合（披露面判据用，批607）
	expertDirs, _ := os.ReadDir(packDir)
	for _, e := range expertDirs {
		if !e.IsDir() || e.Name() == "skills" || e.Name() == "prompts" || e.Name() == "mcp" || e.Name() == "credentials" || e.Name() == "knowledge" || e.Name() == "render" {
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
			// 变更面（破坏性）授权对账（批次四四三/四四四装包知情同意闸镜像）：
			// allow 空 = 该 server 全量授予，含破坏性工具的 server 按最坏记 server/*，
			// 装包/升级被平台闸拦（bq-markettool API 通道不带 consent 参数，装机必 400）。
			if len(t.Allow) == 0 {
				if destructiveTools[t.Server] != nil {
					add("ERROR", "%s tools 对 %s 用空 allow 全量授予（含破坏性受审执行，装包被知情同意闸拦）——请显式列出只读白名单（批次二百六十四收敛范式）", d.Slug, t.Server)
				} else {
					add("WARN", "%s tools 对 %s 空 allow=全量授予（新工具注册时静默扩权）——请显式列出工具", d.Slug, t.Server)
				}
			} else {
				// 显式 allow：命中破坏面清单 ⇒ 未登记豁免 ERROR（新增破坏面授权必须
				// 在 destructiveExemptions 登记理由）；已登记 ⇒ WARN 保持知情同意要求可见。
				var hits []string
				for _, tool := range t.Allow {
					if destructiveTools[t.Server][tool] {
						hits = append(hits, tool)
					}
				}
				if len(hits) > 0 {
					if _, ok := destructiveExemptions[packName+"/"+d.Slug]; ok {
						add("WARN", "%s 显式授予变更面工具 %s.%s（已登记设计内执行档豁免）——该包装/升级时站点须知情同意（allow_destructive_tools）", d.Slug, t.Server, strings.Join(hits, "/"))
					} else {
						add("ERROR", "%s 显式授予变更面（破坏性）工具 %s.%s——装包/升级被知情同意闸拦；若系设计内执行档请在 validator destructiveExemptions 登记理由", d.Slug, t.Server, strings.Join(hits, "/"))
					}
				}
			}
			toolServers = append(toolServers, t.Server)
		}
		v.expertTools[d.Slug] = toolServers
		for _, srv := range toolServers {
			consumedServers[srv] = true
		}
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

	return expertSlugs, skillRefs, consumedServers
}

func (v *validator) checkPackProvides(packDir, packName string, m packManifest, expertSlugs map[string]string, skillRefs map[string][]string, present map[string]bool, add func(level, format string, a ...any)) {
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

}
