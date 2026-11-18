// 本文件由批607 validator 重构拆出（纯代码搬移，判据逐字未动）。
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// mountCheck 记一处「专家↔技能」绑定：chainSlug 空 = 专家挂载，非空 = 链步骤钉扎。
type mountCheck struct {
	pack, expert, skill, chainSlug string
}

// checkPackAssets 校验 P1 新增包资产：knowledge/（manifest + frontmatter）与
// render/charts.yaml（受限路径 + 图表类型）；另核对 mcp write_tools 只引用已声明工具。
// 与扁鹊平台 agents.loadPackKnowledge/loadPackRenderRules 同口径，CI 先拦。
func (v *validator) checkPackAssets(packDir, packName string, add func(level, format string, a ...any)) {
	knowledgeDir := filepath.Join(packDir, "knowledge")
	if info, err := os.Stat(knowledgeDir); err == nil && info.IsDir() {
		raw, err := os.ReadFile(filepath.Join(knowledgeDir, "manifest.yaml"))
		if err != nil {
			add("ERROR", "knowledge/manifest.yaml: %v", err)
		} else {
			var m struct {
				APIVersion int    `yaml:"api_version"`
				Pack       string `yaml:"pack"`
				Docs       []struct {
					Path  string `yaml:"path"`
					Layer string `yaml:"layer"`
				} `yaml:"docs"`
			}
			if err := yaml.Unmarshal(raw, &m); err != nil {
				add("ERROR", "knowledge/manifest.yaml 解析: %v", err)
			} else {
				if m.APIVersion != wantAPIVersion {
					add("ERROR", "knowledge/manifest.yaml api_version=%d 与契约 %d 不兼容", m.APIVersion, wantAPIVersion)
				}
				if m.Pack != "" && m.Pack != packName {
					add("ERROR", "knowledge/manifest.yaml pack=%q 与目录名 %q 不一致", m.Pack, packName)
				}
				if len(m.Docs) == 0 {
					add("ERROR", "knowledge/manifest.yaml docs 为空（删除 knowledge/ 目录=零知识，不留空壳）")
				}
				for _, doc := range m.Docs {
					rel := strings.TrimSpace(doc.Path)
					if rel == "" {
						add("ERROR", "knowledge/manifest.yaml 存在空 path")
						continue
					}
					content, rerr := os.ReadFile(filepath.Join(knowledgeDir, filepath.FromSlash(rel)))
					if rerr != nil {
						add("ERROR", "knowledge/%s 读取: %v", rel, rerr)
						continue
					}
					text := strings.ReplaceAll(string(content), "\r\n", "\n")
					if !strings.HasPrefix(text, "---\n") {
						add("ERROR", "knowledge/%s 缺 frontmatter（--- 开头）", rel)
						continue
					}
					rest := text[4:]
					end := strings.Index(rest, "\n---")
					if end < 0 {
						add("ERROR", "knowledge/%s frontmatter 未闭合", rel)
						continue
					}
					var fm struct {
						Title string `yaml:"title"`
						Layer string `yaml:"layer"`
						Scope string `yaml:"scope"`
					}
					if yerr := yaml.Unmarshal([]byte(rest[:end]), &fm); yerr != nil {
						add("ERROR", "knowledge/%s frontmatter 解析: %v", rel, yerr)
						continue
					}
					if strings.TrimSpace(fm.Title) == "" {
						add("ERROR", "knowledge/%s frontmatter 缺 title", rel)
					}
					layer := strings.TrimSpace(fm.Layer)
					if layer == "" {
						layer = strings.TrimSpace(doc.Layer)
					}
					if layer == "" {
						add("ERROR", "knowledge/%s 缺 layer（frontmatter 或 manifest docs.layer）", rel)
					}
					if fm.Scope != "" && fm.Scope != "global" {
						add("ERROR", "knowledge/%s scope=%q 非法（本期仅支持 global）", rel, fm.Scope)
					}
				}
			}
		}
	}

	renderDir := filepath.Join(packDir, "render")
	if info, err := os.Stat(renderDir); err == nil && info.IsDir() {
		raw, err := os.ReadFile(filepath.Join(renderDir, "charts.yaml"))
		if err != nil {
			add("ERROR", "render/charts.yaml: %v", err)
		} else {
			var m struct {
				APIVersion int `yaml:"api_version"`
				Rules      []struct {
					ID    string `yaml:"id"`
					Match struct {
						Tool string `yaml:"tool"`
					} `yaml:"match"`
					Foreach string `yaml:"foreach"`
					Chart   struct {
						Type        string `yaml:"type"`
						Title       string `yaml:"title"`
						Unit        string `yaml:"unit"`
						Series      string `yaml:"series"`
						SeriesLabel string `yaml:"seriesLabel"`
						Items       []struct {
							Label string `yaml:"label"`
							Value string `yaml:"value"`
						} `yaml:"items"`
						Events []struct {
							Time   string `yaml:"time"`
							Label  string `yaml:"label"`
							Desc   string `yaml:"desc"`
							Status string `yaml:"status"`
						} `yaml:"events"`
					} `yaml:"chart"`
				} `yaml:"rules"`
			}
			if err := yaml.Unmarshal(raw, &m); err != nil {
				add("ERROR", "render/charts.yaml 解析: %v", err)
			} else {
				if m.APIVersion != wantAPIVersion {
					add("ERROR", "render/charts.yaml api_version=%d 与契约 %d 不兼容", m.APIVersion, wantAPIVersion)
				}
				if len(m.Rules) == 0 {
					add("ERROR", "render/charts.yaml rules 为空（删除 render/ 目录=零规则，不留空壳）")
				}
				checkPath := func(ruleID, field, value string, required bool) {
					value = strings.TrimSpace(value)
					if value == "" {
						if required {
							add("ERROR", "render/charts.yaml 规则 %s: %s 不能为空", ruleID, field)
						}
						return
					}
					if strings.HasPrefix(value, "$") && !renderPathRe.MatchString(value) {
						add("ERROR", "render/charts.yaml 规则 %s: %s 路径非法 %q", ruleID, field, value)
					}
				}
				seen := map[string]bool{}
				for _, rule := range m.Rules {
					id := strings.TrimSpace(rule.ID)
					if id == "" {
						add("ERROR", "render/charts.yaml 存在空规则 id")
						continue
					}
					if seen[id] {
						add("ERROR", "render/charts.yaml 规则 id 重复: %s", id)
					}
					seen[id] = true
					if strings.TrimSpace(rule.Match.Tool) == "" {
						add("ERROR", "render/charts.yaml 规则 %s 缺 match.tool", id)
					}
					switch rule.Chart.Type {
					case "bar", "line", "pie", "timeline", "topo":
					default:
						add("ERROR", "render/charts.yaml 规则 %s: type=%q 非法", id, rule.Chart.Type)
					}
					checkPath(id, "foreach", rule.Foreach, false)
					checkPath(id, "title", rule.Chart.Title, true)
					checkPath(id, "unit", rule.Chart.Unit, false)
					checkPath(id, "series", rule.Chart.Series, false)
					checkPath(id, "seriesLabel", rule.Chart.SeriesLabel, false)
					for i, item := range rule.Chart.Items {
						checkPath(id, fmt.Sprintf("items[%d].label", i), item.Label, true)
						checkPath(id, fmt.Sprintf("items[%d].value", i), item.Value, true)
					}
					for i, event := range rule.Chart.Events {
						checkPath(id, fmt.Sprintf("events[%d].time", i), event.Time, false)
						checkPath(id, fmt.Sprintf("events[%d].label", i), event.Label, true)
						checkPath(id, fmt.Sprintf("events[%d].desc", i), event.Desc, false)
						checkPath(id, fmt.Sprintf("events[%d].status", i), event.Status, false)
					}
				}
			}
		}
	}

	// write_tools 只能引用同文件 tools 清单里已声明的工具（与平台 loader 同口径）。
	if files, err := os.ReadDir(filepath.Join(packDir, "mcp")); err == nil {
		for _, f := range files {
			if f.IsDir() || !strings.HasSuffix(f.Name(), ".yaml") {
				continue
			}
			raw, rerr := os.ReadFile(filepath.Join(packDir, "mcp", f.Name()))
			if rerr != nil {
				add("ERROR", "mcp/%s: %v", f.Name(), rerr)
				continue
			}
			var doc struct {
				Tools []struct {
					Name string `yaml:"name"`
				} `yaml:"tools"`
				WriteTools []string `yaml:"write_tools"`
			}
			if yerr := yaml.Unmarshal(raw, &doc); yerr != nil {
				add("ERROR", "mcp/%s 解析: %v", f.Name(), yerr)
				continue
			}
			declared := map[string]bool{}
			for _, tool := range doc.Tools {
				declared[tool.Name] = true
			}
			for _, tool := range doc.WriteTools {
				tool = strings.TrimSpace(tool)
				if tool == "" {
					continue
				}
				if !declared[tool] {
					add("ERROR", "mcp/%s write_tools 含未声明工具 %q", f.Name(), tool)
				}
			}
		}
	}
}

// mcpToolSets 读取包内 mcp/*.yaml 的工具面：server → 全部工具 / write_tools。

// mcpToolSets 读取包内 mcp/*.yaml 的工具面：server → 全部工具 / write_tools。
func mcpToolSets(packDir string) (map[string]map[string]bool, map[string]map[string]bool) {
	all := map[string]map[string]bool{}
	write := map[string]map[string]bool{}
	entries, err := os.ReadDir(filepath.Join(packDir, "mcp"))
	if err != nil {
		return all, write
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}
		server := strings.TrimSuffix(entry.Name(), ".yaml")
		raw, rerr := os.ReadFile(filepath.Join(packDir, "mcp", entry.Name()))
		if rerr != nil {
			continue
		}
		var doc struct {
			Tools []struct {
				Name string `yaml:"name"`
			} `yaml:"tools"`
			WriteTools []string `yaml:"write_tools"`
		}
		if yaml.Unmarshal(raw, &doc) != nil {
			continue
		}
		if all[server] == nil {
			all[server] = map[string]bool{}
			write[server] = map[string]bool{}
		}
		for _, tool := range doc.Tools {
			all[server][tool.Name] = true
		}
		for _, tool := range doc.WriteTools {
			write[server][strings.TrimSpace(tool)] = true
		}
	}
	return all, write
}

func (v *validator) checkPackCoupledFiles(packDir string, expertSlugs map[string]string, add func(level, format string, a ...any)) {
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

}
