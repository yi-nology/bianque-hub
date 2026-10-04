---
name: sentry-observability
description: 将 Sentry 与您的可观测性栈集成——日志、指标、APM 和仪表板。 用于将 Sentry 连接到 winston/pino/structlog、将错误与业务 指标关联、在 Sentry 性能监控与 Datadog/New Relic 之间做选择、构建 Sentry Discover 仪表板，或通过额外上下文将事件链接到外部工具。 触发词："sentry observability"、"sentry logging"、"sentry metrics"、"sentry grafana"、 "sentry datadog correlation"、"sentry discover dashboard"。
mode: on_demand
version: 1.1.0
maturity: experimental
---

# Sentry 可观测性集成

## 概述

将 Sentry 接入您的日志、指标、APM 和仪表板工具链，使每个错误都携带完整上下文，每个指标都能关联到根因事件。本技能涵盖三个集成层：结构化日志（winston、pino、structlog）与 Sentry 事件 ID 关联、带有错误率跟踪的业务指标，以及通过 Sentry Discover、Grafana Webhook 和 APM 工具的跨工具链接。

另请参阅：[日志集成详情](references/logging-integration.md) | [指标模式](references/metrics-integration.md) | [APM 工具交叉链接](references/apm-tool-integration.md)

## 前置条件

- 已安装 Sentry SDK v8+（Node.js 使用 `@sentry/node`，Python 使用 `sentry-sdk`）
- 至少配置了一个结构化日志记录器（winston、pino 或 structlog）
- 环境中可用的 Sentry 项目 DSN（`SENTRY_DSN`）
- 可访问的仪表板平台（Sentry Discover、Grafana 或 Datadog）
- 已确定的告警路由策略（谁接收分页、警告发送到哪里）

## 说明

### 步骤 1 —— 将 Sentry 事件 ID 附加到结构化日志

核心模式：每个触发 Sentry 事件的日志行都携带事件 ID，每个 Sentry 事件都携带日志上下文。这在您的日志聚合器和 Sentry 之间创建了双向链接。

**Winston（Node.js）—— 自定义传输：**

```typescript
import winston from 'winston';
import * as Sentry from '@sentry/node';

class SentryTransport extends winston.Transport {
  log(info: any, callback: () => void) {
    setImmediate(callback);

    if (info.level === 'error' || info.level === 'fatal') {
      const error = info.error instanceof Error
        ? info.error
        : new Error(info.message);

      Sentry.withScope((scope) => {
        scope.setTag('logger', 'winston');
        scope.setContext('log_entry', {
          level: info.level,
          timestamp: info.timestamp,
          service: info.service,
        });
        const eventId = Sentry.captureException(error);
        info.sentry_event_id = eventId;
        info.sentry_url = `https://${process.env.SENTRY_ORG}.sentry.io/issues/?query=${eventId}`;
      });
    }
  }
}

const logger = winston.createLogger({
  defaultMeta: { service: 'api-gateway' },
  transports: [
    new winston.transports.Console({ format: winston.format.json() }),
    new SentryTransport(),
  ],
});
```

**Pino（Node.js）—— 钩子模式：**

```typescript
import pino from 'pino';
import * as Sentry from '@sentry/node';

const logger = pino({
  hooks: {
    logMethod(inputArgs, method, level) {
      if (level >= 50) { // 50 = error, 60 = fatal
        const [obj, msg] = typeof inputArgs[0] === 'object'
          ? [inputArgs[0], inputArgs[1]]
          : [{}, inputArgs[0]];

        Sentry.withScope((scope) => {
          scope.setTag('logger', 'pino');
          const eventId = Sentry.captureException(
            obj.err instanceof Error ? obj.err : new Error(String(msg))
          );
          if (typeof inputArgs[0] === 'object') {
            inputArgs[0].sentry_event_id = eventId;
          }
        });
      }
      return method.apply(this, inputArgs);
    },
  },
});
```

Python structlog 集成，请参阅 [logging-integration.md](references/logging-integration.md)。

**请求 ID 关联中间件：**

```typescript
import { randomUUID } from 'crypto';
import * as Sentry from '@sentry/node';

app.use((req, res, next) => {
  const requestId = (req.headers['x-request-id'] as string) || randomUUID();
  req.requestId = requestId;
  res.setHeader('x-request-id', requestId);
  Sentry.setTag('request_id', requestId);
  req.log = logger.child({ requestId, path: req.path });
  next();
});
```

### 步骤 2 —— 将错误与业务指标和 APM 关联

将 Sentry 事件连接到您的指标管道，并决定何时仅使用 Sentry 性能监控就足够，何时需要添加 Datadog 或 New Relic。

**Sentry 自定义指标（内置，无需额外工具）：**

```typescript
import * as Sentry from '@sentry/node';

// 计数器 —— 跟踪错误率以及业务事件
Sentry.metrics.increment('checkout.attempted', 1, {
  tags: { payment_provider: 'stripe', plan: 'enterprise' },
});

Sentry.metrics.increment('checkout.failed', 1, {
  tags: { payment_provider: 'stripe', failure_reason: 'timeout' },
});

// 分布 —— 跟踪延迟并关联错误
Sentry.metrics.distribution('api.response_time', responseTimeMs, {
  tags: { endpoint: '/api/orders', status_code: String(res.statusCode) },
  unit: 'millisecond',
});

// 计量器 —— 跟踪队列深度、连接池大小
Sentry.metrics.gauge('db.pool.active', pool.activeCount, {
  tags: { database: 'primary' },
});

// 集合 —— 跟踪事件期间受影响的用户
Sentry.metrics.set('incident.affected_users', userId, {
  tags: { incident: 'payment-outage-2026-03' },
});
```

Prometheus 双写模式，请参阅 [metrics-integration.md](references/metrics-integration.md)。

**何时使用 Sentry 性能监控 vs Datadog/New Relic：**

| 场景 | 使用 Sentry Performance | 使用 Datadog/New Relic |
|----------|----------------------|----------------------|
| 前后端统一视图 | 是 —— 统一的错误 + 性能追踪 | 如果 Sentry 已覆盖您的技术栈则过于复杂 |
| 基础设施指标（CPU、内存） | 否 —— Sentry 不收集基础设施指标 | 是 —— 原生主机代理收集 |
| 100+ 自定义指标系列 | 查询约束有限 | 是 —— 专为高基数构建 |
| 预算受限，< 5 个服务 | 是 —— 一个工具，一份账单 | 不必要的成本 |

**通过 `beforeSend` 进行 Datadog + Sentry 交叉链接：**

```typescript
import tracer from 'dd-trace';
import * as Sentry from '@sentry/node';

// dd-trace 必须在 @sentry/node 之前初始化
Sentry.init({
  dsn: process.env.SENTRY_DSN,
  beforeSend(event) {
    const span = tracer.scope().active();
    if (span) {
      const traceId = span.context().toTraceId();
      event.tags = { ...event.tags, 'dd.trace_id': traceId };
      event.contexts = {
        ...event.contexts,
        datadog: {
          trace_url: `https://app.datadoghq.com/apm/trace/${traceId}`,
          trace_id: traceId,
        },
      };
    }
    return event;
  },
});
```

New Relic 关联模式，请参阅 [apm-tool-integration.md](references/apm-tool-integration.md)。

### 步骤 3 —— 构建仪表板并连接外部工具

使用 Sentry Discover 进行错误分析，设置 Grafana Webhook 以统一仪表板，并通过 `setContext` 将 Sentry 事件链接到外部工具。

**通过 `Sentry.setContext('monitoring', ...)` 链接所有工具：**

```typescript
import * as Sentry from '@sentry/node';

function setMonitoringContext(req: Request) {
  const traceId = Sentry.getActiveSpan()?.spanContext().traceId;
  const spanId = Sentry.getActiveSpan()?.spanContext().spanId;
  const requestId = req.headers['x-request-id'] as string || crypto.randomUUID();

  // setContext 在 Sentry 事件侧边栏中创建一个命名部分
  Sentry.setContext('monitoring', {
    traceId,
    spanId,
    requestId,
    grafana_dashboard: `https://grafana.example.com/d/abc123?var-trace_id=${traceId}`,
    kibana_logs: `https://kibana.example.com/app/logs?query=request_id:${requestId}`,
    datadog_trace: traceId
      ? `https://app.datadoghq.com/apm/trace/${traceId}`
      : undefined,
  });

  Sentry.setTag('request_id', requestId);
  Sentry.setTag('trace_id', traceId || 'none');
  Sentry.setTag('deployment', process.env.DEPLOYMENT_ID || 'unknown');
}

app.use((req, res, next) => {
  setMonitoringContext(req);
  next();
});
```

**通过 Sentry Webhook 进行 Grafana 集成：**

在 设置 > 集成 > 内部集成 中配置。将 Webhook URL 指向将 Sentry 事件转换为 Grafana 注释的接收器：

```typescript
// 接收 Sentry Webhook，创建 Grafana 注释
app.post('/sentry-to-grafana', async (req, res) => {
  const { event } = req.body;
  if (!event) return res.status(200).send('ignored');

  await fetch(`${process.env.GRAFANA_URL}/api/annotations`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${process.env.GRAFANA_API_KEY}`,
    },
    body: JSON.stringify({
      dashboardUID: process.env.GRAFANA_DASHBOARD_UID,
      panelId: 1,
      time: new Date(event.datetime).getTime(),
      tags: ['sentry', event.level, event.project],
      text: `**${event.title}**\nLevel: ${event.level}\n[View in Sentry](${event.web_url})`,
    }),
  });

  res.status(201).json({ status: 'annotation_created' });
});
```

**跨工具告警路由：**

```
问题告警："关键生产错误"
  当：首次看到事件
  如果：级别为 fatal 且环境为 production
  则：PagerDuty（严重）+ #alerts-critical Slack + Grafana 注释
  频率：每个问题一次

指标告警："错误率激增"
  当：5 分钟内错误数 > 50
  则：PagerDuty（高）+ #alerts-production Slack + Webhook 到 Grafana
  解决：10 分钟内错误数 < 5

指标告警："延迟回归"
  当：/api/* 的 p95(transaction.duration) > 2000ms 持续 10 分钟
  则：#alerts-performance Slack + 通过 Webhook 创建 JIRA 工单
  解决：p95 < 1000ms 持续 15 分钟
```

## 输出

完成这些步骤后，您将拥有：

- Winston/pino/structlog 将错误转发到 Sentry，并在日志行中标记事件 ID
- Sentry 自定义指标（计数器、计量器、分布、集合）跟踪业务 KPI
- `beforeSend` 钩子将 Sentry 事件链接到 Datadog 追踪和 New Relic 事务
- `Sentry.setContext('monitoring', { traceId, spanId })` 将每个事件链接到外部工具 URL
- 从 Sentry Webhook 在基础设施仪表板上创建的 Grafana 注释
- 分层告警路由：致命错误呼叫值班人员，警告发送到 Slack，延迟问题创建工单

## 错误处理

| 错误 | 原因 | 解决方案 |
|-------|-------|----------|
| 日志中缺少 Sentry 事件 ID | 传输/处理器未连接 | 验证 `SentryTransport` 是否在 Winston 传输中或 `sentry_processor` 是否在 structlog 链中 |
| `beforeSend` 静默丢弃事件 | 处理器抛出或返回 `undefined` | 将 `beforeSend` 包装在 try/catch 中，错误路径始终返回 `event` |
| Grafana 注释未出现 | Webhook URL 错误或 API 密钥过期 | 首先用 `curl -X POST` 测试 Webhook；检查 Grafana API 密钥是否有注释写入权限 |
| Datadog 追踪 ID 不匹配 | `dd-trace` 未在 Sentry 之前初始化 | 在 `@sentry/node` 之前导入并初始化 `dd-trace` |
| Sentry 指标不可见 | 计划未启用该功能 | 自定义指标需要 Business 计划或更高版本 |
| 日志中的重复事件 | SDK 自动捕获和传输都触发 | 使用 `beforeSend` 去重，或禁用已处理错误的自动捕获 |
| Webhook 负载格式更改 | Sentry API 版本升级 | 将 Webhook 固定到 API v0；在接收器中验证负载形状 |

## 示例

**示例 1 —— 带有日志关联的全栈请求追踪**

请求："我们 Express API 中的每个错误都应该同时出现在我们的日志聚合器和 Sentry 中，并带有交叉链接。"

```typescript
import express from 'express';
import pino from 'pino';
import * as Sentry from '@sentry/node';
import { randomUUID } from 'crypto';

Sentry.init({ dsn: process.env.SENTRY_DSN, tracesSampleRate: 0.1 });
const logger = pino({ level: 'info' });
const app = express();

app.use((req, res, next) => {
  const requestId = (req.headers['x-request-id'] as string) || randomUUID();
  req.requestId = requestId;
  res.setHeader('x-request-id', requestId);
  Sentry.setTag('request_id', requestId);
  req.log = logger.child({ requestId, path: req.path, method: req.method });
  next();
});

app.get('/api/orders/:id', async (req, res) => {
  try {
    const order = await db.orders.findById(req.params.id);
    if (!order) {
      req.log.warn({ orderId: req.params.id }, 'Order not found');
      return res.status(404).json({ error: 'Not found' });
    }
    res.json(order);
  } catch (error) {
    Sentry.withScope((scope) => {
      scope.setContext('request', { params: req.params, requestId: req.requestId });
      const eventId = Sentry.captureException(error);
      req.log.error({
        err: error,
        sentry_event_id: eventId,
        sentry_url: `https://sentry.io/issues/?query=${eventId}`,
      }, 'Order fetch failed');
    });
    res.status(500).json({ error: 'Internal error', request_id: req.requestId });
  }
});

Sentry.setupExpressErrorHandler(app);
```

**示例 2 —— 带有 Sentry 错误注释的 Grafana 仪表板**

请求："在我们的 Grafana 延迟仪表板上显示 Sentry 错误作为注释。"

结果：创建一个 Sentry 内部集成（设置 > 集成 > 内部集成），其 Webhook URL 指向您的注释接收器。接收器将 Sentry 事件负载转换为 Grafana 注释 API 调用。错误事件以垂直标记形式出现在延迟图上，带有可点击的链接返回 Sentry 问题。

请参阅 [examples.md](references/examples.md) 获取更多集成模式。

## 资源

- [Sentry + OpenTelemetry](https://docs.sentry.io/platforms/javascript/guides/node/tracing/instrumentation/opentelemetry/) —— SDK v8 OTel 桥接
- [自定义指标](https://docs.sentry.io/product/explore/metrics/) —— 计数器、计量器、分布、集合
- [Sentry Discover 查询](https://docs.sentry.io/product/explore/discover-queries/) —— 自定义事件分析
- [Webhook 集成](https://docs.sentry.io/organization/integrations/integration-platform/webhooks/) —— 出站事件 Webhook
- [PagerDuty 集成](https://docs.sentry.io/organization/integrations/notification-incidents/pagerduty/) —— 事件路由
- [Slack 集成](https://docs.sentry.io/organization/integrations/notification-incidents/slack/) —— 告警频道

## 后续步骤

- 配置 sentry-performance-tracing 以进行事务级插桩
- 设置 sentry-release-management 以将部署与错误回归关联
- 添加 sentry-ci-integration 以在问题到达生产环境之前捕获回归
- 查看 sentry-cost-tuning 以在可观测性覆盖范围扩大时管理事件量
