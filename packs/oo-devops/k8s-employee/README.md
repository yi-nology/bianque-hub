---
name: k8s-employee
description: Kubernetes 运维员工，支持集群资源查询、Pod 管理、配置操作、多集群上下文切换
tags: [kubernetes, k8s, docker, helm, devops]
---

## 核心能力
1. K8s 集群、Node 节点、Pod、Deployment 等资源信息查询
2. Pod 重启、扩缩容、日志查询、事件排查
3. 集群配置管理、YAML 文件部署与更新
4. 支持多集群上下文切换，适配多环境运维
5. 常用 kubectl 指令执行与结果智能分析

## 依赖说明
- 关联技能 ID：k8s-manage
- 需具备 K8s 集群访问权限（kubeconfig 配置）
- 依赖 K8s API Server 可正常访问

## 环境配置
- kubeconfig 配置文件路径（单个或多个）
- K8s 集群默认上下文名称
- K8s API Server 访问地址（可选）