---
name: k8s-helm
description: 管理 Helm charts、releases 和 repositories。用于 Helm 安装、升级、回滚、chart 开发和 release 管理。
mode: on_demand
version: 1.1.1
maturity: experimental
---

# Helm Chart 管理

使用 kubectl-mcp-server 的 16 个 Helm 工具进行全面的 Helm v3 操作。

## 何时应用

在以下情况下使用此技能：
- 用户提到："helm"、"chart"、"release"、"values"、"repository"
- 操作：安装 charts、升级 releases、回滚
- 关键词："package"、"template"、"lint"、"repo add"

## 优先级规则

| 优先级 | 规则 | 影响 | 工具 |
|----------|------|--------|-------|
| 1 | 安装前先 template (dry run) | 严重 | `template_helm_chart` |
| 2 | 先检查现有的 releases | 严重 | `list_helm_releases` |
| 3 | 打包前先 lint charts | 高 | `lint_helm_chart` |
| 4 | 升级前注意 revision | 高 | `get_helm_history` |
| 5 | 升级后验证 values | 中 | `get_helm_values` |
| 6 | 搜索前先更新 repos | 低 | `update_helm_repos` |

## 快速参考

| 任务 | 工具 | 示例 |
|------|------|---------|
| 安装 Chart | `install_helm_chart` | `install_helm_chart(name, chart, namespace)` |
| 升级 Release | `upgrade_helm_release` | `upgrade_helm_release(name, chart, namespace, values)` |
| 回滚 | `rollback_helm_release` | `rollback_helm_release(name, namespace, revision)` |
| 列出 releases | `list_helm_releases` | `list_helm_releases(namespace)` |
| 获取 values | `get_helm_values` | `get_helm_values(name, namespace)` |
| Template (dry run) | `template_helm_chart` | `template_helm_chart(name, chart, namespace)` |

## 安装 Chart

```python
install_helm_chart(
    name="my-release",
    chart="bitnami/nginx",
    namespace="web",
    values={"replicaCount": 3, "service.type": "LoadBalancer"}
)
```

## 升级 Release

```python
upgrade_helm_release(
    name="my-release",
    chart="bitnami/nginx",
    namespace="web",
    values={"replicaCount": 5}
)
```

## 回滚 Release

```python
rollback_helm_release(
    name="my-release",
    namespace="web",
    revision=1
)
```

## 卸载 Release

```python
uninstall_helm_chart(name="my-release", namespace="web")
```

## Release 管理

### 列出 Releases

```python
list_helm_releases(namespace="web")

list_helm_releases()
```

### 获取 Release 详情

```python
get_helm_release(name="my-release", namespace="web")
```

### Release 历史

```python
get_helm_history(name="my-release", namespace="web")
```

### 获取 Release Values

```python
get_helm_values(name="my-release", namespace="web")
```

### 获取 Release Manifest

```python
get_helm_manifest(name="my-release", namespace="web")
```

## 仓库管理

### 添加仓库

```python
add_helm_repo(name="bitnami", url="https://charts.bitnami.com/bitnami")
```

### 列出仓库

```python
list_helm_repos()
```

### 更新仓库

```python
update_helm_repos()
```

### 搜索 Charts

```python
search_helm_charts(keyword="nginx")

search_helm_charts(keyword="postgres", repo="bitnami")
```

## Chart 开发

### Template Chart (Dry Run)

```python
template_helm_chart(
    name="my-release",
    chart="./my-chart",
    namespace="test",
    values={"key": "value"}
)
```

### Lint Chart

```python
lint_helm_chart(chart="./my-chart")
```

### 打包 Chart

```python
package_helm_chart(chart="./my-chart", destination="./packages")
```

## 常见工作流

### 新应用部署

```python
add_helm_repo(name="bitnami", url="...")

search_helm_charts(keyword="postgresql")

template_helm_chart(...)

install_helm_chart(...)

get_helm_release(...)
```

### 安全升级（带回滚保护）

```python
get_helm_history(name, namespace)

upgrade_helm_release(name, chart, namespace, values)

rollback_helm_release(name, namespace, revision)
```

### 多环境部署

```python
install_helm_chart(
    name="app",
    chart="./charts/app",
    namespace="dev",
    values={"replicas": 1},
    context="development"
)

install_helm_chart(
    name="app",
    chart="./charts/app",
    namespace="staging",
    values={"replicas": 2},
    context="staging"
)

install_helm_chart(
    name="app",
    chart="./charts/app",
    namespace="prod",
    values={"replicas": 5},
    context="production"
)
```

## Chart 结构

查看 [references/CHART-STRUCTURE.md](references/CHART-STRUCTURE.md) 了解组织 Helm charts 的最佳实践。

## 故障排除

查看 [TROUBLESHOOTING.md](TROUBLESHOOTING.md) 了解常见问题。

### Release 卡在 Pending 状态

```python
get_helm_release(name, namespace)

get_pods(namespace, label_selector="app.kubernetes.io/instance=<release>")
```

### 安装失败

```python
get_helm_history(name, namespace)

get_events(namespace)

uninstall_helm_chart(name, namespace)
```

### Values 未应用

```python
get_helm_values(name, namespace)

template_helm_chart(...)

upgrade_helm_release(...)
```

## 脚本

查看 [scripts/lint-chart.sh](scripts/lint-chart.sh) 了解自动化 chart 验证。

## 最佳实践

1. **始终先 template**
   ```python
   template_helm_chart(name, chart, namespace, values)
   ```

2. **使用语义化版本**
   ```python
   install_helm_chart(..., version="1.2.3")
   ```

3. **将 Values 存储在 Git 中**
   - `values-dev.yaml`
   - `values-staging.yaml`
   - `values-prod.yaml`

4. **命名空间隔离**
   - 每个 release 一个命名空间
   - 更容易清理和 RBAC

## 先决条件

- **Helm CLI**: 所有 Helm 操作的必需工具
  ```bash
  # 供应链纪律：不直灌管道安装（下载即执行会盲跑远端脚本）。两步式——下载后校验
  # checksum 再执行；或走包管理器（apt install helm / dnf install helm）。
  curl -fsSL -o /tmp/get-helm-3 https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3
  sha256sum /tmp/get-helm-3   # 与 helm 发布页 get-helm-3.sha256sum 比对一致后再 bash 执行
  ```

## 相关技能

- [k8s-deploy](../k8s-deploy/SKILL.md) - 部署策略
- [k8s-gitops](../k8s-gitops/SKILL.md) - GitOps Helm releases
