// 本文件由批607 validator 重构拆出（纯代码搬移，判据逐字未动）。
package main

import "regexp"

// packUserConfigOption pack.yaml user_config[] 项（批601，形态与 bianque 批598E
// PackUserConfigOption 同构；key/label/type/required/sensitive/default）。
type packUserConfigOption struct {
	Key       string `yaml:"key"`
	Label     string `yaml:"label"`
	Type      string `yaml:"type"`
	Required  bool   `yaml:"required"`
	Sensitive bool   `yaml:"sensitive"`
	Default   any    `yaml:"default"`
}

// checkPackUserConfig user_config 声明面校验（批601，判据与 bianque 批598E
// validatePackUserConfig 逐字同源：key 法度/唯一/type 词表/default 类型匹配。
// 源头挡非法声明——hub CI 拒收即进不了市场；平台装载期是第二道同判据闸，
// 两处判据文字互指，改一处必须同步另一处）。

// checkPackUserConfig user_config 声明面校验（批601，判据与 bianque 批598E
// validatePackUserConfig 逐字同源：key 法度/唯一/type 词表/default 类型匹配。
// 源头挡非法声明——hub CI 拒收即进不了市场；平台装载期是第二道同判据闸，
// 两处判据文字互指，改一处必须同步另一处）。
func checkPackUserConfig(add func(level, format string, a ...any), opts []packUserConfigOption) {
	seen := map[string]bool{}
	for i := range opts {
		o := &opts[i]
		if !userConfigKeyRe.MatchString(o.Key) {
			add("ERROR", "pack.yaml user_config[%d]: 键 %q 非法（须 ^[a-z][a-z0-9_]{0,63}$）", i, o.Key)
			continue
		}
		if seen[o.Key] {
			add("ERROR", "pack.yaml user_config: 键 %q 重复声明", o.Key)
			continue
		}
		seen[o.Key] = true
		if !userConfigTypes[o.Type] {
			add("ERROR", "pack.yaml user_config[%s]: 类型 %q 非法（词表 string|bool|number）", o.Key, o.Type)
			continue
		}
		if o.Default == nil {
			continue
		}
		if !userConfigDefaultMatches(o.Type, o.Default) {
			add("ERROR", "pack.yaml user_config[%s]: default 值 %v 与声明类型 %q 不匹配", o.Key, o.Default, o.Type)
		}
	}
}

// userConfigKeyRe 键法度（bianque packs_load.go packUserConfigKeyRe 同源）。
var userConfigKeyRe = regexp.MustCompile(`^[a-z][a-z0-9_]{0,63}$`)

// userConfigTypes 类型词表（bianque packs_load.go packUserConfigTypes 同源）。
var userConfigTypes = map[string]bool{"string": true, "bool": true, "number": true}

// userConfigDefaultMatches default 类型匹配（yaml.v3 解码面：整数=int、
// 小数=float64；bool/string 直判——bianque userConfigDefaultMatches 同源）。

// userConfigDefaultMatches default 类型匹配（yaml.v3 解码面：整数=int、
// 小数=float64；bool/string 直判——bianque userConfigDefaultMatches 同源）。
func userConfigDefaultMatches(typ string, v any) bool {
	switch typ {
	case "string":
		_, ok := v.(string)
		return ok
	case "bool":
		_, ok := v.(bool)
		return ok
	case "number":
		switch v.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
			return true
		}
	}
	return false
}

// packDepends pack.yaml 顶层 depends_on（bianque 批次四百四十二装包环境前提校验对齐）。
// 平台在 Install/Upgrade 暂存后、激活前判定：技能依赖核目标实例技能库的存在性+版本区间
// （不满足即拒装并删除暂存，allow_missing_deps 才放行）；MCP 依赖只核「已配置且启用」
// （min_version 无版本源不判，属展示面字段）。技能依赖是 skill.Decl 双形态（裸串 /
// {name, version, optional}）；**pack 级 optional 被平台忽略、一律按必需判**——声明即误描。
