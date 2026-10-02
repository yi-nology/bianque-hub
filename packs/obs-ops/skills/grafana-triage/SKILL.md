---
name: grafana-triage
description: Grafana 自诊方法论：面板无数据三层反推（数据源→查询→时间范围）、数据源连通与权限面、插件与容器健康、查询性能面——固定顺序定位与判读基准（ask-ops 只读采集面）。
mode: on_demand
version: 0.1.0
maturity: experimental
requires_mcp:            # 采集依赖的工具面（装载期对账依据；旧 bianque-tools 二进制缺这些工具时应先重建）
  - server: ask-ops
    tools: [run_readonly_command, run_readonly_commands]
---

## 触发条件

- 症状关键词：面板无数据、图表空白、grafana排障、数据源报错、面板报错、dashboard打不开
- 组合场景：数据源侧（Prometheus/ES）本体故障分别转 prometheus-triage / es-triage——本技能先判「锅在哪层」

## 数据来源（ask-ops 只读面）

- Grafana API（GET）：`/api/health`（本体健康）、`/api/datasources`（配置面——**返回体含凭证字段，引用前必须脱敏**）、`/api/datasources/<id>/health`（连通性探测）；
- 面板查询面：具体 dashboard JSON（GET `/api/dashboards/uid/<uid>`）看查询语句与数据源绑定、面板 inspector 报错文案（用户提供或浏览器侧取证）；
- 主机侧：grafana 容器/进程状态、日志尾部（server.log）；
- 凭证由主机侧已配置环境注入，对话不回显。

## 分诊路径（按层反推，固定顺序）

1. **本体健康**：`/api/health` 不通= Grafana 本体问题（容器/进程/入口网络，联动 os-basics / k8s-ops）；通了进第 2 步。
2. **数据源面**：`/api/datasources/<id>/health` 逐个探测 → 失败三分：**地址不通**（URL 配错/网络隔离/数据源服务挂了——转对应组件技能）、**认证失败**（凭证轮换后未更新，改配置是变更建议）、**版本/协议不匹配**（数据源升级后 API 形态变化）。
3. **查询与面板面**：数据源健康但面板空 → 三分：**查询错误**（面板 inspector 的 error 文案：语法/指标不存在——指标缺失常是采集断，联动 prometheus-triage）、**时间范围错配**（面板默认时间窗与数据保留期不符：查 30 天但 retention 只 15 天）、**变量与模板**（dashboard 变量取值为空导致查询集合为空，面板显示 No data 但查询本身正常）。
4. **插件与渲染面**：特定面板类型（插件）渲染失败而查询正常（浏览器 console 证据）；截图/导出功能挂属 image renderer 插件面，不影响数据正确性，降级为体验问题。
5. **查询性能面**：面板能出但极慢 → 慢在数据源侧（查询本身重，联动 prometheus-triage 存储面）vs Grafana 侧（并发查询堆积、容器资源不足）——用「直连数据源同样慢吗」分离变量。

## 判读基准

- 「面板无数据」禁止跳层定论：必须给出「数据源层健康 + 查询层判读」的穿透证据链；
- 单 dashboard 异常 vs 全局异常先分型（全局=数据源/本体面，单点=面板配置面）；
- 时间范围与 retention 的错配要写明两侧数值口径（查的窗口 vs 保留窗口）。

## 输出要求

- 每个结论附 API/日志关键行证据（数据源配置引用一律脱敏）；变更类动作（改数据源配置、升级插件、调 retention）标注影响面并 requires_approval；数据不足输出「需补充采集」清单（如面板 inspector 完整报错、浏览器 console 输出），不臆测。
