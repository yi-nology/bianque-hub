---
name: alicloud-ops
description: 阿里云ECS云资源全局监控实战Skill，基于阿里云OpenAPI实现对ECS实例的全生命周期管理、批量监控、自动化运维。无需SSH登录即可实现跨地域、跨实例的统一监控和管理。
mode: on_demand
version: 1.1.0
maturity: experimental
---

# 阿里云 ECS OpenAPI 全局监控运维指南

## 简介

本 Skill 专注于通过**阿里云 OpenAPI** 实现对 ECS 云资源的**全局监控和管理**，适用于：

- **批量管理**多台 ECS 实例
- **跨地域**统一监控
- **无公网/内网环境**下的资源管理
- **自动化运维**场景
- **安全审计**和合规管理

相比 SSH 直连，OpenAPI 方式具有以下优势：
- ✅ 无需暴露公网端口，更安全
- ✅ 批量操作数百台实例
- ✅ 跨地域全局视图
- ✅ 与阿里云监控、日志、告警深度集成
- ✅ 审计日志完整，可追溯

---

## 监控管理方法对比

| 方式 | 适用场景 | 优点 | 缺点 |
|------|----------|------|------|
| **SSH 直连** | 单机深度运维 | 功能完整、实时性好 | 需开放公网/内网端口，存在安全风险 |
| **OpenAPI 接入** ⭐ | 云资源全局监控 | 批量管理、安全审计、跨地域视图 | 依赖阿里云 API 能力 |
| **代理 上报** | 内网/无公网 ECS | 内网通信、数据自主可控 | 需部署维护 代理 |
| **会话管理** | 免公网登录 | 无需开放端口、临时授权 | 仅限交互式操作 |

---

## 前置要求

### 1. 阿里云 CLI 安装配置

```bash
# 手动安装
# Windows
powershell -Command "Invoke-WebRequest -Uri https://aliyuncli.alicdn.com/aliyun-cli-windows-latest-amd64.zip -OutFile aliyun-cli.zip"

# Linux/Mac
curl -fsSL https://aliyuncli.alicdn.com/aliyun-cli-linux-latest-amd64.tgz -o /tmp/aliyun-cli.tgz
mkdir -p ~/.local/bin
tar -xzf /tmp/aliyun-cli.tgz -C /tmp
mv /tmp/aliyun ~/.local/bin/aliyun
chmod +x ~/.local/bin/aliyun

# 验证安装
aliyun version
```

### 2. 配置阿里云凭证（RAM 子账号推荐）

```bash
# 配置默认 profile
aliyun configure set \
  --profile default \
  --mode AK \
  --access-key-id <YourAccessKeyId> \
  --access-key-secret <YourAccessKeySecret> \
  --region cn-hangzhou

# 查看已配置的 profile
aliyun configure list
```

### 3. 环境变量方式（推荐用于脚本/CICD）

```bash
export ALICLOUD_ACCESS_KEY_ID=<YourAccessKeyId>
export ALICLOUD_ACCESS_KEY_SECRET=<YourAccessKeySecret>
export ALICLOUD_REGION_ID=cn-hangzhou
```

### 4. RAM 权限配置（最小权限原则）

```json
{
  "Version": "1",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ecs:Describe*",
        "ecs:StartInstance",
        "ecs:StopInstance",
        "ecs:RebootInstance",
        "ecs:CreateImage",
        "ecs:DescribeSecurityGroups",
        "cms:Describe*",
        "vpc:DescribeVpcs",
        "vpc:DescribeVSwitches"
      ],
      "Resource": "*"
    }
  ]
}
```

---

## 场景 1：云资源全局资产盘点

### 场景描述
获取全地域 ECS 资产清单，建立云资源全局视图

### 跨地域实例查询

```bash
#!/bin/bash
# 全局资产盘点脚本

echo "========== 阿里云 ECS 全局资产盘点 =========="
echo "生成时间: $(date '+%Y-%m-%d %H:%M:%S')"
echo ""

# 定义需要查询的地域列表
REGIONS=(
    "cn-hangzhou"    # 华东1
    "cn-shanghai"    # 华东2
    "cn-beijing"     # 华北2
    "cn-shenzhen"    # 华南1
    "cn-chengdu"     # 西南1
    "cn-hongkong"    # 香港
)

# 汇总数据
TOTAL_INSTANCES=0
TOTAL_RUNNING=0
TOTAL_STOPPED=0

echo "【地域分布统计】"
printf "%-15s %-10s %-10s %-10s %-15s\n" "地域" "总数" "运行中" "已停止" "总CPU核数"

for REGION in "${REGIONS[@]}"; do
    # 查询该地域实例
    RESULT=$(aliyun ecs DescribeInstances \
        --RegionId $REGION \
        --PageSize 100 2>/dev/null)

    COUNT=$(echo "$RESULT" | jq -r '.TotalCount // 0')
    RUNNING=$(echo "$RESULT" | jq -r '.Instances.Instance[]? | select(.Status == "Running") | .InstanceId' | wc -l)
    STOPPED=$(echo "$RESULT" | jq -r '.Instances.Instance[]? | select(.Status == "Stopped") | .InstanceId' | wc -l)
    CPU_CORES=$(echo "$RESULT" | jq -r '[.Instances.Instance[]? | .Cpu] | add // 0')

    TOTAL_INSTANCES=$((TOTAL_INSTANCES + COUNT))
    TOTAL_RUNNING=$((TOTAL_RUNNING + RUNNING))
    TOTAL_STOPPED=$((TOTAL_STOPPED + STOPPED))

    printf "%-15s %-10s %-10s %-10s %-15s\n" "$REGION" "$COUNT" "$RUNNING" "$STOPPED" "$CPU_CORES"
done

echo ""
echo "【全局汇总】"
echo "  实例总数: $TOTAL_INSTANCES"
echo "  运行中: $TOTAL_RUNNING"
echo "  已停止: $TOTAL_STOPPED"
echo "  运行率: $(awk "BEGIN {printf \"%.1f%%\", ($TOTAL_RUNNING/$TOTAL_INSTANCES)*100}")"
```

### 生成详细资产报表

```bash
#!/bin/bash
# 生成 CSV 格式的资产报表

OUTPUT_FILE="ecs_inventory_$(date +%Y%m%d).csv"
REGIONS=("cn-hangzhou" "cn-shanghai" "cn-beijing" "cn-shenzhen")

# CSV 表头
echo "实例ID,实例名称,地域,可用区,状态,实例类型,CPU,内存(GB),公网IP,内网IP,创建时间,到期时间,操作系统" > "$OUTPUT_FILE"

for REGION in "${REGIONS[@]}"; do
    echo "正在查询地域: $REGION..."

    aliyun ecs DescribeInstances \
        --RegionId $REGION \
        --PageSize 100 \
        --output cols=InstanceId,InstanceName,ZoneId,Status,InstanceType,Cpu,Memory,PublicIpAddress.IpAddress,NetworkInterfaces.NetworkInterface[]?.PrimaryIpAddress,CreatedTime,ExpiredTime,OSType \
        2>/dev/null | while read line; do
        # 处理并追加到 CSV
        echo "$REGION,$line" >> "$OUTPUT_FILE"
    done
done

echo "资产报表已生成: $OUTPUT_FILE"
```

### 标签维度资源分组

```bash
#!/bin/bash
# 按标签统计资源分布

echo "========== 按标签维度统计资源 =========="

# 查询带有特定标签的实例
aliyun ecs DescribeInstances \
    --RegionId cn-hangzhou \
    --Tag.1.Key=Environment \
    --Tag.1.Value=Production \
    --output cols=InstanceId,InstanceName,Status rows=all

# 统计各环境资源数量
echo -e "\n【环境分布统计】"
for ENV in Production Staging Development Testing; do
    COUNT=$(aliyun ecs DescribeInstances \
        --RegionId cn-hangzhou \
        --Tag.1.Key=Environment \
        --Tag.1.Value=$ENV \
        --output cols=InstanceId 2>/dev/null | wc -l)
    echo "  $ENV: $COUNT 台"
done

# 按业务线统计
echo -e "\n【业务线分布统计】"
for BUSINESS in Web API Database Cache; do
    COUNT=$(aliyun ecs DescribeInstances \
        --RegionId cn-hangzhou \
        --Tag.1.Key=BusinessLine \
        --Tag.1.Value=$BUSINESS \
        --output cols=InstanceId 2>/dev/null | wc -l)
    echo "  $BUSINESS: $COUNT 台"
done
```

---

## 场景 2：全局监控数据采集

### 场景描述
无需登录实例，通过 OpenAPI 批量获取监控指标

### CPU 使用率全局监控

```bash
#!/bin/bash
# CPU 使用率监控脚本

REGION="cn-hangzhou"
PERIOD=3600  # 1小时粒度
START_TIME=$(date -u -d '1 hour ago' '+%Y-%m-%dT%H:%M:%SZ')
END_TIME=$(date -u '+%Y-%m-%dT%H:%M:%SZ')

echo "========== ECS CPU 使用率监控 =========="
echo "时间范围: $START_TIME ~ $END_TIME"
echo ""

# 获取所有运行中实例
INSTANCES=$(aliyun ecs DescribeInstances \
    --RegionId $REGION \
    --Status Running \
    --PageSize 100)

# 遍历查询每个实例的 CPU 使用率
for INSTANCE_ID in $(echo "$INSTANCES" | jq -r '.Instances.Instance[].InstanceId'); do
    INSTANCE_NAME=$(echo "$INSTANCES" | jq -r ".Instances.Instance[] | select(.InstanceId == \"$INSTANCE_ID\") | .InstanceName")

    # 查询 CPU 监控数据
    METRICS=$(aliyun cms DescribeMetricList \
        --Namespace acs_ecs_dashboard \
        --MetricName CPUUtilization \
        --Dimensions "[{\"instanceId\":\"$INSTANCE_ID\"}]" \
        --StartTime "$START_TIME" \
        --EndTime "$END_TIME" \
        --Period $PERIOD 2>/dev/null)

    # 解析最新数据点
    LATEST_VALUE=$(echo "$METRICS" | jq -r '.Datapoints | fromjson? | sort_by(.timestamp) | last | .Average // "N/A"')
    MAX_VALUE=$(echo "$METRICS" | jq -r '.Datapoints | fromjson? | max_by(.Maximum) | .Maximum // "N/A"')

    # 高亮显示高 CPU 实例
    if (( $(echo "$LATEST_VALUE > 80" | bc -l 2>/dev/null) )); then
        echo "[警告] $INSTANCE_ID ($INSTANCE_NAME) - 当前: ${LATEST_VALUE}%, 峰值: ${MAX_VALUE}%"
    else
        echo "[正常] $INSTANCE_ID ($INSTANCE_NAME) - 当前: ${LATEST_value}%, 峰值: ${MAX_VALUE}%"
    fi
done
```

### 内存使用率监控（需安装云监控插件）

```bash
#!/bin/bash
# 内存使用率监控脚本

REGION="cn-hangzhou"
START_TIME=$(date -u -d '1 hour ago' '+%Y-%m-%dT%H:%M:%SZ')
END_TIME=$(date -u '+%Y-%m-%dT%H:%M:%SZ')

echo "========== ECS 内存使用率监控 =========="

# 查询内存使用率指标
aliyun ecs DescribeInstances \
    --RegionId $REGION \
    --Status Running \
    --PageSize 100 | jq -r '.Instances.Instance[].InstanceId' | while read INSTANCE_ID; do

    # memory_usedutilization 需要云监控插件
    METRICS=$(aliyun cms DescribeMetricList \
        --Namespace acs_ecs_dashboard \
        --MetricName memory_usedutilization \
        --Dimensions "[{\"instanceId\":\"$INSTANCE_ID\"}]" \
        --StartTime "$START_TIME" \
        --EndTime "$END_TIME" \
        --Period 3600 2>/dev/null)

    VALUE=$(echo "$METRICS" | jq -r '.Datapoints | fromjson? | last | .Average // "N/A"')

    if [ "$VALUE" != "N/A" ]; then
        echo "$INSTANCE_ID - 内存使用率: ${VALUE}%"
    fi
done
```

### 磁盘使用率监控

```bash
#!/bin/bash
# 磁盘使用率监控脚本

REGION="cn-hangzhou"
START_TIME=$(date -u -d '1 hour ago' '+%Y-%m-%dT%H:%M:%SZ')
END_TIME=$(date -u '+%Y-%m-%dT%H:%M:%SZ')

echo "========== ECS 磁盘使用率监控 =========="

# 查询磁盘使用率（需要云监控插件）
aliyun ecs DescribeInstances \
    --RegionId $REGION \
    --Status Running \
    --PageSize 100 | jq -r '.Instances.Instance[].InstanceId' | while read INSTANCE_ID; do

    # 查询磁盘使用率（根分区）
    METRICS=$(aliyun cms DescribeMetricList \
        --Namespace acs_ecs_dashboard \
        --MetricName diskusage_utilization \
        --Dimensions "[{\"instanceId\":\"$INSTANCE_ID\", \"device\":\"/dev/vda1\"}]" \
        --StartTime "$START_TIME" \
        --EndTime "$END_TIME" \
        --Period 3600 2>/dev/null)

    VALUE=$(echo "$METRICS" | jq -r '.Datapoints | fromjson? | last | .Average // "N/A"')

    # 报警阈值
    if (( $(echo "$VALUE > 85" | bc -l 2>/dev/null) )); then
        echo "[警告] $INSTANCE_ID - 磁盘使用率: ${VALUE}%"
    elif [ "$VALUE" != "N/A" ]; then
        echo "[正常] $INSTANCE_ID - 磁盘使用率: ${VALUE}%"
    fi
done
```

### 网络流量监控

```bash
#!/bin/bash
# 网络流量监控脚本

REGION="cn-hangzhou"
INSTANCE_ID="i-bp1xxxxxxxxxxxxxx"
START_TIME=$(date -u -d '24 hours ago' '+%Y-%m-%dT%H:%M:%SZ')
END_TIME=$(date -u '+%Y-%m-%dT%H:%M:%SZ')

echo "========== ECS 网络流量监控 =========="
echo "实例ID: $INSTANCE_ID"
echo "时间范围: 过去24小时"
echo ""

# 公网出带宽
echo "【公网出带宽峰值】"
aliyun cms DescribeMetricList \
    --Namespace acs_ecs_dashboard \
    --MetricName InternetOutRate \
    --Dimensions "[{\"instanceId\":\"$INSTANCE_ID\"}]" \
    --StartTime "$START_TIME" \
    --EndTime "$END_TIME" \
    --Period 3600 2>/dev/null | jq -r '.Datapoints | fromjson? | max_by(.Maximum) | "峰值: \(.Maximum) Kbps, 时间: \(.timestamp)"'

# 公网入带宽
echo -e "\n【公网入带宽峰值】"
aliyun cms DescribeMetricList \
    --Namespace acs_ecs_dashboard \
    --MetricName InternetInRate \
    --Dimensions "[{\"instanceId\":\"$INSTANCE_ID\"}]" \
    --StartTime "$START_TIME" \
    --EndTime "$END_TIME" \
    --Period 3600 2>/dev/null | jq -r '.Datapoints | fromjson? | max_by(.Maximum) | "峰值: \(.Maximum) Kbps, 时间: \(.timestamp)"'

# VPC 内网流量
echo -e "\n【VPC 内网出带宽峰值】"
aliyun cms DescribeMetricList \
    --Namespace acs_ecs_dashboard \
    --MetricName IntranetOutRate \
    --Dimensions "[{\"instanceId\":\"$INSTANCE_ID\"}]" \
    --StartTime "$START_TIME" \
    --EndTime "$END_TIME" \
    --Period 3600 2>/dev/null | jq -r '.Datapoints | fromjson? | max_by(.Maximum) | "峰值: \(.Maximum) Kbps, 时间: \(.timestamp)"'
```

### 综合监控大盘

```bash
#!/bin/bash
# 综合监控数据收集

REGION="cn-hangzhou"
OUTPUT_FILE="monitoring_report_$(date +%Y%m%d_%H%M).json"

echo "========== 生成综合监控报告 =========="

# 获取运行中实例列表
INSTANCES=$(aliyun ecs DescribeInstances --RegionId $REGION --Status Running --PageSize 100)

# 构建报告数据
REPORT=$(echo "$INSTANCES" | jq -r '.Instances.Instance[]?.InstanceId' | while read ID; do
    START_TIME=$(date -u -d '1 hour ago' '+%Y-%m-%dT%H:%M:%SZ')
    END_TIME=$(date -u '+%Y-%m-%dT%H:%M:%SZ')

    # 并行查询多个指标
    CPU=$(aliyun cms DescribeMetricList --Namespace acs_ecs_dashboard --MetricName CPUUtilization --Dimensions "[{\"instanceId\":\"$ID\"}]" --StartTime "$START_TIME" --EndTime "$END_TIME" --Period 3600 2>/dev/null | jq -r '.Datapoints | fromjson? | last | .Average // 0')
    MEMORY=$(aliyun cms DescribeMetricList --Namespace acs_ecs_dashboard --MetricName memory_usedutilization --Dimensions "[{\"instanceId\":\"$ID\"}]" --StartTime "$START_TIME" --EndTime "$END_TIME" --Period 3600 2>/dev/null | jq -r '.Datapoints | fromjson? | last | .Average // 0')

    echo "{\"instanceId\":\"$ID\",\"cpu\":$CPU,\"memory\":$MEMORY}"
done | jq -s '.')

# 保存报告
echo "$REPORT" > "$OUTPUT_FILE"
echo "监控报告已保存: $OUTPUT_FILE"

# 生成摘要
echo -e "\n【监控摘要】"
echo "$REPORT" | jq -r '
    group_by(.cpu > 80) |
    {"high_cpu": (map(select(.cpu > 80)) | length),
     "normal": (map(select(.cpu <= 80)) | length),
     "avg_cpu": (map(.cpu) | add / length)}
'
```

---

## 场景 3：批量实例生命周期管理

### 场景描述
通过 OpenAPI 批量启停实例，实现自动化运维

### 批量启动实例

```bash
#!/bin/bash
# 批量启动实例脚本

REGION="cn-hangzhou"
INSTANCE_IDS=("i-bp1xxxxxxxxx1" "i-bp1xxxxxxxxx2" "i-bp1xxxxxxxxx3")

echo "========== 批量启动实例 =========="

for INSTANCE_ID in "${INSTANCE_IDS[@]}"; do
    echo "正在启动实例: $INSTANCE_ID..."

    RESULT=$(aliyun ecs StartInstance \
        --RegionId $REGION \
        --InstanceId $INSTANCE_ID 2>&1)

    if [ $? -eq 0 ]; then
        echo "  ✓ 启动指令已发送"
    else
        echo "  ✗ 启动失败: $RESULT"
    fi
done

# 等待实例进入 Running 状态
echo -e "\n等待实例启动完成..."
sleep 30

for INSTANCE_ID in "${INSTANCE_IDS[@]}"; do
    STATUS=$(aliyun ecs DescribeInstances \
        --RegionId $REGION \
        --InstanceIds "[\"$INSTANCE_ID\"]" \
        --output cols=Status rows=1 2>/dev/null)
    echo "  $INSTANCE_ID: $STATUS"
done
```

### 批量停止实例（优雅停机）

```bash
#!/bin/bash
# 批量停止实例脚本

REGION="cn-hangzhou"
INSTANCE_IDS=("i-bp1xxxxxxxxx1" "i-bp1xxxxxxxxx2")
STOP_MODE="Stop"  # Stop 或 ForceStop

echo "========== 批量停止实例 =========="
echo "停止模式: $STOP_MODE"

for INSTANCE_ID in "${INSTANCE_IDS[@]}"; do
    echo "正在停止实例: $INSTANCE_ID..."

    if [ "$STOP_MODE" = "ForceStop" ]; then
        RESULT=$(aliyun ecs StopInstance \
            --RegionId $REGION \
            --InstanceId $INSTANCE_ID \
            --ForceStop true 2>&1)
    else
        RESULT=$(aliyun ecs StopInstance \
            --RegionId $REGION \
            --InstanceId $INSTANCE_ID \
            --ForceStop false 2>&1)
    fi

    if [ $? -eq 0 ]; then
        echo "  ✓ 停止指令已发送"
    else
        echo "  ✗ 停止失败: $RESULT"
    fi
done

# 等待实例停止
echo -e "\n等待实例停止..."
for i in {1..30}; do
    ALL_STOPPED=true
    for INSTANCE_ID in "${INSTANCE_IDS[@]}"; do
        STATUS=$(aliyun ecs DescribeInstances \
            --RegionId $REGION \
            --InstanceIds "[\"$INSTANCE_ID\"]" \
            --output cols=Status rows=1 2>/dev/null)

        if [ "$STATUS" != "Stopped" ]; then
            ALL_STOPPED=false
            break
        fi
    done

    if $ALL_STOPPED; then
        echo "  所有实例已停止"
        break
    fi

    echo "  等待中... ($i/30)"
    sleep 10
done
```

### 定时批量操作（开发测试环境节省成本）

```bash
#!/bin/bash
# 定时启停脚本 - 用于节省开发测试环境成本

REGION="cn-hangzhou"
TAG_KEY="Environment"
TAG_VALUE="Development"
ACTION=$1  # start 或 stop

if [ -z "$ACTION" ]; then
    echo "用法: $0 [start|stop]"
    exit 1
fi

echo "========== 定时批量操作 =========="
echo "操作: $ACTION"
echo "目标标签: $TAG_KEY=$TAG_VALUE"

# 查询带标签的实例
INSTANCE_IDS=$(aliyun ecs DescribeInstances \
    --RegionId $REGION \
    --Tag.1.Key=$TAG_KEY \
    --Tag.1.Value=$TAG_VALUE \
    --PageSize 100 2>/dev/null | jq -r '.Instances.Instance[].InstanceId')

if [ -z "$INSTANCE_IDS" ]; then
    echo "未找到匹配的实例"
    exit 0
fi

echo "找到 $(echo "$INSTANCE_IDS" | wc -l) 个实例"

# 执行批量操作
for INSTANCE_ID in $INSTANCE_IDS; do
    case $ACTION in
        start)
            echo "启动实例: $INSTANCE_ID"
            aliyun ecs StartInstance --RegionId $REGION --InstanceId $INSTANCE_ID > /dev/null 2>&1
            ;;
        stop)
            echo "停止实例: $INSTANCE_ID"
            aliyun ecs StopInstance --RegionId $REGION --InstanceId $INSTANCE_ID > /dev/null 2>&1
            ;;
        *)
            echo "未知操作: $ACTION"
            exit 1
            ;;
    esac
done

echo "操作完成"
```

### Crontab 配置示例

```bash
# 编辑 crontab
crontab -e

# 每天早上 9 点启动开发环境
0 9 * * 1-5 /path/to/batch_control.sh start >> /var/log/ecs_control.log 2>&1

# 每天晚上 8 点停止开发环境
0 20 * * 1-5 /path/to/batch_control.sh stop >> /var/log/ecs_control.log 2>&1
```

---

## 场景 4：安全组批量管理

### 场景描述
统一配置和管理安全组规则，确保网络安全合规

### 安全组规则审计

```bash
#!/bin/bash
# 安全组规则审计脚本

REGION="cn-hangzhou"

echo "========== 安全组规则审计 =========="

# 获取所有安全组
SECURITY_GROUPS=$(aliyun ecs DescribeSecurityGroups --RegionId $REGION --PageSize 50)

echo "$SECURITY_GROUPS" | jq -r '.SecurityGroups.SecurityGroup[].SecurityGroupId' | while read SG_ID; do
    SG_NAME=$(aliyun ecs DescribeSecurityGroupAttribute \
        --RegionId $REGION \
        --SecurityGroupId $SG_ID \
        --Direction ingress 2>/dev/null | jq -r '.SecurityGroupName')

    echo -e "\n【安全组: $SG_NAME ($SG_ID)】"

    # 查询入方向规则
    RULES=$(aliyun ecs DescribeSecurityGroupAttribute \
        --RegionId $REGION \
        --SecurityGroupId $SG_ID \
        --Direction ingress 2>/dev/null)

    # 检查是否有 0.0.0.0/0 的高危端口规则
    echo "$RULES" | jq -r '.Permissions.Permission[]? | select(.SourceCidrIp == "0.0.0.0/0") | "  规则: \(.IpProtocol):\(.PortRange) 源IP: \(.SourceCidrIp)"' | while read RULE; do
        if echo "$RULE" | grep -qE "22|3389|3306|6379|27017"; then
            echo "  [警告] $RULE (高危端口开放到公网)"
        else
            echo "  [信息] $RULE"
        fi
    done
done
```

### 批量添加安全组规则

```bash
#!/bin/bash
# 批量添加安全组规则

REGION="cn-hangzhou"
SECURITY_GROUP_IDS=("sg-bp1xxxxxxxxx1" "sg-bp1xxxxxxxxx2")

# 新规则参数
IP_PROTOCOL="tcp"
PORT_RANGE="8080/8080"
SOURCE_CIDR="10.0.0.0/8"  # 仅允许内网访问
POLICY="accept"
PRIORITY=1
DESCRIPTION="Internal API Access"

echo "========== 批量添加安全组规则 =========="

for SG_ID in "${SECURITY_GROUP_IDS[@]}"; do
    echo "正在配置安全组: $SG_ID..."

    RESULT=$(aliyun ecs AuthorizeSecurityGroup \
        --RegionId $REGION \
        --SecurityGroupId $SG_ID \
        --IpProtocol $IP_PROTOCOL \
        --PortRange $PORT_RANGE \
        --SourceCidrIp $SOURCE_CIDR \
        --Policy $POLICY \
        --Priority $PRIORITY \
        --Description "$DESCRIPTION" 2>&1)

    if [ $? -eq 0 ]; then
        echo "  ✓ 规则添加成功"
    else
        echo "  ✗ 添加失败: $RESULT"
    fi
done
```

### 安全组规则清理

```bash
#!/bin/bash
# 清理过期或测试用的安全组规则

REGION="cn-hangzhou"
SECURITY_GROUP_ID="sg-bp1xxxxxxxxxxxxxx"

echo "========== 安全组规则清理 =========="
echo "安全组: $SECURITY_GROUP_ID"

# 查询现有规则
RULES=$(aliyun ecs DescribeSecurityGroupAttribute \
    --RegionId $REGION \
    --SecurityGroupId $SECURITY_GROUP_ID \
    --Direction ingress 2>/dev/null)

# 删除包含 "test" 或 "temp" 描述的规则
echo "$RULES" | jq -r '.Permissions.Permission[]? | select(.Description | test("test|temp|临时"; "i")) | "\(.IpProtocol)|\(.PortRange)|\(.SourceCidrIp)"' | while IFS='|' read -r PROTOCOL PORT SOURCE; do
    echo "删除规则: $PROTOCOL $PORT from $SOURCE"

    aliyun ecs RevokeSecurityGroup \
        --RegionId $REGION \
        --SecurityGroupId $SECURITY_GROUP_ID \
        --IpProtocol $PROTOCOL \
        --PortRange $PORT \
        --SourceCidrIp "$SOURCE" 2>/dev/null

    if [ $? -eq 0 ]; then
        echo "  ✓ 已删除"
    else
        echo "  ✗ 删除失败"
    fi
done
```

---

## 场景 5：镜像管理与备份

### 场景描述
通过 OpenAPI 批量创建镜像、复制镜像、管理镜像生命周期

### 批量创建系统镜像

```bash
#!/bin/bash
# 批量创建系统镜像备份

REGION="cn-hangzhou"
DATE_SUFFIX=$(date +%Y%m%d)

echo "========== 批量创建系统镜像 =========="

# 从文件读取需要备份的实例列表
INSTANCE_FILE="backup_instances.txt"

if [ ! -f "$INSTANCE_FILE" ]; then
    echo "错误: 实例列表文件 $INSTANCE_FILE 不存在"
    echo "请创建文件，每行一个实例ID，格式: 实例ID|镜像名称|描述"
    exit 1
fi

while IFS='|' read -r INSTANCE_ID IMAGE_NAME IMAGE_DESC; do
    echo "处理实例: $INSTANCE_ID"

    # 创建镜像（不重启实例）
    RESULT=$(aliyun ecs CreateImage \
        --RegionId $REGION \
        --InstanceId $INSTANCE_ID \
        --ImageName "${IMAGE_NAME}_${DATE_SUFFIX}" \
        --Description "$IMAGE_DESC" \
        --NoReboot true 2>&1)

    if [ $? -eq 0 ]; then
        IMAGE_ID=$(echo "$RESULT" | jq -r '.ImageId')
        echo "  ✓ 镜像创建中，ID: $IMAGE_ID"
    else
        echo "  ✗ 创建失败: $RESULT"
    fi
done < "$INSTANCE_FILE"

echo -e "\n镜像创建任务已提交，请通过控制台或 API 查看进度"
```

### 镜像跨地域复制

```bash
#!/bin/bash
# 镜像跨地域复制，实现灾备

REGION="cn-hangzhou"
IMAGE_ID="m-bp1xxxxxxxxxxxxxx"
TARGET_REGIONS=("cn-beijing" "cn-shanghai" "cn-shenzhen")

echo "========== 镜像跨地域复制 =========="
echo "源镜像: $IMAGE_ID"
echo "源地域: $REGION"
echo ""

for TARGET_REGION in "${TARGET_REGIONS[@]}"; do
    echo "复制到地域: $TARGET_REGION..."

    RESULT=$(aliyun ecs CopyImage \
        --RegionId $REGION \
        --ImageId $IMAGE_ID \
        --DestinationRegionId $TARGET_REGION \
        --DestinationImageName "Backup_$(date +%Y%m%d)_${TARGET_REGION}" 2>&1)

    if [ $? -eq 0 ]; then
        NEW_IMAGE_ID=$(echo "$RESULT" | jq -r '.ImageId')
        echo "  ✓ 复制任务已启动，新镜像ID: $NEW_IMAGE_ID"
    else
        echo "  ✗ 复制失败: $RESULT"
    fi
done
```

### 镜像生命周期管理

```bash
#!/bin/bash
# 自动清理过期镜像

REGION="cn-hangzhou"
RETENTION_DAYS=30

echo "========== 镜像生命周期管理 =========="
echo "保留策略: ${RETENTION_DAYS}天"
echo ""

# 获取自定义镜像列表
IMAGES=$(aliyun ecs DescribeImages \
    --RegionId $REGION \
    --ImageOwnerAlias self \
    --PageSize 100 2>/dev/null)

# 计算过期时间
CUTOFF_DATE=$(date -d "${RETENTION_DAYS} days ago" +%Y-%m-%d)

echo "过期截止时间: $CUTOFF_DATE"
echo ""

# 遍历并删除过期镜像
echo "$IMAGES" | jq -r ".Images.Image[]? | select(.CreationTime < \"${CUTOFF_DATE}T\") | \"\(.ImageId)|\(.ImageName)|\(.CreationTime)\"" | while IFS='|' read -r IMG_ID IMG_NAME CREATE_TIME; do
    echo "发现过期镜像:"
    echo "  ID: $IMG_ID"
    echo "  名称: $IMG_NAME"
    echo "  创建时间: $CREATE_TIME"

    # 检查镜像是否被使用
    USAGE=$(aliyun ecs DescribeInstances \
        --RegionId $REGION \
        --ImageId $IMG_ID \
        --PageSize 1 2>/dev/null | jq -r '.TotalCount')

    if [ "$USAGE" -gt 0 ]; then
        echo "  [跳过] 镜像正在被 $USAGE 个实例使用"
        continue
    fi

    # 删除镜像
    RESULT=$(aliyun ecs DeleteImage \
        --RegionId $REGION \
        --ImageId $IMG_ID \
        --Force true 2>&1)

    if [ $? -eq 0 ]; then
        echo "  ✓ 已删除"
    else
        echo "  ✗ 删除失败: $RESULT"
    fi
    echo ""
done
```

---

## 场景 6：事件监控与告警

### 场景描述
通过 OpenAPI 获取系统事件，实现主动运维

### 查询待处理事件

```bash
#!/bin/bash
# 查询系统事件和告警

REGION="cn-hangzhou"

echo "========== 系统事件查询 =========="

# 查询实例系统事件
EVENTS=$(aliyun ecs DescribeInstanceHistoryEvents \
    --RegionId $REGION \
    --PageSize 50 2>/dev/null)

echo "【近期系统事件】"
echo "$EVENTS" | jq -r '.InstanceSystemEventSet.InstanceSystemEventType[]? | "[\(.EventPublishTime)] \(.EventType.Name) - 实例: \(.InstanceId) - 状态: \(.EventStatus.Name)"'

# 查询计划内维护事件
echo -e "\n【计划内维护】"
aliyun ecs DescribeInstances \
    --RegionId $REGION \
    --InstanceChargeType PostPaid \
    --PageSize 50 2>/dev/null | jq -r '.Instances.Instance[]? | select(.OperationLocks.LockReason[]?.LockReason == "security") | "警告: 实例 \(.InstanceId) 存在安全锁定"'
```

### 自动化事件响应

```bash
#!/bin/bash
# 自动响应系统事件

REGION="cn-hangzhou"
LOG_FILE="/var/log/ecs_events.log"

echo "========== 自动化事件响应 =========="
echo "$(date): 开始事件检查" >> "$LOG_FILE"

# 查询待处理事件
EVENTS=$(aliyun ecs DescribeInstanceHistoryEvents \
    --RegionId $REGION \
    --EventStatus Scheduled \
    --PageSize 50 2>/dev/null)

# 处理需要用户响应的事件
echo "$EVENTS" | jq -r '.InstanceSystemEventSet.InstanceSystemEventType[]? | select(.EventType.Name | contains("Reboot")) | "\(.InstanceId)|\(.EventId)"' | while IFS='|' read -r INSTANCE_ID EVENT_ID; do
    echo "发现需要处理的实例重启事件:"
    echo "  实例ID: $INSTANCE_ID"
    echo "  事件ID: $EVENT_ID"

    # 发送通知（这里可以集成钉钉/企业微信/邮件）
    echo "$(date): 实例 $INSTANCE_ID 计划重启，事件ID: $EVENT_ID" >> "$LOG_FILE"

    # 可选：自动同意重启
    # aliyun ecs AcceptInquiredSystemEvent --RegionId $REGION --EventId $EVENT_ID
done
```

---

## 场景 7：自动化运维工作流

### 场景描述
结合多个 API 实现复杂的自动化运维场景

### 自动扩缩容示例

```bash
#!/bin/bash
# 基于监控指标的简单自动扩容逻辑

REGION="cn-hangzhou"
INSTANCE_TYPE="ecs.g7.xlarge"  # 扩容目标规格
SCALE_THRESHOLD=80  # CPU 阈值
echo "========== 自动扩容检查 =========="

# 获取运行中实例
INSTANCES=$(aliyun ecs DescribeInstances \
    --RegionId $REGION \
    --Status Running \
    --PageSize 100)

# 检查每个实例的 CPU 使用率
for INSTANCE_ID in $(echo "$INSTANCES" | jq -r '.Instances.Instance[].InstanceId'); do
    START_TIME=$(date -u -d '10 minutes ago' '+%Y-%m-%dT%H:%M:%SZ')
    END_TIME=$(date -u '+%Y-%m-%dT%H:%M:%SZ')

    AVG_CPU=$(aliyun cms DescribeMetricList \
        --Namespace acs_ecs_dashboard \
        --MetricName CPUUtilization \
        --Dimensions "[{\"instanceId\":\"$INSTANCE_ID\"}]" \
        --StartTime "$START_TIME" \
        --EndTime "$END_TIME" \
        --Period 300 2>/dev/null | jq -r '.Datapoints | fromjson? | map(.Average) | add / length')

    if (( $(echo "$AVG_CPU > $SCALE_THRESHOLD" | bc -l 2>/dev/null) )); then
        echo "实例 $INSTANCE_ID CPU 使用率: ${AVG_CPU}% - 超过阈值"
        echo "触发扩容流程..."

        # 实际扩容逻辑：
        # 1. 创建扩容实例
        # 2. 配置负载均衡
        # 3. 通知管理员

        # 发送告警
        echo "$(date): 实例 $INSTANCE_ID 需要扩容，当前CPU: ${AVG_CPU}%" >> /var/log/scale_events.log
    fi
done
```

### 运维巡检报告生成

```bash
#!/bin/bash
# 生成每日运维巡检报告

REGION="cn-hangzhou"
REPORT_DATE=$(date +%Y-%m-%d)
REPORT_FILE="daily_report_${REPORT_DATE}.html"

echo "========== 生成每日巡检报告 =========="

# 开始生成 HTML 报告
cat > "$REPORT_FILE" << 'EOF'
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>ECS 每日巡检报告</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        h1 { color: #333; }
        h2 { color: #666; border-bottom: 2px solid #ddd; padding-bottom: 10px; }
        table { border-collapse: collapse; width: 100%; margin: 20px 0; }
        th, td { border: 1px solid #ddd; padding: 12px; text-align: left; }
        th { background-color: #4CAF50; color: white; }
        tr:nth-child(even) { background-color: #f2f2f2; }
        .warning { color: #ff9800; font-weight: bold; }
        .danger { color: #f44336; font-weight: bold; }
        .success { color: #4CAF50; }
    </style>
</head>
<body>
EOF

echo "<h1>ECS 每日巡检报告 - $REPORT_DATE</h1>" >> "$REPORT_FILE"

# 1. 资源概览
echo "<h2>1. 资源概览</h2>" >> "$REPORT_FILE"
TOTAL_INSTANCES=$(aliyun ecs DescribeInstances --RegionId $REGION --PageSize 1 | jq -r '.TotalCount')
RUNNING_INSTANCES=$(aliyun ecs DescribeInstances --RegionId $REGION --Status Running --PageSize 1 | jq -r '.TotalCount')
STOPPED_INSTANCES=$((TOTAL_INSTANCES - RUNNING_INSTANCES))

cat >> "$REPORT_FILE" << EOF
<table>
<tr><th>指标</th><th>数量</th></tr>
<tr><td>实例总数</td><td>$TOTAL_INSTANCES</td></tr>
<tr><td>运行中</td><td class="success">$RUNNING_INSTANCES</td></tr>
<tr><td>已停止</td><td>$STOPPED_INSTANCES</td></tr>
</table>
EOF

# 2. 异常实例列表
echo "<h2>2. 高负载实例 (CPU > 80%)</h2>" >> "$REPORT_FILE"
echo "<table><tr><th>实例ID</th><th>实例名称</th><th>CPU使用率</th></tr>" >> "$REPORT_FILE"

START_TIME=$(date -u -d '1 hour ago' '+%Y-%m-%dT%H:%M:%SZ')
END_TIME=$(date -u '+%Y-%m-%dT%H:%M:%SZ')

aliyun ecs DescribeInstances --RegionId $REGION --Status Running --PageSize 100 | jq -r '.Instances.Instance[].InstanceId' | while read ID; do
    NAME=$(aliyun ecs DescribeInstances --RegionId $REGION --InstanceIds "[\"$ID\"]" | jq -r '.Instances.Instance[0].InstanceName')
    CPU=$(aliyun cms DescribeMetricList --Namespace acs_ecs_dashboard --MetricName CPUUtilization --Dimensions "[{\"instanceId\":\"$ID\"}]" --StartTime "$START_TIME" --EndTime "$END_TIME" --Period 3600 2>/dev/null | jq -r '.Datapoints | fromjson? | last | .Average // 0')

    if (( $(echo "$CPU > 80" | bc -l 2>/dev/null) )); then
        echo "<tr><td>$ID</td><td>$NAME</td><td class="danger">${CPU}%</td></tr>" >> "$REPORT_FILE"
    fi
done

echo "</table>" >> "$REPORT_FILE"

# 结束 HTML
echo "</body></html>" >> "$REPORT_FILE"

echo "巡检报告已生成: $REPORT_FILE"
```

---

## 常用命令速查表

### ECS 实例管理

```bash
# 实例查询
aliyun ecs DescribeInstances                    # 查询实例列表
aliyun ecs DescribeInstanceAttribute            # 查询实例详情
aliyun ecs DescribeAvailableResource            # 查询可用资源

# 实例操作
aliyun ecs StartInstance --InstanceId <id>      # 启动实例
aliyun ecs StopInstance --InstanceId <id>       # 停止实例
aliyun ecs RebootInstance --InstanceId <id>     # 重启实例
aliyun ecs DeleteInstance --InstanceId <id>     # 释放实例

# 实例配置
aliyun ecs ModifyInstanceAttribute              # 修改实例属性
aliyun ecs ModifyInstanceSpec                   # 变更实例规格
aliyun ecs RenewInstance                        # 续费实例
```

### 监控数据查询

```bash
# 云监控 API
aliyun cms DescribeMetricList                   # 查询监控数据
aliyun cms DescribeMetricMetaList               # 查询监控指标列表
aliyun cms DescribeAlarmHistory                 # 查询告警历史
aliyun cms DescribeAlarms                       # 查询告警规则

# 常用监控指标
CPUUtilization              # CPU 使用率
memory_usedutilization      # 内存使用率（需安装插件）
diskusage_utilization       # 磁盘使用率（需安装插件）
InternetOutRate            # 公网出带宽
InternetInRate             # 公网入带宽
IntranetOutRate            # 内网出带宽
IntranetInRate             # 内网入带宽
```

### 镜像管理

```bash
aliyun ecs DescribeImages                       # 查询镜像列表
aliyun ecs CreateImage                          # 创建自定义镜像
aliyun ecs CopyImage                            # 复制镜像
aliyun ecs DeleteImage                          # 删除镜像
aliyun ecs ModifyImageSharePermission          # 共享镜像
```

### 安全组管理

```bash
aliyun ecs DescribeSecurityGroups               # 查询安全组列表
aliyun ecs DescribeSecurityGroupAttribute       # 查询安全组规则
aliyun ecs CreateSecurityGroup                  # 创建安全组
aliyun ecs AuthorizeSecurityGroup               # 添加安全组规则
aliyun ecs RevokeSecurityGroup                  # 删除安全组规则
aliyun ecs DeleteSecurityGroup                  # 删除安全组
```

---

## 安全与最佳实践

### 1. 最小权限原则

```json
{
  "Version": "1",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ecs:Describe*",
        "ecs:StartInstance",
        "ecs:StopInstance"
      ],
      "Resource": "*",
      "Condition": {
        "StringEquals": {
          "ecs:tag/Environment": "Development"
        }
      }
    }
  ]
}
```

### 2. 密钥安全管理

```bash
# 不要硬编码密钥
# ❌ 错误示例
aliyun ecs DescribeInstances --access-key-id xxx --access-key-secret xxx

# ✅ 正确做法
# 1. 使用环境变量
export ALICLOUD_ACCESS_KEY_ID=$AK_ID
export ALICLOUD_ACCESS_KEY_SECRET=$AK_SECRET

# 2. 使用配置文件
aliyun configure set --profile production

# 3. 使用 RAM 角色（ECS 实例内）
# 实例绑定 RAM 角色后，自动获取临时凭证
```

### 3. 操作审计

```bash
# 开启操作审计
aliyun actiontrail CreateTrail \
    --Name ecs-audit \
    --SlsProjectArn acs:log:cn-hangzhou:*:project/ecs-audit \
    --EventRW Write \
    --TrailRegion cn-hangzhou

# 查询操作记录
aliyun actiontrail LookupEvents \
    --StartTime 2026-01-01T00:00:00Z \
    --EndTime 2026-01-31T23:59:59Z \
    --ServiceName ECS
```

### 4. 错误处理与重试

```bash
#!/bin/bash
# 带重试机制的 API 调用

api_call_with_retry() {
    local MAX_RETRIES=3
    local RETRY_DELAY=5
    local attempt=1

    while [ $attempt -le $MAX_RETRIES ]; do
        RESULT=$("$@" 2>&1)
        EXIT_CODE=$?

        if [ $EXIT_CODE -eq 0 ]; then
            echo "$RESULT"
            return 0
        fi

        # 检查是否可重试的错误
        if echo "$RESULT" | grep -q "Throttling\|RateLimit\|ServiceUnavailable"; then
            echo "API 限流，${RETRY_DELAY}秒后重试... ($attempt/$MAX_RETRIES)"
            sleep $RETRY_DELAY
            RETRY_DELAY=$((RETRY_DELAY * 2))
            attempt=$((attempt + 1))
        else
            echo "API 调用失败: $RESULT"
            return 1
        fi
    done

    echo "达到最大重试次数，操作失败"
    return 1
}

# 使用示例
api_call_with_retry aliyun ecs DescribeInstances --RegionId cn-hangzhou
```

---

## 输出和报告

- 监控数据保存到 `output/alicloud-ops/monitoring/`
- 资产报表保存到 `output/alicloud-ops/inventory/`
- 巡检报告保存到 `output/alicloud-ops/reports/`
- 操作日志记录到 `/var/log/ecs-ops.log`

## 参考资料

- [阿里云 CLI 官方文档](https://help.aliyun.com/document_detail/121541.html)
- [阿里云 ECS API 参考](https://help.aliyun.com/document_detail/25484.html)
- [阿里云监控 API](https://help.aliyun.com/document_detail/163824.html)
- [RAM 访问控制](https://help.aliyun.com/document_detail/28627.html)
