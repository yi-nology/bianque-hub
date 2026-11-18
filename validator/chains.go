// 本文件由批607 validator 重构拆出（纯代码搬移，判据逐字未动）。
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

type chainStep struct {
	Type        string `yaml:"type"`
	Agent       string `yaml:"agent"`
	Instruction string `yaml:"instruction"`
	Skill       any    `yaml:"skill"`
}

// graphNode 图模型链的节点（v3；spec 2026-10-22 §5.1）。
//
// 只镜像**内容契约**要用到的字段：链内执行单元引用谁、钉扎哪个技能、
// 有没有指令。图的拓扑（边/环/可达性/变量引用/预算）由平台装载期深判，
// 本仓不重复实现（见 chainNodeViews）。

// graphNode 图模型链的节点（v3；spec 2026-10-22 §5.1）。
//
// 只镜像**内容契约**要用到的字段：链内执行单元引用谁、钉扎哪个技能、
// 有没有指令。图的拓扑（边/环/可达性/变量引用/预算）由平台装载期深判，
// 本仓不重复实现（见 chainNodeViews）。
type graphNode struct {
	ID          string `yaml:"id"`
	Type        string `yaml:"type"`
	Title       string `yaml:"title"`
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
	// Version 链形状版本：0/缺省 = v2 树模型（steps）；≥3 = 图模型（nodes+edges）。
	Version int         `yaml:"version"`
	Nodes   []graphNode `yaml:"nodes"`
}

// isGraph 是否图模型链（与平台 ChainDef.IsGraph 同口径：version≥3 或有节点）。

// isGraph 是否图模型链（与平台 ChainDef.IsGraph 同口径：version≥3 或有节点）。
func (c chainDef) isGraph() bool { return c.Version >= 3 || len(c.Nodes) > 0 }

// chainNodeView 链内一个执行单元的归一视图：v2 步骤与 v3 节点共用同一套内容判据。

// chainNodeView 链内一个执行单元的归一视图：v2 步骤与 v3 节点共用同一套内容判据。
type chainNodeView struct {
	pos         string // 定位串：v2「步骤1」、v3「节点 n1」
	kind        string
	agent       string
	instruction string
	skill       any
}

// chainNodeViews 把链的步骤/节点归一成同一套判据的视图。
//
// v2（steps）与 v3（nodes+edges，spec 2026-10-22 §5.1）是**两种链形状**，
// 但**内容契约同源**：链内执行单元只能引用叶子专家、禁 engine 专家、
// 技能钉扎须能在本包找到。平台侧的次序也是「先按 v2 原样校验再迁移」，
// 故本仓对两种形状用同一套内容判据。
//
// 图的**拓扑**判据（边指向不存在的节点 / 有环 / 可达性 / 变量引用 /
// 节点与层数预算）**不在这里重复实现**——那是平台装载期的深判，
// 在此重写一份就是两份真相，且必然漂移。

// chainNodeViews 把链的步骤/节点归一成同一套判据的视图。
//
// v2（steps）与 v3（nodes+edges，spec 2026-10-22 §5.1）是**两种链形状**，
// 但**内容契约同源**：链内执行单元只能引用叶子专家、禁 engine 专家、
// 技能钉扎须能在本包找到。平台侧的次序也是「先按 v2 原样校验再迁移」，
// 故本仓对两种形状用同一套内容判据。
//
// 图的**拓扑**判据（边指向不存在的节点 / 有环 / 可达性 / 变量引用 /
// 节点与层数预算）**不在这里重复实现**——那是平台装载期的深判，
// 在此重写一份就是两份真相，且必然漂移。
func chainNodeViews(c chainDef) []chainNodeView {
	if c.isGraph() {
		out := make([]chainNodeView, 0, len(c.Nodes))
		for _, n := range c.Nodes {
			kind := n.Type
			if kind == "" {
				kind = "agent"
			}
			pos := n.ID
			if pos == "" {
				pos = "?" // 缺 id 是平台装载期的判据；此处不重复判，只为定位不空
			}
			out = append(out, chainNodeView{
				pos:         "节点 " + pos,
				kind:        kind,
				agent:       n.Agent,
				instruction: n.Instruction,
				skill:       n.Skill,
			})
		}
		return out
	}
	out := make([]chainNodeView, 0, len(c.Steps))
	for i, st := range c.Steps {
		kind := st.Type
		if kind == "" {
			kind = "agent"
		}
		out = append(out, chainNodeView{
			pos:         fmt.Sprintf("步骤%d", i+1),
			kind:        kind,
			agent:       st.Agent,
			instruction: st.Instruction,
			skill:       st.Skill,
		})
	}
	return out
}

func (v *validator) checkPackChain(packDir, packName string, expertSlugs map[string]string, present map[string]bool, add func(level, format string, a ...any)) {
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
			views := chainNodeViews(c)
			if len(views) == 0 {
				if c.isGraph() {
					add("ERROR", "chain %s 无节点", c.Slug)
				} else {
					add("ERROR", "chain %s 无步骤", c.Slug)
				}
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
			for _, nv := range views {
				if nv.kind != "agent" {
					continue
				}
				if strings.HasPrefix(nv.agent, "workflow/") {
					add("ERROR", "chain %s %s 引用 %q 非法：链步骤只能是叶子专家（治理闭环由平台衔接）", c.Slug, nv.pos, nv.agent)
				}
				if k, ok := expertSlugs[nv.agent]; ok && k == "engine" {
					add("ERROR", "chain %s %s 引用 engine 专家 %q（链步骤禁 engine）", c.Slug, nv.pos, nv.agent)
				} else if !ok {
					add("WARN", "chain %s %s 引用 %q 不在本包（跨包引用合法，请确认目标包已安装）", c.Slug, nv.pos, nv.agent)
				}
				if pin := skillRefName(nv.skill); pin != "" {
					if !present[pin] {
						add("WARN", "chain %s %s 钉扎技能 %q 不在本包（跨包引用合法，请确认目标包已安装且技能名无误）", c.Slug, nv.pos, pin)
					}
					v.mountChecks = append(v.mountChecks, mountCheck{pack: packName, expert: nv.agent, skill: pin, chainSlug: c.Slug})
				}
				if nv.instruction == "" && nv.skill == nil {
					add("WARN", "chain %s %s 无 instruction 且未钉扎技能", c.Slug, nv.pos)
				}
			}
		}
	}
}
