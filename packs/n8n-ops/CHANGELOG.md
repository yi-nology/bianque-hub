# CHANGELOG

## 0.3.0 (2026-10-10)

- **n8n-platform-analyst 工具授权最小权限收敛**（bianque 批次二百六十四，除 oo-devops 外全 hub 清扫）：
  `ask-ops allow: []`（=全部 14 工具，含破坏性 run_approved_command(s)）改显式 12 工具
  只读白名单（十结构化采集器+probe_host+run_readonly 只读 CLI 对）。修正三处不合理：
  ①授权面与 prompt 能力声明矛盾（分析专家=只读采集判读域，prompt 零变更工具引用，
  处置建议只随报告进审批）；②`allow: []` 使工具集含变更类，agentrun 重试守卫整步骤
  收紧（RetryAfterMutation=false，429 限流拒重试环节 failed 实弹在案，bianque 侧
  agentkit v0.14.6 豁免缝治标、授权收敛治本）；③能力/路由事实源失真——工具清单是
  派工与路由的能力面，超授即误描述。域 CLI 采集面（run_readonly 白名单内 mysql/psql/
  redis-cli/kafka/docker/kubectl 只读子命令等）完整保留，零能力损失。

## 0.2.3 (2026-10-06)

- 深审修复批：n8n-container-restart 编目 title 错字（task biner→task runner）；
  n8n-execution-triage「running 不结束」补 queue 模式关键口径（超时由 worker 进程
  执行，env 只配 main 不配 worker 是「超时不生效」高频根因，0.1.1→0.2.1）。

## 0.2.2 (2026-10-03)

- 批次九十四「诊断+处方+受审执行」接入：新增编目变更块 `n8n-container-restart`（实例/worker
  容器重启，params: container，verify: docker ps --filter；前置纪律：先有根因结论，重启只是
  让处置生效）；n8n-instance-triage 升 0.1.1、n8n-queue-mode-triage 升 0.1.2，frontmatter
  登记 `provides_changes` 并在输出要求给 change_ref 处方口径；专家提示词处方纪律升为编目
  优先（docker exec 类 n8n CLI 与生命周期大动作仍走自由 steps）。

## 0.2.1 (2026-10-03)

- 专家 prompt 输出契约对齐平台 schema（实弹 e2e 发现）：recommendation.steps 每条必须是
  单条可执行 shell 命令（schema 白名单校验拦说明性内容——0.2.0 首轮实弹六条 steps 全被
  拒收重试失败），影响面/回退路径等说明一律移 needs_followup。

## 0.2.0 (2026-10-03)

- 采集面升级：n8n 公共 API 事实改走 datasources 新源四工具（n8n_list_workflows /
  n8n_get_workflow / n8n_list_executions / n8n_get_execution，配套平台批次：datasources
  server 增 n8n 源、凭证面增 n8n_url/n8n_api_key 字段）——API key 经凭证面注入，
  替代 0.1.0 的主机侧 `$N8N_API_KEY` env 约定；四技能 requires_mcp 显式声明、专家
  tools 授权、prompt 采集纪律同步（execution/webhook/queue-mode/lifecycle 升 0.1.1，
  instance-triage 采集面未变保持 0.1.0）。
- ask-ops 面收敛为主机侧事实：健康端点、/metrics、docker/kubectl 面、webhook URL 探测。

## 0.1.0 (2026-10-02)

- 首发（原生编写，非 oo-devops 血缘）：n8n-instance-triage / n8n-execution-triage /
  n8n-webhook-triage / n8n-queue-mode-triage / n8n-lifecycle-ops 五技能 +
  imports/n8n-platform-analyst 专家（ask-ops 只读面，与 cicd/mw/obs 相邻域路由错位）。
- 方法论基于 n8n 官方自托管文档（docs.n8n.io /deploy/host-n8n 系列）与社区排障案例
  （community.n8n.io）整理，收录 2.x 稳定线与 3.0 破坏性变更基线（task runners 默认、
  MySQL 移除、npm 装法退役、binaryData→storage 更名）、健康端点两级语义、队列模式
  消费者四同核、webhook 404 三查、加密密钥/备份三件套纪律。
