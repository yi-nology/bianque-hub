---
name: k8s-operations
description: kubectl 操作，用于应用、修补、删除和在 Kubernetes 资源上执行命令。在修改资源、在 Pod 中运行命令或管理资源生命周期时使用。
mode: on_demand
version: 1.1.0
maturity: experimental
---

# kubectl 操作

使用 kubectl-mcp-server 的操作工具执行 kubectl 命令。

## 何时应用

在以下情况下使用此技能：
- 用户提到："apply"、"patch"、"delete"、"exec"、"scale"、"rollout"
- 操作：修改资源、运行命令、扩缩容工作负载
- 关键词："update"、"change"、"modify"、"run command"、"restart"

## 优先级规则

| 优先级 | 规则 | 影响 | 工具 |
|----------|------|--------|-------|
| 1 | 生产环境 apply 前先 dry run | 严重 | `kubectl_apply(dry_run=True)` |
| 2 | patch 前检查当前状态 | 高 | `describe_*` 工具 |
| 3 | 除非必要否则避免强制删除 | 高 | `kubectl_delete` |
| 4 | 更改后验证 rollout 状态 | 中 | `kubectl_rollout_status` |

## 快速参考

| 任务 | 工具 | 示例 |
|------|------|---------|
| 应用 manifest | `kubectl_apply` | `kubectl_apply(manifest=yaml)` |
| 修补资源 | `kubectl_patch` | `kubectl_patch(type, name, namespace, patch)` |
| 删除资源 | `kubectl_delete` | `kubectl_delete(type, name, namespace)` |
| 执行命令 | `kubectl_exec` | `kubectl_exec(pod, namespace, command)` |
| 扩缩容 deployment | `scale_deployment` | `scale_deployment(name, namespace, replicas)` |

## 应用资源

```python
kubectl_apply(manifest="""
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
  namespace: default
spec:
  replicas: 3
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:latest
""")

kubectl_apply(file_path="/path/to/manifest.yaml")

kubectl_apply(manifest="...", dry_run=True)
```

## 修补资源

```python
kubectl_patch(
    resource_type="Deployment",
    name="nginx",
    namespace="default",
    patch={"spec": {"replicas": 5}}
)

kubectl_patch(
    resource_type="Deployment",
    name="nginx",
    namespace="default",
    patch=[{"op": "replace", "path": "/spec/replicas", "value": 5}],
    patch_type="json"
)

kubectl_patch(
    resource_type="Service",
    name="my-svc",
    namespace="default",
    patch={"metadata": {"annotations": {"key": "value"}}},
    patch_type="merge"
)
```

## 删除资源

```python
kubectl_delete(resource_type="Pod", name="my-pod", namespace="default")

kubectl_delete(
    resource_type="pods",
    namespace="default",
    label_selector="app=test"
)

kubectl_delete(
    resource_type="Pod",
    name="stuck-pod",
    namespace="default",
    force=True,
    grace_period=0
)
```

## 执行命令

```python
kubectl_exec(
    pod="my-pod",
    namespace="default",
    command="ls -la /app"
)

kubectl_exec(
    pod="my-pod",
    namespace="default",
    container="sidecar",
    command="cat /etc/config/settings.yaml"
)

kubectl_exec(
    pod="my-pod",
    namespace="default",
    command="sh -c 'curl -s localhost:8080/health'"
)
```

## 扩缩容资源

```python
scale_deployment(name="nginx", namespace="default", replicas=5)

scale_deployment(name="nginx", namespace="default", replicas=0)

kubectl_scale(
    resource_type="statefulset",
    name="mysql",
    namespace="default",
    replicas=3
)
```

## Rollout 管理

```python
kubectl_rollout_status(
    resource_type="Deployment",
    name="nginx",
    namespace="default"
)

kubectl_rollout_history(
    resource_type="Deployment",
    name="nginx",
    namespace="default"
)

kubectl_rollout_restart(
    resource_type="Deployment",
    name="nginx",
    namespace="default"
)

rollback_deployment(name="nginx", namespace="default", revision=1)
```

## 标签和注解

```python
kubectl_label(
    resource_type="Pod",
    name="my-pod",
    namespace="default",
    labels={"env": "production"}
)

kubectl_annotate(
    resource_type="Deployment",
    name="nginx",
    namespace="default",
    annotations={"description": "Main web server"}
)
```

## 相关技能

- [k8s-deploy](../k8s-deploy/SKILL.md) - 部署策略
- [k8s-helm](../k8s-helm/SKILL.md) - Helm 操作
