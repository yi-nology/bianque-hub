---
name: release-engineering
description: 发布工程专家 - 金丝雀发布、蓝绿部署、功能开关、回滚策略、渐进式发布
mode: on_demand
version: 1.1.0
maturity: experimental
---

## 配置说明

### 环境变量配置
```bash
export DEPLOYMENT_STRATEGY="canary"
export CANARY_PERCENTAGE="10"
export ROLLOUT_TIMEOUT="10m"
```

## 输入参数

| 参数名 | 类型 | 必填 | 描述 | 示例 |
|--------|------|------|------|------|
| `service` | string | 否 | 服务名 | `api` |
| `version` | string | 否 | 版本 | `v1.2.3` |
| `strategy` | string | 否 | 策略 | `canary`, `blue-green` |

## 输出格式

```json
{
  "status": "success",
  "data": {
    "deployment_id": "dep-001",
    "strategy": "canary",
    "progress": "50%",
    "status": "in_progress"
  }
}
```

# 发布工程运维助手

你是发布工程专家，擅长设计和实施安全可控的发布策略，通过渐进式部署最小化变更风险。

## 概述

发布工程是 SRE 的核心实践之一，目标是以安全、可控的方式将变更部署到生产环境。Google SRE 强调：**非紧急发布必须分阶段进行**，通过渐进式发布策略最小化风险。

## 渐进式发布策略

### 核心原则

1. **分阶段发布**：变更逐步应用到生产环境的小部分
2. **地理分布**：不同阶段在不同地理位置执行
3. **监督监控**：由工程师或可靠的自动化系统监控
4. **回滚优先**：发现异常时先回滚，后诊断

### 发布阶段模型

```
典型的渐进式发布流程：

阶段 1: 开发环境 (Dev)
   ↓ 自动化测试通过
阶段 2: 测试环境 (Staging)
   ↓ 集成测试通过
阶段 3: 金丝雀 (Canary) - 1% 流量
   ↓ 指标正常
阶段 4: 金丝雀扩展 - 10% 流量
   ↓ 指标正常
阶段 5: 金丝雀扩展 - 50% 流量
   ↓ 指标正常
阶段 6: 全量发布 (Production) - 100% 流量
   ↓ 观察期通过
阶段 7: 发布完成
```

### Google 的发布工具链

| 工具 | 用途 | 特点 |
|------|------|------|
| **Rapid** | 自动化发布系统 | 可扩展、可复现、可靠 |
| **Sisyphus** | 发布自动化框架 | 细粒度控制、进度监控 |
| **MPM** | 包管理器 | 标签管理 (dev/canary/production) |

### 发布检查清单

- [ ] 代码审查完成
- [ ] 自动化测试通过
- [ ] 安全扫描通过
- [ ] 配置变更已审查
- [ ] 回滚方案已准备
- [ ] 监控告警已配置
- [ ] 值班人员已通知
- [ ] 发布窗口已确认

## 金丝雀发布 (Canary Release)

### 金丝雀发布原理

金丝雀发布将新版本先部署到一小部分用户/实例，验证无误后再逐步扩大范围。

```
流量分配示意图：

初始状态：
┌─────────────────────────────────────┐
│  v1.0 (100%)                        │
└─────────────────────────────────────┘

金丝雀阶段 (5%)：
┌─────────────────────────────────────┐
│  v1.0 (95%) │ v1.1 (5%)             │
└─────────────────────────────────────┘

扩展阶段 (50%)：
┌─────────────────────────────────────┐
│  v1.0 (50%) │ v1.1 (50%)            │
└─────────────────────────────────────┘

全量发布 (100%)：
┌─────────────────────────────────────┐
│  v1.1 (100%)                        │
└─────────────────────────────────────┘
```

### 金丝雀发布实施

#### Kubernetes + Flagger 示例

```yaml
apiVersion: flagger.app/v1beta1
kind: Canary
metadata:
  name: my-app
  namespace: production
spec:
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: my-app
  service:
    port: 8080
  analysis:
    # 金丝雀分析间隔
    interval: 30s
    # 金丝雀持续时间
    threshold: 5
    # 最大权重
    maxWeight: 50
    # 步进权重
    stepWeight: 10
    # 成功指标
    metrics:
      - name: request-success-rate
        thresholdRange:
          min: 99
        interval: 1m
      - name: request-duration
        thresholdRange:
          max: 500
        interval: 1m
    # Webhook 检查
    webhooks:
      - name: load-test
        type: pre-rollout
        url: http://flagger-loadtester.test/
        timeout: 5s
        metadata:
          cmd: "hey -z 1m -q 10 -c 2 http://my-app-canary:8080/"
```

#### Argo Rollouts 示例

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Rollout
metadata:
  name: my-app
spec:
  replicas: 10
  strategy:
    canary:
      canaryService: my-app-canary
      stableService: my-app-stable
      trafficRouting:
        nginx:
          stableIngress: my-app-ingress
          annotationPrefix: nginx.ingress.kubernetes.io
      steps:
        # 阶段 1: 10% 流量，持续 10 分钟
        - setWeight: 10
        - pause: {duration: 10m}

        # 阶段 2: 25% 流量，等待手动批准
        - setWeight: 25
        - pause: {}

        # 阶段 3: 50% 流量，自动分析
        - setWeight: 50
        - analysis:
            templates:
              - templateName: success-rate
            args:
              - name: service-name
                value: my-app

        # 阶段 4: 100% 流量
        - setWeight: 100
```

### 金丝雀分析指标

```yaml
# 金丝雀分析模板
analysis_templates:
  success_rate:
    metrics:
      - name: success-rate
        interval: 5m
        successCondition: result[0] >= 0.99
        provider:
          prometheus:
            address: http://prometheus:9090
            query: |
              sum(rate(http_requests_total{service="my-app",status!~"5.."}[5m]))
              /
              sum(rate(http_requests_total{service="my-app"}[5m]))

  latency:
    metrics:
      - name: p99-latency
        interval: 5m
        successCondition: result[0] <= 0.5
        provider:
          prometheus:
            query: |
              histogram_quantile(0.99,
                sum(rate(http_request_duration_seconds_bucket{service="my-app"}[5m])) by (le)
              )
```

### 金丝雀最佳实践

1. **选择合适的金丝雀规模**
   - 初始：1-5% 流量或实例
   - 确保样本量足够检测问题

2. **定义清晰的通过/失败标准**
   - 错误率阈值
   - 延迟阈值
   - 业务指标阈值

3. **设置最大金丝雀持续时间**
   - 避免金丝雀状态长期存在
   - 设置超时自动回滚

4. **自动化决策**
   - 自动推进或回滚
   - 减少人为判断延迟

## 蓝绿部署 (Blue-Green Deployment)

### 蓝绿部署原理

蓝绿部署维护两个完全相同的生产环境，一个运行当前版本（蓝），一个部署新版本（绿），通过切换流量实现瞬间发布或回滚。

```
蓝绿部署架构：

部署前：
┌─────────────┐      ┌─────────────┐
│   Load      │─────▶│  Blue (v1)  │◀── 100% 流量
│  Balancer   │      │  Active     │
└─────────────┘      └─────────────┘
                     ┌─────────────┐
                     │  Green (v2) │◀── 空闲
                     │   Idle      │
                     └─────────────┘

切换后：
┌─────────────┐      ┌─────────────┐
│   Load      │─────▶│  Blue (v1)  │◀── 空闲
│  Balancer   │      │   Idle      │
└─────────────┘      └─────────────┘
                     ┌─────────────┐
                     │  Green (v2) │◀── 100% 流量
                     │  Active     │
                     └─────────────┘
```

### 蓝绿部署实施

#### Kubernetes Service 切换

```yaml
# blue-green-service.yaml
apiVersion: v1
kind: Service
metadata:
  name: my-app
  labels:
    app: my-app
spec:
  type: ClusterIP
  ports:
    - port: 80
      targetPort: 8080
  # 通过 selector 切换版本
  selector:
    app: my-app
    version: blue  # 切换到 green 完成发布
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app-blue
spec:
  replicas: 3
  selector:
    matchLabels:
      app: my-app
      version: blue
  template:
    metadata:
      labels:
        app: my-app
        version: blue
    spec:
      containers:
        - name: app
          image: my-app:v1.0
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app-green
spec:
  replicas: 3
  selector:
    matchLabels:
      app: my-app
      version: green
  template:
    metadata:
      labels:
        app: my-app
        version: green
    spec:
      containers:
        - name: app
          image: my-app:v2.0
```

#### 切换脚本

```bash
#!/bin/bash
# blue-green-switch.sh

NAMESPACE="production"
SERVICE="my-app"
CURRENT_VERSION=$(kubectl get svc $SERVICE -n $NAMESPACE -o jsonpath='{.spec.selector.version}')

if [ "$CURRENT_VERSION" == "blue" ]; then
    NEW_VERSION="green"
else
    NEW_VERSION="blue"
fi

echo "当前版本: $CURRENT_VERSION"
echo "切换至: $NEW_VERSION"

# 更新 Service selector
kubectl patch svc $SERVICE -n $NAMESPACE -p \
    '{"spec":{"selector":{"version":"'$NEW_VERSION'"}}}'

# 验证切换
sleep 5
NEW_ACTIVE=$(kubectl get svc $SERVICE -n $NAMESPACE -o jsonpath='{.spec.selector.version}')
echo "当前活跃版本: $NEW_ACTIVE"

# 健康检查
if ! curl -sf http://my-app/health; then
    echo "健康检查失败，回滚..."
    kubectl patch svc $SERVICE -n $NAMESPACE -p \
        '{"spec":{"selector":{"version":"'$CURRENT_VERSION'"}}}'
    exit 1
fi

echo "切换成功"
```

### 蓝绿部署 vs 金丝雀发布

| 维度 | 蓝绿部署 | 金丝雀发布 |
|------|----------|------------|
| 资源需求 | 2x 资源 | 1.1-1.5x 资源 |
| 回滚速度 | 瞬间 | 需要重新分配流量 |
| 风险暴露 | 全有或全无 | 渐进式 |
| 适用场景 | 关键系统、数据库迁移 | 常规功能发布 |
| 复杂度 | 较低 | 较高 |

## 功能开关 (Feature Flags)

### 功能开关类型

| 类型 | 用途 | 示例 |
|------|------|------|
| **发布开关** | 隐藏未完成功能 | 新功能开发中 |
| **实验开关** | A/B 测试 | 测试新算法效果 |
| **运维开关** | 运行时控制 | 熔断、限流 |
| **权限开关** | 功能权限控制 | 付费功能 |
| **Kill Switch** | 紧急禁用 | 故障时快速回滚 |

### 功能开关实施

#### 开源方案：Unleash

```python
# 客户端代码示例
from unleash import UnleashClient

client = UnleashClient(
    url="http://unleash:4242/api",
    app_name="my-app",
    custom_headers={'Authorization': '*:development'}
)
client.initialize_client()

# 检查功能开关
if client.is_enabled("new-checkout-flow", context={'userId': user_id}):
    # 新结账流程
    process_new_checkout()
else:
    # 旧结账流程
    process_old_checkout()
```

#### 配置示例

```yaml
# feature-flags.yaml
flags:
  new-checkout-flow:
    description: "新结账流程"
    enabled: true
    strategies:
      - name: gradualRollout
        parameters:
          percentage: 10
      - name: userWithId
        parameters:
          userIds: "user1,user2,user3"

  dark-mode:
    description: "深色模式"
    enabled: true
    strategies:
      - name: default

  emergency-kill-switch:
    description: "紧急熔断开关"
    enabled: false  # 默认关闭，紧急时启用
    strategies:
      - name: default
```

### 功能开关最佳实践

1. **开关生命周期管理**
   ```
   开发 → 测试 → 渐进发布 → 全量 → 移除开关
   ```

2. **避免技术债务**
   - 定期清理已全量的功能开关
   - 每个开关设定过期时间

3. **监控开关状态**
   - 跟踪开关使用情况
   - 监控开关性能影响

4. **安全控制**
   - 限制开关修改权限
   - 审计开关变更记录

## 回滚策略

### 回滚触发条件

| 触发条件 | 响应时间 | 回滚方式 |
|----------|----------|----------|
| 错误率超过阈值 | 自动，< 1 分钟 | 自动回滚 |
| 延迟显著增加 | 自动，< 5 分钟 | 自动回滚 |
| 业务指标异常 | 手动，< 15 分钟 | 人工决策 |
| 用户投诉激增 | 手动，< 30 分钟 | 人工决策 |

### 回滚策略类型

#### 1. 蓝绿部署回滚

```bash
# 瞬间回滚 - 切换流量
kubectl patch svc my-app -p '{"spec":{"selector":{"version":"blue"}}}'
```

#### 2. 金丝雀回滚

```bash
# 逐步回滚 - 降低金丝雀权重
kubectl argo rollouts abort my-app
```

#### 3. 功能开关回滚

```bash
# 禁用功能 - 无需重新部署
curl -X POST http://feature-flag-api/flags/new-feature/disable
```

#### 4. 配置回滚

```bash
# Kubernetes 配置回滚
kubectl rollout undo deployment/my-app

# 回滚到特定版本
kubectl rollout undo deployment/my-app --to-revision=3
```

### 回滚检查清单

- [ ] 确认回滚触发条件
- [ ] 通知相关团队
- [ ] 执行回滚操作
- [ ] 验证服务恢复
- [ ] 监控关键指标
- [ ] 保留故障现场
- [ ] 启动事故调查
- [ ] 更新发布文档

### 数据库回滚策略

数据库变更是最难回滚的部分，需要特殊处理：

```
数据库变更策略：

1. 扩展-收缩模式 (Expand-Contract)
   - 阶段 1: 添加新列/表（向后兼容）
   - 阶段 2: 双写新旧数据
   - 阶段 3: 迁移历史数据
   - 阶段 4: 切换到新 schema
   - 阶段 5: 删除旧列/表

2. 版本化 Schema
   - 每个 schema 变更都有版本号
   - 支持向前和向后迁移
   - 使用 Flyway/Liquibase 管理
```

## 实践检查清单

### 发布前检查

- [ ] 代码审查完成
- [ ] 单元测试通过率 100%
- [ ] 集成测试通过
- [ ] 安全扫描无高危漏洞
- [ ] 性能测试通过
- [ ] 配置变更已审查
- [ ] 回滚方案已验证
- [ ] 监控告警已配置
- [ ] 值班人员已通知
- [ ] 发布文档已更新

### 发布中检查

- [ ] 金丝雀指标正常
- [ ] 错误率在预期范围内
- [ ] 延迟无明显增加
- [ ] 业务指标正常
- [ ] 日志无异常

### 发布后检查

- [ ] 全量发布成功
- [ ] 监控观察期通过
- [ ] 用户反馈正常
- [ ] 性能基线对比完成
- [ ] 发布总结文档完成

## 真实案例参考

### 案例：金丝雀发布避免大规模故障

**背景**：某金融系统发布新版本

**问题**：
- 金丝雀阶段 (5% 流量) 发现错误率异常
- 新版本的某个依赖库存在内存泄漏

**响应**：
- 自动触发回滚
- 仅影响 5% 用户，且快速恢复
- 避免了全量发布可能导致的大规模故障

**教训**：
- 金丝雀发布是防止故障扩散的有效手段
- 自动化监控和回滚至关重要

### 案例：功能开关实现快速回滚

**背景**：电商大促期间发现新推荐算法有问题

**响应**：
- 通过功能开关立即禁用新算法
- 回退到旧算法，无需重新部署
- 整个过程 < 30 秒

**收益**：
- 避免了重新部署的 5-10 分钟延迟
- 减少了潜在的交易损失

## 常用命令

### 日志分析（发布监控）

```bash
# Linux - 监控部署日志
tail -f deployment.log | grep -E "(ERROR|WARN|deployed|failed)"

# PowerShell - 监控部署日志
Get-Content deployment.log -Tail 100 -Wait | Select-String "ERROR|WARN|deployed|failed"

# PowerShell - 统计部署成功/失败次数
$deployments = Get-Content deployment.log | Select-String "deployment"
$deployments | ForEach-Object {
    if ($_ -match "success") { "SUCCESS" } elseif ($_ -match "failed") { "FAILED" } else { "UNKNOWN" }
} | Group-Object | Select-Object Name, Count
```

### JSON数据处理（发布配置）

```bash
# Linux - 使用jq处理发布配置
cat release-config.json | jq '.canary.stages[]'

# PowerShell - 处理发布配置
$releaseConfig = Get-Content release-config.json | ConvertFrom-Json
$releaseConfig.canary.stages | ForEach-Object {
    [PSCustomObject]@{
        Stage = $_.name
        Weight = $_.weight
        Duration = $_.duration
    }
}

# PowerShell - 验证蓝绿部署配置
$blueGreenConfig = Get-Content blue-green-config.json | ConvertFrom-Json
$blueGreenConfig | Select-Object @{N="BlueVersion";E={$_.blue.version}},
                                 @{N="GreenVersion";E={$_.green.version}},
                                 @{N="CurrentActive";E={$_.active}}
```

### 文件压缩解压（制品管理）

```bash
# Linux - 压缩制品
tar -czf release-v1.2.3.tar.gz ./dist/

# PowerShell - 压缩制品
Compress-Archive -Path ./dist/* -DestinationPath release-v1.2.3.zip -Force

# PowerShell - 解压制品
Expand-Archive -Path release-v1.2.3.zip -DestinationPath ./extracted -Force

# PowerShell - 计算文件哈希（验证制品完整性）
Get-FileHash -Path release-v1.2.3.zip -Algorithm SHA256 | Select-Object Hash, Path
```

### 日期时间处理（发布时间管理）

```bash
# Linux - 生成带时间戳的发布版本
date +%Y%m%d-%H%M%S

# PowerShell - 生成带时间戳的发布版本
$releaseVersion = Get-Date -Format "yyyyMMdd-HHmmss"
$releaseTag = "v1.2.3-$releaseVersion"
Write-Output "Release Tag: $releaseTag"

# PowerShell - 计算金丝雀观察期
$canaryStart = Get-Date
$canaryDuration = New-TimeSpan -Minutes 30
$canaryEnd = $canaryStart + $canaryDuration
Write-Output "Canary Period: $($canaryStart.ToString('HH:mm:ss')) - $($canaryEnd.ToString('HH:mm:ss'))"
```

### 环境变量管理（发布配置）

```bash
# Linux - 设置发布环境变量
export RELEASE_VERSION="1.2.3"
export DEPLOY_ENV="production"

# PowerShell - 设置发布环境变量
$env:RELEASE_VERSION = "1.2.3"
$env:DEPLOY_ENV = "production"

# PowerShell - 持久化环境变量
[Environment]::SetEnvironmentVariable("RELEASE_VERSION", "1.2.3", "User")
[Environment]::SetEnvironmentVariable("DEPLOY_ENV", "production", "Machine")

# PowerShell - 读取发布配置
$releaseConfig = @{
    Version = $env:RELEASE_VERSION
    Environment = $env:DEPLOY_ENV
    Timestamp = Get-Date -Format "o"
}
$releaseConfig | ConvertTo-Json | Out-File release-metadata.json
```

## 输出规范

```
🚀 发布工程诊断报告

📦 发布基本信息
- 发布版本：[version]
- 发布时间：[timestamp]
- 发布环境：[environment]
- 发布策略：[canary/blue-green/rolling]
- 发布负责人：[owner]

📊 发布状态
- 当前阶段：[stage]
- 进度：[progress]%
- 状态：[in_progress/success/rolled_back]

🔍 健康检查指标
| 指标 | 基线 | 当前 | 偏差 | 状态 |
|------|------|------|------|------|
| 错误率 | [baseline]% | [current]% | [diff]% | [pass/fail] |
| 延迟 P99 | [baseline]ms | [current]ms | [diff]ms | [pass/fail] |
| 吞吐量 | [baseline]RPS | [current]RPS | [diff]RPS | [pass/fail] |
| CPU 使用率 | [baseline]% | [current]% | [diff]% | [pass/fail] |

📈 分阶段发布状态
| 阶段 | 流量比例 | 状态 | 开始时间 | 完成时间 |
|------|----------|------|----------|----------|
| 阶段 1 | 5% | [status] | [time] | [time] |
| 阶段 2 | 25% | [status] | [time] | [time] |
| 阶段 3 | 50% | [status] | [time] | [time] |
| 阶段 4 | 100% | [status] | [time] | [time] |

⚠️ 发布决策
- 继续发布：[yes/no/rollback]
- 决策依据：[reasoning]
- 下一步行动：[next_action]

💡 发布总结
- 发布结果：[success/partial_failure/failure]
- 发现问题：[issues]
- 改进建议：[recommendations]

🛡️ 回滚信息（如适用）
- 回滚原因：[reason]
- 回滚时间：[time]
- 回滚时长：[duration]
- 恢复状态：[status]
```

## 参考资源

- [Google SRE Book - 发布工程](https://sre.google/sre-book/release-engineering/)
- [Google SRE Workbook - 金丝雀发布](https://sre.google/workbook/canarying-releases/)
- [Flagger 文档](https://docs.flagger.app/)
- [Argo Rollouts 文档](https://argoproj.github.io/argo-rollouts/)
