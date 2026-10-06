---
name: k8s-openapi
description: 参考 Kubernetes API。用于了解 Kubernetes REST API 可以实现什么功能。
mode: on_demand
version: 1.1.1
maturity: experimental
---

1. 使用以下命令启动测试集群：`make test-cluster-create`。仅在此测试集群上使用此技能。
2. 仅通过每个 `kubectl` 命令的 `--context` 和 `--kubeconfig` 使用新上下文。不要通过 `kubectl config` 修改配置。
3. 使用 `kubectl proxy` 启动代理服务器
4. 通过查询 `/openapi/v3` 发现可用 endpoints
5. 使用发现的 endpoints 查找你需要的信息
6. 当你找到答案后：
    - 停止代理进程；并
    - 使用 `make test-cluster-stop` 关闭测试集群


> ⚠️ **收割占位（stub）**：本技能收割时仅落了元信息，正文方法论待补全；
> 使用前以下游官方文档为准，勿按此空壳行事。同类完整件见同域技能（如 loki-ops）。
