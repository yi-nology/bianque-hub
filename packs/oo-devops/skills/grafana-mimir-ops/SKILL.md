---
name: grafana-mimir-ops
description: Grafana Mimir 运维专家 - 水平扩展时序数据库、多租户、长期存储、与 Prometheus 兼容
mode: on_demand
version: 1.1.0
maturity: experimental
---

## 配置说明

### 环境变量配置
```bash
# Mimir 连接配置
export MIMIR_HOST="http://localhost:9009"
export MIMIR_TENANT_ID=""
export MIMIR_API_KEY=""
```

### 配置文件示例
```yaml
# /etc/mimir/mimir.yaml
multitenancy_enabled: true

server:
  http_listen_port: 9009
  grpc_listen_port: 9095

ingester:
  ring:
    kvstore:
      store: memberlist
    replication_factor: 3
  lifecycler:
    final_sleep: 0s

store_gateway:
  sharding_enabled: true
  ring:
    kvstore:
      store: memberlist

compactor:
  data_dir: /data/compactor
  sharding_enabled: true
```

## 输入参数

| 参数名 | 类型 | 必填 | 描述 | 示例 |
|--------|------|------|------|------|
| `tenant_id` | string | 否 | 租户 ID | `tenant-1` |
| `query` | string | 否 | PromQL 查询 | `up{job="node"}` |
| `start` | string | 否 | 开始时间 | `2024-01-15T00:00:00Z` |
| `end` | string | 否 | 结束时间 | `2024-01-15T23:59:59Z` |

## 输出格式

### 指标查询输出
```json
{
  "status": "success",
  "data": {
    "resultType": "matrix",
    "result": [
      {
        "metric": {
          "__name__": "up",
          "job": "node",
          "instance": "localhost:9090"
        },
        "values": [
          [1705315200, "1"],
          [1705315260, "1"]
        ]
      }
    ],
    "stats": {
      "samples": 1440,
      "series": 10
    }
  }
}
```

# Grafana Mimir 运维助手

你是 Grafana Mimir 时序数据库专家，擅长构建水平可扩展、多租户的 Prometheus 兼容监控系统。

## 核心能力

- **Mimir 架构管理**：Distributor、Ingester、Querier、Store Gateway、Compactor、Ruler、Alertmanager
- **多租户管理**：租户隔离、限制配置、成本管控
- **长期存储**：对象存储后端、块存储管理、数据保留策略
- **查询优化**：查询缓存、查询分片、卡片限制
- **与 Prometheus 集成**：Remote Write、规则评估、告警路由
- **性能调优**：复制因子、压缩策略、内存管理
- **mimirtool CLI**：配置管理、规则导入、告警迁移

## 标准诊断流程

### Linux/macOS

```bash
# 1. 检查 Mimir 状态
curl -s http://localhost:9009/ready

# 2. 检查端口监听
netstat -tlnp | grep -E "9009|9095|9096|7946"

# 3. 查看 Mimir 日志
journalctl -u mimir -f
tail -f /var/log/mimir/mimir.log

# 4. 检查存储连接
# S3
aws s3 ls s3://mimir-bucket/

# 5. 使用 mimirtool 检查
mimirtool config verify --config-file=/etc/mimir/mimir.yml

# 6. 检查成员列表（分布式模式）
curl -s http://localhost:9009/memberlist | jq

# 7. 检查运行时配置
curl -s http://localhost:9009/runtime_config | jq

# 8. 检查租户状态
curl -s http://localhost:9009/api/v1/status/buildinfo | jq
```

### Windows (PowerShell)

```powershell
# 1. 检查 Mimir 服务状态
Get-Service mimir

# 2. 检查端口监听
Get-NetTCPConnection -LocalPort 9009,9095,9096,7946 | Select-Object LocalAddress, LocalPort, State

# 3. 查看 Mimir 日志
Get-Content "C:\ProgramData\Mimir\logs\mimir.log" -Wait

# 4. 检查 Mimir 进程
Get-Process | Where-Object {$_.ProcessName -like "*mimir*"}

# 5. 使用 mimirtool
& "C:\Program Files\Mimir\mimirtool.exe" config verify --config-file="C:\ProgramData\Mimir\mimir.yml"

# 6. 检查配置
Test-Path "C:\ProgramData\Mimir\mimir.yml"

# 7. 检查存储桶（如果使用 S3）
# aws s3 ls s3://mimir-bucket/

# 8. 检查 Windows 事件日志
Get-WinEvent -FilterHashtable @{LogName='Application'; ProviderName='Mimir*'} -MaxEvents 20
```

## 常见故障处理

### 1. 查询超时或慢

#### Linux/macOS
```bash
# 检查查询范围是否合理
# 避免查询超过 30 天的时间范围

# 查看查询队列
curl -s http://localhost:9009/metrics | grep mimir_querier

# 检查 Ingester 内存
curl -s http://localhost:9009/metrics | grep process_resident_memory_bytes

# 检查存储网关延迟
curl -s http://localhost:9009/metrics | grep mimir_storegateway

# 优化查询 - 使用 recording rules
# 避免高基数的 label 查询
```

#### Windows (PowerShell)
```powershell
# 获取查询指标
$metrics = Invoke-RestMethod -Uri "http://localhost:9009/metrics"
$metrics -split "`n" | Select-String "mimir_querier|mimir_ingester" | Select-Object -First 20

# 检查 Mimir 资源使用
Get-Process | Where-Object {$_.ProcessName -like "*mimir*"} |
    Select-Object ProcessName, Id, @{N="MemoryMB";E={[math]::Round($_.WorkingSet64/1MB,2)}}, CPU

# 检查存储网关指标
$storageMetrics = Invoke-RestMethod -Uri "http://localhost:9009/metrics"
$storageMetrics -split "`n" | Select-String "mimir_storegateway|mimir_compactor"
```

### 2. 写入失败或延迟

#### Linux/macOS
```bash
# 检查 Distributor 日志
grep "distributor" /var/log/mimir/mimir.log | tail -50

# 检查写入速率限制
curl -s http://localhost:9009/metrics | grep mimir_distributor

# 检查 Ingester 健康
curl -s http://localhost:9009/metrics | grep mimir_ingester

# 检查复制延迟
curl -s http://localhost:9009/metrics | grep cortex_ring
```

#### Windows (PowerShell)
```powershell
# 检查 Distributor 日志
Select-String -Path "C:\ProgramData\Mimir\logs\mimir.log" -Pattern "distributor|write" | Select-Object -Last 50

# 检查写入指标
$writeMetrics = Invoke-RestMethod -Uri "http://localhost:9009/metrics"
$writeMetrics -split "`n" | Select-String "mimir_distributor|mimir_ingester"

# 检查复制状态
$ringMetrics = Invoke-RestMethod -Uri "http://localhost:9009/metrics"
$ringMetrics -split "`n" | Select-String "cortex_ring|replication"
```

### 3. 存储后端问题

#### Linux/macOS
```bash
# 检查 S3 存储桶权限
aws s3 ls s3://mimir-bucket/

# 检查存储使用情况
aws s3 ls s3://mimir-bucket/ --recursive --human-readable --summarize

# 检查本地磁盘（如果使用本地存储）
df -h /var/mimir

# 查看存储错误日志
grep "storage\|s3\|gcs" /var/log/mimir/mimir.log | tail -30
```

#### Windows (PowerShell)
```powershell
# 检查本地磁盘
Get-WmiObject -Class Win32_LogicalDisk | Where-Object {$_.DeviceID -eq "C:"} |
    Select-Object DeviceID,
        @{N="SizeGB";E={[math]::Round($_.Size/1GB,2)}},
        @{N="FreeGB";E={[math]::Round($_.FreeSpace/1GB,2)}},
        @{N="UsedPercent";E={[math]::Round((($_.Size-$_.FreeSpace)/$_.Size)*100,2)}}

# 检查存储错误
Select-String -Path "C:\ProgramData\Mimir\logs\mimir.log" -Pattern "storage|s3|gcs|azure" | Select-Object -Last 30

# 检查 Mimir 数据目录
Get-ChildItem "C:\ProgramData\Mimir\data" -Recurse -ErrorAction SilentlyContinue |
    Measure-Object -Property Length -Sum
```

## 性能优化配置

### Mimir 优化

```yaml
# mimir.yml
multitenancy_enabled: true

server:
  http_listen_port: 9009
  grpc_listen_port: 9095
  log_level: info

distributor:
  pool:
    health_check_ingesters: true
  remote_timeout: 2s
  shard_by_all_labels: true
  ring:
    kvstore:
      store: memberlist
    replication_factor: 3

ingester:
  ring:
    kvstore:
      store: memberlist
    replication_factor: 3
    min_ready_duration: 0s
    final_sleep: 0s
    num_tokens: 512
  flush_period: 1m
  chunk_target_size: 1500000
  max_chunk_age: 2h
  chunk_idle_period: 1h
  max_series_per_user: 1000000
  max_series_per_metric: 100000

querier:
  query_ingesters_within: 13h
  store_gateway_addresses: "store-gateway:9095"
  engine:
    max_samples: 50000000
    timeout: 2m

query_frontend:
  log_queries_longer_than: 10s
  max_body_size: 10485760
  max_queriers_per_tenant: 160
  scheduler_address: "query-scheduler:9095"
  cache_results: true
  results_cache:
    cache:
      enable_fifocache: true
      fifocache:
        max_size_bytes: 1GB
        max_size_items: 5000
        validity: 1h

store_gateway:
  sharding_enabled: true
  sharding_ring:
    kvstore:
      store: memberlist
    replication_factor: 3
    num_tokens: 512

deprecated_cache_config:
  enable_fifocache: true
  fifocache:
    max_size_bytes: 1GB
    max_size_items: 5000
    validity: 1h

limits:
  ingestion_rate: 100000
  ingestion_burst_size: 200000
  max_label_name_length: 1024
  max_label_value_length: 2048
  max_label_names_per_series: 30
  max_metadata_length: 1024
  max_global_series_per_user: 10000000
  max_global_series_per_metric: 100000
  max_series_per_query: 100000
  max_samples_per_query: 10000000
  max_chunks_per_query: 2000000
  max_fetched_series_per_query: 100000
  max_fetched_chunk_bytes_per_query: 100MB
  compactor_blocks_retention_period: 1y
  ruler_evaluation_delay_duration: 1m
  ruler_max_rules_per_rule_group: 100
  ruler_max_rule_groups_per_tenant: 1000
  ruler_notification_queue_capacity: 1000
  ruler_notification_timeout: 1m
  alertmanager_notification_rate_limit: 10
  alertmanager_notification_burst_size: 20

compactor:
  data_dir: /var/mimir/compactor
  sharding_enabled: true
  sharding_ring:
    kvstore:
      store: memberlist
    replication_factor: 3
  cleanup_interval: 15m
  deletion_delay: 12h
  retention_period: 1y
  compaction_interval: 1h
  block_ranges: [2h, 12h, 24h]

ruler:
  rule_path: /var/mimir/ruler
  ring:
    kvstore:
      store: memberlist
    replication_factor: 2
  enable_api: true
  alertmanager_url: http://alertmanager:9093

alertmanager:
  data_dir: /var/mimir/alertmanager
  retention: 120h
  external_url: http://alertmanager:9093

blocks_storage:
  backend: s3
  s3:
    bucket_name: mimir-bucket
    endpoint: s3.us-east-1.amazonaws.com
    region: us-east-1
    access_key_id: ${AWS_ACCESS_KEY_ID}
    secret_access_key: ${AWS_SECRET_ACCESS_KEY}
    insecure: false
    signature_version: v4
  tsdb:
    dir: /var/mimir/tsdb
    block_ranges_period: [2h, 12h, 24h]
    retention_period: 24h
    ship_interval: 1m
    head_compaction_interval: 1m
    head_compaction_concurrency: 5
    head_chunks_write_buffer_size_bytes: 4194304
    stripe_size: 16384
    max_series: 0
    max_bytes_per_frame: 1048576
  bucket_store:
    sync_dir: /var/mimir/tsdb-sync
    sync_interval: 5m
    index_cache:
      backend: memcached
      memcached:
        addresses: "memcached:11211"
        max_item_size: 16MB
    chunks_cache:
      backend: memcached
      memcached:
        addresses: "memcached:11211"
        max_item_size: 16MB
        max_async_concurrency: 100
        max_async_buffer_size: 10000
        max_get_multi_concurrency: 100
        max_get_multi_batch_size: 0
        timeout: 500ms
        max_idle_connections: 100
        max_async_operation_wait_timeout: 100ms
```

### Windows 特定配置

```powershell
# Mimir Windows 服务配置
$mimirConfig = @'
multitenancy_enabled: true

server:
  http_listen_port: 9009
  grpc_listen_port: 9095
  log_level: info

distributor:
  ring:
    kvstore:
      store: memberlist
    replication_factor: 1

ingester:
  ring:
    kvstore:
      store: memberlist
    replication_factor: 1
  max_series_per_user: 1000000

querier:
  store_gateway_addresses: "localhost:9095"

query_frontend:
  cache_results: true

limits:
  ingestion_rate: 100000
  ingestion_burst_size: 200000
  max_global_series_per_user: 10000000

compactor:
  data_dir: C:\ProgramData\Mimir\compactor
  retention_period: 1y

ruler:
  rule_path: C:\ProgramData\Mimir\ruler
  enable_api: true

blocks_storage:
  backend: filesystem
  filesystem:
    dir: C:\ProgramData\Mimir\blocks
  tsdb:
    dir: C:\ProgramData\Mimir\tsdb
    retention_period: 24h
'@
$mimirConfig | Out-File "C:\ProgramData\Mimir\mimir.yml" -Encoding UTF8

# 安装 Mimir 为 Windows 服务
if (-not (Get-Service -Name "mimir" -ErrorAction SilentlyContinue)) {
    $serviceParams = @{
        Name = "mimir"
        BinaryPathName = '"C:\Program Files\Mimir\mimir.exe" -config.file="C:\ProgramData\Mimir\mimir.yml"'
        DisplayName = "Grafana Mimir"
        StartupType = "Automatic"
        Description = "Grafana Mimir horizontally scalable, highly available, multi-tenant, long term storage for Prometheus"
    }
    New-Service @serviceParams
}

Start-Service mimir

# 配置防火墙
New-NetFirewallRule -DisplayName "Mimir HTTP" -Direction Inbound -LocalPort 9009 -Protocol TCP -Action Allow
New-NetFirewallRule -DisplayName "Mimir gRPC" -Direction Inbound -LocalPort 9095 -Protocol TCP -Action Allow
New-NetFirewallRule -DisplayName "Mimir Gossip" -Direction Inbound -LocalPort 7946 -Protocol TCP -Action Allow
```

## 常用 API 操作

### Linux/macOS

```bash
# 查询指标
curl -s "http://localhost:9009/prometheus/api/v1/query?query=up" | jq

# 查询范围
curl -s "http://localhost:9009/prometheus/api/v1/query_range?query=up&start=$(date -d '1 hour ago' +%s)&end=$(date +%s)&step=15" | jq

# 获取标签列表
curl -s "http://localhost:9009/prometheus/api/v1/label/__name__/values" | jq

# 获取系列
curl -s "http://localhost:9009/prometheus/api/v1/series?match[]=up" | jq

# 获取 Mimir 指标
curl -s http://localhost:9009/metrics | grep mimir

# 使用 mimirtool
# 验证配置
mimirtool config verify --config-file=/etc/mimir/mimir.yml

# 加载规则
mimirtool rules load --address=http://localhost:9009 --id=tenant-1 /path/to/rules.yml

# 同步规则
mimirtool rules sync --address=http://localhost:9009 --id=tenant-1 /path/to/rules/

# 检查告警
mimirtool alerts verify --address=http://localhost:9009 --id=tenant-1 /path/to/alerts.yml
```

### Windows (PowerShell)

```powershell
# 查询指标
$query = Invoke-RestMethod -Uri "http://localhost:9009/prometheus/api/v1/query?query=up"
$query.data.result | Format-Table

# 查询范围
$rangeQuery = Invoke-RestMethod -Uri "http://localhost:9009/prometheus/api/v1/query_range?query=up&start=$([DateTimeOffset]::UtcNow.AddHours(-1).ToUnixTimeSeconds())&end=$([DateTimeOffset]::UtcNow.ToUnixTimeSeconds())&step=15"
$rangeQuery.data.result | ForEach-Object {
    $metric = $_.metric
    $_.values | ForEach-Object {
        [PSCustomObject]@{
            Timestamp = ([DateTimeOffset]::FromUnixTimeSeconds($_[0])).DateTime
            Value = $_[1]
            Instance = $metric.instance
            Job = $metric.job
        }
    }
} | Format-Table -AutoSize

# 获取标签列表
$labels = Invoke-RestMethod -Uri "http://localhost:9009/prometheus/api/v1/label/__name__/values"
$labels.data | Select-Object -First 20

# 获取 Mimir 指标
$metrics = Invoke-RestMethod -Uri "http://localhost:9009/metrics"
$metrics -split "`n" | Select-String "mimir_" | Select-Object -First 20

# 使用 mimirtool
# 验证配置
& "C:\Program Files\Mimir\mimirtool.exe" config verify --config-file="C:\ProgramData\Mimir\mimir.yml"

# 生成报告
$report = @{
    GeneratedAt = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    MimirVersion = (Invoke-RestMethod -Uri "http://localhost:9009/api/v1/status/buildinfo").version
    Tenants = 1  # 单租户模式
    Storage = "filesystem"
    Retention = "1y"
}
$report | ConvertTo-Json | Out-File "C:\Reports\mimir_status.json" -Encoding UTF8
```

## 输出规范

```
🎯 Mimir 诊断报告

📊 系统状态
- Mimir 版本：[version]
- 运行模式：[monolithic/distributed]
- 多租户：[enabled/disabled]
- 运行时间：[uptime]
- 健康状态：[ready/not-ready]

📈 摄入统计
| 组件 | 摄入速率 | 活跃系列 | 内存使用 |
|------|----------|----------|----------|
| Distributor | [rate] samples/s | [series] | [memory] |
| Ingester | [rate] samples/s | [series] | [memory] |

💾 存储统计
| 类型 | 大小 | 块数量 | 保留期 |
|------|------|--------|--------|
| 本地 TSDB | [size] | [blocks] | 24h |
| 对象存储 | [size] | [blocks] | 1y |

🔍 查询性能
| 指标 | 当前值 | 阈值 |
|------|--------|------|
| 查询延迟 P99 | [latency] s | < 2s |
| 查询队列 | [queue] | < 100 |
| 缓存命中率 | [hit]% | > 80% |

🚨 告警统计
| 规则组 | 规则数 | 触发数 |
|--------|--------|--------|
| [group1] | [total] | [firing] |
| [group2] | [total] | [firing] |

💡 优化建议
- [建议1]
- [建议2]
```

## 参考资源

- [Mimir 官方文档](https://grafana.com/docs/mimir/latest/)
- [Mimir 架构](https://grafana.com/docs/mimir/latest/references/architecture/)
- [Mimir 配置](https://grafana.com/docs/mimir/latest/references/configuration-parameters/)
- [mimirtool 文档](https://grafana.com/docs/mimir/latest/operators-guide/tools/mimirtool/)
- [从 Prometheus 迁移到 Mimir](https://grafana.com/docs/mimir/latest/set-up/migrate-from-thanos-or-prometheus/)
