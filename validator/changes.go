// 本文件由批607 validator 重构拆出（纯代码搬移，判据逐字未动）。
package main

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// changeDef 编目变更块（<包>/changes/<slug>.yaml；平台 agents.ChangeDef 同构镜像，
// 校验子集）。治理红线：包侧只声明编目，审批/执行/验证归平台。
type changeDef struct {
	APIVersion          int      `yaml:"api_version"`
	Slug                string   `yaml:"slug"`
	Title               string   `yaml:"title"`
	RequestType         string   `yaml:"request_type"`
	Risk                int      `yaml:"risk"`
	Kind                string   `yaml:"kind"`
	Server              string   `yaml:"server"`
	Tool                string   `yaml:"tool"`
	Rollback            string   `yaml:"rollback"`
	Commands            []string `yaml:"commands"`
	Params              []struct {
		Name   string   `yaml:"name"`
		Kind   string   `yaml:"kind"`
		Values []string `yaml:"values"`
	} `yaml:"params"`
	VerifyReadonlyTools []struct {
		Tool string `yaml:"tool"`
	} `yaml:"verify_readonly_tools"`
}

// changeTainted 变更模板结构层快检（平台 agents.changeTainted 同构镜像：命令替换
// /控制字符/未闭合引号拒；管道/重定向属审批文本合法组成）。
func changeTainted(s string) string {
	inSingle := false
	runes := []rune(s)
	for i := range runes {
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

func (v *validator) checkPackChanges(packDir, packName string, add func(level, format string, a ...any)) {
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
			if cd.Slug == "" || cd.Title == "" || cd.RequestType == "" {
				add("ERROR", "changes/%s 缺 slug/title/request_type（变更块最小契约）", e.Name())
				continue
			}
			if cd.Risk < 1 || cd.Risk > 4 {
				add("ERROR", "changes/%s risk=%d 越界（1-4）", e.Name(), cd.Risk)
			}
			kind := cd.Kind
			if kind == "" {
				kind = "command"
			}
			switch kind {
			case "command":
				if len(cd.Commands) == 0 {
					add("ERROR", "changes/%s kind=command 缺 commands", e.Name())
					continue
				}
				for _, tmpl := range cd.Commands {
					if msg := changeTainted(tmpl); msg != "" {
						add("ERROR", "changes/%s 命令模板非法: %s", e.Name(), msg)
					}
				}
			case "mcp_tool":
				if len(cd.Commands) > 0 {
					add("ERROR", "changes/%s kind=mcp_tool 不能声明 commands", e.Name())
				}
				if cd.Server == "" || cd.Tool == "" {
					add("ERROR", "changes/%s kind=mcp_tool 缺 server/tool", e.Name())
					continue
				}
				all, write := mcpToolSets(packDir)
				if !write[cd.Server][cd.Tool] {
					add("ERROR", "changes/%s mcp_tool %s/%s 不在 mcp/%s.yaml 的 write_tools", e.Name(), cd.Server, cd.Tool, cd.Server)
				}
				for _, call := range cd.VerifyReadonlyTools {
					if call.Tool == "" {
						add("ERROR", "changes/%s verify_readonly_tools 含空 tool", e.Name())
						continue
					}
					if !all[cd.Server][call.Tool] || write[cd.Server][call.Tool] {
						add("ERROR", "changes/%s verify_readonly_tools %s/%s 不是只读工具", e.Name(), cd.Server, call.Tool)
					}
				}
			default:
				add("ERROR", "changes/%s kind=%q 非法（command|mcp_tool）", e.Name(), kind)
			}
			// params[].kind 必须落在平台装载器接受集（agents.ChangeParam*：
			// 缺省 string / string_list / enum 需 values）——批 633 补：此前只校验
			// 条目 kind，kylin-insight 0.4.2 的 kind: list 过检后被平台拒载。
			for _, p := range cd.Params {
				switch p.Kind {
				case "", "string", "string_list":
				case "enum":
					if len(p.Values) == 0 {
						add("ERROR", "changes/%s 参数 %q kind=enum 但缺 values", e.Name(), p.Name)
					}
				default:
					add("ERROR", "changes/%s 参数 %q kind=%q 非法（string|string_list|enum）", e.Name(), p.Name, p.Kind)
				}
			}
			if strings.TrimSpace(cd.Rollback) == "" {
				add("WARN", "changes/%s 缺 rollback（审批卡回退说明，人读必需）", e.Name())
			}
			if prev, dup := v.changeOwner[cd.Slug]; dup {
				add("ERROR", "变更块 slug %q 跨包重复（%s 与 %s；平台装载同规则 fail）", cd.Slug, prev, packName)
			} else {
				v.changeOwner[cd.Slug] = packName
			}
		}
	}

}
