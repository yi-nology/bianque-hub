# CHANGELOG

## 0.2.0 (2026-10-02)

- 14 技能补 requires_mcp 声明（ask-ops：run_readonly_command/run_readonly_commands）——
  采集依赖显式化，装载期可对账（旧 bianque-tools 二进制缺工具时先重建再装包）。
- 六位专家新增「采集脱敏」纪律：docker inspect Env / kubectl describe 环境变量等
  凭证面字段引用前打码；Secret 材料面（kubectl get secrets、config view --raw）
  已被工具层拒收，禁绕过。

## 0.1.3 (2026-10-02)

- 采集契约升级：例检面板改用 ask-ops 批量形态 `run_readonly_commands`（≤10 条
  单会话收口、逐命令退出码、整批先审后发）。

## 0.1.2 (2026-10-02)

- mq-analyst 症状词「发布超时」改「消息发布超时」：与 cicd-ops 的发布流水线语义
  同优先级撞词（真歧义，平台装载会失败）——校验器新规则（跨包同优先级路由词/
  症状重复拦截）首发命中即修复。

## 0.1.1 (2026-10-02)

- 采集契约对齐平台 ask-ops 新工具 `run_readonly_command`（受审只读命令：白名单
  + 注入防线 + 尾部截断/退出码），专家 prompt 点名调用方式。

## 0.1.0 (2026-10-02)

- 首发（oo-devops 拆分重构第一批）：redis-triage（三技能合并重写）、
  redis-hotkey-bigkey（原创）、kafka-triage / rabbitmq-triage（重写去工具耦合）
  + imports/cache-analyst、imports/mq-analyst 双专家
  + workflow/mw-quick-audit 中间件例检链（技能钉扎）。
