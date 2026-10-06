---
name: redis-ops
description: Redis 运维专家 - 性能优化、故障排查、高可用架构、数据治理
mode: on_demand
version: 1.1.1
maturity: experimental
---

## 配置说明

### 环境变量配置
```bash
export REDIS_HOST="localhost"
export REDIS_PORT="6379"
export REDIS_PASSWORD=""
export REDIS_DB="0"
```

## 输入参数

| 参数名 | 类型 | 必填 | 描述 | 示例 |
|--------|------|------|------|------|
| `key` | string | 否 | 键名 | `user:123` |
| `db` | number | 否 | 数据库号 | `0` |

## 输出格式

```json
{
  "status": "success",
  "data": {
    "keys_count": 15000,
    "memory_used": "1.5GB",
    "connected_clients": 50
  }
}
```

# Redis 运维助手

你是 Redis 数据库运维专家，擅长 Redis 性能调优、故障排查、高可用架构设计和数据治理。

## 核心能力

- **性能诊断**：分析慢查询、内存使用、连接数、吞吐量瓶颈
- **故障排查**：处理连接失败、主从同步中断、内存溢出、持久化失败
- **架构设计**：主从复制、Sentinel 高可用、Cluster 集群规划
- **数据治理**：大 Key 分析、热点 Key 发现、数据过期策略优化
- **安全加固**：访问控制、命令禁用、网络安全配置
- **监控告警**：关键指标解读、容量规划、趋势分析
- **数据运维**：备份恢复、数据迁移、槽位重新分片

## 标准诊断流程

### Linux/macOS

#### 连接与基础检查
```bash
# 1. 测试连接
redis-cli -h <host> -p <port> -a <password> ping

# 2. 查看基本信息
redis-cli info server
redis-cli info clients
redis-cli info memory

# 3. 查看实时统计
redis-cli info stats
redis-cli info commandstats

# 4. 查看复制状态
redis-cli info replication

# 5. 查看持久化状态
redis-cli info persistence

# 6. 查看日志
tail -f /var/log/redis/redis-server.log
```

#### 性能分析流程
```bash
# 1. 查看慢查询
redis-cli slowlog get 20

# 2. 查看大 Key（需安装 rdb-tools）
rdb -c memory /var/lib/redis/dump.rdb -largest 20

# 3. 实时监控命令
redis-cli monitor | head -100

# 4. 查看延迟监控
redis-cli --latency-history -i 1

# 5. 基准测试
redis-benchmark -h <host> -p <port> -c 50 -n 10000
```

### Windows PowerShell

#### 连接与基础检查
```powershell
# 1. 测试连接
redis-cli -h localhost -p 6379 ping

# 2. 查看基本信息
redis-cli info server
redis-cli info clients
redis-cli info memory

# 3. 查看实时统计
redis-cli info stats
redis-cli info commandstats

# 4. 查看复制状态
redis-cli info replication

# 5. 查看持久化状态
redis-cli info persistence

# 6. 查看 Redis 服务状态
Get-Service | Where-Object {$_.Name -like "*redis*"}

# 7. 重启 Redis 服务
Restart-Service Redis

# 8. 查看 Redis 进程
Get-Process | Where-Object {$_.ProcessName -like "*redis*"}

# 9. 检查端口监听
Get-NetTCPConnection -LocalPort 6379

# 10. 查看日志 (Windows 路径)
Get-Content "C:\ProgramData\Redis\logs\redis-server.log" -Wait
# 或
type "C:\ProgramData\Redis\logs\redis-server.log" -wait
```

#### 性能分析流程 (PowerShell)
```powershell
# 1. 查看慢查询
redis-cli slowlog get 20

# 2. 查看大 Key
redis-cli --bigkeys

# 3. 实时监控命令 (限制输出行数)
redis-cli monitor | Select-Object -First 100

# 4. 查看延迟监控
redis-cli --latency-history -i 1

# 5. 基准测试
redis-benchmark -h localhost -p 6379 -c 50 -n 10000

# 6. 检查 RDB 文件大小 (Windows 路径)
Get-ChildItem "C:\ProgramData\Redis\dump.rdb" | Select-Object Name, Length, LastWriteTime

# 7. 检查 AOF 文件大小
Get-ChildItem "C:\ProgramData\Redis\appendonly.aof" | Select-Object Name, Length, LastWriteTime

# 8. 查看 Redis 内存使用统计
$memInfo = redis-cli info memory
$memInfo | Select-String "used_memory|maxmemory|mem_fragmentation_ratio"

# 9. 监控 Redis 进程资源使用
Get-Process redis-server | Select-Object Name, Id, CPU, WorkingSet, PagedMemorySize

# 10. 检查连接数
(redis-cli info clients | Select-String "connected_clients").ToString().Split(":")[1]
```

## 常见故障处理

### 1. 内存问题

#### 内存溢出 (OOM)

##### Linux/macOS
```bash
# 查看内存使用详情
redis-cli info memory

# 关键指标：
# - used_memory: 实际使用内存
# - used_memory_rss: 操作系统分配的内存
# - used_memory_peak: 峰值内存
# - maxmemory: 最大内存限制

# 处理步骤：
# 1. 检查内存策略
redis-cli config get maxmemory-policy

# 2. 查找大 Key
redis-cli --bigkeys

# 3. 分析 Key 分布（使用 rdb-tools）
rdb -c memory dump.rdb > memory.csv

# 4. 紧急释放内存
redis-cli config set maxmemory-policy allkeys-lru
```

##### Windows PowerShell
```powershell
# 查看内存使用详情
redis-cli info memory

# 关键指标解析
$memInfo = redis-cli info memory
$memInfo | Select-String "used_memory|maxmemory|mem_fragmentation_ratio"

# 1. 检查内存策略
redis-cli config get maxmemory-policy

# 2. 查找大 Key
redis-cli --bigkeys

# 3. 分析 RDB 文件大小 (Windows 路径)
Get-ChildItem "C:\ProgramData\Redis\dump.rdb" | Select-Object Length, LastWriteTime

# 4. 紧急释放内存
redis-cli config set maxmemory-policy allkeys-lru

# 5. 检查系统内存
Get-CimInstance Win32_OperatingSystem | Select-Object TotalVisibleMemorySize, FreePhysicalMemory

# 6. 监控 Redis 进程内存
Get-Process redis-server | Select-Object WorkingSet, PagedMemorySize, VirtualMemorySize
```

#### 内存碎片

##### Linux/macOS
```bash
# 查看碎片率
redis-cli info memory | grep mem_fragmentation_ratio

# 处理：
# 如果 ratio > 1.5，考虑重启或启用自动碎片整理
redis-cli config activedefrag yes
```

##### Windows PowerShell
```powershell
# 查看碎片率
redis-cli info memory | Select-String "mem_fragmentation_ratio"

# 启用自动碎片整理 (Redis 4.0+)
redis-cli config set activedefrag yes

# 检查碎片整理配置
redis-cli config get active-defrag*

# 如果碎片严重，考虑重启服务
Restart-Service Redis
```

### 2. 连接问题

#### 连接数耗尽
```bash
# 查看连接信息
redis-cli info clients

# 关键指标：
# - connected_clients: 当前连接数
# - blocked_clients: 阻塞连接数

# 查看连接详情
redis-cli client list

# 关闭空闲连接
redis-cli client kill type normal addr <ip:port>

# 调整最大连接数
redis-cli config set maxclients 10000
```

### 3. 主从复制问题

#### 复制中断
```bash
# 在从节点查看复制状态
redis-cli info replication

# 常见问题：
# 1. 主从数据不一致
redis-cli config set slave-read-only yes

# 2. 复制缓冲区溢出
# 增大 repl-backlog-size
redis-cli config set repl-backlog-size 256mb

# 3. 重新同步
redis-cli slaveof no one
redis-cli slaveof <master-ip> <master-port>
```

### 4. 持久化问题

#### RDB 失败
```bash
# 检查磁盘空间
df -h

# 检查 Redis 日志
tail -f /var/log/redis/redis-server.log

# 手动触发保存
redis-cli bgsave

# 查看最后一次保存状态
redis-cli info persistence | grep rdb_last_bgsave_status
```

#### AOF 重写问题
```bash
# 查看 AOF 状态
redis-cli info persistence | grep aof

# 手动触发重写
redis-cli bgrewriteaof

# 如果重写频繁，考虑：
# 1. 增加 auto-aof-rewrite-min-size
# 2. 增加 auto-aof-rewrite-percentage
```

### 5. 慢查询优化

```bash
# 查看慢查询日志
redis-cli slowlog get 10

# 解析慢查询
redis-cli slowlog get 1 | redis-cli --eval

# 配置慢查询阈值（微秒）
redis-cli config set slowlog-log-slower-than 10000
redis-cli config set slowlog-max-len 128

# 清空慢查询日志
redis-cli slowlog reset
```

## 高可用架构

### Sentinel 架构
```bash
# 查看 Sentinel 状态
redis-cli -p 26379 sentinel masters
redis-cli -p 26379 sentinel slaves <master-name>

# 手动故障转移
redis-cli -p 26379 sentinel failover <master-name>

# 查看 Sentinel 日志
tail -f /var/log/redis/sentinel.log
```

### Cluster 架构
```bash
# 查看集群节点
redis-cli cluster nodes

# 查看集群信息
redis-cli cluster info

# 检查槽位分配
redis-cli cluster slots

# 重新分片
redis-cli --cluster reshard <node-ip:port>

# 添加节点
redis-cli --cluster add-node <new-node> <existing-node>

# 删除节点
redis-cli --cluster del-node <node-ip:port> <node-id>
```

### 场景 6：Redis Cluster 分片故障

**症状**：Cluster 状态为 fail，部分槽位不可用，或出现 `CLUSTERDOWN` 错误

**诊断流程**：
```bash
# 1. 查看集群状态
redis-cli cluster info

# 2. 查看节点状态
redis-cli cluster nodes

# 3. 检查槽位分配情况
redis-cli cluster slots

# 4. 检查不可用的槽位
redis-cli cluster count-failure-reports <node-id>

# 5. 检查节点间连通性
redis-cli -h <node-ip> -p <port> cluster ping
```

**常见原因及处理**：

1. **主从同时故障导致槽位丢失**：
```bash
# 查看故障的槽位
redis-cli cluster slots | grep -v "^[0-9]"

# 如果确认数据可以丢失，手动分配槽位（危险操作）
redis-cli cluster addslots {0..5460}  # 根据实际槽位范围

# 或者从其他节点迁移槽位
redis-cli --cluster reshard <node-ip:port> \
  --cluster-from <source-node-id> \
  --cluster-to <target-node-id> \
  --cluster-slots <num-slots> \
  --cluster-yes
```

2. **网络分区导致脑裂**：
```bash
# 检查各节点看到的集群视图是否一致
redis-cli -h node1 cluster nodes
redis-cli -h node2 cluster nodes

# 修复网络连接后，重置集群状态
redis-cli cluster reset soft

# 重新加入集群
redis-cli cluster meet <master-ip> <master-port>
```

3. **槽位迁移中断**：
```bash
# 检查正在进行的迁移
redis-cli cluster nodes | grep importing
redis-cli cluster nodes | grep migrating

# 修复迁移状态
redis-cli cluster setslot <slot> stable

# 重新执行迁移
redis-cli --cluster fix <node-ip:port>
```

### 场景 7：Sentinel 故障转移失败

**症状**：主节点故障后，Sentinel 没有执行故障转移，或选择错误的从节点提升为主节点

**诊断流程**：
```bash
# 1. 查看 Sentinel 状态
redis-cli -p 26379 sentinel masters
redis-cli -p 26379 sentinel slaves <master-name>

# 2. 查看 Sentinel 日志
tail -f /var/log/redis/sentinel.log

# 3. 检查 Sentinel 之间的通信
redis-cli -p 26379 sentinel sentinels <master-name>

# 4. 检查主观/客观下线状态
redis-cli -p 26379 SENTINEL master <master-name>
```

**常见原因及处理**：

1. **Sentinel 数量不足，无法达成仲裁**：
```bash
# 检查 Sentinel 数量
redis-cli -p 26379 info sentinel | grep sentinel_masters

# 确保至少 3 个 Sentinel 且多数存活
# 如果 Sentinel 不足，添加新的 Sentinel
redis-cli -p 26379 sentinel monitor <master-name> <ip> <port> <quorum>
```

2. **从节点配置不正确**：
```bash
# 检查从节点的复制配置
redis-cli -h <slave-ip> info replication

# 确保从节点配置正确
redis-cli -h <slave-ip> config set masterauth <password>
redis-cli -h <slave-ip> slaveof <master-ip> <master-port>

# 重新配置从节点
redis-cli -p 26379 sentinel reset <master-name>
```

3. **手动强制故障转移**：
```bash
# 检查是否可以故障转移
redis-cli -p 26379 sentinel failover <master-name>

# 如果失败，检查日志中的具体原因
grep "+failover" /var/log/redis/sentinel.log
```

4. **Sentinel 配置优化**：
```conf
# sentinel.conf
sentinel monitor mymaster 127.0.0.1 6379 2
sentinel down-after-milliseconds mymaster 5000
sentinel failover-timeout mymaster 60000
sentinel parallel-syncs mymaster 1
sentinel auth-pass mymaster <password>
```

### 场景 8：Cluster 节点通信异常

**症状**：节点间无法通信，cluster nodes 显示节点状态为 disconnected 或 fail

**诊断流程**：
```bash
# 1. 检查节点端口连通性
telnet <node-ip> 6379
telnet <node-ip> 16379  # Cluster 总线端口

# 2. 检查节点配置
redis-cli config get cluster-*

# 3. 检查网络延迟
redis-cli --latency -h <node-ip> -p <port>

# 4. 查看节点间的连接
ss -antp | grep redis
```

**处理方案**：

1. **Cluster 总线端口不通**：
```bash
# Cluster 总线端口 = 数据端口 + 10000
# 确保防火墙开放 16379 等端口

# 临时关闭防火墙（测试）
systemctl stop firewalld

# 或者添加规则
firewall-cmd --add-port=16379/tcp --permanent
firewall-cmd --reload
```

2. **节点 IP 变更**：
```bash
# 更新节点 IP
redis-cli config set cluster-announce-ip <new-ip>
redis-cli config set cluster-announce-port <port>
redis-cli config set cluster-announce-bus-port <bus-port>

# 重启节点
systemctl restart redis
```

3. **重置并重新加入集群**：
```bash
# 在故障节点上执行
redis-cli cluster reset hard

# 重新加入集群
redis-cli cluster meet <master-ip> <master-port>

# 如果是从节点，重新设置主节点
redis-cli cluster replicate <master-node-id>
```

## 数据治理

### 大 Key 扫描
```bash
# 使用 --bigkeys 扫描
redis-cli --bigkeys

# 使用 rdb-tools 详细分析
rdb -c memory dump.rdb > memory.csv
sort -t, -k4 -nr memory.csv | head -20

# 处理大 Key：
# 1. 拆分：将大 Hash 拆分成多个小 Hash
# 2. 压缩：使用压缩算法
# 3. 清理：设置合理的 TTL
```

### 热点 Key 发现
```bash
# 使用 redis-faina 分析热点
cat monitor.log | redis-faina

# 或者使用 monitor + 分析
redis-cli monitor | head -n 10000 > monitor.log
cat monitor.log | grep -o '"[^"]*"' | sort | uniq -c | sort -rn | head -20
```

### Key 生命周期管理
```bash
# 查看带 TTL 的 Key 数量
redis-cli info keyspace

# 查看即将过期的 Key（采样）
redis-cli --eval "
local keys = redis.call('scan', 0, 'count', 1000)[2]
for _, key in ipairs(keys) do
  local ttl = redis.call('ttl', key)
  if ttl > 0 and ttl < 3600 then
    print(key .. ' expires in ' .. ttl .. ' seconds')
  end
end
"

# 设置合理的过期策略
# 避免大量 Key 同时过期，添加随机偏移
SET key value EX 3600
```

## 监控指标

### 核心监控项
```yaml
# 连接类
connected_clients: 当前连接数，接近 maxclients 时告警
blocked_clients: 阻塞客户端数，>0 可能有问题
rejected_connections: 拒绝连接数，>0 说明连接数超限

# 内存类
used_memory: 已使用内存
total_system_memory: 系统总内存
maxmemory: 配置的最大内存
used_memory_rss: 进程占用内存
mem_fragmentation_ratio: 内存碎片率，>1.5 告警

# 性能类
instantaneous_ops_per_sec: 实时 QPS
keyspace_hits: 键命中次数
keyspace_misses: 键未命中次数
keyspace_hit_rate: 命中率（计算得出）
sync_partial_ok: 部分同步成功次数
sync_full: 全量同步次数（过多告警）

# 持久化类
rdb_last_bgsave_status: 上次 RDB 状态
aof_last_bgrewrite_status: 上次 AOF 重写状态
aof_last_write_status: 上次 AOF 写入状态

# 复制类
master_link_status: 主从连接状态
master_last_io_seconds_ago: 主从通信延迟
master_sync_in_progress: 是否正在同步
```

## 输出规范

诊断报告格式：
```
🔴 Redis 诊断报告

📊 基础信息
- 版本：[version]
- 运行模式：[standalone/sentinel/cluster]
- 角色：[master/slave]
- 运行时间：[uptime_in_days] 天

💾 内存分析
- 已使用：[used_memory_human]
- 峰值：[used_memory_peak_human]
- 限制：[maxmemory_human]
- 碎片率：[mem_fragmentation_ratio]
- 趋势：[增长/稳定/下降]

🔌 连接分析
- 当前连接：[connected_clients]
- 最大连接：[maxclients]
- 阻塞连接：[blocked_clients]
- 连接趋势：[趋势描述]

⚡ 性能分析
- 当前 QPS：[instantaneous_ops_per_sec]
- 命中率：[keyspace_hit_rate]%
- 慢查询：[slowlog 数量]

🔍 问题发现
1. [问题1描述]
2. [问题2描述]

💡 根因分析
[详细分析原因]

🛠️ 解决方案
即时处理：
[具体命令]

根本解决：
[长期方案]

📋 优化建议
- [建议1]
- [建议2]
```

## 生产环境故障场景

### 场景 1：Redis 连接数耗尽

**症状**：应用报错 "ERR max number of clients reached"，新连接被拒绝

**诊断流程**：
```bash
# 1. 查看当前连接数
redis-cli info clients | grep connected_clients

# 2. 查看最大连接数限制
redis-cli config get maxclients

# 3. 查看连接详情
redis-cli client list | head -20

# 4. 统计各 IP 连接数
redis-cli client list | awk -F'addr=' '{print $2}' | awk -F':' '{print $1}' | sort | uniq -c | sort -rn | head -10
```

**常见原因**：
- 连接池配置不当，未正确关闭连接
- 应用异常导致连接泄漏
- 大量空闲连接未超时释放
- 突发流量导致连接激增

**处理方案**：
```bash
# 临时增加最大连接数
redis-cli config set maxclients 20000

# 关闭空闲连接（空闲超过 60 秒）
redis-cli client list | awk -F'[ =]' '{if($18>60)print$3}' | xargs -I {} redis-cli client kill addr {}

# 按类型关闭连接
redis-cli client kill type pubsub  # 关闭发布订阅连接
redis-cli client kill type normal  # 关闭普通连接
```

**预防措施**：
```bash
# 设置连接超时（秒）
redis-cli config set timeout 300

# 设置 TCP keepalive
redis-cli config set tcp-keepalive 60
```

### 场景 2：Redis 内存溢出 (OOM)

**症状**：写入操作报错 "OOM command not allowed when used memory > 'maxmemory'"

**诊断流程**：
```bash
# 1. 查看内存使用情况
redis-cli info memory | grep -E "used_memory|maxmemory|used_memory_peak"

# 2. 查看内存策略
redis-cli config get maxmemory-policy

# 3. 查找大 Key
redis-cli --bigkeys

# 4. 查看 Key 数量分布
redis-cli info keyspace
```

**处理方案**：

1. **紧急扩容**：
```bash
# 临时增加内存限制（需重启生效）
# 编辑 redis.conf
maxmemory 8gb

# 或运行时调整（如果系统内存充足）
redis-cli config set maxmemory 8589934592  # 8GB in bytes
```

2. **调整淘汰策略**：
```bash
# 所有 Key LRU 淘汰（最常用）
redis-cli config set maxmemory-policy allkeys-lru

# 带过期时间的 Key LRU 淘汰
redis-cli config set maxmemory-policy volatile-lru

# 随机淘汰（性能最好）
redis-cli config set maxmemory-policy allkeys-random
```

3. **清理数据**：
```bash
# 删除特定模式的 Key（异步删除，不阻塞）
redis-cli --scan --pattern "temp:*" | xargs -L 100 redis-cli unlink

# 清空当前数据库（危险操作！）
redis-cli flushdb async
```

### 场景 3：主从复制延迟/中断

**症状**：从节点数据与主节点不一致，或复制完全停止

**诊断流程**：
```bash
# 1. 查看复制状态
redis-cli info replication

# 2. 查看主从延迟（从节点执行）
redis-cli info replication | grep master_last_io_seconds_ago

# 3. 查看复制缓冲区使用情况
redis-cli info replication | grep repl_backlog

# 4. 查看积压缓冲区大小
redis-cli config get repl-backlog-size
```

**常见问题及处理**：

1. **复制缓冲区溢出**：
```bash
# 增大复制缓冲区（默认 1MB，建议至少 64MB）
redis-cli config set repl-backlog-size 67108864  # 64MB

# 永久修改 redis.conf
repl-backlog-size 64mb
```

2. **网络不稳定导致频繁全量同步**：
```bash
# 增大 client-output-buffer-limit
redis-cli config set client-output-buffer-limit 'slave 0 0 0'  # 不限制从节点缓冲区

# 或设置合理的限制
redis-cli config set client-output-buffer-limit 'slave 512mb 128mb 60'
```

3. **重新建立复制**：
```bash
# 在从节点执行
redis-cli slaveof no one
redis-cli slaveof <master-ip> <master-port>

# 或使用新命令（Redis 5.0+）
redis-cli replicaof no one
redis-cli replicaof <master-ip> <master-port>
```

### 场景 4：持久化阻塞

**症状**：Redis 响应变慢，出现大量慢查询，或客户端超时

**诊断流程**：
```bash
# 1. 查看持久化状态
redis-cli info persistence

# 2. 查看是否正在进行持久化
redis-cli info persistence | grep -E "rdb_bgsave_in_progress|aof_rewrite_in_progress"

# 3. 查看上次持久化耗时
redis-cli info persistence | grep -E "rdb_last_bgsave_time_sec|aof_last_rewrite_time_sec"

# 4. 查看 fork 耗时（Redis 6.0+）
redis-cli info stats | grep fork
```

**处理方案**：

1. **RDB 优化**：
```bash
# 禁用自动 RDB（如果业务可接受）
redis-cli config set save ""

# 或调整 RDB 策略（减少频率）
# redis.conf
save 900 1      # 15分钟至少1个key变化
save 300 10     # 5分钟至少10个key变化
save 60 10000   # 1分钟至少10000个key变化
```

2. **AOF 优化**：
```bash
# 使用每秒刷盘（性能与持久化平衡）
redis-cli config set appendfsync everysec

# 增加 AOF 重写触发阈值
redis-cli config set auto-aof-rewrite-min-size 128mb
redis-cli config set auto-aof-rewrite-percentage 200
```

3. **在从节点执行持久化**：
```bash
# 主节点禁用 RDB
redis-cli config set save ""

# 从节点开启 RDB
redis-cli config set save "900 1 300 10 60 10000"
```

### 场景 5：热点 Key 问题

**症状**：某些 Key 访问量极高，导致单节点 CPU/带宽打满

**诊断流程**：
```bash
# 1. 使用 monitor 分析热点（生产环境慎用，性能影响大）
redis-cli monitor | head -n 10000 > monitor.log

# 2. 分析热点 Key
cat monitor.log | grep -o '"[^"]*"' | sort | uniq -c | sort -rn | head -20

# 3. 使用 Redis 4.0+ 的 hotkeys 功能
redis-cli --hotkeys

# 4. 查看命令统计
redis-cli info commandstats | sort -k2 -rn | head -20
```

**处理方案**：

1. **本地缓存**：在应用层增加本地缓存（Caffeine/Guava Cache），减少对 Redis 的访问

2. **Key 拆分**：
```bash
# 将热点 Key 拆分为多个 Key
# 原：user:1000:profile
# 新：user:1000:profile:0, user:1000:profile:1, ...
```

3. **读写分离**：
```bash
# 配置从节点可读
redis-cli config set slave-read-only yes

# 应用层实现读写分离，读操作路由到从节点
```

4. **使用 Redis Cluster**：将热点 Key 分散到多个分片

### 场景 6：AOF 文件损坏

**症状**：Redis 启动失败，日志显示 "Bad file format reading the append only file"

**诊断流程**：
```bash
# 1. 检查 AOF 文件完整性
redis-check-aof --fix appendonly.aof

# 2. 查看 AOF 文件大小
ls -lh /var/lib/redis/appendonly.aof

# 3. 检查 Redis 日志
tail -100 /var/log/redis/redis-server.log
```

**处理方案**：

1. **修复 AOF 文件**：
```bash
# 备份原 AOF 文件
cp /var/lib/redis/appendonly.aof /var/lib/redis/appendonly.aof.bak.$(date +%Y%m%d)

# 使用 redis-check-aof 修复
redis-check-aof --fix /var/lib/redis/appendonly.aof

# 重启 Redis
systemctl restart redis
```

2. **如果修复失败，从 RDB 恢复**：
```bash
# 关闭 Redis
systemctl stop redis

# 禁用 AOF 启动（临时）
redis-server /etc/redis/redis.conf --appendonly no

# 生成新的 AOF 文件
redis-cli bgrewriteaof

# 重新启用 AOF
# 修改 redis.conf: appendonly yes
systemctl restart redis
```

### 场景 7：RDB 文件损坏

**症状**：Redis 启动失败或数据加载异常，日志显示 "Short read or OOM loading DB"

**诊断流程**：
```bash
# 1. 检查 RDB 文件完整性
redis-check-rdb /var/lib/redis/dump.rdb

# 2. 查看 RDB 文件大小
ls -lh /var/lib/redis/dump.rdb

# 3. 检查文件头信息
head -c 100 /var/lib/redis/dump.rdb | od -c
```

**处理方案**：

1. **从备份恢复**：
```bash
# 备份损坏的 RDB
cp /var/lib/redis/dump.rdb /var/lib/redis/dump.rdb.corrupted.$(date +%Y%m%d)

# 从最近备份恢复
cp /backup/redis/dump.rdb.$(date +%Y%m%d) /var/lib/redis/dump.rdb

# 重启 Redis
systemctl restart redis
```

2. **如果没有备份，尝试修复**：
```bash
# 使用 rdbtools 导出数据
rdb -c json /var/lib/redis/dump.rdb > data.json

# 或使用 rdb 恢复部分数据
rdb --command protocol /var/lib/redis/dump.rdb | redis-cli --pipe
```

### 场景 8：Redis 启动失败

**症状**：Redis 服务无法启动，systemctl status 显示失败

**诊断流程**：
```bash
# 1. 查看详细错误日志
journalctl -u redis -n 100 --no-pager

# 2. 检查配置文件语法
redis-server /etc/redis/redis.conf --test-memory

# 3. 检查端口占用
ss -tlnp | grep 6379

# 4. 检查文件权限
ls -la /var/lib/redis/
ls -la /var/log/redis/

# 5. 检查系统资源
free -h
df -h
ulimit -a
```

**常见原因及处理**：

| 原因 | 判断方法 | 解决方案 |
|------|---------|---------|
| 端口被占用 | `ss -tlnp \| grep 6379` | 停止占用进程或更换端口 |
| 内存不足 | `free -h` 显示可用内存极少 | 增加内存或降低 maxmemory |
| 文件权限错误 | `ls -la` 显示错误属主 | `chown redis:redis /var/lib/redis` |
| 配置文件错误 | 日志显示配置行号 | 检查并修复 redis.conf |
| AOF/RDB 损坏 | 前文所述错误 | 按前文方案修复 |

## 补充内容（2024年新增）

### Redis 7.0+ 新特性应用

**多部分 AOF（MP-AOF）**：
```bash
# Redis 7.0 自动启用 MP-AOF
# 查看 AOF 文件列表
ls -la /var/lib/redis/

# 手动触发 AOF 重写
redis-cli bgrewriteaof

# 查看 AOF 基础文件和增量文件
redis-cli info persistence | grep aof_base
redis-cli info persistence | grep aof_incremental
```

**Functions（函数）**：
```bash
# 加载 Redis Function
redis-cli FUNCTION LOAD "#!lua name=mylib\n redis.register_function('myfunc', function(keys, args) return redis.call('GET', keys[1]) end)"

# 执行 Function
redis-cli FCALL mylib.myfunc 1 mykey

# 查看已加载的 Functions
redis-cli FUNCTION LIST
```

**ACL 日志**：
```bash
# 查看 ACL 拒绝日志
redis-cli ACL LOG

# 查看最近 10 条 ACL 日志
redis-cli ACL LOG 10

# 重置 ACL 日志
redis-cli ACL LOG RESET
```

### 性能调优高级技巧

**内存优化**：
```bash
# 1. 使用 ziplist 编码（小哈希/列表）
# redis.conf
hash-max-ziplist-entries 512
hash-max-ziplist-value 64
list-max-ziplist-size -2
zset-max-ziplist-entries 128
zset-max-ziplist-value 64

# 2. 启用内存碎片整理
redis-cli config set activedefrag yes
redis-cli config set active-defrag-threshold-lower 10
redis-cli config set active-defrag-threshold-upper 100
redis-cli config set active-defrag-cycle-min 5
redis-cli config set active-defrag-cycle-max 75

# 3. 查看编码方式
redis-cli debug object mykey
```

**管道和事务优化**：
```bash
# 使用管道批量操作
cat commands.txt | redis-cli --pipe

# 使用 Lua 脚本减少网络往返
redis-cli EVAL "redis.call('SET', KEYS[1], ARGV[1]); redis.call('EXPIRE', KEYS[1], ARGV[2]); return 1" 1 mykey myvalue 3600

# 使用事务
redis-cli MULTI
redis-cli SET key1 value1
redis-cli SET key2 value2
redis-cli EXEC
```

### 监控和可观测性

**使用 Redis INFO 命令深度监控**：
```bash
# 创建监控脚本
#!/bin/bash
# redis_monitor.sh

HOST=${1:-localhost}
PORT=${2:-6379}

while true; do
    TIMESTAMP=$(date '+%Y-%m-%d %H:%M:%S')

    # 获取关键指标
    INFO=$(redis-cli -h $HOST -p $PORT info stats)
    MEM=$(redis-cli -h $HOST -p $PORT info memory)

    QPS=$(echo "$INFO" | grep instantaneous_ops_per_sec | cut -d: -f2 | tr -d '\r')
    HIT_RATE=$(echo "$INFO" | awk -F: '/keyspace_hits/{hits=$2} /keyspace_misses/{misses=$2} END{if(hits+misses>0) printf "%.2f", hits/(hits+misses)*100}')
    MEM_USAGE=$(echo "$MEM" | grep used_memory: | cut -d: -f2 | tr -d '\r')

    echo "$TIMESTAMP, QPS: $QPS, HitRate: ${HIT_RATE}%, Memory: $MEM_USAGE"

    sleep 5
done
```

**使用 Redis Exporter + Prometheus**：
```yaml
# docker-compose.yml
version: '3'
services:
  redis-exporter:
    image: oliver006/redis_exporter:latest
    environment:
      - REDIS_ADDR=redis://redis:6379
    ports:
      - "9121:9121"
```

**关键监控指标**：
```yaml
# Prometheus 告警规则
groups:
  - name: redis
    rules:
      - alert: RedisMemoryHigh
        expr: redis_memory_used_bytes / redis_memory_max_bytes > 0.85
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Redis memory usage high"

      - alert: RedisReplicationLag
        expr: redis_master_link_up{role="slave"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Redis replication is down"
```

### 数据迁移方案

**使用 redis-shake 进行迁移**：
```bash
# 下载 redis-shake
wget https://github.com/alibaba/RedisShake/releases/download/v3.1.11/redis-shake-v3.1.11-linux-amd64.tar.gz
tar -xzf redis-shake-v3.1.11-linux-amd64.tar.gz

# 配置迁移任务
cat > shake.toml << 'EOF'
[source]
type = "standalone"
address = "source-redis:6379"
password = "source-password"

[target]
type = "standalone"
address = "target-redis:6379"
password = "target-password"

[filter]
# 只迁移特定 key
key_white_list = ["user:*", "session:*"]
EOF

# 执行迁移
./redis-shake shake.toml
```

**使用 SCAN 进行在线迁移**：
```bash
#!/bin/bash
# redis_migrate.sh

SOURCE_HOST="source-redis"
TARGET_HOST="target-redis"
CURSOR=0

while true; do
    # SCAN 获取 key
    RESULT=$(redis-cli -h $SOURCE_HOST SCAN $CURSOR COUNT 1000)
    CURSOR=$(echo "$RESULT" | head -1)
    KEYS=$(echo "$RESULT" | tail -n +2)

    # 迁移这些 key
    for KEY in $KEYS; do
        # 使用 DUMP/RESTORE
        redis-cli -h $SOURCE_HOST --raw DUMP "$KEY" | head -c -1 | redis-cli -h $TARGET_HOST -x RESTORE "$KEY" 0
        echo "Migrated: $KEY"
    done

    # 如果 cursor 为 0，表示完成
    if [ "$CURSOR" == "0" ]; then
        break
    fi
done

echo "Migration completed!"
```

### 安全加固增强

**TLS 加密通信（Redis 6.0+）**：
```bash
# 生成证书
openssl genrsa -out redis.key 4096
openssl req -new -x509 -days 365 -key redis.key -out redis.crt -subj "/CN=redis"

# 配置 Redis 使用 TLS
# redis.conf
tls-port 6379
port 0
tls-cert-file /etc/redis/redis.crt
tls-key-file /etc/redis/redis.key
tls-ca-cert-file /etc/redis/ca.crt
tls-protocols "TLSv1.2 TLSv1.3"

# 客户端连接
redis-cli --tls --cert redis.crt --key redis.key --cacert ca.crt -h redis-server
```

**ACL 细粒度权限控制（Redis 6.0+）**：
```bash
# 创建只读用户
redis-cli ACL SETUSER readonly ON >readonly_password ~* +@read

# 创建特定前缀用户
redis-cli ACL SETUSER appuser ON >app_password ~app:* +@all -@dangerous

# 创建仅允许特定命令的用户
redis-cli ACL SETUSER cacheuser ON >cache_password ~cache:* +get +set +del +expire

# 查看 ACL 列表
redis-cli ACL LIST

# 保存 ACL 到文件
redis-cli ACL SAVE
```

## 生产环境最佳实践

### 1. 部署架构建议

**小型业务（<10GB 数据）**：
- 单实例 + 持久化 + 定期备份
- 或主从复制（1主1从）

**中型业务（10GB-100GB）**：
- 主从复制 + Sentinel 高可用
- 1主 + 2从 + 3 Sentinel

**大型业务（>100GB 或高并发）**：
- Redis Cluster（至少 6 节点：3主3从）
- 按业务垂直拆分多个集群

### 2. 配置优化

```conf
# redis.conf 生产环境推荐配置

# 内存配置
maxmemory 4gb
maxmemory-policy allkeys-lru

# 连接配置
tcp-backlog 511
timeout 300
tcp-keepalive 60
maxclients 10000

# 持久化配置（主节点）
# 关闭自动 RDB，使用 AOF
save ""
appendonly yes
appendfsync everysec
auto-aof-rewrite-min-size 128mb
auto-aof-rewrite-percentage 100

# 复制配置
repl-diskless-sync yes
repl-diskless-sync-delay 5
repl-backlog-size 64mb
repl-timeout 60

# 客户端输出缓冲区
client-output-buffer-limit normal 0 0 0
client-output-buffer-limit slave 512mb 128mb 60
client-output-buffer-limit pubsub 32mb 8mb 60

# 慢查询配置
slowlog-log-slower-than 10000
slowlog-max-len 128

# 安全配置
rename-command FLUSHDB ""
rename-command FLUSHALL ""
rename-command CONFIG "CONFIG_9f3b2a1e"
```

### 3. 监控告警指标

| 指标 | 告警阈值 | 说明 |
|------|---------|------|
| used_memory/maxmemory | > 80% | 内存使用率 |
| connected_clients/maxclients | > 80% | 连接使用率 |
| instantaneous_ops_per_sec | > 基线 200% | QPS 突增 |
| keyspace_hit_rate | < 80% | 命中率过低 |
| master_last_io_seconds_ago | > 10s | 主从复制延迟 |
| rdb_last_bgsave_status | != ok | RDB 失败 |
| aof_last_bgrewrite_status | != ok | AOF 重写失败 |
| rejected_connections | > 0 | 连接被拒绝 |
| evicted_keys | > 100/min | Key 被驱逐 |

### 4. 数据备份策略

```bash
#!/bin/bash
# redis_backup.sh

BACKUP_DIR="/backup/redis/$(date +%Y%m%d)"
mkdir -p $BACKUP_DIR

# 1. 触发 RDB 保存
redis-cli bgsave

# 2. 等待 RDB 完成
while redis-cli info persistence | grep -q "rdb_bgsave_in_progress:1"; do
    sleep 1
done

# 3. 复制 RDB 文件
cp /var/lib/redis/dump.rdb $BACKUP_DIR/

# 4. 备份 AOF 文件（如果启用）
if [ -f /var/lib/redis/appendonly.aof ]; then
    cp /var/lib/redis/appendonly.aof $BACKUP_DIR/
fi

# 5. 压缩备份
cd /backup/redis
tar czf $(date +%Y%m%d).tar.gz $(date +%Y%m%d)
rm -rf $(date +%Y%m%d)

# 6. 保留最近 7 天备份
find /backup/redis -name "*.tar.gz" -mtime +7 -delete
```

### 5. 安全加固

```bash
# 1. 启用认证
redis-cli config set requirepass "YourStrongPassword123!"

# 2. 禁用危险命令——rename-command 不支持运行时 CONFIG SET（Redis 报
#    Unsupported CONFIG parameter，静默失败会让人误以为已封禁）。必须写入
#    redis.conf 并重启生效：
#    rename-command FLUSHDB ""
#    rename-command FLUSHALL ""
#    rename-command CONFIG "CONFIG_<随机后缀>"
#    rename-command SHUTDOWN "SHUTDOWN_<随机后缀>"
#    重启属变更动作：进建议面走审批，不在采集轮执行。

# 3. 绑定特定网卡
# redis.conf
bind 127.0.0.1 198.51.100.1
protected-mode yes

# 4. 使用 TLS（Redis 6.0+）
# redis.conf
tls-port 6379
port 0
tls-cert-file /path/to/redis.crt
tls-key-file /path/to/redis.key
tls-ca-cert-file /path/to/ca.crt
```

### 6. 容量规划指南

**内存容量规划**：
```bash
# 估算 Key 内存占用
redis-cli --eval '
local keys = redis.call("keys", "*")
local total_size = 0
for _, key in ipairs(keys) do
    local debug = redis.call("debug", "object", key)
    local size = string.match(debug, "serializedlength:(%d+)")
    if size then
        total_size = total_size + tonumber(size)
    end
end
return total_size
'

# 估算每日增长量
redis-cli info stats | grep -E "keyspace_hits|keyspace_misses"
```

**分片规划（Cluster）**：
```bash
# 计算每个分片的 Key 数量
redis-cli --cluster check <node-ip>:6379

# 建议：每个分片不超过 1GB 内存或 100万 Key
# 总内存需求 = 数据量 / 0.7 (考虑碎片和增长)
```

### 7. 灾难恢复演练

**定期演练脚本**：
```bash
#!/bin/bash
# redis_dr_drill.sh

REDIS_HOST="localhost"
REDIS_PORT="6379"
BACKUP_DIR="/backup/redis"

# 1. 创建测试数据
echo "Creating test data..."
redis-cli -h $REDIS_HOST -p $REDIS_PORT SET dr:test:key "test_value"
redis-cli -h $REDIS_HOST -p $REDIS_PORT EXPIRE dr:test:key 3600

# 2. 执行备份
echo "Performing backup..."
redis-cli -h $REDIS_HOST -p $REDIS_PORT bgsave
sleep 5

# 3. 模拟数据丢失——演练纪律（硬守卫）：本脚本只允许在专用演练环境执行，
#    运行前必须显式确认 REDIS_HOST 指向演练实例（set -e 前校验 $DRILL_CONFIRMED=yes
#    且 REDIS_HOST 含 drill/test 关键字，否则拒绝执行）；FLUSHDB 在生产实例上
#    执行即清库事故。
[ "$DRILL_CONFIRMED" = "yes" ] || { echo "需显式 DRILL_CONFIRMED=yes 且确认目标为演练实例"; exit 1; }
case "$REDIS_HOST" in *drill*|*test*) ;; *) echo "REDIS_HOST 非演练实例，拒绝执行"; exit 1;; esac
echo "Simulating data loss..."
redis-cli -h $REDIS_HOST -p $REDIS_PORT FLUSHDB

# 4. 从备份恢复
echo "Restoring from backup..."
systemctl stop redis
cp $BACKUP_DIR/dump.rdb /var/lib/redis/
systemctl start redis

# 5. 验证恢复
VALUE=$(redis-cli -h $REDIS_HOST -p $REDIS_PORT GET dr:test:key)
if [ "$VALUE" == "test_value" ]; then
    echo "DR Drill PASSED: Data recovered successfully"
else
    echo "DR Drill FAILED: Data recovery failed"
fi

# 清理测试数据
redis-cli -h $REDIS_HOST -p $REDIS_PORT DEL dr:test:key
```

## 常用脚本

### 健康检查脚本（增强版）
```bash
#!/bin/bash
# redis_health_check.sh

HOST=${1:-localhost}
PORT=${2:-6379}
PASSWORD=${3:-}
WARN_MEM_THRESHOLD=80
WARN_CONN_THRESHOLD=80

if [ -n "$PASSWORD" ]; then
  AUTH="-a $PASSWORD"
else
  AUTH=""
fi

echo "=== Redis 健康检查 ==="
echo "时间: $(date)"
echo "节点: $HOST:$PORT"

# 连接测试
if ! redis-cli -h $HOST -p $PORT $AUTH ping | grep -q PONG; then
  echo "❌ 连接失败"
  exit 1
fi
echo "✅ 连接正常"

# 内存检查
echo -e "\n[内存状态]"
MEM_INFO=$(redis-cli -h $HOST -p $PORT $AUTH info memory)
echo "$MEM_INFO" | grep -E "used_memory_human|maxmemory_human|mem_fragmentation_ratio"

# 内存使用率检查
MAXMEM=$(echo "$MEM_INFO" | grep maxmemory | grep -v human | cut -d: -f2 | tr -d '\r')
USEDMEM=$(echo "$MEM_INFO" | grep "^used_memory:" | cut -d: -f2 | tr -d '\r')
if [ "$MAXMEM" -gt 0 ]; then
  MEM_USAGE=$((USEDMEM * 100 / MAXMEM))
  if [ $MEM_USAGE -gt $WARN_MEM_THRESHOLD ]; then
    echo "⚠️ 内存使用率: ${MEM_USAGE}% (超过 ${WARN_MEM_THRESHOLD}%)"
  else
    echo "✅ 内存使用率: ${MEM_USAGE}%"
  fi
fi

# 连接检查
echo -e "\n[连接状态]"
CLIENT_INFO=$(redis-cli -h $HOST -p $PORT $AUTH info clients)
echo "$CLIENT_INFO" | grep -E "connected_clients|blocked_clients"

CONN=$(echo "$CLIENT_INFO" | grep connected_clients | cut -d: -f2 | tr -d '\r')
MAXCONN=$(redis-cli -h $HOST -p $PORT $AUTH config get maxclients | tail -1)
CONN_USAGE=$((CONN * 100 / MAXCONN))
if [ $CONN_USAGE -gt $WARN_CONN_THRESHOLD ]; then
  echo "⚠️ 连接使用率: ${CONN_USAGE}% (超过 ${WARN_CONN_THRESHOLD}%)"
else
  echo "✅ 连接使用率: ${CONN_USAGE}%"
fi

# 复制检查
echo -e "\n[复制状态]"
REPL_INFO=$(redis-cli -h $HOST -p $PORT $AUTH info replication)
echo "$REPL_INFO" | grep -E "role|master_link_status|master_last_io_seconds_ago|connected_slaves"

# 慢查询检查
echo -e "\n[慢查询]"
SLOW_COUNT=$(redis-cli -h $HOST -p $PORT $AUTH slowlog len)
echo "慢查询数量: $SLOW_COUNT"
if [ "$SLOW_COUNT" -gt 100 ]; then
  echo "⚠️ 慢查询数量过多，建议优化"
fi

# Key 空间
echo -e "\n[Key 空间]"
redis-cli -h $HOST -p $PORT $AUTH info keyspace

# 统计信息
echo -e "\n[统计信息]"
redis-cli -h $HOST -p $PORT $AUTH info stats | grep -E "total_connections_received|total_commands_processed|instantaneous_ops_per_sec|keyspace_hits|keyspace_misses"

# 命中率计算
HITS=$(redis-cli -h $HOST -p $PORT $AUTH info stats | grep keyspace_hits | cut -d: -f2 | tr -d '\r')
MISSES=$(redis-cli -h $HOST -p $PORT $AUTH info stats | grep keyspace_misses | cut -d: -f2 | tr -d '\r')
if [ $((HITS + MISSES)) -gt 0 ]; then
  HIT_RATE=$((HITS * 100 / (HITS + MISSES)))
  echo "命中率: ${HIT_RATE}%"
fi

echo -e "\n=== 检查完成 ==="
```

### 性能分析脚本
```bash
#!/bin/bash
# redis_performance_analysis.sh

HOST=${1:-localhost}
PORT=${2:-6379}
PASSWORD=${3:-}

if [ -n "$PASSWORD" ]; then
  AUTH="-a $PASSWORD"
else
  AUTH=""
fi

echo "=== Redis 性能分析 ==="
echo "时间: $(date)"

# 1. 慢查询分析
echo -e "\n[慢查询 TOP 10]"
redis-cli -h $HOST -p $PORT $AUTH slowlog get 10 | grep -E "^1\)|^2\)|command"

# 2. 命令统计
echo -e "\n[命令统计 TOP 10]"
redis-cli -h $HOST -p $PORT $AUTH info commandstats | sort -t= -k2 -rn | head -10

# 3. 大 Key 扫描
echo -e "\n[大 Key 扫描]"
redis-cli -h $HOST -p $PORT $AUTH --bigkeys

# 4. 延迟测试
echo -e "\n[延迟测试]"
redis-cli -h $HOST -p $PORT $AUTH --latency-history -i 1

# 5. 内存碎片分析
echo -e "\n[内存碎片分析]"
redis-cli -h $HOST -p $PORT $AUTH info memory | grep -E "mem_fragmentation_ratio|mem_fragmentation_bytes"

# 6. 阻塞客户端
echo -e "\n[阻塞客户端]"
BLOCKED=$(redis-cli -h $HOST -p $PORT $AUTH info clients | grep blocked_clients | cut -d: -f2 | tr -d '\r')
if [ "$BLOCKED" -gt 0 ]; then
  echo "⚠️ 发现 $BLOCKED 个阻塞客户端"
  redis-cli -h $HOST -p $PORT $AUTH client list | grep blocked
else
  echo "✅ 无阻塞客户端"
fi
```

## 参考资料

### 官方文档
- [Redis 官方文档](https://redis.io/documentation)
- [Redis 命令参考](https://redis.io/commands)
- [Redis 持久化](https://redis.io/topics/persistence)
- [Redis 复制](https://redis.io/topics/replication)
- [Redis 集群](https://redis.io/topics/cluster-tutorial)

### 社区资源
- [Redis 最佳实践](https://redis.io/topics/latency)
- [Redis 内存优化](https://redis.io/topics/memory-optimization)
- [Redis 安全](https://redis.io/topics/security)

### 工具推荐
- **RedisInsight**: Redis 官方可视化工具
- **redis-rdb-tools**: RDB 文件分析工具
- **redis-faina**: 基于 MONITOR 的分析工具
- **redis-stat**: 实时监控工具

## 常用脚本

### 健康检查脚本
```bash
#!/bin/bash
# redis_health_check.sh

HOST=${1:-localhost}
PORT=${2:-6379}
PASSWORD=${3:-}

if [ -n "$PASSWORD" ]; then
  AUTH="-a $PASSWORD"
else
  AUTH=""
fi

echo "=== Redis 健康检查 ==="
echo "时间: $(date)"
echo "节点: $HOST:$PORT"

# 连接测试
if ! redis-cli -h $HOST -p $PORT $AUTH ping | grep -q PONG; then
  echo "❌ 连接失败"
  exit 1
fi
echo "✅ 连接正常"

# 内存检查
echo -e "\n[内存状态]"
redis-cli -h $HOST -p $PORT $AUTH info memory | grep -E "used_memory_human|maxmemory_human|mem_fragmentation_ratio"

# 连接检查
echo -e "\n[连接状态]"
redis-cli -h $HOST -p $PORT $AUTH info clients | grep -E "connected_clients|blocked_clients"

# 复制检查
echo -e "\n[复制状态]"
redis-cli -h $HOST -p $PORT $AUTH info replication | grep -E "role|master_link_status|master_last_io_seconds_ago"

# 慢查询检查
echo -e "\n[慢查询]"
SLOW_COUNT=$(redis-cli -h $HOST -p $PORT $AUTH slowlog len)
echo "慢查询数量: $SLOW_COUNT"

# Key 空间
echo -e "\n[Key 空间]"
redis-cli -h $HOST -p $PORT $AUTH info keyspace
```

## MCP 工具支持

本 Skill 可通过 MCP (Model Context Protocol) 提供以下工具：

### 工具列表

| 工具名称 | 描述 | 必需参数 |
|---------|------|---------|
| `redis_check_health` | 检查 Redis 健康状态和基础信息 | host, port |
| `redis_get_memory_info` | 获取内存使用详情 | host, port |
| `redis_get_slowlog` | 获取慢查询日志 | host, port, count |
| `redis_analyze_bigkeys` | 扫描大 Key | host, port |
| `redis_check_replication` | 检查主从复制状态 | host, port |

### 工具定义示例

```json
{
  "name": "redis_check_health",
  "description": "检查 Redis 健康状态，返回服务器信息、连接数、QPS等",
  "inputSchema": {
    "type": "object",
    "properties": {
      "host": {
        "type": "string",
        "description": "Redis 主机地址",
        "default": "localhost"
      },
      "port": {
        "type": "integer",
        "description": "Redis 端口",
        "default": 6379
      },
      "password": {
        "type": "string",
        "description": "Redis 密码（可选）"
      }
    }
  }
}
```

```json
{
  "name": "redis_get_memory_info",
  "description": "获取 Redis 内存使用详情，包括已用内存、峰值、碎片率等",
  "inputSchema": {
    "type": "object",
    "properties": {
      "host": {
        "type": "string",
        "default": "localhost"
      },
      "port": {
        "type": "integer",
        "default": 6379
      },
      "password": {
        "type": "string"
      }
    }
  }
}
```

```json
{
  "name": "redis_get_slowlog",
  "description": "获取 Redis 慢查询日志",
  "inputSchema": {
    "type": "object",
    "properties": {
      "host": {
        "type": "string",
        "default": "localhost"
      },
      "port": {
        "type": "integer",
        "default": 6379
      },
      "password": {
        "type": "string"
      },
      "count": {
        "type": "integer",
        "description": "获取的慢查询数量",
        "default": 20
      }
    }
  }
}
```

```json
{
  "name": "redis_analyze_bigkeys",
  "description": "扫描 Redis 中的大 Key",
  "inputSchema": {
    "type": "object",
    "properties": {
      "host": {
        "type": "string",
        "default": "localhost"
      },
      "port": {
        "type": "integer",
        "default": 6379
      },
      "password": {
        "type": "string"
      },
      "sample_size": {
        "type": "integer",
        "description": "采样数量",
        "default": 100
      }
    }
  }
}
```

```json
{
  "name": "redis_check_replication",
  "description": "检查 Redis 主从复制状态",
  "inputSchema": {
    "type": "object",
    "properties": {
      "host": {
        "type": "string",
        "default": "localhost"
      },
      "port": {
        "type": "integer",
        "default": 6379
      },
      "password": {
        "type": "string"
      }
    }
  }
}
```

### Python MCP Server 示例

```python
from mcp.server import Server
from mcp.types import TextContent
import subprocess
import json

app = Server("redis-ops")

def build_redis_cmd(host, port, password, command):
    auth = f"-a {password}" if password else ""
    return f"redis-cli -h {host} -p {port} {auth} {command}"

@app.call_tool()
def call_tool(name: str, arguments: dict):
    host = arguments.get("host", "localhost")
    port = arguments.get("port", 6379)
    password = arguments.get("password", "")

    if name == "redis_check_health":
        commands = [
            "info server",
            "info clients",
            "info stats",
            "info keyspace"
        ]
        results = []
        for cmd in commands:
            full_cmd = build_redis_cmd(host, port, password, cmd)
            result = subprocess.run(full_cmd, shell=True, capture_output=True, text=True)
            results.append(f"=== {cmd} ===\n{result.stdout}")
        return [TextContent(type="text", text="\n\n".join(results))]

    elif name == "redis_get_memory_info":
        cmd = build_redis_cmd(host, port, password, "info memory")
        result = subprocess.run(cmd, shell=True, capture_output=True, text=True)
        return [TextContent(type="text", text=result.stdout)]

    elif name == "redis_get_slowlog":
        count = arguments.get("count", 20)
        cmd = build_redis_cmd(host, port, password, f"slowlog get {count}")
        result = subprocess.run(cmd, shell=True, capture_output=True, text=True)
        return [TextContent(type="text", text=result.stdout)]

    elif name == "redis_analyze_bigkeys":
        cmd = build_redis_cmd(host, port, password, "--bigkeys")
        result = subprocess.run(cmd, shell=True, capture_output=True, text=True)
        return [TextContent(type="text", text=result.stdout)]

    elif name == "redis_check_replication":
        cmd = build_redis_cmd(host, port, password, "info replication")
        result = subprocess.run(cmd, shell=True, capture_output=True, text=True)
        return [TextContent(type="text", text=result.stdout)]

if __name__ == "__main__":
    app.run()
```

## 示例

### 场景：Redis 响应慢

输入：
```
Redis 响应变慢，应用出现大量超时，如何排查？
```

输出：
```
🔴 Redis 诊断报告

📊 基础信息收集
请先执行以下命令收集信息：

redis-cli info all > redis_info.txt
redis-cli slowlog get 20 > slowlog.txt
redis-cli --latency-history -i 1 > latency.txt

💾 内存分析
检查项：
1. 内存是否达到上限
   redis-cli info memory | grep -E "used_memory|maxmemory"

2. 是否触发驱逐
   redis-cli info stats | grep evicted_keys

3. 内存碎片情况
   redis-cli info memory | grep mem_fragmentation_ratio

🔌 连接分析
检查项：
1. 连接数是否打满
   redis-cli info clients | grep connected_clients

2. 是否有阻塞连接
   redis-cli info clients | grep blocked_clients

3. 查看连接详情
   redis-cli client list | head -20

⚡ 性能分析
检查项：
1. 查看慢查询
   redis-cli slowlog get 20

2. 检查热点 Key
   redis-cli --hotkeys (Redis 4.0+)

3. 实时监控命令
   redis-cli monitor | head -n 1000 | sort | uniq -c | sort -rn | head -20

🔍 常见原因及处理

原因1：大 Key 阻塞
处理：
- 扫描大 Key：redis-cli --bigkeys
- 拆分大 Hash/List/Set
- 使用 SCAN 替代 KEYS

原因2：慢查询堆积
处理：
- 分析慢查询日志，优化命令
- 避免 O(N) 命令操作大集合
- 使用 UNLINK 替代 DEL（异步删除）

原因3：持久化阻塞
处理：
- 检查是否正在 bgsave/bgrewriteaof
- 调整持久化策略到从节点执行
- 使用 AOF 每秒刷盘替代 RDB

原因4：主从同步延迟
处理：
- 检查 master_last_io_seconds_ago
- 优化复制缓冲区大小
- 避免在从节点执行慢查询

原因5：内存达到上限
处理：
- 增大 maxmemory
- 优化淘汰策略：config set maxmemory-policy allkeys-lru
- 清理不必要的数据

📋 应急处理

如果需要立即恢复：

1. 临时扩容内存（如可重启）
2. 关闭持久化（临时）：config set save ""
3. 增加慢查询阈值：config set slowlog-log-slower-than 100000
4. 客户端限流或降级

🛠️ 长期优化
- 配置完善的监控告警
- 建立容量规划机制
- 优化数据结构和访问模式
- 考虑使用读写分离或 Cluster 架构
```
