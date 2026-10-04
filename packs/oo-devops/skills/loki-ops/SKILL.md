---
name: loki-ops
description: Grafana Loki 运维专家 - 日志聚合、标签索引、查询优化、分布式部署、与 Grafana 集成
mode: on_demand
version: 1.1.0
maturity: experimental
---

## 配置说明

### 环境变量配置
```bash
# Loki 连接配置
export LOKI_HOST="http://localhost:3100"
export LOKI_USERNAME=""
export LOKI_PASSWORD=""
export LOKI_ORG_ID=""

# LogCLI 配置
export LOKI_ADDR="http://localhost:3100"
export LOKI_USERNAME=""
export LOKI_PASSWORD=""
```

### 配置文件示例
```yaml
# /etc/loki/local-config.yaml
auth_enabled: false

server:
  http_listen_port: 3100
  grpc_listen_port: 9096

common:
  path_prefix: /tmp/loki
  storage:
    filesystem:
      chunks_directory: /tmp/loki/chunks
      rules_directory: /tmp/loki/rules
  replication_factor: 1
  ring:
    instance_addr: 127.0.0.1
    kvstore:
      store: inmemory

schema_config:
  configs:
    - from: 2020-10-24
      store: boltdb-shipper
      object_store: filesystem
      schema: v11
      index:
        prefix: index_
        period: 24h

limits_config:
  retention_period: 168h
  max_query_length: 721h
```

## 输入参数

| 参数名 | 类型 | 必填 | 描述 | 示例 |
|--------|------|------|------|------|
| `query` | string | 是 | LogQL 查询语句 | `{job="nginx"} |= "error"` |
| `start` | string | 否 | 开始时间 | `2024-01-15T00:00:00Z` |
| `end` | string | 否 | 结束时间 | `2024-01-15T23:59:59Z` |
| `limit` | number | 否 | 返回条数 | `100` |

## 输出格式

### 日志查询输出
```json
{
  "status": "success",
  "data": {
    "resultType": "streams",
    "result": [
      {
        "stream": {
          "job": "nginx",
          "level": "error"
        },
        "values": [
          ["1705315200000000000", "2024/01/15 12:00:00 [error] request failed"]
        ]
      }
    ],
    "stats": {
      "summary": {
        "bytesProcessedPerSecond": 1048576,
        "linesProcessedPerSecond": 10000
      }
    }
  }
}
```

# Grafana Loki 运维助手

你是 Grafana Loki 日志聚合专家，擅长构建水平可扩展、高成本效益的日志系统，与 Prometheus 标签理念一致。

## 核心能力

- **Loki 架构管理**：Distributor、Ingester、Querier、Ruler、Compactor
- **部署模式**：单体模式、简单可扩展模式、分布式微服务模式
- **日志收集**：Promtail、Fluent Bit、Docker Driver、Lambda Promtail
- **标签策略**：高效标签设计、基数控制、结构化元数据
- **查询优化**：LogQL 语法、查询范围限制、缓存策略
- **存储管理**：对象存储配置、索引管理、保留策略
- **告警规则**：Loki Ruler、Recording Rules、与 Alertmanager 集成

## 标准诊断流程

### Linux/macOS

```bash
# 1. 检查 Loki 状态
curl -s http://localhost:3100/ready

# 2. 检查 Promtail 状态
curl -s http://localhost:9080/ready

# 3. 查看 Loki 日志
journalctl -u loki -f
tail -f /var/log/loki/loki.log

# 4. 查看 Promtail 日志
journalctl -u promtail -f
tail -f /var/log/promtail/promtail.log

# 5. 检查端口监听
netstat -tlnp | grep -E "3100|9095|9096|7946"

# 6. 检查存储连接
# S3
aws s3 ls s3://loki-bucket/

# 7. 使用 logcli 检查日志
logcli labels
logcli query '{job="varlogs"}'

# 8. 检查成员列表（分布式模式）
curl -s http://localhost:3100/memberlist | jq
```

### Windows (PowerShell)

```powershell
# 1. 检查 Loki 服务状态
Get-Service loki

# 2. 检查 Promtail 服务状态
Get-Service promtail

# 3. 查看 Loki 日志
Get-Content "C:\ProgramData\Loki\logs\loki.log" -Wait

# 4. 查看 Promtail 日志
Get-Content "C:\ProgramData\Promtail\logs\promtail.log" -Tail 100

# 5. 检查端口监听
Get-NetTCPConnection -LocalPort 3100,9095,9096,7946 | Select-Object LocalAddress, LocalPort, State

# 6. 检查 Loki 进程
Get-Process | Where-Object {$_.ProcessName -like "*loki*"}

# 7. 使用 logcli
& "C:\Program Files\Loki\logcli.exe" labels
& "C:\Program Files\Loki\logcli.exe" query '{job="varlogs"}'

# 8. 检查配置
Test-Path "C:\ProgramData\Loki\loki.yml"
Test-Path "C:\ProgramData\Promtail\promtail.yml"
```

## 常见故障处理

### 1. 查询超时或慢

#### Linux/macOS
```bash
# 检查查询范围是否合理
# 避免查询超过 7 天的日志

# 查看查询队列
curl -s http://localhost:3100/metrics | grep loki_querier

# 检查 Ingester 内存
curl -s http://localhost:3100/metrics | grep process_resident_memory_bytes

# 优化查询 - 使用时间过滤器和标签过滤器
logcli query '{job="varlogs"} |= "error"' --from="2024-01-01T00:00:00Z" --to="2024-01-01T01:00:00Z"

# 检查对象存储延迟
curl -s http://localhost:3100/metrics | grep loki_storage
```

#### Windows (PowerShell)
```powershell
# 获取查询指标
$metrics = Invoke-RestMethod -Uri "http://localhost:3100/metrics"
$metrics -split "`n" | Select-String "loki_querier|loki_ingester" | Select-Object -First 20

# 检查 Loki 资源使用
Get-Process | Where-Object {$_.ProcessName -like "*loki*"} |
    Select-Object ProcessName, Id, @{N="MemoryMB";E={[math]::Round($_.WorkingSet64/1MB,2)}}, CPU

# 使用 logcli 优化查询
& "C:\Program Files\Loki\logcli.exe" query '{job="varlogs"} |= "error"' --limit=100

# 检查存储指标
$storageMetrics = Invoke-RestMethod -Uri "http://localhost:3100/metrics"
$storageMetrics -split "`n" | Select-String "loki_storage|loki_bigtable|loki_aws_s3"
```

### 2. Ingester 内存过高

#### Linux/macOS
```bash
# 检查 Ingester 内存使用
curl -s http://ingester:3100/metrics | grep process_resident_memory_bytes

# 查看 Flush 队列
curl -s http://ingester:3100/metrics | grep loki_ingester_flush_queue_length

# 调整 Flush 间隔
# loki.yml: ingester.flush_period

# 手动触发 Flush
# 重启 Ingester（会触发 Flush）
systemctl restart loki-ingester
```

#### Windows (PowerShell)
```powershell
# 检查 Ingester 内存
$ingesterMetrics = Invoke-RestMethod -Uri "http://ingester:3100/metrics"
$memoryLine = $ingesterMetrics -split "`n" | Select-String "process_resident_memory_bytes"
$memoryBytes = [double]($memoryLine -replace ".* ", "")
Write-Output "Ingester Memory: $([math]::Round($memoryBytes/1MB,2)) MB"

# 查看 Flush 队列
$flushQueue = $ingesterMetrics -split "`n" | Select-String "loki_ingester_flush_queue_length"
Write-Output "Flush Queue: $flushQueue"

# 重启 Ingester 服务
Restart-Service loki-ingester
```

### 3. Promtail 无法推送日志

#### Linux/macOS
```bash
# 检查 Promtail 配置
cat /etc/promtail/promtail.yml | grep -A5 "clients"

# 测试 Loki 连通性
curl -s http://loki:3100/ready

# 查看 Promtail 位置文件
cat /var/lib/promtail/positions.yml

# 检查目标抓取状态
curl -s http://localhost:9080/metrics | grep promtail_targets

# 重启 Promtail
systemctl restart promtail
```

#### Windows (PowerShell)
```powershell
# 检查 Promtail 配置
Select-String -Path "C:\ProgramData\Promtail\promtail.yml" -Pattern "clients|loki" -Context 2

# 测试 Loki 连通性
Test-NetConnection -ComputerName loki -Port 3100

# 查看位置文件
Get-Content "C:\ProgramData\Promtail\positions.yml"

# 检查目标状态
$promtailMetrics = Invoke-RestMethod -Uri "http://localhost:9080/metrics"
$promtailMetrics -split "`n" | Select-String "promtail_targets|promtail_files"

# 重启 Promtail
Restart-Service promtail
```

## 性能优化配置

### Loki 优化

```yaml
# loki.yml
auth_enabled: false

server:
  http_listen_port: 3100
  grpc_listen_port: 9095
  log_level: info

ingester:
  lifecycler:
    address: 127.0.0.1
    ring:
      kvstore:
        store: inmemory
      replication_factor: 1
    final_sleep: 0s
  chunk_idle_period: 5m
  chunk_retain_period: 30s
  max_transfer_retries: 0
  max_chunk_age: 1h
  chunk_target_size: 1048576
  chunk_encoding: snappy

schema_config:
  configs:
    - from: 2024-01-01
      store: boltdb-shipper
      object_store: filesystem
      schema: v11
      index:
        prefix: index_
        period: 24h

storage_config:
  boltdb_shipper:
    active_index_directory: /loki/boltdb-shipper-active
    cache_location: /loki/boltdb-shipper-cache
    cache_ttl: 24h
  filesystem:
    directory: /loki/chunks

limits_config:
  enforce_metric_name: false
  reject_old_samples: true
  reject_old_samples_max_age: 168h
  ingestion_rate_mb: 16
  ingestion_burst_size_mb: 32
  per_stream_rate_limit: 3MB
  per_stream_rate_limit_burst: 15MB
  max_query_length: 721h
  max_query_parallelism: 32
  max_entries_limit_per_query: 5000

chunk_store_config:
  max_look_back_period: 0s

table_manager:
  retention_deletes_enabled: true
  retention_period: 720h

frontend:
  max_outstanding_per_tenant: 2048
  compress_responses: true

cache:
  enable_fifocache: true
  fifocache:
    max_size_bytes: 500MB
    max_size_items: 5000
    validity: 24h
```

### Promtail 优化

```yaml
# promtail.yml
server:
  http_listen_port: 9080
  grpc_listen_port: 0
  log_level: info

positions:
  filename: /var/lib/promtail/positions.yml
  sync_period: 10s

clients:
  - url: http://loki:3100/loki/api/v1/push
    batchwait: 1s
    batchsize: 1048576
    timeout: 10s
    backoff_config:
      min_period: 100ms
      max_period: 5s
      max_retries: 10
    external_labels:
      host: ${HOSTNAME}
      cluster: production

scrape_configs:
  - job_name: varlogs
    static_configs:
      - targets:
          - localhost
        labels:
          job: varlogs
          __path__: /var/log/*.log
    pipeline_stages:
      - json:
          expressions:
            level: level
            msg: message
      - labels:
          level:
      - timestamp:
          source: time
          format: RFC3339

  - job_name: docker
    static_configs:
      - targets:
          - localhost
        labels:
          job: docker
          __path__: /var/lib/docker/containers/*/*.log
    pipeline_stages:
      - json:
          expressions:
            output: log
            stream: stream
            attrs:
      - json:
          source: attrs
          expressions:
            tag:
      - timestamp:
          source: time
          format: RFC3339
      - output:
          source: output
```

### Windows 特定配置

```powershell
# Promtail Windows Event Log 配置
$windowsConfig = @'
server:
  http_listen_port: 9080

positions:
  filename: C:\ProgramData\Promtail\positions.yml

clients:
  - url: http://loki:3100/loki/api/v1/push
    external_labels:
      host: ${COMPUTERNAME}
      os: windows

scrape_configs:
  - job_name: windows-eventlog
    windows_events:
      use_incoming_timestamp: false
      bookmark_path: C:\ProgramData\Promtail\bookmark.xml
      eventlog_name: "Application"
      xpath_query: "*"
      labels:
        job: windows-eventlog
    relabel_configs:
      - source_labels: ['computer']
        target_label: 'host'
'@
$windowsConfig | Out-File "C:\ProgramData\Promtail\promtail-windows.yml" -Encoding UTF8

# 安装 Promtail 为 Windows 服务
if (-not (Get-Service -Name "promtail" -ErrorAction SilentlyContinue)) {
    $serviceParams = @{
        Name = "promtail"
        BinaryPathName = '"C:\Program Files\Promtail\promtail.exe" -config.file="C:\ProgramData\Promtail\promtail.yml"'
        DisplayName = "Promtail Log Collector"
        StartupType = "Automatic"
        Description = "Grafana Promtail log collector for Loki"
    }
    New-Service @serviceParams
}

Start-Service promtail

# 配置防火墙
New-NetFirewallRule -DisplayName "Promtail HTTP" -Direction Inbound -LocalPort 9080 -Protocol TCP -Action Allow

# 创建日志收集监控脚本
$monitorScript = @'
$lokiEndpoint = "http://loki:3100"
$promtailEndpoint = "http://localhost:9080"

try {
    $lokiHealth = Invoke-RestMethod -Uri "$lokiEndpoint/ready" -TimeoutSec 5
    $promtailMetrics = Invoke-RestMethod -Uri "$promtailEndpoint/metrics" -TimeoutSec 5

    $lines = $promtailMetrics -split "`n"
    $targetCount = ($lines | Select-String "promtail_targets_active_total").Count

    if ($targetCount -eq 0) {
        Write-EventLog -LogName Application -Source "Promtail" -EventId 3001 -EntryType Warning -Message "No active Promtail targets"
    }
} catch {
    Write-EventLog -LogName Application -Source "Promtail" -EventId 3002 -EntryType Error -Message "Promtail health check failed: $_"
}
'@
$monitorScript | Out-File "C:\ProgramData\Promtail\scripts\health_check.ps1" -Encoding UTF8
```

## 常用 API 操作

### Linux/macOS

```bash
# 使用 logcli 查询日志
# 安装 logcli
curl -fSL -o "/usr/local/bin/logcli" "https://github.com/grafana/loki/releases/download/v2.9.0/logcli-linux-amd64.zip"
chmod +x /usr/local/bin/logcli

# 设置环境变量
export LOKI_ADDR=http://localhost:3100

# 查询标签
logcli labels
logcli labels job

# 查询日志
logcli query '{job="varlogs"}'
logcli query '{job="varlogs"} |= "error"'
logcli query '{job="varlogs"} |= "error" != "debug"'

# 范围查询
logcli query '{job="varlogs"} |= "error"' --since=1h
logcli query '{job="varlogs"} |= "error"' --from="2024-01-01T00:00:00Z" --to="2024-01-02T00:00:00Z"

# 统计查询
logcli query 'sum(rate({job="varlogs"} |= "error" [1m])) by (level)'

# 获取 Loki 指标
curl -s http://localhost:3100/metrics | grep loki

# 检查存储
curl -s http://localhost:3100/loki/api/v1/status/buildinfo
```

### Windows (PowerShell)

```powershell
# 设置 logcli 环境
$env:LOKI_ADDR = "http://localhost:3100"

# 查询标签
& "C:\Program Files\Loki\logcli.exe" labels
& "C:\Program Files\Loki\logcli.exe" labels job

# 查询日志
& "C:\Program Files\Loki\logcli.exe" query '{job="varlogs"}'
& "C:\Program Files\Loki\logcli.exe" query '{job="varlogs"} |= "error"'

# 范围查询
& "C:\Program Files\Loki\logcli.exe" query '{job="varlogs"} |= "error"' --since=1h

# 使用 REST API 直接查询
$query = '{job="varlogs"} |= "error"'
$encodedQuery = [Uri]::EscapeDataString($query)
$logs = Invoke-RestMethod -Uri "http://localhost:3100/loki/api/v1/query_range?query=$encodedQuery&limit=100"

$logs.data.result | ForEach-Object {
    $stream = $_.stream
    $_.values | ForEach-Object {
        [PSCustomObject]@{
            Timestamp = ([DateTimeOffset]::FromUnixTimeMilliseconds($_[0]/1000000)).DateTime
            Message = $_[1]
            Job = $stream.job
            Level = $stream.level
        }
    }
} | Format-Table -AutoSize

# 获取 Loki 指标
$metrics = Invoke-RestMethod -Uri "http://localhost:3100/metrics"
$metrics -split "`n" | Select-String "loki_ingester|loki_querier|loki_distributor" | Select-Object -First 20

# 检查构建信息
$buildInfo = Invoke-RestMethod -Uri "http://localhost:3100/loki/api/v1/status/buildinfo"
$buildInfo | Format-List

# 生成日志分析报告
$report = @{
    GeneratedAt = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    LokiVersion = (Invoke-RestMethod -Uri "http://localhost:3100/loki/api/v1/status/buildinfo").version
    TotalStreams = (& "C:\Program Files\Loki\logcli.exe" series '{job=~".+"}' 2>$null).Count
    RecentErrors = (& "C:\Program Files\Loki\logcli.exe" query '{job="varlogs"} |= "error"' --limit=1000 2>$null).Count
}
$report | ConvertTo-Json | Out-File "C:\Reports\loki_status.json" -Encoding UTF8
```

## 输出规范

```
🔥 Loki 诊断报告

📊 系统状态
- Loki 版本：[version]
- 运行模式：[monolithic/simple-scalable/distributed]
- 运行时间：[uptime]
- 健康状态：[ready/not-ready]

📈 摄入统计
| 组件 | 摄入速率 | 活跃流 | 内存使用 |
|------|----------|--------|----------|
| Distributor | [rate] MB/s | [streams] | [memory] |
| Ingester | [rate] MB/s | [streams] | [memory] |

🔍 查询性能
| 指标 | 当前值 | 阈值 | 状态 |
|------|--------|------|------|
| 查询延迟 P99 | [latency]s | <30s | [ok/warn] |
| 查询队列 | [queue] | <100 | [ok/warn] |
| 缓存命中率 | [hit]% | >80% | [ok/warn] |

💾 存储统计
| 类型 | 大小 | 保留期 | 状态 |
|------|------|--------|------|
| 索引 | [size] | [retention] | [ok] |
| 块存储 | [size] | [retention] | [ok] |
| 压缩后 | [size] | - | [ok] |

🚨 活动告警
| 规则 | 状态 | 最后触发 |
|------|------|----------|
| [rule1] | [firing/resolved] | [time] |

💡 优化建议
- [建议1]
- [建议2]
```

## 参考资源

- [Loki 官方文档](https://grafana.com/docs/loki/latest/)
- [LogQL 查询语言](https://grafana.com/docs/loki/latest/logql/)
- [Loki 架构](https://grafana.com/docs/loki/latest/fundamentals/architecture/)
- [Promtail 配置](https://grafana.com/docs/loki/latest/clients/promtail/)
- [Loki 存储](https://grafana.com/docs/loki/latest/operations/storage/)
