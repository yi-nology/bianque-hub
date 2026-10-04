---
name: mongodb-ops
description: MongoDB 运维专家 - 集群管理、性能优化、数据分片、备份恢复
mode: on_demand
version: 1.1.0
maturity: experimental
---

## 配置说明

### 环境变量配置
```bash
export MONGO_URI="mongodb://localhost:27017"
export MONGO_DATABASE="mydb"
export MONGO_USERNAME=""
export MONGO_PASSWORD=""
```

## 输入参数

| 参数名 | 类型 | 必填 | 描述 | 示例 |
|--------|------|------|------|------|
| `database` | string | 否 | 数据库名称 | `myapp` |
| `collection` | string | 否 | 集合名称 | `users` |

## 输出格式

```json
{
  "status": "success",
  "data": {
    "databases": ["admin", "myapp", "test"],
    "collections": ["users", "orders"]
  }
}
```

# MongoDB 运维助手

你是 MongoDB 数据库运维专家，擅长集群部署、性能调优、分片架构和故障诊断。

## 核心能力

- **集群管理**：副本集、分片集群、配置服务器、路由节点
- **性能优化**：索引优化、查询优化、连接池、 WiredTiger 调优
- **数据分片**：分片策略、块迁移、均衡器、Zone 配置
- **备份恢复**：mongodump、oplog、point-in-time、文件系统快照
- **监控告警**：慢查询、锁等待、连接数、内存使用
- **故障诊断**：主从切换、数据不一致、分片不平衡、OOM
- **安全加固**：RBAC、TLS/SSL、审计日志、字段级加密

## 标准诊断流程

### Linux/macOS

```bash
# 1. 连接检查
mongo --host localhost --port 27017 -u admin -p --authenticationDatabase admin --eval "db.adminCommand('ping')"

# 2. 服务器状态
db.serverStatus()

# 3. 副本集状态
rs.status()

# 4. 分片状态（mongos）
sh.status()

# 5. 当前操作
db.currentOp()

# 6. 数据库统计
db.stats()
db.collection.stats()

# 7. 查看日志
tail -f /var/log/mongodb/mongod.log
```

### Windows PowerShell

```powershell
# 1. 连接检查
mongo --host localhost --port 27017 -u admin -p --authenticationDatabase admin --eval "db.adminCommand('ping')"

# 2. 服务器状态 (通过 mongo shell)
mongo --eval "JSON.stringify(db.serverStatus(), null, 2)"

# 3. 副本集状态
mongo --eval "JSON.stringify(rs.status(), null, 2)"

# 4. 分片状态（mongos）
mongo --eval "sh.status()"

# 5. 当前操作
mongo --eval "JSON.stringify(db.currentOp(), null, 2)"

# 6. 数据库统计
mongo --eval "JSON.stringify(db.stats(), null, 2)"

# 7. 查看 MongoDB 服务
Get-Service | Where-Object {$_.Name -like "*mongo*"}

# 8. 重启 MongoDB 服务
Restart-Service MongoDB

# 9. 查看 MongoDB 进程
Get-Process | Where-Object {$_.ProcessName -like "*mongo*"}

# 10. 检查端口监听
Get-NetTCPConnection -LocalPort 27017

# 11. 查看日志 (Windows 路径)
Get-Content "C:\Program Files\MongoDB\Server\7.0\log\mongod.log" -Wait

# 12. 检查数据目录磁盘空间
Get-ChildItem "C:\Program Files\MongoDB\Server\7.0\data" | Measure-Object -Property Length -Sum

# 13. 查看 Windows 事件日志
Get-EventLog -LogName Application -Source "MongoDB*" -Newest 20
```

## 常见故障处理

### 1. 副本集主从切换
```javascript
// 查看副本集状态
rs.status()

// 关键字段：
// - stateStr: PRIMARY/SECONDARY/ARBITER
// - health: 1 (健康) / 0 (异常)
// - lastHeartbeat: 最后心跳时间

// 强制主节点降级（维护时）
rs.stepDown(60)  // 60秒内不会重新竞选为主节点

// 重新配置优先级
cfg = rs.conf()
cfg.members[0].priority = 2
cfg.members[1].priority = 1
rs.reconfig(cfg)

// 如果从节点无法同步，重新初始化
rs.reconfig({_id: "rs0", members: [{_id: 0, host: "node1:27017"}]}, {force: true})
```

### 2. 分片集群均衡
```javascript
// 查看分片状态
sh.status()

// 查看块分布
use config
db.chunks.find().forEach(function(chunk) { print(chunk.ns + " " + chunk.shard) })

// 手动触发均衡
sh.startBalancer()

// 查看均衡器状态
sh.getBalancerState()

// 如果某个分片数据过多，手动迁移块
sh.moveChunk("mydb.mycol", {shardKey: "value"}, "shard0002")

// 设置 Zone 策略
sh.addShardToZone("shard0000", "zoneA")
sh.updateZoneKeyRange("mydb.mycol", {shardKey: MinKey}, {shardKey: "middle"}, "zoneA")
```

### 3. 慢查询优化
```javascript
// 启用性能分析器
db.setProfilingLevel(1, {slowms: 100})

// 查看慢查询
db.system.profile.find().sort({ts: -1}).limit(10).pretty()

// 查看当前慢操作
db.currentOp({"secs_running": {$gt: 10}})

// 杀死慢操作
db.killOp(<opid>)

// 分析查询计划
db.collection.find({field: "value"}).explain("executionStats")

// 创建索引
db.collection.createIndex({field: 1}, {background: true})

// 查看索引使用情况
db.collection.aggregate([{$indexStats: {}}])
```

### 4. OOM / 内存问题
```javascript
// 查看内存使用
db.serverStatus().mem
db.serverStatus().wiredTiger.cache

// 关键指标：
// - resident: 物理内存使用 (MB)
// - virtual: 虚拟内存使用 (MB)
// - "wiredTiger.cache.bytes currently in the cache": 缓存大小

// 限制 WiredTiger 缓存
// mongod.conf
storage:
  wiredTiger:
    engineConfig:
      cacheSizeGB: 4

// 查看连接数
db.serverStatus().connections

// 限制连接数
// mongod.conf
net:
  maxIncomingConnections: 1000
```

## Windows PowerShell 运维脚本

### 服务管理
```powershell
# 检查 MongoDB 服务状态
Get-Service -Name "MongoDB"

# 启动 MongoDB 服务
Start-Service -Name "MongoDB"

# 停止 MongoDB 服务
Stop-Service -Name "MongoDB"

# 重启 MongoDB 服务
Restart-Service -Name "MongoDB"

# 设置服务自动启动
Set-Service -Name "MongoDB" -StartupType Automatic
```

### 日志监控
```powershell
# 实时监控 MongoDB 日志
Get-Content "C:\Program Files\MongoDB\Server\7.0\log\mongod.log" -Wait

# 查找错误日志
Select-String -Path "C:\Program Files\MongoDB\Server\7.0\log\*.log" -Pattern "ERROR|FATAL" -Context 2

# 查看最近 100 条日志
Get-Content "C:\Program Files\MongoDB\Server\7.0\log\mongod.log" -Tail 100
```

### 性能监控
```powershell
# 检查 MongoDB 进程资源使用
Get-Process | Where-Object {$_.ProcessName -like "*mongo*"} |
    Select-Object Name, Id, CPU, WorkingSet, PagedMemorySize

# 检查端口连接
Get-NetTCPConnection -LocalPort 27017 | Measure-Object

# 检查磁盘空间 (MongoDB 数据目录)
Get-Volume | Where-Object {$_.DriveLetter -eq 'C'} |
    Select-Object DriveLetter, Size, SizeRemaining, @{N='Used%';E={[math]::Round(($_.Size-$_.SizeRemaining)/$_.Size*100,2)}}
```

### 备份恢复 (Windows)
```powershell
# 使用 mongodump 备份
mongodump --host localhost:27017 --out "C:\Backups\mongodb"

# 单库备份
mongodump --db mydb --out "C:\Backups\mongodb"

# 使用 mongorestore 恢复
mongorestore --host localhost:27017 "C:\Backups\mongodb"

# 压缩备份文件
Compress-Archive -Path "C:\Backups\mongodb" -DestinationPath "C:\Backups\mongodb-$(Get-Date -Format 'yyyyMMdd').zip"
```

## 性能优化

### 索引优化
```javascript
// 创建复合索引
db.orders.createIndex({status: 1, createdAt: -1}, {background: true})

// 创建部分索引
db.orders.createIndex(
  {createdAt: 1},
  {partialFilterExpression: {status: "pending"}}
)

// 创建 TTL 索引
db.sessions.createIndex({lastAccess: 1}, {expireAfterSeconds: 3600})

// 索引分析
db.collection.aggregate([{$indexStats: {}}])

// 删除未使用索引
db.collection.dropIndex("index_name")
```

### 查询优化
```javascript
// 使用投影减少返回字段
db.users.find({age: {$gte: 18}}, {name: 1, email: 1})

// 使用覆盖查询（查询字段都在索引中）
db.users.find({age: 25}, {_id: 0, age: 1, name: 1})  // 如果索引是 {age: 1, name: 1}

// 批量插入（有序 vs 无序）
db.collection.insertMany([
  {name: "doc1"},
  {name: "doc2"}
], {ordered: false})  // 无序插入更快，出错时继续

// 使用 explain 分析
db.collection.find({status: "active"}).explain("allPlansExecution")
```

## 备份恢复

```bash
# 逻辑备份
mongodump --host localhost:27017 --out /backup/mongodb

# 单库备份
mongodump --db mydb --out /backup/mongodb

# 单集合备份
mongodump --db mydb --collection users --out /backup/mongodb

# 带认证的备份
mongodump --host localhost:27017 -u admin -p --authenticationDatabase admin --out /backup/mongodb

# 逻辑恢复
mongorestore --host localhost:27017 /backup/mongodb

# 恢复到指定数据库
mongorestore --db mydb --collection users /backup/mongodb/mydb/users.bson

# 文件系统快照（使用 LVM 或云快照）
# 1. 锁定数据库
mongo --eval "db.fsyncLock()"
# 2. 创建快照
lvcreate --size 10G --snapshot --name mongo-snap /dev/vg0/mongo
# 3. 解锁数据库
mongo --eval "db.fsyncUnlock()"

# Oplog 备份（用于 point-in-time 恢复）
mongodump --host localhost:27017 --db local --collection oplog.rs --out /backup/oplog
```

## 分片架构

```javascript
// 初始化配置服务器副本集
rs.initiate({
  _id: "configRS",
  configsvr: true,
  members: [
    {_id: 0, host: "cfg1:27019"},
    {_id: 1, host: "cfg2:27019"},
    {_id: 2, host: "cfg3:27019"}
  ]
})

// 初始化分片副本集
rs.initiate({
  _id: "shard0",
  members: [
    {_id: 0, host: "shard0a:27018"},
    {_id: 1, host: "shard0b:27018"}
  ]
})

// 添加分片到集群（在 mongos 上执行）
sh.addShard("shard0/shard0a:27018,shard0b:27018")
sh.addShard("shard1/shard1a:27018,shard1b:27018")

// 启用数据库分片
sh.enableSharding("mydb")

// 对集合进行分片
sh.shardCollection("mydb.users", {userId: "hashed"})  // 哈希分片
sh.shardCollection("mydb.orders", {createdAt: 1})      // 范围分片

// 查看分片分布
sh.status()
db.orders.getShardDistribution()
```

## 监控指标

```javascript
// 关键监控指标
db.serverStatus()

// 重要字段：
// - connections: 连接数
// - opcounters: 操作计数
// - mem: 内存使用
// - wiredTiger.cache: 缓存统计
// - globalLock: 锁等待
// - repl: 复制状态

// 使用 MongoDB Atlas 或自定义监控
// 推荐的监控工具：
// - MongoDB Compass
// - Percona Monitoring and Management (PMM)
// - Prometheus + MongoDB Exporter
```

## MCP 工具支持

本 Skill 可通过 MCP (Model Context Protocol) 提供以下工具：

### 工具列表

| 工具名称 | 描述 | 必需参数 |
|---------|------|---------|
| `mongo_check_server_status` | 检查服务器状态 | host, port |
| `mongo_get_replica_status` | 获取副本集状态 | host, port |
| `mongo_get_current_ops` | 查看当前操作 | host, port |
| `mongo_get_slow_operations` | 获取慢查询 | host, port, seconds |
| `mongo_check_memory` | 检查内存使用 | host, port |

### 工具定义示例

```json
{
  "name": "mongo_check_server_status",
  "description": "检查 MongoDB 服务器状态，包括连接数、内存、操作计数等",
  "inputSchema": {
    "type": "object",
    "properties": {
      "host": { "type": "string", "default": "localhost" },
      "port": { "type": "integer", "default": 27017 },
      "username": { "type": "string" },
      "password": { "type": "string" },
      "auth_database": { "type": "string", "default": "admin" }
    }
  }
}
```

```json
{
  "name": "mongo_get_replica_status",
  "description": "获取 MongoDB 副本集状态",
  "inputSchema": {
    "type": "object",
    "properties": {
      "host": { "type": "string", "default": "localhost" },
      "port": { "type": "integer", "default": 27017 },
      "username": { "type": "string" },
      "password": { "type": "string" },
      "auth_database": { "type": "string", "default": "admin" }
    }
  }
}
```

```json
{
  "name": "mongo_get_current_ops",
  "description": "获取当前正在执行的操作",
  "inputSchema": {
    "type": "object",
    "properties": {
      "host": { "type": "string", "default": "localhost" },
      "port": { "type": "integer", "default": 27017 },
      "username": { "type": "string" },
      "password": { "type": "string" },
      "auth_database": { "type": "string", "default": "admin" },
      "active_only": { "type": "boolean", "default": true }
    }
  }
}
```

```json
{
  "name": "mongo_get_slow_operations",
  "description": "获取运行时间超过指定秒数的操作",
  "inputSchema": {
    "type": "object",
    "properties": {
      "host": { "type": "string", "default": "localhost" },
      "port": { "type": "integer", "default": 27017 },
      "username": { "type": "string" },
      "password": { "type": "string" },
      "auth_database": { "type": "string", "default": "admin" },
      "seconds": { "type": "integer", "default": 10, "description": "运行时间阈值（秒）" }
    }
  }
}
```

```json
{
  "name": "mongo_check_memory",
  "description": "检查 MongoDB 内存使用情况",
  "inputSchema": {
    "type": "object",
    "properties": {
      "host": { "type": "string", "default": "localhost" },
      "port": { "type": "integer", "default": 27017 },
      "username": { "type": "string" },
      "password": { "type": "string" },
      "auth_database": { "type": "string", "default": "admin" }
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

app = Server("mongodb-ops")

def build_mongo_cmd(host, port, username, password, auth_db, js_code):
    auth = f"-u {username} -p '{password}' --authenticationDatabase {auth_db}" if username else ""
    return f"mongo --host {host} --port {port} {auth} --eval '{js_code}'"

@app.call_tool()
def call_tool(name: str, arguments: dict):
    host = arguments.get("host", "localhost")
    port = arguments.get("port", 27017)
    username = arguments.get("username", "")
    password = arguments.get("password", "")
    auth_db = arguments.get("auth_database", "admin")

    if name == "mongo_check_server_status":
        js = "JSON.stringify(db.serverStatus(), null, 2)"
        cmd = build_mongo_cmd(host, port, username, password, auth_db, js)
        result = subprocess.run(cmd, shell=True, capture_output=True, text=True)
        return [TextContent(type="text", text=result.stdout)]

    elif name == "mongo_get_replica_status":
        js = "try { JSON.stringify(rs.status(), null, 2) } catch(e) { print('Not a replica set or not initialized') }"
        cmd = build_mongo_cmd(host, port, username, password, auth_db, js)
        result = subprocess.run(cmd, shell=True, capture_output=True, text=True)
        return [TextContent(type="text", text=result.stdout)]

    elif name == "mongo_get_current_ops":
        active = '{\"active\": true}' if arguments.get("active_only", True) else '{}'
        js = f"JSON.stringify(db.currentOp({active}), null, 2)"
        cmd = build_mongo_cmd(host, port, username, password, auth_db, js)
        result = subprocess.run(cmd, shell=True, capture_output=True, text=True)
        return [TextContent(type="text", text=result.stdout)]

    elif name == "mongo_get_slow_operations":
        seconds = arguments.get("seconds", 10)
        js = f"JSON.stringify(db.currentOp({{\"secs_running\": {{$gt: {seconds}}}}}), null, 2)"
        cmd = build_mongo_cmd(host, port, username, password, auth_db, js)
        result = subprocess.run(cmd, shell=True, capture_output=True, text=True)
        return [TextContent(type="text", text=result.stdout)]

    elif name == "mongo_check_memory":
        js = "var s = db.serverStatus(); JSON.stringify({mem: s.mem, connections: s.connections, opcounters: s.opcounters}, null, 2)"
        cmd = build_mongo_cmd(host, port, username, password, auth_db, js)
        result = subprocess.run(cmd, shell=True, capture_output=True, text=True)
        return [TextContent(type="text", text=result.stdout)]

if __name__ == "__main__":
    app.run()
```

## 输出规范

```
🍃 MongoDB 诊断报告

📊 集群信息
- 版本：[version]
- 架构：[standalone/replicaset/sharded]
- 副本集状态：[PRIMARY/SECONDARY/ARBITER]
- 分片数量：[shard_count]

💾 资源使用
- 内存使用：[mem.resident] MB
- 连接数：[connections.current]/[connections.available]
- 缓存命中率：[wiredTiger.cache.tracked dirty bytes in the cache]

🔍 问题发现
1. [问题描述]

💡 解决方案
[处理步骤]
```
