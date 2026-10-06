---
name: loki-triage
description: Loki 日志分诊方法论：标签面断流归因、查询空结果三分型、429 限频与语法错判读、store/ingester 后端压力——固定顺序定位与判读基准（datasources Loki 查询面）。
mode: on_demand
version: 0.1.1
maturity: experimental
requires_mcp:            # 采集依赖的工具面（装载期对账依据；旧 bianque-tools 二进制缺这些工具时应先重建）
  - server: datasources
    tools: [loki_query_range, loki_labels]
  - server: ask-ops      # 主机侧补充采集（采集器进程/配置、存储端指标 curl 补采）；obs-analyst 已授权
    tools: [run_readonly_command, run_readonly_commands]
---

## 触发条件

- 症状关键词：日志查不到、日志断流、loki 超时、loki 429、日志检索慢、loki排障、日志空结果
- 组合场景：应用层异常先归业务域包，本技能只管「日志看不看得到、查不查得动」；Grafana
  Explore 无数据先经 grafana-triage 排除展示层，再回本技能查 Loki 查询面

## 数据来源（datasources Loki 只读查询面）

- `loki_labels`：列当前可用标签名——**一切查询前先跑它**，标签面是 Loki 分诊的第一现场；
- `loki_query_range`：LogQL 流查询（`query` 必填；`since` 缺省 1h；`limit` 缺省 100，
  上限 1000）。结果按流（stream+values）返回，空结果与报错信封都是结论素材；
- 未配置 Loki 源时工具报「未配置（BQ_DATASOURCES_CONFIG）」——转凭证面补 loki_url，
  不臆测后端状态；
- Loki 本体主机侧指标（ingester/store/内存）需要时走 ask-ops 只读命令 `curl <loki>/metrics`
  补充，凭证由主机侧已配置环境注入，对话不回显。

## 分诊路径（固定顺序，先标签后查询再后端）

1. **标签面**：`loki_labels` 返回空或缺预期标签（如 job/app）→ 日志**根本没进来**：
   采集端（promtail/alloy/otel-collector 系）没在送——转主机面查采集器进程与配置（ask-ops），
   本技能止步于「标签缺失=采集断」，不猜采集端原因。
2. **查询空结果三分型**：标签在但 `query_range` 空结果，按序排查——
   **选择器写错**（用了不存在的标签/值；对照 labels 输出修正）→ **时间窗错**（`since`
   给太窄、或日志时间戳与当前时钟漂移）→ **retention 切掉**（窗内数据已被过期清理，
   询问保留策略口径）。空结果必须写清用了哪个选择器+哪个时间窗，禁笼统说「没日志」。
3. **报错面**：HTTP 429（`rate limit` / `too many outstanding requests`）= 限频，**退避
   重试而非报错**——先窄化选择器+行过滤（`|= error`）降量，仍 429 才上报调 limits 配置；
   400 带解析错误行 = LogQL 语法错（管道/正则引号最常见），按报错行修正重试。
4. **后端压力面**：查询持续超时/5xx、大窗必挂小窗能过 = store/ingester 压力——看
   对象存储/文件系统水位、`loki_ingester` 相关 metrics（ask-ops 补采）；compactor 卡死
   与 retention 堆积是常见根因。后端扩容/配置变更标注影响面，requires_approval。

## 判读基准

- 高流量流**先窄化再拉**：先 `{job="x"} |= "关键词"` 小窗试探，确认有料再放宽窗——
  一次全量拉爆的不只是 limit，还有下游上下文预算；
- 「日志断流」结论必须区分**写入断**（labels 缺失/流停在旧时间戳）与**查询断**（labels
  在、查询报错）——两者处置方向完全不同（采集端 vs 后端）；
- 时间戳口径：Loki 按摄入时间索引、按事件时间戳排序，时钟漂移的采集器会产生「未来/
  过去」日志，查询窗对不上时先核对时间戳分布再下结论；
- 数量级口径：limit 100 拿满=流很大，不是异常本身；判断异常看**模式**（错误率突增、
  特定标签集中）而非单条。

## 输出要求

- 每个结论附 LogQL 查询原文+关键返回行证据；「查不到」结论必须带标签/时间窗口径；
- 变更类动作（调 limits、扩存储、改 retention、动采集器配置）标注影响面并
  requires_approval；数据不足输出「需补充采集」清单（如采集器配置原文、loki limits 段），
  不臆测。
