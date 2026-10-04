---
name: sql-server
description: Microsoft SQL Server 数据库日常运维管理。当用户需要运行 SQL 查询、检查 schema、管理表、监控性能、处理备份或通过 sqlcmd 或连接字符串管理 SQL Server 数据库时使用。
mode: on_demand
version: 1.1.0
maturity: experimental
---

# SQL Server

Microsoft SQL Server 管理 —— 查询、schema 检查、插入、更新和性能监控。

## 设置

```bash
export SQLCMDSERVER="localhost"
export SQLCMDUSER="sa"
export SQLCMDPASSWORD="yourpassword"   # 避免密码出现在 shell 历史记录中
export SQLCMDDBNAME="mydb"
```

使用 `sqlcmd` 连接：

```bash
sqlcmd -S "$SQLCMDSERVER" -U "$SQLCMDUSER" -P "$SQLCMDPASSWORD" -d "$SQLCMDDBNAME"
# 或内联：
sqlcmd -S localhost -U sa -P pass -d mydb
# 或带显式端口：
sqlcmd -S "tcp:myserver.database.windows.net,1433" -U user -P pass -d mydb
```

不进入交互式 shell 运行单个查询：

```bash
sqlcmd -S localhost -U sa -P pass -d mydb -Q "SELECT TOP 10 * FROM Orders"
```

运行 SQL 脚本文件：

```bash
sqlcmd -S localhost -U sa -P pass -d mydb -i script.sql
```

## 安装

```bash
# Linux (Debian/Ubuntu) —— mssql-tools
curl https://packages.microsoft.com/keys/microsoft.asc | sudo apt-key add -
curl https://packages.microsoft.com/config/ubuntu/22.04/prod.list \
  | sudo tee /etc/apt/sources.list.d/mssql-release.list
sudo apt-get update
sudo ACCEPT_EULA=Y apt-get install -y mssql-tools unixodbc-dev
echo 'export PATH="$PATH:/opt/mssql-tools/bin"' >> ~/.bashrc
source ~/.bashrc

# mssql-cli（带自动补全的交互式替代方案）
pip install mssql-cli
```

## 常用 sqlcmd 选项

| 选项 | 描述 |
|---|---|
| `-S` | 服务器（`host` 或 `host,port` 或 `tcp:host,port`） |
| `-U` | 用户名 |
| `-P` | 密码（优先使用 `SQLCMDPASSWORD` 环境变量） |
| `-d` | 数据库 |
| `-E` | 使用 Windows 集成认证 |
| `-Q` | 运行查询并退出 |
| `-q` | 运行查询，保持交互模式 |
| `-i` | 输入 SQL 文件 |
| `-o` | 输出文件 |
| `-h -1` | 移除列标题 |
| `-s ","` | 设置列分隔符（例如用于 CSV 输出） |
| `-W` | 移除列的尾部空格 |

## 常用操作

### 查询

```sql
SELECT TOP 10 * FROM Users;
SELECT column1, column2 FROM TableName WHERE condition ORDER BY column1 DESC;
```

### 插入 / 更新 / 删除

```sql
INSERT INTO Users (Name, Email) VALUES ('Alice', 'alice@example.com');
UPDATE Users SET Email = 'new@example.com' WHERE Id = 1;
DELETE FROM Users WHERE Id = 1;
```

### Schema 检查

```sql
-- 列出当前数据库中的所有表
SELECT TABLE_SCHEMA, TABLE_NAME FROM INFORMATION_SCHEMA.TABLES
WHERE TABLE_TYPE = 'BASE TABLE' ORDER BY TABLE_SCHEMA, TABLE_NAME;

-- 描述表（列、类型、可空性）
SELECT COLUMN_NAME, DATA_TYPE, CHARACTER_MAXIMUM_LENGTH, IS_NULLABLE
FROM INFORMATION_SCHEMA.COLUMNS
WHERE TABLE_NAME = 'Orders'
ORDER BY ORDINAL_POSITION;

-- 列出表上的索引
SELECT i.name AS index_name, i.type_desc, c.name AS column_name
FROM sys.indexes i
JOIN sys.index_columns ic ON i.object_id = ic.object_id AND i.index_id = ic.index_id
JOIN sys.columns c ON ic.object_id = c.object_id AND ic.column_id = c.column_id
WHERE OBJECT_NAME(i.object_id) = 'Orders';

-- 列出外键
SELECT
  fk.name AS fk_name,
  OBJECT_NAME(fk.parent_object_id) AS parent_table,
  c.name AS parent_column,
  OBJECT_NAME(fk.referenced_object_id) AS ref_table,
  rc.name AS ref_column
FROM sys.foreign_keys fk
JOIN sys.foreign_key_columns fkc ON fk.object_id = fkc.constraint_object_id
JOIN sys.columns c ON fkc.parent_object_id = c.object_id AND fkc.parent_column_id = c.column_id
JOIN sys.columns rc ON fkc.referenced_object_id = rc.object_id AND fkc.referenced_column_id = rc.column_id;
```

### Schema 变更

```sql
-- 创建表
CREATE TABLE Orders (
  Id        INT IDENTITY(1,1) PRIMARY KEY,
  UserId    INT NOT NULL REFERENCES Users(Id),
  Total     DECIMAL(10,2),
  CreatedAt DATETIME2 DEFAULT SYSDATETIME()
);

-- 添加 / 删除列
ALTER TABLE Orders ADD Status NVARCHAR(50) DEFAULT 'pending';
ALTER TABLE Orders DROP COLUMN Status;

-- 创建索引
CREATE INDEX IX_Orders_UserId ON Orders(UserId);
-- 非聚集覆盖索引：
CREATE NONCLUSTERED INDEX IX_Orders_Status ON Orders(Status) INCLUDE (Total, CreatedAt);
```

### 事务

```sql
BEGIN TRANSACTION;
  UPDATE Accounts SET Balance = Balance - 100 WHERE Id = 1;
  UPDATE Accounts SET Balance = Balance + 100 WHERE Id = 2;
COMMIT;
-- 或 ROLLBACK; 撤销
```

## 性能与监控

```sql
-- 当前运行的查询
SELECT
  r.session_id,
  r.status,
  r.cpu_time,
  r.total_elapsed_time,
  r.logical_reads,
  t.text AS query_text
FROM sys.dm_exec_requests r
CROSS APPLY sys.dm_exec_sql_text(r.sql_handle) t
WHERE r.session_id != @@SPID;

-- 消耗 CPU 最多的查询（来自计划缓存）
SELECT TOP 10
  qs.total_worker_time / qs.execution_count AS avg_cpu_time,
  qs.execution_count,
  SUBSTRING(st.text, (qs.statement_start_offset/2)+1,
    ((CASE qs.statement_end_offset WHEN -1 THEN DATALENGTH(st.text)
      ELSE qs.statement_end_offset END - qs.statement_start_offset)/2)+1) AS query_text
FROM sys.dm_exec_query_stats qs
CROSS APPLY sys.dm_exec_sql_text(qs.sql_handle) st
ORDER BY avg_cpu_time DESC;

-- 活动连接
SELECT session_id, login_name, status, host_name, program_name, database_id, cpu_time
FROM sys.dm_exec_sessions
WHERE is_user_process = 1;

-- 表大小
SELECT
  t.NAME AS table_name,
  p.rows AS row_count,
  SUM(a.total_pages) * 8 AS total_kb,
  SUM(a.used_pages) * 8 AS used_kb
FROM sys.tables t
JOIN sys.indexes i ON t.object_id = i.object_id
JOIN sys.partitions p ON i.object_id = p.object_id AND i.index_id = p.index_id
JOIN sys.allocation_units a ON p.partition_id = a.container_id
GROUP BY t.NAME, p.rows
ORDER BY total_kb DESC;

-- 解释查询计划
SET STATISTICS IO ON;
SET STATISTICS TIME ON;
SELECT * FROM Orders WHERE UserId = 42;

-- 或查看估计计划而不运行（SSMS/Azure Data Studio 语法）
-- 查询前缀：SET SHOWPLAN_TEXT ON; GO
```

## 备份与恢复

```bash
# 备份（通过 sqlcmd 使用 T-SQL）
sqlcmd -S localhost -U sa -P pass -Q \
  "BACKUP DATABASE [mydb] TO DISK = N'/var/opt/mssql/backup/mydb.bak' WITH NOFORMAT, INIT, STATS=10"

# 恢复
sqlcmd -S localhost -U sa -P pass -Q \
  "RESTORE DATABASE [mydb] FROM DISK = N'/var/opt/mssql/backup/mydb.bak' WITH REPLACE, STATS=10"
```

## 批量导入/导出 (BCP)

```bash
# 导出表到 CSV
bcp mydb.dbo.Orders out orders.csv -S localhost -U sa -P pass -c -t ','

# 导入 CSV 到表
bcp mydb.dbo.Orders in orders.csv -S localhost -U sa -P pass -c -t ','

# 导出查询结果到 CSV
bcp "SELECT * FROM mydb.dbo.Orders WHERE Status = 'pending'" queryout pending.csv \
  -S localhost -U sa -P pass -c -t ','
```

## 安全规则

1. **始终确认** 再运行 `DELETE`、`DROP` 或 `TRUNCATE`
2. **始终在 schema 迁移前备份**
3. **对多步数据更改使用事务**
4. 在大表上运行之前使用 `SET STATISTICS IO ON` / 执行计划预览查询成本
5. 优先选择有针对性的索引而非广泛的索引 —— 过度索引会减慢写入速度
6. 使用 `SQLCMDPASSWORD` 环境变量代替 `-P` 标志，以避免密码出现在 shell 历史记录中

## 参考

完整 `sqlcmd` 文档：<https://learn.microsoft.com/en-us/sql/tools/sqlcmd/sqlcmd-utility>
