# CHANGELOG

## 0.2.3 (2026-10-06)

- 深审修复批：mw-quick-audit 链第二步 instruction 改「按部署形态分面板」（原钉扎 kafka-triage
  却让 RabbitMQ 环境也收到 Kafka 分诊路径——方法论错配）；mw-redis-unlink-key 编目补
  REDISCLI_AUTH 执行侧凭证载体约定（requirepass 实例原模板 NOAUTH 失败；禁 -a 裸密码）；
  kafka-triage 两处错字（症状/目录）并升 0.1.2；redis-triage 版本补账 0.1.1（requires_mcp
  系 0.2.0 批次补录未升版）。
- 215 实弹修复：cache-analyst/mq-analyst 输出铁律补「steps 每条=单条可执行 shell 命令」
  纪律（实测 mq-analyst 报告 steps 带中文说明被平台 schema 拒收，好诊断被兜底成失败报告；
  说明性内容归 needs_followup——对齐 n8n-ops 0.2.1 实弹教训口径）。
- 215 实弹修复：cache-analyst/mq-analyst max_iterations 12→24（实测多面板采集+委派
  路径把 12 轮预算打爆，agentrun exceeds max iterations → 整会话失败兜底；上限只在
  需要时消耗，不影响短会话）。



## 0.2.2 (2026-10-03)

- 批次九十四「诊断+处方+受审执行」接入：新增编目变更块三件——`mw-redis-unlink-key`
  （大 key 异步删除，params: host/port/key，verify: EXISTS）、`mw-rabbitmq-purge-queue`
  （清空队列，params: queue，verify: list_queues 归零；仅限积压可丢弃场景）、
  `mw-kafka-reset-offsets-latest`（位点重置到最新跳过积压，params: brokers/group，
  verify: --describe）；三技能 frontmatter 登记 `provides_changes` 并在输出要求给
  change_ref 处方口径（技能各升 0.1.1）；cache/mq 双专家提示词处方纪律升为编目优先。

## 0.2.1 (2026-10-02)

- 0.2.0 条目数字按包内实情改写：本包为 4 技能、2 专家（原条目误抄批次合计「14 技能/六位专家」）。

## 0.2.0 (2026-10-02)

- 4 技能（redis-triage/redis-hotkey-bigkey/kafka-triage/rabbitmq-triage）补 requires_mcp 声明
  （ask-ops：run_readonly_command/run_readonly_commands）——
  采集依赖显式化，装载期可对账（旧 bianque-tools 二进制缺工具时先重建再装包）。
- 两位专家（cache-analyst/mq-analyst）新增「采集脱敏」纪律：docker inspect Env / kubectl describe 环境变量等
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
