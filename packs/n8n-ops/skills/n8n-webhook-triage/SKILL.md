---
name: n8n-webhook-triage
description: n8n webhook 触发面分诊方法论：404 not registered 三查（active 状态/URL 语义/重注册）、Forbidden maybe CSRF 反代配置面、test 与 production URL 注册语义、queue 模式下 webhook 路由错位——固定顺序定位与判读基准（datasources n8n 查询面 + ask-ops 只读采集面）。
mode: on_demand
version: 0.1.1
maturity: experimental
requires_mcp:            # 采集依赖的工具面（装载期对账依据；旧 bianque-tools 二进制缺这些工具时应先重建）
  - server: datasources
    tools: [n8n_list_workflows, n8n_get_workflow]
  - server: ask-ops
    tools: [run_readonly_command, run_readonly_commands]
---

## 触发条件

- 症状关键词：webhook 不触发、webhook 404、The requested webhook is not registered、test 通 production 不通、Forbidden maybe CSRF、登录后操作被拒、回调收不到
- 组合场景：对端系统侧的出网问题（回调没发出来）不在本技能——先经平台网络域确认「请求到没到 n8n」再进本技能

## 数据来源（datasources n8n 查询面 + ask-ops 主机面）

- **工作流状态面走 `datasources` 工具面**（`n8n_list_workflows` 看 active 分布、
  `n8n_get_workflow` 看单工作流定义与 Webhook 节点的 path/method 配置；API key 经凭证面
  注入，对话不收集/回显）；未配置 n8n 源时工具报「未配置」——转凭证面补 n8n_url +
  n8n_api_key，不臆测；
- 探测面（ask-ops）：`curl -s -i '<webhook-url>'`（原样回显状态行；URL 带 `&`/`?` 一律引号形态；
  405 Method Not Allowed 本身说明路径已注册，是**健康信号**）；
- 配置面（ask-ops）：`docker inspect <容器>` 读 env 键名清单（`N8N_HOST`/`N8N_PROTOCOL`/
  `N8N_WEBHOOK_URL`/`N8N_PROXY_HOPS` 现值——值非敏感可引用，凭证类值打码）；
- 日志面（ask-ops）：`docker logs --tail 200 <容器>`（webhook 注册/404 记录）。

## 分诊路径（固定顺序，404 三查 → 注册面 → CSRF → 路由面）

1. **404「not registered」三查**（高频报错原文：The requested webhook is not registered）：
   ① workflow 是否 **active**——production URL（`/webhook/*`）只在激活态注册，test URL
   （`/webhook-test/*`）只在编辑器「Listen for test event」期间临时注册，**test 通而
   production 404 的第一嫌疑就是没激活**；② URL 与 method 是否用对（production 路径/
   POST vs GET）；③ 用错端点前缀（`N8N_ENDPOINT_WEBHOOK*` 改过路径的情况）。
2. **active=true 仍 404**：属已知 bug 形态（active 但未注册，多见于升级/迁移后）——按序
   恢复：停用再激活 → 重启实例 → 经 API 完整 PUT workflow 强制重注册；v1→v2 迁移不干净
   时需重建 Webhook 节点。全部是变更动作：进建议面 requires_approval。
3. **Forbidden — maybe CSRF**：先核对配置面三件套与实际访问 URL 是否一致——
   `N8N_HOST`（反代后必须是对外域名）、`N8N_PROTOCOL`、`N8N_WEBHOOK_URL`（webhook 对外
   基础 URL；旧名 `WEBHOOK_URL` 已废弃）；Cloudflare Flexible SSL 类协议错位与代理未转发
   `X-Forwarded-Proto` 是高频根因，多级反代补 `N8N_PROXY_HOPS`。纯 HTTP 内网环境
   `N8N_SECURE_COOKIE`（默认 true，仅 HTTPS 传 cookie）会把人锁在登录页——如实施例属
   内网明文形态，列入建议面并注明安全代价。
4. **queue 模式路由错位**：LB 必须把 `/webhook/*`、`/webhook-waiting/*` 路由到 webhook
   processor（或未设 processor 时由 main 承接），`/webhook-test/*` **必须**路由到 main
   ——test 请求进了 worker/processor 就是 404/不响应；`N8N_DISABLE_PRODUCTION_MAIN_PROCESS=true`
   时 main 不再接生产 webhook，LB 分流规则必须同步，否则整体 404。

## 判读基准

- 「不触发」先分**请求没到**（网络/对端面，出本技能）与**到了被拒**（404/403，本技能）
  ——curl 探测状态行是分界证据；
- test/production 的注册语义差异是判读基石：test 通不能作为 production 健康的证据；
- 405 响应 = 路径在注册表里（换 method 即可），与 404（未注册）严格区分；
- webhook 生产端点是无认证公网入口（认证靠工作流自身设计），暴露面收紧建议（网关侧
  限源/鉴权）注明这是安全加固而非故障修复。

## 输出要求

- 每个结论附探测状态行/API active 字段/env 现值证据；变更类动作（停启激活、重启实例、
  改 env、调整 LB 分流）标注影响面并 requires_approval；数据不足输出「需补充采集」清单
  （如反代配置原文、LB 分流规则、完整 webhook URL 样本），不臆测。
