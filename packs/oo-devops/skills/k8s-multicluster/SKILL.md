---
name: k8s-multicluster
description: 管理多个 Kubernetes 集群，切换上下文，执行跨集群操作。在处理多个集群、比较环境或管理集群生命周期时使用。
mode: on_demand
version: 1.1.0
maturity: experimental
---

# 多集群 Kubernetes 管理

使用 kubectl-mcp-server 的多集群支持进行跨集群操作和上下文管理。

## 何时应用

在以下情况下使用此技能：
- 用户提到："cluster"、"context"、"multi-cluster"、"cross-cluster"
- 操作：切换上下文、比较集群、联邦部署
- 关键词："different environment"、"production vs staging"、"all clusters"

## 优先级规则

| 优先级 | 规则 | 影响 | 工具 |
|----------|------|--------|-------|
| 1 | 生产环境始终指定上下文 | 严重 | `context` 参数 |
| 2 | 切换前先列出上下文 | 高 | `list_contexts_tool` |
| 3 | 升级前先比较 | 中 | `compare_namespaces` |
| 4 | 使用命名约定 | 低 | `prod-*`、`staging-*` |

## 快速参考

| 任务 | 工具 | 示例 |
|------|------|---------|
| 列出上下文 | `list_contexts_tool` | `list_contexts_tool()` |
| 查看 kubeconfig | `kubeconfig_view` | `kubeconfig_view()` |
| 列出 CAPI 集群 | `capi_clusters_list_tool` | `capi_clusters_list_tool(namespace)` |
| 获取 CAPI kubeconfig | `capi_cluster_kubeconfig_tool` | `capi_cluster_kubeconfig_tool(name, namespace)` |

## 上下文管理

### 列出可用上下文

```python
list_contexts_tool()
```

### 查看当前上下文

```python
kubeconfig_view()
```

### 切换上下文

CLI: `kubectl-mcp-server context <context-name>`

## 跨集群操作

所有 kubectl-mcp-server 工具都支持 `context` 参数：

```python
get_pods(namespace="default", context="production-cluster")

get_pods(namespace="default", context="staging-cluster")
```

## 常见多集群模式

### 比较环境

```python
compare_namespaces(
    namespace1="production",
    namespace2="staging",
    resource_type="Deployment",
    context="production-cluster"
)
```

### 并行查询

同时查询多个集群：

```python
get_pods(namespace="app", context="prod-us-east")
get_pods(namespace="app", context="prod-eu-west")

get_pods(namespace="app", context="development")
```

### 跨集群健康检查

```python
for context in ["prod-1", "prod-2", "staging"]:
    get_nodes(context=context)
    get_pods(namespace="kube-system", context=context)
```

## Cluster API (CAPI) 管理

用于管理集群生命周期：

### 列出托管集群

```python
capi_clusters_list_tool(namespace="capi-system")
```

### 获取集群详情

```python
capi_cluster_get_tool(name="prod-cluster", namespace="capi-system")
```

### 获取工作负载集群 Kubeconfig

```python
capi_cluster_kubeconfig_tool(name="prod-cluster", namespace="capi-system")
```

### 机器管理

```python
capi_machines_list_tool(namespace="capi-system")
capi_machinedeployments_list_tool(namespace="capi-system")
```

### 扩缩容集群

```python
capi_machinedeployment_scale_tool(
    name="prod-cluster-md-0",
    namespace="capi-system",
    replicas=5
)
```

详见 [CONTEXT-SWITCHING.md](CONTEXT-SWITCHING.md) 了解详细模式。

## 多集群 Helm

部署 charts 到指定集群：

```python
install_helm_chart(
    name="nginx",
    chart="bitnami/nginx",
    namespace="web",
    context="production-cluster"
)

list_helm_releases(
    namespace="web",
    context="staging-cluster"
)
```

## 多集群 GitOps

### 跨集群 Flux

```python
flux_kustomizations_list_tool(
    namespace="flux-system",
    context="cluster-1"
)

flux_reconcile_tool(
    kind="kustomization",
    name="apps",
    namespace="flux-system",
    context="cluster-2"
)
```

### 跨集群 ArgoCD

```python
argocd_apps_list_tool(namespace="argocd", context="management-cluster")
```

## 联邦模式

### Secret 同步

```python
get_secrets(namespace="app", context="source-cluster")

kubectl_apply(secret_manifest, namespace="app", context="target-cluster")
```

### 跨集群服务发现

使用 Cilium ClusterMesh 或 Istio 多集群：

```python
cilium_nodes_list_tool(context="cluster-1")
istio_proxy_status_tool(context="cluster-2")
```

## 最佳实践

1. **命名约定**: 使用描述性上下文名称 (`prod-us-east-1`、`staging-eu-west-1`)
2. **访问控制**: 不同环境使用不同的 kubeconfig
3. **始终指定上下文**: 避免意外的跨集群操作
4. **集群分组**: 按用途组织 (`prod-*`、`staging-*`、`dev-*`)

## 相关技能

- [k8s-troubleshoot](../k8s-troubleshoot/SKILL.md) - 跨集群调试
- [k8s-gitops](../k8s-gitops/SKILL.md) - GitOps 多集群
- [k8s-capi](../k8s-capi/SKILL.md) - Cluster API 管理
