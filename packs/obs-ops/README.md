# obs-ops：可观测自诊包（社区版）

「监控系统的监控」——当监控/日志/看板栈自己病了（监控失明），本包接管诊断。
三技能 + 一专家，全部只读采集。

| 资产 | 说明 |
|---|---|
| `skills/prometheus-triage` | target 掉线三分、采集质量（超时/乱序）、规则与告警失效、TSDB 基数与 remote write 积压四面板（oo-devops prometheus/prometheus-alerts 合并重写） |
| `skills/grafana-triage` | 面板无数据三层反推（数据源→查询→时间范围）、插件面、查询性能面（oo-devops grafana-ops 面重写） |
| `skills/es-triage` | 集群红黄定性、allocation explain 五类归因、磁盘水位三级保护（flood_stage 只读锁）、压力面（oo-devops elasticsearch-ops/elk-ops 合并重写） |
| `obs-analyst/` | 可观测自诊专家（ask-ops 只读面） |

## 路由错位（重要）

- **用监控数据查业务问题**（现在的负载、某指标曲线）→ 平台监控查询入口；
- **被监控的业务系统异常** → 对应领域包（os-basics / k8s-ops / mw-ops / db-ops）；
- **监控栈本体故障**（target 掉线、面板无数据、ES 集群红）→ 本包。
- 三条入口用复合路由词（监控失明/采集断点/面板无数据…）与平台查询面词
  （prometheus/监控指标/指标查询）刻意错开，避免互相劫持。

本包无链：监控栈故障是事件驱动排障（失明场景），无例检语义，走专家路由即可。

## 工具面

无专用 MCP 依赖：各组件 REST GET / `_cat` API 经平台 `ask-ops` 采集面在目标运维
主机执行，凭证走主机侧已配置环境，对话不回显（数据源凭证字段一律脱敏后引用）。

## 安装

```bash
go run ./cmd/bq-markettool install --url <本仓> --pack obs-ops --api <扁鹊实例>
```

## 改造说明

源自 openocta 收割的 oo-devops 市场包。原版技能假定 PROMETHEUS_URL /
REDIS_EXPORTER_URL 等环境变量注入与未供给的 MCP 工具面——本版全部剥离，改为
ask-ops 只读采集 + 明确的 API 面口径。loki/skywalking/zabbix 等技能族暂未搬运，
按需在后续版本承接。
