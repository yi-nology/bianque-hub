---
name: k8s-node-diagnosis
description: K8s 节点异常（NotReady/资源压力/调度失败）诊断方法论：k8sgpt 节点面定位四类压力与调度失败归因，区分节点层与工作负载层。
mode: on_demand
version: 0.3.0
maturity: experimental
requires_mcp:
  - server: k8sgpt
    tools: [get-resource, get-logs, list-events]
---
## 触发条件

- 节点 NotReady、节点资源压力告警（CPU/内存/PID/磁盘）、Pod 大面积 Pending、
  调度失败事件、节点上的 Pod 被驱逐（Evicted）；
- 组合场景：工作负载层异常先定位到节点传导时（k8s-ops/k8s-workload-triage 下游）。

## 数据来源

- k8sgpt 三工具：`list-events`（事件链）、`get-resource`（Node/Pod 状态与资源）、
  `get-logs`（容器日志，previous=true 取崩溃前输出）；
- 本技能只消费已声明工具；节点宿主侧命令（dmesg/journalctl 等）不在本技能面，
  需要时标注需人工执行或走宿主只读命令面。

## 方法论（固定顺序）

1. **事件链**：`list-events` 看命名空间近期事件（驱逐/探针失败/亲和冲突/镜像拉取），
   按 involvedObject 归并到节点；
2. **节点条件**：`get-resource`（Node）看 Ready/MemoryPressure/DiskPressure/
   PIDPressure 四条件与 allocatable 余量，定位问题节点与压力类型；
3. **负载对位**：`get-resource`（Pod）看问题 Pod 的 QoS 与 limits，区分「节点资源
   被个别负载吃满」与「节点本身异常」；
4. **日志佐证**：`get-logs` 拉问题容器日志（previous=true 看崩溃前输出）；
5. **逐类排除**：资源压力按 CPU/内存/PID/磁盘四类逐一排除，每类给证据行；
6. **归因分层**：结论必须区分「节点层」（kubelet/容器运行时/内核/硬件）与
   「工作负载层」（单负载资源越界/探针误配）。

## 判读基准

- Ready=False + 心跳事件缺失 → kubelet/网络面；MemoryPressure=True → 按 Pod
  内存用量排序找越界者，驱逐事件（Evicted）与此对位；
- DiskPressure=True → 镜像/容器日志/emptyDir 三类占用大户，镜像清理属节点
  变更动作不进本技能处方；
- PIDPressure=True → 高危信号（进程表耗尽影响全节点），建议面直接给端口占用
  进程清单；
- Pod Pending + FailedScheduling（Insufficient x）→ 资源面；无事件 → 调度器/
  亲和面。

## 输出要求

- 结论给「压力类型/问题节点/问题负载/归因层级」四要素，每要素带事件或资源
  证据行；
- 涉及节点处置（排水/重启 kubelet/清镜像）只进建议面走审批，不在本轮执行。
