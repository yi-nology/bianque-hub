---
name: mysql-triage
description: MySQL 五类故障分诊方法论：连接打满、锁与事务堵塞、慢查询定位、容量与 binlog 增长、缓冲池与抖动判读——固定顺序定位路径与判读基准（ask-ops 只读采集面）。
mode: on_demand
version: 0.1.0
maturity: experimental
---

## 触发条件

- 症状关键词：数据库连接打满、连不上 mysql、锁表、行锁堵塞、查询堆积、数据库慢、磁盘涨、binlog 堆积
- 组合场景：主从延迟/复制断开转 mysql-replication 专项；宿主机资源异常先经 os-basics 排除

## 数据来源（ask-ops 只读面）

- `SHOW GLOBAL STATUS LIKE '<pattern>'`（Threads_%、Aborted_%、Innodb_%、Slow_queries）、`SHOW VARIABLES`（max_connections、innodb_buffer_pool_size、long_query_time）；
- `SHOW PROCESSLIST`（完整态用 `SELECT ... FROM information_schema.PROCESSLIST` 限行）、`SHOW ENGINE INNODB STATUS`；
- `performance_schema.events_statements_summary_by_digest`（Top SQL 面）、`information_schema.INNODB_TRX` / `data_lock_waits`（8.0；5.7 用 information_schema.innodb_lock_waits）；
- 容量面：`du` 数据目录/bdir 段、`SHOW BINARY LOGS` 总量；凭证由主机侧已配置客户端注入，对话不回显。

## 分诊路径（按症状类定位，固定顺序）

1. **连接打满**：`Threads_connected` 对 `max_connections` → `Aborted_connects` 增速（失败握手：凭证/网络/back_log）→ `SHOW PROCESSLIST` 看 Time 长的 Sleep 堆积来源（应用连接池泄漏 vs 空闲超时失配）。归因三分：池泄漏 / max_connections 过低 / 认证与网络失败。
2. **锁与事务堵塞**：`INNODB_TRX` 找长事务（trx_started 最早者=嫌疑源头）→ `data_lock_waits` 确认等待链（谁阻塞谁）→ `SHOW ENGINE INNODB STATUS` 的 TRANSACTIONS 段交叉。判读：行锁等 MDL 锁要区分（processlist 中「Waiting for table metadata lock」=MDL，源头常是未提交事务/长查询）；间隙锁导致的「锁范围扩大」需结合事务隔离级别口径。
3. **慢查询定位**：`Slow_queries` 增速 → digest 表按 SUM_TIMER_WAIT 排序取 Top（注意 TIMER_SUM 与 CALLS 组合看：偶发大查询 vs 高频小查询处理方向不同）→ 对 Top digest 取样 `EXPLAIN` 口径。禁直接对业务库执行未知 SQL。
4. **容量与 binlog 增长**：数据目录 du 定位大户（ibdata/表空间/binlog/relaylog/tmp）→ `SHOW BINARY LOGS` 看总量对 `binlog_expire_logs_seconds`；relaylog 堆积常指向复制回放慢（转 mysql-replication）。
5. **缓冲池与抖动判读**：`Innodb_buffer_pool_read_requests/reads` 算命中率 → 命中率低 + 大量物理读 = 池不足或冷数据扫；`Innodb_row_lock_time_avg` 与 `Threads_running` 锯齿 = 行锁热点或突发并发。

## 判读基准

- 计数器结论必须给「时间窗+增量」口径（两次采样差分），禁单点快照定论；
- Sleep 连接多不必然是问题：结合 wait_timeout 与业务形态（常驻池）判读，只有 Time 远超池空闲上限且量级持续增长才构成泄漏结论；
- MDL 等待的源头事务要在 PROCESSLIST/INNODB_TRX 双侧对上才下结论；
- KILL 连接/事务是变更动作：只建议并标注「将终止哪些会话」的影响面。

## 输出要求

- 每个结论附 SQL 输出关键行证据；变更类动作（KILL、调参、清 binlog、重启）标注影响面并 requires_approval；数据不足输出「需补充采集」清单（如业务侧慢日志原文、监控基线）。
