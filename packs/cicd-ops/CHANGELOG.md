# CHANGELOG

## 0.2.1 (2026-10-02)

- 0.2.0 条目数字按包内实情改写：本包为 4 技能、1 专家（原条目误抄批次合计「14 技能/六位专家」）。

## 0.2.0 (2026-10-02)

- 4 技能（jenkins-pipeline-triage/gitlab-ci-triage/harbor-triage/argocd-sync-triage）补 requires_mcp 声明
  （ask-ops：run_readonly_command/run_readonly_commands）——
  采集依赖显式化，装载期可对账（旧 bianque-tools 二进制缺工具时先重建再装包）。
- pipeline-analyst 专家新增「采集脱敏」纪律：docker inspect Env / kubectl describe 环境变量等
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

- 首发（oo-devops 拆分重构第一批）：jenkins-pipeline-triage / gitlab-ci-triage /
  harbor-triage / argocd-sync-triage 四技能（收割重写）
  + imports/pipeline-analyst 专家（ask-ops 只读面）。
