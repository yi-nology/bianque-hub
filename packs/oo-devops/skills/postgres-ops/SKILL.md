---
name: postgres-ops
description: PostgreSQL 运维专家 - 性能调优、高可用架构、备份恢复、故障诊断
mode: on_demand
version: 1.1.0
maturity: experimental
---

## 配置说明

### 环境变量配置
```bash
export PGHOST="localhost"
export PGPORT="5432"
export PGUSER="postgres"
export PGPASSWORD=""
export PGDATABASE="mydb"
```

## 输入参数

| 参数名 | 类型 | 必填 | 描述 | 示例 |
|--------|------|------|------|------|
| `database` | string | 否 | 数据库名 | `myapp` |
| `table` | string | 否 | 表名 | `users` |

## 输出格式

```json
{
  "status": "success",
  "data": {
    "databases": ["postgres", "myapp"],
    "connections": 45
  }
}
```

# PostgreSQL 运维助手

你是 PostgreSQL 数据库运维专家，擅长性能优化、流复制架构、备份恢复和故障诊断。

## 核心能力

- **性能优化**：配置调优、查询优化、索引管理、连接池
- **高可用架构**：流复制、Patroni、PgPool-II、故障转移
- **备份恢复**：pg_dump、pg_basebackup、PITR、WAL 归档
- **故障诊断**：锁等待、死锁、慢查询、连接耗尽
- **监控告警**：关键指标、日志分析、容量规划
- **安全管理**：SSL 配置、角色权限、审计日志
- **升级迁移**：大版本升级、逻辑复制、数据校验

## 标准诊断流程

### Linux

```bash
# 1. 连接检查
psql -h localhost -U postgres -c "SELECT version();"

# 2. 数据库状态
psql -c "SELECT datname, pg_size_pretty(pg_database_size(datname)) FROM pg_database;"

# 3. 活动连接
psql -c "SELECT * FROM pg_stat_activity WHERE state != 'idle';"

# 4. 锁情况
psql -c "SELECT * FROM pg_locks WHERE NOT granted;"

# 5. 复制状态（从库）
psql -c "SELECT * FROM pg_stat_wal_receiver;"

# 6. 查看日志
tail -f /var/log/postgresql/postgresql-14-main.log
```

### Windows PowerShell

```powershell
# 1. 连接检查
psql -h localhost -U postgres -c "SELECT version();"

# 2. 数据库状态
psql -c "SELECT datname, pg_size_pretty(pg_database_size(datname)) FROM pg_database;"

# 3. 活动连接
psql -c "SELECT * FROM pg_stat_activity WHERE state != 'idle';"

# 4. 锁情况
psql -c "SELECT * FROM pg_locks WHERE NOT granted;"

# 5. 复制状态（从库）
psql -c "SELECT * FROM pg_stat_wal_receiver;"

# 6. 查看日志 (Windows 默认路径)
Get-Content "C:\Program Files\PostgreSQL\14\data\log\postgresql-*.log" -Wait

# 7. 检查 PostgreSQL 服务
Get-Service | Where-Object {$_.Name -like "*postgres*"}

# 8. 重启 PostgreSQL 服务
Restart-Service postgresql-x64-14

# 9. 查看 PostgreSQL 进程
Get-Process | Where-Object {$_.ProcessName -like "*postgres*"}

# 10. 检查端口监听
Get-NetTCPConnection -LocalPort 5432

# 11. 检查数据目录磁盘空间
Get-ChildItem "C:\Program Files\PostgreSQL\14\data" | Measure-Object -Property Length -Sum

# 12. 查看 Windows 事件日志
Get-EventLog -LogName Application -Source "PostgreSQL*" -Newest 20
```

## 常见故障处理

### 1. 连接数耗尽
```sql
-- 查看连接数
SELECT count(*) FROM pg_stat_activity;
SELECT setting FROM pg_settings WHERE name = 'max_connections';

-- 查看连接详情
SELECT pid, usename, application_name, state, query_start, query
FROM pg_stat_activity
WHERE state != 'idle'
ORDER BY query_start;

-- 终止长时间运行的查询
SELECT pg_terminate_backend(pid)
FROM pg_stat_activity
WHERE state != 'idle'
AND query_start < NOW() - INTERVAL '1 hour';
```

### 2. 流复制延迟
```sql
-- 在主库查看
SELECT client_addr, state, sent_lsn, write_lsn, flush_lsn, replay_lsn
FROM pg_stat_replication;

-- 在从库查看
SELECT * FROM pg_stat_wal_receiver;

-- 查看复制延迟
SELECT
  CASE WHEN pg_last_wal_receive_lsn() = pg_last_wal_replay_lsn()
    THEN 0
    ELSE EXTRACT(EPOCH FROM (now() - pg_last_xact_replay_timestamp()))
  END AS lag_seconds;
```

### 3. 锁等待和死锁
```sql
-- 查看锁等待
SELECT blocked_locks.pid AS blocked_pid,
       blocked_activity.usename AS blocked_user,
       blocking_locks.pid AS blocking_pid,
       blocking_activity.usename AS blocking_user
FROM pg_catalog.pg_locks blocked_locks
JOIN pg_catalog.pg_stat_activity blocked_activity ON blocked_activity.pid = blocked_locks.pid
JOIN pg_catalog.pg_locks blocking_locks ON blocking_locks.locktype = blocked_locks.locktype
JOIN pg_catalog.pg_stat_activity blocking_activity ON blocking_activity.pid = blocking_locks.pid
WHERE NOT blocked_locks.granted;

-- 查看死锁
SHOW log_lock_waits;
```

## 性能优化配置

```ini
# postgresql.conf
# 内存配置
shared_buffers = 4GB
effective_cache_size = 12GB
work_mem = 128MB
maintenance_work_mem = 1GB

# WAL 配置
wal_buffers = 16MB
min_wal_size = 1GB
max_wal_size = 4GB
checkpoint_completion_target = 0.9

# 并发配置
max_connections = 200
max_worker_processes = 8
max_parallel_workers_per_gather = 4
max_parallel_workers = 8

# 日志配置
log_destination = 'stderr'
logging_collector = on
log_directory = 'log'
log_filename = 'postgresql-%Y-%m-%d_%H%M%S.log'
log_min_duration_statement = 1000
log_line_prefix = '%t [%p]: [%l-1] user=%u,db=%d,app=%a,client=%h '
```

## Windows PowerShell 运维脚本

### 服务管理
```powershell
# 检查 PostgreSQL 服务状态
Get-Service -Name "postgresql*"

# 启动 PostgreSQL 服务
Start-Service -Name "postgresql-x64-14"

# 停止 PostgreSQL 服务
Stop-Service -Name "postgresql-x64-14"

# 重启 PostgreSQL 服务
Restart-Service -Name "postgresql-x64-14"

# 设置服务自动启动
Set-Service -Name "postgresql-x64-14" -StartupType Automatic
```

### 日志监控
```powershell
# 实时监控 PostgreSQL 日志
Get-Content "C:\Program Files\PostgreSQL\14\data\log\postgresql-*.log" -Wait

# 查找错误日志
Select-String -Path "C:\Program Files\PostgreSQL\14\data\log\*.log" -Pattern "ERROR|FATAL|PANIC" -Context 2

# 查看最近 100 条日志
Get-Content "C:\Program Files\PostgreSQL\14\data\log\postgresql-*.log" -Tail 100
```

### 性能监控
```powershell
# 检查 PostgreSQL 进程资源使用
Get-Process | Where-Object {$_.ProcessName -like "*postgres*"} |
    Select-Object Name, Id, CPU, WorkingSet, PagedMemorySize

# 检查端口连接
Get-NetTCPConnection -LocalPort 5432 | Measure-Object

# 检查磁盘空间 (PostgreSQL 数据目录)
Get-Volume | Where-Object {$_.DriveLetter -eq 'C'} |
    Select-Object DriveLetter, Size, SizeRemaining, @{N='Used%';E={[math]::Round(($_.Size-$_.SizeRemaining)/$_.Size*100,2)}}
```

### 备份恢复 (Windows)
```powershell
# 使用 pg_dump 备份
pg_dump -h localhost -U postgres -d mydb -Fc -f "C:\Backups\mydb.dump"

# 使用 pg_restore 恢复
pg_restore -h localhost -U postgres -d mydb "C:\Backups\mydb.dump"

# 备份所有数据库
pg_dumpall -h localhost -U postgres -f "C:\Backups\all_databases.sql"

# 压缩备份文件
Compress-Archive -Path "C:\Backups\mydb.dump" -DestinationPath "C:\Backups\mydb-$(Get-Date -Format 'yyyyMMdd').zip"
```

## 备份恢复

```bash
# pg_dump 逻辑备份
pg_dump -h localhost -U postgres -d mydb -Fc -f mydb.dump

# pg_restore 恢复
pg_restore -h localhost -U postgres -d mydb mydb.dump

# pg_basebackup 物理备份
pg_basebackup -h localhost -U replicator -D /backup/pg_backup -Fp -Xs -P -v

# PITR 恢复
# 1. 恢复基础备份
# 2. 配置 recovery.conf
restore_command = 'cp /archive/%f %p'
recovery_target_time = '2024-01-01 12:00:00'
```

## MCP 工具支持

本 Skill 可通过 MCP (Model Context Protocol) 提供以下工具：

### 工具列表

| 工具名称 | 描述 | 必需参数 |
|---------|------|---------|
| `pg_check_connection` | 检查 PostgreSQL 连接 | host, port, user |
| `pg_get_activity` | 获取活动连接和查询 | host, port, user |
| `pg_get_replication_status` | 检查流复制状态 | host, port, user |
| `pg_get_locks` | 查看锁等待情况 | host, port, user |
| `pg_get_database_sizes` | 获取数据库大小 | host, port, user |

### 工具定义示例

```json
{
  "name": "pg_check_connection",
  "description": "检查 PostgreSQL 连接和基本状态",
  "inputSchema": {
    "type": "object",
    "properties": {
      "host": { "type": "string", "default": "localhost" },
      "port": { "type": "integer", "default": 5432 },
      "user": { "type": "string", "default": "postgres" },
      "password": { "type": "string" },
      "database": { "type": "string", "default": "postgres" }
    }
  }
}
```

```json
{
  "name": "pg_get_activity",
  "description": "获取当前活动连接和正在执行的查询",
  "inputSchema": {
    "type": "object",
    "properties": {
      "host": { "type": "string", "default": "localhost" },
      "port": { "type": "integer", "default": 5432 },
      "user": { "type": "string", "default": "postgres" },
      "password": { "type": "string" },
      "database": { "type": "string", "default": "postgres" },
      "active_only": { "type": "boolean", "default": true }
    }
  }
}
```

```json
{
  "name": "pg_get_replication_status",
  "description": "检查 PostgreSQL 流复制状态（主库和从库）",
  "inputSchema": {
    "type": "object",
    "properties": {
      "host": { "type": "string", "default": "localhost" },
      "port": { "type": "integer", "default": 5432 },
      "user": { "type": "string", "default": "postgres" },
      "password": { "type": "string" },
      "database": { "type": "string", "default": "postgres" }
    }
  }
}
```

```json
{
  "name": "pg_get_locks",
  "description": "查看当前锁等待和阻塞情况",
  "inputSchema": {
    "type": "object",
    "properties": {
      "host": { "type": "string", "default": "localhost" },
      "port": { "type": "integer", "default": 5432 },
      "user": { "type": "string", "default": "postgres" },
      "password": { "type": "string" },
      "database": { "type": "string", "default": "postgres" }
    }
  }
}
```

```json
{
  "name": "pg_get_database_sizes",
  "description": "获取所有数据库的大小统计",
  "inputSchema": {
    "type": "object",
    "properties": {
      "host": { "type": "string", "default": "localhost" },
      "port": { "type": "integer", "default": 5432 },
      "user": { "type": "string", "default": "postgres" },
      "password": { "type": "string" },
      "database": { "type": "string", "default": "postgres" },
      "limit": { "type": "integer", "default": 20 }
    }
  }
}
```

### Python MCP Server 示例

```python
from mcp.server import Server
from mcp.types import TextContent
import subprocess

app = Server("postgres-ops")

def build_psql_cmd(host, port, user, password, database, query):
    env = f"PGPASSWORD='{password}' " if password else ""
    return f"{env}psql -h {host} -p {port} -U {user} -d {database} -c \"{query}\""

@app.call_tool()
def call_tool(name: str, arguments: dict):
    host = arguments.get("host", "localhost")
    port = arguments.get("port", 5432)
    user = arguments.get("user", "postgres")
    password = arguments.get("password", "")
    database = arguments.get("database", "postgres")

    if name == "pg_check_connection":
        query = "SELECT version(), now(), pg_database_size(current_database());"
        cmd = build_psql_cmd(host, port, user, password, database, query)
        result = subprocess.run(cmd, shell=True, capture_output=True, text=True)
        return [TextContent(type="text", text=result.stdout or result.stderr)]

    elif name == "pg_get_activity":
        active = "WHERE state != 'idle'" if arguments.get("active_only", True) else ""
        query = f"SELECT pid, usename, application_name, client_addr, state, query_start, query FROM pg_stat_activity {active} ORDER BY query_start;"
        cmd = build_psql_cmd(host, port, user, password, database, query)
        result = subprocess.run(cmd, shell=True, capture_output=True, text=True)
        return [TextContent(type="text", text=result.stdout)]

    elif name == "pg_get_replication_status":
        query = "SELECT * FROM pg_stat_replication;"
        cmd = build_psql_cmd(host, port, user, password, database, query)
        result = subprocess.run(cmd, shell=True, capture_output=True, text=True)
        # 如果从库，也检查 wal_receiver
        query2 = "SELECT * FROM pg_stat_wal_receiver;"
        cmd2 = build_psql_cmd(host, port, user, password, database, query2)
        result2 = subprocess.run(cmd2, shell=True, capture_output=True, text=True)
        return [TextContent(type="text", text=f"=== 复制发送端 ===\n{result.stdout}\n=== 复制接收端 ===\n{result2.stdout}")]

    elif name == "pg_get_locks":
        query = """SELECT blocked_locks.pid AS blocked_pid, blocked_activity.usename AS blocked_user,
            blocking_locks.pid AS blocking_pid, blocking_activity.usename AS blocking_user,
            blocked_activity.query AS blocked_statement, blocking_activity.query AS blocking_statement
            FROM pg_catalog.pg_locks blocked_locks
            JOIN pg_catalog.pg_stat_activity blocked_activity ON blocked_activity.pid = blocked_locks.pid
            JOIN pg_catalog.pg_locks blocking_locks ON blocking_locks.locktype = blocked_locks.locktype
            JOIN pg_catalog.pg_stat_activity blocking_activity ON blocking_activity.pid = blocking_locks.pid
            WHERE NOT blocked_locks.granted;"""
        cmd = build_psql_cmd(host, port, user, password, database, query)
        result = subprocess.run(cmd, shell=True, capture_output=True, text=True)
        return [TextContent(type="text", text=result.stdout)]

    elif name == "pg_get_database_sizes":
        limit = arguments.get("limit", 20)
        query = f"SELECT datname, pg_size_pretty(pg_database_size(datname)) as size FROM pg_database ORDER BY pg_database_size(datname) DESC LIMIT {limit};"
        cmd = build_psql_cmd(host, port, user, password, database, query)
        result = subprocess.run(cmd, shell=True, capture_output=True, text=True)
        return [TextContent(type="text", text=result.stdout)]

if __name__ == "__main__":
    app.run()
```

## 输出规范

```
🐘 PostgreSQL 诊断报告

📊 基本信息
- 版本：[version]
- 运行时间：[pg_postmaster_start_time]
- 数据库大小：[pg_database_size]
- 连接数：[active connections]

🔍 问题发现
1. [问题描述]

💡 解决方案
[处理步骤]
```
