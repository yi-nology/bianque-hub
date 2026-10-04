---
name: db-readonly
description: 针对 MySQL 或 PostgreSQL 运行安全的只读查询，用于数据检查、报告和故障排查。当用户要求读取 tables、检查 schema、统计 rows、采样数据或导出查询结果而不修改数据时使用。
mode: on_demand
version: 1.1.0
maturity: experimental
---

# db-readonly

本 skill 仅用于数据库读取任务。

## 本 skill 能做什么

- 使用 connection env vars 连接 **PostgreSQL** 或 **MySQL**
- 仅执行 **SELECT / WITH / EXPLAIN** 查询
- 可选将输出保存为 CSV/TSV/JSON
- 阻止 risky SQL（`INSERT`、`UPDATE`、`DELETE`、`DROP`、`ALTER` 等）

## Connection env vars

### PostgreSQL

- `PGHOST`
- `PGPORT`（可选，默认 5432）
- `PGDATABASE`
- `PGUSER`
- `PGPASSWORD`

### MySQL

- `MYSQL_HOST`
- `MYSQL_PORT`（可选，默认 3306）
- `MYSQL_DATABASE`
- `MYSQL_USER`
- `MYSQL_PASSWORD`

## 运行

使用脚本：

- `scripts/db_readonly.sh postgres "SELECT now();"`
- `scripts/db_readonly.sh mysql "SELECT NOW();"`

导出示例：

- `scripts/db_readonly.sh postgres "SELECT * FROM users LIMIT 100" --format csv --out /tmp/users.csv`

## 安全规则

1. 拒绝非 read SQL。
2. 对探索性查询优先使用 `LIMIT`。
3. 当用户要求 updates/deletes/schema changes 时，询问明确确认，不要通过本 skill 运行。
4. 避免从 env vars 打印 secrets。

## Reference

- Query cookbook: `references/query-cookbook.md`
