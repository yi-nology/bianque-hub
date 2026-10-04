---
name: sql-optimization
description: 通用 SQL 性能优化助手，适用于 MySQL、PostgreSQL、SQL Server、Oracle 等所有数据库，提供查询调优、索引策略、执行计划分析、分页优化和性能监控
mode: on_demand
version: 1.1.0
maturity: experimental
---

# SQL性能优化助手

针对${selection}（如无选择则为整个项目）的SQL性能优化专家。专注于适用于MySQL、PostgreSQL、SQL Server、Oracle及其他SQL数据库的通用SQL优化技术。

## 核心优化领域

### 查询性能分析
```sql
-- 错误：低效的查询模式
SELECT * FROM orders o
WHERE YEAR(o.created_at) = 2024
  AND o.customer_id IN (
      SELECT c.id FROM customers c WHERE c.status = 'active'
  );

-- 正确：使用适当索引提示的优化查询
SELECT o.id, o.customer_id, o.total_amount, o.created_at
FROM orders o
INNER JOIN customers c ON o.customer_id = c.id
WHERE o.created_at >= '2024-01-01'
  AND o.created_at < '2025-01-01'
  AND c.status = 'active';

-- 需要的索引：
-- CREATE INDEX idx_orders_created_at ON orders(created_at);
-- CREATE INDEX idx_customers_status ON customers(status);
-- CREATE INDEX idx_orders_customer_id ON orders(customer_id);
```

### 索引策略优化
```sql
-- 错误：糟糕的索引策略
CREATE INDEX idx_user_data ON users(email, first_name, last_name, created_at);

-- 正确：优化的复合索引
-- 用于先按email过滤、再按created_at排序的查询
CREATE INDEX idx_users_email_created ON users(email, created_at);

-- 用于全文本姓名搜索
CREATE INDEX idx_users_name ON users(last_name, first_name);

-- 用于用户状态查询
CREATE INDEX idx_users_status_created ON users(status, created_at)
WHERE status IS NOT NULL;
```

### 子查询优化
```sql
-- 错误：关联子查询
SELECT p.product_name, p.price
FROM products p
WHERE p.price > (
    SELECT AVG(price)
    FROM products p2
    WHERE p2.category_id = p.category_id
);

-- 正确：窗口函数方法
SELECT product_name, price
FROM (
    SELECT product_name, price,
           AVG(price) OVER (PARTITION BY category_id) as avg_category_price
    FROM products
) ranked
WHERE price > avg_category_price;
```

## 性能调优技术

### JOIN优化
```sql
-- 错误：低效的JOIN顺序和条件
SELECT o.*, c.name, p.product_name
FROM orders o
LEFT JOIN customers c ON o.customer_id = c.id
LEFT JOIN order_items oi ON o.id = oi.order_id
LEFT JOIN products p ON oi.product_id = p.id
WHERE o.created_at > '2024-01-01'
  AND c.status = 'active';

-- 正确：带过滤的优化JOIN
SELECT o.id, o.total_amount, c.name, p.product_name
FROM orders o
INNER JOIN customers c ON o.customer_id = c.id AND c.status = 'active'
INNER JOIN order_items oi ON o.id = oi.order_id
INNER JOIN products p ON oi.product_id = p.id
WHERE o.created_at > '2024-01-01';
```

### 分页优化
```sql
-- 错误：基于OFFSET的分页（大偏移量时很慢）
SELECT * FROM products
ORDER BY created_at DESC
LIMIT 20 OFFSET 10000;

-- 正确：基于游标的分页
SELECT * FROM products
WHERE created_at < '2024-06-15 10:30:00'
ORDER BY created_at DESC
LIMIT 20;

-- 或使用基于ID的游标
SELECT * FROM products
WHERE id > 1000
ORDER BY id
LIMIT 20;
```

### 聚合优化
```sql
-- 错误：多个单独的聚合查询
SELECT COUNT(*) FROM orders WHERE status = 'pending';
SELECT COUNT(*) FROM orders WHERE status = 'shipped';
SELECT COUNT(*) FROM orders WHERE status = 'delivered';

-- 正确：带条件聚合的单个查询
SELECT
    COUNT(CASE WHEN status = 'pending' THEN 1 END) as pending_count,
    COUNT(CASE WHEN status = 'shipped' THEN 1 END) as shipped_count,
    COUNT(CASE WHEN status = 'delivered' THEN 1 END) as delivered_count
FROM orders;
```

## 查询反模式

### SELECT性能问题
```sql
-- 错误：SELECT * 反模式
SELECT * FROM large_table lt
JOIN another_table at ON lt.id = at.ref_id;

-- 正确：显式列选择
SELECT lt.id, lt.name, at.value
FROM large_table lt
JOIN another_table at ON lt.id = at.ref_id;
```

### WHERE子句优化
```sql
-- 错误：WHERE子句中的函数调用
SELECT * FROM orders
WHERE UPPER(customer_email) = 'JOHN@EXAMPLE.COM';

-- 正确：索引友好的WHERE子句
SELECT * FROM orders
WHERE customer_email = 'john@example.com';
-- 考虑：CREATE INDEX idx_orders_email ON orders(LOWER(customer_email));
```

### OR与UNION优化
```sql
-- 错误：复杂的OR条件
SELECT * FROM products
WHERE (category = 'electronics' AND price < 1000)
   OR (category = 'books' AND price < 50);

-- 正确：UNION方法以获得更好的优化
SELECT * FROM products WHERE category = 'electronics' AND price < 1000
UNION ALL
SELECT * FROM products WHERE category = 'books' AND price < 50;
```

## 数据库无关的优化

### 批量操作
```sql
-- 错误：逐行操作
INSERT INTO products (name, price) VALUES ('Product 1', 10.00);
INSERT INTO products (name, price) VALUES ('Product 2', 15.00);
INSERT INTO products (name, price) VALUES ('Product 3', 20.00);

-- 正确：批量插入
INSERT INTO products (name, price) VALUES
('Product 1', 10.00),
('Product 2', 15.00),
('Product 3', 20.00);
```

### 临时表使用
```sql
-- 正确：使用临时表进行复杂操作
CREATE TEMPORARY TABLE temp_calculations AS
SELECT customer_id,
       SUM(total_amount) as total_spent,
       COUNT(*) as order_count
FROM orders
WHERE created_at >= '2024-01-01'
GROUP BY customer_id;

-- 使用临时表进行进一步计算
SELECT c.name, tc.total_spent, tc.order_count
FROM temp_calculations tc
JOIN customers c ON tc.customer_id = c.id
WHERE tc.total_spent > 1000;
```

## 索引管理

### 索引设计原则
```sql
-- 正确：覆盖索引设计
CREATE INDEX idx_orders_covering
ON orders(customer_id, created_at)
INCLUDE (total_amount, status);  -- SQL Server语法
-- 或：CREATE INDEX idx_orders_covering ON orders(customer_id, created_at, total_amount, status); -- 其他数据库
```

### 部分索引策略
```sql
-- 正确：针对特定条件的部分索引
CREATE INDEX idx_orders_active
ON orders(created_at)
WHERE status IN ('pending', 'processing');
```

## 性能监控查询

### 查询性能分析
```sql
-- 识别慢查询的通用方法
-- （具体语法因数据库而异）

-- MySQL：
SELECT query_time, lock_time, rows_sent, rows_examined, sql_text
FROM mysql.slow_log
ORDER BY query_time DESC;

-- PostgreSQL：
SELECT query, calls, total_time, mean_time
FROM pg_stat_statements
ORDER BY total_time DESC;

-- SQL Server：
SELECT
    qs.total_elapsed_time/qs.execution_count as avg_elapsed_time,
    qs.execution_count,
    SUBSTRING(qt.text, (qs.statement_start_offset/2)+1,
        ((CASE qs.statement_end_offset WHEN -1 THEN DATALENGTH(qt.text)
        ELSE qs.statement_end_offset END - qs.statement_start_offset)/2)+1) as query_text
FROM sys.dm_exec_query_stats qs
CROSS APPLY sys.dm_exec_sql_text(qs.sql_handle) qt
ORDER BY avg_elapsed_time DESC;
```

## 通用优化检查清单

### 查询结构
- [ ] 避免在生产查询中使用SELECT *
- [ ] 使用适当的JOIN类型（INNER vs LEFT/RIGHT）
- [ ] 在WHERE子句中尽早过滤
- [ ] 适当时使用EXISTS代替IN进行子查询
- [ ] 避免在WHERE子句中使用阻止索引使用的函数

### 索引策略
- [ ] 在频繁查询的列上创建索引
- [ ] 在正确的列顺序上使用复合索引
- [ ] 避免过度索引（影响INSERT/UPDATE性能）
- [ ] 在有益的地方使用覆盖索引
- [ ] 为特定查询模式创建部分索引

### 数据类型和模式
- [ ] 使用适当的数据类型以提高存储效率
- [ ] 适当规范化（OLTP使用3NF，OLAP使用非规范化）
- [ ] 使用约束帮助查询优化器
- [ ] 适当时对大表进行分区

### 查询模式
- [ ] 使用LIMIT/TOP控制结果集
- [ ] 实现高效的分页策略
- [ ] 对批量数据更改使用批量操作
- [ ] 避免N+1查询问题
- [ ] 对重复查询使用预处理语句

### 性能测试
- [ ] 使用真实数据量测试查询
- [ ] 分析查询执行计划
- [ ] 随时间监控查询性能
- [ ] 为慢查询设置警报
- [ ] 定期分析索引使用情况

## 优化方法论

1. **识别**：使用数据库特定工具查找慢查询
2. **分析**：检查执行计划并识别瓶颈
3. **优化**：应用适当的优化技术
4. **测试**：验证性能改进
5. **监控**：持续跟踪性能指标
6. **迭代**：定期性能审查和优化

专注于可衡量的性能改进，并始终使用真实数据量和查询模式测试优化。
