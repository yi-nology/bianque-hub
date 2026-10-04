---
name: opengauss-ops
description: openGauss 数据库运维指南 - 安装部署、集群启停、状态检查、性能监控、日志清理、备份恢复、参数调优。用于 openGauss 数据库的日常运维管理。
mode: on_demand
version: 1.1.0
maturity: experimental
---

# openGauss 运维

> **官网**: https://opengauss.org/

## 安装部署

### 系统要求
- CentOS 7.6+ / openEuler 20.03 LTS+
- 内存 ≥ 8GB
- 磁盘 ≥ 100GB

### 一键安装脚本
```bash
# 下载安装包
wget https://opengauss.obs.cn-south-1.myhuaweicloud.com/5.0.0/x86/openGauss-5.0.0-CentOS-64bit-all.tar.gz

# 创建用户
groupadd dbgroup
useradd -g dbgroup omm

# 解压安装
tar -zxvf openGauss-5.0.0-CentOS-64bit-all.tar.gz
cd openGauss-5.0.0-CentOS-64bit-all

# 执行安装
./install.sh -w "YourPassword123" --multinode
```

### 配置文件示例
```xml
<!-- /opt/opengauss/data/postgresql.conf -->
listen_addresses = '*'
port = 5432
max_connections = 100
shared_buffers = 1GB
effective_cache_size = 3GB
work_mem = 16MB
maintenance_work_mem = 256MB
```

---

## 集群启停

### 启动集群
```bash
# 切换到 omm 用户
su - omm

# 启动集群
gs_om -t start

# 或启动单个节点
gs_ctl start -D /opt/opengauss/data
```

### 停止集群
```bash
# 正常停止
gs_om -t stop

# 强制停止
gs_om -t stop -m fast

# 立即停止
gs_ctl stop -D /opt/opengauss/data -m immediate
```

### 重启集群
```bash
gs_om -t restart
```

---

## 状态检查

### 集群状态
```bash
# 查看集群整体状态
gs_om -t status

# 查看详细状态
gs_om -t status --detail

# 查看节点状态
gs_check -i CheckClusterState
```

### 数据库状态
```sql
-- 查看数据库运行状态
SELECT * FROM pg_stat_activity WHERE state = 'active';

-- 查看连接数
SELECT count(*) FROM pg_stat_activity;

-- 查看数据库大小
SELECT datname, pg_size_pretty(pg_database_size(datid)) as size
FROM pg_stat_database;
```

### 复制状态（主备）
```sql
-- 查看复制延迟
SELECT * FROM pg_stat_replication;

-- 查看复制槽
SELECT * FROM pg_replication_slots;
```

---

## 性能监控

### 系统监控
```bash
# CPU 和内存
gs_checkperf -i CheckCPU

# IO 性能
gs_checkperf -i CheckIO

# 网络性能
gs_checkperf -i CheckNetwork
```

### SQL 性能
```sql
-- 慢查询分析
SELECT * FROM pg_stat_statements
ORDER BY total_time DESC
LIMIT 10;

-- 锁等待
SELECT * FROM pg_locks WHERE NOT granted;

-- 表统计信息
SELECT schemaname, relname, n_live_tup, n_dead_tup
FROM pg_stat_user_tables
ORDER BY n_live_tup DESC;
```

### WDR 报告生成
```sql
-- 创建快照
SELECT create_wdr_snapshot();

-- 生成 WDR 报告（指定快照 ID）
SELECT generate_wdr_report(1, 2, 'local');
```

---

## 日志清理

### 日志目录
- 运行日志：`/var/log/opengauss/omm/`
- 审计日志：`/var/log/opengauss/pg_audit/`
- 归档日志：`$GAUSSHOME/arch`

### 自动清理配置
```bash
# 清理 7 天前的日志
find /var/log/opengauss -name "*.log" -mtime +7 -delete

# 清理审计日志
gs_guc reload -D /opt/opengauss/data -c "audit_directory='/var/log/opengauss/pg_audit'"
```

### 日志轮转
```bash
# 手动切换日志
gs_ctl switchlog -D /opt/opengauss/data
```

---

## 备份恢复

### 物理备份
```bash
# 全量备份
gs_basebackup -D /backup/full -Fp -Xstream -P -v

# 增量备份（基于时间线）
gs_basebackup -D /backup/incr -Fp -Xstream -P -v --incremental
```

### 逻辑备份
```bash
# 导出整个数据库
gs_dump -U omm -W password dbname -f /backup/dbname.sql

# 导出指定表
gs_dump -U omm -W password dbname -t table_name -f /backup/table.sql

# 并行导出
gs_dump -U omm -W password dbname -Fd -f /backup/dbname_dir -j 4
```

### 恢复
```bash
# 物理恢复
# 1. 停止数据库
gs_ctl stop -D /opt/opengauss/data

# 2. 清理数据目录
rm -rf /opt/opengauss/data/*

# 3. 恢复备份
cp -r /backup/full/* /opt/opengauss/data/

# 4. 启动数据库
gs_ctl start -D /opt/opengauss/data

# 逻辑恢复
gsql -U omm -d dbname -f /backup/dbname.sql
```

### 时间点恢复 (PITR)
```bash
# 配置归档
recovery_target_time = '2024-03-25 12:00:00'
recovery_target_timeline = 'latest'

# 启动恢复模式
gs_ctl start -D /opt/opengauss/data -m standby
```

---

## 参数调优

### 内存参数
```sql
-- 共享缓冲区（推荐物理内存 25%）
ALTER SYSTEM SET shared_buffers = '2GB';

-- 有效缓存大小（推荐物理内存 75%）
ALTER SYSTEM SET effective_cache_size = '6GB';

-- 工作内存
ALTER SYSTEM SET work_mem = '32MB';

-- 维护工作内存
ALTER SYSTEM SET maintenance_work_mem = '512MB';
```

### 并发参数
```sql
-- 最大连接数
ALTER SYSTEM SET max_connections = 200;

-- 并行工作者
ALTER SYSTEM SET max_parallel_workers = 8;
ALTER SYSTEM SET max_parallel_workers_per_gather = 4;
```

### 日志参数
```sql
-- 慢查询阈值（毫秒）
ALTER SYSTEM SET log_min_duration_statement = 1000;

-- 记录锁等待
ALTER SYSTEM SET log_lock_waits = on;
```

### 应用配置
```bash
# 重载配置
gs_ctl reload -D /opt/opengauss/data

# 或 SQL 方式
SELECT pg_reload_conf();
```

---

## 常用命令速查

| 操作 | 命令 |
|------|------|
| 启动 | `gs_om -t start` |
| 停止 | `gs_om -t stop` |
| 状态 | `gs_om -t status` |
| 备份 | `gs_basebackup -D /backup -Fp` |
| 导出 | `gs_dump -U omm dbname -f backup.sql` |
| 连接 | `gsql -U omm -d postgres -p 5432` |
| 查看配置 | `SHOW shared_buffers;` |
| 修改配置 | `ALTER SYSTEM SET param = value;` |

---

## 最佳实践

- 定期执行 `gs_check` 进行健康检查
- 启用 `pg_stat_statements` 跟踪慢查询
- 配置合适的 `shared_buffers`（物理内存 25%）
- 使用 WDR 报告进行性能分析
- 定期清理审计和日志文件
- 主备环境确保同步延迟 < 1 秒
