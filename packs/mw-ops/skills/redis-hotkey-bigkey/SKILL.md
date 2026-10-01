---
name: redis-hotkey-bigkey
description: Redis 热Key与大Key定位方法论：全量抽样扫描、嫌疑 key 定点体检、行为面交叉验证三段式，附拆分/打散/异步删除处置口径（ask-ops 只读采集面）。
mode: on_demand
version: 0.1.0
maturity: experimental
---

## 触发条件

- 症状关键词：热key、大key、内存不明增长、单节点倾斜、慢命令集中在少数 key、击穿嫌疑
- 组合场景：redis-triage 分诊第 2/3 类命中数据面嫌疑后进入本专项

## 数据来源（ask-ops 只读面）

- `redis-cli --bigkeys`（内置 SCAN 抽样，生产可用，给各类型 top）；
- `redis-cli --hotkeys`（仅 maxmemory-policy 为 LFU 系时可用，先 `CONFIG GET maxmemory-policy` 核实）；
- 定点命令：`MEMORY USAGE <key> SAMPLES 5`、`OBJECT ENCODING/FREQ/IDLETIME <key>`；
- 行为面：`INFO commandstats` 调用集中度、`SLOWLOG` 中同一 key 反复出现、应用侧访问日志。
- 禁用 `MONITOR` 与 `KEYS *`（全量流量镜像/O(N) 扫描）。

## 方法论（固定顺序）

1. **全量抽样面**：`--bigkeys` 拿各类型（string/hash/list/set/zset）最大 key 榜单；LFU 策略核实通过后补 `--hotkeys` 榜单，不满足则记录「热 key 口径缺失，改走第 3 步」。
2. **定点体检面**：对嫌疑 key 逐个 `MEMORY USAGE`（字节数）、`OBJECT ENCODING`（ziplist/skiplist 等编码形态）、`OBJECT IDLETIME`、LFU 下 `OBJECT FREQ`（访问频率对数刻度，数值越大越热）。
3. **行为交叉面**：`INFO commandstats` 看主命令调用占比；slowlog 时间线与业务 key 命中是否重合；无 LFU 口径时以应用侧访问日志采样替代（请求结论里明示证据等级降档）。

## 判读基准

- 大 key 参考：string > 1MB、集合类元素 > 5000 或序列化后 > 1MB 即列入处置建议（阈值需结合实例规格与基线，写结论时带口径）；
- 热 key 参考：单 key QPS 占实例总 QPS >10%，或绝对值超过单核处理能力明显比例（数万 QPS 级）；
- 删除引发的不一致：fragmentation_ratio 在大批删除后短期走高属正常回收形态，不要误判为泄漏；
- 集群模式单分片 CPU/带宽倾斜 + 该分片 key 命中集中，才构成「热 key 导致分片倾斜」结论。

## 输出要求

- 结论给 top 榜单证据行 + 证据等级（全量抽样/定点/行为面）；处置建议分只读（观察基线）与变更（大 key 拆分、过期时间打散、UNLINK 异步删除、热 key 本地缓存/副本读）两类，变更类标注影响面并 requires_approval；`DEL` 大集合禁用（同步阻塞），建议一律 `UNLINK`。
