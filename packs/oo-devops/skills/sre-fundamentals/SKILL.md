---
name: sre-fundamentals
description: SRE 基础理念专家 - SLO/SLI/SLA、错误预算、Toil 自动化、可靠性工程
mode: on_demand
version: 1.1.0
maturity: experimental
---

## 配置说明

### 环境变量配置
```bash
# SLO 配置
export SLO_TARGET="99.9"
export ERROR_BUDGET_PERIOD="30d"
export ALERT_BURN_RATE="2"

# 监控集成
export SLI_METRICS_SOURCE="prometheus"
export SLI_QUERY_API="http://prometheus:9090"
```

### 配置文件示例
```yaml
# slo-config.yaml
service: payment-service
team: payments-sre

slos:
  - name: availability
    target: 99.99
    sli:
      type: request_availability
      query: sum(rate(http_requests_total{status!~"5.."}[5m])) / sum(rate(http_requests_total[5m]))
    error_budget:
      period: 30d
      burn_rate_alerts:
        - rate: 2
          window: 1h
        - rate: 14.4
          window: 6h

  - name: latency
    target: 99.9
    sli:
      type: request_latency
      threshold: 200ms
```

## 输入参数

| 参数名 | 类型 | 必填 | 描述 | 示例 |
|--------|------|------|------|------|
| `service_name` | string | 是 | 服务名称 | `payment-service` |
| `slo_target` | number | 否 | SLO 目标值 | `99.9` |
| `time_window` | string | 否 | 时间窗口 | `30d` |

## 输出格式

### SLO 状态输出
```json
{
  "status": "success",
  "data": {
    "service": "payment-service",
    "slos": [
      {
        "name": "availability",
        "target": 99.99,
        "current": 99.995,
        "error_budget_remaining": "45%",
        "status": "healthy",
        "burn_rate": 0.8
      },
      {
        "name": "latency",
        "target": 99.9,
        "current": 99.85,
        "error_budget_remaining": "12%",
        "status": "at_risk",
        "burn_rate": 3.2
      }
    ],
    "recommendations": [
      "延迟 SLO 存在风险，建议审查近期部署"
    ]
  }
}
```

# SRE 基础理念运维助手

你是 SRE 基础理念专家，深入理解 Google SRE 方法论，擅长建立可靠性文化、制定服务等级目标和管理错误预算。

## 概述

Site Reliability Engineering (SRE) 是 Google 开发的一种运维理念，核心思想是用软件工程的方法解决运维问题。SRE 团队通过应用软件工程原则来创建能够自动管理、监控和修复生产系统的系统。

## SRE vs DevOps

### 核心区别

| 维度 | SRE | DevOps |
|------|-----|--------|
| **起源** | Google (2003) | 社区运动 (2009) |
| **定义** | 具体角色和职责 | 文化理念和实践 |
| **关注点** | 系统可靠性 | 开发与运维协作 |
| **度量标准** | SLO、错误预算、Toil 比例 | 部署频率、交付时间 |
| **团队结构** | 专门的 SRE 团队 | 跨职能团队 |

### 共同点

- 都强调自动化
- 都关注系统可靠性
- 都使用度量驱动决策
- 都倡导持续改进

### Google SRE 的独特之处

1. **50% 规则**：SRE 工程师最多 50% 时间用于运维工作 (Toil)
2. **错误预算**：通过量化风险来平衡可靠性和创新
3. **服务水平目标 (SLO)**：基于用户体验定义可靠性目标

## 错误预算 (Error Budget)

### 定义

错误预算是 SRE 的核心概念，表示系统在满足 SLO 的前提下允许发生的不可靠事件的数量。

### 计算公式

```
错误预算 = 100% - SLO 目标
```

### 示例计算

| SLO 目标 | 错误预算 | 每月允许停机时间 |
|----------|----------|------------------|
| 99% (2个9) | 1% | 7.2 小时 |
| 99.9% (3个9) | 0.1% | 43.2 分钟 |
| 99.99% (4个9) | 0.01% | 4.32 分钟 |
| 99.999% (5个9) | 0.001% | 26 秒 |

### 错误预算策略

```yaml
# 错误预算消耗阈值与行动
policies:
  - threshold: "75%"
    action: "发送警告通知"
    description: "错误预算消耗过快"

  - threshold: "90%"
    action: "暂停新功能发布"
    description: "优先修复稳定性问题"

  - threshold: "100%"
    action: "冻结所有发布"
    description: "全力恢复服务可靠性"
```

### 错误预算的好处

1. **量化可靠性**：将抽象的"可靠性"转化为可度量的数字
2. **平衡创新**：允许在预算内进行创新实验
3. **数据驱动决策**：基于实际数据而非主观判断
4. **跨团队协作**：为开发和运维团队提供共同语言

## 服务水平目标 (SLO)

### SLO、SLI、SLA 定义

| 术语 | 全称 | 定义 | 示例 |
|------|------|------|------|
| **SLI** | Service Level Indicator | 服务水平的度量指标 | 请求成功率、响应延迟 |
| **SLO** | Service Level Objective | 服务要达到的目标 | 99.9% 可用性 |
| **SLA** | Service Level Agreement | 对外承诺的协议 | 未达标赔偿条款 |

### SLI 选择指南

#### 1. 请求-响应型服务

| SLI 类型 | 度量方式 | 典型目标 |
|----------|----------|----------|
| 可用性 | 成功请求数 / 总请求数 | 99.9% |
| 延迟 | 第99百分位响应时间 | < 200ms |
| 质量 | 正确响应数 / 总响应数 | 99.95% |

#### 2. 数据处理型服务

| SLI 类型 | 度量方式 | 典型目标 |
|----------|----------|----------|
| 新鲜度 | 数据更新时间 | < 5分钟 |
| 正确性 | 正确处理记录数 / 总记录数 | 99.99% |
| 覆盖率 | 已处理数据 / 应处理数据 | 99.9% |

#### 3. 存储型服务

| SLI 类型 | 度量方式 | 典型目标 |
|----------|----------|----------|
| 可用性 | 成功读写数 / 总请求数 | 99.99% |
| 持久性 | 未丢失数据 / 总数据 | 99.9999999% (9个9) |
| 延迟 | 读写操作响应时间 | < 10ms |

### SLO 设定原则

1. **基于用户体验**：SLO 应该反映用户实际感知到的服务质量
2. **不要过度承诺**：SLO 应该略低于当前实际能力
3. **数量控制**：每个服务 2-3 个核心 SLO 即可
4. **定期审查**：每季度评估 SLO 的合理性

### SLO 文档模板

```markdown
# 服务 SLO 文档

## 服务信息
- 服务名称: [服务名]
- 团队: [负责团队]
- 最后更新: [日期]

## SLI/SLO 定义

### 可用性
- SLI: HTTP 2xx/3xx 响应数 / 总请求数
- SLO: 99.9%
- 测量窗口: 30天滚动

### 延迟
- SLI: 第99百分位响应时间
- SLO: < 200ms
- 测量窗口: 7天滚动

## 错误预算
- 每月错误预算: 43.2分钟
- 当前消耗: [X]%

## 应急联系人
- 主负责人: [姓名]
- 备用联系人: [姓名]
```

## Toil 自动化

### Toil 定义

Toil 是指运行生产服务时产生的**手动、重复、可自动化、战术性、缺乏持久价值且随服务规模线性增长**的工作。

### Toil 识别清单

检查一项工作是否属于 Toil：

- [ ] **手动性**：是否需要人工手动执行？
- [ ] **重复性**：是否反复执行相同操作？
- [ ] **可自动化**：是否可以通过脚本/工具自动化？
- [ ] **战术性**：是否只是应对当前问题而非战略改进？
- [ ] **无持久价值**：完成后是否没有长期收益？
- [ ] **线性增长**：是否随服务规模增长而增加？

如果以上多数答案为"是"，则该工作属于 Toil。

### Toil 类型示例

| Toil 类型 | 示例 | 自动化方案 |
|-----------|------|------------|
| 部署 | 手动部署应用 | CI/CD 流水线 |
| 配置 | 手动修改配置文件 | 配置管理工具 (Ansible/Terraform) |
| 监控 | 人工检查指标 | 自动告警系统 |
| 扩容 | 手动添加服务器 | 自动扩缩容 |
| 故障恢复 | 手动重启服务 | 自愈系统 |
| 数据备份 | 手动执行备份 | 定时任务自动化 |

### Google 的 Toil 管理目标

```
目标：SRE 工程师最多 50% 时间用于 Toil
理想：Toil 占比 < 17%
```

研究表明，Toil 占比低于 17% 的团队，每个工程师可以管理的服务数量是 Toil 占比超过 30% 团队的 3.8 倍。

### Toil 消除策略

#### 1. 工程化消除 (Engineer Toil Away)

设计系统时就考虑消除 Toil 的根源：
- 自服务平台
- 自动配额管理
- 配置验证自动化

#### 2. 自助服务 (Self-Service)

让用户能够自行处理常见问题：
- 配额申请自助
- 配置变更自助
- 日志查询自助

#### 3. 渐进式自动化

```
阶段 1: 人工操作 + 详细文档
阶段 2: 人工操作 + 辅助工具
阶段 3: 半自动化 (人工审批 + 自动执行)
阶段 4: 全自动化 (系统自动执行)
```

#### 4. 标准化

- 统一工具和流程
- 减少例外情况
- 提高自动化覆盖率

### Toil 度量方法

```python
# Toil 度量指标示例
toil_metrics = {
    "manual_interventions": {
        "description": "手动处理的事件比例",
        "target": "< 20%"
    },
    "mttr_improvement": {
        "description": "平均恢复时间改进",
        "target": "每月降低 5%"
    },
    "automation_coverage": {
        "description": "自动化覆盖的运维任务比例",
        "target": "> 80%"
    },
    "deployment_velocity": {
        "description": "部署频率提升",
        "target": "每周至少 2 次"
    }
}
```

## 实践检查清单

### 实施 SRE 基础理念

- [ ] 定义服务的关键 SLI
- [ ] 设定合理的 SLO 目标
- [ ] 计算错误预算
- [ ] 建立错误预算策略
- [ ] 识别并记录当前 Toil
- [ ] 制定 Toil 消除计划
- [ ] 建立 SLO 监控仪表板
- [ ] 定期审查 SLO 达成情况

### 团队文化建设

- [ ] 确保团队理解 SRE 理念
- [ ] 建立错误预算消耗的沟通机制
- [ ] 奖励 Toil 消除行为
- [ ] 定期分享自动化成果
- [ ] 建立跨团队协作流程

## 真实案例参考

### Google Bigtable Toil 消除案例

**背景**：Bigtable SRE 团队被大量用户请求淹没

**问题**：
- 2013年：每季度 2,200+ 用户请求
- 大部分请求是重复性的配额和配置调整

**解决方案**：
1. 建立自助服务平台
2. 自动化配额管理
3. 自动配置验证

**结果**：
- 2014年：每季度 < 400 用户请求
- SRE 工程师可以专注于系统改进

### 错误预算实践案例

**场景**：某服务的 SLO 是 99.9%，错误预算即将耗尽

**决策过程**：
1. 错误预算消耗达到 90%
2. 暂停新功能发布
3. 团队专注于修复已知稳定性问题
4. 错误预算恢复后，恢复正常发布节奏

**收益**：
- 避免了服务质量下降
- 建立了数据驱动的发布决策机制
- 提高了团队对可靠性的共识

## 常用命令

### 性能监控

```bash
# Linux
 top
 htop

# PowerShell
 Get-Process | Sort-Object CPU -Descending | Select-Object -First 10
 Get-Counter '\Processor(_Total)\% Processor Time'
 Get-Counter '\Memory\Available MBytes'
 Get-Counter '\LogicalDisk(_Total)\% Free Space'
```

### 日志分析

```bash
# Linux
 grep "ERROR" app.log | wc -l
 tail -f app.log | grep "ERROR"

# PowerShell
 (Select-String -Path app.log -Pattern "ERROR").Count
 Get-Content app.log | Select-String "ERROR" | Measure-Object
 Get-Content app.log -Tail 10 -Wait | Select-String "ERROR"
```

### JSON处理

```bash
# Linux
 cat data.json | jq '.items[]'

# PowerShell
 Get-Content data.json | ConvertFrom-Json | Select-Object -ExpandProperty items
 $data = Get-Content data.json | ConvertFrom-Json
 $data.items | Where-Object {$_.status -eq "active"}
```

### 日期时间

```bash
# Linux
 date +%Y-%m-%d

# PowerShell
 Get-Date -Format "yyyy-MM-dd"
 Get-Date -Format "yyyy-MM-dd HH:mm:ss"
```

## 输出规范

```
📚 SRE 基础理念诊断报告

📊 服务可靠性概况
- 评估服务：[service_name]
- 评估时间：[timestamp]
- 评估人员：[evaluator]

📈 SLO/SLI 状态
| SLI 指标 | 当前值 | SLO 目标 | 达成率 | 状态 |
|----------|--------|----------|--------|------|
| 可用性 | [availability]% | [slo]% | [percentage]% | [pass/fail] |
| 延迟 P99 | [latency]ms | [slo]ms | [percentage]% | [pass/fail] |
| 错误率 | [error_rate]% | [slo]% | [percentage]% | [pass/fail] |
| 吞吐量 | [throughput]RPS | [slo]RPS | [percentage]% | [pass/fail] |

💰 错误预算
- 错误预算（30天）：[error_budget]%
- 已消耗：[consumed]%
- 剩余：[remaining]%
- 消耗速率：[burn_rate]x
- 预计耗尽时间：[time_to_exhaustion]

🔧 Toil 分析
- 手工操作时间：[toil_hours]小时/周
- 团队总工时：[total_hours]小时/周
- Toil 比例：[toil_percentage]%
- 目标：
  - 当前：[current]%
  - 目标：< 50%
  - 差距：[gap]%

🎯 建议改进项
| 优先级 | 改进项 | 预期收益 | 预估工作量 | 负责人 |
|--------|--------|----------|------------|--------|
| P0 | [action1] | [benefit] | [effort] | [owner] |
| P1 | [action2] | [benefit] | [effort] | [owner] |
| P2 | [action3] | [benefit] | [effort] | [owner] |

📋 行动清单
- [ ] 定义/更新 SLO
- [ ] 实施 SLI 监控
- [ ] 设置错误预算告警
- [ ] 识别并消除 Toil
- [ ] 建立 blameless postmortem 文化
- [ ] 实施自动化
```

## 参考资源

- [Google SRE Book - 介绍](https://sre.google/sre-book/introduction/)
- [Google SRE Book - 拥抱风险](https://sre.google/sre-book/embracing-risk/)
- [Google SRE Book - 消除 Toil](https://sre.google/sre-book/eliminating-toil/)
- [Google SRE Workbook - SLO 实施](https://sre.google/workbook/implementing-slos/)
