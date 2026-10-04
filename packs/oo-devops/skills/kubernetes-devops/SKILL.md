---
name: kubernetes-devops
description: 生产级 Kubernetes 清单(manifest)生成技能。支持 Deployment、StatefulSet、CronJob、Service、Ingress、ConfigMap、Secret 和 PVC，包含安全上下文(security context)、健康检查(health check)和资源管理(resource management)。触发词：Kubernetes 清单(manifest)、Deployment、Service、Ingress、ConfigMap、DevOps
mode: on_demand
version: 1.1.0
maturity: experimental
---

# Kubernetes

生产就绪的 Kubernetes 清单生成，涵盖 Deployment、StatefulSet、CronJob、Service、Ingress、ConfigMap、Secret 和 PVC，包含安全上下文、健康检查和资源管理。


## 安装

### OpenClaw / Moltbot / Clawbot

```bash
npx clawhub@latest install kubernetes
```


## 何时使用

| 场景 | 示例 |
|----------|---------|
| 创建 Deployment 清单 | 新的微服务需要 Deployment + Service |
| 定义网络资源 | ClusterIP、LoadBalancer、包含 TLS 的 Ingress |
| 管理配置 | ConfigMap 用于应用配置，Secret 用于凭证 |
| 有状态工作负载 | 使用 StatefulSet + PVC 的数据库 |
| 定时任务 | 用于批处理的 CronJob |
| 多环境设置 | 用于 dev/staging/prod 的 Kustomize overlay |

## 工作负载选择

| 工作负载类型 | 资源 | 何时使用 |
|---------------|----------|-------------|
| 无状态应用 | Deployment | Web 服务器、API、微服务 |
| 有状态应用 | StatefulSet | 数据库、消息队列、缓存 |
| 一次性任务 | Job | 迁移、数据导入 |
| 定时任务 | CronJob | 备份、报告、清理 |
| 每节点代理 | DaemonSet | 日志收集器、监控代理 |

## Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
  namespace: production
  labels:
    app.kubernetes.io/name: my-app
    app.kubernetes.io/version: "1.0.0"
    app.kubernetes.io/component: backend
spec:
  replicas: 3
  selector:
    matchLabels:
      app.kubernetes.io/name: my-app
  template:
    metadata:
      labels:
        app.kubernetes.io/name: my-app
        app.kubernetes.io/version: "1.0.0"
    spec:
      securityContext:
        runAsNonRoot: true
        runAsUser: 1000
        fsGroup: 1000
        seccompProfile:
          type: RuntimeDefault
      containers:
        - name: my-app
          image: registry.example.com/my-app:1.0.0
          ports:
            - containerPort: 8080
              name: http
          resources:
            requests:
              cpu: 250m
              memory: 256Mi
            limits:
              cpu: 500m
              memory: 512Mi
          securityContext:
            allowPrivilegeEscalation: false
            readOnlyRootFilesystem: true
            capabilities:
              drop: [ALL]
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
            - name: LOG_LEVEL
              valueFrom:
                configMapKeyRef:
                  name: my-app-config
                  key: LOG_LEVEL
            - name: DB_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: my-app-secret
                  key: DATABASE_PASSWORD
```

## Services

### ClusterIP（内部）

```yaml
apiVersion: v1
kind: Service
metadata:
  name: my-app
  namespace: production
spec:
  type: ClusterIP
  selector:
    app.kubernetes.io/name: my-app
  ports:
    - name: http
      port: 80
      targetPort: 8080
      protocol: TCP
```

### LoadBalancer（外部）

```yaml
apiVersion: v1
kind: Service
metadata:
  name: my-app-lb
  namespace: production
  annotations:
    service.beta.kubernetes.io/aws-load-balancer-type: nlb
spec:
  type: LoadBalancer
  selector:
    app.kubernetes.io/name: my-app
  ports:
    - name: http
      port: 80
      targetPort: 8080
```

### Service 类型快速参考

| 类型 | 范围 | 用例 |
|------|-------|----------|
| ClusterIP | 集群内部 | 服务间通信 |
| NodePort | 通过节点 IP 外部访问 | 开发/测试、本地部署 |
| LoadBalancer | 通过云负载均衡器外部访问 | 生产外部访问 |
| ExternalName | DNS 别名 | 映射到外部服务 |

## Ingress

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: my-app
  namespace: production
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/rate-limit: "100"
spec:
  ingressClassName: nginx
  tls:
    - hosts: [app.example.com]
      secretName: app-tls
  rules:
    - host: app.example.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: my-app
                port:
                  number: 80
```

## ConfigMap 和 Secret

### ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-app-config
  namespace: production
data:
  LOG_LEVEL: info
  APP_MODE: production
  DATABASE_HOST: db.internal.svc.cluster.local
  app.properties: |
    server.port=8080
    server.host=0.0.0.0
```

### Secret

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: my-app-secret
  namespace: production
type: Opaque
stringData:
  DATABASE_PASSWORD: "changeme"
  API_KEY: "secret-api-key"
```

> **重要：** 切勿将明文 Secret 提交到 Git。生产环境请使用 Sealed Secrets、External Secrets Operator 或 Vault。

## 持久化存储

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: my-app-data
  namespace: production
spec:
  accessModes: [ReadWriteOnce]
  storageClassName: gp3
  resources:
    requests:
      storage: 10Gi
```

挂载到容器：

```yaml
containers:
  - name: app
    volumeMounts:
      - name: data
        mountPath: /var/lib/app
volumes:
  - name: data
    persistentVolumeClaim:
      claimName: my-app-data
```

| 访问模式 | 缩写 | 用例 |
|-------------|-------------|----------|
| ReadWriteOnce | RWO | 单 Pod 数据库 |
| ReadOnlyMany | ROX | 共享配置/静态资源 |
| ReadWriteMany | RWX | 多 Pod 共享存储 |

## 安全上下文

### Pod 级别

```yaml
spec:
  securityContext:
    runAsNonRoot: true
    runAsUser: 1000
    fsGroup: 1000
    seccompProfile:
      type: RuntimeDefault
```

### 容器级别

```yaml
securityContext:
  allowPrivilegeEscalation: false
  readOnlyRootFilesystem: true
  capabilities:
    drop: [ALL]
```

### 安全检查清单

| 检查 | 状态 |
|-------|--------|
| `runAsNonRoot: true` | 必需 |
| `allowPrivilegeEscalation: false` | 必需 |
| `readOnlyRootFilesystem: true` | 推荐 |
| `capabilities.drop: [ALL]` | 必需 |
| `seccompProfile: RuntimeDefault` | 推荐 |
| 特定镜像标签（永远不要使用 `:latest`） | 必需 |
| 设置资源请求和限制 | 必需 |

## 标准标签

```yaml
metadata:
  labels:
    app.kubernetes.io/name: my-app
    app.kubernetes.io/instance: my-app-prod
    app.kubernetes.io/version: "1.0.0"
    app.kubernetes.io/component: backend
    app.kubernetes.io/part-of: my-system
    app.kubernetes.io/managed-by: kubectl
```

## 清单组织

### 选项 1 — 单独文件

```
manifests/
├── configmap.yaml
├── secret.yaml
├── deployment.yaml
├── service.yaml
└── pvc.yaml
```

### 选项 2 — Kustomize

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
    ├── kustomization.yaml
    └── resource-patch.yaml
```

## 验证

```bash
# 客户端 dry run
kubectl apply -f manifest.yaml --dry-run=client

# 服务端验证
kubectl apply -f manifest.yaml --dry-run=server

# 使用 kube-score 检查
kube-score score manifest.yaml

# 使用 kube-linter 检查
kube-linter lint manifest.yaml
```

## 故障排查快速参考

| 问题 | 诊断 | 修复 |
|---------|-----------|-----|
| Pod 卡在 `Pending` | `kubectl describe pod` — 检查事件 | 修复资源请求、节点容量、PVC 绑定 |
| `ImagePullBackOff` | 镜像名称/标签错误或缺少拉取密钥 | 验证镜像存在，添加 `imagePullSecrets` |
| `CrashLoopBackOff` | 应用启动时崩溃 | 检查日志：`kubectl logs <pod> --previous` |
| Service 不可达 | 选择器不匹配 | 验证 `kubectl get endpoints <svc>` 非空 |
| ConfigMap 未加载 | 名称不匹配或错误的命名空间 | 检查名称匹配且命名空间正确 |
| Readiness 探针失败 | 错误的路径或端口 | 验证健康端点在容器内可用 |
| OOMKilled | 内存限制过低 | 增加 `resources.limits.memory` |

## 永远不要做

| 反模式 | 原因 | 替代方案 |
|-------------|-----|------------|
| 使用 `:latest` 镜像标签 | 不可复现的部署 | 固定精确版本：`image:1.2.3` |
| 跳过资源限制 | Pod 可能耗尽节点资源 | 始终设置 `requests` 和 `limits` |
| 以 root 运行 | 容器逃逸 = 完全访问主机 | 设置 `runAsNonRoot: true` + `USER` |
| 提交明文 Secret | 凭证将永远留在 Git 历史记录中 | 使用 Sealed Secrets / External Secrets / Vault |
| 跳过健康检查 | K8s 无法检测不健康的 Pod | 始终配置 liveness + readiness 探针 |
| 省略标签 | 无法过滤、选择或组织 | 使用标准 `app.kubernetes.io/*` 标签 |
| 生产环境单副本 | 更新期间零可用性 | 高可用性最少使用 `replicas: 3` |
| 在容器中硬编码配置 | 配置更改需要重新构建 | 使用 ConfigMap 和 Secret |

## 资源和参考

### 资源（模板）

| 模板 | 描述 |
|----------|-------------|
| [assets/deployment-template.yaml](assets/deployment-template.yaml) | 包含安全和探针的生产环境 Deployment |
| [assets/service-template.yaml](assets/service-template.yaml) | ClusterIP、LoadBalancer、NodePort 示例 |
| [assets/configmap-template.yaml](assets/configmap-template.yaml) | 包含数据类型的 ConfigMap |
| [assets/statefulset-template.yaml](assets/statefulset-template.yaml) | 包含 Headless Service + PVC 的 StatefulSet |
| [assets/cronjob-template.yaml](assets/cronjob-template.yaml) | 包含并发控制和历史记录的 CronJob |
| [assets/ingress-template.yaml](assets/ingress-template.yaml) | 包含 TLS、速率限制、CORS 的 Ingress |

### 参考

| 参考 | 描述 |
|-----------|-------------|
| [references/deployment-spec.md](references/deployment-spec.md) | 详细的 Deployment 规范 |
| [references/service-spec.md](references/service-spec.md) | Service 类型和网络详情 |
