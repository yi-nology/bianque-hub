---
name: tidb-ops
description: TiDB 分布式数据库运维指南 - TiUP/TiDB Operator 集群部署、扩缩容、备份恢复、数据迁移、日常监控、故障定位。用于 TiDB 数据库的日常运维管理。
mode: on_demand
version: 1.1.0
maturity: experimental
---

# TiDB 运维

> **官网**: https://pingcap.com/
> **文档**: https://docs.pingcap.com/

## 集群部署

### TiUP 部署

#### 安装 TiUP
```bash
# 安装 TiUP
curl --proto '=https' --tlsv1.2 -sSf https://tiup-mirrors.pingcap.com/install.sh | sh

# 配置环境变量
source ~/.bashrc

# 安装集群组件
tiup cluster
```

#### 初始化集群拓扑
```yaml
# topology.yaml
global:
  user: "tidb"
  ssh_port: 22
  deploy_dir: "/tidb-deploy"
  data_dir: "/tidb-data"

server_configs:
  tidb:
    log.slow-threshold: 300
  tikv:
    readpool.storage.use-unified-pool: false
    readpool.coprocessor.use-unified-pool: true
  pd:
    replication.enable-placement-rules: true
    replication.location-labels: ["zone", "dc", "rack", "host"]

pd_servers:
  - host: 198.51.100.1
  - host: 198.51.100.2
  - host: 198.51.100.3

tidb_servers:
  - host: 198.51.100.1
  - host: 198.51.100.2

tikv_servers:
  - host: 198.51.100.4
  - host: 198.51.100.5
  - host: 198.51.100.6

monitoring_servers:
  - host: 198.51.100.7

grafana_servers:
  - host: 198.51.100.7

alertmanager_servers:
  - host: 198.51.100.7
```

#### 部署集群
```bash
# 检查集群配置
tiup cluster check ./topology.yaml --user root

# 自动修复配置
tiup cluster check ./topology.yaml --apply --user root

# 部署集群
tiup cluster deploy tidb-cluster v8.0.0 ./topology.yaml --user root

# 启动集群
tiup cluster start tidb-cluster

# 查看集群状态
tiup cluster display tidb-cluster
```

### TiDB Operator 部署 (Kubernetes)

#### 安装 CRD
```bash
# 添加 Helm Repo
helm repo add pingcap https://charts.pingcap.org/
helm repo update

# 安装 TiDB Operator
kubectl create namespace tidb-admin
helm install --namespace tidb-admin tidb-operator pingcap/tidb-operator --version v1.6.0
```

#### 部署 TiDB 集群
```yaml
# tidb-cluster.yaml
apiVersion: pingcap.com/v1alpha1
kind: TidbCluster
metadata:
  name: basic
  namespace: tidb-cluster
spec:
  version: "v8.0.0"
  timezone: "Asia/Shanghai"
  configUpdateStrategy: RollingUpdate
  pvReclaimPolicy: Retain
  enableMonitor: true
  helper:
    image: busybox:1.34.1
  pd:
    baseImage: pingcap/pd
    replicas: 3
    requests:
      storage: "10Gi"
    config: |
      [replication]
        enable-placement-rules = true
  tikv:
    baseImage: pingcap/tikv
    replicas: 3
    requests:
      storage: "100Gi"
    config: |
      [storage]
        reserve-space = "0"
  tidb:
    baseImage: pingcap/tidb
    replicas: 2
    service:
      type: ClusterIP
      exposeStatus: true
    config: |
      [log]
        slow-threshold = 300
```

```bash
# 创建命名空间并部署
kubectl create namespace tidb-cluster
kubectl apply -f tidb-cluster.yaml

# 查看 Pod 状态
kubectl get pods -n tidb-cluster -o wide
```

---

## 扩缩容

### TiUP 扩缩容

#### 扩容 TiKV
```bash
# 编辑 scale-out.yaml
vi scale-out.yaml
```

```yaml
tikv_servers:
  - host: 198.51.100.8
  - host: 198.51.100.9
```

```bash
# 执行扩容
tiup cluster scale-out tidb-cluster scale-out.yaml

# 检查扩容进度
tiup cluster display tidb-cluster
```

#### 缩容 TiKV
```bash
# 缩容前确保数据已迁移完毕
tiup cluster scale-in tidb-cluster -N 198.51.100.8

# 强制缩容（慎用）
tiup cluster scale-in tidb-cluster -N 198.51.100.8 --force
```

#### 扩容 TiDB
```yaml
tidb_servers:
  - host: 198.51.100.10
```

```bash
tiup cluster scale-out tidb-cluster scale-out.yaml
```

### TiDB Operator 扩缩容

```yaml
# 修改 tidb-cluster.yaml
spec:
  tikv:
    replicas: 5  # 从 3 扩容到 5
```

```bash
kubectl apply -f tidb-cluster.yaml

# 查看扩容进度
kubectl get pods -n tidb-cluster -w
```

---

## 备份恢复

### 使用 BR 工具

#### 全量备份
```bash
# 执行备份
tiup br backup full \
  --pd "198.51.100.1:2379" \
  --storage "s3://tidb-backup/full" \
  --s3.region "cn-north-1" \
  --ratelimit 120 \
  --log-file backup.log

# 本地存储备份
tiup br backup full \
  --pd "198.51.100.1:2379" \
  --storage "local:///backup/tidb/full" \
  --ratelimit 120
```

#### 增量备份
```bash
tiup br backup incremental \
  --pd "198.51.100.1:2379" \
  --storage "s3://tidb-backup/incr" \
  --lastbackupts $(cat backupts.txt) \
  --ratelimit 120
```

#### 恢复数据
```bash
# 全量恢复
tiup br restore full \
  --pd "198.51.100.1:2379" \
  --storage "s3://tidb-backup/full" \
  --ratelimit 128

# 指定库表恢复
tiup br restore db \
  --pd "198.51.100.1:2379" \
  --db "testdb" \
  --storage "s3://tidb-backup/full"
```

### TiDB Operator 备份

```yaml
# backup.yaml
apiVersion: pingcap.com/v1alpha1
kind: Backup
metadata:
  name: full-backup
  namespace: tidb-cluster
spec:
  cluster:
    name: basic
    namespace: tidb-cluster
  s3:
    provider: aws
    secretName: s3-secret
    region: cn-north-1
    bucket: tidb-backup
  storageClassName: standard
  storageSize: 100Gi
```

```bash
kubectl apply -f backup.yaml
```

---

## 数据迁移

### TiDB Lightning 导入

```toml
# tidb-lightning.toml
[lightning]
level = "info"
file = "tidb-lightning.log"
index-concurrency = 2
table-concurrency = 6

[tikv-importer]
backend = "local"
sorted-kv-dir = "/mnt/ssd/sorted-kv-dir"

[mydumper]
data-source-dir = "/data/export"

[tidb]
host = "198.51.100.1"
port = 4000
user = "root"
password = ""
status-port = 10080
pd-addr = "198.51.100.1:2379"
```

```bash
# 执行导入
tiup lightning -c tidb-lightning.toml
```

### TiDB Data Migration (DM)

#### 部署 DM 集群
```bash
# 创建拓扑文件
vi dm-topology.yaml
```

```yaml
global:
  user: "tidb"
  ssh_port: 22
  deploy_dir: "/dm-deploy"
  data_dir: "/dm-data"

master_servers:
  - host: 198.51.100.11

worker_servers:
  - host: 198.51.100.12
  - host: 198.51.100.13

monitoring_servers:
  - host: 198.51.100.14
```

```bash
tiup dm deploy dm-cluster v8.0.0 ./dm-topology.yaml --user root
tiup dm start dm-cluster
```

#### 配置迁移任务
```yaml
# task.yaml
name: "mysql-to-tidb"
task-mode: all

mysql-instances:
  - source-id: "mysql-01"
    block-allow-list: "bw-list"
    syncer-config-name: "syncer-01"

black-white-list:
  bw-list:
    do-dbs: ["db1", "db2"]

target-database:
  host: "198.51.100.1"
  port: 4000
  user: "root"
  password: ""
```

```bash
# 启动迁移任务
tiup dmctl --master-addr 198.51.100.11:8261 start-task task.yaml

# 查看任务状态
tiup dmctl --master-addr 198.51.100.11:8261 query-status
```

---

## 日常监控

### Grafana 监控

```bash
# 访问 Grafana
curl http://198.51.100.7:3000
# 默认账号: admin / admin
```

**核心监控指标**:
- **集群概览**: QPS、TPS、延迟、连接数
- **TiDB**: 解析耗时、编译耗时、执行耗时
- **TiKV**: CPU、内存、磁盘 IO、Region 分布
- **PD**: Leader 切换、Store 状态、调度状态

### SQL 诊断

```sql
-- 查看慢查询
SELECT * FROM information_schema.slow_query
WHERE time > '2024-03-01 00:00:00'
ORDER BY query_time DESC
LIMIT 10;

-- 查看正在执行的 SQL
SHOW PROCESSLIST;

-- 查看执行计划
EXPLAIN ANALYZE SELECT * FROM t WHERE id = 1;
```

### 系统表查询

```sql
-- 查看集群拓扑
SELECT * FROM information_schema.cluster_info;

-- 查看 Store 状态
SELECT * FROM information_schema.tikv_store_status;

-- 查看热点 Region
SELECT * FROM information_schema.tikv_region_status
ORDER BY written_bytes DESC
LIMIT 10;
```

---

## 故障定位

### 日志查看

```bash
# TiDB 日志
tiup cluster audit log tidb-cluster

# 查看指定节点日志
tiup cluster log tidb-cluster -N 198.51.100.1

# TiKV 日志
tail -f /tidb-deploy/tikv-20160/log/tikv.log
```

### 常见问题处理

#### Region 不平衡
```bash
# 查看 Store 分布
tiup ctl:v8.0.0 pd -u http://198.51.100.1:2379 store

# 手动调度
# 将 Region 1 的 Leader 迁移到 Store 2
tiup ctl:v8.0.0 pd -u http://198.51.100.1:2379 operator add transfer-leader 1 2
```

#### TiKV 节点故障
```bash
# 下线故障节点
tiup cluster scale-in tidb-cluster -N 198.51.100.4

# 等待数据迁移完成
tiup cluster display tidb-cluster

# 重新上线新节点
tiup cluster scale-out tidb-cluster scale-out.yaml
```

#### 慢查询处理
```sql
-- 捕获慢查询
SELECT digest, query, query_time, plan
FROM information_schema.statements_summary
WHERE query_time > 10
ORDER BY query_time DESC;

-- 添加索引优化
ALTER TABLE table_name ADD INDEX idx_column(column);
```

---

## 常用命令速查

| 操作 | 命令 |
|------|------|
| 部署集群 | `tiup cluster deploy <name> <version> <topology>` |
| 启动集群 | `tiup cluster start <name>` |
| 停止集群 | `tiup cluster stop <name>` |
| 查看状态 | `tiup cluster display <name>` |
| 连接数据库 | `mysql -h<host> -P4000 -uroot` |
| 扩容 | `tiup cluster scale-out <name> <scale-out.yaml>` |
| 缩容 | `tiup cluster scale-in <name> -N <node>` |
| 备份 | `tiup br backup full --pd <pd> --storage <path>` |
| 升级 | `tiup cluster upgrade <name> <version>` |

---

## 最佳实践

- 生产环境至少部署 3 个 TiKV、3 个 PD、2 个 TiDB
- TiKV 建议独占磁盘，避免与系统盘混用
- 启用 Prometheus + Grafana 监控
- 定期进行全量备份，增量备份频率根据业务需求
- 大表提前设计分区策略
- 使用 TiDB Dashboard 进行可视化诊断
- 慢查询阈值建议设置为 300ms
- 定期分析表 `ANALYZE TABLE table_name`
