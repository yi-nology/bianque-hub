// 本文件由批607 validator 重构拆出（纯代码搬移，判据逐字未动）。
package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/Masterminds/semver/v3"
)

// packDepends pack.yaml 顶层 depends_on（bianque 批次四百四十二装包环境前提校验对齐）。
// 平台在 Install/Upgrade 暂存后、激活前判定：技能依赖核目标实例技能库的存在性+版本区间
// （不满足即拒装并删除暂存，allow_missing_deps 才放行）；MCP 依赖只核「已配置且启用」
// （min_version 无版本源不判，属展示面字段）。技能依赖是 skill.Decl 双形态（裸串 /
// {name, version, optional}）；**pack 级 optional 被平台忽略、一律按必需判**——声明即误描。
type packDepends struct {
	Skills []packDepSkill `yaml:"skills"`
	MCP    []packDepMCP   `yaml:"mcp"`
}

type packDepSkill struct {
	Name     string `yaml:"name"`
	Version  string `yaml:"version"`
	Optional bool   `yaml:"optional"`
}

// UnmarshalYAML 兼容裸串形态（平台 skill.Decl 同款双形态）。

type packDepMCP struct {
	Name       string `yaml:"name"`
	MinVersion string `yaml:"min_version"`
}

// mcpServer 工具面声明（agent.yaml 的 tools 与 SKILL.md 的 requires_mcp 同构）。

// mcpServer 工具面声明（agent.yaml 的 tools 与 SKILL.md 的 requires_mcp 同构）。
type mcpServer struct {
	Server string   `yaml:"server"`
	Tools  []string `yaml:"tools"`
	Allow  []string `yaml:"allow"`
}

// checkDepends pack.yaml depends_on 环境前提闭包核对（bianque 批次四百四十二装包闸
// 对齐）：平台装包/升级时对目标实例判定，不满足即拒装（allow_missing_deps 才放行）——
// hub 在 CI 提前拦。
//   - 技能依赖：名字必须能在本仓技能空间解析（本仓都解析不了 ⇒ 从本仓同步的实例
//     更不会有，装包必被拒）；版本区间表达式必须可被平台同款解析器（Masterminds/semver，
//     skillkit skill.VersionInRange 逐字同款语义）解析，且须覆盖本仓现值——区间合法但
//     现值不满足 = 本仓两个包互相不兼容，凡同装必被闸拦。pack 级 optional 被平台忽略
//     （一律按必需判），声明即误描 → WARN。
//   - MCP 依赖：平台只核「已配置且启用」（conf ⊕ 各包声明 ⊕ DB 覆盖层），hub 看不到
//     实例配置面——只做命名法度核对；本仓无任何包声明该 server 且非宿主面时 WARN 提示
//     站点确认。min_version 平台不判（无版本源），不在此做漂移核对。
func (v *validator) checkDepends() {
	packs := make([]string, 0, len(v.packDependsOn))
	for name := range v.packDependsOn {
		packs = append(packs, name)
	}
	slices.Sort(packs)
	for _, pack := range packs {
		deps := v.packDependsOn[pack]
		for _, dep := range deps.Skills {
			if strings.TrimSpace(dep.Name) == "" {
				v.add("ERROR", pack, "depends_on.skills 含空条目（平台判定按缺失处理）")
				continue
			}
			if dep.Optional {
				v.add("WARN", pack, "depends_on.skills %q 声明 optional——pack 级依赖的 optional 被平台忽略（一律按必需判），声明即误描", dep.Name)
			}
			owner, ok := v.skillOwner[dep.Name]
			if !ok {
				v.add("ERROR", pack, "depends_on.skills %q 不存在于本仓任何包——装包环境前提闸会因实例技能库无此技能拒装（跨包依赖请确认目标包在本仓且先装）", dep.Name)
				continue
			}
			if dep.Version == "" {
				continue // 区间空 = 任意版本（skill.VersionInRange 同款语义）
			}
			sat, err := versionInRange(v.skillVersions[dep.Name], dep.Version)
			if err != nil {
				v.add("ERROR", pack, "depends_on.skills %q 版本区间非法: %v（平台按不满足判，装包必拒）", dep.Name, err)
				continue
			}
			if !sat {
				v.add("ERROR", pack, "depends_on.skills %q 版本区间 %q 不满足本仓现值 %s（属 %s）——从本仓同步的实例同装必被环境前提闸拒", dep.Name, dep.Version, v.skillVersions[dep.Name], owner)
			}
		}
		for _, dep := range deps.MCP {
			if strings.TrimSpace(dep.Name) == "" {
				v.add("ERROR", pack, "depends_on.mcp 含空条目")
				continue
			}
			if !toolServerAllowRe.MatchString(dep.Name) {
				v.add("ERROR", pack, "depends_on.mcp name=%q 命名非法（小写中划线）", dep.Name)
			}
			if hostSideMCP[dep.Name] || v.declaredInAnyPack(dep.Name) {
				continue
			}
			v.add("WARN", pack, "depends_on.mcp %q 本仓无包声明且非宿主面——平台只核实例「已配置且启用」（conf 或控制台 MCP 页激活），装包前请确认目标实例已配置", dep.Name)
		}
	}
}

// declaredInAnyPack 该 server 是否被本仓任一包的 mcp/ 契约声明。

// declaredInAnyPack 该 server 是否被本仓任一包的 mcp/ 契约声明。
func (v *validator) declaredInAnyPack(server string) bool {
	for _, set := range v.packMCP {
		if set[server] {
			return true
		}
	}
	return false
}

// versionInRange 平台 skillkit skill.VersionInRange 逐字同款语义（独立实现，漂移以
// 平台侧为准绳同步）：expr 空=任意（恒真）；区间/版本非法返回错误。

// versionInRange 平台 skillkit skill.VersionInRange 逐字同款语义（独立实现，漂移以
// 平台侧为准绳同步）：expr 空=任意（恒真）；区间/版本非法返回错误。
func versionInRange(version, expr string) (bool, error) {
	if strings.TrimSpace(expr) == "" {
		return true, nil
	}
	c, err := semver.NewConstraint(expr)
	if err != nil {
		return false, fmt.Errorf("版本区间非法 %q", expr)
	}
	ver, err := semver.NewVersion(version)
	if err != nil {
		return false, fmt.Errorf("版本 %q 非 SemVer", version)
	}
	return c.Check(ver), nil
}

// changeTainted 变更模板结构层快检（平台 agents.changeTainted 同构镜像：命令替换
// /控制字符/未闭合引号拒；管道/重定向属审批文本合法组成）。
