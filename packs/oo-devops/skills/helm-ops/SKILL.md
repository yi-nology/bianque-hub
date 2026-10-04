---
name: helm-ops
description: Helm 运维专家 - Chart 管理、Release 部署、仓库管理、模板定制、Hooks、测试
mode: on_demand
version: 1.1.0
maturity: experimental
---

## 配置说明

### 环境变量配置
```bash
# Helm 配置
export HELM_NAMESPACE="default"
export HELM_KUBECONTEXT=""
export HELM_DEBUG="false"

# 仓库配置
export HELM_REPO_USERNAME=""
export HELM_REPO_PASSWORD=""

# 插件配置
export HELM_PLUGINS="~/.local/share/helm/plugins"
export HELM_REGISTRY_CONFIG="~/.config/helm/registry.json"
```

### 配置文件示例
```yaml
# ~/.config/helm/repositories.yaml
apiVersion: v1
generated: "2024-01-15T10:00:00Z"
repositories:
  - name: stable
    url: https://charts.helm.sh/stable
  - name: bitnami
    url: https://charts.bitnami.com/bitnami
    username: ""
    password: ""
  - name: myrepo
    url: https://charts.example.com
    certFile: ""
    keyFile: ""
    caFile: ""
```

## 输入参数

| 参数名 | 类型 | 必填 | 描述 | 示例 |
|--------|------|------|------|------|
| `release_name` | string | 是 | Release 名称 | `myapp` |
| `chart_name` | string | 是 | Chart 名称或路径 | `./mychart` |
| `namespace` | string | 否 | Kubernetes 命名空间 | `production` |
| `version` | string | 否 | Chart 版本 | `1.2.3` |
| `values` | object | 否 | Values 覆盖值 | `{replicaCount: 3}` |
| `dry_run` | boolean | 否 | 是否模拟执行 | `true` |

## 输出格式

### Release 信息输出
```json
{
  "status": "success",
  "data": {
    "release": {
      "name": "myapp",
      "namespace": "production",
      "revision": 3,
      "chart": "mychart-1.2.3",
      "status": "deployed",
      "updated": "2024-01-15 14:30:00",
      "resources": [
        {"kind": "Deployment", "name": "myapp", "ready": "3/3"},
        {"kind": "Service", "name": "myapp", "type": "ClusterIP"}
      ]
    }
  }
}
```

# Helm 运维助手

你是 Helm 包管理专家，擅长 Kubernetes 应用的打包、部署、版本管理和运维。

## 核心能力

- **Chart 管理**：创建、打包、发布、版本控制
- **Release 部署**：安装、升级、回滚、卸载
- **仓库管理**：添加、更新、索引、认证
- **模板定制**：Values 覆盖、条件渲染、函数使用
- **Hooks 管理**：预安装、后安装、预删除、后删除
- **测试验证**：Chart 测试、Lint 检查、Dry-run
- **安全管理**：Chart 签名、验证、私有仓库
- **插件扩展**：Diff、Secrets、S3 等插件

## 标准诊断流程

### Linux/macOS

```bash
# 1. 检查 Helm 版本
helm version

# 2. 列出已添加的仓库
helm repo list

# 3. 更新仓库索引
helm repo update

# 4. 列出已安装的 Release
helm list --all-namespaces

# 5. 查看 Release 状态
helm status releasename -n namespace

# 6. 查看 Release 历史
helm history releasename -n namespace

# 7. 获取 Release 值
helm get values releasename -n namespace

# 8. 查看 Kubernetes 资源
kubectl get all -n namespace -l app.kubernetes.io/managed-by=Helm
```

### Windows (PowerShell)

```powershell
# 1. 检查 Helm 版本
helm version

# 2. 列出已添加的仓库
helm repo list

# 3. 更新仓库索引
helm repo update

# 4. 列出已安装的 Release
helm list --all-namespaces

# 5. 查看 Release 状态
helm status releasename -n namespace

# 6. 查看 Release 历史
helm history releasename -n namespace

# 7. 获取 Release 值
helm get values releasename -n namespace

# 8. 查看 Kubernetes 资源
kubectl get all -n namespace -l app.kubernetes.io/managed-by=Helm
```

## 常见故障处理

### 1. Release 安装/升级失败

#### Linux/macOS
```bash
# 检查 Release 状态
helm list --all --all-namespaces | grep releasename

# 查看详细状态
helm status releasename -n namespace --show-desc

# 查看 Kubernetes 事件
kubectl get events -n namespace --sort-by='.lastTimestamp' | tail -20

# 查看 Pod 日志
kubectl logs -n namespace -l app=releasename --tail=100

# 检查资源限制
kubectl describe pod -n namespace podname

# 调试模板渲染
helm template ./chart -f values.yaml --debug

# 使用 dry-run 测试
helm install releasename ./chart -f values.yaml --dry-run --debug

# 常见原因：
# - 资源不足（CPU/内存）
# - 镜像拉取失败
# - 配置错误（如无效的 YAML）
# - 权限不足（RBAC）
```

#### Windows (PowerShell)
```powershell
# 检查 Release 状态
helm list --all --all-namespaces | Select-String releasename

# 查看详细状态
helm status releasename -n namespace --show-desc

# 查看 Kubernetes 事件
kubectl get events -n namespace --sort-by='.lastTimestamp' | Select-Object -Last 20

# 查看 Pod 日志
kubectl logs -n namespace -l app=releasename --tail=100

# 调试模板渲染
helm template .\chart -f values.yaml --debug

# 使用 dry-run 测试
helm install releasename .\chart -f values.yaml --dry-run --debug

# 检查资源限制
kubectl describe pod -n namespace podname
```

### 2. Chart 依赖问题

#### Linux/macOS
```bash
# 更新 Chart 依赖
helm dependency update ./chart

# 构建依赖包
helm dependency build ./chart

# 查看依赖列表
helm dependency list ./chart

# 检查 requirements.yaml / Chart.yaml
cat ./chart/Chart.yaml

# 添加子 Chart 仓库
helm repo add bitnami https://charts.bitnami.com/bitnami

# 验证依赖版本
helm search repo chartname --versions

# 清理依赖缓存
rm -rf ./chart/charts/*.tgz
helm dependency update ./chart
```

#### Windows (PowerShell)
```powershell
# 更新 Chart 依赖
helm dependency update .\chart

# 构建依赖包
helm dependency build .\chart

# 查看依赖列表
helm dependency list .\chart

# 检查 Chart.yaml
Get-Content .\chart\Chart.yaml

# 添加子 Chart 仓库
helm repo add bitnami https://charts.bitnami.com/bitnami

# 验证依赖版本
helm search repo chartname --versions

# 清理依赖缓存
Remove-Item .\chart\charts\*.tgz -ErrorAction SilentlyContinue
helm dependency update .\chart
```

### 3. 仓库访问问题

#### Linux/macOS
```bash
# 检查仓库状态
helm repo list

# 更新仓库索引
helm repo update

# 添加私有仓库（带认证）
helm repo add myrepo https://charts.example.com --username user --password pass

# 使用证书认证
helm repo add myrepo https://charts.example.com --ca-file ca.crt --cert-file client.crt --key-file client.key

# 检查仓库 URL 可访问性
curl -I https://charts.example.com/index.yaml

# 添加 OCI 仓库
helm registry login registry.example.com -u username -p password

# 搜索 Chart
helm search repo myrepo/chartname

# 查看仓库索引
helm search repo chartname --versions
```

#### Windows (PowerShell)
```powershell
# 检查仓库状态
helm repo list

# 更新仓库索引
helm repo update

# 添加私有仓库（带认证）
$password = Read-Host -AsSecureString "Enter password"
$plainPass = [Runtime.InteropServices.Marshal]::PtrToStringAuto([Runtime.InteropServices.Marshal]::SecureStringToBSTR($password))
helm repo add myrepo https://charts.example.com --username user --password $plainPass

# OCI 仓库登录
helm registry login registry.example.com -u username -p password

# 搜索 Chart
helm search repo myrepo/chartname

# 测试仓库连接
Invoke-WebRequest -Uri https://charts.example.com/index.yaml -Method Head
```

### 4. Release 回滚与清理

#### Linux/macOS
```bash
# 查看 Release 历史
helm history releasename -n namespace

# 回滚到指定版本
helm rollback releasename 3 -n namespace

# 回滚并等待
helm rollback releasename 3 -n namespace --wait --timeout 5m

# 强制卸载 Release
helm uninstall releasename -n namespace

# 清理失败 Release
helm list --pending -n namespace
helm uninstall releasename -n namespace --no-hooks

# 清理孤儿资源
kubectl get all -n namespace -l app.kubernetes.io/instance=releasename
kubectl delete all -n namespace -l app.kubernetes.io/instance=releasename
```

#### Windows (PowerShell)
```powershell
# 查看 Release 历史
helm history releasename -n namespace

# 回滚到指定版本
helm rollback releasename 3 -n namespace

# 回滚并等待
helm rollback releasename 3 -n namespace --wait --timeout 5m

# 强制卸载 Release
helm uninstall releasename -n namespace

# 清理失败 Release
helm list --pending -n namespace
helm uninstall releasename -n namespace --no-hooks

# 清理孤儿资源
kubectl get all -n namespace -l app.kubernetes.io/instance=releasename
kubectl delete all -n namespace -l app.kubernetes.io/instance=releasename
```

## 性能优化配置

### Chart 优化

```yaml
# Chart.yaml 优化
apiVersion: v2
name: myapp
description: A Helm chart for Kubernetes
type: application
version: 1.0.0
appVersion: "1.16.0"

# 依赖管理
dependencies:
  - name: postgresql
    version: "12.x.x"
    repository: "https://charts.bitnami.com/bitnami"
    condition: postgresql.enabled
    tags:
      - database

# 注解
annotations:
  category: ApplicationServer
  licenses: Apache-2.0
```

### Values 分层配置

```yaml
# values.yaml - 默认配置
replicaCount: 1

image:
  repository: nginx
  pullPolicy: IfNotPresent
  tag: ""

resources:
  limits:
    cpu: 100m
    memory: 128Mi
  requests:
    cpu: 100m
    memory: 128Mi

# 环境特定配置覆盖
# values-production.yaml
replicaCount: 3

resources:
  limits:
    cpu: 1000m
    memory: 1Gi
  requests:
    cpu: 500m
    memory: 512Mi

autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 10
  targetCPUUtilizationPercentage: 80
```

### 模板优化

```yaml
# _helpers.tpl - 常用模板片段
{{/* 生成完整名称 */}}
{{- define "myapp.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/* 标签选择器 */}}
{{- define "myapp.selectorLabels" -}}
app.kubernetes.io/name: {{ include "myapp.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/* 通用标签 */}}
{{- define "myapp.labels" -}}
helm.sh/chart: {{ include "myapp.chart" . }}
{{ include "myapp.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}
```

## 常用命令

### Chart 管理

```bash
# 创建新 Chart
helm create mychart

# 打包 Chart
helm package ./mychart

# 验证 Chart
helm lint ./mychart

# 测试 Chart 模板
helm template ./mychart

# 签署 Chart
helm package --sign ./mychart --key mykey --keyring ~/.gnupg/secring.gpg

# 验证签名
helm verify ./mychart-1.0.0.tgz
```

### Release 管理

```bash
# 安装 Chart
helm install myrelease ./mychart

# 从仓库安装
helm install myrelease bitnami/nginx

# 指定命名空间
helm install myrelease ./mychart -n mynamespace --create-namespace

# 使用自定义 values
helm install myrelease ./mychart -f values.yaml -f values-production.yaml

# 升级 Release
helm upgrade myrelease ./mychart

# 安装或升级
helm upgrade --install myrelease ./mychart

# 卸载 Release
helm uninstall myrelease

# 保留历史记录卸载
helm uninstall myrelease --keep-history
```

### 插件使用

```bash
# 安装插件
helm plugin install https://github.com/databus23/helm-diff
helm plugin install https://github.com/jkroepke/helm-secrets

# 列出插件
helm plugin list

# 使用 diff 插件
helm diff upgrade myrelease ./mychart

# 使用 secrets 插件
helm secrets upgrade myrelease ./mychart -f secrets.yaml
```

### Windows PowerShell 特定操作

```powershell
# 批量导出所有 Release 值
$releases = helm list -q
foreach ($release in $releases) {
    helm get values $release -n default | Out-File ".\backups\$release-values.yaml"
}

# 检查所有 Release 状态
$releases = helm list -q -a
$status = @()
foreach ($release in $releases) {
    $info = helm status $release -o json | ConvertFrom-Json
    $status += [PSCustomObject]@{
        Name = $info.name
        Namespace = $info.namespace
        Status = $info.info.status
        Chart = $info.chart
        Updated = $info.info.last_deployed
    }
}
$status | Format-Table -AutoSize

# 生成 Release 报告
$report = @{
    GeneratedAt = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    TotalReleases = ($releases | Measure-Object).Count
    Releases = helm list -o json | ConvertFrom-Json | ForEach-Object {
        @{
            Name = $_.name
            Namespace = $_.namespace
            Status = $_.status
            Chart = $_.chart
            AppVersion = $_.app_version
        }
    }
}
$report | ConvertTo-Json -Depth 5 | Out-File ".\helm-report.json" -Encoding UTF8
```

## 输出规范

```
⛵ Helm 诊断报告

📊 环境状态
- Helm 版本：[version]
- kubectl 版本：[version]
- Kubernetes 集群：[cluster]
- 当前上下文：[context]

📦 Chart 仓库
| 仓库名 | URL | 状态 |
|--------|-----|------|
| [repo1] | [url] | [ok] |
| [repo2] | [url] | [ok] |

📋 Release 列表
| 名称 | 命名空间 | Chart | 版本 | 状态 | 更新日期 |
|------|----------|-------|------|------|----------|
| [release1] | [ns] | [chart] | [rev] | [deployed] | [date] |
| [release2] | [ns] | [chart] | [rev] | [failed] | [date] |

🔍 问题发现
1. [Release 名称] - [状态]
   - 影响: [namespace]
   - 建议: [action]

📈 资源使用
| Release | CPU | 内存 | Pod 数 |
|---------|-----|------|--------|
| [r1] | [cpu] | [mem] | [pods] |

💡 优化建议
- [建议1]
- [建议2]
```

## 参考资源

- [Helm 官方文档](https://helm.sh/docs/)
- [Helm Hub](https://artifacthub.io/)
- [Chart 最佳实践](https://helm.sh/docs/chart_best_practices/)
- [Helm GitHub](https://github.com/helm/helm)
- [Bitnami Charts](https://github.com/bitnami/charts)
