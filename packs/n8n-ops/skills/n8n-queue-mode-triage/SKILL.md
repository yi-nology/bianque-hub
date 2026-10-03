---
name: n8n-queue-mode-triage
description: n8n 队列模式分诊方法论：消费者四同核（同 Redis/同模式/同版本/同密钥）、worker 不消费归因、队列四指标判读、Redis 面与 webhook processor 路由、multi-main——固定顺序定位与判读基准（datasources n8n 查询面 + ask-ops 只读采集面）。
mode: on_demand
version: 0.1.1
maturity: experimental
requires_mcp:            # 采集依赖的工具面（装载期对账依据；旧 bianque-tools 二进制缺这些工具时应先重建）
  - server: datasources
    tools: [n8n_list_executions]
  - server: ask-ops
    tools: [run_readonly_command, run_readonly_commands]
---

## 触发条件

- 症状关键词：worker 不消费、队列堆积、任务全在排队、queue mode 排障、加 worker 没用、执行不开始、multi-main
- 组合场景：单工作流卡住/失败率先走 n8n-execution-triage；本技能管**消费面结构性问题**（有活没人干/有人不干活）；Redis 本体故障深挖走 mw-ops（本技能只判 n8n 视角的队列连通与水位）

## 数据来源（datasources n8n 查询面 + ask-ops 主机面）

- **执行侧证走 `datasources` 工具面**（`n8n_list_executions` 看 queued/waiting 堆积的
  执行侧分布——队列指标说「堆了多少」，执行列表说「堆的是什么」；未配置 n8n 源时工具报
  「未配置」——转凭证面补 n8n_url + n8n_api_key，不臆测）；
- worker 健康面（ask-ops）：`curl -s http://127.0.0.1:5679/healthz`（worker 端口按部署实情；需
  `QUEUE_HEALTH_CHECK_ACTIVE=true` 才开启，未开启时报连接拒绝是**配置形态**不是死证据）；
- 指标面（ask-ops）：main 与 worker 各自暴露 `/metrics`——官方列名四指标
  `n8n_scaling_mode_queue_jobs_waiting`（gauge）、`_active`（gauge）、`_completed`、
  `_failed`（counter）；multi-main 下另有 `instance_role_leader`（谁是 leader）；
- 进程面：`docker ps`（main/worker/webhook processor 容器清单与版本 tag）、
  `docker inspect <容器>`（env 键名核对：`EXECUTIONS_MODE`、`QUEUE_BULL_REDIS_*`、
  `N8N_ENCRYPTION_KEY` **只核有无/是否一致，值不进报告**）；
- 日志面：`docker logs --tail 200 <worker>`（Redis 连接失败、退出前的
  timeout threshold 记录）。

## 分诊路径（固定顺序，先消费者存在性再队列再依赖）

1. **消费者四同核**（worker 不消费的第一嫌疑组，逐项核对）：
   ① **同 Redis**——`QUEUE_BULL_REDIS_HOST/PORT/DB/用户` 与 main 完全一致（db 号不同=
   静默不消费的经典形态）；② **同模式**——两边 `EXECUTIONS_MODE=queue`；③ **同版本**——
   main/worker/webhook processor 必须同版（**升级后只重启 main 不重启 worker** 是社区
   最高频坑，任务会永远 queued）；④ **同加密密钥**——`N8N_ENCRYPTION_KEY` 不一致时
   worker 取到任务解不开 credentials，执行必挂。
2. **队列指标判读**：waiting 持续增长且 active=0 = 无消费者/消费者全卡（回第 1 步）；
   waiting 涨、active 也涨 = 消费力不足（并发与 worker 数配比问题，扩容属建议面）；
   `_failed` 增速突增转执行面技能看错误样本。
3. **Redis 面与退出形态**：Redis 宕机时 worker 在 `QUEUE_BULL_REDIS_TIMEOUT_THRESHOLD`
   （默认 10000ms）后**自行退出**——worker 大批消失的时间线与 Redis 事件对齐即定性；
   Redis 内存水位要预留 webhook 响应中转（`N8N_WEBHOOK_RESPONSE_RELAY_SIZE_MAX` 默认
   64MiB，按 ~1.5× 预算；超大响应 2.34+ 可开 offload 到 S3/Azure，属建议面）。
4. **并发与背压**：worker 并发是 CLI 启动参数 `--concurrency`（默认 10，官方建议 ≥5），
   进程内并发上限另有 `N8N_CONCURRENCY_PRODUCTION_LIMIT`（-1 不限）；并发调太低 + 多
   worker 会打爆 Postgres 连接池（`DB_POSTGRESDB_POOL_SIZE` 默认 2）——并发/worker 数/
   连接池三方要一起看，配比调整是变更建议。
5. **结构性边界核验**（定性前必查）：queue mode **不支持 sqlite**（sqlite + queue 声明
   本身就是配置错误）；binary data filesystem 模式在 queue mode 下不可用（需外部存储）；
   multi-main（Enterprise）需 `N8N_MULTI_MAIN_SETUP_ENABLED=true` 且全部 main 同版本 +
   LB sticky session，定时任务/清理由 leader 独占——双 main 都在跑定时任务的「重复触发」
   先核 leader 归属。

## 判读基准

- 「不消费」结论必须落到四同核的具体一项（哪一项不一致、证据是什么），禁笼统说队列有问题；
- 版本/密钥不一致是**结构性根因**，优先于一切性能类解释；
- worker 重启/扩容/改并发/换 Redis 都是变更动作：标注影响面走审批；
- regular→queue 模式切换前有 pending 执行未清，旧任务不触发是已知形态——切换窗口证据
  优先采信。

## 输出要求

- 每个结论附容器清单/指标行/env 键名核对证据；数据不足输出「需补充采集」清单（如
  docker-compose/helm values 中 queue 段原文、Redis 侧连通性证据），不臆测。
