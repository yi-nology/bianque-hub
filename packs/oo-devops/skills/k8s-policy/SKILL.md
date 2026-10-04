---
name: k8s-policy
description: 使用 Kyverno 和 Gatekeeper 进行 Kubernetes 策略管理。在执行安全策略、验证资源或审计策略合规性时使用。
mode: on_demand
version: 1.1.0
maturity: experimental
---

# Kubernetes 策略管理

使用 kubectl-mcp-server 的 Kyverno 和 Gatekeeper 工具管理策略。

## 何时应用

在以下情况下使用此技能：
- 用户提及："Kyverno"、"Gatekeeper"、"OPA"、"policy"、"compliance"
- 操作：执行策略、检查违规、策略审计
- 关键词："require labels"、"block privileged"、"validate"、"enforce"

## 优先级规则

| 优先级 | 规则 | 影响 | 工具 |
|----------|------|--------|-------|
| 1 | 首先检测策略引擎 | 严重 | `kyverno_detect_tool`、`gatekeeper_detect_tool` |
| 2 | 强制执行前使用审计模式 | 高 | validationFailureAction |
| 3 | 检查策略报告中的违规 | 高 | `kyverno_clusterpolicyreports_list_tool` |
| 4 | 审查约束模板 | 中 | `gatekeeper_constrainttemplates_list_tool` |

## 快速参考

| 任务 | 工具 | 示例 |
|------|------|---------|
| 列出 Kyverno 集群策略 | `kyverno_clusterpolicies_list_tool` | `kyverno_clusterpolicies_list_tool()` |
| 获取 Kyverno 策略 | `kyverno_clusterpolicy_get_tool` | `kyverno_clusterpolicy_get_tool(name)` |
| 列出 Gatekeeper 约束 | `gatekeeper_constraints_list_tool` | `gatekeeper_constraints_list_tool()` |
| 获取约束 | `gatekeeper_constraint_get_tool` | `gatekeeper_constraint_get_tool(kind, name)` |

## Kyverno

### 检测安装

```python
kyverno_detect_tool()
```

### 列出策略

```python
kyverno_clusterpolicies_list_tool()

kyverno_policies_list_tool(namespace="default")
```

### 获取策略详情

```python
kyverno_clusterpolicy_get_tool(name="require-labels")
kyverno_policy_get_tool(name="require-resources", namespace="default")
```

### 策略报告

```python
kyverno_clusterpolicyreports_list_tool()

kyverno_policyreports_list_tool(namespace="default")
```

### 常见 Kyverno 策略

```python
kubectl_apply(manifest="""
apiVersion: kyverno.io/v1
kind: ClusterPolicy
metadata:
  name: require-labels
spec:
  validationFailureAction: Enforce
  rules:
  - name: require-app-label
    match:
      resources:
        kinds:
        - Pod
    validate:
      message: "Label 'app' is required"
      pattern:
        metadata:
          labels:
            app: "?*"
""")

kubectl_apply(manifest="""
apiVersion: kyverno.io/v1
kind: ClusterPolicy
metadata:
  name: require-limits
spec:
  validationFailureAction: Enforce
  rules:
  - name: require-cpu-memory
    match:
      resources:
        kinds:
        - Pod
    validate:
      message: "CPU and memory limits are required"
      pattern:
        spec:
          containers:
          - resources:
              limits:
                cpu: "?*"
                memory: "?*"
""")
```

## Gatekeeper (OPA)

### 检测安装

```python
gatekeeper_detect_tool()
```

### 列出约束

```python
gatekeeper_constraints_list_tool()

gatekeeper_constrainttemplates_list_tool()
```

### 获取约束详情

```python
gatekeeper_constraint_get_tool(
    kind="K8sRequiredLabels",
    name="require-app-label"
)

gatekeeper_constrainttemplate_get_tool(name="k8srequiredlabels")
```

### 常见 Gatekeeper 策略

```python
kubectl_apply(manifest="""
apiVersion: templates.gatekeeper.sh/v1
kind: ConstraintTemplate
metadata:
  name: k8srequiredlabels
spec:
  crd:
    spec:
      names:
        kind: K8sRequiredLabels
      validation:
        openAPIV3Schema:
          type: object
          properties:
            labels:
              type: array
              items:
                type: string
  targets:
  - target: admission.k8s.gatekeeper.sh
    rego: |
      package k8srequiredlabels
      violation[{"msg": msg}] {
        provided := {label | input.review.object.metadata.labels[label]}
        required := {label | label := input.parameters.labels[_]}
        missing := required - provided
        count(missing) > 0
        msg := sprintf("Missing labels: %v", [missing])
      }
""")

kubectl_apply(manifest="""
apiVersion: constraints.gatekeeper.sh/v1beta1
kind: K8sRequiredLabels
metadata:
  name: require-app-label
spec:
  match:
    kinds:
    - apiGroups: [""]
      kinds: ["Pod"]
  parameters:
    labels: ["app", "env"]
""")
```

## 策略审计工作流

```python
kyverno_detect_tool()
kyverno_clusterpolicies_list_tool()
kyverno_clusterpolicyreports_list_tool()
```

## 先决条件

- **Kyverno**: Kyverno 工具所需
  ```bash
  kubectl create -f https://github.com/kyverno/kyverno/releases/latest/download/install.yaml
  ```
- **Gatekeeper**: Gatekeeper 工具所需
  ```bash
  kubectl apply -f https://raw.githubusercontent.com/open-policy-agent/gatekeeper/master/deploy/gatekeeper.yaml
  ```

## 相关技能

- [k8s-security](../k8s-security/SKILL.md) - RBAC 和安全
- [k8s-operations](../k8s-operations/SKILL.md) - 应用策略
