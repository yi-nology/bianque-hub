// 本文件由批607 validator 重构拆出（纯代码搬移，判据逐字未动）。
package main

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
)

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
	slices.SortFunc(out, func(a, b finding) int { return strings.Compare(a.pack, b.pack) })
	return out
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
	slices.SortFunc(out, func(a, b finding) int { return strings.Compare(a.pack, b.pack) })
	return out
}

// checkPackAssets 校验 P1 新增包资产：knowledge/（manifest + frontmatter）与
// render/charts.yaml（受限路径 + 图表类型）；另核对 mcp write_tools 只引用已声明工具。
// 与扁鹊平台 agents.loadPackKnowledge/loadPackRenderRules 同口径，CI 先拦。
