---
name: k8s-cilium
description: 用于 Kubernetes 的 Cilium 和 Hubble 网络可观测性。在以下情况下使用：管理网络策略、观察流量流或使用基于 eBPF 的网络进行连接故障排除。
mode: on_demand
version: 1.1.0
maturity: experimental
---

# Cilium & Hubble 网络可观测性

使用 kubectl-mcp-server 的 Cilium 工具（8 个工具）管理基于 eBPF 的网络。

## 何时应用

在以下情况下使用此技能：
- 用户提到："Cilium", "Hubble", "eBPF", "network policy", "flow"
- 操作：网络策略管理、流量观察、L7 过滤
- 关键词："network security", "traffic flow", "dropped packets", "connectivity"

## 优先级规则

| Priority | 规则 | Impact | Tools |
|----------|------|--------|-------|
| 1 | 首先检测 Cilium 安装 | Critical | `cilium_detect_tool` |
| 2 | 检查 agent 状态以确保健康 | High | `cilium_status_tool` |
| 3 | 使用 Hubble 进行 flow 调试 | High | `hubble_flows_query_tool` |
| 4 | 从默认 deny 开始 | Medium | CiliumNetworkPolicy |

## 快速参考

| Task | Tool | Example |
|------|------|---------|
| 检测 Cilium | `cilium_detect_tool` | `cilium_detect_tool()` |
| Agent 状态 | `cilium_status_tool` | `cilium_status_tool()` |
| 列表 policies | `cilium_policies_list_tool` | `cilium_policies_list_tool(namespace)` |
| Query flows | `hubble_flows_query_tool` | `hubble_flows_query_tool(namespace)` |

## 检查安装

```python
cilium_detect_tool()
```

## Cilium 状态

```python
cilium_status_tool()
```

## 网络 Policies

### 列表 Policies

```python
cilium_policies_list_tool(namespace="default")
```

### 获取策略详细信息

```python
cilium_policy_get_tool(name="allow-web", namespace="default")
```

### 创建 Cilium 网络策略

```python
kubectl_apply(manifest="""
apiVersion: cilium.io/v2
kind: CiliumNetworkPolicy
metadata:
  name: allow-web
  namespace: default
spec:
  endpointSelector:
    matchLabels:
      app: web
  ingress:
  - fromEndpoints:
    - matchLabels:
        app: frontend
    toPorts:
    - ports:
      - port: "80"
        protocol: TCP
  egress:
  - toEndpoints:
    - matchLabels:
        app: database
    toPorts:
    - ports:
      - port: "5432"
        protocol: TCP
""")
```

## Endpoints

```python
cilium_endpoints_list_tool(namespace="default")
```

## Identities

```python
cilium_identities_list_tool()
```

## Nodes

```python
cilium_nodes_list_tool()
```

## Hubble Flow 可观测性

```python
hubble_flows_query_tool(
    namespace="default",
    pod="my-pod",
    last="5m"
)

hubble_flows_query_tool(
    namespace="default",
    verdict="DROPPED"
)

hubble_flows_query_tool(
    namespace="default",
    type="l7"
)
```

## 创建 L7 策略

```python
kubectl_apply(manifest="""
apiVersion: cilium.io/v2
kind: CiliumNetworkPolicy
metadata:
  name: api-policy
  namespace: default
spec:
  endpointSelector:
    matchLabels:
      app: api
  ingress:
  - fromEndpoints:
    - matchLabels:
        app: frontend
    toPorts:
    - ports:
      - port: "8080"
        protocol: TCP
      rules:
        http:
        - method: GET
          path: "/api/v1/.*"
        - method: POST
          path: "/api/v1/users"
""")
```

## 集群 Mesh

```python
kubectl_apply(manifest="""
apiVersion: cilium.io/v2
kind: CiliumClusterwideNetworkPolicy
metadata:
  name: allow-cross-cluster
spec:
  endpointSelector:
    matchLabels:
      app: shared-service
  ingress:
  - fromEntities:
    - cluster
    - remote-node
""")
```

## 故障排除 Workflows

### Pod 无法访问 Service

```python
cilium_status_tool()
cilium_endpoints_list_tool(namespace)
cilium_policies_list_tool(namespace)
hubble_flows_query_tool(namespace, pod, verdict="DROPPED")
```

### 策略不工作

```python
cilium_policy_get_tool(name, namespace)
cilium_endpoints_list_tool(namespace)
hubble_flows_query_tool(namespace)
```

### 网络性能问题

```python
cilium_status_tool()
cilium_nodes_list_tool()
hubble_flows_query_tool(namespace, type="l7")
```

## 最佳实践

1. **从默认 deny 开始**: 创建 baseline deny-all 策略
2. **一致使用标签**: 策略依赖标签选择器
3. **使用 Hubble 监控**: 在策略更改之前/之后观察 flows
4. **在 staging 中测试**: 验证策略不会破坏连接

## 先决条件

- **Cilium**: 所有 Cilium 工具的必需组件
  ```bash
  cilium install
  ```

## 相关 Skills

- [k8s-networking](../k8s-networking/SKILL.md) - 标准 K8s 网络
- [k8s-security](../k8s-security/SKILL.md) - 安全策略
- [k8s-service-mesh](../k8s-service-mesh/SKILL.md) - Istio Service mesh
