---
name: rabbitmq-triage
description: RabbitMQ 故障分诊方法论：队列堆积三态（ready 高/unacked 高/无消费者）、连接与限流状态、内存磁盘水位告警、集群脑裂一致性、镜像与仲裁队列健康——固定顺序定位与判读基准（ask-ops 只读采集面）。
mode: on_demand
version: 0.1.0
maturity: experimental
requires_mcp:            # 采集依赖的工具面（装载期对账依据；旧 bianque-tools 二进制缺这些工具时应先重建）
  - server: ask-ops
    tools: [run_readonly_command, run_readonly_commands]
---

## 触发条件

- 症状关键词：队列堆积、消费者掉线、rabbitmq连不上、发布被阻、内存告警、磁盘告警、脑裂、镜像队列异常
- 组合场景：Erlang 节点宿主资源异常先经 os-basics 排除；K8s 内部署先经 k8s-ops 排除工作负载层

## 数据来源（ask-ops 只读面）

- `rabbitmqctl list_queues name messages_ready messages_unacknowledged consumers`（堆积三态主面板）；
- `rabbitmqctl list_connections name state send_pend`、`rabbitmqctl list_channels number consumers`；
- `rabbitmqctl status`（mem_used/mem_limit、disk_free/disk_free_limit、告警开关）、`rabbitmq-diagnostics cluster_status`（分区线与运行中节点）；
- management 插件在位时可 GET `/api/overview`、`/api/queues`（只读，curl 走已配置凭证）。

## 分诊路径（固定顺序）

1. **队列堆积三态**：`list_queues` 逐队列判读：ready 高且 consumers=0=**消费者掉线**（追部署与订阅端）；unacked 高=**消费者持有不确认**（处理慢/卡死/预取过大）；ready 高且 consumers 在位=**消费速率不足**（扩消费者/调 prefetch 属变更建议）。
2. **连接与限流**：`list_connections` 看 state：flow=服务端对发布方限流（正常节流形态，持续 flow 才构成结论）；blocking/blocked=内存或磁盘告警触发（去第 3 步）；send_pend 持续大=网络出口积压。
3. **水位告警**：`status` 对比 mem_used/mem_limit 与 disk_free/disk_free_limit：触发后集群**阻塞发布方**（这是保护机制，不是故障本体）——先归因水位来源（内存：队列堆积/消息体大/插件；磁盘：持久化日志增长/未消费堆积），解法多为先治堆积再解水位。清理磁盘与重启节点是变更动作，只建议。
4. **集群一致性**：`cluster_status` 查 partitions：非空=**脑裂在案**（记录分区线与两侧节点，处置属变更类建议）；同时核对 quorum/镜像队列的 leader 所在节点——leader 集中在单节点会制造倾斜。
5. **队列形态面**：classic 镜像队列（ha-mode）与 quorum 队列混布时分别核对同步状态（`list_queues` 加同步参数或 management API）；镜像不同步是 failover 风险，不是当前故障，列为 WARN。

## 判读基准

- unacked 高与 prefetch 大相关：判读前先了解或标注 prefetch 未知（需应用侧确认），不要直接断言消费卡死；
- 短时 ready 波动是流量形态；持续单调增长且消费速率不升才是堆积结论；
- 脑裂结论必须给出分区线证据（哪些节点互相不可见），禁仅凭「某节点响应慢」定性；
- 内存水位告警阈值默认相对值（如物理内存比例），结论里写明实际 limit 数值口径。

## 输出要求

- 结论附命令输出关键行证据，归因区分「消费侧 / 连接面 / 水位保护 / 集群一致性」；变更类动作（清队列、重启节点、磁盘清理、调 prefetch/镜像策略）标注影响面并 requires_approval；数据不足输出「需补充采集」清单（如应用侧消费确认日志）。
