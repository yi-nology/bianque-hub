# CHANGELOG

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
