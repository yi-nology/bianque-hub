---
name: k8s-events-timeline
description: K8s 事件时间线分析方法论：从 list-events 输出构建事件链、识别因果起点、区分噪音信号（Normal 级）与故障信号（Warning 级）。
mode: on_demand
version: 0.1.0
maturity: experimental
requires_mcp:
  - server: k8sgpt
    tools: [list-events, list-namespaces]
---

## 触发条件

- 症状关键词：集群日志、事件排查、k8s 日志、事件时间线、「为什么 pod 被删/驱逐/重启」
- 组合场景：任何工作负载异常的佐证环节（k8s-workload-triage 第 3/5 步深化）

## 方法论（固定顺序）

1. **全量拉取**：`list-events`（按命名空间过滤；未指明 ns 时先 `list-namespaces` 圈定涉事 ns）；
2. **分级过滤**：Warning 级全留；Normal 级只留四类信号——Scheduled（调度决策）、Pulled/Created/Started（生命周期翻转）、Killing（终止发起）、SuccessfulCreate/Delete（控制器动作）；
3. **建时间线**：按 lastTimestamp 排序，标注 involvedObject（kind/name）；同一对象的事件串成链；
4. **找因果起点**：时间线上第一条 Warning（或异常 Normal 翻转）之前 5 分钟内的关联事件（同节点/同控制器）为候选根因窗口；
5. **计数归并**：count 字段大（>10）的重复事件降权为「持续状态」而非「新事件」，避免时间线被刷屏淹没。

## 判读基准

- `BackOff` 事件跟着 `Pulled` 失败 → 镜像/凭证面；`FailedScheduling` → 资源/亲和面；`Evicted` → 节点压力面；
- `Killing` 的 reason/消息字段指出发起者（liveness probe / preemption / eviction / OOM）——这是「谁杀的 pod」的直接答案；
- events 只有 1 小时保留期（缺省 etcd ttl）：查不到历史事件≠没发生过，改查 workload 的 lastState 与节点 dmesg。

## 输出要求

- 时间线以「HH:MM 事件（对象）→ 影响」列表呈现；结论指明因果起点事件与传播链，不得只罗列事件。
