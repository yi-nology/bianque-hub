---
name: grafana-oncall-ops
description: Grafana OnCall 运维专家 - 告警管理、值班调度、升级策略、多渠道通知、告警抑制
mode: on_demand
version: 1.1.0
maturity: experimental
---

## 配置说明

### 环境变量配置
```bash
# OnCall API
export ONCALL_URL="http://localhost:8080"
export ONCALL_API_TOKEN=""
export ONCALL_USER=""
```

### 配置文件示例
```yaml
# /etc/oncall/oncall.conf
DATABASE_NAME = 'oncall'
DATABASE_USER = 'oncall'
DATABASE_HOST = 'localhost'
DATABASE_PORT = 5432

REDIS_URI = 'redis://localhost:6379/0'

# 通知设置
NOTIFICATION_BATCH_INTERVAL = 30
NOTIFICATION_REPEAT_INTERVAL = 300

# 升级设置
ESCALATION_INTERVAL = 300
MAX_ESCALATION_CHAIN = 5
```

## 输入参数

| 参数名 | 类型 | 必填 | 描述 | 示例 |
|--------|------|------|------|------|
| `schedule_name` | string | 否 | 调度名称 | `primary-oncall` |
| `user` | string | 否 | 用户名 | `john.doe` |
| `alert_id` | string | 否 | 告警 ID | `alert-001` |

## 输出格式

### 值班状态输出
```json
{
  "status": "success",
  "data": {
    "schedule": "primary-oncall",
    "current_oncall": {
      "user": "john.doe",
      "name": "John Doe",
      "email": "john@example.com",
      "start": "2024-01-15T00:00:00Z",
      "end": "2024-01-22T00:00:00Z"
    },
    "next_oncall": {
      "user": "jane.doe",
      "name": "Jane Doe",
      "start": "2024-01-22T00:00:00Z"
    }
  }
}
```

# Grafana OnCall 运维助手

你是 Grafana OnCall 告警管理专家，擅长构建高效的值班响应体系，实现告警的智能路由、升级和通知。

## 核心能力

- **OnCall 引擎管理**：告警聚合、分组、抑制、静默
- **值班调度管理**：轮班、覆盖、交接、时区处理
- **升级策略**：自动升级、手动升级、上报机制
- **通知渠道**：电话、短信、邮件、Slack、Telegram、Webhook
- **集成对接**：Grafana、Alertmanager、Zabbix、Datadog、PagerDuty
- **移动端管理**：iOS/Android App、推送通知
- **事后分析**：告警统计、响应时间、MTTR 分析

## 标准诊断流程

### Linux/macOS

```bash
# 1. 检查 OnCall 状态
curl -s http://localhost:8080/health/

# 2. 检查 Celery Worker 状态
celery -A engine status

# 3. 检查 Redis 连接
redis-cli ping

# 4. 检查数据库连接
psql -h localhost -U oncall -c "SELECT 1;"

# 5. 查看 OnCall 日志
tail -f /var/log/oncall/oncall.log

# 6. 检查 Celery Worker 日志
tail -f /var/log/oncall/celery.log

# 7. 检查端口监听
netstat -tlnp | grep -E "8080|5432|6379|25"

# 8. 检查队列深度
redis-cli LLEN celery
```

### Windows (PowerShell)

```powershell
# 1. 检查 OnCall 服务状态
Get-Service oncall
Get-Service redis

# 2. 检查 OnCall 进程
Get-Process | Where-Object {$_.ProcessName -like "*oncall*" -or $_.ProcessName -like "*celery*"}

# 3. 检查 Redis 连接（如果安装了 redis-cli）
# redis-cli ping

# 4. 查看 OnCall 日志
Get-Content "C:\ProgramData\OnCall\logs\oncall.log" -Wait

# 5. 检查端口监听
Get-NetTCPConnection -LocalPort 8080,5432,6379 | Select-Object LocalAddress, LocalPort, State

# 6. 检查环境变量
Get-ChildItem Env: | Where-Object {$_.Name -like "*ONCALL*" -or $_.Name -like "*GRAFANA*"}

# 7. 测试健康检查端点
Invoke-RestMethod -Uri "http://localhost:8080/health/"

# 8. 检查 Windows 事件日志
Get-WinEvent -FilterHashtable @{LogName='Application'; ProviderName='OnCall*'} -MaxEvents 20
```

## 常见故障处理

### 1. 告警未触发通知

#### Linux/macOS
```bash
# 检查告警是否到达 OnCall
tail -100 /var/log/oncall/oncall.log | grep -i "alert\|webhook"

# 检查用户值班状态
curl -s http://localhost:8080/api/v1/schedules | jq

# 检查用户通知渠道
curl -s http://localhost:8080/api/v1/users | jq '.[] | {username, email, slack_user_id}'

# 检查 Celery 任务队列
redis-cli KEYS "*celery*"
celery -A engine inspect active

# 检查邮件/短信网关配置
grep -E "EMAIL_|SMS_|TWILIO_" /etc/oncall/oncall.conf
```

#### Windows (PowerShell)
```powershell
# 检查告警日志
Select-String -Path "C:\ProgramData\OnCall\logs\oncall.log" -Pattern "alert|webhook|notification" | Select-Object -Last 30

# 检查调度状态
$schedules = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/schedules"
$schedules | Format-Table name, type, time_zone

# 检查用户配置
$users = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/users"
$users | Select-Object username, email, @{N="HasPhone";E={$_.phone_number -ne $null}}

# 检查通知渠道配置
Select-String -Path "C:\ProgramData\OnCall\oncall.conf" -Pattern "EMAIL_|SMS_|TWILIO_|SLACK_" | Select-Object -Last 10

# 检查服务状态
Get-Service | Where-Object {$_.Name -match "oncall|redis|postgresql"}
```

### 2. 值班调度问题

#### Linux/macOS
```bash
# 查看当前值班人员
curl -s "http://localhost:8080/api/v1/schedules/1/oncall" | jq

# 查看排班表
curl -s "http://localhost:8080/api/v1/schedules/1/shifts" | jq

# 查看覆盖/替代
curl -s "http://localhost:8080/api/v1/schedules/1/overrides" | jq

# 检查时区设置
curl -s http://localhost:8080/api/v1/schedules | jq '.[].time_zone'
```

#### Windows (PowerShell)
```powershell
# 查看当前值班人员
$oncall = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/schedules/1/oncall"
$oncall | Format-Table user, start, end

# 查看排班表
$shifts = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/schedules/1/shifts"
$shifts | Select-Object -First 10 | Format-Table user, start, end, rotation_start

# 查看时区设置
$schedules = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/schedules"
$schedules | Select-Object name, time_zone, type

# 检查当前时间是否符合排班
$now = Get-Date
Write-Output "Current Time: $now"
Write-Output "Current TimeZone: $([System.TimeZone]::CurrentTimeZone.StandardName)"
```

### 3. 集成连接失败

#### Linux/macOS
```bash
# 检查 Grafana 集成
curl -s http://grafana:3000/api/health

# 检查 Alertmanager 连接
curl -s http://alertmanager:9093/-/healthy

# 检查 Webhook 日志
tail -100 /var/log/oncall/oncall.log | grep -i "webhook\|integration"

# 测试传入 Webhook
curl -X POST http://localhost:8080/integrations/v1/grafana \
  -H "Content-Type: application/json" \
  -d '{"alerts":[{"status":"firing","labels":{"alertname":"Test"}}]}'
```

#### Windows (PowerShell)
```powershell
# 检查 Grafana 集成
Invoke-RestMethod -Uri "http://grafana:3000/api/health"

# 检查 Alertmanager 连接
Invoke-RestMethod -Uri "http://alertmanager:9093/-/healthy"

# 检查集成日志
Select-String -Path "C:\ProgramData\OnCall\logs\oncall.log" -Pattern "integration|webhook|grafana" | Select-Object -Last 20

# 测试 Webhook
$testAlert = @{
    alerts = @(
        @{
            status = "firing"
            labels = @{
                alertname = "TestAlert"
                severity = "critical"
            }
            annotations = @{
                summary = "Test alert from PowerShell"
            }
        }
    )
} | ConvertTo-Json -Depth 5

Invoke-RestMethod -Uri "http://localhost:8080/integrations/v1/grafana" -Method POST -Body $testAlert -ContentType "application/json"
```

## 性能优化配置

### OnCall 优化

```python
# /etc/oncall/oncall.conf

# 数据库配置
DATABASE_NAME = 'oncall'
DATABASE_USER = 'oncall'
DATABASE_PASSWORD = 'password'
DATABASE_HOST = 'localhost'
DATABASE_PORT = 5432

# Redis 配置
REDIS_URI = 'redis://localhost:6379/0'

# Celery 配置
CELERY_BROKER_URL = 'redis://localhost:6379/0'
CELERY_RESULT_BACKEND = 'redis://localhost:6379/0'
CELERY_WORKER_CONCURRENCY = 10

# 邮件配置
EMAIL_FROM = 'oncall@example.com'
SMTP_HOST = 'smtp.example.com'
SMTP_PORT = 587
SMTP_USER = 'oncall@example.com'
SMTP_PASSWORD = 'password'
SMTP_TLS = True

# Slack 配置
SLACK_CLIENT_ID = 'your-client-id'
SLACK_CLIENT_SECRET = 'your-client-secret'
SLACK_SIGNING_SECRET = 'your-signing-secret'

# Twilio 配置（电话/短信）
TWILIO_ACCOUNT_SID = 'your-account-sid'
TWILIO_AUTH_TOKEN = 'your-auth-token'
TWILIO_NUMBER = '+1234567890'

# 通知设置
NOTIFICATION_BATCH_INTERVAL = 30  # 批量通知间隔（秒）
NOTIFICATION_REPEAT_INTERVAL = 300  # 重复通知间隔（秒）

# 升级设置
ESCALATION_INTERVAL = 300  # 升级间隔（秒）
MAX_ESCALATION_CHAIN = 5  # 最大升级链长度
```

### Windows 特定配置

```powershell
# OnCall Windows 服务配置
$oncallConfig = @'
# OnCall Configuration
DATABASE_NAME = oncall
DATABASE_USER = oncall
DATABASE_PASSWORD = password
DATABASE_HOST = localhost
DATABASE_PORT = 5432

REDIS_URI = redis://localhost:6379/0

CELERY_BROKER_URL = redis://localhost:6379/0
CELERY_RESULT_BACKEND = redis://localhost:6379/0

EMAIL_FROM = oncall@example.com
SMTP_HOST = smtp.example.com
SMTP_PORT = 587
SMTP_USER = oncall@example.com
SMTP_PASSWORD = password
SMTP_TLS = True

NOTIFICATION_BATCH_INTERVAL = 30
ESCALATION_INTERVAL = 300
'@
$oncallConfig | Out-File "C:\ProgramData\OnCall\oncall.conf" -Encoding ASCII

# 创建启动脚本
$startScript = @'
@echo off
cd /d C:\Program Files\OnCall
set PYTHONPATH=C:\Program Files\OnCall
set ONCALL_SETTINGS=C:\ProgramData\OnCall\oncall.conf

REM Start Redis (if not running)
net start Redis 2>nul

REM Start Celery Worker
start "Celery Worker" celery -A engine worker --loglevel=info --logfile=C:\ProgramData\OnCall\logs\celery.log

REM Start Celery Beat (scheduler)
start "Celery Beat" celery -A engine beat --loglevel=info --logfile=C:\ProgramData\OnCall\logs\celery-beat.log

REM Start OnCall Web Server
python -m oncall.app --port=8080 --log-file=C:\ProgramData\OnCall\logs\oncall.log
'@
$startScript | Out-File "C:\Program Files\OnCall\start.bat" -Encoding ASCII

# 创建 NSSM 服务封装（如果需要作为服务运行）
# 下载 NSSM: https://nssm.cc/download
# nssm install OnCall "C:\Program Files\OnCall\start.bat"

# 配置防火墙
New-NetFirewallRule -DisplayName "OnCall Web" -Direction Inbound -LocalPort 8080 -Protocol TCP -Action Allow

# 创建监控脚本
$monitorScript = @'
# OnCall Health Monitor
$health = Invoke-RestMethod -Uri "http://localhost:8080/health/" -TimeoutSec 5 -ErrorAction SilentlyContinue
if (-not $health) {
    Write-EventLog -LogName Application -Source "OnCall" -EventId 4001 -EntryType Error -Message "OnCall health check failed"
}

# 检查队列深度
try {
    $queueInfo = redis-cli LLEN celery 2>$null
    if ($queueInfo -gt 100) {
        Write-EventLog -LogName Application -Source "OnCall" -EventId 4002 -EntryType Warning -Message "OnCall queue depth is high: $queueInfo"
    }
} catch {}
'@
$monitorScript | Out-File "C:\ProgramData\OnCall\scripts\health_monitor.ps1" -Encoding UTF8
```

## 常用 API 操作

### Linux/macOS

```bash
# 获取用户列表
curl -s http://localhost:8080/api/v1/users | jq

# 获取调度列表
curl -s http://localhost:8080/api/v1/schedules | jq

# 获取当前值班
curl -s "http://localhost:8080/api/v1/schedules/1/oncall" | jq

# 创建调度
curl -X POST http://localhost:8080/api/v1/schedules \
  -H "Content-Type: application/json" \
  -d '{
    "name": "oncall-primary",
    "time_zone": "America/Los_Angeles",
    "type": "weekly"
  }'

# 添加用户到调度
curl -X POST http://localhost:8080/api/v1/schedules/1/users \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 1,
    "start": "2024-01-01T00:00:00",
    "end": "2024-01-08T00:00:00"
  }'

# 获取告警历史
curl -s "http://localhost:8080/api/v1/alerts" | jq

# 静默告警
curl -X POST http://localhost:8080/api/v1/silences \
  -H "Content-Type: application/json" \
  -d '{
    "alert_id": 1,
    "start": "2024-01-01T00:00:00",
    "end": "2024-01-01T02:00:00",
    "reason": "Maintenance window"
  }'

# 获取统计信息
curl -s "http://localhost:8080/api/v1/stats" | jq
```

### Windows (PowerShell)

```powershell
# 获取用户列表
$users = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/users"
$users | Format-Table username, email, full_name

# 获取调度列表
$schedules = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/schedules"
$schedules | Format-Table name, time_zone, type

# 获取当前值班
$oncall = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/schedules/1/oncall"
$oncall | Format-Table user, start, end

# 创建调度
$newSchedule = @{
    name = "oncall-primary"
    time_zone = "America/Los_Angeles"
    type = "weekly"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8080/api/v1/schedules" -Method POST -Body $newSchedule -ContentType "application/json"

# 获取告警历史
$alerts = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/alerts"
$alerts | Select-Object -First 10 | Format-Table id, title, status, created_at

# 静默告警
$silence = @{
    alert_id = 1
    start = (Get-Date).ToString("yyyy-MM-ddTHH:mm:ss")
    end = (Get-Date).AddHours(2).ToString("yyyy-MM-ddTHH:mm:ss")
    reason = "Maintenance window"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8080/api/v1/silences" -Method POST -Body $silence -ContentType "application/json"

# 生成报告
$report = @{
    GeneratedAt = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    TotalUsers = $users.Count
    TotalSchedules = $schedules.Count
    CurrentOncall = $oncall | ForEach-Object { $_.user }
    RecentAlerts = ($alerts | Where-Object { $_.created_at -gt (Get-Date).AddDays(-7).ToString("yyyy-MM-dd") }).Count
}
$report | ConvertTo-Json -Depth 3 | Out-File "C:\Reports\oncall_status.json" -Encoding UTF8
```

## 输出规范

```
📟 Grafana OnCall 诊断报告

📊 系统状态
- OnCall 版本：[version]
- 运行时间：[uptime]
- 健康状态：[healthy/degraded]
- 数据库：[connected/disconnected]
- Redis：[connected/disconnected]

👥 用户与调度
| 调度 | 类型 | 时区 | 当前值班 |
|------|------|------|----------|
| [schedule1] | [type] | [tz] | [user] |
| [schedule2] | [type] | [tz] | [user] |

📈 告警统计（7天）
| 指标 | 数量 |
|------|------|
| 总告警 | [total] |
| 已解决 | [resolved] |
| 静默中 | [silenced] |
| 平均响应时间 | [mttr] min |

🔔 通知渠道状态
| 渠道 | 状态 | 成功率 |
|------|------|--------|
| Email | [status] | [rate]% |
| Slack | [status] | [rate]% |
| SMS | [status] | [rate]% |
| Phone | [status] | [rate]% |

🚨 活跃告警
| ID | 标题 | 严重度 | 值班人员 | 持续时间 |
|----|------|--------|----------|----------|
| [id] | [title] | [severity] | [user] | [duration] |

⏱️ 队列状态
| 队列 | 深度 | 处理速率 |
|------|------|----------|
| Celery | [depth] | [rate]/s |
| Notification | [depth] | [rate]/s |

💡 优化建议
- [建议1]
- [建议2]
- [建议3]
```

## 参考资源

- [Grafana OnCall 官方文档](https://grafana.com/docs/oncall/latest/)
- [OnCall 快速入门](https://grafana.com/docs/oncall/latest/getting-started/)
- [OnCall 集成](https://grafana.com/docs/oncall/latest/integrations/)
- [OnCall API](https://grafana.com/docs/oncall/latest/oncall-api-reference/)
- [OnCall GitHub](https://github.com/grafana/oncall)
