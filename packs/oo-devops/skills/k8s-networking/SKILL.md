---
name: k8s-networking
description: Kubernetes 网络管理，包括 services、ingresses、endpoints 和 network policies。在配置连接性、负载均衡或网络隔离时使用。
mode: on_demand
version: 1.1.0
maturity: experimental
---

# Kubernetes 网络

使用 kubectl-mcp-server 的网络工具管理 Kubernetes 网络资源。

## 何时应用

在以下情况下使用此技能：
- 用户提到："service"、"ingress"、"endpoints"、"network policy"、"load balancer"
- 操作：暴露应用、配置路由、网络隔离
- 关键词："connectivity"、"DNS"、"traffic"、"port"、"firewall"

## 优先级规则

| 优先级 | 规则 | 影响 | 工具 |
|----------|------|--------|-------|
| 1 | 排查 service 问题前先检查 endpoints | 严重 | `get_endpoints` |
| 2 | 验证 Service selector 匹配 Pod labels | 高 | `get_services`、`get_pods` |
| 3 | 检查网络策略进行隔离 | 高 | `get_network_policies` |
| 4 | 从 Pod 内测试 DNS 解析 | 中 | `kubectl_exec` |

## 快速参考

| 任务 | 工具 | 示例 |
|------|------|---------|
| 列出 services | `get_services` | `get_services(namespace)` |
| 检查 backends | `get_endpoints` | `get_endpoints(namespace)` |
| 列出 ingresses | `get_ingresses` | `get_ingresses(namespace)` |
| 网络策略 | `get_network_policies` | `get_network_policies(namespace)` |

## Services

```python
get_services(namespace="default")

describe_service(name="my-service", namespace="default")

create_service(
    name="my-service",
    namespace="default",
    selector={"app": "my-app"},
    ports=[{"port": 80, "targetPort": 8080}]
)

create_service(
    name="my-lb",
    namespace="default",
    type="LoadBalancer",
    selector={"app": "my-app"},
    ports=[{"port": 443, "targetPort": 8443}]
)
```

## Endpoints

```python
get_endpoints(namespace="default")
```

## Ingress

```python
get_ingresses(namespace="default")

describe_ingress(name="my-ingress", namespace="default")

kubectl_apply(manifest="""
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: my-ingress
  namespace: default
spec:
  rules:
  - host: app.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: my-service
            port:
              number: 80
""")
```

## Network Policies

```python
get_network_policies(namespace="default")

describe_network_policy(name="deny-all", namespace="default")

kubectl_apply(manifest="""
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: deny-all
  namespace: default
spec:
  podSelector: {}
  policyTypes:
  - Ingress
  - Egress
""")

kubectl_apply(manifest="""
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-web
  namespace: default
spec:
  podSelector:
    matchLabels:
      app: web
  ingress:
  - from:
    - podSelector:
        matchLabels:
          app: frontend
    ports:
    - port: 80
""")
```

## 排查连接问题

```python
get_endpoints(namespace="default")

get_network_policies(namespace="default")

kubectl_exec(
    pod="debug-pod",
    namespace="default",
    command="nslookup my-service.default.svc.cluster.local"
)
```

## 相关技能

- [k8s-service-mesh](../k8s-service-mesh/SKILL.md) - Istio 流量管理
- [k8s-cilium](../k8s-cilium/SKILL.md) - Cilium 网络策略
