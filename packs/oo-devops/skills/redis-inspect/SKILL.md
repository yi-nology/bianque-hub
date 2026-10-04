---
name: redis-inspect
description: 检查 Redis 缓存键、值和 TTL 以进行调试。支持主缓存和系统缓存。用于调试缓存问题、检查缓存值和监控缓存状态。默认只读。
mode: on_demand
version: 1.1.0
maturity: experimental
---

# Redis 缓存检查器

使用此技能检查 Redis 缓存状态以进行调试。

## 运行命令

```bash
node .claude/skills/redis-inspect/query.mjs <command> [options]
```

### 命令

| 命令 | 描述 |
|---------|-------------|
| `get <key>` | 获取字符串值 |
| `keys <pattern>` | 查找匹配模式的键（使用 * 作为通配符） |
| `ttl <key>` | 获取 TTL（-1 = 无过期，-2 = 未找到） |
| `type <key>` | 获取键的类型 |
| `exists <key>` | 检查键是否存在 |
| `hgetall <key>` | 从 hash 获取所有字段 |
| `hget <key> <field>` | 获取特定的 hash 字段 |
| `scard <key>` | 获取 set 基数（计数） |
| `smembers <key>` | 获取所有 set 成员 |
| `llen <key>` | 获取 list 长度 |
| `lrange <key>` | 获取 list 元素 |
| `del <key>` | 删除键（需要 --writable） |
| `info` | 获取 Redis 服务器信息 |

### 选项

| 标志 | 描述 |
|------|-------------|
| `--sys` | 使用系统缓存而非主缓存 |
| `--writable` | 允许写操作（del 必需） |
| `--json` | 输出原始 JSON |
| `--limit <n>` | 限制结果（默认：100） |

## 缓存类型

项目有两个 Redis 实例：

| 缓存 | 标志 | 环境变量 | 用途 |
|-------|------|--------------|---------|
| **主缓存** | （默认） | `REDIS_URL` | 常规缓存，集群模式，可丢失 |
| **系统缓存** | `--sys` | `REDIS_SYS_URL` | 持久系统值，单节点 |

### 主缓存（默认）
常规应用缓存。如果丢失，这里的数据可以重新生成。
- 用户会话
- 缓存的查询
- 临时数据
- 速率限制计数器

### 系统缓存（--sys）
持久系统配置和状态。更关键的数据。
- 功能标志
- 生成限制/状态
- 系统权限
- 作业状态
- 事件配置

## 示例

```bash
# 查找匹配模式的键
node .claude/skills/redis-inspect/query.mjs keys "user:*" --limit 20
node .claude/skills/redis-inspect/query.mjs keys "packed:caches:*"

# 获取值
node .claude/skills/redis-inspect/query.mjs get "session:data2:123456"

# 检查系统缓存值
node .claude/skills/redis-inspect/query.mjs --sys get "system:features"
node .claude/skills/redis-inspect/query.mjs --sys hgetall "system:entity-moderation"

# 检查 TTL
node .claude/skills/redis-inspect/query.mjs ttl "generation:count:123"

# 检查 hash
node .claude/skills/redis-inspect/query.mjs hgetall "packed:caches:cosmetics"
node .claude/skills/redis-inspect/query.mjs hget "system:entity-moderation" "entities"

# 检查 set 大小
node .claude/skills/redis-inspect/query.mjs scard "queues:seen-images"

# 获取服务器信息
node .claude/skills/redis-inspect/query.mjs info
node .claude/skills/redis-inspect/query.mjs --sys info
```

## 常见键模式

### 主缓存
| 模式 | 描述 |
|---------|-------------|
| `user:*` | 用户数据 |
| `session:*` | 会话数据 |
| `packed:caches:*` | 打包/压缩的缓存数据 |
| `packed:user:*` | 打包的用户缓存 |
| `generation:*` | 生成相关的缓存 |
| `tag:*` | 标签缓存 |

### 系统缓存
| 模式 | 描述 |
|---------|-------------|
| `system:*` | 系统配置 |
| `generation:*` | 生成限制/状态 |
| `download:limits` | 下载限制 |
| `job:*` | 作业状态 |
| `event:*` | 事件配置 |
| `new-order:*` | New Order 游戏状态 |
| `daily-challenge:*` | 每日挑战配置 |

## 调试技巧

```bash
# 检查用户会话是否存在
node .claude/skills/redis-inspect/query.mjs keys "session:data2:*" --limit 10

# 检查生成状态
node .claude/skills/redis-inspect/query.mjs --sys get "generation:status"

# 检查功能标志
node .claude/skills/redis-inspect/query.mjs --sys hgetall "system:features"

# 检查缓存内存使用情况
node .claude/skills/redis-inspect/query.mjs info
```

## 写操作

写操作需要 `--writable` 标志和用户批准：

```bash
# 删除特定键（需要批准）
node .claude/skills/redis-inspect/query.mjs del "some:key" --writable
```

**重要**：在使用 `--writable` 之前始终询问用户许可。
