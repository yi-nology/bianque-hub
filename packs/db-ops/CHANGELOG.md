# CHANGELOG

## 0.3.0 (2026-10-10)

- **mysql-analyst/pg-analyst 工具授权最小权限收敛**（bianque 批次二百六十四，除 oo-devops 外全 hub 清扫）：
  `ask-ops allow: []`（=全部 14 工具，含破坏性 run_approved_command(s)）改显式 12 工具
  只读白名单（十结构化采集器+probe_host+run_readonly 只读 CLI 对）。修正三处不合理：
  ①授权面与 prompt 能力声明矛盾（分析专家=只读采集判读域，prompt 零变更工具引用，
  处置建议只随报告进审批）；②`allow: []` 使工具集含变更类，agentrun 重试守卫整步骤
  收紧（RetryAfterMutation=false，429 限流拒重试环节 failed 实弹在案，bianque 侧
  agentkit v0.14.6 豁免缝治标、授权收敛治本）；③能力/路由事实源失真——工具清单是
  派工与路由的能力面，超授即误描述。域 CLI 采集面（run_readonly 白名单内 mysql/psql/
  redis-cli/kafka/docker/kubectl 只读子命令等）完整保留，零能力损失。

## 0.2.3 (2026-10-06)

- 深审修复批：pg-triage 凭证口径收敛到 .pgpass 单一通道（PGPASSWORD 前缀形态会连命令
  进审计面，与「对话不回显」纪律相抵）；mysql-triage 容量面「bdir」占位名明确为 binlog
  目录（log_bin 路径）。技能版本补账：mysql-triage/mysql-replication 的 requires_mcp
  系 0.2.0 批次补录（升 0.1.1）、pg-triage 本次口径修正（0.1.1→0.1.2）。经平台 Render
  语义核实（string_list 值逗号分隔、逐值展开成多行命令），pg-kill-idle-backend 模板
  语义无误，不动。
- 215 实弹口径前移：mysql-analyst/pg-analyst 输出铁律补「steps 每条=单条可执行 shell 命令」
  纪律（mw-ops 实弹同型问题的预防性收口；说明性内容归 needs_followup）。
- 215 实弹口径前移：mysql-analyst/pg-analyst max_iterations 12→24（同 mw-ops 实弹
  根因：多面板分诊的诚实采集轮次超过旧预算）。



## 0.2.2 (2026-10-02)

- 批次九十四「诊断+处方+受审执行」首批：新增编目变更块 `pg-kill-idle-backend`（终止 idle in transaction 后端，params: pids 清单）；pg-triage 技能 frontmatter 登记 `provides_changes` 并在输出要求里给 change_ref 处方口径（编目优先、自由 steps 兜底），技能升版 0.1.1。

## 0.2.1 (2026-10-02)

- 0.2.0 条目数字按包内实情改写：本包为 3 技能、2 专家（原条目误抄批次合计「14 技能/六位专家」）。

## 0.2.0 (2026-10-02)

- 3 技能（mysql-triage/mysql-replication/pg-triage）补 requires_mcp 声明
  （ask-ops：run_readonly_command/run_readonly_commands）——
  采集依赖显式化，装载期可对账（旧 bianque-tools 二进制缺工具时先重建再装包）。
- 两位专家（mysql-analyst/pg-analyst）新增「采集脱敏」纪律：docker inspect Env / kubectl describe 环境变量等
  凭证面字段引用前打码；Secret 材料面（kubectl get secrets、config view --raw）
  已被工具层拒收，禁绕过。

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
