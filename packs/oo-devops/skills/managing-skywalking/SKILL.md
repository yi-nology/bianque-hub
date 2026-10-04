---
name: managing-skywalking
description: 使用 Skywalking 时 —— apache SkyWalking 应用性能 监控平台，用于分布式追踪、服务网格可观测性、 指标聚合和日志分析。涵盖服务拓扑、端点 性能、追踪分析、告警管理和基础设施监控。 用于监控服务健康、调查分布式追踪、 分析端点延迟或查看 SkyWalking 告警。
mode: on_demand
version: 1.1.0
maturity: experimental
---

# SkyWalking 监控 Skill

使用 SkyWalking GraphQL API 查询、分析和管理 Apache SkyWalking 可观测性数据。

## API 概述

SkyWalking 在 `https://<SKYWALKING_HOST>/graphql` 使用 GraphQL API。

### 核心 Helper 函数

```bash
#!/bin/bash

sw_gql() {
    local query="$1"
    curl -s -X POST "${SKYWALKING_URL}/graphql" \
        -H "Content-Type: application/json" \
        ${SKYWALKING_AUTH:+-H "Authorization: Bearer $SKYWALKING_AUTH"} \
        -d "{\"query\": $(echo "$query" | jq -Rs .)}"
}

sw_duration() {
    local hours="${1:-1}"
    local end=$(date -u +"%Y-%m-%d %H%M")
    local start=$(date -u -d "${hours} hours ago" +"%Y-%m-%d %H%M")
    echo "{start: \"${start}\", end: \"${end}\", step: HOUR}"
}
```

## 强制：发现优先模式

**始终在查询前发现服务、端点和实例。**

### 阶段 1：发现

```bash
#!/bin/bash
DUR=$(sw_duration 1)

echo "=== Services ==="
sw_gql "{
    getAllServices(duration: ${DUR}) {
        id name group
    }
}" | jq -r '.data.getAllServices[] | "\(.id)\t\(.name)\t\(.group // "default")"' | head -20

echo ""
echo "=== Service Instances ==="
SERVICE_ID="${1:-}"
[ -n "$SERVICE_ID" ] && sw_gql "{
    getServiceInstances(serviceId: \"${SERVICE_ID}\", duration: ${DUR}) {
        id name language instanceUUID
    }
}" | jq -r '.data.getServiceInstances[] | "\(.id)\t\(.name)\t\(.language // "unknown")"' | head -20

echo ""
echo "=== Endpoints (Top 20) ==="
[ -n "$SERVICE_ID" ] && sw_gql "{
    findEndpoint(serviceId: \"${SERVICE_ID}\", keyword: \"\", limit: 20) {
        id name
    }
}" | jq -r '.data.findEndpoint[] | "\(.id)\t\(.name)"' | head -20

echo ""
echo "=== Active Alarms ==="
sw_gql "{
    getAlarm(duration: ${DUR}, paging: {pageNum: 1, pageSize: 15}) {
        msgs {
            id message startTime scope
        }
    }
}" | jq -r '.data.getAlarm.msgs[] | "\(.startTime)\t\(.scope)\t\(.message[0:60])"' | head -15
```

### 阶段 2：分析

```bash
#!/bin/bash
DUR=$(sw_duration 1)
SERVICE_ID="${1:?Service ID required}"

echo "=== Service Health (Golden Signals) ==="
sw_gql "{
    readMetricsValues(condition: {name: \"service_resp_time\", entity: {scope: Service, serviceName: \"${SERVICE_ID}\"}}, duration: ${DUR}) {
        values { values { value } }
    }
}" | jq -r '.data.readMetricsValues.values.values | map(.value) | "Avg Response Time: \(add / length)ms"'

sw_gql "{
    readMetricsValues(condition: {name: \"service_sla\", entity: {scope: Service, serviceName: \"${SERVICE_ID}\"}}, duration: ${DUR}) {
        values { values { value } }
    }
}" | jq -r '.data.readMetricsValues.values.values | map(.value) | "Success Rate: \(add / length / 100)%"'

sw_gql "{
    readMetricsValues(condition: {name: \"service_cpm\", entity: {scope: Service, serviceName: \"${SERVICE_ID}\"}}, duration: ${DUR}) {
        values { values { value } }
    }
}" | jq -r '.data.readMetricsValues.values.values | map(.value) | "Calls/min: \(add / length)"'

echo ""
echo "=== Topology (Dependencies) ==="
sw_gql "{
    getServiceTopology(serviceId: \"${SERVICE_ID}\", duration: ${DUR}) {
        nodes { id name type }
        calls { source target detectPoints }
    }
}" | jq -r '.data.getServiceTopology.calls[] | "\(.source) -> \(.target)\t\(.detectPoints | join(","))"' | head -15

echo ""
echo "=== Slow Endpoints ==="
sw_gql "{
    sortMetrics(condition: {name: \"endpoint_avg\", topN: 15, order: DES, parentService: \"${SERVICE_ID}\"}, duration: ${DUR}) {
        name value
    }
}" | jq -r '.data.sortMetrics[] | "\(.value)ms\t\(.name)"' | head -15
```

## 输出规则
- **TOKEN 效率**: 目标 ≤50 行 —— 使用 `topN` 和 `pageSize` 限制结果
- Duration 格式: `{start: "YYYY-MM-DD HHmm", end: "YYYY-MM-DD HHmm", step: HOUR}`
- Metric 名称: `service_resp_time`, `service_sla`, `service_cpm`, `endpoint_avg`
- 使用 `sortMetrics` 进行 top-N 分析，而不是获取所有 metric 值

## 输出格式

将结果作为结构化报告呈现：
```
Managing Skywalking Report
══════════════════════════
Resources discovered: [count]

Resource       Status    Key Metric    Issues
──────────────────────────────────────────────
[name]         [ok/warn] [value]       [findings]

Summary: [total] resources | [ok] healthy | [warn] warnings | [crit] critical
Action Items: [list of prioritized findings]
```

目标 ≤50 行输出。使用表格进行多资源比较。

## 反幻觉规则

1. **永远不要假设资源名称** —— 始终在阶段 2 之前通过 CLI/API 在阶段 1 中发现。
2. **永远不要伪造 metric 名称或维度** —— 对照服务文档或 `--help` 输出验证。
3. **永远不要混合 CLI 命令** —— 确认您针对的是哪个版本/API。
4. **始终使用发现 → 验证 → 分析链** —— 引用的每个资源必须首先被发现。
5. **始终优雅地处理空结果** —— 空响应是有效数据，不是需要重试的错误。

## 反辩解

| 快捷方式 | 反驳 | 原因 |
|----------|---------|-----|
| "我会跳过发现，检查已知资源" | 始终先运行阶段 1 发现 | 资源名称会变化，新资源会出现 —— 假设的名称会导致错误 |
| "用户只要求快速检查" | 遵循完整的发现 → 分析流程 | 快速检查会遗漏关键问题；结构化分析会捕获静默故障 |
| "默认配置可能没问题" | 显式审计配置 | 默认值通常会禁用日志记录、安全和优化功能 |
| "这不为此需要 metrics" | 始终检查相关指标（如果可用） | API/CLI 响应显示当前状态；指标揭示趋势和间歇性问题 |
| "我无法访问那个" | 尝试命令并报告实际错误 | 假设的权限失败会阻止有用的调查；实际错误具有信息性 |
