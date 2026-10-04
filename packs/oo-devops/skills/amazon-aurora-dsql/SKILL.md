---
name: amazon-aurora-dsql
description: 使用 Aurora DSQL 构建 - 管理架构、执行查询，并处理 DSQL 特定要求的迁移。在开发可扩展或分布式数据库/应用程序或用户请求 DSQL 时使用。
mode: on_demand
version: 1.1.0
maturity: experimental
---

# Amazon Aurora DSQL 技能

Aurora DSQL 是一个无服务器、兼容 PostgreSQL 的分布式 SQL 数据库。此技能通过 MCP 工具提供直接的数据库交互、架构管理、迁移支持和多租户模式。

**关键能力：**
- 通过 MCP 工具直接执行查询
- 具有 DSQL 约束的架构管理
- 迁移支持和安全的架构演进
- 多租户隔离模式
- 基于 IAM 的身份验证

---

## 参考文件

根据需要加载这些文件以获取详细指导：

### [development-guide.md](references/development-guide.md)
**何时：** 在实施架构更改或数据库操作之前始终加载
**包含：** DDL 规则、连接模式、事务限制、安全最佳实践

### MCP：
#### [mcp-setup.md](mcp/mcp-setup.md)
**何时：** 在使用或更新 DSQL MCP 服务器时始终加载以获取指导
**包含：** 使用 2 种配置选项设置 DSQL MCP 服务器的说明，如 [.mcp.json](mcp/.mcp.json) 中所示
1. 仅文档工具
2. 数据库操作（需要集群端点）

#### [mcp-tools.md](mcp/mcp-tools.md)
**何时：** 需要详细的 MCP 工具语法和示例时加载
**包含：** 工具参数、详细示例、使用模式

### [language.md](references/language.md)
**何时：** 进行特定语言实现选择时必须加载
**包含：** Python/JS/Go/Java/Rust 的驱动选择、框架模式、连接代码

### [dsql-examples.md](references/dsql-examples.md)
**何时：** 查找具体实现示例时加载
**包含：** 代码示例、仓库模式、多租户实现

### [troubleshooting.md](references/troubleshooting.md)
**何时：** 调试错误或意外行为时加载
**包含：** 常见陷阱、错误消息、解决方案

### [onboarding.md](references/onboarding.md)
**何时：** 用户明确请求"开始使用 DSQL"或类似短语时
**包含：** 新用户的交互式分步指南

### [access-control.md](references/access-control.md)
**何时：** 创建数据库角色、授予权限、为应用程序设置架构或处理敏感数据时必须加载
**包含：** 作用域角色设置、IAM 到数据库角色映射、敏感数据的架构分离、角色设计模式

### [ddl-migrations.md](references/ddl-migrations.md)
**何时：** 尝试执行 DROP COLUMN、RENAME COLUMN、ALTER COLUMN TYPE 或 DROP CONSTRAINT 功能时必须加载
**包含：** 表重建模式、大表的批量迁移、数据验证

### [mysql-to-dsql-migrations.md](references/mysql-to-dsql-migrations.md)
**何时：** 从 MySQL 迁移到 DSQL 或将 MySQL DDL 转换为 DSQL 兼容等效项时必须加载
**包含：** MySQL 数据类型映射、DDL 操作转换、AUTO_INCREMENT/ENUM/SET/FOREIGN KEY 迁移模式、通过表重建实现 ALTER TABLE ALTER COLUMN 和 DROP COLUMN

---

## 可用的 MCP 工具

`aurora-dsql` MCP 服务器提供这些工具：

**数据库操作：**
1. **readonly_query** - 执行 SELECT 查询（返回字典列表）
2. **transact** - 在事务中执行 DDL/DML 语句（接受 SQL 语句列表）
3. **get_schema** - 获取特定表的结构

**文档与知识：**
4. **dsql_search_documentation** - 搜索 Aurora DSQL 文档
5. **dsql_read_documentation** - 读取特定文档页面
6. **dsql_recommend** - 获取 DSQL 最佳实践建议

**注意：** 没有 `list_tables` 工具。使用 `readonly_query` 查询 information_schema。

参见 [mcp-setup.md](mcp/mcp-setup.md) 获取详细设置说明。
参见 [mcp-tools.md](mcp/mcp-tools.md) 获取详细用法和示例。

---

## 可用的 CLI 脚本

用于集群管理和直接 psql 连接的 Bash 脚本。所有脚本位于 [scripts/](scripts/)。

**集群管理：**
- **create-cluster.sh** - 使用可选标签创建新的 DSQL 集群
- **delete-cluster.sh** - 带确认提示删除集群
- **list-clusters.sh** - 列出区域中的所有集群
- **cluster-info.sh** - 获取详细的集群信息

**数据库连接：**
- **psql-connect.sh** - 使用自动 IAM 身份验证令牌生成连接到 DSQL

**快速示例：**
```bash
./scripts/create-cluster.sh --region us-east-1
export CLUSTER=abc123def456
./scripts/psql-connect.sh
```

参见 [scripts/README.md](scripts/README.md) 获取详细用法。

---

## 快速开始

### 1. 列出表并浏览架构
```
使用 readonly_query 查询 information_schema 列出表
使用 get_schema 了解表结构
```

### 2. 查询数据
```
使用 readonly_query 执行 SELECT 查询
多租户应用始终在 WHERE 子句中包含 tenant_id
仔细验证输入（无参数化查询可用）
```

### 3. 执行架构更改
```
使用 transact 工具配合 SQL 语句列表
遵循每个事务一个 DDL 的规则
始终在单独的事务中使用 CREATE INDEX ASYNC
```

---

## 常见工作流

### 工作流 1：创建多租户架构

**目标：** 创建具有适当租户隔离的新表

**步骤：**
1. 使用 transact 创建带 tenant_id 列的主表
2. 在单独的 transact 调用中在 tenant_id 上创建异步索引
3. 为常见查询模式创建复合索引（单独的 transact 调用）
4. 使用 get_schema 验证架构

**关键规则：**
- 在所有表中包含 tenant_id
- 使用 CREATE INDEX ASYNC（永远不要同步）
- 每个 DDL 在自己的 transact 调用中：`transact(["CREATE TABLE ..."])`
- 将数组/JSON 存储为 TEXT

### 工作流 2：安全数据迁移

**目标：** 安全地添加带有默认值的新列

**步骤：**
1. 使用 transact 添加列：`transact(["ALTER TABLE ... ADD COLUMN ..."])`
2. 在单独的 transact 调用中批量填充现有行（少于 3,000 行）
3. 使用 readonly_query 配合 COUNT 验证迁移
4. 如需，使用 transact 为新列创建异步索引

**关键规则：**
- 先添加列，稍后填充
- 绝不在 ALTER TABLE 中添加 DEFAULT
- 在单独的 transact 调用中批量更新少于 3,000 行
- 每个 ALTER TABLE 在自己的事务中

### 工作流 3：应用层引用完整性

**目标：** 安全地插入/删除具有父子关系的记录

**INSERT 步骤：**
1. 使用 readonly_query 验证父项存在
2. 如果未找到父项则抛出错误
3. 使用 transact 配合父项引用插入子记录

**DELETE 步骤：**
1. 使用 readonly_query 检查依赖记录（COUNT）
2. 如果存在依赖项则返回错误
3. 如果安全则使用 transact 删除记录

### 工作流 4：带租户隔离的查询

**目标：** 检索限定于特定租户的数据

**步骤：**
1. 始终在 WHERE 子句中包含 tenant_id
2. 验证并清理 tenant_id 输入（无参数化查询可用！）
3. 使用 readonly_query 配合已验证的 tenant_id
4. 绝不允许跨租户数据访问

**关键规则：**
- 在构建 SQL 之前验证所有输入（SQL 注入风险！）
- 所有查询包含 WHERE tenant_id = 'validated-value'
- 在应用层拒绝跨租户访问
- 对租户 ID 使用白名单或正则表达式验证

### 工作流 5：设置作用域数据库角色

**目标：** 创建应用程序特定的数据库角色，而不是使用 `admin` 角色

**必须加载 [access-control.md](references/access-control.md) 获取详细指导。**

**步骤：**
1. 以 `admin` 连接（唯一应该使用 admin 的时候）
2. 使用 `CREATE ROLE <name> WITH LOGIN` 创建数据库角色
3. 为每个数据库角色创建具有 `dsql:DbConnect` 的 IAM 角色
4. 使用 `AWS IAM GRANT` 将数据库角色映射到 IAM 角色
5. 为敏感数据创建专用架构（例如 `users_schema`）
6. 按角色授予架构和表权限
7. 应用程序使用 `generate-db-connect-auth-token` 连接（不是 admin 变体）

**关键规则：**
- 应用程序连接始终使用作用域数据库角色
- 必须将用户 PII 和敏感数据放在专用架构中，而不是 `public`
- 应用程序 IAM 角色始终使用 `dsql:DbConnect`
- 应该为每个服务组件创建单独的角色（只读、读写、用户服务等）

### 工作流 6：表重建 DDL 迁移

**目标：** 使用表重建模式执行 DROP COLUMN、RENAME COLUMN、ALTER COLUMN TYPE 或 DROP CONSTRAINT。

**必须加载 [ddl-migrations.md](references/ddl-migrations.md) 获取详细指导。**

**步骤：**
1. 必须使用 `readonly_query` 验证表存在并获取行数
2. 必须使用 `get_schema` 获取当前架构
3. 必须使用 `transact` 创建具有所需结构的新表
4. 必须迁移数据（对于超过 3,000 行的表，以 500-1,000 行为批量）
5. 必须在继续之前验证行数匹配
6. 必须交换表：删除原始表，重命名新表
7. 必须使用 `CREATE INDEX ASYNC` 重新创建索引

**规则：**
- 对于超过 3,000 行的表必须使用批量处理
- 首选 500-1,000 行的批量以获得最佳吞吐量
- 必须在类型更改之前验证数据兼容性（如果不兼容则中止）
- 在新表验证之前不得删除原始表
- 交换表后必须使用 ASYNC 重新创建所有索引

### 工作流 6：MySQL 到 DSQL 架构迁移

**目标：** 将 MySQL 表架构和 DDL 操作迁移到 DSQL 兼容等效项，包括数据类型映射、ALTER TABLE ALTER COLUMN 和 DROP COLUMN 操作。

**必须加载 [mysql-to-dsql-migrations.md](references/mysql-to-dsql-migrations.md) 获取详细指导。**

**步骤：**
1. 必须将所有 MySQL 数据类型映射到 DSQL 等效项（例如 AUTO_INCREMENT → UUID/IDENTITY/SEQUENCE、ENUM → 带 CHECK 的 VARCHAR、JSON → TEXT）
2. 必须删除 MySQL 特定功能（ENGINE、FOREIGN KEY、ON UPDATE CURRENT_TIMESTAMP、FULLTEXT INDEX）
3. 必须为删除的功能实现应用层替代方案（引用完整性、时间戳更新）
4. 对于 `ALTER TABLE ... ALTER COLUMN col datatype` 或 `MODIFY COLUMN`：必须使用表重建模式
5. 对于 `ALTER TABLE ... DROP COLUMN col`：必须使用表重建模式
6. 必须将所有索引创建转换为单独事务中的 `CREATE INDEX ASYNC`
7. 必须在类型更改之前验证数据兼容性（如果不兼容则中止）

**规则：**
- 对于 ALTER COLUMN 和 DROP COLUMN 必须使用表重建模式（不直接支持）
- 必须用应用层引用完整性替换 FOREIGN KEY
- 必须用 VARCHAR 和 CHECK 约束替换 ENUM
- 必须用 TEXT（逗号分隔）替换 SET
- 必须用 TEXT 替换 JSON 列
- 必须将 AUTO_INCREMENT 转换为 UUID、IDENTITY 列或 SEQUENCE（不支持 SERIAL）
- 必须用 CHECK (col >= 0) 替换 UNSIGNED 整数
- 对于超过 3,000 行的表必须使用批量处理
- 在新表验证之前不得删除原始表

---

## 最佳实践

- **应该首先阅读指南** - 在进行架构更改之前检查 [development_guide.md](references/development-guide.md)
- **应该使用首选语言模式** - 检查 [language.md](references/language.md)
- **应该直接执行查询** - 优先使用 MCP 工具进行临时查询
- **必需：遵循 DDL 指南** - 参见 [DDL 规则](references/development-guide.md#schema-ddl-rules)
- **应重复生成新令牌** - 参见 [连接限制](references/development-guide.md#connection-rules)
- **始终使用 ASYNC 索引** - `CREATE INDEX ASYNC` 是强制性的
- **必须将数组/JSON 序列化为 TEXT** - 将数组/JSON 存储为 TEXT（逗号分隔、JSON.stringify）
- **始终批量处理少于 3,000 行** - 保持事务限制
- **必需：使用白名单、正则表达式和引号转义清理 SQL 输入** - 参见 [输入验证](mcp/mcp-tools.md#input-validation-critical)
- **必须遵循正确的应用层模式** - 当需要多租户隔离或应用引用完整性时；参见 [应用层模式](references/development-guide.md#application-layer-patterns)
- **必需使用 DELETE 进行截断** - DELETE 是唯一支持的截断操作
- **应该测试任何迁移** - 在生产环境之前在开发集群上验证 DDL
- **规划水平扩展** - DSQL 旨在优化大规模性能而不降低延迟；参见 [水平扩展](references/development-guide.md#horizontal-scaling-best-practice)
- **生产应用程序中应该使用连接池** - 参见 [连接池](references/development-guide.md#connection-pooling-recommended)
- **应该使用故障排除指南进行调试：** - 始终参考 [troubleshooting.md](references/troubleshooting.md) 中的资源和指南
- **应用程序始终使用作用域角色** - 使用 `dsql:DbConnect` 创建数据库角色；参见 [访问控制](references/access-control.md)

---

## 其他资源

- [Aurora DSQL 文档](https://docs.aws.amazon.com/aurora-dsql/latest/userguide/)
- [代码示例仓库](https://github.com/aws-samples/aurora-dsql-samples)
- [PostgreSQL 兼容性](https://docs.aws.amazon.com/aurora-dsql/latest/userguide/working-with-postgresql-compatibility.html)
- [IAM 身份验证指南](https://docs.aws.amazon.com/aurora-dsql/latest/userguide/using-database-and-iam-roles.html)
- [CloudFormation 资源](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-dsql-cluster.html)
