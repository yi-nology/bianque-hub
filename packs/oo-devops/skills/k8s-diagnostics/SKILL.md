---
name: k8s-diagnostics
description: Kubernetes 集群诊断专家 - 指标分析、健康检查、资源对比与集群状态评估
mode: on_demand
version: 1.1.0
maturity: experimental
---

# Kubernetes Diagnostics

Analyze cluster health and compare resources using kubectl-mcp-server's diagnostic tools.

## 何时应用

在以下情况下使用此技能:
- 用户 mentions: "metrics", "health checks", "compare", "analysis", "capacity"
- Operations: cluster health assessment, environment comparison, resource analysis
- Keywords: "how much", "usage", "difference between", "capacity planning"

## Priority Rules

| Priority | 规则 | Impact | Tools |
|----------|------|--------|-------|
| 1 | 检查 metrics-server 之前 using metrics | 严重 | `get_resource_metrics` |
| 2 | Run health checks 之前 deployments | 高 | `cluster_health_check` |
| 3 | Compare staging vs prod 之前 releases | 中 | `compare_namespaces` |
| 4 | Document baseline metrics | 低 | `get_nodes_summary` |

## 快速参考

| Task | Tool | Example |
|------|------|---------|
| Cluster health | `cluster_health_check` | `cluster_health_check()` |
| Pod metrics | `get_resource_metrics` | `get_resource_metrics(namespace)` |
| Node summary | `get_nodes_summary` | `get_nodes_summary()` |
| Compare envs | `compare_namespaces` | `compare_namespaces(ns1, ns2, type)` |
| List CRDs | `list_crds` | `list_crds()` |

## Resource Metrics

```python
get_resource_metrics(namespace="default")

get_node_metrics()

get_top_pods(namespace="default", sort_by="cpu")

get_top_pods(namespace="default", sort_by="memory")
```

## Cluster Health Checks

```python
cluster_health_check()

get_cluster_info()
```

## Compare Environments

```python
compare_namespaces(
    namespace1="staging",
    namespace2="production",
    resource_type="Deployment"
)

compare_namespaces(
    namespace1="default",
    namespace2="default",
    resource_type="Deployment",
    context1="staging-cluster",
    context2="prod-cluster"
)
```

## API Discovery

```python
get_api_versions()

check_crd_exists(crd_name="certificates.cert-manager.io")

list_crds()
```

## Resource Analysis

```python
get_nodes_summary()

kubeconfig_view()

list_contexts_tool()
```

## Diagnostic Workflows

### Cluster Overview

```python
cluster_health_check()
get_nodes_summary()
get_events(namespace="")
list_crds()
```

### Pre-Deployment Checks

```python
get_resource_metrics(namespace="production")
get_nodes_summary()
compare_namespaces(namespace1="staging", namespace2="prod", resource_type="Deployment")
```

### Post-incident Analysis

```python
get_events(namespace)
get_pod_logs(name, namespace, previous=True)
get_resource_metrics(namespace)
describe_node(name)
```

## 相关 Skills

- [k8s-troubleshoot](../k8s-troubleshoot/SKILL.md) - 调试 issues
- [k8s-cost](../k8s-cost/SKILL.md) - Cost analysis
- [k8s-incident](../k8s-incident/SKILL.md) - Incident response
