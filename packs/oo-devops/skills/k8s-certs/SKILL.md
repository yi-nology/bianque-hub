---
name: k8s-certs
description: 使用 cert-manager 进行 Kubernetes 证书管理。在以下情况下使用：管理 TLS 证书、配置 issuers 或排除证书问题。
mode: on_demand
version: 1.1.0
maturity: experimental
---

# 使用 cert-manager 进行证书管理

使用 kubectl-mcp-server 的 cert-manager 工具管理 TLS 证书。

## 何时应用

在以下情况下使用此技能：
- 用户提到："certificate", "cert-manager", "TLS", "SSL", "issuer", "Let's Encrypt"
- 操作：创建证书、配置 issuers、调试证书问题
- 关键词："https", "secure", "encrypt", "renew", "expiring"

## 优先级规则

| Priority | 规则 | Impact | Tools |
|----------|------|--------|-------|
| 1 | 首先检测 cert-manager | Critical | `certmanager_detect_tool` |
| 2 | 测试使用 staging issuer | High | Test with letsencrypt-staging |
| 3 | 在证书之前检查 issuer | High | `certmanager_clusterissuers_list_tool` |
| 4 | 监控证书 expiry | Medium | `certmanager_certificate_get_tool` |

## 快速参考

| Task | Tool | Example |
|------|------|---------|
| 检测 cert-manager | `certmanager_detect_tool` | `certmanager_detect_tool()` |
| 列表 certificates | `certmanager_certificates_list_tool` | `certmanager_certificates_list_tool(namespace)` |
| 获取证书 | `certmanager_certificate_get_tool` | `certmanager_certificate_get_tool(name, namespace)` |
| 列表 issuers | `certmanager_clusterissuers_list_tool` | `certmanager_clusterissuers_list_tool()` |

## 检查安装

```python
certmanager_detect_tool()
```

## Certificates

### 列表 Certificates

```python
certmanager_certificates_list_tool(namespace="default")
```

### 获取证书详细信息

```python
certmanager_certificate_get_tool(
    name="my-tls",
    namespace="default"
)
```

### 创建证书

```python
kubectl_apply(manifest="""
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: my-tls
  namespace: default
spec:
  secretName: my-tls-secret
  issuerRef:
    name: letsencrypt-prod
    kind: ClusterIssuer
  dnsNames:
  - app.example.com
  - www.example.com
""")
```

## Issuers

### 列表 Issuers

```python
certmanager_issuers_list_tool(namespace="default")

certmanager_clusterissuers_list_tool()
```

### 获取 Issuer 详细信息

```python
certmanager_issuer_get_tool(name="my-issuer", namespace="default")
certmanager_clusterissuer_get_tool(name="letsencrypt-prod")
```

### 创建 Let's Encrypt Issuer

```python
kubectl_apply(manifest="""
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-staging
spec:
  acme:
    server: https://acme-staging-v02.api.letsencrypt.org/directory
    email: admin@example.com
    privateKeySecretRef:
      name: letsencrypt-staging-key
    solvers:
    - http01:
        ingress:
          class: nginx
""")

kubectl_apply(manifest="""
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: admin@example.com
    privateKeySecretRef:
      name: letsencrypt-prod-key
    solvers:
    - http01:
        ingress:
          class: nginx
""")
```

### 创建 Self-Signed Issuer

```python
kubectl_apply(manifest="""
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: selfsigned
spec:
  selfSigned: {}
""")
```

## 证书 Requests

```python
certmanager_certificaterequests_list_tool(namespace="default")

certmanager_certificaterequest_get_tool(
    name="my-tls-xxxxx",
    namespace="default"
)
```

## 故障排除

### 证书 Not Ready

```python
certmanager_certificate_get_tool(name, namespace)
certmanager_certificaterequests_list_tool(namespace)
get_events(namespace)
```

### Issuer Not Ready

```python
certmanager_clusterissuer_get_tool(name)
get_events(namespace="cert-manager")
```

## Ingress 集成

```python
kubectl_apply(manifest="""
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: my-ingress
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
spec:
  tls:
  - hosts:
    - app.example.com
    secretName: app-tls
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

## 先决条件

- **cert-manager**: 所有证书工具的必需组件
  ```bash
  kubectl apply -f https://github.com/cert-manager/cert-manager/releases/latest/download/cert-manager.yaml
  ```

## 相关 Skills

- [k8s-networking](../k8s-networking/SKILL.md) - Ingress 配置
- [k8s-security](../k8s-security/SKILL.md) - 安全最佳实践
