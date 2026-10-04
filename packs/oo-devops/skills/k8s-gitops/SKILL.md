---
name: k8s-gitops
description: GitOps 工作流管理专家 - 基于 Flux/ArgoCD 实现状态同步、应用管理与故障排查
mode: on_demand
version: 1.1.0
maturity: experimental
---

# Kubernetes GitOps

GitOps workflows using Flux and ArgoCD with kubectl-mcp-server tools.

## 何时应用

在以下情况下使用此技能:
- 用户 mentions: "Flux", "ArgoCD", "GitOps", "sync", "reconcile"
- Operations: checking 同步状态, triggering reconciliation, drift detection
- Keywords: "out of sync", "deploy from git", "continuous delivery"

## Priority Rules

| Priority | 规则 | Impact | Tools |
|----------|------|--------|-------|
| 1 | 检查 source readiness 之前 故障排除 | 严重 | `flux_sources_list_tool` |
| 2 | 验证 同步状态 之前 deployments | 高 | `argocd_app_get_tool` |
| 3 | Reconcile 之后 git changes | 中 | `flux_reconcile_tool` |
| 4 | Suspend 之前 manual changes | 低 | `flux_suspend_tool` |

## 快速参考

| Task | Tool | Example |
|------|------|---------|
| List Flux kustomizations | `flux_kustomizations_list_tool` | `flux_kustomizations_list_tool(namespace)` |
| Reconcile Flux | `flux_reconcile_tool` | `flux_reconcile_tool(kind, name, namespace)` |
| List ArgoCD apps | `argocd_apps_list_tool` | `argocd_apps_list_tool(namespace)` |
| Sync ArgoCD | `argocd_sync_tool` | `argocd_sync_tool(name, namespace)` |

## Flux CD

### 检查 Flux 状态

```python
flux_kustomizations_list_tool(namespace="flux-system")
flux_helmreleases_list_tool(namespace)
flux_sources_list_tool(namespace="flux-system")
```

### Reconcile Resources

```python
flux_reconcile_tool(
    kind="kustomization",
    name="my-app",
    namespace="flux-system"
)

flux_reconcile_tool(
    kind="helmrelease",
    name="my-chart",
    namespace="default"
)
```

### Suspend/Resume

```python
flux_suspend_tool(kind="kustomization", name="my-app", namespace="flux-system")

flux_resume_tool(kind="kustomization", name="my-app", namespace="flux-system")
```

See [FLUX.md](FLUX.md) for detailed Flux workflows.

## ArgoCD

### List Applications

```python
argocd_apps_list_tool(namespace="argocd")
```

### Get App 状态

```python
argocd_app_get_tool(name="my-app", namespace="argocd")
```

### Sync 应用

```python
argocd_sync_tool(name="my-app", namespace="argocd")
```

### Refresh App

```python
argocd_refresh_tool(name="my-app", namespace="argocd")
```

See [ARGOCD.md](ARGOCD.md) for detailed ArgoCD workflows.

## GitOps 故障排除

### Flux Not Syncing

| Symptom | 检查 | Resolution |
|---------|-------|------------|
| Source not ready | `flux_sources_list_tool()` | 检查 git credentials |
| Kustomization 失败 | `flux_kustomizations_list_tool()` | 检查 manifest errors |
| HelmRelease 失败 | `flux_helmreleases_list_tool()` | 检查 values, chart version |

### ArgoCD Out of Sync

| Symptom | 检查 | Resolution |
|---------|-------|------------|
| OutOfSync | `argocd_app_get_tool()` | Manual sync or 检查 auto-sync |
| Degraded | 检查 health 状态 | Fix unhealthy resources |
| Unknown | Refresh app | `argocd_refresh_tool()` |

## 环境 Promotion

### With Flux Kustomizations

```python
flux_reconcile_tool(kind="kustomization", name="staging", namespace="flux-system")

flux_reconcile_tool(kind="kustomization", name="production", namespace="flux-system")
```

### With ArgoCD

```python
argocd_sync_tool(name="app-staging", namespace="argocd")

argocd_app_get_tool(name="app-staging", namespace="argocd")

argocd_sync_tool(name="app-production", namespace="argocd")
```

## Multi-集群 GitOps

管理 GitOps across clusters:

```python
flux_kustomizations_list_tool(namespace="flux-system", context="cluster-1")
flux_kustomizations_list_tool(namespace="flux-system", context="cluster-2")

flux_reconcile_tool(
    kind="kustomization",
    name="apps",
    namespace="flux-system",
    context="production-cluster"
)
```

## Drift Detection

Compare live state with desired:

```python
argocd_app_get_tool(name="my-app", namespace="argocd")

flux_kustomizations_list_tool(namespace="flux-system")
```

## 先决条件

- **Flux**: 必需 for Flux tools
  ```bash
  flux install
  ```
- **ArgoCD**: 必需 for ArgoCD tools
  ```bash
  kubectl create namespace argocd
  kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
  ```

## 相关 Skills

- [k8s-deploy](../k8s-deploy/SKILL.md) - Standard deployments
- [k8s-helm](../k8s-helm/SKILL.md) - Helm chart management
