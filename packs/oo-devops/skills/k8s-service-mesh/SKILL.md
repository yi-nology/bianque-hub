---
name: k8s-service-mesh
description: 管理 Istio 服务网格以进行流量管理、安全和可观测性。用于流量切换、金丝雀发布、mTLS 和服务网格故障排除。
mode: on_demand
version: 1.1.0
maturity: experimental
---

# Kubernetes 服务网格 (Istio)

使用 kubectl-mcp-server 的 Istio/Kiali 工具进行流量管理、安全和可观测性。

## 何时应用

在以下情况下使用此技能：
- 用户提到："Istio"、"service mesh"、"mTLS"、"VirtualService"、"traffic shifting"
- 操作：流量管理、金丝雀部署、安全策略
- 关键词："sidecar"、"proxy"、"traffic split"、"mutual TLS"

## 优先级规则

| 优先级 | 规则 | 影响 | 工具 |
|----------|------|--------|-------|
| 1 | 首先检测 Istio 安装 | 严重 | `istio_detect_tool` |
| 2 | 更改前运行分析 | 高 | `istio_analyze_tool` |
| 3 | 检查代理状态以进行同步 | 高 | `istio_proxy_status_tool` |
| 4 | 验证 sidecar 注入 | 中 | `istio_sidecar_status_tool` |

## 快速参考

| 任务 | 工具 | 示例 |
|------|------|---------|
| 检测 Istio | `istio_detect_tool` | `istio_detect_tool()` |
| 分析配置 | `istio_analyze_tool` | `istio_analyze_tool(namespace)` |
| 代理状态 | `istio_proxy_status_tool` | `istio_proxy_status_tool()` |
| 列出 VirtualServices | `istio_virtualservices_list_tool` | `istio_virtualservices_list_tool(namespace)` |

## 快速状态检查

### 检测 Istio 安装

```python
istio_detect_tool()
```

### 检查代理状态

```python
istio_proxy_status_tool()
istio_sidecar_status_tool(namespace)
```

### 分析配置

```python
istio_analyze_tool(namespace)
```

## 流量管理

### VirtualServices

列出和检查：

```python
istio_virtualservices_list_tool(namespace)
istio_virtualservice_get_tool(name, namespace)
```

查看 [TRAFFIC-SHIFTING.md](TRAFFIC-SHIFTING.md) 了解金丝雀和蓝绿模式。

### DestinationRules

```python
istio_destinationrules_list_tool(namespace)
```

### Gateways

```python
istio_gateways_list_tool(namespace)
```

## 流量切换模式

### 金丝雀发布（基于权重）

90/10 分割的 VirtualService：

```yaml
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: my-service
spec:
  hosts:
  - my-service
  http:
  - route:
    - destination:
        host: my-service
        subset: stable
      weight: 90
    - destination:
        host: my-service
        subset: canary
      weight: 10
```

应用和验证：

```python
kubectl_apply(vs_yaml, namespace)
istio_virtualservice_get_tool("my-service", namespace)
```

### 基于 Header 的路由

路由 beta 用户：

```yaml
http:
- match:
  - headers:
      x-user-type:
        exact: beta
  route:
  - destination:
      host: my-service
      subset: canary
- route:
  - destination:
      host: my-service
      subset: stable
```

## 安全 (mTLS)

查看 [MTLS.md](MTLS.md) 了解详细的 mTLS 配置。

### PeerAuthentication（mTLS 模式）

```python
istio_peerauthentications_list_tool(namespace)
```

### AuthorizationPolicy

```python
istio_authorizationpolicies_list_tool(namespace)
```

## 可观测性

### 代理指标

```python
istio_proxy_status_tool()
```

### Hubble（Cilium 集成）

如果将 Cilium 与 Istio 一起使用：

```python
hubble_flows_query_tool(namespace)
cilium_endpoints_list_tool(namespace)
```

## 故障排除

### Sidecar 未注入

```python
istio_sidecar_status_tool(namespace)
```

### 流量未路由

```python
istio_analyze_tool(namespace)
istio_virtualservice_get_tool(name, namespace)
istio_destinationrules_list_tool(namespace)
istio_proxy_status_tool()
```

### mTLS 失败

```python
istio_peerauthentications_list_tool(namespace)
```

### 常见问题

| 症状 | 检查 | 解决方案 |
|---------|-------|------------|
| 503 错误 | `istio_analyze_tool()` | 修复 VirtualService/DestinationRule |
| 无 sidecar | `istio_sidecar_status_tool()` | 标记命名空间 |
| 配置未应用 | `istio_proxy_status_tool()` | 等待同步或重启 pod |

## 多集群服务网格

Istio 多集群设置：

```python
istio_proxy_status_tool(context="primary")
istio_virtualservices_list_tool(namespace, context="primary")

istio_proxy_status_tool(context="remote")
```

## 先决条件

- **Istio**：所有 Istio 工具都需要
  ```bash
  istioctl install --set profile=demo
  ```

## 相关技能

- [k8s-deploy](../k8s-deploy/SKILL.md) - 带流量切换的部署
- [k8s-security](../k8s-security/SKILL.md) - 授权策略
