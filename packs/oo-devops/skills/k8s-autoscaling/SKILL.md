---
name: k8s-autoscaling
description: Kubernetes 自动扩缩容专家 - HPA/VPA 弹性伸缩、KEDA 事件驱动伸缩、容量管理
mode: on_demand
version: 1.1.0
maturity: experimental
---

# Kubernetes Autoscaling

使用 HPA, VPA, 和 KEDA 配合 kubectl-mcp-server 工具进行全面的 autoscaling。

## 何时应用

在以下情况下使用此技能：
- 用户提到："HPA", "VPA", "KEDA", "autoscale", "scale to zero"
- 操作：配置 autoscaling, 检查 scaling 状态
- 关键词："scale automatically", "event-driven", "right-size"

## 优先级规则

| Priority | 规则 | Impact | Tools |
|----------|------|--------|-------|
| 1 | 首先验证 metrics-server for HPA | Critical | `get_resource_metrics` |
| 2 | 在 HPA 之前设置资源 requests | Critical | `describe_pod` |
| 3 | 使用 KEDA 进行 scale-to-zero | High | `keda_scaledobjects_list_tool` |
| 4 | 检查 VPA recommendations | Medium | `get_resource_recommendations` |

## 快速参考

| Task | Tool | Example |
|------|------|---------|
| 列表 KEDA ScaledObjects | `keda_scaledobjects_list_tool` | `keda_scaledobjects_list_tool(namespace)` |
| 获取 ScaledObject | `keda_scaledobject_get_tool` | `keda_scaledobject_get_tool(name, namespace)` |
| 列表 ScaledJobs | `keda_scaledjobs_list_tool` | `keda_scaledjobs_list_tool(namespace)` |
| 检查 KEDA | `keda_detect_tool` | `keda_detect_tool()` |

## HPA (Horizontal Pod Autoscaler)

基于 CPU 的基本 scaling：

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: my-app-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: my-app
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
```

应用和验证：

```python
kubectl_apply(hpa_yaml, namespace)
get_hpa(namespace)
```

## VPA (Vertical Pod Autoscaler)

Right-size 资源 requests：

```yaml
apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: my-app-vpa
spec:
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: my-app
  updatePolicy:
    updateMode: "Auto"
```

## KEDA (Event-Driven Autoscaling)

### 检测 KEDA 安装

```python
keda_detect_tool()
```

### 列表 ScaledObjects

```python
keda_scaledobjects_list_tool(namespace)
keda_scaledobject_get_tool(name, namespace)
```

### 列表 ScaledJobs

```python
keda_scaledjobs_list_tool(namespace)
```

### Trigger 认证

```python
keda_triggerauths_list_tool(namespace)
keda_triggerauth_get_tool(name, namespace)
```

### KEDA-Managed HPAs

```python
keda_hpa_list_tool(namespace)
```

参见 [KEDA-TRIGGERS.md](KEDA-TRIGGERS.md) 了解 trigger 配置。

## Common KEDA Triggers

### Queue-Based Scaling (AWS SQS)

```yaml
apiVersion: keda.sh/v1alpha1
kind: ScaledObject
metadata:
  name: sqs-scaler
spec:
  scaleTargetRef:
    name: queue-processor
  minReplicaCount: 0
  maxReplicaCount: 100
  triggers:
  - type: aws-sqs-queue
    metadata:
      queueURL: https://sqs.region.amazonaws.com/...
      queueLength: "5"
```

### Cron-Based Scaling

```yaml
triggers:
- type: cron
  metadata:
    timezone: America/New_York
    start: 0 8 * * 1-5
    end: 0 18 * * 1-5
    desiredReplicas: "10"
```

### Prometheus Metrics

```yaml
triggers:
- type: prometheus
  metadata:
    serverAddress: http://prometheus:9090
    metricName: http_requests_total
    query: sum(rate(http_requests_total{app="myapp"}[2m]))
    threshold: "100"
```

## Scaling Strategies

| Strategy | Tool | Use Case |
|----------|------|----------|
| CPU/Memory | HPA | 稳定的流量模式 |
| 自定义 metrics | HPA v2 | 业务指标 |
| Event-driven | KEDA | 队列处理, cron |
| Vertical | VPA | Right-size requests |
| Scale to zero | KEDA | 成本节省, 空闲工作负载 |

## Cost-Optimized Autoscaling

### Scale to Zero with KEDA

为空闲工作负载降低成本：

```python
keda_scaledobjects_list_tool(namespace)
```

### Right-Size with VPA

获取 recommendations 并应用：

```python
get_resource_recommendations(namespace)
```

## 故障排除

### HPA Not Scaling

```python
get_hpa(namespace)
get_pod_metrics(name, namespace)
describe_pod(name, namespace)
```

### KEDA Not Triggering

```python
keda_scaledobject_get_tool(name, namespace)
get_events(namespace)
```

### Common Issues

| Symptom | 检查 | Resolution |
|---------|-------|------------|
| HPA unknown | Metrics server | 安装 metrics-server |
| KEDA no scaling | Trigger auth | 检查 TriggerAuthentication |
| VPA not updating | Update mode | 设置 updateMode: Auto |
| Scale down slow | Stabilization | 调整 stabilizationWindowSeconds |

## 最佳实践

1. **Always 设置资源 Requests** - HPA 需要 requests 来计算利用率
2. **Use Multiple Metrics** - 组合 CPU + 自定义 metrics 以提高准确性
3. **Stabilization Windows** - 使用 scaleDown stabilization 防止波动
4. **Scale to Zero Carefully** - 考虑冷启动时间

## 相关 Skills

- [k8s-cost](../k8s-cost/SKILL.md) - 成本优化
- [k8s-troubleshoot](../k8s-troubleshoot/SKILL.md) - 调试 scaling issues
