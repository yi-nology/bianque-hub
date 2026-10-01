---
name: mysql-replication
description: MySQL 复制链诊断方法论：双线程（IO/SQL）定位、延迟三源归因（拉取慢/回放慢/大事务）、GTID 断点与常见错误码判读、并行回放与半同步状态核查（ask-ops 只读采集面）。
mode: on_demand
version: 0.1.0
maturity: experimental
---

## 触发条件

- 症状关键词：主从延迟、主从断开、复制报错、从库追不平、binlog 不同步、GTID 断点
- 组合场景：mysql-triage 分诊第 4 类命中 relaylog 堆积后进入本专项；切换/重建从库属变更动作，本技能只产出方案面

## 数据来源（ask-ops 只读面）

- 从库：`SHOW REPLICA STATUS\G`（8.0.22+；旧版 `SHOW SLAVE STATUS\G`）——双线程状态、Seconds_Behind_Source、Last_IO/SQL_Errno 与 Error、Retrieved/Gtid_Set 对 Executed_Gtid_Set；
- 主库：`SHOW MASTER STATUS`（或 `SHOW BINARY LOG STATUS`）、`SELECT @@gtid_mode`；
- 并行回放：`SELECT * FROM performance_schema.replication_applier_status_by_worker LIMIT <n>`；
- 半同步：`SHOW GLOBAL STATUS LIKE 'Rpl_semi_sync%'`（插件在位时）。

## 方法论（固定顺序）

1. **双线程定位**：Replica_IO_Running / Replica_SQL_Running 双标志先定性——IO 停=**拉取断**（网络/主库 binlog 被清/认证失效，看 Last_IO_Error）；SQL 停=**回放断**（数据冲突，看 Last_SQL_Errno）；双 Running 但延迟=进第 2 步。
2. **延迟三源归因**：Seconds_Behind_Source 持续大时三分：**拉取慢**（Retrieved 领先 Executed 很多=回放慢；Retrieved 也不涨=拉取侧，查网络与主库压力）、**回放慢**（单线程回放瓶颈或 slave_parallel_workers=0，看 worker 表）、**大事务/无事务时钟基准**（Seconds_Behind=0 突跳的锯齿常是大事务回放瞬间造成，回放期间统计失真）。判定必须给 Retrieved/Executed 双侧位点差证据。
3. **错误码判读**（回放断场景）：1032（记录不存在——从库有脏写或位点错）、1062（主键冲突——双写或位点回退）、1594（中继日志损坏）；处置口径：跳过错误属**数据丢失风险动作**，必须先核对 GTID 断点与业务影响，只输出方案建议，禁直接执行。
4. **GTID 与位点口径**：`gtid_mode=ON` 时用 Executed_Gtid_Set 主从对比找断点事务段；OFF 时用 Relay_Master_Log_File/Exec_Master_Log_Pos 口径，结论里写明模式（口径不同，修复路径不同）。
5. **并行与半同步核查**：worker 表看各 worker 队列积压分布；半同步状态（Rpl_semi_sync_master_status/clients）核查「半同步退化成异步」——退化后主库崩溃窗口内有丢失风险，列 WARN 与建议。

## 判读基准

- Seconds_Behind_Source 是「事件时间戳差」不是「真实延迟」，回放大事务期间读数失真——结论要结合 binlog 事务尺度判读；
- 延迟 <1s 的波动属正常；持续增长且不收敛才是异常结论；
- 主从数据一致性无法用状态页证明：只能输出「位点一致性」结论，禁说「数据完全一致」。

## 输出要求

- 每个结论附状态页关键行证据（双 Running 标志、位点差、错误码原文）；变更类动作（跳错、重建从库、切主、启停复制）标注「影响面+数据风险」并 requires_approval；数据不足输出「需补充采集」清单，不臆测。
