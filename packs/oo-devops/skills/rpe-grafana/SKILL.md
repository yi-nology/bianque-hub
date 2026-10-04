---
name: rpe-grafana
description: 无需了解底层查询即可从 Grafana 仪表板读取当前值。使用场景：询问 Grafana 仪表板中可见的值（传感器读数、指标、统计信息）。通过仪表板和面板名称导航——无需 PromQL/SQL。不适用于：写入 Grafana、管理操作或原始查询执行。
mode: on_demand
version: 1.1.0
maturity: experimental
---

# Grafana 技能

无需编写查询即可从任何 Grafana 仪表板读取当前值。该插件通过仪表板和面板名称导航，提取面板的现有查询配置，并返回简洁摘要——无需 PromQL、SQL 或数据源知识。

适用于任何 Grafana 数据源（Prometheus、InfluxDB、MySQL 等）。

## 使用场景

✅ **使用此技能的场景：**

- 询问 Grafana 仪表板中可见的值
- 列出可用的仪表板或面板
- 通过面板名称检索指标的当前值或最近值

## 不使用场景

❌ **不要使用此技能的场景：**

- 写入、修改或创建仪表板 → 使用 Grafana UI
- 管理操作（用户、数据源配置、告警）→ 直接使用 Grafana API
- 需要执行不受现有面板支持的任意查询

## 设置

添加到您的 `openclaw.json`：

```json
{
  "plugins": {
    "entries": {
      "rpe-grafana": {
        "enabled": true,
        "config": {
          "url": "http://your-grafana:3000",
          "user": "your-username",
          "password": "your-password"
        }
      }
    }
  }
}
```

或设置环境变量：

- `GRAFANA_URL` - Grafana 基础 URL
- `GRAFANA_USER` - 用户名
- `GRAFANA_PASSWORD` - 密码或 API 密钥

## 工具

### grafana_list_dashboards

列出所有可用仪表板。

**参数：** 无

**返回：** `[{ uid, title }]`

### grafana_list_panels

列出仪表板中的所有面板。

**参数：**
- `dashboard_uid`（必需）- 来自 `grafana_list_dashboards` 的仪表板 UID

**返回：** `[{ id, title }]`

### grafana_query_panel

读取特定面板的当前数据。从仪表板获取面板的查询配置并通过 Grafana 的数据源 API 执行——无需查询语言知识。

**参数：**
- `dashboard_uid`（必需）- 仪表板 UID
- `panel_id`（必需）- 来自 `grafana_list_panels` 的面板 ID
- `from`（可选）- 时间范围开始（默认：`now-1h`）
- `to`（可选）- 时间范围结束（默认：`now`）

**返回：** `[{ refId, name, lastValue, unit }]`

## 典型工作流程

1. `grafana_list_dashboards` → 查找仪表板 UID
2. `grafana_list_panels` → 通过标题查找面板 ID
3. `grafana_query_panel` → 获取当前值

## 注意事项

- 需要具有读取权限的 Grafana 用户（Viewer 角色即可）
- 仪表板 UID 是稳定的标识符；面板 ID 在仪表板内唯一
- 行面板会自动扁平化
