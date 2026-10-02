# CHANGELOG

## 0.3.1 (2026-10-02)

- 补 loki-triage 接线（0.3.0 遗漏，技能已发布但无入口可达）：obs-analyst 挂载该技能、
  tools 增授 datasources（loki_query_range/loki_labels）、提示词职责/采集纪律/判读依据
  补 Loki 面、路由词补 loki排障/日志断流/logql查询失败。
- 0.2.0 条目数字按包内实情改写：本包当时为 3 技能、1 专家（原条目误抄批次合计）。

## 0.3.0 (2026-10-02)

- 新增 loki-triage 技能：Loki 日志分诊方法论（标签面断流/查询空结果三分型/429 限频/store 压力），
  采集面钉 datasources server 新工具 loki_query_range/loki_labels（配套平台批次九十：datasources
  扩五源）。

## 0.2.0 (2026-10-02)

- 3 技能（prometheus-triage/grafana-triage/es-triage）补 requires_mcp 声明
  （ask-ops：run_readonly_command/run_readonly_commands）——
  采集依赖显式化，装载期可对账（旧 bianque-tools 二进制缺工具时先重建再装包）。
- obs-analyst 专家新增「采集脱敏」纪律：docker inspect Env / kubectl describe 环境变量等
  凭证面字段引用前打码；Secret 材料面（kubectl get secrets、config view --raw）
  已被工具层拒收，禁绕过。

## 0.1.3 (2026-10-02)

- 采集契约升级：例检面板改用 ask-ops 批量形态 `run_readonly_commands`（≤10 条
  单会话收口、逐命令退出码、整批先审后发）。

## 0.1.2 (2026-10-02)

- 处方修正（对齐 run_readonly_command 引号感知守卫）：harbor-triage 证书探测去
  `2>/dev/null`（stderr 本就不采集，重定向会被拒）；es-triage `_cat` URL 带 `&`
  查询参数时给引号形态示范。

## 0.1.1 (2026-10-02)

- 采集契约对齐平台 ask-ops 新工具 `run_readonly_command`（受审只读命令：白名单
  + 注入防线 + 尾部截断/退出码），专家 prompt 点名调用方式。

## 0.1.0 (2026-10-02)

- 首发（oo-devops 拆分重构第一批）：prometheus-triage（合并重写）、grafana-triage
  （重写）、es-triage（合并重写）+ imports/obs-analyst 专家（ask-ops 只读面，
  与平台监控查询入口路由错位）。
