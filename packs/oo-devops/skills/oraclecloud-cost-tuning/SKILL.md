---
name: oraclecloud-cost-tuning
description: 使用 Usage API 跟踪 OCI 支出并设置预算警报。 在监控 Oracle Cloud 成本、创建预算、按区间或服务分析支出或优化 Universal Credits 消耗时使用。 使用 "oraclecloud cost"、"oci budget"、"oci usage api"、"oci spending"、"oracle cloud cost tuning" 触发。
mode: on_demand
version: 1.1.0
maturity: experimental
---

# Oracle Cloud 成本调优

## 概述

使用 Usage API 以编程方式跟踪 OCI 支出，并在 Universal Credits 意外耗尽之前设置预算警报。OCI 定价因形状、区域和承诺级别而异，控制台中的成本分析工具深藏且令人困惑。本技能使用 Usage API 按区间、服务和形状查询支出，创建带有警报规则的预算，并涵盖优化策略，包括 Always Free 层级资源、抢占式实例和预留容量。

**目的：** 通过代码获取 OCI 支出的可见性，设置主动预算警报，并识别成本优化机会。

## 前置条件

- 在 `~/.oci/config` 中配置了 API 签名密钥的 **OCI 租户**
- 安装了 `pip install oci` 的 **Python 3.8+**
- 用于租户范围成本查询的 **租户 OCID**（根区间）
- 授予租户中 `read usage-reports` 权限的 **IAM 策略**
- 用于预算警报交付的 **通知主题 OCID**（参见 `oraclecloud-observability`）

## 说明

### 步骤 1：使用 Usage API 查询使用情况

Usage API 返回按可配置维度细分的成本和用量数据：

```python
import oci
from datetime import datetime, timedelta

config = oci.config.from_file("~/.oci/config")
usage_api = oci.usage_api.UsageapiClient(config)

# 查询最近 30 天按服务细分的支出
response = usage_api.request_summarized_usages(
    oci.usage_api.models.RequestSummarizedUsagesDetails(
        tenant_id=config["tenancy"],
        time_usage_started=(datetime.utcnow() - timedelta(days=30)).isoformat() + "Z",
        time_usage_ended=datetime.utcnow().isoformat() + "Z",
        granularity="DAILY",
        query_type="COST",
        group_by=["service"]
    )
)

total_cost = 0.0
for item in response.data.items:
    cost = item.computed_amount or 0
    total_cost += cost
    if cost > 0:
        print(f"{item.service}: ${cost:.2f} ({item.currency})")

print(f"\n30 天总支出: ${total_cost:.2f}")
```

### 步骤 2：按区间和形状细分成本

识别哪些区间和形状正在推高您的账单：

```python
# 按区间统计成本
response = usage_api.request_summarized_usages(
    oci.usage_api.models.RequestSummarizedUsagesDetails(
        tenant_id=config["tenancy"],
        time_usage_started=(datetime.utcnow() - timedelta(days=30)).isoformat() + "Z",
        time_usage_ended=datetime.utcnow().isoformat() + "Z",
        granularity="MONTHLY",
        query_type="COST",
        group_by=["compartmentName", "skuName"]
    )
)

for item in response.data.items:
    cost = item.computed_amount or 0
    if cost > 1.0:  # 过滤噪音
        print(f"{item.compartment_name} | {item.sku_name}: ${cost:.2f}")
```

### 步骤 3：创建带有警报规则的预算

预算会在支出超过阈值之前发出警告。通过 SDK 创建它们，而不是在控制台中搜索：

```python
budget_client = oci.budget.BudgetClient(config)

# 为特定区间创建月度预算
budget = budget_client.create_budget(
    oci.budget.models.CreateBudgetDetails(
        compartment_id=config["tenancy"],
        target_type="COMPARTMENT",
        targets=["ocid1.compartment.oc1..example"],
        amount=500.0,
        reset_period="MONTHLY",
        display_name="开发环境预算",
        description="开发区间的月度预算"
    )
).data

print(f"预算已创建: {budget.id}")

# 在 80% 阈值添加警报规则
budget_client.create_alert_rule(
    budget_id=budget.id,
    create_alert_rule_details=oci.budget.models.CreateAlertRuleDetails(
        type="ACTUAL",
        threshold=80.0,
        threshold_type="PERCENTAGE",
        display_name="80% 警告",
        recipients="oncall@example.com",
        message="开发预算已达到 500 美元月度限制的 80%。"
    )
)

# 在 95% 添加严重警报
budget_client.create_alert_rule(
    budget_id=budget.id,
    create_alert_rule_details=oci.budget.models.CreateAlertRuleDetails(
        type="ACTUAL",
        threshold=95.0,
        threshold_type="PERCENTAGE",
        display_name="95% 严重",
        recipients="oncall@example.com",
        message="严重：开发预算达到 95%。立即检查。"
    )
)
print("警报规则已创建：80% 警告 + 95% 严重")
```

### 步骤 4：使用预测支出预测预算

设置基于预测的警报，如果当前消耗率将超过预算则发出警告：

```python
budget_client.create_alert_rule(
    budget_id=budget.id,
    create_alert_rule_details=oci.budget.models.CreateAlertRuleDetails(
        type="FORECAST",
        threshold=100.0,
        threshold_type="PERCENTAGE",
        display_name="预测超支",
        recipients="oncall@example.com",
        message="根据当前使用情况，预测支出将超过月度预算。"
    )
)
print("预测警报已创建：如果消耗率超过预算则发出警告")
```

### 步骤 5：成本优化策略

应用这些策略来降低 OCI 支出：

**Always Free 层级资源** —— 免费运行开发/测试工作负载：
- 2 个 AMD 计算 VM（每个 1/8 OCPU，1 GB）或 4 个 Arm A1 VM（总共 24 GB）
- 总共 200 GB 块存储，10 GB 对象存储
- 1 个自治数据库（20 GB），10 Mbps 负载均衡器
- 监控：5 亿摄取数据点，10 亿检索数据点

**抢占式实例** —— 对于容错批处理作业可节省高达 50%：

```python
compute = oci.core.ComputeClient(config)

compute.launch_instance(
    oci.core.models.LaunchInstanceDetails(
        compartment_id="ocid1.compartment.oc1..example",
        availability_domain="Uocm:US-ASHBURN-AD-1",
        shape="VM.Standard.E4.Flex",
        shape_config=oci.core.models.LaunchInstanceShapeConfigDetails(
            ocpus=4.0,
            memory_in_gbs=16.0
        ),
        preemptible_instance_config=oci.core.models.PreemptibleInstanceConfigDetails(
            preemption_action=oci.core.models.TerminatePreemptionAction(
                type="TERMINATE",
                preserve_boot_volume=False
            )
        ),
        source_details=oci.core.models.InstanceSourceViaImageDetails(
            image_id="ocid1.image.oc1..example",
            source_type="image"
        ),
        create_vnic_details=oci.core.models.CreateVnicDetails(
            subnet_id="ocid1.subnet.oc1..example"
        ),
        display_name="batch-worker-preemptible"
    )
)
print("抢占式实例已启动 —— 可节省高达 50% 成本")
```

**预留容量** —— 承诺 1 年或 3 年以获得可预测的折扣。节省金额因形状和期限而异，通常为按需定价的 30-60% 折扣。

## 输出

成功完成将产生：
- Usage API 查询显示按服务、区间和 SKU 细分的成本
- 在 80% 和 95% 阈值带有警报规则的月度预算
- 用于预测超支的基于预测的警报
- 成本优化模式：Always Free 资源、抢占式实例和预留容量指南

## 错误处理

| 错误 | 代码 | 原因 | 解决方案 |
|-------|------|-------|----------|
| NotAuthenticated | 401 | API 密钥错误或租户错误 | 验证 `~/.oci/config` 字段与您的租户匹配 |
| NotAuthorizedOrNotFound | 404 | 缺少 `read usage-reports` 策略 | 添加：`Allow group X to read usage-reports in tenancy` |
| TooManyRequests | 429 | Usage API 速率受限 | 降低查询频率；缓存每日结果 |
| InvalidParameter | 400 | 日期范围或粒度无效 | 确保日期为带 Z 后缀的 ISO 格式；使用 DAILY 或 MONTHLY |
| InternalError | 500 | OCI Usage API 服务问题 | 检查 [OCI 状态](https://ocistatus.oraclecloud.com) 并重试 |
| ServiceError status -1 | N/A | 大型成本查询超时 | 缩小日期范围或减少 `group_by` 维度 |

## 示例

**使用 OCI CLI 快速成本检查：**

```bash
# 列出租户中的所有预算
oci budgets budget list --compartment-id $OCI_TENANCY_OCID

# 获取当月支出摘要
oci usage-api usage-summary request-summarized-usages \
  --tenant-id $OCI_TENANCY_OCID \
  --time-usage-started "$(date -u -d '30 days ago' +%Y-%m-%dT%H:%M:%SZ)" \
  --time-usage-ended "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  --granularity MONTHLY \
  --query-type COST
```

**每日成本跟踪脚本：**

```python
import oci
from datetime import datetime, timedelta

config = oci.config.from_file("~/.oci/config")
usage_api = oci.usage_api.UsageapiClient(config)

# 昨日支出
yesterday = datetime.utcnow() - timedelta(days=1)
response = usage_api.request_summarized_usages(
    oci.usage_api.models.RequestSummarizedUsagesDetails(
        tenant_id=config["tenancy"],
        time_usage_started=yesterday.replace(hour=0, minute=0).isoformat() + "Z",
        time_usage_ended=yesterday.replace(hour=23, minute=59).isoformat() + "Z",
        granularity="DAILY",
        query_type="COST",
        group_by=["service"]
    )
)

daily_total = sum(i.computed_amount or 0 for i in response.data.items)
print(f"昨日支出: ${daily_total:.2f}")
for item in sorted(response.data.items, key=lambda x: x.computed_amount or 0, reverse=True)[:5]:
    print(f"  {item.service}: ${item.computed_amount or 0:.2f}")
```

## 资源

- [OCI 成本分析](https://docs.oracle.com/en-us/iaas/Content/Billing/Concepts/costanalysisoverview.htm) —— 控制台成本工具文档
- [OCI 预算](https://docs.oracle.com/en-us/iaas/Content/Billing/Concepts/budgetsoverview.htm) —— 预算创建和警报规则
- [OCI 定价](https://www.oracle.com/cloud/pricing/) —— 当前按服务和形状定价
- [OCI Always Free](https://www.oracle.com/cloud/free/) —— 永久免费层级资源
- [OCI Python SDK](https://docs.oracle.com/en-us/iaas/tools/python/latest/) —— SDK 参考

## 后续步骤

成本监控到位后，查看 `oraclecloud-performance-tuning` 以根据实际指标调整实例大小，或参见 `oraclecloud-observability` 以通过与基础设施警报相同的通知主题路由预算警报。
