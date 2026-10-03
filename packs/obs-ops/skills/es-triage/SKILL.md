---
name: es-triage
description: Elasticsearch 分诊方法论：集群红黄定性、未分配分片归因（allocation explain）、磁盘水位三级保护、节点掉线、线程池拒绝与堆压力——固定顺序定位与判读基准（ask-ops 只读采集面）。
mode: on_demand
version: 0.1.1
maturity: experimental
requires_mcp:            # 采集依赖的工具面（装载期对账依据；旧 bianque-tools 二进制缺这些工具时应先重建）
  - server: ask-ops
    tools: [run_readonly_command, run_readonly_commands]
provides_changes:        # 编目变更块（批次九十四「受审执行」）：本技能方法论覆盖的处方编目
  - obs-es-index-unblock-readonly # 解除 flood_stage 只读锁（params: host/port/index）——处方优先引用编目而非自由 steps
  - obs-es-allocation-enable      # 恢复分片分配（params: host/port）——升级/维护遗忘项高频处置
---

## 触发条件

- 症状关键词：es集群红、集群状态黄、分片未分配、es慢、日志写不进、es节点掉线、elasticsearch排障
- 组合场景：作为 Grafana/日志查询后端时先经 grafana-triage 判层归属；K8s 内部署的节点异常先经 k8s-ops 排除工作负载层

## 数据来源（ask-ops 只读面）

- 集群面（GET，URL 带 `&` 查询参数时整个 URL 必须加引号）：`/_cluster/health`（status/unassigned_shards/active_shards_percent）、`/_cluster/allocation/explain`（未分配原因——**只读解释端点**）、`/_cat/nodes?v`（角色/堆/磁盘水位列）、`/_cat/shards`（分片分布，如 `curl -s 'http://<es>:9200/_cat/shards?v&h=index,shard,prirep,state,unassigned.reason'`）；
- 节点面（GET）：`/_nodes/stats`（jvm 压力、thread_pool 拒绝数、breaker 触发）、`/_cat/indices?v`（索引级健康与大小）；
- 主机侧：ES 进程 CPU/内存、数据目录 df、日志尾部；
- 凭证由主机侧已配置环境注入，对话不回显。

## 分诊路径（按层定位，固定顺序）

1. **集群状态定性**：`/_cluster/health`——**red**（有主分片缺失，读写都受损）优先于 **yellow**（仅副本缺失，读正常写正常）。red 时第 2 步必须先做。
2. **未分配归因**：`/_cluster/allocation/explain` 拿权威原因 → 常见五类：**磁盘水位**（见第 3 步）、**节点数/专区不足**（副本放不下：节点数 < 副本需求或 zone awareness 不满足）、**分片过大恢复超时**（恢复中 give-up，重试是变更建议）、**版本不兼容**（滚动升级期间高低版本混布，某些只分配到特定版本）、**显式禁用了分配**（cluster.routing.allocation.enabled 被设为 none——升级/维护后的遗忘项，高发）。
3. **磁盘水位三级**：`/_cat/nodes` 的磁盘列对三级线（low/high/flood_stage，默认 85%/90%/95%）——**flood_stage 会强制索引只读**（写入报错「read-only allow_delete」），此时「日志写不进」的真因是磁盘而不是 ES 本体；解除只读（改 watermark 或清磁盘后 reroute）是变更建议。
4. **节点掉线**：`/_cat/nodes` 对比预期拓扑 → 掉线节点的分片会被集群重平衡（IO 波动是次生现象）；归因在主机/容器层（联动 os-basics / k8s-ops）；**脑裂防护下的少数派节点会自我关闭**——看到单节点「疑似死亡」先查是否被投票保护，不要急着拉起。
5. **压力面**：thread_pool 拒绝数（search/write bulk 持续增长=请求过载或慢查询拖累）、JVM old GC 频率与 heap 使用率（>85% 持续=堆压力，breaker 会先拦大查询）；索引级 `/_cat/indices` 找异常大户（超大索引/疯狂滚动的新索引）。

## 判读基准

- yellow 是**可用状态**（副本缺失不影响读写）：只有持续不收敛或副本长期无法分配才列异常；
- red 必须给出「哪些索引的哪些主分片」证据（`/_cat/shards` 中未分配的主分片清单），禁笼统说「集群红了」；
- flood_stage 只读锁的解除顺序是「先治磁盘、再解除锁」——直接解锁而不清磁盘会立刻复发；
- bulk 拒绝与客户端重试风暴互为因果，结论要给时间对齐证据。

## 输出要求

- 每个结论附 API 输出关键行证据；变更类动作（reroute、改水位、清索引、扩节点、解除只读）标注影响面并 requires_approval；数据不足输出「需补充采集」清单（如完整 allocation explain、节点日志），不臆测。

**受审执行处方（批次九十四）**：两块高频 ES 处置优先引用编目变更块——flood 只读锁解除用 `recommendation.change_ref: obs-es-index-unblock-readonly` + `change_params: {host: "<es 地址>", port: "<端口>", index: "<索引名或 _all>"}`（前置铁律：先治磁盘再解锁）；分片分配恢复用 `obs-es-allocation-enable` + `change_params: {host: "<es 地址>", port: "<端口>"}`（cluster.routing.allocation.enabled=none 遗忘项）；均不给自由 steps，变更后 `_settings`/`_cluster/settings` 自动验证；reroute、改水位、清索引、扩节点仍走自由 steps + requires_approval（审批卡标「未编目」）。编目未安装的站点照常走自由 steps，语义不变。
