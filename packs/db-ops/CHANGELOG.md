# CHANGELOG

## 0.1.3 (2026-10-02)

- 采集契约升级：例检面板改用 ask-ops 批量形态 `run_readonly_commands`（≤10 条
  单会话收口、逐命令退出码、整批先审后发）。

## 0.1.2 (2026-10-02)

- pg-triage 补 sudo -u postgres 采集路径提示（对齐 run_readonly_command 的目标
  用户切换支持与 mysql -p/psql -W 交互挂起守卫）。

## 0.1.1 (2026-10-02)

- 采集契约对齐平台 ask-ops 新工具 `run_readonly_command`（受审只读命令：白名单
  + 注入防线 + 尾部截断/退出码），专家 prompt 点名调用方式。

## 0.1.0 (2026-10-02)

- 首发（oo-devops 拆分重构第一批）：mysql-triage（合并重写）、mysql-replication
  （重写）、pg-triage（重写）+ imports/mysql-analyst、imports/pg-analyst 双专家
  + workflow/db-quick-audit 数据库例检链（技能钉扎）。
