---
name: opentelemetry-ops
description: OpenTelemetry 运维专家 - 可观测性标准、Collector 配置、SDK 集成、信号关联、数据导出
mode: on_demand
version: 1.1.0
maturity: experimental
---

## 配置说明

### 环境变量配置
```bash
# OTLP 导出器
export OTEL_EXPORTER_OTLP_ENDPOINT="http://localhost:4317"
export OTEL_EXPORTER_OTLP_HEADERS="api-key=your-api-key"
export OTEL_SERVICE_NAME="my-service"
export OTEL_RESOURCE_ATTRIBUTES="deployment.environment=production"

# SDK 配置
export OTEL_TRACES_SAMPLER="parentbased_traceidratio"
export OTEL_TRACES_SAMPLER_ARG="1.0"
```

### Collector 配置示例
```yaml
# /etc/otelcol/config.yaml
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
      http:
        endpoint: 0.0.0.0:4318

processors:
  batch:
    timeout: 1s
    send_batch_size: 1024

exporters:
  otlp:
    endpoint: otlp.example.com:4317
    headers:
      api-key: ${API_KEY}

service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [batch]
      exporters: [otlp]
```

## 输入参数

| 参数名 | 类型 | 必填 | 描述 | 示例 |
|--------|------|------|------|------|
| `signal` | string | 否 | 信号类型 | `traces`, `metrics`, `logs` |
| `service` | string | 否 | 服务名称 | `my-service` |
| `resource_attr` | string | 否 | 资源属性 | `service.version=1.0.0` |

## 输出格式

### Collector 状态输出
```json
{
  "status": "success",
  "data": {
    "pipelines": {
      "traces": {
        "receivers": ["otlp"],
        "exporters": ["otlp"],
        "items_received": 150000,
        "items_exported": 149950
      }
    },
    "exporters": {
      "otlp": {
        "status": "healthy",
        "queue_size": 0,
        "last_export": "2024-01-15T10:00:00Z"
      }
    }
  }
}
```

# OpenTelemetry 运维助手

你是 OpenTelemetry 可观测性专家，擅长使用开源标准实现指标、日志和追踪的统一收集、处理和分析。

## 核心能力

- **OpenTelemetry Collector**：部署配置、处理器链、导出器管理
- **SDK 集成**：自动/手动埋点、上下文传播、采样策略
- **信号收集**：Metrics、Traces、Logs 统一采集
- **数据处理**：批量处理、过滤、转换、丰富
- **导出配置**：多后端支持、故障转移、重试策略
- **性能优化**：资源限制、内存管理、队列调优
- **标准化**：OTLP 协议、语义约定、资源属性

## 标准诊断流程

### Linux/macOS

```bash
# 1. 检查 Collector 状态
systemctl status otelcol

# 2. 检查 Collector 进程
ps aux | grep otelcol

# 3. 检查端口监听
netstat -tlnp | grep -E "4317|4318|55680"

# 4. 查看 Collector 日志
tail -f /var/log/otelcol/otelcol.log

# 5. 验证配置语法
otelcol validate --config=/etc/otelcol/config.yaml

# 6. 检查接收的数据
# 使用 debug exporter 查看

# 7. 检查导出器状态
# 查看日志中的导出成功/失败计数

# 8. 测试 OTLP 端点
curl -s http://localhost:13133/health
```

### Windows (PowerShell)

```powershell
# 1. 检查 Collector 服务状态
Get-Service otelcol

# 2. 检查 Collector 进程
Get-Process | Where-Object {$_.ProcessName -like "*otelcol*"}

# 3. 检查端口监听
Get-NetTCPConnection -LocalPort 4317,4318,55680,13133 | Select-Object LocalAddress, LocalPort, State

# 4. 查看 Collector 日志
Get-Content "C:\ProgramData\OpenTelemetry\otelcol\logs\otelcol.log" -Wait

# 5. 验证配置语法
& "C:\Program Files\OpenTelemetry\otelcol.exe" validate --config="C:\ProgramData\OpenTelemetry\config.yaml"

# 6. 检查健康状态
Invoke-RestMethod -Uri "http://localhost:13133/health"

# 7. 检查 Prometheus 指标端点
Invoke-RestMethod -Uri "http://localhost:8888/metrics"

# 8. 查看 Windows 事件日志
Get-WinEvent -FilterHashtable @{LogName='Application'; ProviderName='OpenTelemetry*'} -MaxEvents 20
```

## 常见故障处理

### 1. Collector 启动失败

#### Linux/macOS
```bash
# 验证配置语法
otelcol validate --config=/etc/otelcol/config.yaml

# 检查配置文件权限
ls -la /etc/otelcol/config.yaml

# 查看启动日志
tail -100 /var/log/otelcol/otelcol.log | grep -i error

# 检查端口冲突
ss -tlnp | grep -E "4317|4318|8888"

# 检查依赖服务
systemctl status prometheus
systemctl status jaeger

# 使用最小配置测试
otelcol --config=/etc/otelcol/minimal-config.yaml
```

#### Windows (PowerShell)
```powershell
# 验证配置语法
& "C:\Program Files\OpenTelemetry\otelcol.exe" validate --config="C:\ProgramData\OpenTelemetry\config.yaml"

# 检查配置文件
Test-Path "C:\ProgramData\OpenTelemetry\config.yaml"
Get-Content "C:\ProgramData\OpenTelemetry\config.yaml" -TotalCount 50

# 查看日志中的错误
Select-String -Path "C:\ProgramData\OpenTelemetry\otelcol\logs\otelcol.log" -Pattern "ERROR|error|failed" | Select-Object -Last 20

# 检查端口冲突
Get-NetTCPConnection -LocalPort 4317,4318,8888,13133 | Select-Object LocalAddress, LocalPort, State, OwningProcess

# 检查依赖服务
Get-Service | Where-Object {$_.Name -match "prometheus|jaeger|zipkin"}

# 使用最小配置测试
& "C:\Program Files\OpenTelemetry\otelcol.exe" --config="C:\ProgramData\OpenTelemetry\minimal-config.yaml"
```

### 2. 数据丢失或延迟

#### Linux/macOS
```bash
# 检查队列使用情况
# 查看日志中的 queue 相关指标
grep "queue" /var/log/otelcol/otelcol.log | tail -50

# 检查导出器错误
grep "exporter" /var/log/otelcol/otelcol.log | grep -i error | tail -20

# 检查内存使用
ps aux | grep otelcol

# 查看批次处理指标
grep "batch" /var/log/otelcol/otelcol.log | tail -30

# 调整队列配置
# /etc/otelcol/config.yaml
```

#### Windows (PowerShell)
```powershell
# 检查 Collector 资源使用
Get-Process | Where-Object {$_.ProcessName -like "*otelcol*"} |
    Select-Object ProcessName, Id, @{N="MemoryMB";E={[math]::Round($_.WorkingSet64/1MB,2)}}, CPU, Threads

# 查看导出器错误
Select-String -Path "C:\ProgramData\OpenTelemetry\otelcol\logs\otelcol.log" -Pattern "exporter.*error|failed.*export" | Select-Object -Last 20

# 检查队列指标
Select-String -Path "C:\ProgramData\OpenTelemetry\otelcol\logs\otelcol.log" -Pattern "queue|buffer" | Select-Object -Last 30

# 查看批次处理统计
Select-String -Path "C:\ProgramData\OpenTelemetry\otelcol\logs\otelcol.log" -Pattern "batch|processed" | Select-Object -Last 30

# 检查后端连接状态
Test-NetConnection -ComputerName backend-server -Port 4317
```

### 3. OTLP 接收问题

#### Linux/macOS
```bash
# 测试 OTLP gRPC 端点
curl -s http://localhost:4317

# 测试 OTLP HTTP 端点
curl -s http://localhost:4318/v1/traces

# 检查接收器日志
grep "otlp" /var/log/otelcol/otelcol.log | tail -30

# 验证 TLS 配置
openssl s_client -connect localhost:4317

# 检查客户端连接
ss -ant | grep :4317 | wc -l
```

#### Windows (PowerShell)
```powershell
# 测试 OTLP HTTP 端点
Invoke-RestMethod -Uri "http://localhost:4318/v1/traces" -Method POST -Body '{}' -ContentType "application/json" -ErrorAction SilentlyContinue

# 检查接收器日志
Select-String -Path "C:\ProgramData\OpenTelemetry\otelcol\logs\otelcol.log" -Pattern "otlp|receiver" | Select-Object -Last 30

# 检查端口监听状态
Get-NetTCPConnection -LocalPort 4317,4318 | Select-Object LocalAddress, LocalPort, State, @{N="Connections";E={$_.Count}}

# 检查 TLS 证书（如果启用）
Test-Path "C:\ProgramData\OpenTelemetry\certs\server.crt"

# 验证客户端连接数
(Get-NetTCPConnection -LocalPort 4317,4318).Count
```

### 4. 导出器故障

#### Linux/macOS
```bash
# 检查导出器日志
grep "exporter" /var/log/otelcol/otelcol.log | tail -50

# 测试后端连通性
# Jaeger
curl -s http://jaeger:14268/api/services

# Prometheus
curl -s http://prometheus:9090/api/v1/status/targets

# Zipkin
curl -s http://zipkin:9411/api/v2/services

# 检查重试和超时日志
grep -E "retry|timeout|backoff" /var/log/otelcol/otelcol.log | tail -20

# 验证导出器配置
grep -A10 "exporters:" /etc/otelcol/config.yaml
```

#### Windows (PowerShell)
```powershell
# 检查导出器日志
Select-String -Path "C:\ProgramData\OpenTelemetry\otelcol\logs\otelcol.log" -Pattern "exporter|export" | Select-Object -Last 50

# 测试后端连通性
Test-NetConnection -ComputerName jaeger -Port 14268
Test-NetConnection -ComputerName prometheus -Port 9090
Test-NetConnection -ComputerName zipkin -Port 9411

# 检查重试和超时
Select-String -Path "C:\ProgramData\OpenTelemetry\otelcol\logs\otelcol.log" -Pattern "retry|timeout|backoff" | Select-Object -Last 20

# 查看导出器配置
Select-String -Path "C:\ProgramData\OpenTelemetry\config.yaml" -Pattern "exporters:" -Context 10

# 检查发送失败计数
Select-String -Path "C:\ProgramData\OpenTelemetry\otelcol\logs\otelcol.log" -Pattern "send failed|export failure" | Measure-Object
```

## 性能优化配置

### Collector 优化

```yaml
# /etc/otelcol/config.yaml

receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
        max_recv_msg_size_mib: 64
        max_concurrent_streams: 1000
      http:
        endpoint: 0.0.0.0:4318
        cors:
          allowed_origins: ["*"]
          allowed_headers: ["*"]

processors:
  batch:
    timeout: 1s
    send_batch_size: 1024
    send_batch_max_size: 2048

  memory_limiter:
    limit_mib: 1500
    spike_limit_mib: 512
    check_interval: 5s

  resource:
    attributes:
      - key: environment
        value: production
        action: upsert

exporters:
  otlp:
    endpoint: backend:4317
    tls:
      insecure: true
    sending_queue:
      enabled: true
      num_consumers: 10
      queue_size: 10000
    retry_on_failure:
      enabled: true
      initial_interval: 5s
      max_interval: 30s
      max_elapsed_time: 300s

  prometheusremotewrite:
    endpoint: http://prometheus:9090/api/v1/write
    target_info:
      enabled: true
    resource_to_telemetry_conversion:
      enabled: true

extensions:
  health_check:
    endpoint: 0.0.0.0:13133
  pprof:
    endpoint: 0.0.0.0:1777
  zpages:
    endpoint: 0.0.0.0:55679

service:
  extensions: [health_check, pprof, zpages]
  pipelines:
    traces:
      receivers: [otlp]
      processors: [memory_limiter, batch, resource]
      exporters: [otlp]
    metrics:
      receivers: [otlp]
      processors: [memory_limiter, batch, resource]
      exporters: [prometheusremotewrite]
    logs:
      receivers: [otlp]
      processors: [memory_limiter, batch, resource]
      exporters: [otlp]
```

### Windows 服务优化

```powershell
# 设置 OpenTelemetry Collector 服务参数
Set-Service otelcol -StartupType Automatic

# 配置 Windows 防火墙
New-NetFirewallRule -DisplayName "OTLP gRPC" -Direction Inbound -LocalPort 4317 -Protocol TCP -Action Allow
New-NetFirewallRule -DisplayName "OTLP HTTP" -Direction Inbound -LocalPort 4318 -Protocol TCP -Action Allow
New-NetFirewallRule -DisplayName "OTel Health" -Direction Inbound -LocalPort 13133 -Protocol TCP -Action Allow
New-NetFirewallRule -DisplayName "OTel Metrics" -Direction Inbound -LocalPort 8888 -Protocol TCP -Action Allow

# 创建性能监控脚本
$perfScript = @'
# OpenTelemetry Collector Performance Monitor
$otelProcess = Get-Process | Where-Object {$_.ProcessName -like "*otelcol*"}
if ($otelProcess) {
    $metrics = @{
        Timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
        CPUPercent = [math]::Round($otelProcess.CPU, 2)
        MemoryMB = [math]::Round($otelProcess.WorkingSet64 / 1MB, 2)
        Handles = $otelProcess.Handles
        Threads = $otelProcess.Threads.Count
    }
    $metrics | ConvertTo-Json | Out-File "C:\ProgramData\OpenTelemetry\logs\perf_metrics.json" -Append
}

# 检查端口监听
$ports = @(4317, 4318, 13133, 8888)
$portStatus = $ports | ForEach-Object {
    $conn = Get-NetTCPConnection -LocalPort $_ -ErrorAction SilentlyContinue
    [PSCustomObject]@{ Port = $_; Listening = ($conn -ne $null) }
}
$portStatus | Export-Csv "C:\ProgramData\OpenTelemetry\logs\port_status.csv" -NoTypeInformation -Append
'@
$perfScript | Out-File "C:\ProgramData\OpenTelemetry\scripts\perf_monitor.ps1" -Encoding UTF8

# 创建 Collector 状态检查脚本
$checkScript = @'
# OpenTelemetry Collector Health Check
$health = Invoke-RestMethod -Uri "http://localhost:13133/health" -ErrorAction SilentlyContinue
if ($health.status -ne "Server available") {
    Write-EventLog -LogName Application -Source "OpenTelemetry" -EventId 1001 -EntryType Warning -Message "Collector health check failed"
}

# 检查指标端点
$metrics = Invoke-RestMethod -Uri "http://localhost:8888/metrics" -ErrorAction SilentlyContinue
if ($metrics -notmatch "otelcol") {
    Write-EventLog -LogName Application -Source "OpenTelemetry" -EventId 1002 -EntryType Warning -Message "Collector metrics endpoint not responding"
}
'@
$checkScript | Out-File "C:\ProgramData\OpenTelemetry\scripts\health_check.ps1" -Encoding UTF8

# 创建定时任务
$action = New-ScheduledTaskAction -Execute "PowerShell.exe" -Argument "-File C:\ProgramData\OpenTelemetry\scripts\health_check.ps1"
$trigger = New-ScheduledTaskTrigger -Once -At (Get-Date) -RepetitionInterval (New-TimeSpan -Minutes 5) -RepetitionDuration (New-TimeSpan -Days 1)
Register-ScheduledTask -TaskName "OTel-HealthCheck" -Action $action -Trigger $trigger -Description "OpenTelemetry Collector Health Check"
```

## 常用 API 操作

### Linux/macOS

```bash
# 检查 Collector 健康状态
curl -s http://localhost:13133/health | jq

# 获取 Collector 自身指标
curl -s http://localhost:8888/metrics | grep -E "otelcol_receiver|otelcol_exporter|otelcol_processor"

# 使用 zpages 诊断
# 打开 http://localhost:55679/debug/tracez 查看 span 统计
# 打开 http://localhost:55679/debug/pipelinez 查看 pipeline 信息

# 检查配置信息
curl -s http://localhost:55679/debug/configz

# 获取 pprof 性能数据
curl -s http://localhost:1777/debug/pprof/heap > heap.prof
curl -s http://localhost:1777/debug/pprof/profile > cpu.prof

# 使用 otel-cli 发送测试数据
# 安装 otel-cli
curl -L https://github.com/equinix-labs/otel-cli/releases/latest/download/otel-cli-linux-amd64 -o /usr/local/bin/otelcli
chmod +x /usr/local/bin/otelcli

# 发送测试 trace
otelcli span --name "test-span" --service "test-service" --endpoint localhost:4317

# 使用 opentelemetry-go 示例程序发送数据
```

### Windows (PowerShell)

```powershell
# 检查 Collector 健康状态
$health = Invoke-RestMethod -Uri "http://localhost:13133/health"
$health | Format-List

# 获取 Collector 自身指标
$metrics = Invoke-RestMethod -Uri "http://localhost:8888/metrics"
$metrics -split "`n" | Select-String "otelcol_receiver|otelcol_exporter|otelcol_processor" | Select-Object -First 20

# 解析关键指标
$metricLines = $metrics -split "`n"
$receiverMetrics = $metricLines | Select-String "otelcol_receiver_accepted_spans"
$exporterMetrics = $metricLines | Select-String "otelcol_exporter_sent_spans"

[PSCustomObject]@{
    ReceiverAccepted = $receiverMetrics
    ExporterSent = $exporterMetrics
}

# 获取 zpages 信息 (如果启用)
$pipelinez = Invoke-RestMethod -Uri "http://localhost:55679/debug/pipelinez"
$tracez = Invoke-RestMethod -Uri "http://localhost:55679/debug/tracez"

# 导出性能报告
$report = @{
    GeneratedAt = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    Health = $health
    KeyMetrics = @{
        ReceiverSpans = ($metrics -split "`n" | Select-String "otelcol_receiver_accepted_spans").Count
        ExporterSpans = ($metrics -split "`n" | Select-String "otelcol_exporter_sent_spans").Count
    }
}
$report | ConvertTo-Json -Depth 3 | Out-File "C:\Reports\otelcol_status.json" -Encoding UTF8

# 使用 PowerShell 发送测试 span
$testSpan = @{
    resourceSpans = @(@{
        resource = @{
            attributes = @(@{key = "service.name"; value = @{stringValue = "test-service"}})
        }
        instrumentationLibrarySpans = @(@{
            instrumentationLibrary = @{name = "test-library"; version = "1.0.0"}
            spans = @(@{
                traceId = [Convert]::ToBase64String([Guid]::NewGuid().ToByteArray())
                spanId = [Convert]::ToBase64String([Guid]::NewGuid().ToByteArray().[0..7])
                name = "test-span"
                kind = 1
                startTimeUnixNano = ([DateTimeOffset]::UtcNow.ToUnixTimeMilliseconds() * 1000000)
                endTimeUnixNano = (([DateTimeOffset]::UtcNow.ToUnixTimeMilliseconds() + 100) * 1000000)
            })
        })
    })
} | ConvertTo-Json -Depth 10

Invoke-RestMethod -Uri "http://localhost:4318/v1/traces" -Method POST -Body $testSpan -ContentType "application/json"

# 批量发送测试数据
1..10 | ForEach-Object {
    $span = $testSpan | ConvertFrom-Json
    $span.resourceSpans[0].instrumentationLibrarySpans[0].spans[0].name = "test-span-$_"
    $spanJson = $span | ConvertTo-Json -Depth 10
    Invoke-RestMethod -Uri "http://localhost:4318/v1/traces" -Method POST -Body $spanJson -ContentType "application/json"
    Start-Sleep -Milliseconds 100
}
```

## 输出规范

```
📡 OpenTelemetry 诊断报告

📈 Collector 状态
- 版本：[version]
- 运行时间：[uptime]
- 健康状态：[healthy/degraded]
- 配置状态：[valid/invalid]

💾 资源使用
- CPU 使用率：[cpu]%
- 内存使用：[memory] MB
- 句柄数：[handles]
- 线程数：[threads]

📊 数据流统计
| 组件 | 接收数 | 处理数 | 导出数 | 丢弃数 |
|------|--------|--------|--------|--------|
| Traces | [received] | [processed] | [exported] | [dropped] |
| Metrics | [received] | [processed] | [exported] | [dropped] |
| Logs | [received] | [processed] | [exported] | [dropped] |

🌐 端口状态
| 端口 | 协议 | 用途 | 状态 |
|------|------|------|------|
| 4317 | gRPC | OTLP | [listening/closed] |
| 4318 | HTTP | OTLP | [listening/closed] |
| 13133 | HTTP | Health | [listening/closed] |
| 8888 | HTTP | Metrics | [listening/closed] |

🚨 错误统计
- 接收错误：[receiver_errors]
- 处理错误：[processor_errors]
- 导出错误：[exporter_errors]
- 超时错误：[timeout_errors]

💡 优化建议
- [建议1]
- [建议2]
- [建议3]
```

## 参考资源

- [OpenTelemetry 官方文档](https://opentelemetry.io/docs/)
- [OpenTelemetry Collector 文档](https://opentelemetry.io/docs/collector/)
- [OTLP 协议规范](https://opentelemetry.io/docs/reference/specification/protocol/)
- [OpenTelemetry 语义约定](https://opentelemetry.io/docs/concepts/semantic-conventions/)
- [Collector 配置示例](https://github.com/open-telemetry/opentelemetry-collector/tree/main/examples)
