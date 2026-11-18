// 本文件由批607 validator 重构拆出（纯代码搬移，判据逐字未动）。
package main

import "slices"

// checkCrossRefs 全包扫描后核对工具面覆盖：技能 requires_mcp 声明的 server 必须
// 出现在挂载它的专家（或链步骤目标专家）的 tools 授权里——「技能钉了面、专家
// 没授权」是发布后不可达事故（obs-ops 0.3.0 loki-triage 即此形态），CI 必拦。
// 目标专家或技能不在本仓（跨包/平台内置）时无法判定，跳过（存在性另有 WARN）。
func (v *validator) checkCrossRefs() {
	for _, mc := range v.mountChecks {
		if _, ok := v.skillOwner[mc.skill]; !ok {
			continue
		}
		granted, known := v.expertTools[mc.expert]
		if !known {
			continue
		}
		for _, srv := range v.skillNeeds[mc.skill] {
			if slices.Contains(granted, srv) {
				continue
			}
			if mc.chainSlug == "" {
				v.add("ERROR", mc.pack, "专家 %s 挂载技能 %s 声明 requires_mcp server %q，但 agent.yaml tools 未授权该 server", mc.expert, mc.skill, srv)
			} else {
				v.add("ERROR", mc.pack, "链 %s 步骤钉扎技能 %s 声明 requires_mcp server %q，但目标专家 %s 的 tools 未授权", mc.chainSlug, mc.skill, srv, mc.expert)
			}
		}
	}
}

// checkExpertInstall 专家可独立安装性（spec 2026-10-05-per-expert-install §6）：
// 子集安装按闭包合成迷你包，闭包只含本包资产——跨包技能引用、无契约又非宿主面的
// tools server 都让该专家装不出去（安装器会拒），CI 提前拦。

// checkChangeRefs 全包扫描后核对 provides_changes 引用闭包：技能声明的编目 slug
// 必须真实存在（本仓任一包）；跨包引用合法但 WARN 提示依赖目标包安装。
// 注意不做工具面交叉核对：受审执行走引擎直取（不经专家 allowlist），与
// requires_mcp 的工具面覆盖语义不同（那是技能采集面的授权）。
func (v *validator) checkChangeRefs() {
	for skill, refs := range v.skillChanges {
		for _, slug := range refs {
			owner, ok := v.changeOwner[slug]
			if !ok {
				v.add("ERROR", v.skillOwner[skill], "技能 %s provides_changes 引用 %q 不存在于本仓任何包（编目引用闭包）", skill, slug)
				continue
			}
			if owner != v.skillOwner[skill] {
				v.add("WARN", v.skillOwner[skill], "技能 %s provides_changes 引用 %q 属于包 %s（跨包引用合法，站点须同时安装）", skill, slug, owner)
			}
		}
	}
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
