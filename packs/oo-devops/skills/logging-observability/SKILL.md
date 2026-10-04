---
name: logging-observability
description: 结构化日志(structured logging)、分布式追踪(distributed tracing)和指标(metrics)收集模式，用于构建可观测(observability)系统。涵盖 JSON 日志、OpenTelemetry、RED/USE 方法、告警(alert)和仪表盘(dashboard)。触发词：可观测性(observability)、结构化日志(structured logging)、分布式追踪(distributed tracing)、指标(metrics)、OpenTelemetry
mode: on_demand
version: 1.1.0
maturity: experimental
---

# Logging & Observability

在三个支柱（logs、metrics、traces）上构建可观测系统的模式。

## Three Pillars

| Pillar | Purpose | Question It Answers | Example |
|--------|---------|---------------------|---------|
| **Logs** | What happened | Why did this request fail? | `{"level":"error","msg":"payment declined","user_id":"u_82"}` |
| **Metrics** | How much / how fast | Is latency increasing? | `http_request_duration_seconds{route="/api/orders"} 0.342` |
| **Traces** | Request flow | Where is the bottleneck? | Span: `api-gateway → auth → order-service → db` |

每个支柱在关联时最强。在每个日志行中嵌入 `trace_id`，以便从日志条目跳转到完整的分布式 trace。

---

## Structured Logging

始终将日志输出为结构化 JSON —— 永远不要使用自由文本字符串。

### Required Fields

| Field | Purpose | Required |
|-------|---------|----------|
| `timestamp` | ISO-8601 with milliseconds | Yes |
| `level` | Severity (DEBUG … FATAL) | Yes |
| `service` | Originating service name | Yes |
| `message` | Human-readable description | Yes |
| `trace_id` | Distributed trace correlation | Yes |
| `span_id` | Current span within trace | Yes |
| `correlation_id` | Business-level correlation (order ID) | When applicable |
| `error` | Structured error object | On errors |
| `context` | Request-specific metadata | Recommended |

### Context Enrichment

在 middleware 级别附加 context，以便下游日志自动继承：

```typescript
app.use((req, res, next) => {
  const ctx = {
    trace_id: req.headers['x-trace-id'] || crypto.randomUUID(),
    request_id: crypto.randomUUID(),
    user_id: req.user?.id,
    method: req.method,
    path: req.path,
  };
  asyncLocalStorage.run(ctx, () => next());
});
```

### Library Recommendations

| Library | Language | Strengths | Perf |
|---------|----------|-----------|------|
| **Pino** | Node.js | Fastest Node logger, low overhead | Excellent |
| **structlog** | Python | Composable processors, context binding | Good |
| **zerolog** | Go | Zero-allocation JSON logging | Excellent |
| **zap** | Go | High performance, typed fields | Excellent |
| **tracing** | Rust | Spans + events, async-aware | Excellent |

选择一个原生输出结构化 JSON 的 logger。避免需要后处理的 logger。

---

## Log Levels

| Level | When to Use | Example |
|-------|-------------|---------|
| **FATAL** | App cannot continue, process will exit | Database connection pool exhausted |
| **ERROR** | Operation failed, needs attention | Payment charge failed: CARD_DECLINED |
| **WARN** | Unexpected but recoverable | Retry 2/3 for upstream timeout |
| **INFO** | Normal business events | Order ORD-1234 placed successfully |
| **DEBUG** | Developer troubleshooting | Cache miss for key user:82:preferences |
| **TRACE** | Very fine-grained (rarely in prod) | Entering validateAddress with payload |

**Rules:** Production default = INFO and above. 如果你记录一个 ERROR，应该有人对其采取行动。每个 FATAL 都应该触发一个 alert。

---

## Distributed Tracing

### OpenTelemetry Setup

始终优先选择 OpenTelemetry 而非 vendor-specific SDKs：

```typescript
import { NodeSDK } from '@opentelemetry/sdk-node';
import { OTLPTraceExporter } from '@opentelemetry/exporter-trace-otlp-http';
import { getNodeAutoInstrumentations } from '@opentelemetry/auto-instrumentations-node';

const sdk = new NodeSDK({
  serviceName: 'order-service',
  traceExporter: new OTLPTraceExporter({
    url: 'http://otel-collector:4318/v1/traces',
  }),
  instrumentations: [getNodeAutoInstrumentations()],
});
sdk.start();
```

### Span Creation

```typescript
const tracer = trace.getTracer('order-service');

async function processOrder(order: Order) {
  return tracer.startActiveSpan('processOrder', async (span) => {
    try {
      span.setAttribute('order.id', order.id);
      span.setAttribute('order.total_cents', order.totalCents);
      await validateInventory(order);
      await chargePayment(order);
      span.setStatus({ code: SpanStatusCode.OK });
    } catch (err) {
      span.setStatus({ code: SpanStatusCode.ERROR, message: err.message });
      span.recordException(err);
      throw err;
    } finally {
      span.end();
    }
  });
}
```

### Context Propagation

- 使用 W3C Trace Context (`traceparent` header) —— OTel 中的默认值
- 跨 HTTP、gRPC 和 message queues 传播
- 对于 async workers：将 `traceparent` 序列化到 job payload 中

### Trace Sampling

| Strategy | Use When |
|----------|----------|
| **Always On** | Low-traffic services, debugging |
| **Probabilistic** (N%) | General production use |
| **Rate-limited** (N/sec) | High-throughput services |
| **Tail-based** | When you need all error traces |

无论 strategy 如何，始终对 100% 的 error traces 进行 sampling。

---

## Metrics Collection

### RED Method (Request-Driven)

为每个服务端点监控以下三个指标：

| Metric | What It Measures | Prometheus Example |
|--------|-----------------|-------------------|
| **Rate** | Requests/sec | `rate(http_requests_total[5m])` |
| **Errors** | Failed request ratio | `rate(http_requests_total{status=~"5.."}[5m])` |
| **Duration** | Response time | `histogram_quantile(0.99, http_request_duration_seconds)` |

### USE Method (Resource-Driven)

对于基础设施组件（CPU、memory、disk、network）：

| Metric | What It Measures | Example |
|--------|-----------------|---------|
| **Utilization** | % resource busy | CPU usage at 78% |
| **Saturation** | Work queued/waiting | 12 requests queued in thread pool |
| **Errors** | Error events on resource | 3 disk I/O errors in last minute |

---

## Monitoring Stack

| Tool | Category | Best For |
|------|----------|----------|
| **Prometheus** | Metrics | Pull-based metrics, alerting rules |
| **Grafana** | Visualisation | Dashboards for metrics, logs, traces |
| **Jaeger** | Tracing | Distributed trace visualisation |
| **Loki** | Logs | Log aggregation (pairs with Grafana) |
| **OpenTelemetry** | Collection | Vendor-neutral telemetry collection |

**Recommendation:** 从 OTel Collector → Prometheus + Grafana + Loki + Jaeger 开始。仅在运维开销证明成本合理时迁移到 SaaS。

---

## Alert Design

### Severity Levels

| Severity | Response Time | Example |
|----------|---------------|---------|
| **P1** | Immediate | Service fully down, data loss |
| **P2** | < 30 min | Error rate > 5%, latency p99 > 5s |
| **P3** | Business hours | Disk > 80%, cert expiring in 7 days |
| **P4** | Best effort | Non-critical deprecation warning |

### Alert Fatigue Prevention

- **Alert on symptoms, not causes** —— "error rate > 5%" 而不是 "pod restarted"
- **Multi-window, multi-burn-rate** —— 既能捕捉突然飙升，也能捕捉缓慢燃烧
- **Require runbook links** —— 每个 alert 必须链接到 diagnosis 和 remediation
- **Review monthly** —— 删除或调整从未触发或总是触发的 alerts
- **Group related alerts** —— 使用 inhibition rules 来抑制 child alerts
- **Set appropriate thresholds** —— 如果 alert 每天触发且被忽略，提高 threshold 或删除

---

## Dashboard Patterns

### Overview Dashboard ("War Room")
- Total requests/sec across all services
- Global error rate (%) with trendline
- p50 / p95 / p99 latency
- Active alerts count by severity
- Deployment markers overlaid on graphs

### Service Dashboard (Per-Service)
- 每个端点的 RED metrics
- Dependency health（upstream/downstream success rates）
- Resource utilisation（CPU、memory、connections）
- Top errors table with count and last seen

---

## Observability Checklist

每个服务必须拥有：

- [ ] 结构化 JSON 日志，具有一致的 schema
- [ ] 在所有请求上传播 correlation / trace IDs
- [ ] 为每个外部端点暴露 RED metrics
- [ ] Health check 端点（`/healthz` 和 `/readyz`）
- [ ] 使用 OpenTelemetry 进行分布式追踪
- [ ] 用于 RED metrics 和 resource utilisation 的 dashboards
- [ ] 用于 error rate、latency 和 saturation 的 alerts，并带有 runbook links
- [ ] 无需重新部署即可在运行时配置 log level
- [ ] PII scrubbing 已验证并测试
- [ ] 为 logs、metrics 和 traces 定义 retention policies

## Anti-Patterns

| Anti-Pattern | Problem | Fix |
|-------------|---------|-----|
| Logging PII | Privacy/compliance violation | Mask or exclude PII; use token references |
| Excessive logging | Storage costs balloon, signal drowns | Log business events, not data flow |
| Unstructured logs | Cannot query or alert on fields | Use structured JSON with consistent schema |
| String interpolation | Breaks structured fields, injection risk | Pass fields as metadata, not in message |
| Missing correlation IDs | Cannot trace across services | Generate and propagate trace_id everywhere |
| Alert storms | On-call fatigue, real issues buried | Use grouping, inhibition, deduplication |
| Metrics with high cardinality | Prometheus OOM, dashboard timeouts | Never use user ID or request ID as label |

## NEVER Do

1. **NEVER log passwords, tokens, API keys, or secrets** —— 即使在 DEBUG 级别
2. **NEVER use console.log / print in production** —— 使用结构化 logger
3. **NEVER use user IDs, emails, or request IDs as metric labels** —— cardinality 会爆炸
4. **NEVER create alerts without a runbook link** —— 无法操作的 alerts 会侵蚀信任
5. **NEVER rely on logs alone** —— 你需要 metrics 和 traces 来实现完整的可观测性
6. **NEVER log request/response bodies by default** —— 仅 opt-in，并带有 PII redaction
7. **NEVER ignore log volume** —— 设置 budgets，并在服务超过每日 quota 时发出 alert
8. **NEVER skip context propagation in async flows** —— 断裂的 traces 比没有 traces 更糟
