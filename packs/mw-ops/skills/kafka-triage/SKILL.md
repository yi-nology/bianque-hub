---
name: kafka-triage
description: Kafka 故障分诊方法论：消费组积压三因子（消费慢/分区不均/生产激增）、分区 Leader 与 ISR 健康、存储水位、消息丢失/重复成因链——固定顺序定位与判读基准（ask-ops 只读采集面）。
mode: on_demand
version: 0.1.0
maturity: experimental
requires_mcp:            # 采集依赖的工具面（装载期对账依据；旧 bianque-tools 二进制缺这些工具时应先重建）
  - server: ask-ops
    tools: [run_readonly_command, run_readonly_commands]
---

## 触发条件

- 痪状关键词：消息积压、lag增长、消费不动、消费者掉线、isr收缩、分区不可用、发布超时、消息丢失、消息重复
- 组合场景：broker 进程/宿主异常先经 os-basics 或 k8s-ops 排除底层；rebalance 反复与消费逻辑超时强相关（需应用侧日志佐证）

## 数据来源（ask-ops 只读面）

- `kafka-consumer-groups.sh --bootstrap-server <bs> --describe --group <g>`：CURRENT-OFFSET / LOG-END-OFFSET / LAG / CONSUMER-ID；
- `kafka-topics.sh --bootstrap-server <bs> --describe --topic <t>`（Leader/Replicas/Isr；全量体检加 `--unavailable-partitions`）；
- `kafka-log-dirs.sh --bootstrap-server <bs> --describe --topic-list <t>`：各 broker 日录容量；
- 以上 CLI 从目标集群运维主机执行，凭证走已配置环境，不在对话回显。

## 分诊路径（固定顺序）

1. **消费积压三因子**：`--describe --group` 拿 lag → 差分定性：消费增速≈0 且 LOG-END 平稳=**消费停摆**（CONSUMER-ID 缺失=组已解散或 rebalance 循环）；消费增速>0 但 < 生产增速=**消费能力不足**（单线程处理慢/批量过小）；LOG-END 激增=**生产激增**（去业务侧归因）。分区级 lag 严重倾斜=分区键设计问题。
2. **分区健康**：`kafka-topics --describe` 逐分区看 Isr 数对 Replicas 数：Isr 收缩=同步副本掉队（追因 broker 负载/网络）；Leader 显示 none=分区不可用；`--unavailable-partitions` 全局扫不可用分区。
3. **存储水位**：`kafka-log-dirs` 看目录容量对磁盘总量：逼近上限时联系 retention 配置（log.retention.hours / segment 大小）与 topic 数据倾斜；磁盘满是 ISR 收缩与 broker 掉线的常见上游因。
4. **消息丢失/重复成因链**（无法直接回查，只做证据链）：生产端配置面（acks=all? retries?）× 分区面（min.insync.replicas 对当前 Isr 数——Isr 跌破 min.insync.replicas 时 acks=all 生产会失败而非丢失）× 消费端提交语义（enable.auto.commit / 手动提交时机 / auto.offset.reset）× rebalance 时点与 lag 突变对齐。输出「成因链假设 + 待验证点」，禁无证据断言。

## 判读基准

- lag 是存量：结论必须写清差分窗口与三因子归属，禁只报「lag 很大」；
- Isr 收缩是瞬态还是持续要看二次采样；持续收缩且无恢复趋势才升级 CRIT；
- rebalance 风暴判据：组内成员 ID 短窗反复变更 + lag 锯齿突变，两者同时成立才定性；
- 消息「乱序」在单分区内才可判（跨分区间无序是设计行为，不是故障）。

## 输出要求

- 结论附命令输出关键行证据，归因区分「消费侧 / broker 侧 / 存储水位 / 配置语义」；变更类动作（重置位点、迁 leader、调 retention、扩分区）标注影响面并 requires_approval；数据不足输出「需补充采集」清单（如应用侧消费日志、生产端配置）。
