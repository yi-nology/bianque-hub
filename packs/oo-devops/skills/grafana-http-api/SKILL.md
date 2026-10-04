---
name: grafana-http-api
description: 与Grafana的HTTP API交互的综合skill，用于管理dashboards、data sources、folders、alerting、annotations、users、teams和organizations。当Claude需要(1)创建、读取、更新或删除Grafana dashboards，(2)管理data sources和connections，(3)配置alerting rules、contact points和notification policies，(4)处理folders和permissions，(5)管理users、teams和service accounts，(6)创建或查询annotations，(7)对data sources执行queries，或任何其他通过API进行的Grafana自动化任务时使用。
mode: on_demand
version: 1.1.0
maturity: experimental
---

# Grafana HTTP API Skill

以编程方式管理Grafana资源，包括dashboards、data sources、alerting、folders、annotations、users、teams和organizations。

## Authentication

### Service Account Token (推荐)

```bash
curl -H "Authorization: Bearer <SERVICE_ACCOUNT_TOKEN>" \
     -H "Content-Type: application/json" \
     https://your-grafana.com/api/dashboards/home
```

### Basic Auth

```bash
curl -u admin:admin https://your-grafana.com/api/org
```

### Multi-Organization Header

使用 `X-Grafana-Org-Id` header来指定目标organization：

```bash
curl -H "Authorization: Bearer <TOKEN>" \
     -H "X-Grafana-Org-Id: 2" \
     https://your-grafana.com/api/org
```

## Quick Reference

### Dashboards

**搜索 dashboards：**

```bash
GET /api/search?type=dash-db&query=<search_term>&tag=<tag>&folderIds=<folder_id>
```

**通过UID获取 dashboard：**

```bash
GET /api/dashboards/uid/<dashboard_uid>
```

**创建/更新 dashboard：**

```bash
POST /api/dashboards/db
{
  "dashboard": {
    "id": null,
    "uid": null,
    "title": "Production Overview",
    "tags": ["templated"],
    "timezone": "browser",
    "schemaVersion": 16,
    "refresh": "25s"
  },
  "folderUid": "l3KqBxCMz",
  "message": "Made changes to xyz",
  "overwrite": false
}
```

**删除 dashboard：**

```bash
DELETE /api/dashboards/uid/<dashboard_uid>
```

### Data Sources

**列出所有 data sources：**

```bash
GET /api/datasources
```

**通过UID获取 data source：**

```bash
GET /api/datasources/uid/<datasource_uid>
```

**创建 data source：**

```bash
POST /api/datasources
{
  "name": "Prometheus",
  "type": "prometheus",
  "url": "http://prometheus:9090",
  "access": "proxy",
  "basicAuth": false,
  "isDefault": true
}
```

**查询 data source：**

```bash
POST /api/ds/query
{
  "queries": [
    {
      "refId": "A",
      "datasource": { "uid": "<datasource_uid>" },
      "expr": "up",
      "intervalMs": 1000,
      "maxDataPoints": 43200
    }
  ],
  "from": "now-1h",
  "to": "now"
}
```

**健康检查：**

```bash
GET /api/datasources/uid/<datasource_uid>/health
```

### Folders

**列出 folders：**

```bash
GET /api/folders
```

**创建 folder：**

```bash
POST /api/folders
{
  "title": "My Folder",
  "uid": "my-folder-uid"
}
```

**获取 folder：**

```bash
GET /api/folders/<folder_uid>
```

**更新 folder：**

```bash
PUT /api/folders/<folder_uid>
{
  "title": "Updated Folder Title",
  "version": 1
}
```

**删除 folder：**

```bash
DELETE /api/folders/<folder_uid>?forceDeleteRules=true
```

### Alerting

**列出所有 alert rules：**

```bash
GET /api/v1/provisioning/alert-rules
```

**获取 alert rule：**

```bash
GET /api/v1/provisioning/alert-rules/<rule_uid>
```

**创建 alert rule：**

```bash
POST /api/v1/provisioning/alert-rules
{
  "title": "High CPU Alert",
  "ruleGroup": "CPU Alerts",
  "folderUID": "<folder_uid>",
  "noDataState": "OK",
  "execErrState": "OK",
  "for": "5m",
  "condition": "B",
  "annotations": { "summary": "CPU usage is high" },
  "labels": { "severity": "warning" },
  "data": [
    {
      "refId": "A",
      "relativeTimeRange": { "from": 600, "to": 0 },
      "datasourceUid": "<datasource_uid>",
      "model": {
        "expr": "100 - (avg(rate(node_cpu_seconds_total{mode=\"idle\"}[5m])) * 100)",
        "refId": "A"
      }
    },
    {
      "refId": "B",
      "relativeTimeRange": { "from": 0, "to": 0 },
      "datasourceUid": "-100",
      "model": {
        "type": "classic_conditions",
        "refId": "B",
        "conditions": [{
          "evaluator": { "type": "gt", "params": [80] },
          "operator": { "type": "and" },
          "query": { "params": ["A"] },
          "reducer": { "type": "avg" }
        }]
      }
    }
  ]
}
```

**列出 contact points：**

```bash
GET /api/v1/provisioning/contact-points
```

**列出 notification policies：**

```bash
GET /api/v1/provisioning/policies
```

**获取 active alerts：**

```bash
GET /api/alertmanager/grafana/api/v2/alerts
```

### Annotations

**查询 annotations：**

```bash
GET /api/annotations?from=<epoch_ms>&to=<epoch_ms>&tags=tag1&tags=tag2&limit=100
```

**创建 annotation：**

```bash
POST /api/annotations
{
  "dashboardUID": "jcIIG-07z",
  "panelId": 2,
  "time": 1507037197339,
  "timeEnd": 1507180805056,
  "tags": ["tag1", "tag2"],
  "text": "Annotation Description"
}
```

**更新 annotation：**

```bash
PUT /api/annotations/<annotation_id>
{
  "time": 1507037197339,
  "timeEnd": 1507180805056,
  "text": "Updated annotation",
  "tags": ["updated-tag"]
}
```

**删除 annotation：**

```bash
DELETE /api/annotations/<annotation_id>
```

### Users & Teams

**搜索 users (admin)：**

```bash
GET /api/users/search?perpage=10&page=1&query=<search_term>
```

**获取当前 user：**

```bash
GET /api/user
```

**列出 teams：**

```bash
GET /api/teams/search?perpage=50&page=1&name=<team_name>
```

**创建 team：**

```bash
POST /api/teams
{
  "name": "MyTeam",
  "email": "team@example.com"
}
```

**添加 user 到 team：**

```bash
POST /api/teams/<team_id>/members
{
  "userId": <user_id>
}
```

### Organizations

**获取当前 organization：**

```bash
GET /api/org
```

**列出所有 organizations (admin)：**

```bash
GET /api/orgs
```

**创建 organization (admin)：**

```bash
POST /api/orgs
{
  "name": "New Org"
}
```

### Service Accounts

**列出 service accounts：**

```bash
GET /api/serviceaccounts/search?perpage=10&page=1
```

**创建 service account：**

```bash
POST /api/serviceaccounts
{
  "name": "automation-sa",
  "role": "Editor"
}
```

**创建 service account token：**

```bash
POST /api/serviceaccounts/<service_account_id>/tokens
{
  "name": "token-name",
  "secondsToLive": 86400
}
```

## Reference Documentation

按领域的详细API文档：

- **[Dashboards API](references/dashboards.md)**：完整的dashboard CRUD、versions、permissions
- **[Data Sources API](references/datasources.md)**：Data source管理、queries、health checks
- **[Alerting API](references/alerting.md)**：Alert rules、contact points、notification policies、silences
- **[Folders API](references/folders.md)**：Folder管理和permissions
- **[Annotations API](references/annotations.md)**：创建、查询、更新annotations
- **[Users & Teams API](references/users_teams.md)**：User管理、team操作
- **[Common Patterns](references/common_patterns.md)**：Error handling、pagination、Python utilities

## Python Helper Library

使用 `scripts/grafana_api.py` 作为可复用的Python client：

```python
from grafana_api import GrafanaAPI

# 初始化 client
grafana = GrafanaAPI(
    base_url="https://your-grafana.com",
    token="your-service-account-token"
)

# 列出 dashboards
dashboards = grafana.search_dashboards(query="production")

# 获取 dashboard 详情
dashboard = grafana.get_dashboard_by_uid("abc123")

# 创建 annotation
grafana.create_annotation(
    dashboard_uid="abc123",
    text="Deployment completed",
    tags=["deploy", "production"]
)

# 查询 data source
result = grafana.query_datasource(
    datasource_uid="prometheus-uid",
    expr="up",
    start="now-1h",
    end="now"
)
```

## Common Operations

### 导出 Dashboard 为 JSON

```bash
curl -H "Authorization: Bearer <TOKEN>" \
     https://your-grafana.com/api/dashboards/uid/<uid> | jq '.dashboard'
```

### 导入 Dashboard

```bash
curl -X POST -H "Authorization: Bearer <TOKEN>" \
     -H "Content-Type: application/json" \
     -d @dashboard.json \
     https://your-grafana.com/api/dashboards/db
```

### 按 Tag 批量删除 Dashboards

```python
dashboards = grafana.search_dashboards(tag="deprecated")
for dash in dashboards:
    grafana.delete_dashboard(dash['uid'])
```

### 克隆 Dashboard 到另一个 Folder

```python
source = grafana.get_dashboard_by_uid("source-uid")
source['dashboard']['id'] = None
source['dashboard']['uid'] = None
source['dashboard']['title'] = f"{source['dashboard']['title']} (Copy)"
source['folderUid'] = "target-folder-uid"
grafana.create_or_update_dashboard(source)
```

## Error Handling

常见HTTP状态码：

- `200`：成功
- `400`：错误请求（invalid JSON、missing required fields）
- `401`：未授权（invalid/missing token）
- `403`：禁止（insufficient permissions）
- `404`：未找到
- `409`：冲突（resource already exists）
- `412`：前置条件失败（version mismatch）
- `422`：无法处理的实体（validation error）

## Tips

1. **使用UIDs而不是IDs**：UIDs可以在Grafana实例之间移植
2. **更新时包含version**：防止覆盖并发更改
3. **谨慎使用 `overwrite: true`**：仅在你想要强制更新时使用
4. **对大量结果进行分页**：使用 `limit` 和 `page` 参数
5. **首先在dev中测试**：始终在non-production实例上测试API调用
6. **Service accounts优于API keys**：API keys在新版Grafana中已弃用
