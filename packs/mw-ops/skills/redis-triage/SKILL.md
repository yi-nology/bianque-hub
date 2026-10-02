---
name: redis-triage
description: Redis 五类故障分诊方法论：内存满/驱逐风暴、延迟毛刺、缓存雪崩击穿穿透、主从复制中断、连接打满——固定顺序定位路径与判读基准（ask-ops 只读采集面）。
mode: on_demand
version: 0.1.0
maturity: experimental
requires_mcp:            # 采集依赖的工具面（装载期对账依据；旧 bianque-tools 二进制缺这些工具时应先重建）
  - server: ask-ops
    tools: [run_readonly_command, run_readonly_commands]
---

## 触发条件

- 症状关键词：redis慢、redis内存满、key被驱逐、缓存命中率下降、redis连不上、主从断了、雪崩、击穿、穿透
- 组合场景：延迟毛刺定位到数据面（转 redis-hotkey-bigkey 专项）；K8s 内部署的 Redis 先经 k8s-ops 工作负载分诊排除宿主层

## 数据来源（ask-ops 只读面）

- `redis-cli -h <host> -p <port> INFO <section>`：memory / stats / replication / clients / persistence 分节；
- `SLOWLOG GET 20`、`INFO commandstats`、`CLIENT LIST`；集群加 `CLUSTER INFO` / `CLUSTER NODES`，哨兵部署在哨兵节点 `SENTINEL masters`；
- 凭证由主机侧已配置客户端注入，对话不回显密码。

## 分诊路径（按症状类定位，固定顺序）

1. **内存满/驱逐风暴**：`INFO memory`（used_memory 对 maxmemory、mem_fragmentation_ratio）→ `INFO stats`（evicted_keys 增速、keyspace_hits/misses 差分）→ 判读：驱逐持续增长=容量不足或过期策略失配；fragmentation_ratio 持续 >1.5 且无大批量删除=碎片膨胀（jemalloc 语境）。
2. **延迟毛刺**：`SLOWLOG GET 20` 先看命令面（O(N) 命令/大批量/MULTI）→ `INFO stats` 的 instantaneous_ops_per_sec 对基线 → `INFO persistence`（rdb_last_bgsave_status / aof_last_write_status / bgsave_in_progress=1 期间 fork 停顿嫌疑）。归因三分：命令复杂度 / 数据面（转专项）/ 持久化 fork。
3. **命中率异常（雪崩/击穿/穿透）**：命中率 = hits/(hits+misses)，必须带时间窗 → miss 激增三分：大批 key 同时过期（雪崩，查业务整点规律与 expires 分布）、热 key 失效（击穿，转专项）、查询不存在的 key（穿透，miss 激增但 keyspace 无对应前缀增长）。
4. **主从复制中断**：主库侧 `INFO replication`（connected_slaves 与各 slave 的 offset 差）、从库侧 master_link_status=down 与 master_repl_offset 差值 → 归因三分：网络（联动主机网络域）/ 主库压力（repl-backlog 溢出）/ 认证与配置变更。切主是变更动作：只产出建议与影响面，禁执行。
5. **连接打满**：`INFO clients`（connected_clients 对 maxclients、blocked_clients）→ `CLIENT LIST` 看 Age/Idle 异常与来源 IP 分布 → 归因三分：应用连接泄漏 / maxclients 过低（`CONFIG GET maxclients` 只读核实）/ BLPOP 类阻塞堆积。

## 判读基准

- 计数器结论必须给「时间窗+增量」口径（两次采样差分或 slowlog 时间戳），禁单点快照定论；
- 命中率下降先排除统计窗内请求结构变化（新业务 key 空间进入）再归因缓存策略；
- 主从 offset 差 <1MB 且 link 正常属正常抖动，不构成异常结论；
- blocked_clients > 0 且持续增长才构成结论，瞬态阻塞是 BLPOP 正常形态。

## 输出要求

- 每个结论附 INFO/slowlog 关键行证据；变更类动作（扩 maxmemory、切主、清 key、改策略）一律标注影响面并 requires_approval；数据不足输出「需补充采集」清单（如应用侧日志、监控基线）。
