# CHANGELOG

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
