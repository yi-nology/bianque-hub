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

func (v *validator) checkPackSkills(packDir, packName string, m packManifest, add func(level, format string, a ...any)) map[string]bool {
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
				v.skillVersions[sf.Name] = sf.Version
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
	return present
}
