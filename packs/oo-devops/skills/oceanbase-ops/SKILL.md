---
name: oceanbase-ops
description: OceanBase 分布式数据库运维指南 - 集群部署、租户管理、性能调优、日志分析、备份恢复、集群扩容、故障处理。用于 OceanBase 数据库的日常运维管理。
mode: on_demand
version: 1.1.0
maturity: experimental
---

# OceanBase 运维

> **官网**: https://www.oceanbase.com/
> **文档**: https://www.oceanbase.com/docs

## 集群部署

### 系统要求
- Linux CentOS 7.9+ / Ubuntu 18.04+
- CPU ≥ 4 核
- 内存 ≥ 16GB
- SSD 磁盘（推荐）

### OBD 部署方式
```bash
# 安装 OBD (OceanBase Deployer)
yum install -y yum-utils
yum-config-manager --add-repo https://mirrors.aliyun.com/oceanbase/OceanBase.repo
yum install -y ob-deploy

# 初始化配置
echo 'export PATH=$PATH:/usr/obd/bin' >> ~/.bashrc
source ~/.bashrc

# 创建集群配置
obd demo
```

### 手动部署配置
```yaml
# obcluster.yaml
user:
  username: admin
  password: your_password
  key_file: /home/admin/.ssh/id_rsa

oceanbase-ce:
  servers:
    - name: server1
      ip: 192.168.1.10
    - name: server2
      ip: 192.168.1.11
    - name: server3
      ip: 192.168.1.12
  global:
    devname: eth0
    mysql_port: 2881
    rpc_port: 2882
    data_dir: /data/observer
    redo_dir: /redo/observer
    zone: zone1
  server1:
    zone: zone1
  server2:
    zone: zone2
  server3:
    zone: zone3
```

### 启动集群
```bash
# 部署集群
obd cluster deploy obcluster -c obcluster.yaml

# 启动集群
obd cluster start obcluster

# 查看状态
obd cluster display obcluster
```

---

## 租户管理

### 创建租户
```sql
-- 创建资源单元
CREATE RESOURCE UNIT S1_unit_config
  MEMORY_SIZE = '2G',
  MAX_CPU = 1,
  MIN_CPU = 1,
  IOPS = 10000,
  LOG_DISK_SIZE = '10G',
  MAX_IOPS = 10000,
  MIN_IOPS = 10000;

-- 创建资源池
CREATE RESOURCE POOL S1_pool
  UNIT = 'S1_unit_config',
  UNIT_NUM = 1;

-- 创建租户
CREATE TENANT IF NOT EXISTS S1
  PRIMARY_ZONE = 'zone1',
  RESOURCE_POOL_LIST = ('S1_pool'),
  CHARSET = 'utf8mb4'
  SET ob_tcp_invited_nodes = '%';
```

### 管理租户
```sql
-- 查看所有租户
SELECT * FROM oceanbase.DBA_OB_TENANTS;

-- 查看租户资源使用
SELECT * FROM oceanbase.GV$OB_UNITS WHERE TENANT_ID = 1001;

-- 修改租户资源
ALTER TENANT S1 SET VARIABLES ob_tcp_invited_nodes = '%';

-- 锁定/解锁租户
ALTER TENANT S1 LOCK;
ALTER TENANT S1 UNLOCK;
```

### 租户删除
```sql
-- 删除租户（谨慎操作）
DROP TENANT S1;

-- 强制删除
DROP TENANT S1 FORCE;
```

---

## 性能调优

### SQL 调优
```sql
-- 查看执行计划
EXPLAIN SELECT * FROM t1 WHERE id = 1;

-- 查看慢查询
SELECT * FROM oceanbase.GV$OB_SQL_AUDIT
WHERE elapsed_time > 1000000
ORDER BY elapsed_time DESC
LIMIT 10;

-- 绑定执行计划
CREATE OUTLINE ol_name ON sql_id USING HINT /*+ INDEX(t1 idx_id) */;
```

### 表优化
```sql
-- 分析表
ANALYZE TABLE table_name;

-- 查看表统计信息
SELECT * FROM oceanbase.DBA_TAB_STATISTICS WHERE table_name = 'TABLE_NAME';

-- 创建索引
CREATE INDEX idx_name ON table_name(column_name);

-- 查看表分布
SELECT * FROM oceanbase.DBA_OB_TABLE_LOCATIONS WHERE table_name = 'TABLE_NAME';
```

### 参数调优
```sql
-- 修改系统变量
ALTER SYSTEM SET memory_limit = '32G';
ALTER SYSTEM SET cpu_count = 16;
ALTER SYSTEM SET cache_wash_threshold = '4G';

-- 租户级别参数
ALTER TENANT S1 SET ob_query_timeout = 10000000;
ALTER TENANT S1 SET ob_trx_timeout = 100000000;
```

---

## 日志分析

### 日志目录
```bash
# 日志路径
/data/observer/log/
├── observer.log      # 主日志
├── election.log      # 选举日志
├── rootservice.log   # RootService 日志
├── trace.log         # 追踪日志
└── clog              # 事务日志
```

### 常用日志查询
```bash
# 查看错误日志
grep "ERROR" /data/observer/log/observer.log

# 查看慢查询日志
grep "slow" /data/observer/log/observer.log

# 查看选举日志
tail -f /data/observer/log/election.log
```

### SQL 诊断
```sql
-- 查看诊断事件
SELECT * FROM oceanbase.GV$OB_DIAGNOSTIC_EVENTS
ORDER BY timestamp DESC
LIMIT 50;

-- 查看等待事件
SELECT event, count(*), avg_wait_time
FROM oceanbase.GV$OB_SYSTEM_EVENT
GROUP BY event
ORDER BY count(*) DESC;
```

---

## 备份恢复

### 物理备份
```sql
-- 配置备份路径
ALTER SYSTEM SET data_backup_dest = 'file:///backup/data';
ALTER SYSTEM SET log_archive_dest = 'file:///backup/archivelog';

-- 开启日志归档
ALTER SYSTEM ARCHIVELOG;

-- 执行全量备份
ALTER SYSTEM BACKUP DATABASE;
```

### 逻辑备份
```bash
# 使用 obdumper 导出
obdumper -h 127.0.0.1 -P 2881 -u root -p password -D dbname \
  -f /backup/export \
  --ddl \
  --data

# 导出指定表
obdumper -h 127.0.0.1 -P 2881 -u root -p password -D dbname \
  -t table1,table2 \
  -f /backup/export
```

### 数据恢复
```sql
-- 查看备份集
SELECT * FROM oceanbase.CDB_OB_BACKUP_SET_DETAILS;

-- 恢复到指定时间点
ALTER SYSTEM RESTORE dest_tenant_name = 'new_tenant' backup_dest = 'file:///backup/data';

-- 取消恢复
ALTER SYSTEM CANCEL RESTORE;
```

### 逻辑导入
```bash
# 使用 obloader 导入
obloader -h 127.0.0.1 -P 2881 -u root -p password -D dbname \
  -f /backup/export \
  --ddl \
  --data
```

---

## 集群扩容

### 添加节点
```bash
# 编辑配置文件，添加新节点
vi obcluster.yaml
# 在 servers 列表中添加新节点 IP

# 扩容集群
obd cluster add obcluster -c obcluster.yaml --server server4

# 启动新节点
obd cluster start obcluster -s server4
```

### 扩容租户资源
```sql
-- 修改资源单元规格
ALTER RESOURCE UNIT S1_unit_config
  MEMORY_SIZE = '4G',
  MAX_CPU = 2,
  MIN_CPU = 2;

-- 扩容资源池单元数
ALTER RESOURCE POOL S1_pool UNIT_NUM = 2;
```

### 副本管理
```sql
-- 查看副本分布
SELECT * FROM oceanbase.DBA_OB_TABLE_LOCATIONS WHERE tenant_id = 1001;

-- 修改副本数
ALTER TENANT S1 SET PRIMARY_ZONE = 'zone1,zone2,zone3';

-- 本地化表
ALTER TABLE table_name LOCALITY = 'F@zone1';
```

---

## 故障处理

### 节点故障
```bash
# 查看节点状态
obd cluster display obcluster

# 重启故障节点
obd cluster restart obcluster -s server1

# 替换故障节点
obd cluster stop obcluster -s server1
obd cluster deploy obcluster -c obcluster.yaml --server server1
obd cluster start obcluster -s server1
```

### 数据修复
```sql
-- 检查集群健康状态
SELECT * FROM oceanbase.DBA_OB_ZONE;
SELECT * FROM oceanbase.DBA_OB_SERVERS;

-- 查看副本状态
SELECT * FROM oceanbase.DBA_OB_TABLE_LOCATIONS WHERE replica_type = 'FULL';

-- 手动切换 Leader
ALTER SYSTEM SWITCH REPLICA LEADER ls = 1001 SERVER = '192.168.1.10:2882';
```

### 性能问题排查
```sql
-- 查看会话
SELECT * FROM oceanbase.GV$OB_PROCESSLIST;

-- 终止会话
KILL 123456;

-- 查看锁等待
SELECT * FROM oceanbase.GV$OB_LOCKS WHERE block = 1;

-- 查看事务
SELECT * FROM oceanbase.GV$OB_TRANSACTION_PARTICIPANTS;
```

---

## 常用命令速查

| 操作 | 命令 |
|------|------|
| 部署集群 | `obd cluster deploy <name> -c <yaml>` |
| 启动集群 | `obd cluster start <name>` |
| 停止集群 | `obd cluster stop <name>` |
| 查看状态 | `obd cluster display <name>` |
| 连接租户 | `mysql -h127.0.0.1 -P2881 -uroot@S1 -p` |
| 查看租户 | `SELECT * FROM oceanbase.DBA_OB_TENANTS;` |
| 备份数据 | `ALTER SYSTEM BACKUP DATABASE;` |
| 查看日志 | `tail -f /data/observer/log/observer.log` |

---

## 最佳实践

- 部署时确保时钟同步（NTP）
- 配置独立的 redo 和 data 磁盘
- 定期执行数据备份和归档
- 监控 QPS、TPS、响应时间等核心指标
- 使用 OCP (OceanBase Cloud Platform) 进行可视化运维
- 大表提前规划分区策略
- 定期分析表和收集统计信息
