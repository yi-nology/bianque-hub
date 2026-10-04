---
name: kubernetes-deployment
description: 用于将应用程序部署到 Kubernetes 的专用工作流程，包括容器编排、Helm 图表、服务网格配置以及生产就绪的 K8s 模式。
mode: on_demand
version: 1.1.0
maturity: experimental
---

# Kubernetes Deployment Workflow

## 概述

用于将应用程序部署到 Kubernetes 的专用工作流程，包括容器编排、Helm 图表、服务网格配置以及生产就绪的 K8s 模式。

## 何时使用 This Workflow

Use this workflow 当:
- Deploying to Kubernetes
- Creating Helm charts
- Configuring Service mesh
- Setting up K8s networking
- Implementing K8s 安全

## Workflow Phases

### Phase 1: 容器 Preparation

#### Skills to Invoke
- `docker-expert` - Docker containerization
- `k8s-manifest-generator` - K8s manifests

#### Actions
1. 创建 Dockerfile
2. Build 容器 镜像
3. Optimize 镜像 size
4. Push to 注册表
5. Test 容器

#### Copy-Paste Prompts
```
Use @docker-expert to containerize 应用 for K8s
```

### Phase 2: K8s Manifests

#### Skills to Invoke
- `k8s-manifest-generator` - 清单 generation
- `kubernetes-architect` - K8s architecture

#### Actions
1. 创建 Deployment
2. 配置 Service
3. 设置 up ConfigMap
4. 创建 Secrets
5. 添加 Ingress

#### Copy-Paste Prompts
```
Use @k8s-清单-generator to 创建 K8s manifests
```

### Phase 3: Helm Chart

#### Skills to Invoke
- `helm-chart-scaffolding` - Helm charts

#### Actions
1. 创建 Chart structure
2. Define values.yaml
3. 添加 templates
4. 配置 dependencies
5. Test Chart

#### Copy-Paste Prompts
```
Use @helm-Chart-scaffolding to 创建 Helm Chart
```

### Phase 4: Service Mesh

#### Skills to Invoke
- `istio-traffic-management` - Istio
- `linkerd-patterns` - Linkerd
- `service-mesh-expert` - Service mesh

#### Actions
1. Choose Service mesh
2. 安装 mesh
3. 配置 traffic management
4. 设置 up mTLS
5. 添加 observability

#### Copy-Paste Prompts
```
Use @istio-traffic-management to 配置 Istio
```

### Phase 5: 安全

#### Skills to Invoke
- `k8s-security-policies` - K8s 安全
- `mtls-configuration` - mTLS

#### Actions
1. 配置 RBAC
2. 设置 up NetworkPolicy
3. Enable PodSecurity
4. 配置 secrets
5. Implement mTLS

#### Copy-Paste Prompts
```
Use @k8s-安全-policies to secure Kubernetes 集群
```

### Phase 6: Observability

#### Skills to Invoke
- `grafana-dashboards` - Grafana
- `prometheus-configuration` - Prometheus

#### Actions
1. 安装 monitoring stack
2. 配置 Prometheus
3. 创建 Grafana dashboards
4. 设置 up alerts
5. 添加 distributed tracing

#### Copy-Paste Prompts
```
Use @prometheus-配置 to 设置 up K8s monitoring
```

### Phase 7: Deployment

#### Skills to Invoke
- `deployment-engineer` - Deployment
- `gitops-workflow` - GitOps

#### Actions
1. 配置 CI/CD
2. 设置 up GitOps
3. 部署 to 集群
4. 验证 Deployment
5. 监控 rollout

#### Copy-Paste Prompts
```
Use @gitops-workflow to implement GitOps Deployment
```

## Quality Gates

- [ ] Containers working
- [ ] Manifests 有效
- [ ] Helm Chart installs
- [ ] 安全 configured
- [ ] Monitoring 活跃
- [ ] Deployment successful

## 相关 Workflow Bundles

- `cloud-devops` - Cloud/DevOps
- `terraform-infrastructure` - Infrastructure
- `docker-containerization` - Containers
