---
name: k8s-debug
description: Kubernetes 调试模式。用于 Pod crashes, CrashLoopBackOff, OOMKilled, ImagePullBackOff, scheduling failures, Deployment issues。
mode: on_demand
version: 1.1.0
maturity: experimental
---

# Kubernetes Debugging Expertise

## 黄金规则：Events 在 Logs 之前

当调试 Kubernetes issues 时，**ALWAYS 首先检查 events**：

1. `get_pod_events` - 显示 scheduling, pulling, starting, probes, OOM
2. 然后 `get_pod_logs` - 应用-level errors

Events 比 logs 更快地解释大多数 crash/scheduling issues。

## 典型调查流程

```
1. list_pods        → 获取 namespace 中 Pod 健康的概述
2. get_pod_events   → 理解 WHY pods 处于它们的状态
3. get_pod_logs     → 仅当 events 不能解释问题时
4. get_pod_resources → 用于性能/资源 issues
5. describe_deployment → 检查 Deployment 状态和 conditions
```

## Common Issue Patterns

### CrashLoopBackOff

**首先检查**: `get_pod_events`

| Event Reason | Likely Cause | 下一步 Step |
|--------------|--------------|-----------|
| OOMKilled | Memory 限制太低或内存泄漏 | 检查 `get_pod_resources`, 增加 limits |
| Error | 应用 crash | 检查 `get_pod_logs` 获取 stack trace |
| BackOff | 重复失败 | 检查 logs 获取 startup errors |

**Checklist**:
- [ ] Memory limits vs actual usage
- [ ] Recent Deployment changes (`get_deployment_history`)
- [ ] Missing config/secrets
- [ ] Dependency failures (database, 外部 services)

### OOMKilled

**首先检查**: `get_pod_events` (确认 OOMKilled)
**然后**: `get_pod_resources` (比较 usage 和 limits)

**Common causes**:
- Memory 限制设置太低对于工作负载
- Memory leak (usage 随时间增加)
- 突然的流量峰值导致内存压力
- 大的请求 payloads 缓存在内存中

### ImagePullBackOff

**首先检查**: `get_pod_events`

**Common causes**:
- 错误的镜像名称或标签
- 没有 imagePullSecrets 的私有 registry
- 来自 registry 的速率限制
- 到达 registry 的网络 issues

### Pending Pods

**首先检查**: `get_pod_events`

**Look for**:
- `FailedScheduling` - 资源不足
- `Unschedulable` - 节点亲和性/taints
- 没有匹配 nodeSelector 的节点

### Readiness/Liveness 探针 Failures

**首先检查**: `describe_pod` (显示探针配置)
**然后**: `get_pod_events` (探针 failure events)
**然后**: `get_pod_logs` (为什么端点没有响应)

### Evicted Pods

**首先检查**: `get_pod_events`

**Causes**:
- 节点资源压力 (disk, memory)
- Priority preemption
- 基于 taint 的 eviction

## Deployment Issues

### Stuck Rollout

```
describe_deployment  → 检查 replicas (desired vs 就绪 vs 可用)
get_deployment_history → 比较 current vs previous revision
get_pod_events → 对于新 ReplicaSet 中的 pods
```

**Common causes**:
- 新 pods 失败 (CrashLoopBackOff)
- Readiness probes 失败
- 阻止调度的资源约束

### 回滚决策

使用 `get_deployment_history` 查看之前的可用版本。

## 错误分类

### Non-Retryable (立即停止)
- 401 Unauthorized - 无效 credentials
- 403 Forbidden - 没有权限
- 404 Not Found - 资源不存在
- "config_required": true - 集成未配置

### Retryable (可以重试一次)
- 429 Too Many Requests
- 500/502/503/504 Server errors
- Timeout
- Connection refused

## 资源调查模式

对于 memory/CPU issues：

```
1. get_pod_resources → 查看 allocation vs usage
2. describe_pod → 查看完整容器 spec
3. get_cloudwatch_metrics/query_datadog_metrics → 历史 usage
4. detect_anomalies on historical data → 查找 issue 开始的时间
```
