---
name: oraclecloud-observability
description: 为 OCI 资源设置程序化监控、日志记录和告警。 在配置 OCI 监控指标、创建告警规则、发布自定义指标或通过日志服务搜索日志时使用。 使用 "oraclecloud observability"、"oci monitoring"、"oci alarms"、"oci logging"、"oracle cloud observability" 触发。
mode: on_demand
version: 1.1.0
maturity: experimental
---

# Oracle Cloud 可观测性

## 概述

使用监控、日志记录和通知服务为 OCI 基础设施设置程序化监控。OCI 控制台将这些功能隐藏在嵌套菜单后面，且状态页面历史上未能确认中断（例如，2026 年 1 月伦敦区域）。本技能构建您通过代码控制的监控 —— 指标查询、告警规则、自定义指标发布和日志搜索 —— 这样您就不会对应该捕获的中断感到惊讶。

**目的：** 创建一个代码驱动的可观测性堆栈，无需依赖 OCI 控制台即可查询指标、触发告警、发布自定义指标和搜索日志。

## 前置条件

- 在 `~/.oci/config` 中配置了 API 签名密钥的 **OCI 租户**
- 安装了 `pip install oci` 的 **Python 3.8+**
- 包含要监控资源的 **区间 OCID**
- 授予目标区间中 `manage alarms` 和 `read metrics` 权限的 **IAM 策略**
- 为告警目标创建的 **通知主题**（或在步骤 4 中创建一个）

## 说明

### 步骤 1：使用 MonitoringClient 查询指标

OCI 为计算、网络、块存储等发布内置指标。以编程方式查询它们：

```python
import oci
from datetime import datetime, timedelta

config = oci.config.from_file("~/.oci/config")
monitoring = oci.monitoring.MonitoringClient(config)

# 查询区间中所有实例的 CPU 利用率
response = monitoring.summarize_metrics_data(
    compartment_id="ocid1.compartment.oc1..example",
    summarize_metrics_data_details=oci.monitoring.models.SummarizeMetricsDataDetails(
        namespace="oci_computeagent",
        query='CpuUtilization[5m]{availabilityDomain = "Uocm:US-ASHBURN-AD-1"}.mean()',
        start_time=(datetime.utcnow() - timedelta(hours=1)).isoformat() + "Z",
        end_time=datetime.utcnow().isoformat() + "Z"
    )
)

for metric in response.data:
    for dp in metric.aggregated_datapoints:
        print(f"{dp.timestamp}: {dp.value:.1f}% CPU")
```

### 步骤 2：创建告警规则

告警在指标超过阈值时触发。通过 SDK 创建它们，这样它们就能在控制台 UI 变更中幸存：

```python
monitoring.create_alarm(
    oci.monitoring.models.CreateAlarmDetails(
        display_name="高 CPU 告警",
        compartment_id="ocid1.compartment.oc1..example",
        metric_compartment_id="ocid1.compartment.oc1..example",
        namespace="oci_computeagent",
        query='CpuUtilization[5m].mean() > 80',
        severity="CRITICAL",
        body="CPU 利用率超过 80% 持续 5 分钟。",
        destinations=["ocid1.onstopic.oc1..example"],
        is_enabled=True,
        pending_duration="PT5M",
        repeat_notification_duration="PT15M"
    )
)
print("告警已创建：高 CPU 告警")
```

### 步骤 3：发布自定义指标

将应用程序级指标推送到 OCI 监控，这样它们就可以触发相同的告警系统：

```python
from datetime import datetime

monitoring.post_metric_data(
    oci.monitoring.models.PostMetricDataDetails(
        metric_data=[
            oci.monitoring.models.MetricDataDetails(
                namespace="custom_app",
                compartment_id="ocid1.compartment.oc1..example",
                name="RequestLatencyMs",
                dimensions={"service": "api-gateway", "endpoint": "/v1/orders"},
                datapoints=[
                    oci.monitoring.models.Datapoint(
                        timestamp=datetime.utcnow().isoformat() + "Z",
                        value=142.5
                    )
                ]
            )
        ]
    )
)
print("自定义指标已发布：RequestLatencyMs = 142.5ms")
```

### 步骤 4：设置通知

创建通知主题和电子邮件订阅以接收告警警报：

```python
notifications = oci.ons.NotificationDataPlaneClient(config)
control_plane = oci.ons.NotificationControlPlaneClient(config)

# 创建主题
topic = control_plane.create_topic(
    oci.ons.models.CreateTopicDetails(
        name="infra-alerts",
        compartment_id="ocid1.compartment.oc1..example",
        description="基础设施告警通知"
    )
).data

# 订阅电子邮件端点
notifications.create_subscription(
    oci.ons.models.CreateSubscriptionDetails(
        topic_id=topic.topic_id,
        compartment_id="ocid1.compartment.oc1..example",
        protocol="EMAIL",
        endpoint="oncall@example.com"
    )
)
print(f"主题已创建：{topic.topic_id}")
```

### 步骤 5：搜索日志

查询 OCI 日志服务以查找基础设施中的特定事件：

```python
logging_search = oci.loggingsearch.LogSearchClient(config)

results = logging_search.search_logs(
    oci.loggingsearch.models.SearchLogsDetails(
        time_start=(datetime.utcnow() - timedelta(hours=1)).isoformat() + "Z",
        time_end=datetime.utcnow().isoformat() + "Z",
        search_query=(
            'search "ocid1.compartment.oc1..example" '
            '| where data.statusCode = 500'
        ),
        is_return_field_info=False
    )
)

for log_entry in results.data.results:
    print(f"{log_entry.data}")
```

### 步骤 6：健康检查探针

使用 OCI 健康检查监控端点可用性：

```python
health = oci.healthchecks.HealthChecksClient(config)

health.create_http_monitor(
    oci.healthchecks.models.CreateHttpMonitorDetails(
        compartment_id="ocid1.compartment.oc1..example",
        display_name="API 健康检查",
        targets=["api.example.com"],
        protocol="HTTPS",
        port=443,
        path="/health",
        interval_in_seconds=30,
        timeout_in_seconds=10,
        is_enabled=True
    )
)
print("健康检查探针已创建：api.example.com/health 每 30 秒")
```

## 输出

成功完成将产生：
- 返回您区间 CPU、内存和网络数据的指标查询
- 当阈值被突破时触发到通知主题的告警规则
- 发布到 OCI 监控的自定义应用程序指标
- 带有电子邮件订阅的告警交付通知主题
- 用于排查 500 错误和其他事件的日志搜索查询
- 用于端点可用性监控的 HTTP 健康检查探针

## 错误处理

| 错误 | 代码 | 原因 | 解决方案 |
|-------|------|-------|----------|
| NotAuthenticated | 401 | API 密钥错误或配置过期 | 验证 `~/.oci/config` 指纹与您的 API 密钥匹配 |
| NotAuthorizedOrNotFound | 404 | 缺少监控的 IAM 策略 | 添加：`Allow group X to manage alarms in compartment Y` |
| TooManyRequests | 429 | 指标查询速率受限 | 降低查询频率；为仪表板缓存结果 |
| InternalError | 500 | OCI 监控服务问题 | 检查 [OCI 状态](https://ocistatus.oraclecloud.com) 并重试 |
| InvalidParameter | 400 | MQL 查询语法错误 | 验证命名空间和指标名称；使用 `list_metrics` 发现可用指标 |
| ServiceError status -1 | N/A | 大型查询请求超时 | 缩小时间窗口或添加维度过滤器 |

## 示例

**使用 OCI CLI 快速指标检查：**

```bash
# 列出可用的指标命名空间
oci monitoring metric list \
  --compartment-id ocid1.compartment.oc1..example \
  --namespace oci_computeagent

# 列出所有告警
oci monitoring alarm list \
  --compartment-id ocid1.compartment.oc1..example
```

**列出命名空间中的所有指标以发现可用内容：**

```python
import oci

config = oci.config.from_file("~/.oci/config")
monitoring = oci.monitoring.MonitoringClient(config)

metrics = monitoring.list_metrics(
    compartment_id="ocid1.compartment.oc1..example",
    list_metrics_details=oci.monitoring.models.ListMetricsDetails(
        namespace="oci_computeagent"
    )
).data

for m in metrics:
    print(f"{m.name} — 维度：{m.dimensions}")
```

## 资源

- [OCI 监控](https://docs.oracle.com/en-us/iaas/Content/Monitoring/home.htm) —— 指标、告警和 MQL 查询语言
- [OCI 日志记录](https://docs.oracle.com/en-us/iaas/Content/Logging/home.htm) —— 集中式日志服务
- [OCI 通知](https://docs.oracle.com/en-us/iaas/Content/Notification/home.htm) —— 通过电子邮件、Slack、PagerDuty 发送告警
- [OCI Python SDK](https://docs.oracle.com/en-us/iaas/tools/python/latest/) —— SDK 参考
- [OCI 已知问题](https://docs.oracle.com/en-us/iaas/Content/knownissues.htm) —— 当前平台问题

## 后续步骤

监控到位后，继续到 `oraclecloud-performance-tuning` 以优化形状和存储性能，或参见 `oraclecloud-cost-tuning` 以设置使用相同通知主题的预算警报。
