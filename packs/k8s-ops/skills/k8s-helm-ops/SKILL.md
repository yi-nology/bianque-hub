---
name: k8s-helm-ops
description: Helm release 排查方法论：release 状态判读（deployed/failed/pending-*）、升级回滚决策、values 漂移核对。纯方法论（宿主侧 helm CLI），诊断只读优先。
mode: on_demand
version: 0.1.0
maturity: experimental
---

> 改造自 oo-devops/k8s-helm（openocta 收割）：原技能绑定 kubectl-mcp-server 的 16 个
> Helm 工具（该 MCP 面未随包供给）——本版去工具耦合，改为通用 helm CLI 方法论，
> 任何运行时可按此执行；后续平台接通 Helm 工具面时再补 requires_mcp 声明。

## 触发条件

- 症状关键词：helm 排查、release 异常、chart 升级失败、helm 回滚
- 组合场景：Helm 部署后的工作负载异常（k8s-workload-triage 上游）

## 方法论（固定顺序）

1. **状态盘点**：`helm list -A -a`——状态四态判读：deployed=正常；failed=上次操作失败但历史版本可用；pending-install/pending-upgrade/pending-rollback=操作中断（helm 2 兼容锁残留，`helm rollback` 或解除 secret 锁）；
2. **历史核对**：`helm history <release> -n <ns>`——看 revision 链：最近一次 revision 的状态与描述；failed 后 chart 往往停在倒数第二个可用 revision；
3. **回滚决策**：应急优先 `helm rollback <release> <n> -n <ns>`（回到最后 deployed 的 revision）；回滚≠修复——回滚后必须查失败原因再决定是否重试升级；
4. **values 漂移**：`helm get values <release> -n <ns> --all` 与期望值比对；注意 `--reuse-values` 陷阱：升级未显式传值时沿用旧值，chart 新参数缺省不生效；
5. **渲染比对**：`helm template` + `kubectl diff`（或 get manifest 比对）定位「chart 变了什么」——比读 chart 源码快。

## 判读基准

- failed 常见三因：values 引用错误（模板渲染失败）/镜像 tag 不存在（部署时失败）/CRD 依赖缺失（v3 需先装 CRD）；
- pending-* 残留锁在 release 的 secret 最后一个字符（sh.helm.release.v1.<name>.v<rev>.lock）；
- 多 release 共存时按 namespace 归位，禁止跨 ns 盲操作。

## 输出要求

- 结论附 revision 链证据；回滚动作必须标注「应急回滚，根因待查」并给出后续核对步骤。
