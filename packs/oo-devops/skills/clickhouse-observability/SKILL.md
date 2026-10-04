---
name: clickhouse-observability
description: OLAP 数据库监控专家 - ClickHouse 监控、查询性能分析、合并健康检测
mode: on_demand
version: 1.1.0
maturity: experimental
---

# ClickHouse 可观测性

## 概述

使用内置系统表、Prometheus 集成、Grafana 仪表板和告警规则为 ClickHouse 设置全面监控。

## 先决条件

- 具有 `system.*` 表访问权限的 ClickHouse 实例
- Prometheus（或兼容的：Grafana Alloy、Victoria Metrics）
- 用于仪表板的 Grafana
- 用于告警的 AlertManager 或 PagerDuty

## 使用说明

### 步骤 1：系统表的关键指标

```sql
-- 实时服务器健康快照
SELECT
    (SELECT count() FROM system.processes) AS running_queries,
    (SELECT value FROM system.metrics WHERE metric = 'MemoryTracking') AS memory_bytes,
    (SELECT value FROM system.metrics WHERE metric = 'Query') AS concurrent_queries,
    (SELECT count() FROM system.merges) AS active_merges,
    (SELECT value FROM system.asynchronous_metrics WHERE metric = 'Uptime') AS uptime_sec;

-- 查询吞吐量（最近一小时，每分钟）
SELECT
    toStartOfMinute(event_time) AS minute,
    count() AS queries,
    countIf(exception_code != 0) AS errors,
    round(avg(query_duration_ms)) AS avg_ms,
    round(quantile(0.95)(query_duration_ms)) AS p95_ms,
    formatReadableSize(sum(read_bytes)) AS total_read
FROM system.query_log
WHERE type IN ('QueryFinish', 'ExceptionWhileProcessing')
  AND event_time >= now() - INTERVAL 1 HOUR
GROUP BY minute ORDER BY minute;

-- 插入吞吐量（最近一小时）
SELECT
    toStartOfMinute(event_time) AS minute,
    count() AS inserts,
    sum(written_rows) AS rows_written,
    formatReadableSize(sum(written_bytes)) AS bytes_written
FROM system.query_log
WHERE type = 'QueryFinish' AND query_kind = 'Insert'
  AND event_time >= now() - INTERVAL 1 HOUR
GROUP BY minute ORDER BY minute;

-- 每个表的 part 数量（合并健康指标）
SELECT database, table, count() AS parts, sum(rows) AS rows,
       formatReadableSize(sum(bytes_on_disk)) AS size
FROM system.parts WHERE active
GROUP BY database, table
HAVING parts > 50
ORDER BY parts DESC;
```

### 步骤 2：Prometheus 集成

**ClickHouse Cloud** 暴露了一个托管的 Prometheus 端点：

```yaml
# prometheus.yml
scrape_configs:
  - job_name: clickhouse-cloud
    metrics_path: /v1/organizations/<ORG_ID>/prometheus
    basic_auth:
      username: <API_KEY_ID>
      password: <API_KEY_SECRET>
    static_configs:
      - targets: ['api.clickhouse.cloud']
    params:
      filtered_metrics: ['true']   # 仅 125 个关键指标
```

**自托管** — 使用 clickhouse-exporter 或内置指标端点：

```yaml
# prometheus.yml
scrape_configs:
  - job_name: clickhouse
    static_configs:
      - targets: ['clickhouse-server:9363']  # 内置 Prometheus 端点
    metrics_path: /metrics
```

```xml
<!-- 在 config.xml 中启用 Prometheus 端点 -->
<prometheus>
    <endpoint>/metrics</endpoint>
    <port>9363</port>
    <metrics>true</metrics>
    <events>true</events>
    <asynchronous_metrics>true</asynchronous_metrics>
</prometheus>
```

### 步骤 3：应用级别指标

```typescript
import { Registry, Counter, Histogram, Gauge } from 'prom-client';

const registry = new Registry();

const queryDuration = new Histogram({
  name: 'clickhouse_query_duration_seconds',
  help: 'ClickHouse 查询持续时间',
  labelNames: ['query_type', 'status'],
  buckets: [0.01, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10],
  registers: [registry],
});

const queryErrors = new Counter({
  name: 'clickhouse_query_errors_total',
  help: 'ClickHouse 查询错误',
  labelNames: ['error_code'],
  registers: [registry],
});

const insertRows = new Counter({
  name: 'clickhouse_insert_rows_total',
  help: '插入 ClickHouse 的总行数',
  labelNames: ['table'],
  registers: [registry],
});

// 插桩的查询包装器
async function instrumentedQuery<T>(
  queryType: string,
  fn: () => Promise<T>,
): Promise<T> {
  const timer = queryDuration.startTimer({ query_type: queryType });
  try {
    const result = await fn();
    timer({ status: 'success' });
    return result;
  } catch (err: any) {
    timer({ status: 'error' });
    queryErrors.inc({ error_code: err.code ?? 'unknown' });
    throw err;
  }
}

// 暴露 /metrics 端点
app.get('/metrics', async (req, res) => {
  res.set('Content-Type', registry.contentType);
  res.send(await registry.metrics());
});
```

### 步骤 4：Grafana 仪表板面板

```json
{
  "panels": [
    {
      "title": "查询速率 (QPS)",
      "type": "timeseries",
      "targets": [{ "expr": "rate(clickhouse_query_duration_seconds_count[5m])" }]
    },
    {
      "title": "查询延迟 P50 / P95 / P99",
      "type": "timeseries",
      "targets": [
        { "expr": "histogram_quantile(0.5, rate(clickhouse_query_duration_seconds_bucket[5m]))" },
        { "expr": "histogram_quantile(0.95, rate(clickhouse_query_duration_seconds_bucket[5m]))" },
        { "expr": "histogram_quantile(0.99, rate(clickhouse_query_duration_seconds_bucket[5m]))" }
      ]
    },
    {
      "title": "错误率",
      "type": "stat",
      "targets": [{
        "expr": "rate(clickhouse_query_errors_total[5m]) / rate(clickhouse_query_duration_seconds_count[5m])"
      }]
    },
    {
      "title": "插入吞吐量 (行/秒)",
      "type": "timeseries",
      "targets": [{ "expr": "rate(clickhouse_insert_rows_total[5m])" }]
    }
  ]
}
```

导入官方 ClickHouse Grafana 仪表板：`https://grafana.com/grafana/dashboards/23415`

### 步骤 5：告警规则

```yaml
# clickhouse-alerts.yml
groups:
  - name: clickhouse
    rules:
      - alert: ClickHouseHighErrorRate
        expr: |
          rate(clickhouse_query_errors_total[5m]) /
          rate(clickhouse_query_duration_seconds_count[5m]) > 0.05
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "ClickHouse 错误率 > 5%"

      - alert: ClickHouseHighLatency
        expr: |
          histogram_quantile(0.95,
            rate(clickhouse_query_duration_seconds_bucket[5m])) > 5
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "ClickHouse P95 延迟 > 5 秒"

      - alert: ClickHouseTooManyParts
        expr: clickhouse_table_parts > 300
        for: 10m
        labels:
          severity: critical
        annotations:
          summary: "表有 > 300 个活跃 parts — 合并落后"

      - alert: ClickHouseMemoryHigh
        expr: clickhouse_server_memory_usage / clickhouse_server_memory_limit > 0.9
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "ClickHouse 内存使用率 > 90%"

      - alert: ClickHouseDiskLow
        expr: clickhouse_disk_free_bytes / clickhouse_disk_total_bytes < 0.15
        for: 10m
        labels:
          severity: critical
        annotations:
          summary: "ClickHouse 磁盘空间 < 15% 可用"
```

### 步骤 6：结构化日志记录

```typescript
import pino from 'pino';

const logger = pino({ name: 'clickhouse' });

// 记录查询性能以进行分析
function logQuery(queryType: string, durationMs: number, rowsRead: number) {
  logger.info({
    service: 'clickhouse',
    query_type: queryType,
    duration_ms: durationMs,
    rows_read: rowsRead,
    status: durationMs > 5000 ? 'slow' : 'ok',
  });
}
```

## 监控的关键系统表

| 表 | 监控内容 | 频率 |
|-------|-----------------|-----------|
| `system.processes` | 运行中的查询、内存使用 | 每 10 秒 |
| `system.query_log` | 查询性能历史 | 每 1 分钟 |
| `system.parts` | Part 数量、合并健康 | 每 1 分钟 |
| `system.merges` | 活跃合并进度 | 每 30 秒 |
| `system.metrics` | 服务器范围的仪表盘（连接数、内存） | 每 10 秒 |
| `system.events` | 累积计数器 | 每 1 分钟 |
| `system.replicas` | 复制延迟 | 每 30 秒 |
| `system.disks` | 磁盘空间 | 每 5 分钟 |

## 错误处理

| 问题 | 原因 | 解决方案 |
|-------|-------|----------|
| 指标端点为空 | Prometheus 未配置 | 在配置中启用 `/metrics` |
| 高基数告警 | 标签值过多 | 降低标签基数 |
| 缺少 query_log 数据 | 日志记录已禁用 | 在配置中设置 `log_queries = 1` |
| 仪表板空白 | 抓取间隔过长 | 使用 10-15 秒抓取间隔 |

## 资源

- [Prometheus 集成](https://clickhouse.com/docs/integrations/prometheus)
- [ClickHouse Grafana 仪表板](https://grafana.com/grafana/dashboards/23415)
- [系统表参考](https://clickhouse.com/docs/operations/system-tables)
- [云监控](https://clickhouse.com/blog/clickhouse-cloud-now-supports-prometheus-monitoring)

## 下一步

有关事件响应，请参阅 `clickhouse-incident-runbook`。
