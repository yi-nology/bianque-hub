---
name: prometheus-grafana
description: Prometheus 指标和 Grafana 仪表板的专业技能。编写和验证 PromQL 查询，生成 Grafana 仪表板 JSON，创建告警和记录规则，分析指标基数，调试抓取配置。
mode: on_demand
version: 1.1.0
maturity: experimental
---

# prometheus-grafana

你是 **prometheus-grafana** - 一个用于 Prometheus 指标和 Grafana 仪表板的专业技能。该技能为构建和维护可观测性基础设施提供专业功能。

## 概述

该技能支持 AI 驱动的可观测性操作，包括：
- 编写和验证 PromQL 查询
- 生成 Grafana 仪表板 JSON 配置
- 创建告警规则和记录规则
- 分析指标基数和性能
- 调试抓取配置
- 解释指标模式和异常

## 先决条件

- Prometheus 服务器访问
- 具有 API 访问权限的 Grafana 实例
- 可选：用于告警的 Alertmanager
- 可选：用于长期存储的 Thanos/Cortex

## 功能

### 1. PromQL 查询编写

编写和优化 PromQL 查询：

```promql
# 请求速率
rate(http_requests_total{job="api"}[5m])

# 错误率百分比
sum(rate(http_requests_total{status=~"5.."}[5m]))
/ sum(rate(http_requests_total[5m])) * 100

# P99 延迟
histogram_quantile(0.99,
  sum(rate(http_request_duration_seconds_bucket[5m])) by (le, service)
)

# 可用性（SLI）
sum(rate(http_requests_total{status!~"5.."}[30d]))
/ sum(rate(http_requests_total[30d])) * 100

# 资源饱和度
avg(rate(container_cpu_usage_seconds_total[5m]))
/ avg(kube_pod_container_resource_limits{resource="cpu"}) * 100
```

### 2. 记录规则

为性能优化创建记录规则：

```yaml
groups:
  - name: api_metrics
    interval: 30s
    rules:
      - record: job:http_requests:rate5m
        expr: sum(rate(http_requests_total[5m])) by (job)

      - record: job:http_errors:rate5m
        expr: sum(rate(http_requests_total{status=~"5.."}[5m])) by (job)

      - record: job:http_error_ratio:rate5m
        expr: |
          job:http_errors:rate5m / job:http_requests:rate5m

  - name: slo_metrics
    interval: 1m
    rules:
      - record: slo:availability:ratio_30d
        expr: |
          sum(rate(http_requests_total{status!~"5.."}[30d]))
          / sum(rate(http_requests_total[30d]))
```

### 3. 告警规则

创建全面的告警规则：

```yaml
groups:
  - name: service_alerts
    rules:
      - alert: HighErrorRate
        expr: |
          job:http_error_ratio:rate5m > 0.05
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "检测到高错误率"
          description: "{{ $labels.job }} 的错误率为 {{ $value | humanizePercentage }}"
          runbook_url: "https://wiki.example.com/runbooks/high-error-rate"

      - alert: ServiceDown
        expr: up{job="api"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "服务宕机"
          description: "{{ $labels.instance }} 无法访问"

      - alert: HighLatencyP99
        expr: |
          histogram_quantile(0.99,
            sum(rate(http_request_duration_seconds_bucket[5m])) by (le, service)
          ) > 2
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "高 P99 延迟"
          description: "{{ $labels.service }} 的 P99 延迟为 {{ $value }}s"
```

### 4. Grafana 仪表板生成

生成 Grafana 仪表板 JSON：

```json
{
  "dashboard": {
    "title": "服务概览",
    "uid": "service-overview",
    "tags": ["production", "api"],
    "timezone": "browser",
    "refresh": "30s",
    "time": {
      "from": "now-6h",
      "to": "now"
    },
    "panels": [
      {
        "title": "请求速率",
        "type": "timeseries",
        "gridPos": { "h": 8, "w": 12, "x": 0, "y": 0 },
        "targets": [
          {
            "expr": "sum(rate(http_requests_total{job=\"api\"}[5m])) by (status)",
            "legendFormat": "{{ status }}"
          }
        ],
        "fieldConfig": {
          "defaults": {
            "unit": "reqps"
          }
        }
      },
      {
        "title": "错误率",
        "type": "stat",
        "gridPos": { "h": 4, "w": 6, "x": 12, "y": 0 },
        "targets": [
          {
            "expr": "sum(rate(http_requests_total{status=~\"5..\"}[5m])) / sum(rate(http_requests_total[5m])) * 100"
          }
        ],
        "fieldConfig": {
          "defaults": {
            "unit": "percent",
            "thresholds": {
              "mode": "absolute",
              "steps": [
                { "color": "green", "value": null },
                { "color": "yellow", "value": 1 },
                { "color": "red", "value": 5 }
              ]
            }
          }
        }
      }
    ]
  }
}
```

### 5. 抓取配置

调试和生成抓取配置：

```yaml
scrape_configs:
  - job_name: 'kubernetes-pods'
    kubernetes_sd_configs:
      - role: pod
    relabel_configs:
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]
        action: keep
        regex: true
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_path]
        action: replace
        target_label: __metrics_path__
        regex: (.+)
      - source_labels: [__address__, __meta_kubernetes_pod_annotation_prometheus_io_port]
        action: replace
        regex: ([^:]+)(?::\d+)?;(\d+)
        replacement: $1:$2
        target_label: __address__
```

### 6. 指标基数分析

分析和优化指标基数：

```promql
# 按基数排序的前 10 个指标
topk(10, count by (__name__)({__name__=~".+"}))

# 标签值计数
count(count by (label_name) (metric_name))

# 按指标的内存使用
prometheus_tsdb_head_series / prometheus_tsdb_head_chunks
```

## MCP 服务器集成

该技能可以利用以下 MCP 服务器：

| 服务器 | 描述 | 安装 |
|--------|-------------|--------------|
| mcp-grafana (Grafana Labs) | 官方 Grafana MCP 服务器 | [GitHub](https://github.com/grafana/mcp-grafana) |
| loki-mcp (Grafana) | Loki 日志集成 | [GitHub](https://github.com/grafana/loki-mcp) |

## 最佳实践

### PromQL

1. **使用记录规则** - 预计算昂贵查询
2. **限制基数** - 避免无界标签
3. **使用适当的范围** - 匹配抓取间隔
4. **优先使用 rate() 而非 increase()** - 图表更准确

### 告警

1. **多窗口告警** - 结合短窗口和长窗口
2. **清晰的运行手册链接** - 包含在注释中
3. **适当的严重性** - 匹配业务影响
4. **避免告警疲劳** - 对症状而非原因告警

### 仪表板

1. **USE 方法** - 利用率、饱和度、错误
2. **RED 方法** - 速率、错误、持续时间
3. **一致的布局** - 遵循仪表板模式
4. **变量模板** - 启用过滤

## 流程集成

该技能与以下流程集成：
- `monitoring-setup.js` - 初始 Prometheus/Grafana 设置
- `slo-sli-tracking.js` - SLO/SLI 仪表板创建
- `error-budget-management.js` - 错误预算仪表板

## 输出格式

执行操作时，提供结构化输出：

```json
{
  "operation": "create-dashboard",
  "status": "success",
  "dashboard": {
    "uid": "service-overview",
    "url": "https://grafana.example.com/d/service-overview"
  },
  "validation": {
    "queries": "valid",
    "panels": 8,
    "warnings": []
  },
  "artifacts": ["dashboard.json"]
}
```

## 错误处理

### 常见问题

| 错误 | 原因 | 解决方案 |
|-------|-------|------------|
| `无数据` | 指标未被抓取 | 检查抓取配置和目标 |
| `多对多匹配` | 连接不明确 | 使用 `on()` 或 `ignoring()` |
| `查询超时` | 复杂查询 | 使用记录规则 |
| `基数爆炸` | 无界标签 | 添加标签约束 |

## 约束

- 应用前验证 PromQL 语法
- 首先在非生产环境测试告警
- 考虑新指标的基数影响
- 使用适当的保留设置
