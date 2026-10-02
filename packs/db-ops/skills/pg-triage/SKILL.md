---
name: pg-triage
description: PostgreSQL 故障分诊方法论：连接与 idle 事务、锁等待链（pg_blocking_pids）、autovacuum 与死元组膨胀、WAL 增长与复制槽滞留、缓存命中与 pg_stat_statements 查询面——固定顺序定位与判读基准（ask-ops 只读采集面）。
mode: on_demand
version: 0.1.0
maturity: experimental
---

## 触发条件

- 症状关键词：pg慢、连接耗尽、idle in transaction、锁等待、死元组膨胀、vacuum堵塞、wal堆积、复制槽滞留、磁盘涨
- 组合场景：宿主资源异常先经 os-basics 排除；跨库（MySQL）场景走 mysql-triage

## 数据来源（ask-ops 只读面）

- `SELECT ... FROM pg_stat_activity`（state 分布、wait_event、xact_start 最早的 idle in transaction）；
- `pg_locks` + `pg_blocking_pids(<pid>)`（等待链）、`pg_stat_user_tables`（n_dead_tup 对 n_live_tup）、`pg_stat_progress_vacuum`（在跑的 vacuum）；
- `pg_replication_slots`（active、restart_lsn 滞留）、`pg_stat_wal`（14+）或 WAL 目录 du、`pg_stat_database`（blks_hit/blks_read 命中率）、`pg_stat_statements`（扩展在位时 Top SQL）；
- 凭证由主机侧已配置 psql 注入（.pgpass/PGPASSWORD），对话不回显；无独立凭证面时用 `sudo -u postgres psql -c '…'`（run_readonly_command 支持 sudo -u 目标用户形态，须在 sudoers 面内）。

## 分诊路径（固定顺序）

1. **连接与 idle 事务**：`pg_stat_activity` 按 state 聚合——active 对 `max_connections` 看余量；`idle in transaction` 且 xact_start 早于阈值=**头号嫌疑**（持锁阻碍 vacuum、撑爆连接、阻碍复制槽推进三害同源），记下最早事务的 pid 与 query。
2. **锁等待链**：`pg_blocking_pids()` 找每个 waiting 进程的阻塞者，构建「谁堵谁」链；判读区分 LockType（行锁/表锁/咨询锁）与 wait_event 形态（Lock/LWLock/IO）。**杀后端是变更动作**：只输出目标 pid 清单建议。
3. **膨胀与 vacuum**：`n_dead_tup/n_live_tup` 比例（>20% 且不收敛=膨胀嫌疑）→ `pg_stat_progress_vacuum` 看 autovacuum 是否在跑/反复被取消（`n_ins_since_vacuum` 口径）→ 归因三分：长事务阻碍 xmin 推进 / autovacuum 参数过保守 / 写入速率超回收速率。VACUUM FULL 与重建表是变更建议（锁表影响面必须写明）。
4. **WAL 与复制槽**：WAL 目录/`pg_stat_wal` 看增速 → `pg_replication_slots` 逐槽看 active=false 或 restart_lsn 滞留——**废弃复制槽是 WAL 无限增长最常见的根因**；滞留结论给槽位字节数口径。删槽是变更动作。
5. **查询面**：`pg_stat_statements` 按 total_exec_time 排序 Top（mean 与 calls 组合看：偶发大查询 vs 高频小查询）→ 共享缓冲命中率 `blks_hit/(blks_hit+blks_read)`（<95% 且物理读大=池不足或冷扫描）→ 冷查询抖动（first read 慢）不构成命中率结论。

## 判读基准

- 统计视图是累计值：结论必须带时间窗（stats_reset 口径或两次采样差分）；
- idle in transaction 的判据是「时长+持锁证据」双条件，短时 idle 属正常事务间隙；
- 膨胀判定给比例口径并注明表大小量级（小表百分比无意义）；
- 复制槽 active=true 但 lag 大，是消费端慢（流复制对端或逻辑订阅端），与废弃槽处置方向不同。

## 输出要求

- 每个结论附查询输出关键行证据；变更类动作（pg_terminate_backend、VACUUM FULL、删槽、调参）标注影响面并 requires_approval；数据不足输出「需补充采集」清单（如需 DBA 开 pg_stat_statements 扩展），不臆测。
