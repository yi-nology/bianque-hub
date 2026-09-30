---
name: k8s-node-diagnosis
description: K8s 节点异常（NotReady/资源压力/调度失败）诊断方法论
mode: on_demand
version: 0.2.0
provides: [k8s-node-diagnosis]
requires_mcp:
  - server: k8sgpt
    tools: [get-resource, get-logs, list-events]
---
## 触发条件

节点 NotReady、Pod Pending/ CrashLoopBackOff、调度失败事件。

## 方法论

1. `list-events` 看命名空间近期事件链（驱逐/探针失败/亲和冲突/镜像拉取）；
2. `get-resource`（Node/Pod）看 Ready/压力条件与资源分配，定位问题对象；
3. `get-logs` 拉问题容器日志（previous=true 看崩溃前输出）；
4. 资源压力区分 CPU/内存/PID/磁盘四类，逐一排除；
5. 结论必须区分「节点层」与「工作负载层」归因。
