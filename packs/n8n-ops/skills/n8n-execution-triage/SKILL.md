---
name: n8n-execution-triage
description: n8n 执行面分诊方法论：卡住执行三态归因（queued 无消费者 / waiting 堆积 / running 不结束）、失败面判读、执行历史膨胀与 prune 三件套——固定顺序定位与判读基准（datasources n8n 查询面 + ask-ops 只读采集面）。
mode: on_demand
version: 0.2.1
maturity: experimental
requires_mcp:            # 采集依赖的工具面（装载期对账依据；旧 bianque-tools 二进制缺这些工具时应先重建）
  - server: datasources
    tools: [n8n_list_executions, n8n_get_execution, n8n_list_workflows]
  - server: ask-ops
    tools: [run_readonly_command, run_readonly_commands]
---

## 触发条件

- 症状关键词：工作流卡住、执行卡住、一直在排队、waiting 堆积、执行失败、失败率涨、执行历史膨胀、n8n 数据库涨
- 组合场景：卡住若伴随「队列深度持续增长/全部执行都在排队」优先转 n8n-queue-mode-triage（消费面结构性问题）；单工作流异常才在本技能深挖

## 数据来源（datasources n8n 查询面 + ask-ops 主机面）

- **n8n 公共 API 统一走 `datasources` 工具面**（`n8n_list_executions` / `n8n_get_execution` /
  `n8n_list_workflows`，契约见 frontmatter；API key 经凭证面注入，对话不收集/回显）：
  `n8n_list_executions`（`status` 过滤 error/success/running/waiting、`workflow_id`、`limit`
  缺省 20 上限 100）看近态分布；`n8n_get_execution` 取单执行明细（含节点级错误与 runData）；
  `n8n_list_workflows` 看 active 面与数量；
- 指标面（ask-ops）：`curl -s http://127.0.0.1:5678/metrics`（需 `N8N_METRICS=true`）——队列四指标
  `n8n_scaling_mode_queue_jobs_waiting / _active / _completed / _failed`（官方文档列名）；
- DB 侧水位（sqlite 文件 du / postgres 表体积走 db-ops 联动口径）：execution 数据是库体积第一大户；
- 未配置 n8n 源时工具报「未配置（BQ_DATASOURCES_CONFIG）」——转凭证面补 n8n_url +
  n8n_api_key，不臆测实例状态。

## 分诊路径（固定顺序，先分布后单点再治理）

1. **分布定性**：拉近端 executions 的 status 分布——**全局排队**（大量 queued/waiting、
   正常执行也在堆）转队列模式技能；**单工作流集中失败**在本技能继续；**总量正常但偶发失败**
   走第 3 步看错误样本。
2. **卡住三态归因**（「卡住」必须先分态，三态根因方向完全不同）：
   - **queued 持续不消费** = 无可用 worker / worker 版本不一致 / Redis 队列断（队列面，转
     n8n-queue-mode-triage）；
   - **waiting 大量堆积** = Wait/Schedule 类节点挂起等待是设计内，但 v2.16 起 Wait 节点有
     已知堆积 bug 形态（社区高频）：堆积量与等待时长超设计预期才定性异常，处置（缩短等待/
     重启 worker）为建议面；
   - **running 长时不结束** = 单执行死循环/无超时保护：核对 `EXECUTIONS_TIMEOUT`（默认 -1
     不限）与 `EXECUTIONS_TIMEOUT_MAX`（默认 3600s）是否生效，Code 节点长循环是高发点；
     **queue 模式注意**：该超时由 worker 进程执行——env 只配在 main 容器、worker 容器
     未同配是社区高频「超时不生效」根因（与本技能「四同核」口径一致，核 worker 侧）。
   手动「Stop」停不掉的执行是已知 bug 形态（queue mode 下不释放 worker）——如实记录，
   恢复手段（重启 worker / 受控更新执行状态）全部进建议面。
3. **失败面**：error 样本按**错误节点 + 错误消息**聚类（credential 失效、对端 4xx/5xx、
   超时、数据形状不符）；错误消息原文是 API 证据，逐类给处置方向，禁只报数量。
4. **膨胀治理核对**（数据库涨的归因收口）：prune 三件套——`EXECUTIONS_DATA_PRUNE`
   （默认 true）、`EXECUTIONS_DATA_MAX_AGE`（默认 336 小时=14 天）、
   `EXECUTIONS_DATA_PRUNE_MAX_COUNT`（默认 10000，0=不限）；软删/硬删两阶段
   （软删间隔默认 60s、硬删间隔 15，完成后 1 小时 buffer 才硬删）。判读「prune 是否真生效」：
   看**现存最老 execution 的时间戳**是否 ≈ MAX_AGE 口径，而不是只看开关。
   减量选项 `EXECUTIONS_DATA_SAVE_ON_SUCCESS=none` 与 `EXECUTIONS_DATA_SAVE_ON_PROGRESS`
   （默认 false，开启会放大写入）按需列为建议。清历史属变更动作：requires_approval。

## 判读基准

- 「卡住」禁当单一结论——必须落到三态之一并给对应证据；
- 失败率突增先对齐**变更窗口**（近期升级/凭证轮换/对端变更）再归因代码；
- 数据库涨先问「涨的是什么」：execution 历史（本技能）vs binary data（外部存储配置面），
  两者处置完全不同；
- 停用再激活工作流、重启实例是恢复类变更（影响所有 active 工作流），标注影响面走审批。

## 输出要求

- 每个结论附 API 返回/指标关键行证据（status 分布数字、错误消息原文、最老执行时间戳）；
  变更类动作（停启工作流、重启 worker、清执行历史、改 prune/超时配置）标注影响面并
  requires_approval；数据不足输出「需补充采集」清单（如 API key 配置状态、部署形态
  regular/queue、prune 环境变量现值），不臆测。
