---
name: k8s-manifest-generator
description: 创建生产就绪的 Kubernetes manifests，包括 Deployments、Services、ConfigMaps 和 Secrets，遵循最佳实践和安全标准。在生成 Kubernetes YAML manifests、创建 K8s 资源或实现生产级 Kubernetes 配置时使用。
mode: on_demand
version: 1.1.0
maturity: experimental
---

# Kubernetes Manifest 生成器

创建生产就绪 Kubernetes manifests（包括 Deployments、Services、ConfigMaps、Secrets 和 PersistentVolumeClaims）的分步指导。

## 用途

此技能提供生成结构良好、安全且生产就绪的 Kubernetes manifests 的全面指导，遵循云原生最佳实践和 Kubernetes 约定。

## 何时使用此技能

在以下情况下使用此技能：

- 创建新的 Kubernetes Deployment manifests
- 定义用于网络连接的 Service 资源
- 生成用于配置管理的 ConfigMap 和 Secret 资源
- 为有状态工作负载创建 PersistentVolumeClaim manifests
- 遵循 Kubernetes 最佳实践和命名约定
- 实现资源限制、健康检查和安全上下文
- 设计用于多环境部署的 manifests

## 分步工作流

### 1. 收集需求

**了解工作负载：**

- 应用类型（无状态/有状态）
- 容器镜像和版本
- 环境变量和配置需求
- 存储需求
- 网络暴露需求（内部/外部）
- 资源需求（CPU、内存）
- 扩展需求
- 健康检查端点

**需要询问的问题：**

- 应用名称和用途是什么？
- 将使用什么容器镜像和标签？
- 应用是否需要持久存储？
- 应用暴露什么端口？
- 是否需要任何 secrets 或配置文件？
- CPU 和内存需求是什么？
- 应用是否需要外部暴露？

### 2. 创建 Deployment Manifest

**遵循此结构：**

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: <app-name>
  namespace: <namespace>
  labels:
    app: <app-name>
    version: <version>
spec:
  replicas: 3
  selector:
    matchLabels:
      app: <app-name>
  template:
    metadata:
      labels:
        app: <app-name>
        version: <version>
    spec:
      containers:
        - name: <container-name>
          image: <image>:<tag>
          ports:
            - containerPort: <port>
              name: http
          resources:
            requests:
              memory: "256Mi"
              cpu: "250m"
            limits:
              memory: "512Mi"
              cpu: "500m"
          livenessProbe:
            httpGet:
              path: /health
              port: http
            initialDelaySeconds: 30
            periodSeconds: 10
          readinessProbe:
            httpGet:
              path: /ready
              port: http
            initialDelaySeconds: 5
            periodSeconds: 5
          env:
            - name: ENV_VAR
              value: "value"
          envFrom:
            - configMapRef:
                name: <app-name>-config
            - secretRef:
                name: <app-name>-secret
```

**应用最佳实践：**

- 始终设置资源请求和限制
- 实现 liveness 和 readiness 探针
- 使用特定镜像标签（永远不要使用 `:latest`）
- 为非 root 用户应用安全上下文
- 使用标签进行组织和选择
- 根据可用性需求设置适当的副本数

**参考：** 查看 `references/deployment-spec.md` 了解详细的 Deployment 选项

### 3. 创建 Service Manifest

**选择适当的 Service 类型：**

**ClusterIP（仅内部）：**

```yaml
apiVersion: v1
kind: Service
metadata:
  name: <app-name>
  namespace: <namespace>
  labels:
    app: <app-name>
spec:
  type: ClusterIP
  selector:
    app: <app-name>
  ports:
    - name: http
      port: 80
      targetPort: 8080
      protocol: TCP
```

**LoadBalancer（外部访问）：**

```yaml
apiVersion: v1
kind: Service
metadata:
  name: <app-name>
  namespace: <namespace>
  labels:
    app: <app-name>
  annotations:
    service.beta.kubernetes.io/aws-load-balancer-type: nlb
spec:
  type: LoadBalancer
  selector:
    app: <app-name>
  ports:
    - name: http
      port: 80
      targetPort: 8080
      protocol: TCP
```

**参考：** 查看 `references/service-spec.md` 了解 Service 类型和网络

### 4. 创建 ConfigMap

**用于应用配置：**

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: <app-name>-config
  namespace: <namespace>
data:
  APP_MODE: production
  LOG_LEVEL: info
  DATABASE_HOST: db.example.com
  # 用于配置文件
  app.properties: |
    server.port=8080
    server.host=0.0.0.0
    logging.level=INFO
```

**最佳实践：**

- 仅将 ConfigMaps 用于非敏感数据
- 将相关配置组织在一起
- 使用有意义的键名称
- 考虑每个组件使用一个 ConfigMap
- 更改时版本化 ConfigMaps

**参考：** 查看 `assets/configmap-template.yaml` 了解示例

### 5. 创建 Secret

**用于敏感数据：**

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: <app-name>-secret
  namespace: <namespace>
type: Opaque
stringData:
  DATABASE_PASSWORD: "changeme"
  API_KEY: "secret-api-key"
  # 用于证书文件
  tls.crt: |
    -----BEGIN CERTIFICATE-----
    ...
    -----END CERTIFICATE-----
  tls.key: |
    <TLS 私钥占位：部署时经密钥管理注入，示例不入库>
```

**安全注意事项：**

- 切勿以明文形式将 secrets 提交到 Git
- 使用 Sealed Secrets、External Secrets Operator 或 Vault
- 定期轮换 secrets
- 使用 RBAC 限制 Secret 访问
- 考虑使用 Secret 类型：`kubernetes.io/tls` 用于 TLS secrets

### 6. 创建 PersistentVolumeClaim（如果需要）

**用于有状态应用：**

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: <app-name>-data
  namespace: <namespace>
spec:
  accessModes:
    - ReadWriteOnce
  storageClassName: gp3
  resources:
    requests:
      storage: 10Gi
```

**在 Deployment 中挂载：**

```yaml
spec:
  template:
    spec:
      containers:
        - name: app
          volumeMounts:
            - name: data
              mountPath: /var/lib/app
      volumes:
        - name: data
          persistentVolumeClaim:
            claimName: <app-name>-data
```

**存储注意事项：**

- 根据性能需求选择适当的 StorageClass
- 对单 Pod 访问使用 ReadWriteOnce
- 对多 Pod 共享存储使用 ReadWriteMany
- 考虑备份策略
- 设置适当的保留策略

### 7. 应用安全最佳实践

**向 Deployment 添加安全上下文：**

```yaml
spec:
  template:
    spec:
      securityContext:
        runAsNonRoot: true
        runAsUser: 1000
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

**安全检查清单：**

- [ ] 以非 root 用户运行
- [ ] 丢弃所有 capabilities
- [ ] 使用只读 root 文件系统
- [ ] 禁用特权提升
- [ ] 设置 seccomp 配置文件
- [ ] 使用 Pod Security Standards

### 8. 添加标签和注解

**标准标签（推荐）：**

```yaml
metadata:
  labels:
    app.kubernetes.io/name: <app-name>
    app.kubernetes.io/instance: <instance-name>
    app.kubernetes.io/version: "1.0.0"
    app.kubernetes.io/component: backend
    app.kubernetes.io/part-of: <system-name>
    app.kubernetes.io/managed-by: kubectl
```

**有用的注解：**

```yaml
metadata:
  annotations:
    description: "应用描述"
    contact: "team@example.com"
    prometheus.io/scrape: "true"
    prometheus.io/port: "9090"
    prometheus.io/path: "/metrics"
```

### 9. 组织多资源 Manifests

**文件组织选项：**

**选项 1：使用 `---` 分隔符的单个文件**

```yaml
# app-name.yaml
---
apiVersion: v1
kind: ConfigMap
...
---
apiVersion: v1
kind: Secret
...
---
apiVersion: apps/v1
kind: Deployment
...
---
apiVersion: v1
kind: Service
...
```

**选项 2：单独文件**

```
manifests/
├── configmap.yaml
├── secret.yaml
├── deployment.yaml
├── service.yaml
└── pvc.yaml
```

**选项 3：Kustomize 结构**

```
base/
├── kustomization.yaml
├── deployment.yaml
├── service.yaml
└── configmap.yaml
overlays/
├── dev/
│   └── kustomization.yaml
└── prod/
    └── kustomization.yaml
```

### 10. 验证和测试

**验证步骤：**

```bash
# Dry-run 验证
kubectl apply -f manifest.yaml --dry-run=client

# 服务器端验证
kubectl apply -f manifest.yaml --dry-run=server

# 使用 kubeval 验证
kubeval manifest.yaml

# 使用 kube-score 验证
kube-score score manifest.yaml

# 使用 kube-linter 检查
kube-linter lint manifest.yaml
```

**测试检查清单：**

- [ ] Manifest 通过 dry-run 验证
- [ ] 所有必填字段都存在
- [ ] 资源限制合理
- [ ] 健康检查已配置
- [ ] 安全上下文已设置
- [ ] 标签遵循约定
- [ ] 命名空间存在或已创建

## 常见模式

### 模式 1：简单无状态 Web 应用

**用例：** 标准 Web API 或微服务

**所需组件：**

- Deployment（3 个副本用于高可用）
- ClusterIP Service
- 用于配置的 ConfigMap
- 用于 API 密钥的 Secret
- HorizontalPodAutoscaler（可选）

**参考：** 查看 `assets/deployment-template.yaml`

### 模式 2：有状态数据库应用

**用例：** 数据库或持久存储应用

**所需组件：**

- StatefulSet（不是 Deployment）
- Headless Service
- PersistentVolumeClaim 模板
- 用于 DB 配置的 ConfigMap
- 用于凭据的 Secret

### 模式 3：后台作业或 Cron

**用例：** 计划任务或批处理

**所需组件：**

- CronJob 或 Job
- 用于作业参数的 ConfigMap
- 用于凭据的 Secret
- 带有 RBAC 的 ServiceAccount

### 模式 4：多容器 Pod

**用例：** 带有 sidecar 容器的应用

**所需组件：**

- 带有多个容器的 Deployment
- 容器之间的共享卷
- 用于设置的 Init 容器
- Service（如果需要）

## 模板

`assets/` 目录中提供以下模板：

- `deployment-template.yaml` - 带有最佳实践的标准 Deployment
- `service-template.yaml` - Service 配置（ClusterIP、LoadBalancer、NodePort）
- `configmap-template.yaml` - 具有不同数据类型的 ConfigMap 示例
- `secret-template.yaml` - Secret 示例（需要生成，不要提交）
- `pvc-template.yaml` - PersistentVolumeClaim 模板

## 参考文档

- `references/deployment-spec.md` - 详细的 Deployment 规范
- `references/service-spec.md` - Service 类型和网络详细信息

## 最佳实践总结

1. **始终设置资源请求和限制** - 防止资源耗尽
2. **实现健康检查** - 确保 Kubernetes 可以管理你的应用
3. **使用特定镜像标签** - 避免不可预测的部署
4. **应用安全上下文** - 以非 root 运行，丢弃 capabilities
5. **使用 ConfigMaps 和 Secrets** - 将配置与代码分离
6. **为所有内容添加标签** - 启用过滤和组织
7. **遵循命名约定** - 使用标准 Kubernetes 标签
8. **应用前验证** - 使用 dry-run 和验证工具
9. **版本化你的 manifests** - 保存在 Git 中进行版本控制
10. **使用注解进行文档说明** - 为其他开发人员添加上下文

## 故障排除

**Pod 未启动：**

- 检查镜像拉取错误：`kubectl describe pod <pod-name>`
- 验证资源可用性：`kubectl get nodes`
- 检查事件：`kubectl get events --sort-by='.lastTimestamp'`

**Service 不可访问：**

- 验证选择器是否匹配 Pod 标签：`kubectl get endpoints <service-name>`
- 检查 Service 类型和端口配置
- 从集群内测试：`kubectl run debug --rm -it --image=busybox -- sh`

**ConfigMap/Secret 未加载：**

- 验证 Deployment 中的名称是否匹配
- 检查命名空间
- 确保资源存在：`kubectl get configmap,secret`

## 后续步骤

创建 manifests 后：

1. 存储在 Git 仓库中
2. 为部署设置 CI/CD 流水线
3. 考虑使用 Helm 或 Kustomize 进行模板化
4. 使用 ArgoCD 或 Flux 实现 GitOps
5. 添加监控和可观测性

## 相关技能

- `helm-chart-scaffolding` - 用于模板化和打包
- `gitops-workflow` - 用于自动化部署
- `k8s-security-policies` - 用于高级安全配置
