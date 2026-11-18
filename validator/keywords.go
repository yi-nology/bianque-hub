// 本文件由批607 validator 重构拆出（纯代码搬移，判据逐字未动）。
package main

import "fmt"

// kwFirst 路由词首登记录（平台 packs.go registerRouteKeyword 同构镜像）。
type kwFirst struct{ prio, pack string }

// registerKeyword 镜像平台跨包冲突规则：同包 OK；跨包同优先级 = 真歧义（ERROR，
// 平台装载会失败——PR 必须在 CI 拦下）；跨层 = 警告（高优先级胜出）。
// 局限：hub 只见本仓包，与平台内置域词（_shared/os-basics 内置版）的冲突在此
// 查不到，靠 CONTRIBUTING 路由词纪律人工把关。

// registerKeyword 镜像平台跨包冲突规则：同包 OK；跨包同优先级 = 真歧义（ERROR，
// 平台装载会失败——PR 必须在 CI 拦下）；跨层 = 警告（高优先级胜出）。
// 局限：hub 只见本仓包，与平台内置域词（_shared/os-basics 内置版）的冲突在此
// 查不到，靠 CONTRIBUTING 路由词纪律人工把关。
func registerKeyword(kwOwner map[string]kwFirst, kw, pack, prio string) (level, msg string, dup bool) {
	first, hit := kwOwner[kw]
	if !hit {
		kwOwner[kw] = kwFirst{prio: prio, pack: pack}
		return "", "", false
	}
	if first.pack == pack {
		return "", "", false
	}
	if first.prio == prio {
		return "ERROR", fmt.Sprintf("关键词/症状 %q 在 %s 与 %s 同优先级 %s 重复（真歧义，平台装载会失败）", kw, first.pack, pack, prio), true
	}
	return "WARN", fmt.Sprintf("关键词/症状 %q 跨层重复（%s@%s 与 %s@%s），高优先级胜出", kw, first.pack, first.prio, pack, prio), true
}

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
