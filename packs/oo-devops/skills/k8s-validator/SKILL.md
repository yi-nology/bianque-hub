---
name: k8s-validator
description: 验证 Kubernetes manifests 的安全性、最佳实践和资源限制
mode: on_demand
version: 1.1.0
maturity: experimental
---

# Kubernetes Manifest Validator 技能

## 概述

验证 Kubernetes manifests，包括使用 OPA/Gatekeeper 进行安全策略检查、使用 kube-linter 进行最佳实践检查以及资源限制验证。

## 能力

- 验证 Kubernetes manifests (YAML/JSON)
- 安全策略检查 (OPA/Gatekeeper)
- 最佳实践检查 (kube-linter, kubeval)
- 资源限制验证
- 网络策略分析
- RBAC 分析
- Pod 安全标准检查

## 目标流程

- iac-review
- devops-architecture-alignment
- resilience-patterns

## 输入模式

```json
{
  "type": "object",
  "required": ["manifestPaths"],
  "properties": {
    "manifestPaths": {
      "type": "array",
      "items": { "type": "string" },
      "description": "Paths to Kubernetes manifests"
    },
    "validators": {
      "type": "array",
      "items": {
        "type": "string",
        "enum": ["kubeval", "kube-linter", "opa", "kubesec"]
      },
      "default": ["kubeval", "kube-linter"]
    },
    "options": {
      "type": "object",
      "properties": {
        "kubernetesVersion": {
          "type": "string",
          "default": "1.28.0"
        },
        "strict": {
          "type": "boolean",
          "default": false
        },
        "customPolicies": {
          "type": "array",
          "description": "Paths to custom OPA policies"
        }
      }
    }
  }
}
```

## 输出模式

```json
{
  "type": "object",
  "properties": {
    "valid": {
      "type": "boolean"
    },
    "manifests": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "path": { "type": "string" },
          "kind": { "type": "string" },
          "name": { "type": "string" },
          "valid": { "type": "boolean" },
          "issues": { "type": "array" }
        }
      }
    },
    "securityFindings": {
      "type": "array"
    },
    "bestPracticeViolations": {
      "type": "array"
    },
    "resourceLimitIssues": {
      "type": "array"
    }
  }
}
```

## 使用示例

```javascript
{
  kind: 'skill',
  skill: {
    name: 'k8s-validator',
    context: {
      manifestPaths: ['k8s/*.yaml'],
      validators: ['kubeval', 'kube-linter', 'kubesec'],
      options: {
        kubernetesVersion: '1.28.0',
        strict: true
      }
    }
  }
}
```
