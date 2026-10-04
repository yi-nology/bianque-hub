---
name: k8s-debug-pods
description: Kubernetes Pod 调试专家 - 诊断 Pending、CrashLoopBackOff、OOMKilled、ImagePullBackOff 等异常状态，检查节点资源与事件
mode: on_demand
version: 1.1.0
maturity: experimental
---

# K8s 调试 Pods

诊断和修复 Kubernetes 上 Kurtosis pods 的问题。

## 快速分类

```bash
# 查看跨 namespaces 的所有 kurtosis-related pods
kubectl get pods -A | grep kurtosis

# 检查问题 pods (not running)
kubectl get pods -A | grep kurtosis | grep -v Running

# 获取特定 Pod 的 events
kubectl describe pod <POD_NAME> -n <NAMESPACE> | tail -30
```

## Common Pod 状态和修复

### Pending — Unschedulable

Pod 由于节点 taints, 资源压力, 或亲和性规则而无法被调度。

```bash
# 检查节点 taints
kubectl get nodes -o custom-columns=NAME:.metadata.name,TAINTS:.spec.taints

# 检查节点 conditions (DiskPressure, MemoryPressure, 等.)
kubectl get nodes -o custom-columns=NAME:.metadata.name,CONDITIONS:.status.conditions[*].type
```

**Fix**: 在 `~/Library/Application Support/kurtosis/kurtosis-config.yml` 的 kurtosis 配置中添加 tolerations 或修复节点 condition。

### ImagePullBackOff

镜像标签在 registry 上不存在。

```bash
# 检查哪个镜像失败
kubectl describe pod <POD_NAME> -n <NAMESPACE> | grep -A5 "Image:"

# 验证镜像在 Docker Hub 上存在
docker manifest inspect <IMAGE>:<TAG>
```

**Fix**: 推送正确的镜像标签，或修复代码中的镜像引用。

### CrashLoopBackOff

容器启动但立即 crash。

```bash
# 检查容器 logs
kubectl logs <POD_NAME> -n <NAMESPACE>
kubectl logs <POD_NAME> -n <NAMESPACE> --previous
```

### Evicted

节点由于资源压力驱逐了 Pod。

```bash
# 检查哪些节点有压力
kubectl get nodes -o custom-columns=NAME:.metadata.name,STATUS:.status.conditions[-1].type

# 清理被驱逐的 pods
kubectl get pods -A | grep Evicted | awk '{print $2 " -n " $1}' | xargs -L1 kubectl delete pod
```

## Kurtosis-specific Pod 类型

| Pod pattern | Component | 镜像源 |
|-------------|-----------|-------------|
| `kurtosis-engine-*` | Engine server | `engine/server/Dockerfile` |
| `kurtosis-api` (在 `kt-*` namespaces 中) | API 容器 (APIC) | `core/server/Dockerfile` |
| `kurtosis-logs-collector-*` | Fluentbit DaemonSet | 从 registry 拉取 |
| `kurtosis-logs-aggregator-*` | Vector Deployment | 从 registry 拉取 |
| `remove-dir-pod-*` | Fluentbit cleanup pods | busybox |
| `files-artifact-expander` (init 容器) | Files artifacts | `core/files_artifacts_expander/Dockerfile` |

## Engine start 失败

如果 `kurtosis engine start` 失败：

1. 检查旧的 kurtosis namespaces 是否存在：`kubectl get ns | grep kurtosis`
2. 删除它们：`kubectl get ns | grep kurtosis | awk '{print $1}' | xargs -r kubectl delete ns`
3. 重试 engine start

## Logs collector issues

Logs collector 是一个在每个节点上运行的 DaemonSet。如果某些节点不健康：

```bash
# 检查 DaemonSet 状态
kubectl get ds -A | grep kurtosis

# 查看哪些 pods 没有运行
kubectl get pods -A | grep logs-collector | grep -v Running
```

有 DiskPressure 或其他 taints 的节点可能不会调度 collector pods — 这是预期的，engine 应该以关于部分降级收集的警告启动。
