---
name: prometheus-triage
description: Prometheus 自诊方法论：target 掉线与采集断点、scrape 超时与乱序、规则与告警失效、TSDB 存储与 remote write 积压四面板——固定顺序定位与判读基准（ask-ops 只读采集面）。
mode: on_demand
version: 0.1.0
maturity: experimental
requires_mcp:            # 采集依赖的工具面（装载期对账依据；旧 bianque-tools 二进制缺这些工具时应先重建）
  - server: ask-ops
    tools: [run_readonly_command, run_readonly_commands]
---

## 触发条件

- 症状关键词：采集断点、target掉线、指标断了、告警失效、prometheus排障、监控失明
- 组合场景：面板无数据先经 grafana-triage 排除展示层，再回本技能查采集层；被采集端业务异常走对应领域包

## 数据来源（ask-ops 只读面）

- Prometheus 自身 API（GET）：`/-/healthy`、`/api/v1/targets?state=active`（每 target 的 health/lastError/lastScrape）、`/api/v1/rules`（规则评估状态与最近错误）；
- 查询面（GET）：`/api/v1/query?query=up`、`scrape_duration_seconds`、`prometheus_tsdb_head_series`、remote write 队列指标（`prometheus_remote_storage_samples_pending` 族）；
- 主机侧：prometheus 进程（内存/CPU/Goroutine 面可用 `curl <prom>/metrics | grep go_goroutines`）、数据目录 df/du、日志尾部；
- 凭证由主机侧已配置环境注入，对话不回显。

## 分诊路径（按面板定位，固定顺序）

1. **target 面**：`/api/v1/targets` 逐 target 看 health=down 的 lastError → 归因三分：**被采集端不通**（connection refused/timeout——被采集服务挂了，转对应领域包并给联动线索）、** exporter 自身死**（端口通但指标面 5xx/无响应）、**配置面**（DNS 解析失败、标签冲突 duplicate sample 报错、证书过期）。
2. **采集质量面**：health=up 但数据断续 → `scrape_duration_seconds` 对 scrape_interval（接近=超时丢点）、日志中「out of order sample」/「duplicate sample for timestamp」（采集端时间回拨或重复标签——多为配置错误）；断续形态要与 target 掉线区分（曲线有洞 vs 全断）。
3. **规则与告警面**：`/api/v1/rules` 看每条规则的 lastError 与 evaluationTime → 规则报错三分：**查询语法/指标缺失**（recording rule 引用了不存在的指标）、**评估超时**（规则过重）、**告警状态机卡死**（pending 永不 firing 的阈值/for 配置问题）。「告警没响」必须先核对规则链（规则正常评估？条件真满足过？发送链路是 Alertmanager 域，属联动线索）。
4. **存储与转发面**：`prometheus_tsdb_head_series` 持续增长（基数爆炸：label 组合失控）→ 内存与 WAL 膨胀、查询变慢连锁；remote write 开启时看 `samples_pending`/`samples_failed` 积压（对端不可达时样本在本地排队，磁盘随之涨）；数据目录水位对 retention 口径。

## 判读基准

- 单 target down 与批量 target 同时 down 处置方向完全不同（单点=被采集端问题；批量=共同网络/Prometheus 本体/采集配置面），必须先分型再归因；
- 「无数据」结论必须写清**指标口径**（哪个 metric、哪个时间窗），禁笼统说「监控挂了」；
- 基数问题给 head_series 数量级口径（百万级 series 才构成结论，十万级看增长率）；
- 重启 Prometheus 会丢内存 head 数据（WAL 之外），恢复动作只建议并标注数据影响。

## 输出要求

- 每个结论附 API/日志关键行证据；变更类动作（改配置、重启、清 WAL、降基数）标注影响面并 requires_approval；数据不足输出「需补充采集」清单（如 scrape 配置原文、Alertmanager 状态），不臆测。
