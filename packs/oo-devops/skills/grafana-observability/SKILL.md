---
name: grafana-observability
description: Grafana 可观测性平台 — 仪表板、Prometheus PromQL、Loki LogQL、告警、事故、OnCall 排班、注释、数据源查询、面板渲染（75+ 工具）。在查询 Grafana 仪表板、为接口指标运行 PromQL、为 syslog 事件搜索 Loki 日志、调查触发告警或检查谁正在值班时使用。
mode: on_demand
version: 1.1.0
maturity: experimental
---

# Grafana 可观测性

## MCP 服务器

| 属性 | 值 |
|----------|-------|
| **Source** | [grafana/mcp-grafana](https://github.com/grafana/mcp-grafana) |
| **Transport** | stdio (默认), SSE, 或 streamable-http |
| **Language** | Go (通过 `uvx mcp-grafana` 运行) |
| **Tools** | 75+ (仪表板、Prometheus、Loki、告警、事故、OnCall、注释、管理员) |
| **Auth** | 服务账户令牌 (首选) 或 用户名/密码 |
| **Requires** | Grafana 9.0+, 具有 Editor 角色或细粒度 RBAC 的服务账户 |

## 如何运行

```bash
# stdio 模式 (默认 — NetClaw 使用)
uvx mcp-grafana

# 只读模式 (防止仪表板/告警修改)
uvx mcp-grafana --disable-write
```

## 环境变量

| 变量 | 必需 | 示例 | 描述 |
|----------|----------|---------|-------------|
| `GRAFANA_URL` | 是 | `http://grafana.example.com:3000` | Grafana 实例 URL |
| `GRAFANA_SERVICE_ACCOUNT_TOKEN` | 是* | `glsa_abc123...` | 服务账户令牌 (首选认证方式) |
| `GRAFANA_USERNAME` | 替代 | `admin` | 基本认证用户名 (令牌替代方案) |
| `GRAFANA_PASSWORD` | 替代 | `changeme` | 基本认证密码 |
| `GRAFANA_ORG_ID` | 否 | `1` | 多组织设置的 组织 ID |

*需要服务账户令牌或用户名/密码之一。

## 主要工具类别

### 仪表板操作

| 工具 | 功能 |
|------|-------------|
| `search_dashboards` | 按标题或元数据查找仪表板 |
| `get_dashboard_summary` | 轻量级概览 (上下文高效 — 首先使用这个) |
| `get_dashboard_by_uid` | 完整仪表板 JSON (较大 — 谨慎使用) |
| `get_dashboard_property` | 通过 JSONPath 提取特定字段 |
| `get_dashboard_panel_queries` | 提取面板查询详情 |
| `update_dashboard` | 创建或修改仪表板 |
| `patch_dashboard` | 有针对性的修改，无需完整 JSON 替换 |

### Prometheus (PromQL)

| 工具 | 功能 |
|------|-------------|
| `query_prometheus` | 执行即时或范围 PromQL 查询 |
| `list_prometheus_metric_names` | 发现可用指标 |
| `list_prometheus_label_names` | 列出与选择器匹配的标签 |
| `list_prometheus_label_values` | 检索特定标签的值 |
| `query_prometheus_histogram` | 计算百分位数 (p50, p90, p95, p99) |
| `list_prometheus_metric_metadata` | 指标类型、帮助文本、单位 |

### Loki (LogQL)

| 工具 | 功能 |
|------|-------------|
| `query_loki_logs` | 针对日志流执行 LogQL 查询 |
| `list_loki_label_names` | 发现可用的日志标签 |
| `list_loki_label_values` | 列出特定日志标签的值 |
| `query_loki_stats` | 流统计 (容量、速率) |
| `query_loki_patterns` | 检测日志结构模式 |

### 告警

| 工具 | 功能 |
|------|-------------|
| `list_alert_rules` | 查看所有 Grafana 和数据源管理的告警规则 |
| `get_alert_rule_by_uid` | 检索特定告警规则详情 |
| `create_alert_rule` | 创建新告警规则 |
| `update_alert_rule` | 修改现有告警规则 |
| `delete_alert_rule` | 删除告警规则 |
| `list_contact_points` | 查看通知端点 (email, Slack, PagerDuty, 等.) |

### 事故管理

| 工具 | 功能 |
|------|-------------|
| `list_incidents` | 查看带过滤的 Grafana 事故 |
| `get_incident` | 单个事故详情 |
| `create_incident` | 创建新事故 |
| `add_activity_to_incident` | 向事故添加时间线条目 |

### OnCall

| 工具 | 功能 |
|------|-------------|
| `list_oncall_schedules` | 查看值班轮班排班表 |
| `get_oncall_shift` | 班次详情 |
| `get_current_oncall_users` | 谁现在正在值班 |
| `list_alert_groups` | 带过滤的 OnCall 告警组 |

### 注释与渲染

| 工具 | 功能 |
|------|-------------|
| `get_annotations` | 查询带时间/标签过滤器的注释 |
| `create_annotation` | 向仪表板/面板添加注释 |
| `get_panel_image` | 将面板或仪表板渲染为 PNG 图像 |
| `generate_deeplink` | 创建准确的 Grafana URL 用于分享 |

### 调查 (Sift)

| 工具 | 功能 |
|------|-------------|
| `list_sift_investigations` | 列出自动调查 |
| `get_sift_investigation` | 调查详情 |
| `find_error_pattern_logs` | 检测日志中错误模式升高 |
| `find_slow_requests` | 通过 Tempo 跟踪识别慢请求 |

---

## 工作流: 网络基础设施监控

在 Grafana 中检查网络设备指标时:

1. **查找仪表板**: `search_dashboards` 使用关键词 (例如, "network", "interface", "BGP")
2. **仪表板概览**: `get_dashboard_summary` 用于面板列表，无需完整 JSON
3. **查询指标**: `query_prometheus` 使用 PromQL 查询特定指标:
   - 接口流量: `rate(ifHCInOctets{instance="router1"}[5m]) * 8`
   - BGP 对等状态: `bgp_peer_state{peer="10.1.1.2"}`
   - CPU 利用率: `device_cpu_utilization{device="core-rtr-01"}`
   - 接口错误: `increase(ifInErrors{device=~".*"}[1h])`
4. **检查告警**: `list_alert_rules` 查看活动告警阈值
5. **搜索日志**: `query_loki_logs` 用于 syslog 或 SNMP trap 数据
6. **报告**: 指标摘要与告警状态和日志关联
7. **GAIT**: 在审计跟踪中记录所有查询

### 示例: 接口利用率检查

```
search_dashboards(title="Network Interfaces")
get_dashboard_summary(uid="abc123")
query_prometheus(expr="rate(ifHCInOctets{device='core-rtr-01'}[5m]) * 8", time_range="1h")
query_prometheus(expr="rate(ifHCOutOctets{device='core-rtr-01'}[5m]) * 8", time_range="1h")
list_alert_rules(folder="Network")
```

## 工作流: 告警调查

在调查 Grafana 告警时:

1. **列出告警**: `list_alert_rules` — 查找触发或待处理的规则
2. **告警详情**: `get_alert_rule_by_uid` — 阈值、条件、数据源
3. **查询指标**: `query_prometheus` — 检查触发告警的指标
4. **搜索日志**: `query_loki_logs` — 关联告警时间附近的日志事件
5. **检查事故**: `list_incidents` — 这是否已被跟踪?
6. **联系点**: `list_contact_points` — 验证通知路由
7. **报告**: 告警分析与根因和指标证据

## 工作流: 事故响应

在响应 Grafana 事故时:

1. **列出血事故**: `list_incidents` — 查找开放事故
2. **事故详情**: `get_incident` — 时间线、严重度、标签
3. **OnCall**: `get_current_oncall_users` — 谁应该被通知
4. **关联指标**: `query_prometheus` — 检查受影响的服务指标
5. **关联日志**: `query_loki_logs` — 查找事故时间附近的错误模式
6. **调查**: `find_error_pattern_logs` — 自动错误模式检测
7. **更新事故**: `add_activity_to_incident` — 将发现添加到时间线
8. **注释**: `create_annotation` — 在相关仪表板上标记事件

## 工作流: 日志分析

在调查存储在 Loki 中的网络日志时:

1. **发现标签**: `list_loki_label_names` — 查找可用标签 (host, severity, facility)
2. **标签值**: `list_loki_label_values` — 枚举主机、严重级别
3. **查询日志**: `query_loki_logs` 使用 LogQL:
   - 按设备: `{host="core-rtr-01"}`
   - 按严重度: `{host="core-rtr-01"} |= "error"`
   - 模式匹配: `{job="syslog"} |~ "BGP|OSPF"`
4. **模式**: `query_loki_patterns` — 检测重复日志结构
5. **统计**: `query_loki_stats` — 日志容量和速率分析

---

## 与其他技能集成

| 技能 | 集成 |
|-------|-------------|
| **pyats-health-check** | 将 pyATS 健康数据与 Grafana 指标和仪表板交叉引用 |
| **pyats-routing** | 将 OSPF/BGP 状态变化与 Grafana 指标时间线关联 |
| **gait-session-tracking** | 在 GAIT 审计跟踪中记录所有 Grafana 查询和发现 |
| **slack-network-alerts** | Grafana 告警通过 Slack + NetClaw 进行自动调查 |
| **servicenow-change-workflow** | 在变更窗口期间注释 Grafana 仪表板; 将事故与 CR 关联 |
| **te-network-monitoring** | 将 ThousandEyes 路径数据与 Grafana 基础设施指标配对 |
| **aws-cloud-monitoring** | 将 Grafana 仪表板与 CloudWatch 数据进行混合可见性比较 |
| **markmap-viz** | 将 Grafana 告警规则层次结构可视化为思维导图 |

---

## 上下文窗口管理

Grafana 仪表板可能是大型 JSON 文档。使用这些策略:

1. **始终从 `get_dashboard_summary` 开始** — 轻量级概览，不是完整 JSON
2. **使用 `get_dashboard_property`** 配合 JSONPath 获取特定字段
3. **避免 `get_dashboard_by_uid`** 除非你需要完整的仪表板定义
4. **使用 `get_dashboard_panel_queries`** 仅提取查询定义

---

## 重要规则

- **首选只读操作** — 在进行任何写入操作之前使用 `search_dashboards`, `get_dashboard_summary`, `query_prometheus`, `query_loki_logs`, `list_alert_rules`
- **仪表板修改需要 ServiceNow CR** — 除非在实验室/开发 Grafana 实例中
- **告警规则变更需要批准** — 创建/更新/删除告警规则会影响生产监控
- **令牌高效查询** — 使用 `get_dashboard_summary` 替代 `get_dashboard_by_uid`, 使用时间范围限制 Prometheus/Loki 结果大小
- **GAIT 审计强制** — 记录所有 Grafana 查询、仪表板修改、告警变更和事故更新
- **查询中无秘密** — 切勿在 PromQL/LogQL 表达式中嵌入凭据或敏感数据

## 错误处理

- **认证失败 (401/403)**: 检查 `~/.openclaw/.env` 中的 `GRAFANA_URL` 和 `GRAFANA_SERVICE_ACCOUNT_TOKEN`。验证服务账户是否具有 Editor 角色或所需的 RBAC 权限。
- **数据源未找到**: 使用 `list_datasources` 发现可用的数据源 UID 和名称。
- **PromQL/LogQL 错误**: 在查询前使用 `list_prometheus_metric_names` 或 `list_loki_label_names` 发现有效的指标/标签名称。
- **仪表板未找到**: 在使用基于 UID 的工具之前，使用 `search_dashboards` 按标题查找仪表板。
- **速率限制**: Grafana 可能会对 API 请求进行速率限制; 间隔大型查询批次。
