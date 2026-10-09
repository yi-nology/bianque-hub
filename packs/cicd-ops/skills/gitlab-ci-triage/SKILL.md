---
name: gitlab-ci-triage
description: GitLab CI 分诊方法论：token 身份与项目可见性预检、job 卡 created/pending 的派活与执行资源两分、runner 掉线、流水线状态语义（allow_failure/manual 红标≠阻断）、MR/commit→流水线→jobs→trace 尾部下钻、配额与对象存储面——固定顺序定位与判读基准（ask-ops 只读采集面）。
mode: on_demand
version: 0.2.0
maturity: experimental
requires_mcp:            # 采集依赖的工具面（装载期对账依据；旧 bianque-tools 二进制缺这些工具时应先重建）
  - server: ask-ops
    tools: [run_readonly_command, run_readonly_commands]
provides_changes:        # 编目变更块（受审执行）：本技能方法论覆盖的处方编目
  - cicd-gitlab-job-retry        # 作业级单发重跑（params: project/job_id）——粒度与流水线级两分
  - cicd-gitlab-pipeline-retry   # 流水线级重跑全部失败作业（params: project/pipeline_id）
  - cicd-gitlab-pipeline-cancel  # 取消进行中流水线（params: project/pipeline_id）
---

## 触发条件

- 症状关键词：runner掉线、流水线卡住、job一直pending、gitlabci排障、gitlab流水线失败、构建失败、runner未接活、流水线红标、job日志拉不到
- 组合场景：镜像推送失败转 harbor-triage；部署阶段失败转 argocd-sync-triage / k8s-ops；GitLab 平台本体（实例 502/推拉开闭/登录）转 gitlab-ops——本技能管流水线作业执行面

## 数据来源（ask-ops 只读面）

- GitLab REST GET（凭证走主机侧已配置 token；URL 含 `?`/`|`/`%` 特殊字符一律整体单引号包裹；`<id>` 用数字项目 id 或 URL 编码路径 `group%2Fproject`，**不是** CLI 占位符 `:id` 形态）：
  - **身份与可见性预检**：`/api/v4/user`（令牌有效 + 确认活跃身份）、`/api/v4/projects/<id>`（令牌是否看得见该项目——一切流水线面结论的前置）；
  - **流水线面**：`/api/v4/projects/<id>/pipelines`（每条 status/sha/ref）、`/api/v4/projects/<id>/pipelines/<pid>`、`/api/v4/projects/<id>/pipelines/<pid>/jobs`（一次拿全 stage/name/status/allow_failure/when/duration——作业明细视图，勿逐 job 单独 GET）、`/api/v4/projects/<id>/merge_requests/<iid>/pipelines`（MR→流水线关联：把失败钉到变更上）、`'/api/v4/projects/<id>/pipelines?sha=<commit_sha>'`（commit→流水线）；
  - **作业日志**：`/api/v4/projects/<id>/jobs/<jid>/trace`——取失败/已结束作业的**尾部**（错误行在末尾，自行 pipe tail 控量）；running 作业的 trace 持续增长，不流式等全文、先看状态分布；
  - **runner 面**：`'/api/v4/runners?scope=online'`、`'/api/v4/runners?scope=offline'`（runner 明细含 tag/locked/version——tag 不匹配判定依据）、`/api/v4/version`；
- runner 主机侧：`gitlab-runner list` / `gitlab-runner status`、config.toml 的 concurrent 限制、runner 进程与宿主资源；
- GitLab 侧健康：sidekiq 队列延迟证据（管理面，通常只有平台侧可见，无证据时标注）。

## 分诊路径（按症状类定位，固定顺序）

0. **身份与可见性预检（恒为第一动作）**：GET `/user` → GET `/projects/<id>` → GET `pipelines`——状态码语义三分：**401**=令牌失效/被轮换（修凭证面，不进流水线归因）、**403**=令牌有效但看不见该项目（授权 scope 面）、**200 空清单**=真的没流水线。「流水线丢了/查不到 job」大半是没做这步预检。
1. **job 卡 created/pending**：`'/api/v4/runners?scope=online'` 确认在用 runner，再按状态两分——**created**（还没被 runner 接走）：派活面三因，job tag 与 runner tags 不交、runner 未分配到该项目/组（locked 归属）、runner 全下线（转第 2 步）；**pending**（runner 已接走但未开跑）：runner 侧资源面——并发饱和（concurrent 配置口径）或执行器环境启动受阻（查 runner 主机 config.toml 与宿主资源）。tag 不匹配是最常见的「pending 永动机」根因。
2. **runner 掉线**：runner 主机 `gitlab-runner status` + 进程/容器状态 → 归因三分：进程死（重启是变更建议）、认证失效（注册令牌被轮换/解除注册，重新注册是变更）、宿主资源不足（联动 os-basics）。批量 runner 同时掉线优先查共同上游（网络/注册中心/GitLab 侧变更）。
3. **流水线级失败分类**：先状态语义再读日志——`allow_failure=true` 的红标与 `manual` 未触发**不构成阻断**（「流水线绿、job 红」即此因）；真阻断 = `allow_failure=false` 的 failed、到期的 manual。阻断作业读 trace 尾部三分——**环境面**（clone 失败：Git 通道/网络，联动 gitlab-ops 推拉面；依赖拉取失败：网络/缓存）、**真失败**（测试/编译断言，交业务修复，不处方重跑）、**资源面**（runner 磁盘满、job 超时 trace 被截断标志）。
4. **配额与令牌面**：项目 CI 分钟数/配额耗尽时 job 直接失败（错误文案证据）；共享 runner vs 自建 runner 归属核对（配额只影响共享面）；runner 明细显示令牌失效回第 2 步认证面。
5. **日志与对象存储面**：trace 端点 404/空返回/超时指向 GitLab 外部依赖的存储面——构建本身可能正常，归因到存储面，不要误判构建失败。

## 判读基准

- pending 结论必须给出「等待时长 + runner 可用性证据」双要素；短时排队是正常并发形态；
- **状态语义先于红标**：阻断判定 = `allow_failure=false` failed / 到期 manual；`allow_failure=true` 红标与未触发 manual 不进阻断清单；
- **created 与 pending 是两个面**：created=派活/匹配面（tag/归属/下线），pending=runner 执行资源面（并发/执行器）——查完第 1 步两分再归因，不混谈；
- **空清单先过身份预检**：「项目没有流水线」的结论只在第 0 步全 200 后可下；403/401 不许转写成「流水线丢失」；
- runner offline 判据是注册心跳超时口径（通常数分钟），瞬时抖动不构成结论；
- CI 分钟数配额只约束共享 runner：自建 runner 不受影响，核对归属后再下结论。

## 输出要求

- 每个结论附 API/日志关键行证据，引用 pipeline_id/job_id/ref/sha 使结论可追溯；数据不足输出「需补充采集」清单（如 job 完整 trace、project id），不臆测。
- **受审执行处方（重跑/取消一律粒度先行）**：处置优先引用编目变更块——单失败作业重跑 → `recommendation.change_ref: cicd-gitlab-job-retry` + `change_params: {project: "<id>", job_id: "<作业id>"}`；流水线级重跑全部失败作业 → `cicd-gitlab-pipeline-retry` + `{project, pipeline_id}`；叫停进行中流水线 → `cicd-gitlab-pipeline-cancel` + `{project, pipeline_id}`。**粒度不混发**：单 job 重跑勿走流水线级（重触发无关作业、消耗共享并发与配额），job 级只取 job id、流水线级只取 pipeline id（GitLab 官方 CLI 把 `ci retry` 钉在 job 粒度、流水线级另走 pipelines 端点——同型口径）。重跑处方的前提 = 第 3 步分类为环境面/资源面瞬态失败；真失败交业务修复，不处方重跑。编目未覆盖的动作（重启 runner、重新注册、调 concurrent、清队列、轮转令牌）进 `recommendation.steps`，requires_approval 恒 true（审批卡标「未编目」）；编目未安装的站点照常走自由 steps，语义不变。
