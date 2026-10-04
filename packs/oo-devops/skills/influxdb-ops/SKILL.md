---
name: influxdb-ops
description: InfluxDB 时序数据库运维专家 - 性能优化、数据保留、集群管理
mode: on_demand
version: 1.1.0
maturity: experimental
---

## 配置说明

### 环境变量配置
```bash
export INFLUXDB_URL="http://localhost:8086"
export INFLUXDB_TOKEN=""
export INFLUXDB_ORG="myorg"
export INFLUXDB_BUCKET="mybucket"
```

## 输入参数

| 参数名 | 类型 | 必填 | 描述 | 示例 |
|--------|------|------|------|------|
| `bucket` | string | 否 | Bucket 名 | `metrics` |
| `measurement` | string | 否 | Measurement | `cpu` |

## 输出格式

```json
{
  "status": "success",
  "data": {
    "buckets": 3,
    "series": 150,
    "points": 1000000
  }
}
```

# InfluxDB 运维助手

你是 InfluxDB 时序数据库运维专家，擅长性能优化、数据保留策略、集群管理和故障排查。

## 核心能力

- **数据库管理**：数据库创建、RP（保留策略）、连续查询、CQ 管理
- **性能优化**：内存配置、索引优化、批量写入、查询优化
- **集群管理**：高可用架构、数据分片、副本配置、集群扩展
- **数据保留**：RP 配置、数据降采样、过期清理、备份恢复
- **监控告警**：内置监控、Grafana 集成、告警规则
- **安全加固**：认证授权、SSL/TLS、访问控制
- **数据迁移**：导出导入、协议兼容、版本升级

## 标准诊断流程

### Linux

```bash
# 1. 检查 InfluxDB 状态
influx -execute "SHOW DIAGNOSTICS"

# 2. 查看数据库列表
influx -execute "SHOW DATABASES"

# 3. 查看保留策略
influx -database mydb -execute "SHOW RETENTION POLICIES"

# 4. 查看统计信息
influx -execute "SHOW STATS"

# 5. 查看日志
tail -f /var/log/influxdb/influxd.log
```

### Windows PowerShell

```powershell
# 1. 检查 InfluxDB 状态
influx -execute "SHOW DIAGNOSTICS"

# 2. 查看数据库列表
influx -execute "SHOW DATABASES"

# 3. 查看保留策略
influx -database mydb -execute "SHOW RETENTION POLICIES"

# 4. 查看统计信息
influx -execute "SHOW STATS"

# 5. 查看 InfluxDB 服务
Get-Service | Where-Object {$_.Name -like "*influx*"}

# 6. 重启 InfluxDB 服务
Restart-Service InfluxDB

# 7. 查看 InfluxDB 进程
Get-Process | Where-Object {$_.ProcessName -like "*influx*"}

# 8. 检查端口监听
Get-NetTCPConnection -LocalPort 8086

# 9. 查看日志 (Windows 路径)
Get-Content "C:\Program Files\InfluxData\InfluxDB\influxd.log" -Wait

# 10. 检查数据目录磁盘空间
Get-ChildItem "C:\Program Files\InfluxData\InfluxDB\data" | Measure-Object -Property Length -Sum

# 11. 查看 Windows 事件日志
Get-EventLog -LogName Application -Source "InfluxDB*" -Newest 20

# 12. 检查配置文件
Get-Content "C:\Program Files\InfluxData\InfluxDB\influxdb.conf"

# 13. 查看 WAL 目录大小
Get-ChildItem "C:\Program Files\InfluxData\InfluxDB\wal" -Recurse | Measure-Object -Property Length -Sum
```

## 常见故障处理

### 1. 内存占用过高
```bash
# 查看内存配置
cat /etc/influxdb/influxdb.conf | grep -A 5 "\[coordinator\]"

# 关键参数：
# max-concurrent-queries = 0
# query-timeout = "0s"
# max-select-point = 0
# max-select-series = 0

# 查看当前查询
influx -execute "SHOW QUERIES"

# 杀死慢查询
influx -execute "KILL QUERY <qid>"

# 优化建议：
# 1. 限制并发查询数
# 2. 设置查询超时
# 3. 增加内存限制
# 4. 优化查询语句
```

### 2. 写入性能问题
```bash
# 查看写入统计
influx -execute "SHOW STATS WRITE"

# 检查 WAL 目录大小
du -sh /var/lib/influxdb/wal

# 优化建议：
# 1. 启用批量写入
# 2. 使用 HTTP 而非 UDP
# 3. 调整 WAL 配置
# 4. 增加缓存大小

# influxdb.conf
[data]
  cache-max-memory-size = "1g"
  cache-snapshot-memory-size = "25m"
  cache-snapshot-write-cold-duration = "10m"
  compact-full-write-cold-duration = "4h"
  max-concurrent-compactions = 3
```

## Windows PowerShell 运维脚本

### 服务管理
```powershell
# 检查 InfluxDB 服务状态
Get-Service -Name "InfluxDB"

# 启动 InfluxDB 服务
Start-Service -Name "InfluxDB"

# 停止 InfluxDB 服务
Stop-Service -Name "InfluxDB"

# 重启 InfluxDB 服务
Restart-Service -Name "InfluxDB"

# 设置服务自动启动
Set-Service -Name "InfluxDB" -StartupType Automatic
```

### 日志监控
```powershell
# 实时监控 InfluxDB 日志
Get-Content "C:\Program Files\InfluxData\InfluxDB\influxd.log" -Wait

# 查找错误日志
Select-String -Path "C:\Program Files\InfluxData\InfluxDB\*.log" -Pattern "ERROR|FATAL" -Context 2

# 查看最近 100 条日志
Get-Content "C:\Program Files\InfluxData\InfluxDB\influxd.log" -Tail 100
```

### 性能监控
```powershell
# 检查 InfluxDB 进程资源使用
Get-Process | Where-Object {$_.ProcessName -like "*influx*"} |
    Select-Object Name, Id, CPU, WorkingSet, PagedMemorySize

# 检查端口连接
Get-NetTCPConnection -LocalPort 8086 | Measure-Object

# 检查磁盘空间 (InfluxDB 数据目录)
$walSize = (Get-ChildItem "C:\Program Files\InfluxData\InfluxDB\wal" -Recurse | Measure-Object -Property Length -Sum).Sum
dataSize = (Get-ChildItem "C:\Program Files\InfluxData\InfluxDB\data" -Recurse | Measure-Object -Property Length -Sum).Sum
Write-Output "WAL Size: $([math]::Round($walSize/1MB,2)) MB"
Write-Output "Data Size: $([math]::Round($dataSize/1MB,2)) MB"

# 检查内存使用
Get-CimInstance Win32_OperatingSystem | Select-Object TotalVisibleMemorySize, FreePhysicalMemory
```

### 备份恢复 (Windows)
```powershell
# 备份数据
influxd backup -portable "C:\Backups\influxdb"

# 恢复数据
influxd restore -portable -db mydb "C:\Backups\influxdb"

# 压缩备份文件
Compress-Archive -Path "C:\Backups\influxdb" -DestinationPath "C:\Backups\influxdb-$(Get-Date -Format 'yyyyMMdd').zip"
```

### 3. 数据丢失或损坏

#### Linux
```bash
# 检查分片状态
influx -execute "SHOW SHARDS"

# 查看分片组
influx -execute "SHOW SHARD GROUPS"

# 备份数据
influxd backup -portable /backup/influxdb

# 恢复数据
influxd restore -portable -db mydb /backup/influxdb
```

#### Windows PowerShell
```powershell
# 检查分片状态
influx -execute "SHOW SHARDS"

# 查看分片组
influx -execute "SHOW SHARD GROUPS"

# 备份数据
influxd backup -portable "C:\Backups\influxdb"

# 恢复数据
influxd restore -portable -db mydb "C:\Backups\influxdb"

# 检查分片文件完整性
Get-ChildItem "C:\Program Files\InfluxData\InfluxDB\data\mydb" -Recurse | Select-Object Name, Length, LastWriteTime
```

# 常见原因：
# - 磁盘空间不足
# - 保留策略删除过期数据
# - 分片损坏
# - 权限问题
```

### 4. 查询超时
```bash
# 分析慢查询
influx -execute "SHOW QUERIES"

# 添加查询限制
influx -execute "SET MAX_ROW_LIMIT = 10000"

# 使用连续查询预处理数据
CREATE CONTINUOUS QUERY cq_1h ON mydb
BEGIN
  SELECT mean(value) INTO mydb.rp_1h.measurement_1h
  FROM mydb.rp_default.measurement
  GROUP BY time(1h),*
END
```

## 配置优化

```toml
# /etc/influxdb/influxdb.conf

[meta]
  dir = "/var/lib/influxdb/meta"

[data]
  dir = "/var/lib/influxdb/data"
  wal-dir = "/var/lib/influxdb/wal"

  # 缓存配置
  cache-max-memory-size = "2g"
  cache-snapshot-memory-size = "50m"
  max-series-per-database = 1000000
  max-values-per-tag = 100000

[coordinator]
  write-timeout = "10s"
  max-concurrent-queries = 10
  query-timeout = "30s"
  max-select-point = 10000000
  max-select-series = 100000

[retention]
  enabled = true
  check-interval = "30m"

[monitor]
  store-enabled = true
  store-database = "_internal"
  store-interval = "10s"

[http]
  enabled = true
  bind-address = ":8086"
  auth-enabled = true
  max-row-limit = 100000
  max-connection-limit = 0

[logging]
  format = "auto"
  level = "info"
  suppress-logo = false
```

## 输出规范

```
⏱️ InfluxDB 诊断报告

📊 数据库状态
- 版本：[version]
- 数据库数：[databases]
- 序列数：[series]
- 写入速率：[write rate]

🔍 问题发现
1. [问题描述]

💡 解决方案
[处理步骤]
```
