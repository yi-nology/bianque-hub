# n8n-ops：n8n 自托管平台运维包（社区版）

当自动化平台的引擎自己病了——实例起不来、工作流执行卡住、webhook 不触发、
worker 不消费、升级/密钥出事故——本包接管诊断。五技能 + 一专家，全部只读采集。

| 资产 | 说明 |
|---|---|
| `skills/n8n-instance-triage` | 实例本体分诊：健康端点两级语义（/healthz vs /healthz/readiness）、起不来/反复重启首错定位、资源占用与 task runner 隔离、日志面（N8N_LOG_*） |
| `skills/n8n-execution-triage` | 执行面分诊：卡住三态归因（queued 无消费者 / waiting 堆积 / running 不结束）、失败面聚类、执行历史膨胀与 prune 三件套核验 |
| `skills/n8n-webhook-triage` | 触发面分诊：404 not registered 三查（active 状态 / test 与 production 注册语义 / 强制重注册）、Forbidden maybe CSRF 反代三件套、queue 模式 webhook 路由错位 |
| `skills/n8n-queue-mode-triage` | 队列模式分诊：消费者四同核（同 Redis/同模式/同版本/同加密密钥）、队列四指标判读（n8n_scaling_mode_queue_jobs_*）、Redis 背压与并发/连接池配比、multi-main |
| `skills/n8n-lifecycle-ops` | 生命周期运维：加密密钥不可丢失语义与轮换纪律、备份三件套完整性（CLI 导出 ≠ 完整恢复）、升级路径与 2.x/3.x 破坏性变更基线、SQLite→Postgres 迁移、license 排障 |
| `n8n-platform-analyst/` | n8n 平台诊断专家（ask-ops 只读面） |

## 路由错位（重要）

- **编写/调试工作流内容**（节点怎么配、表达式怎么写）→ 使用侧问题，不走本包；
- **底座异常**：K8s 容器/调度面 → k8s-ops；PostgreSQL/Redis 本体 → db-ops/mw-ops；
- **n8n 平台本体故障**（执行卡住、webhook 404、worker 不消费）→ 本包。
- 入口词全部带 n8n/自动化平台/工作流限定词，与 cicd-ops（流水线）、mw-ops（消息队列）
  语义相邻域刻意错开，避免互相劫持。

本包无链：n8n 故障是事件驱动排障（触发/执行/队列三面），无例检语义，走专家路由即可。

## 工具面

无专用 MCP 依赖：健康端点（`/healthz`、`/healthz/readiness`）、`/metrics`、公共 API
（`/api/v1/*`，`X-N8N-API-KEY` 走主机侧已配置环境）与容器/进程面（docker/kubectl 只读
子命令）统一经平台 `ask-ops` 采集面在目标运维主机执行，凭证不回显。`docker exec` 与
n8n CLI（export/import/license 类）不在只读白名单，一律作为审批后宿主侧动作进建议面。

## 版本基线

方法论以 **2.x 稳定线**为基线（task runners 默认启用、MySQL/MariaDB 已移除、
`publish:workflow` 取代 `update:workflow`），并收录 3.0 破坏性变更预警（npm 装法退役、
`binaryData`→`storage` 更名、internal task runner 废弃）。n8n 约每周发一个 minor、无 LTS，
官方建议至少每月升级——版本对齐（main/worker/webhook processor 同版）是队列模式的
硬约束。

## 安装

```bash
go run ./cmd/bq-markettool install --url <本仓> --pack n8n-ops --api <扁鹊实例>
```

## 来源说明

本包为**原生编写**（非 oo-devops 血缘）：方法论基于 n8n 官方自托管文档
（docs.n8n.io 的 deploy/host-n8n 系列：queue mode / task runners / 环境变量 /
备份升级 / 安全基线）与社区排障案例（community.n8n.io 高频故障帖）整理重写，
剥离一切环境注入假设，适配 ask-ops 只读采集面与扁鹊输出协议。
官方文档 URL 2026 年已迁移到 `/deploy/host-n8n/` 路径（旧 `/hosting/` 部分已 404）。
