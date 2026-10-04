---
name: k8s-operator-change
description: Kubernetes Operator 变更专家 - CRD Schema 修改、控制面协调逻辑调整与部署配置更新
mode: on_demand
version: 1.1.0
maturity: experimental
---

# Kubernetes Operator 变更

## 触发器

- 更改 operator APIs、CRDs 或面向控制器的配置转换
- 更改 semantic-router 的 Kubernetes Deployment control-plane behavior

## 工作流

1. 读取变更表面、功能完整检查清单和模块边界以获取 operator 上下文
2. 修改 operator APIs、CRDs 或面向控制器的配置转换
3. 运行 `make agent-report ENV=cpu CHANGED_FILES="..."` 以识别受影响的表面
4. 运行 `make agent-ci-gate CHANGED_FILES="..."` 以验证所有约束
5. 验证 operator APIs、CRDs 和面向路由器的转换保持一致

## 陷阱

- CRD schema 变更、转换更新和 reconciliation logic 通常需要一起落地以避免 split-brain 行为。
- 当 leaving the router-side config contract ambiguous 时不要修复 Kubernetes 表示。

## 必读

- [docs/agent/change-surfaces.md](../../../../docs/agent/change-surfaces.md)
- [docs/agent/feature-complete-checklist.md](../../../../docs/agent/feature-complete-checklist.md)
- [docs/agent/module-boundaries.md](../../../../docs/agent/module-boundaries.md)

## 标准命令

- `make agent-report ENV=cpu CHANGED_FILES="..."`
- `make agent-ci-gate CHANGED_FILES="..."`

## 验收

- Operator APIs、CRDs 和面向路由器的转换保持一致
- 当 behavior changes 时更新相关的面向 Kubernetes 的验证
