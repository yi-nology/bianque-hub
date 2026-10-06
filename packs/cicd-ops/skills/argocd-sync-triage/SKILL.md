---
name: argocd-sync-triage
description: ArgoCD 同步分诊方法论：SyncFailed / OutOfSync / Degraded 三态定位、集群连接与凭证面、hook 卡住、RBAC 与 controller 面错位——固定顺序定位与判读基准（ask-ops 只读采集面）。
mode: on_demand
version: 0.1.2
maturity: experimental
requires_mcp:            # 采集依赖的工具面（装载期对账依据；旧 bianque-tools 二进制缺这些工具时应先重建）
  - server: ask-ops
    tools: [run_readonly_command, run_readonly_commands]
provides_changes:        # 编目变更块（批次九十四「受审执行」）：本技能方法论覆盖的处方编目
  - cicd-argocd-app-sync     # 手动同步（params: app）——处方优先引用编目而非自由 steps
  - cicd-argocd-app-rollback # 回滚到历史 revision（params: app/revision）——同上
---

## 触发条件

- 症状关键词：argo同步失败、outofsync、应用不健康、gitops不同步、argocd排障、发布不生效
- 组合场景：同步成功但业务异常是**运行态**问题，交 k8s-ops 工作负载分诊；本技能管 GitOps 交付面

## 数据来源（ask-ops 只读面）

- `argocd app list`（全应用健康/同步态面板）、`argocd app get <app>`（目标态细节：health 状态、sync 状态、条件与事件摘要）；
- `argocd app diff <app>`（只读差异：期望态 vs 实际态）；
- ArgoCD 服务侧（K8s 里部署时经 kubectl 只读）：application-controller / repo-server / server 组件状态与日志尾部；
- 凭证由主机侧已配置 argocd CLI 登录态注入，对话不回显。

## 分诊路径（按症状类定位，固定顺序）

1. **三态定性**：`argocd app list` 圈异常应用 → `app get` 逐个定性：**SyncFailed**=同步动作失败（去第 2/3 步）、**OutOfSync（持续）**=实际态与 Git 期望态不一致（去第 4 步 diff）、**Degraded（health）**=同步成功但资源不健康（运行态问题，交 k8s-ops，给联动线索）。三态混报时按 SyncFailed 优先处理（同步不成立，其余态是次生）。
2. **集群连接与凭证面**：SyncFailed 的 error 含连接类文案（timeout/refused/认证失败）→ 核对目标集群注册状态与凭证有效期（server 侧 cluster 列表证据）；集群凭证轮换后未更新是高发根因。
3. **repo-server 与清单面**：error 含 manifest/render 类失败 → 归因三分：Git 仓库不可达/凭证失效、Helm/Kustomize 渲染错误（模板参数）、CRD 缺失（目标集群无对应 CRD，渲染通过但 apply 会被 API 拒）。
4. **OutOfSync 归因**：`argocd app diff` 逐资源看差异——**手工漂移**（有人直接改了集群，diff 为少量字段）、**控制器观测漂移**（未配 `ignoreDifferences`（应用级）或 `resource.customizations`（argocd-cm 级）豁免 webhook 注入字段/finalizer 的 jsonPointers 时，这类字段噪音常态存在）、**Git 变更未同步**（同步策略 auto 关闭/被暂停，paused 标志）。漂移回滚（sync）会覆盖手工变更，影响面必须写明。
5. **hook 与 controller 面**：Progressing 卡住看 PreSync/PostSync hook 资源状态（hook 卡死=同步流水线悬置）；app-controller 日志的 Reconcile 超时与队列积压（大仓库应用数过多）属上游容量面。

## 判读基准

- OutOfSync 短窗内出现属正常（auto sync 的收敛间隔），持续 >2 个同步周期不收敛才是结论；
- 「同步成功」与「应用健康」是两个正交状态，禁混用；
- sync/rollback 均为变更动作：只建议并标注「覆盖哪些实际态变更」的影响面，恒 requires_approval。

## 输出要求

- 每个结论附 CLI 输出关键行证据；变更类动作（sync、rollback、refresh、更新集群凭证）标注影响面并 requires_approval；数据不足输出「需补充采集」清单（如 controller 日志全文），不臆测。

**受审执行处方（批次九十四）**：sync/rollback 处方优先引用编目变更块——sync 用 `recommendation.change_ref: cicd-argocd-app-sync` + `change_params: {app: "<应用名>"}`；rollback 用 `cicd-argocd-app-rollback` + `change_params: {app: "<应用名>", revision: "<app history 的 rollback-version>"}`（均不给自由 steps；sync 前提是 diff 影响面已进结论；变更后 `app get`/`app history` 自动验证）；refresh、更新集群凭证、暂停自动同步仍走自由 steps + requires_approval（审批卡标「未编目」）。编目未安装的站点照常走自由 steps，语义不变。
