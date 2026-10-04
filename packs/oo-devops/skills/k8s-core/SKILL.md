---
name: k8s-core
description: Kubernetes 核心资源管理专家 - Pod、Namespace、ConfigMap、Secret、Node 的查看、检查与管理
mode: on_demand
version: 1.1.0
maturity: experimental
---

# Core Kubernetes Resources

使用 kubectl-mcp-server 的 core 工具管理 fundamental Kubernetes 对象。

## 何时应用

在以下情况下使用此技能：
- 用户提到："pods", "namespaces", "configmaps", "secrets", "nodes", "events"
- 操作：listing 资源, describing 对象, creating/deleting 资源
- 关键词："show me", "list", "get", "describe", "create", "delete"

## 优先级规则

| Priority | 规则 | Impact | Tools |
|----------|------|--------|-------|
| 1 | 在操作之前检查 namespace 是否存在 | Critical | `get_namespaces` |
| 2 | 永远不要以纯文本暴露 secrets | Critical | 小心处理 `get_secret` 输出 |
| 3 | 使用标签进行过滤 | High | `label_selector` 参数 |
| 4 | 在更改之后检查 events | Medium | `get_events` |

## 快速参考

| Task | Tool | Example |
|------|------|---------|
| 列表 pods | `get_pods` | `get_pods(namespace="default")` |
| Describe Pod | `describe_pod` | `describe_pod(name, namespace)` |
| 获取 logs | `get_pod_logs` | `get_pod_logs(name, namespace)` |
| 列表 namespaces | `get_namespaces` | `get_namespaces()` |
| 获取 ConfigMap | `get_configmap` | `get_configmap(name, namespace)` |
| 列表 nodes | `get_nodes` | `get_nodes()` |

## Pods

```python
get_pods(namespace="default")
get_pods(namespace="kube-system", label_selector="app=nginx")

describe_pod(name="my-pod", namespace="default")

get_pod_logs(name="my-pod", namespace="default")
get_pod_logs(name="my-pod", namespace="default", previous=True)

delete_pod(name="my-pod", namespace="default")
```

## Namespaces

```python
get_namespaces()

create_namespace(name="my-namespace")

delete_namespace(name="my-namespace")
```

## ConfigMaps

```python
get_configmaps(namespace="default")

get_configmap(name="my-config", namespace="default")

create_configmap(
    name="app-config",
    namespace="default",
    data={"key": "value", "config.yaml": "setting: true"}
)
```

## Secrets

```python
get_secrets(namespace="default")

get_secret(name="my-secret", namespace="default")

create_secret(
    name="db-credentials",
    namespace="default",
    data={"username": "admin", "password": "secret123"}
)
```

## Nodes

```python
get_nodes()

describe_node(name="node-1")

get_nodes_summary()

cordon_node(name="node-1")
uncordon_node(name="node-1")

drain_node(name="node-1", ignore_daemonsets=True)
```

## Events

```python
get_events(namespace="default")

get_events(namespace="default", field_selector="involvedObject.name=my-pod")
```

## 多集群支持

所有工具支持 `context` 参数：

```python
get_pods(namespace="default", context="production-cluster")
get_nodes(context="staging-cluster")
```

## 相关 Skills

- [k8s-troubleshoot](../k8s-troubleshoot/SKILL.md) - 调试 failing pods
- [k8s-operations](../k8s-operations/SKILL.md) - kubectl apply/patch/delete
