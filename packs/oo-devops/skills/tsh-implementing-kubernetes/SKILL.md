---
name: tsh-implementing-kubernetes
description: Kubernetes 部署模式和集群管理专家，涵盖 Helm Chart、工作负载配置、扩展策略和集群资源管理
mode: on_demand
version: 1.1.0
maturity: experimental
---

# Kubernetes 模式

## 使用场景

- 将应用程序部署到 Kubernetes
- 设计 Deployment、StatefulSet 或 Job 配置
- 实现自动扩展（HPA、VPA、KEDA）
- 创建或修改 Helm Chart
- 设置入口、网络和服务网格
- 配置资源请求、限制和 QoS

## 项目检测

检查项目使用哪些 Kubernetes 工具：
- `helm/` 或 `Chart.yaml` → Helm Chart
- `kustomize/` 或 `kustomization.yaml` → Kustomize
- `k8s/` 或 `kubernetes/` 包含 `*.yaml` → 原始清单
- `skaffold.yaml` → 用于本地开发的 Skaffold
- `argocd/` 或 `Application` 资源 → ArgoCD GitOps
- `flux-system/` 或 `Kustomization` CRD → Flux GitOps

使用 `context7` 查找 Kubernetes API 版本和语法。

## 工作负载类型决策

| 工作负载类型 | 使用场景 |
|---------------|----------|
| **Deployment** | 无状态应用、Web 服务器、API |
| **StatefulSet** | 数据库、需要稳定标识的有状态应用 |
| **DaemonSet** | 节点级代理（日志、监控） |
| **Job** | 一次性任务、批处理 |
| **CronJob** | 定时周期性任务 |

## 部署配置

### 资源管理

```yaml
resources:
  requests:    # 调度器用于放置
    memory: "256Mi"
    cpu: "100m"
  limits:      # Kubelet 强制执行这些
    memory: "512Mi"
    cpu: "500m"
```

**规则：**
- 始终设置请求（调度必需）
- 设置内存限制以防止 OOM 影响节点
- CPU 限制可选（可能导致节流）
- 请求:限制比例为 1:2 是良好的起点

### QoS 类别

| 类别 | 条件 | 驱逐优先级 |
|-------|-----------|-------------------|
| **Guaranteed** | 请求 == 限制（所有容器） | 最后驱逐 |
| **Burstable** | 请求 < 限制 | 中等 |
| **BestEffort** | 无请求或限制 | 首先驱逐 |

**规则：** 生产工作负载应为 Guaranteed 或 Burstable，绝不能是 BestEffort。

### 探针配置

```yaml
livenessProbe:      # 失败时重启容器
  httpGet:
    path: /healthz
    port: 8080
  initialDelaySeconds: 15
  periodSeconds: 10
  failureThreshold: 3

readinessProbe:     # 失败时从 Service 移除
  httpGet:
    path: /ready
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 5
  failureThreshold: 3

startupProbe:       # 启动完成前延迟存活检查
  httpGet:
    path: /healthz
    port: 8080
  failureThreshold: 30
  periodSeconds: 10
```

**规则：**
- 始终配置 readinessProbe（优雅的流量处理）
- 对启动缓慢的应用使用 startupProbe（替代较长的 initialDelaySeconds）
- livenessProbe 应检查应用健康，而非依赖项

### Pod 中断预算

```yaml
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: api-pdb
spec:
  minAvailable: 2        # 或 maxUnavailable: 1
  selector:
    matchLabels:
      app: api
```

**规则：** 始终为生产工作负载创建 PDB，以确保节点驱逐期间的可用性。

### Pod 反亲和性

```yaml
affinity:
  podAntiAffinity:
    preferredDuringSchedulingIgnoredDuringExecution:
      - weight: 100
        podAffinityTerm:
          labelSelector:
            matchLabels:
              app: api
          topologyKey: kubernetes.io/hostname
```

**规则：** 跨节点/区域分布副本以实现高可用性。

## 扩展策略

### 水平 Pod 自动扩展器（HPA）

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: api-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: api
  minReplicas: 2
  maxReplicas: 10
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 70
  behavior:
    scaleDown:
      stabilizationWindowSeconds: 300  # 防止波动
```

### 扩展决策矩阵

| 扩展类型 | 使用场景 | 工具 |
|--------------|----------|------|
| 基于 CPU | 通用计算工作负载 | HPA |
| 基于内存 | 内存密集型应用 | HPA |
| 自定义指标 | 队列深度、请求速率 | HPA + Prometheus Adapter |
| 事件驱动 | 消息队列、定时作业 | KEDA |
| 垂直 | 调整请求/限制大小 | VPA |

## Helm Chart 结构

```
mychart/
├── Chart.yaml          # Chart 元数据
├── values.yaml         # 默认值
├── values-dev.yaml     # 环境覆盖
├── values-prod.yaml
├── templates/
│   ├── _helpers.tpl    # 模板助手
│   ├── deployment.yaml
│   ├── service.yaml
│   ├── ingress.yaml
│   ├── hpa.yaml
│   ├── pdb.yaml
│   └── configmap.yaml
└── charts/             # 依赖
```

### Helm 最佳实践

```yaml
# values.yaml - 使用结构化默认值
replicaCount: 2

image:
  repository: myapp
  tag: ""  # 在 CI 中覆盖，不要在这里设置
  pullPolicy: IfNotPresent

resources:
  requests:
    memory: "256Mi"
    cpu: "100m"
  limits:
    memory: "512Mi"

# 启用/禁用可选组件
autoscaling:
  enabled: true
  minReplicas: 2
  maxReplicas: 10
```

**规则：**
- 不要在 values.yaml 中硬编码镜像标签（在 CI 中设置）
- 对资源名称使用 `{{ include "mychart.fullname" . }}`
- 提供合理的默认值，按环境覆盖

## 入口配置

### 入口类决策

| 入口控制器 | 使用场景 |
|-------------------|----------|
| **nginx-ingress** | 通用，广泛支持 |
| **AWS ALB** | AWS 原生，与 WAF/ACM 集成 |
| **Traefik** | 简单设置，自动 HTTPS |
| **Istio Gateway** | 已使用服务网格 |

### 入口示例（nginx）

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: api-ingress
  annotations:
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
spec:
  ingressClassName: nginx
  tls:
    - hosts:
        - api.example.com
      secretName: api-tls
  rules:
    - host: api.example.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: api
                port:
                  number: 80
```

## 安全上下文

```yaml
securityContext:
  runAsNonRoot: true
  runAsUser: 1000
  runAsGroup: 1000
  fsGroup: 1000
  seccompProfile:
    type: RuntimeDefault

containers:
  - name: app
    securityContext:
      allowPrivilegeEscalation: false
      readOnlyRootFilesystem: true
      capabilities:
        drop:
          - ALL
```

**规则：** 在生产环境中始终以非 root 用户运行，并授予最小权限。

## 流程

1. **发现上下文** → 检查现有的 K8s 清单、Helm Chart、Kustomize
2. **选择工作负载类型** → 根据需求选择 Deployment、StatefulSet、Job
3. **配置资源** → 根据性能分析或估算设置请求/限制
4. **添加探针** → 配置就绪、存活和启动探针
5. **启用扩展** → 根据扩展需求添加 HPA/KEDA
6. **添加弹性** → PDB、Pod 反亲和性、拓扑分布
7. **配置安全** → 安全上下文、网络策略
8. **验证** → `kubectl apply --dry-run=server`、helm template

## 检查清单

- [ ] 定义了资源请求和限制
- [ ] 配置了就绪和存活探针
- [ ] 为生产环境创建了 PodDisruptionBudget
- [ ] 配置了 Pod 反亲和性或拓扑分布
- [ ] 为可变工作负载配置了 HPA
- [ ] 配置了非 root 用户的安全上下文
- [ ] 镜像拉取策略适当（生产中绝不使用 `latest`）
- [ ] 标签一致（`app`、`version`、`environment`）
- [ ] 按环境的命名空间隔离

## 反模式

| 不要 | 要 |
|-------|-----|
| 使用 `latest` 镜像标签 | 固定特定版本或 SHA |
| 跳过资源请求 | 始终设置请求以供调度 |
| 生产环境单副本 | 最少 2 个副本并带 PDB |
| 以 root 运行 | 使用非 root 用户并授予最小权限 |
| 缺少就绪探针 | 配置探针以实现优雅的流量处理 |
| 生产环境使用 `kubectl apply` | 使用 ArgoCD/Flux 进行 GitOps |
| 在清单中硬编码值 | 使用 Helm values 或 Kustomize 覆盖 |
| 忽略 Pod 驱逐 | 设置 PDB 以保持可用性 |

## 相关技能

- `tsh-implementing-observability` - 用于 K8s 监控和日志设置
- `tsh-implementing-ci-cd` - 用于 K8s 部署流水线
- `tsh-managing-secrets` - 用于 K8s 密钥管理模式
- `tsh-implementing-terraform-modules` - 用于配置 K8s 集群
