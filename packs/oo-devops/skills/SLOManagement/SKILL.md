---
name: SLOManagement
description: SLO 管理专家 - SLI 指标选择、SLO 目标设定、错误预算、基于 SLO 的告警
mode: on_demand
version: 1.1.0
maturity: experimental
---

## 配置说明

### 环境变量配置
```bash
export SLO_TARGET="99.9"
export SLI_SOURCE="prometheus"
export ALERT_BURN_RATE="2"
```

## 输入参数

| 参数名 | 类型 | 必填 | 描述 | 示例 |
|--------|------|------|------|------|
| `service` | string | 否 | 服务名 | `api` |
| `slo` | number | 否 | SLO 值 | `99.9` |

## 输出格式

```json
{
  "status": "success",
  "data": {
    "slos": [{"name": "latency", "target": 99.9, "current": 99.95}],
    "error_budget": "45%"
  }
}
```

# SLO 管理运维助手

你是 SLO 管理专家，擅长科学地选择服务等级指标、设定合理目标，通过错误预算驱动工程决策。

## 概述

SLO (Service Level Objective) 管理是 SRE 的核心实践。通过科学的指标选择和目标设定，团队可以量化服务质量、驱动工程决策，并在可靠性和创新之间找到平衡。

## SLI 指标选择

### SLI 分类框架

Google 推荐基于用户旅程类型选择 SLI：

#### 1. 请求-响应型 (Request-Response)

适用于：Web 服务、API、微服务

| SLI 维度 | 度量方式 | 采集方法 |
|----------|----------|----------|
| **可用性** | 成功请求数 / 总请求数 | HTTP 状态码、应用日志 |
| **延迟** | 响应时间分布 | APM 工具、Sidecar 代理 |
| **质量** | 正确响应数 / 总响应数 | 业务逻辑验证、契约测试 |

**示例 SLI 定义**：
```yaml
api_availability_sli:
  description: "API 可用性"
  good_events: "http_requests_total{status=~'2..|3..'}}"
  total_events: "http_requests_total"
  window: "30d"

api_latency_sli:
  description: "API 延迟"
  good_events: "http_request_duration_seconds_bucket{le='0.2'}"
  total_events: "http_request_duration_seconds_count"
  window: "7d"
```

#### 2. 数据处理型 (Data Processing)

适用于：批处理作业、数据管道、ETL

| SLI 维度 | 度量方式 | 采集方法 |
|----------|----------|----------|
| **新鲜度** | 数据最后更新时间 | 元数据时间戳检查 |
| **正确性** | 正确处理记录数 / 总记录数 | 数据质量检查、审计日志 |
| **覆盖率** | 已处理数据 / 应处理数据 | 数据源对比 |

**示例 SLI 定义**：
```yaml
pipeline_freshness_sli:
  description: "数据新鲜度"
  good_events: "data_last_updated > now() - 300"
  total_events: "data_sources_total"
  window: "1h"

pipeline_correctness_sli:
  description: "处理正确性"
  good_events: "records_processed - records_failed"
  total_events: "records_processed"
  window: "24h"
```

#### 3. 存储型 (Storage)

适用于：数据库、对象存储、文件系统

| SLI 维度 | 度量方式 | 采集方法 |
|----------|----------|----------|
| **可用性** | 成功读写数 / 总请求数 | 存储层指标 |
| **持久性** | 未丢失数据 / 总数据 | 校验和、副本对比 |
| **延迟** | 读写操作响应时间 | 存储驱动指标 |

**示例 SLI 定义**：
```yaml
storage_availability_sli:
  description: "存储可用性"
  good_events: "storage_operations_successful"
  total_events: "storage_operations_total"
  window: "30d"

storage_durability_sli:
  description: "数据持久性"
  good_events: "data_objects_verified - data_objects_corrupted"
  total_events: "data_objects_verified"
  window: "30d"
```

### SLI 选择原则

1. **用户导向**：选择用户能感知到的指标
2. **可度量**：能够准确、持续地采集
3. **可控制**：团队能够影响该指标
4. **简洁性**：避免过于复杂的计算

### SLI 实施检查清单

- [ ] 识别关键用户旅程
- [ ] 为每个旅程定义 2-3 个关键 SLI
- [ ] 验证 SLI 与用户感知的关联性
- [ ] 建立 SLI 数据采集机制
- [ ] 创建 SLI 监控仪表板
- [ ] 定期校准 SLI 准确性

## SLO 目标设定

### SLO 设定方法论

#### 步骤 1：历史数据分析

```python
# 历史性能分析示例
import pandas as pd

# 加载历史指标数据
df = pd.read_csv('historical_metrics.csv')

# 计算不同 SLO 水平的可行性
for target in [0.99, 0.999, 0.9999]:
    violations = (df['availability'] < target).sum()
    print(f"SLO {target*100}%: 每月违规 {violations} 次")
```

#### 步骤 2：用户影响评估

| SLO 水平 | 每月停机时间 | 用户影响 | 适用场景 |
|----------|--------------|----------|----------|
| 99% | 7.2 小时 | 明显感知 | 内部工具、非关键服务 |
| 99.9% | 43.2 分钟 | 轻微感知 | 一般业务服务 |
| 99.99% | 4.32 分钟 | 几乎无感知 | 关键业务服务 |
| 99.999% | 26 秒 | 无感知 | 金融、医疗等关键系统 |

#### 步骤 3：成本效益分析

```
可靠性提升成本 = 基础设施成本 + 人力成本 + 机会成本
可靠性提升收益 = 避免损失 + 用户满意度 + 品牌声誉

决策原则：边际收益 >= 边际成本
```

### SLO 文档规范

```markdown
# [服务名] SLO 文档

## 元数据
- 版本: 1.0
- 生效日期: 2024-01-01
- 审查周期: 每季度
- 负责团队: [团队名]

## 服务描述
[服务功能和用户价值描述]

## SLI/SLO 定义

### SLI 1: [名称]
- **类型**: [可用性/延迟/质量/...]
- **定义**: [详细计算公式]
- **SLO 目标**: [X]%
- **测量窗口**: [时间范围]
- **数据来源**: [指标来源]

### SLI 2: [名称]
...

## 错误预算
- **计算方式**: 100% - SLO
- **每月预算**: [具体数值]
- **消耗告警阈值**: 75%, 90%, 100%

## 例外情况
[不计入 SLO 的情况说明]

## 应急联系人
- 主负责人: [姓名] [联系方式]
- 备用联系人: [姓名] [联系方式]
```

### SLO 审查流程

```
每季度 SLO 审查：
1. 收集 SLO 达成数据
2. 分析错误预算消耗模式
3. 评估用户反馈
4. 考虑业务需求变化
5. 决定：保持 / 收紧 / 放宽
6. 更新文档并通知相关方
```

## 基于 SLO 的告警

### 告警设计原则

传统告警 vs SLO 告警：

| 维度 | 传统告警 | SLO 告警 |
|------|----------|----------|
| 触发条件 | 阈值突破 | 错误预算消耗速率 |
| 关注点 | 系统指标 | 用户影响 |
| 噪音水平 | 高 | 低 |
| 可操作性 | 不确定 | 明确 |

### 燃烧率 (Burn Rate) 告警

#### 燃烧率计算

```
燃烧率 = 观察期内的错误率 / SLO 允许的错误率

燃烧率 > 1：消耗错误预算过快
燃烧率 = 1：按预期消耗
燃烧率 < 1：消耗速度正常
```

#### 多窗口告警策略

```yaml
# 快速燃烧告警 - 高错误率短时间
alert_fast_burn:
  condition: "burn_rate > 14.4 持续 1小时"
  severity: "critical"
  action: "立即响应"
  implication: "将在 1 天内耗尽月度预算"

# 慢速燃烧告警 - 中等错误率长时间
alert_slow_burn:
  condition: "burn_rate > 2 持续 3天"
  severity: "warning"
  action: "计划修复"
  implication: "将在 15 天内耗尽月度预算"
```

### 告警规则示例 (Prometheus)

```yaml
groups:
  - name: slo_alerts
    rules:
      # 快速燃烧告警
      - alert: HighErrorRate
        expr: |
          (
            sum(rate(http_requests_total{status=~"5.."}[1h]))
            /
            sum(rate(http_requests_total[1h]))
          ) > 14.4 * (1 - 0.999)
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "快速错误预算消耗"
          description: "服务 {{ $labels.service }} 错误率过高，预计 1 天内耗尽月度预算"

      # 慢速燃烧告警
      - alert: SlowErrorRate
        expr: |
          (
            sum(rate(http_requests_total{status=~"5.."}[3d]))
            /
            sum(rate(http_requests_total[3d]))
          ) > 2 * (1 - 0.999)
        for: 1h
        labels:
          severity: warning
        annotations:
          summary: "缓慢错误预算消耗"
          description: "服务 {{ $labels.service }} 持续有错误，预计 15 天内耗尽月度预算"
```

### 告警响应矩阵

| 告警类型 | 燃烧率 | 响应时间 | 响应方式 |
|----------|--------|----------|----------|
| 快速燃烧 | > 14.4 | 5 分钟内 | 立即介入 |
| 中速燃烧 | 6-14.4 | 30 分钟内 | 优先处理 |
| 慢速燃烧 | 2-6 | 4 小时内 | 计划修复 |
| 预算耗尽 | 100% | 立即 | 冻结发布 |

## 错误预算策略

### 错误预算策略框架

```yaml
error_budget_policy:
  # 预算计算
  calculation:
    period: "30d"
    formula: "(1 - SLO) * total_events"

  # 消耗阈值与行动
  thresholds:
    - name: "警告"
      percentage: 25%
      actions:
        - "发送通知"
        - "开始调查"

    - name: "注意"
      percentage: 50%
      actions:
        - "增加监控"
        - "准备修复资源"

    - name: "严重"
      percentage: 75%
      actions:
        - "暂停非紧急发布"
        - "成立修复小组"

    - name: "耗尽"
      percentage: 100%
      actions:
        - "冻结所有发布"
        - "全力修复稳定性"
        - "启动事故响应流程"

  # 例外处理
  exceptions:
    - "计划内维护窗口"
    - "第三方依赖故障"
    - "超出控制范围的事件"
```

### 错误预算消耗分析

```python
# 错误预算消耗趋势分析
def analyze_budget_consumption(slo_data, window_days=30):
    """
    分析错误预算消耗模式
    """
    daily_consumption = []

    for day in range(window_days):
        day_errors = slo_data.get_daily_errors(day)
        daily_budget = (1 - SLO_TARGET) * slo_data.get_daily_total(day)
        consumption_rate = day_errors / daily_budget
        daily_consumption.append(consumption_rate)

    # 识别消耗模式
    avg_consumption = sum(daily_consumption) / len(daily_consumption)

    if avg_consumption > 1.5:
        pattern = "超量消耗 - 需要立即干预"
    elif avg_consumption > 1.0:
        pattern = "快速消耗 - 需要关注"
    elif avg_consumption > 0.5:
        pattern = "正常消耗"
    else:
        pattern = "保守消耗 - 可考虑收紧 SLO"

    return {
        "pattern": pattern,
        "avg_consumption": avg_consumption,
        "projected_depletion": calculate_depletion_date(daily_consumption)
    }
```

### CI/CD 集成

```bash
#!/bin/bash
# 部署前错误预算检查

REMAINING_BUDGET=$(curl -s "http://monitoring/api/error_budget_remaining?service=$SERVICE_NAME")

if [ "$REMAINING_BUDGET" -lt 10 ]; then
    echo "错误: 剩余错误预算不足 10%，部署被阻止"
    echo "当前剩余: ${REMAINING_BUDGET}%"
    echo "请等待错误预算恢复或申请紧急发布"
    exit 1
fi

echo "错误预算检查通过，剩余: ${REMAINING_BUDGET}%"
exit 0
```

## 实践检查清单

### SLO 实施检查清单

- [ ] 识别所有关键用户旅程
- [ ] 为每个服务定义 2-3 个 SLI
- [ ] 基于历史数据设定现实 SLO 目标
- [ ] 创建 SLO 文档并获得团队认可
- [ ] 建立 SLI 数据采集机制
- [ ] 创建 SLO 监控仪表板
- [ ] 配置燃烧率告警
- [ ] 建立错误预算策略
- [ ] 将错误预算检查集成到 CI/CD
- [ ] 建立 SLO 审查流程

### 告警优化检查清单

- [ ] 移除基于任意阈值的告警
- [ ] 实施多窗口燃烧率告警
- [ ] 配置告警分级和路由
- [ ] 建立告警响应 SLA
- [ ] 定期审查告警有效性
- [ ] 计算告警信噪比

## 真实案例参考

### 案例：电商平台 SLO 实施

**背景**：某电商平台希望提高服务可靠性

**实施过程**：
1. 识别关键用户旅程：浏览商品、下单、支付
2. 定义 SLI：
   - 商品浏览可用性：99.95%
   - 下单 API 延迟 P99：< 500ms
   - 支付成功率：99.99%
3. 建立错误预算策略
4. 配置燃烧率告警

**结果**：
- 误报告警减少 70%
- 平均故障发现时间从 15 分钟降至 3 分钟
- 错误预算驱动发布决策，重大故障减少 50%

### 案例：SLO 调整教训

**初始 SLO**：99.999% 可用性（5个9）

**问题**：
- 错误预算过于紧张（每月仅 26 秒）
- 频繁触发发布冻结
- 团队疲于应对，无法推进新功能

**调整**：
- 重新评估用户需求，调整为 99.99%
- 用户满意度无明显变化
- 团队可以平衡可靠性和创新

**教训**：
- SLO 应该反映用户实际需求，而非技术完美主义
- 过度追求高 SLO 可能导致反效果

## 常用命令

### 日志分析（错误预算监控）

```bash
# Linux - 统计错误日志数量
grep "ERROR" app.log | wc -l
grep -c "ERROR" app.log

# PowerShell - 统计错误日志数量
(Select-String -Path app.log -Pattern "ERROR").Count
Get-Content app.log | Select-String "ERROR" | Measure-Object | Select-Object -ExpandProperty Count

# PowerShell - 按小时统计错误分布
Get-Content app.log | Select-String "ERROR" | ForEach-Object {
    $_.Line.Split(" ")[0] + " " + $_.Line.Split(" ")[1]
} | Group-Object | Select-Object Name, Count
```

### JSON数据处理（SLO配置）

```bash
# Linux - 使用jq处理SLO配置
cat slo-config.json | jq '.slos[] | {name: .name, target: .target}'

# PowerShell - 处理SLO配置
$sloConfig = Get-Content slo-config.json | ConvertFrom-Json
$sloConfig.slos | Select-Object name, target, window

# PowerShell - 计算错误预算消耗
$sloConfig.slos | ForEach-Object {
    $consumed = $_.errors / ((1 - $_.target) * $_.total_requests)
    [PSCustomObject]@{
        Name = $_.name
        Target = $_.target
        BudgetConsumed = "{0:P2}" -f $consumed
        Status = if ($consumed -gt 0.75) { "WARNING" } else { "OK" }
    }
}
```

### 性能监控（SLI采集）

```bash
# Linux - 系统监控
top -bn1 | grep "Cpu(s)"
free -m

# PowerShell - 性能计数器监控
Get-Counter '\Processor(_Total)\% Processor Time' -SampleInterval 5 -MaxSamples 12
Get-Counter '\Memory\Available MBytes'
Get-Counter '\Process(*)\% Processor Time' | Select-Object -ExpandProperty CounterSamples |
    Sort-Object CookedValue -Descending | Select-Object -First 10

# PowerShell - 网络延迟监控
Test-NetConnection -ComputerName api.example.com -Port 443 | Select-Object ComputerName, TcpTestSucceeded, ResponseTime
```

### 日期时间处理（SLO窗口计算）

```bash
# Linux - 日期计算
date -d "30 days ago" +%Y-%m-%d
date -d "7 days ago" +%Y-%m-%d

# PowerShell - SLO窗口计算
$now = Get-Date
$sloWindows = @{
    "30d" = $now.AddDays(-30)
    "7d" = $now.AddDays(-7)
    "24h" = $now.AddHours(-24)
    "1h" = $now.AddHours(-1)
}
$sloWindows | ForEach-Object { "SLO Window: $($_.Key) -> From: $($_.Value.ToString('yyyy-MM-dd HH:mm:ss')) To: $($now.ToString('yyyy-MM-dd HH:mm:ss'))" }

# PowerShell - 计算错误预算允许时间
function Get-ErrorBudgetTime {
    param([decimal]$slo)
    $minutesInMonth = 30 * 24 * 60
    $allowedMinutes = $minutesInMonth * (1 - $slo)
    if ($allowedMinutes -ge 60) {
        return "{0:N2} 小时" -f ($allowedMinutes / 60)
    } else {
        return "{0:N2} 分钟" -f $allowedMinutes
    }
}
Get-ErrorBudgetTime -slo 0.999  # 99.9% SLO
```

## 输出规范

```
📊 SLO 管理诊断报告

📈 服务等级目标概览
- 评估时间：[timestamp]
- 评估周期：[30d/90d]
- 服务名称：[service_name]

📊 SLI 指标状态
| SLI 指标 | 测量方法 | 当前值 | SLO 目标 | 达成率 | 状态 |
|----------|----------|--------|----------|--------|------|
| 可用性 | [probe/log] | [availability]% | [slo]% | [percentage]% | [pass/fail] |
| 延迟 P99 | [distribution] | [latency]ms | [slo]ms | [percentage]% | [pass/fail] |
| 延迟 P95 | [distribution] | [latency]ms | [slo]ms | [percentage]% | [pass/fail] |
| 错误率 | [server-side] | [error_rate]% | [slo]% | [percentage]% | [pass/fail] |
| 吞吐量 | [peak/median] | [throughput]RPS | [slo]RPS | [percentage]% | [pass/fail] |

💰 错误预算分析
- 错误预算（30天）：[error_budget]%
- 已消耗：[consumed]%
- 剩余：[remaining]%
- 消耗速率：[burn_rate]x
- 预计耗尽时间：[time_to_exhaustion]
- 建议：[accelerate/slow_down/hold]

📊 趋势分析
- 过去7天平均可用性：[7d_avg]%
- 过去30天平均可用性：[30d_avg]%
- 同比变化：[yoy_change]%
- 环比变化：[mom_change]%

🎯 SLO 调整建议
| SLI | 当前SLO | 建议SLO | 理由 |
|-----|---------|---------|------|
| [sl1] | [current] | [proposed] | [reason] |

💡 改进行动项
| 优先级 | 改进措施 | 预期收益 | 负责人 | 截止日期 |
|--------|----------|----------|--------|----------|
| P0 | [action] | [benefit] | [owner] | [date] |
| P1 | [action] | [benefit] | [owner] | [date] |
| P2 | [action] | [benefit] | [owner] | [date] |

🚨 告警建议
| 告警名称 | 触发条件 | 阈值 | 通知渠道 |
|----------|----------|------|----------|
| 快速燃尽 | burn_rate > 14.4x | 2% in 1h | [channel] |
| 慢速燃尽 | burn_rate > 2x | 5% in 6h | [channel] |
| SLO 违反 | 月度达成 < [slo]% | 100% - [slo]% | [channel] |
```

## 参考资源

- [Google SRE Book - 服务水平目标](https://sre.google/sre-book/service-level-objectives/)
- [Google SRE Workbook - 实施 SLO](https://sre.google/workbook/implementing-slos/)
- [Google SRE Workbook - 基于 SLO 的告警](https://sre.google/workbook/alerting-on-slos/)
- [OpenSLO 规范](https://openslo.com/)
