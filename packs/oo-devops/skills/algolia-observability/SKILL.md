---
name: algolia-observability
description: SaaS 搜索监控专家 - Algolia 可观测性、搜索性能分析、全文检索监控
mode: on_demand
version: 1.1.0
maturity: experimental
---

# Algolia 可观测性

## 概述

Algolia 在仪表板中提供内置分析功能，但生产系统需要应用级别的可观测性：延迟直方图、错误率计数器、分布式链路追踪和告警。本技能为 `algoliasearch` v5 客户端集成了 Prometheus、OpenTelemetry 和结构化日志记录。

## 关键监控指标

| 指标 | 类型 | 重要性 |
|--------|------|---------------|
| 搜索延迟 (P50/P95/P99) | 直方图 | 用户体验、SLA 合规性 |
| 搜索请求/秒 | 计数器 | 容量规划、成本跟踪 |
| 按类型划分的错误率 | 计数器 | 在用户报告之前检测 API 问题 |
| 索引新鲜度（最后更新时间） | 仪表盘 | 数据管道健康状态 |
| 记录数 | 仪表盘 | 成本监控、数据完整性 |

## 使用说明

### 步骤 1：插桩的 Algolia 客户端包装器

```typescript
// src/algolia/instrumented-client.ts
import { algoliasearch, ApiError } from 'algoliasearch';
import { Counter, Histogram, Gauge, Registry } from 'prom-client';

const registry = new Registry();

const searchLatency = new Histogram({
  name: 'algolia_search_duration_seconds',
  help: 'Algolia 搜索请求持续时间（秒）',
  labelNames: ['index', 'status'],
  buckets: [0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5],
  registers: [registry],
});

const searchTotal = new Counter({
  name: 'algolia_search_requests_total',
  help: 'Algolia 搜索请求总数',
  labelNames: ['index', 'status'],
  registers: [registry],
});

const searchErrors = new Counter({
  name: 'algolia_errors_total',
  help: '按类型划分的 Algolia 错误总数',
  labelNames: ['index', 'error_type', 'status_code'],
  registers: [registry],
});

const indexRecords = new Gauge({
  name: 'algolia_index_records',
  help: 'Algolia 索引中的记录数',
  labelNames: ['index'],
  registers: [registry],
});

const client = algoliasearch(process.env.ALGOLIA_APP_ID!, process.env.ALGOLIA_ADMIN_KEY!);

export async function instrumentedSearch<T = any>(
  indexName: string,
  searchParams: Record<string, any>
) {
  const timer = searchLatency.startTimer({ index: indexName });

  try {
    const result = await client.searchSingleIndex<T>({ indexName, searchParams });
    timer({ status: 'success' });
    searchTotal.inc({ index: indexName, status: 'success' });
    return result;
  } catch (error) {
    timer({ status: 'error' });
    searchTotal.inc({ index: indexName, status: 'error' });

    if (error instanceof ApiError) {
      searchErrors.inc({
        index: indexName,
        error_type: error.status === 429 ? 'rate_limit' : 'api_error',
        status_code: String(error.status),
      });
    } else {
      searchErrors.inc({
        index: indexName,
        error_type: 'network',
        status_code: '0',
      });
    }
    throw error;
  }
}

// 定期索引统计信息收集（每 5 分钟运行一次）
export async function collectIndexMetrics() {
  const { items } = await client.listIndices();
  for (const idx of items) {
    indexRecords.set({ index: idx.name }, idx.entries || 0);
  }
}

export { registry };
```

### 步骤 2：Prometheus 指标端点

```typescript
// src/api/metrics.ts (Express 示例)
import express from 'express';
import { registry, collectIndexMetrics } from '../algolia/instrumented-client';

const app = express();

app.get('/metrics', async (_req, res) => {
  res.set('Content-Type', registry.contentType);
  res.send(await registry.metrics());
});

// 每 5 分钟收集一次索引统计信息
setInterval(collectIndexMetrics, 5 * 60 * 1000);
```

### 步骤 3：OpenTelemetry 分布式链路追踪

```typescript
// src/algolia/tracing.ts
import { trace, SpanStatusCode, type Span } from '@opentelemetry/api';

const tracer = trace.getTracer('algolia-service', '1.0.0');

export async function tracedSearch<T>(
  indexName: string,
  query: string,
  searchParams: Record<string, any> = {}
): Promise<T> {
  return tracer.startActiveSpan(`algolia.search ${indexName}`, async (span: Span) => {
    span.setAttribute('algolia.index', indexName);
    span.setAttribute('algolia.query', query);
    span.setAttribute('algolia.hitsPerPage', searchParams.hitsPerPage || 20);

    try {
      const result = await client.searchSingleIndex<T>({
        indexName,
        searchParams: { query, ...searchParams },
      });

      span.setAttribute('algolia.nbHits', result.nbHits);
      span.setAttribute('algolia.processingTimeMS', result.processingTimeMS);
      span.setStatus({ code: SpanStatusCode.OK });
      return result as T;
    } catch (error: any) {
      span.setStatus({ code: SpanStatusCode.ERROR, message: error.message });
      span.recordException(error);
      throw error;
    } finally {
      span.end();
    }
  });
}
```

### 步骤 4：结构化日志记录

```typescript
// src/algolia/logger.ts
import pino from 'pino';

const logger = pino({ name: 'algolia', level: process.env.LOG_LEVEL || 'info' });

export function logSearch(params: {
  index: string;
  query: string;
  nbHits: number;
  processingTimeMS: number;
  page: number;
  userId?: string;
}) {
  logger.info({
    event: 'algolia.search',
    index: params.index,
    query: params.query,
    hits: params.nbHits,
    latency_ms: params.processingTimeMS,
    page: params.page,
    user: params.userId,
  });
}

export function logSearchError(params: {
  index: string;
  query: string;
  error: string;
  statusCode?: number;
}) {
  logger.error({
    event: 'algolia.search.error',
    index: params.index,
    query: params.query,
    error: params.error,
    status_code: params.statusCode,
  });
}
```

### 步骤 5：告警规则 (Prometheus AlertManager)

```yaml
# alerts/algolia.yml
groups:
  - name: algolia
    rules:
      - alert: AlgoliaHighErrorRate
        expr: |
          rate(algolia_errors_total[5m]) /
          rate(algolia_search_requests_total[5m]) > 0.05
        for: 5m
        labels: { severity: warning }
        annotations:
          summary: "Algolia 错误率 > 5%，持续 5 分钟"

      - alert: AlgoliaHighLatency
        expr: |
          histogram_quantile(0.95,
            rate(algolia_search_duration_seconds_bucket[5m])
          ) > 0.5
        for: 5m
        labels: { severity: warning }
        annotations:
          summary: "Algolia P95 搜索延迟 > 500ms"

      - alert: AlgoliaRateLimited
        expr: rate(algolia_errors_total{error_type="rate_limit"}[5m]) > 0
        for: 2m
        labels: { severity: critical }
        annotations:
          summary: "Algolia 返回 429 速率限制错误"

      - alert: AlgoliaIndexStale
        expr: algolia_index_records == 0
        for: 10m
        labels: { severity: warning }
        annotations:
          summary: "Algolia 索引有 0 条记录 — 可能同步失败"
```

## Grafana 仪表板查询

```
# 搜索速率: rate(algolia_search_requests_total[5m])
# 错误率: rate(algolia_errors_total[5m]) / rate(algolia_search_requests_total[5m])
# P50 延迟: histogram_quantile(0.5, rate(algolia_search_duration_seconds_bucket[5m]))
# P95 延迟: histogram_quantile(0.95, rate(algolia_search_duration_seconds_bucket[5m]))
# 每个索引的记录数: algolia_index_records
```

## 错误处理

| 问题 | 原因 | 解决方案 |
|-------|-------|----------|
| 指标缺失 | 客户端未插桩 | 使用 `instrumentedSearch` 包装器 |
| 高基数 | 标签值过多 | 不要将查询文本用作标签 |
| 链路追踪断层 | 缺少上下文传播 | 确保 OTel 上下文流经异步操作 |
| 告警风暴 | 阈值过于敏感 | 添加最少 `for: 5m` 持续时间 |

## 资源

- [Prometheus 客户端](https://www.npmjs.com/package/prom-client)
- [OpenTelemetry JS](https://opentelemetry.io/docs/languages/js/)
- [Algolia 仪表板分析](https://www.algolia.com/doc/guides/getting-analytics/search-analytics/)
- [pino 日志记录器](https://getpino.io/)

## 下一步

有关事件响应，请参阅 `algolia-incident-runbook`。
