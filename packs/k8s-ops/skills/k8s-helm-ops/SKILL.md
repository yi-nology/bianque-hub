---
name: k8s-helm-ops
description: Helm release 排查方法论：release 状态判读（deployed/failed/pending-*）、升级回滚决策、values 漂移核对。宿主侧只读 CLI 采集（ask-ops 受审通道），诊断只读优先。
mode: on_demand
version: 0.1.3
maturity: experimental
requires_mcp:            # 采集依赖的工具面（装载期对账依据；旧 bianque-tools 二进制缺这些工具时应先重建）
  - server: ask-ops
    tools: [run_readonly_command]
provides_changes:        # 编目变更块（批次九十四「受审执行」）：本技能方法论覆盖的处方编目
  - k8s-helm-rollback    # release 应急回滚（params: release/revision/namespace）——处方优先引用编目而非自由 steps
---

> 改造自 oo-devops/k8s-helm（openocta 收割）：原技能绑定 kubectl-mcp-server 的 16 个
> Helm 工具（该 MCP 面未随包供给）——本版去工具耦合，改为通用 helm CLI 方法论，
> 任何运行时可按此执行。宿主侧 helm 只读子命令（list/history/get/values/show）
> 经 ask-ops `run_readonly_command` 受审执行（子命令在白名单内时；被拒不绕过，
> 降级为方法论输出）；Helm 专用 MCP 工具面接通前不另声明。

## 数据来源（ask-ops 只读面）

- 只读子命令经 ask-ops `run_readonly_command` 受审执行；`helm rollback/uninstall/install/upgrade` 等变更动作不处方化执行，一律进建议面走审批。

## 触发条件

- 症状关键词：helm 排查、release 异常、chart 升级失败、helm 回滚
- 组合场景：Helm 部署后的工作负载异常（k8s-workload-triage 上游）

## 方法论（固定顺序）

1. **状态盘点**：`helm list -A -a`——状态四态判读：deployed=正常；failed=上次操作失败但历史版本可用；pending-install/pending-upgrade/pending-rollback=操作中断（helm 2 兼容锁残留，`helm rollback` 或解除 secret 锁）；
2. **历史核对**：`helm history <release> -n <ns>`——看 revision 链：最近一次 revision 的状态与描述；failed 后 chart 往往停在倒数第二个可用 revision；
3. **回滚决策**：应急回滚走编目变更块 `k8s-helm-rollback`（release/namespace/revision 三参；revision 取 history 链上最后 deployed 的 revision）；回滚≠修复——回滚后必须查失败原因再决定是否重试升级；
4. **values 漂移**：`helm get values <release> -n <ns> --all` 与期望值比对；注意 `--reuse-values` 陷阱：升级未显式传值时沿用旧值，chart 新参数缺省不生效；
5. **渲染比对**：`helm template` + `kubectl diff`（或 get manifest 比对）定位「chart 变了什么」——比读 chart 源码快。

## 判读基准

- failed 常见三因：values 引用错误（模板渲染失败）/镜像 tag 不存在（部署时失败）/CRD 依赖缺失（v3 需先装 CRD）；
- pending-* 残留锁在 release 的 secret 最后一个字符（sh.helm.release.v1.<name>.v<rev>.lock）；
- 多 release 共存时按 namespace 归位，禁止跨 ns 盲操作。

## 输出要求

- 结论附 revision 链证据；回滚动作必须标注「应急回滚，根因待查」并给出后续核对步骤。

**受审执行处方（批次九十四）**：应急回滚处方优先引用编目变更块——报告 `recommendation.change_ref: k8s-helm-rollback` + `change_params: {release: "<release 名>", revision: "<目标 revision，取 history 链上最后 deployed>", namespace: "<ns>"}`（不给自由 steps，命令本体由编目模板固定；变更后 `helm history` 自动验证新 revision 落链）；其余变更动作（uninstall/升级重试/解 pending 锁）仍走自由 steps + requires_approval（审批卡标「未编目」）。编目未安装的站点照常走自由 steps，语义不变。
