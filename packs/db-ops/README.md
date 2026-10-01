# db-ops：数据库诊断包（社区版）

MySQL 与 PostgreSQL 双专家 + 三技能 + 一链。全部方法论面向**只读采集 + 诊断判读**，
KILL 连接/跳错/VACUUM FULL 等高危动作一律走平台审批面。

| 资产 | 说明 |
|---|---|
| `skills/mysql-triage` | 五类分诊：连接打满、锁与事务、慢查询、容量与 binlog、缓冲池抖动（oo-devops mysql-ops/mysql-dba 面合并重写） |
| `skills/mysql-replication` | 复制链专项：双线程定位、延迟三源归因、GTID 断点、1032/1062 错误处置口径（重写） |
| `skills/pg-triage` | PG 分诊：idle 事务、锁等待链、膨胀与 vacuum、WAL 与复制槽滞留、查询面（oo-devops postgres 面重写） |
| `mysql-analyst/` | MySQL 诊断专家（ask-ops 只读面） |
| `pg-analyst/` | PostgreSQL 诊断专家（ask-ops 只读面） |
| `chain.yaml` | `workflow/db-quick-audit` 数据库例检链（MySQL→PG 两步，技能钉扎） |

## 与其他包的错位

- Redis 等 KV 缓存归 **mw-ops**；ES（日志/检索面）归 **obs-ops**；
- 中间件到数据库的连锁故障（如缓存击穿打穿 DB）由调用侧专家在结论里给联动线索，
  不跨包代诊。

## 工具面

无专用 MCP 依赖：mysql / psql 只读 SQL 经平台 `ask-ops` 采集面在目标运维主机执行，
凭证走主机侧已配置客户端（~/.my.cnf / .pgpass 等），对话不回显。

## 安装

```bash
go run ./cmd/bq-markettool install --url <本仓> --pack db-ops --api <扁鹊实例>
```

## 改造说明

源自 openocta 收割的 oo-devops 市场包。取数据库域高价值技能收割重写：剥离原版
环境变量注入与 MCP 假设，统一扁鹊契约。mongodb/tidb/oceanbase 等技能族暂未搬运，
按需在后续版本承接。
