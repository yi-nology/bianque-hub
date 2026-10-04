---
name: clscli
description: 1. 安装 clscli (Homebrew)：
mode: on_demand
version: 1.1.0
maturity: experimental
---

# CLS Skill

查询和分析腾讯云 CLS 日志。

## 设置
1. 安装 clscli (Homebrew)：
    ```bash
    brew tap dbwang0130/clscli
    brew install dbwang0130/clscli/clscli
    ```
2. 获取凭证和区域列表：https://cloud.tencent.com/document/api/614/56474
3. 设置环境变量（与腾讯云 API 通用参数相同）：
    ```bash
    export TENCENTCLOUD_SECRET_ID="your-secret-id"
    export TENCENTCLOUD_SECRET_KEY="your-secret-key"
    ```
4. 通过 `--region` 指定区域（例如 ap-guangzhou）。

## 用法

!重要：如果你不知道日志主题，请先列出主题。

### 列出日志主题
列出区域中的主题以确定查询/上下文要使用的 `--region` 和主题 ID。

```bash
clscli topics --region <region> [--topic-name name] [--logset-name name] [--logset-id id] [--limit 20] [--offset 0]
```
示例：`--output=json`, `--output=csv`, `-o topics.csv`

| 选项 | 必需 | 描述 |
|--------|----------|-------------|
| --region | 是 | CLS 区域，例如 ap-guangzhou |
| --topic-name | 否 | 按主题名称过滤（模糊匹配） |
| --logset-name | 否 | 按日志集名称过滤（模糊匹配） |
| --logset-id | 否 | 按日志集 ID 过滤 |
| --limit | 否 | 页面大小，默认 20，最大 100 |
| --offset | 否 | 分页偏移量，默认 0 |
| --output, -o | 否 | 输出：json、csv 或文件路径 |

输出列：Region, TopicId, TopicName, LogsetId, CreateTime, StorageType。

### 通过查询获取日志
```bash
clscli query -q "[query condition] | [SQL statement]" --region <region> -t <TopicId> --last 1h
```
示例：
- 时间：`--last 1h`, `--last 30m`；或 `--from`/`--to` (Unix ms)
- 多个主题：`--topics <id1>,<id2>` 或多个 `-t <id>`
- 自动分页和上限：`--max 5000`（分页直到 5000 条日志或 ListOver）
- 输出：`--output=json`, `--output=csv`, `-o result.json`（写入文件）

| 选项 | 必需 | 描述 |
|--------|----------|-------------|
| --region | 是 | CLS 区域，例如 ap-guangzhou |
| -q, --query | 是 | 查询条件或 SQL，例如 `level:ERROR` 或 `* \| select count(*) as cnt` |
| -t, --topic | -t/--topics 之一 | 单个日志主题 ID |
| --topics | -t/--topics 之一 | 逗号分隔的主题 ID，最多 50 个 |
| --last | --last/--from/--to 之一 | 时间范围，例如 1h, 30m, 24h |
| --from, --to | --last/--from/--to 之一 | 开始/结束时间 (Unix ms) |
| --limit | 否 | 每次请求的日志数，默认 100，最大 1000 |
| --max | 否 | 最大总日志数；非零时，自动分页直到达到或 ListOver |
| --output, -o | 否 | 输出：json、csv 或文件路径 |
| --sort | 否 | 排序：asc 或 desc，默认 desc |

#### 查询条件语法

支持两种语法：
- **CQL** (CLS Query Language)：CLS 特定的日志查询语法，易于使用，推荐。
- **Lucene**：开源 Lucene 语法；不是为日志搜索设计的，对特殊字符、大小写、通配符有更多限制；不推荐。

##### CQL 语法
| 语法 | 描述 |
|--------|-------------|
| `key:value` | 键值搜索；字段 (key) 包含 value 的日志，例如 `level:ERROR` |
| `value` | 全文搜索；包含 value 的日志，例如 `ERROR` |
| `AND` | 逻辑与，不区分大小写，例如 `level:ERROR AND pid:1234` |
| `OR` | 逻辑或，不区分大小写，例如 `level:ERROR OR level:WARNING`, `level:(ERROR OR WARNING)` |
| `NOT` | 逻辑非，不区分大小写，例如 `level:ERROR NOT pid:1234`, `level:ERROR AND NOT pid:1234` |
| `()` | 用于优先级分组，例如 `level:(ERROR OR WARNING) AND pid:1234`。**注意：没有括号时，AND 的优先级高于 OR。** |
| `"  "` | 短语搜索；双引号字符串，单词和顺序必须匹配，例如 `name:"john Smith"`。短语内无逻辑运算符。 |
| `'  '` | 短语搜索；单引号，与 `""` 相同；当短语包含双引号时使用，例如 `body:'user_name:"bob"'` |
| `*` | 通配符；零个或多个字符，例如 `host:www.test*.com`。无前缀通配符。 |
| `>`, `>=`, `<`, `<=`, `=` | 数值的范围运算符，例如 `status>400`, `status:>=400` |
| `\` | 转义；转义字符按字面意思处理。转义值中的空格、`:`, `()`, `>`, `=`, `<`, `"`, `'`, `*`。 |
| `key:*` | text：字段存在（任何值）。long/double：字段存在且为数字，例如 `response_time:*` |
| `key:""` | text：字段存在且为空。long/double：值不是数字或字段缺失，例如 `response_time:""` |

#### SQL 语句语法
| 语法 | 描述 |
|--------|-------------|
| SELECT | 从表中选择；来自与查询条件匹配的当前日志主题的数据 |
| AS | 列 (KEY) 的别名 |
| GROUP BY | 与聚合函数一起使用，按一列或多列 (KEY) 分组 |
| ORDER BY | 按 KEY 对结果集排序 |
| LIMIT | 限制行数，默认 100，最大 1M |
| WHERE | 过滤原始数据 |
| HAVING | 在 GROUP BY 之后、ORDER BY 之前过滤；WHERE 过滤原始数据 |
| 嵌套子查询 | 一个 SELECT 嵌套在另一个 SELECT 中，用于多步分析 |
| SQL 函数 | 更丰富的分析：IP 地理、时间格式、字符串拆分/连接、JSON 提取、数学、去重计数等。 |


### 描述日志上下文

检索给定日志周围的日志上下文。

```bash
clscli context <PkgId> <PkgLogId> --region <region> -t <TopicId>
```
示例：`--output=json`, `--output=csv`, `-o context.json`（写入文件）

| 选项 | 必需 | 类型 | 描述 | 示例 |
|--------|----------|------|-------------|---------|
| --region | 是 | String | CLS 区域 | ap-guangzhou |
| -t, --topic | 是 | String | 日志主题 ID | - |
| PkgId | 是 | String | 日志包 ID，即 SearchLog Results[].PkgId | 528C1318606EFEB8-1A7 |
| PkgLogId | 是 | Integer | 包内的索引，即 SearchLog Results[].PkgLogId | 65536 |
| --output, -o | 否 | - | 输出：json、csv 或文件路径 | - |
