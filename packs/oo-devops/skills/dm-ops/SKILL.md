---
name: dm-ops
description: 达梦数据库（DM）运维指南 - 安装部署、实例管理、用户权限、备份恢复、性能监控、SQL 调优、故障排查。用于达梦数据库的日常运维管理。
mode: on_demand
version: 1.1.0
maturity: experimental
---

# DM 达梦数据库运维

> **官网**: https://www.dameng.com/
> **文档**: https://eco.dameng.com/

## 安装部署

### 系统要求
- Linux: CentOS 7.6+ / RedHat 7.6+ / 麒麟 V10
- Windows: Windows Server 2016+
- 内存 ≥ 2GB
- 磁盘 ≥ 50GB

### Linux 安装

```bash
# 创建用户
groupadd dinstall
useradd -g dinstall -m -d /home/dmdba -s /bin/bash dmdba

# 设置密码
passwd dmdba

# 创建安装目录
mkdir -p /dm8
chown -R dmdba:dinstall /dm8

# 挂载 ISO 或解压安装包
mount -o loop dm8_20240116_x86_rh7.iso /mnt

# 图形化安装
su - dmdba
/mnt/DMInstall.bin

# 或命令行安装
/mnt/DMInstall.bin -i
```

### 静默安装配置
```xml
<!-- /dm8/install.xml -->
<?xml version="1.0"?>
<DATABASE>
  <LANGUAGE>zh</LANGUAGE>
  <TIME_ZONE>+08:00</TIME_ZONE>
  <KEY></KEY>
  <INSTALL_TYPE>0</INSTALL_TYPE>
  <INSTALL_PATH>/dm8</INSTALL_PATH>
  <INIT_DB>1</INIT_DB>
  <DB_PARAMS>
    <PATH>/dm8/data</PATH>
    <DB_NAME>DAMENG</DB_NAME>
    <INSTANCE_NAME>DMSERVER</INSTANCE_NAME>
    <PORT_NUM>5236</PORT_NUM>
    <SYSDBA_PWD>Dameng123</SYSDBA_PWD>
    <EXTENT_SIZE>16</EXTENT_SIZE>
    <PAGE_SIZE>8</PAGE_SIZE>
    <LOG_SIZE>256</LOG_SIZE>
    <CASE_SENSITIVE>1</CASE_SENSITIVE>
    <CHARSET>1</CHARSET>
  </DB_PARAMS>
</DATABASE>
```

### 初始化数据库
```bash
# 手动初始化实例
dminit PATH=/dm8/data DB_NAME=DAMENG INSTANCE_NAME=DMSERVER PORT_NUM=5236

# 查看帮助
dminit help
```

---

## 实例管理

### 启停实例

```bash
# 启动实例
dmserver /dm8/data/DAMENG/dm.ini

# 后台启动
dmserver /dm8/data/DAMENG/dm.ini &

# 使用服务启动
systemctl start DmServiceDMSERVER

# 停止实例（正常关闭）
systemctl stop DmServiceDMSERVER

# 强制停止（慎用）
systemctl kill -s 9 DmServiceDMSERVER
```

### 服务注册

```bash
# 注册系统服务
/dm8/script/root/dm_service_installer.sh -t dmserver -dm_ini /dm8/data/DAMENG/dm.ini -p DMSERVER

# 卸载服务
/dm8/script/root/dm_service_uninstaller.sh -n DmServiceDMSERVER
```

### 实例状态查看

```sql
-- 查看实例状态
SELECT * FROM V$INSTANCE;

-- 查看数据库状态
SELECT STATUS$ FROM V$DATABASE;

-- 查看启动时间
SELECT STARTUP_TIME FROM V$INSTANCE;

-- 查看实例名和端口
SELECT INSTANCE_NAME, HOST_NAME, PORT_NUM FROM V$INSTANCE;
```

### 多实例管理

```bash
# 创建第二个实例
dminit PATH=/dm8/data2 DB_NAME=DAMENG2 INSTANCE_NAME=DMSERVER2 PORT_NUM=5237

# 注册第二个服务
/dm8/script/root/dm_service_installer.sh -t dmserver -dm_ini /dm8/data2/DAMENG2/dm.ini -p DMSERVER2

# 启动第二个实例
systemctl start DmServiceDMSERVER2
```

---

## 用户权限

### 用户管理

```sql
-- 创建用户
CREATE USER "appuser" IDENTIFIED BY "Dameng@123"
DEFAULT TABLESPACE "MAIN"
DEFAULT INDEX TABLESPACE "MAIN";

-- 修改密码
ALTER USER "appuser" IDENTIFIED BY "NewPass@456";

-- 锁定/解锁用户
ALTER USER "appuser" ACCOUNT LOCK;
ALTER USER "appuser" ACCOUNT UNLOCK;

-- 删除用户
DROP USER "appuser" CASCADE;
```

### 权限授予

```sql
-- 授予 DBA 权限
GRANT DBA TO "appuser";

-- 授予对象权限
GRANT SELECT ON "SYSDBA"."T_EMPLOYEE" TO "appuser";
GRANT INSERT, UPDATE, DELETE ON "SYSDBA"."T_EMPLOYEE" TO "appuser";

-- 授予系统权限
GRANT CREATE TABLE TO "appuser";
GRANT CREATE VIEW TO "appuser";
GRANT CREATE PROCEDURE TO "appuser";

-- 授予列级权限
GRANT SELECT("EMPLOYEE_ID", "EMPLOYEE_NAME") ON "SYSDBA"."T_EMPLOYEE" TO "appuser";
```

### 角色管理

```sql
-- 创建角色
CREATE ROLE "app_role";

-- 给角色授权
GRANT CREATE TABLE TO "app_role";
GRANT CREATE VIEW TO "app_role";
GRANT SELECT ANY TABLE TO "app_role";

-- 授予角色给用户
GRANT "app_role" TO "appuser";

-- 设置用户默认角色
ALTER USER "appuser" DEFAULT ROLE "app_role";

-- 撤销角色
REVOKE "app_role" FROM "appuser";

-- 删除角色
DROP ROLE "app_role";
```

### 权限查询

```sql
-- 查看用户权限
SELECT * FROM DBA_SYS_PRIVS WHERE GRANTEE = 'appuser';
SELECT * FROM DBA_ROLE_PRIVS WHERE GRANTEE = 'appuser';
SELECT * FROM DBA_TAB_PRIVS WHERE GRANTEE = 'appuser';

-- 查看角色权限
SELECT * FROM ROLE_SYS_PRIVS WHERE ROLE = 'app_role';
SELECT * FROM ROLE_TAB_PRIVS WHERE ROLE = 'app_role';

-- 查看当前用户权限
SELECT * FROM USER_SYS_PRIVS;
SELECT * FROM USER_ROLE_PRIVS;
```

---

## 备份恢复

### 物理备份

#### 冷备份
```bash
# 停止数据库
systemctl stop DmServiceDMSERVER

# 备份数据目录
cp -r /dm8/data /backup/data_$(date +%Y%m%d)

# 启动数据库
systemctl start DmServiceDMSERVER
```

#### 热备份（联机备份）
```sql
-- 配置归档
ALTER DATABASE MOUNT;
ALTER DATABASE ARCHIVELOG;
ALTER DATABASE OPEN;

-- 查看归档配置
SELECT * FROM V$ARCHIVED_LOG;

-- 全库备份
BACKUP DATABASE FULL TO '/backup/full.bak' COMPRESSED;

-- 增量备份
BACKUP DATABASE INCREMENT WITH BACKUPDIR '/backup' TO '/backup/incr.bak';

-- 表空间备份
BACKUP TABLESPACE "MAIN" TO '/backup/main_ts.bak';

-- 表备份
BACKUP TABLE "SYSDBA"."T_EMPLOYEE" TO '/backup/employee.bak';
```

#### 备份管理
```sql
-- 查看备份集
SELECT * FROM V$BACKUPSET;

-- 删除备份集
SP_DROP_BACKUPSET('/backup/old.bak');

-- 添加备份描述
BACKUP DATABASE FULL TO '/backup/full.bak' BACKUPINFO '每周全量备份';
```

### 逻辑备份

```bash
# 全库导出
dexp SYSDBA/Dameng123 DIRECTORY=/backup FILE=full.dmp FULL=Y

# 按用户导出
dexp SYSDBA/Dameng123 DIRECTORY=/backup FILE=user.dmp OWNER=appuser

# 按表导出
dexp SYSDBA/Dameng123 DIRECTORY=/backup FILE=table.dmp TABLES=T_EMPLOYEE,T_DEPARTMENT

# 按模式导出
dexp SYSDBA/Dameng123 DIRECTORY=/backup FILE=schema.dmp SCHEMAS=app_schema

# 带条件导出
dexp SYSDBA/Dameng123 DIRECTORY=/backup FILE=filter.dmp TABLES=T_EMPLOYEE QUERY="WHERE SALARY > 5000"
```

### 数据恢复

#### 物理恢复
```sql
-- 整库恢复
cd /dm8/bin
./dmrestore INI_PATH=/dm8/data/DAMENG/dm.ini FILE=/backup/full.bak

-- 表空间恢复
RESTORE TABLESPACE "MAIN" FROM '/backup/main_ts.bak';

-- 时间点恢复（PITR）
RESTORE DATABASE FROM '/backup/full.bak' WITH ARCHIVEDIR '/dm8/arch' UNTIL TIME '2024-03-25 12:00:00';
```

#### 逻辑恢复
```bash
# 全库导入
dimp SYSDBA/Dameng123 DIRECTORY=/backup FILE=full.dmp FULL=Y

# 用户导入
dimp SYSDBA/Dameng123 DIRECTORY=/backup FILE=user.dmp OWNER=appuser

# 表导入（覆盖存在表）
dimp SYSDBA/Dameng123 DIRECTORY=/backup FILE=table.dmp TABLES=T_EMPLOYEE TABLE_EXISTS_ACTION=REPLACE

# 导入时忽略约束错误
dimp SYSDBA/Dameng123 DIRECTORY=/backup FILE=full.dmp FULL=Y IGNORE=Y
```

---

## 性能监控

### 系统监控

```sql
-- 查看数据库性能指标
SELECT * FROM V$SYSTEM_INFO;

-- 查看内存使用情况
SELECT * FROM V$BUFFERPOOL;
SELECT * FROM V$MEM_POOL;

-- 查看线程信息
SELECT * FROM V$THREADS;

-- 查看会话信息
SELECT * FROM V$SESSIONS WHERE STATE = 'ACTIVE';
```

### SQL 监控

```sql
-- 查看当前执行的 SQL
SELECT SESS_ID, SQL_TEXT, TIME_USED FROM V$SESSIONS WHERE SQL_TEXT IS NOT NULL;

-- 查看历史 SQL（需开启 SQL 日志）
SELECT * FROM V$SQL_HISTORY ORDER BY EXEC_TIME DESC;

-- 查看执行计划
EXPLAIN SELECT * FROM T_EMPLOYEE WHERE EMPLOYEE_ID = 1001;

-- 查看慢查询（超过 1 秒）
SELECT * FROM V$SQL_HISTORY WHERE TIME_USED > 1000000 ORDER BY TIME_USED DESC;
```

### 锁和阻塞

```sql
-- 查看当前锁
SELECT * FROM V$LOCK;

-- 查看阻塞会话
SELECT * FROM V$DEADLOCK_HISTORY;

-- 查看等待事件
SELECT * FROM V$SYSTEM_EVENT ORDER BY TOTAL_WAITS DESC;

-- 杀掉会话
SP_CLOSE_SESSION(SESSION_ID);
```

### 表空间监控

```sql
-- 查看表空间使用率
SELECT
    TABLESPACE_NAME,
    TOTAL_SIZE / 1024 / 1024 AS TOTAL_MB,
    FREE_SIZE / 1024 / 1024 AS FREE_MB,
    (TOTAL_SIZE - FREE_SIZE) / 1024 / 1024 AS USED_MB,
    ROUND((TOTAL_SIZE - FREE_SIZE) * 100.0 / TOTAL_SIZE, 2) AS USED_PCT
FROM DBA_TABLESPACES;

-- 查看数据文件
SELECT * FROM DBA_DATA_FILES;

-- 扩展表空间
ALTER TABLESPACE "MAIN" ADD DATAFILE '/dm8/data/DAMENG/MAIN02.DBF' SIZE 1024 AUTOEXTEND ON NEXT 100 MAXSIZE 20480;
```

---

## SQL 调优

### 索引优化

```sql
-- 创建索引
CREATE INDEX IDX_EMP_SALARY ON T_EMPLOYEE(SALARY);

-- 复合索引
CREATE INDEX IDX_EMP_DEPT_SAL ON T_EMPLOYEE(DEPARTMENT_ID, SALARY);

-- 查看索引使用情况
SELECT * FROM V$OBJECT_USAGE WHERE INDEX_NAME = 'IDX_EMP_SALARY';

-- 重建索引
ALTER INDEX IDX_EMP_SALARY REBUILD;

-- 收集统计信息
DBMS_STATS.GATHER_TABLE_STATS('SYSDBA', 'T_EMPLOYEE');

-- 查看表统计信息
SELECT * FROM DBA_TABLES WHERE TABLE_NAME = 'T_EMPLOYEE';
```

### SQL 提示

```sql
-- 使用索引提示
SELECT /*+ INDEX(T_EMPLOYEE IDX_EMP_SALARY) */ * FROM T_EMPLOYEE WHERE SALARY > 5000;

-- 使用并行提示
SELECT /*+ PARALLEL(4) */ * FROM LARGE_TABLE;

-- 使用 HASH JOIN
SELECT /*+ USE_HASH(t1 t2) */ * FROM T1 t1, T2 t2 WHERE t1.ID = t2.ID;

-- 禁用索引
SELECT /*+ NO_INDEX(T_EMPLOYEE) */ * FROM T_EMPLOYEE WHERE EMPLOYEE_ID = 1;
```

### 参数调优

```sql
-- 修改内存参数
SP_SET_PARA_VALUE(2, 'MEMORY_POOL', 2000);
SP_SET_PARA_VALUE(2, 'BUFFER_POOLS', 50);

-- 修改会话参数
ALTER SYSTEM SET 'SORT_BUF_SIZE' = 20000;

-- 查看参数
SELECT * FROM V$PARAMETER WHERE NAME LIKE '%BUFFER%';

-- 查看动态参数
SELECT * FROM V$DM_INI WHERE TYPE = 'SYS';
```

---

## 故障排查

### 日志分析

```bash
# 查看数据库日志
tail -f /dm8/log/dm_dmserver_202403.log

# 查看告警日志
grep "ERROR" /dm8/log/dm_dmserver_*.log

# 查看 SQL 日志
tail -f /dm8/log/log_commit.log
```

### 常见问题

#### 连接问题
```sql
-- 查看最大连接数
SELECT PARA_VALUE FROM V$DM_INI WHERE PARA_NAME = 'MAX_SESSIONS';

-- 修改最大连接数
SP_SET_PARA_VALUE(2, 'MAX_SESSIONS', 1000);

-- 查看当前连接数
SELECT COUNT(*) FROM V$SESSIONS;
```

#### 锁等待处理
```sql
-- 查找阻塞会话
SELECT
    s.SESS_ID,
    s.SQL_TEXT,
    s.CLNT_IP,
    l.BLOCKED,
    l.LTYPE
FROM V$SESSIONS s
JOIN V$LOCK l ON s.SESS_ID = l.TRX_ID
WHERE l.BLOCKED = 1;

-- 终止阻塞会话
SP_CLOSE_SESSION(139815801157128);
```

#### 空间不足
```sql
-- 查看表空间使用率
SELECT TABLESPACE_NAME,
       ROUND((USED_SPACE*8192)/1024/1024,2) AS USED_MB,
       ROUND((TABLESPACE_SIZE*8192)/1024/1024,2) AS TOTAL_MB,
       ROUND((USED_SPACE/TABLESPACE_SIZE)*100,2) AS USED_PCT
FROM DBA_TABLESPACES;

-- 扩展表空间
ALTER TABLESPACE "MAIN" RESIZE DATAFILE '/dm8/data/DAMENG/MAIN.DBF' TO 10240;
```

### 诊断工具

```bash
# 一致性检查
dmchk /dm8/data/DAMENG

# 查看数据文件信息
dmfldr SYSDBA/Dameng123 CONTROL='/dm8/check.ctl'

# 导出错误信息
dmrdc DUMP ERROR FROM '/dm8/log/dm_dmserver_202403.log' TO '/dm8/error.txt'
```

---

## 常用命令速查

| 操作 | 命令 |
|------|------|
| 启动实例 | `systemctl start DmServiceDMSERVER` |
| 停止实例 | `systemctl stop DmServiceDMSERVER` |
| 连接数据库 | `disql SYSDBA/Dameng123@localhost:5236` |
| 初始化实例 | `dminit PATH=/dm8/data DB_NAME=DAMENG` |
| 全量备份 | `BACKUP DATABASE FULL TO '/backup/full.bak'` |
| 逻辑导出 | `dexp SYSDBA/pwd FILE=full.dmp FULL=Y` |
| 逻辑导入 | `dimp SYSDBA/pwd FILE=full.dmp FULL=Y` |
| 查看版本 | `SELECT * FROM V$VERSION;` |
| 查看参数 | `SELECT * FROM V$PARAMETER;` |

---

## 最佳实践

- 生产环境定期执行全量备份，建议每日一次
- 开启归档模式以支持时间点恢复
- 定期收集统计信息以保持执行计划准确
- 关键表建立合适的索引，避免全表扫描
- 监控表空间使用率，超过 80% 及时扩容
- 定期清理历史 SQL 日志和审计日志
- 使用普通用户进行应用连接，避免使用 SYSDBA
- 定期修改数据库密码，遵守安全规范
