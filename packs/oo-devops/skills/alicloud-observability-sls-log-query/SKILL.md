---
name: alicloud-observability-sls-log-query
description: 使用 query|analysis 语法与 Python SDK 在阿里云日志服务 (SLS) 中查询与排障日志。支持时间范围限定的日志搜索、错误调查与根因分析。触发词：SLS、日志查询(log query)、日志排障(log troubleshooting)、根因分析(root cause analysis)、query|analysis。
mode: on_demand
version: 1.1.0
maturity: experimental
---

Category: service

# SLS 日志查询与排障

使用 SLS 的 query|analysis 语法与 Python SDK 做日志检索、过滤与统计分析。

## 前置条件

- 安装 SDK（建议在虚拟环境中，避免 PEP 668 限制）：

```bash
python3 -m venv .venv
. .venv/bin/activate
python -m pip install -U aliyun-log-python-sdk
```
- 配置环境变量：
  - `ALIBABA_CLOUD_ACCESS_KEY_ID`
  - `ALIBABA_CLOUD_ACCESS_KEY_SECRET`
  - `SLS_ENDPOINT` (如 `cn-hangzhou.log.aliyuncs.com`)
  - `SLS_PROJECT`
  - `SLS_LOGSTORE`

## Query 组成

- 查询语句：用于过滤日志（如 `status:500`）。
- 分析语句：用于统计聚合，格式为 `查询语句|分析语句`。
- 示例：`* | SELECT status, count(*) AS pv GROUP BY status`

详细语法见 `references/query-syntax.md`。

## 快速开始 (Python SDK)

```python
import os
import time
from aliyun.log import LogClient, GetLogsRequest

client = LogClient(
    os.environ["SLS_ENDPOINT"],
    os.environ["ALIBABA_CLOUD_ACCESS_KEY_ID"],
    os.environ["ALIBABA_CLOUD_ACCESS_KEY_SECRET"],
)

project = os.environ["SLS_PROJECT"]
logstore = os.environ["SLS_LOGSTORE"]

query = "status:500"
start_time = int(time.time()) - 15 * 60
end_time = int(time.time())

request = GetLogsRequest(project, logstore, start_time, end_time, query=query)
response = client.get_logs(request)
for log in response.get_logs():
    print(log.contents)
```

## 脚本快速开始

```bash
python skills/observability/sls/alicloud-observability-sls-log-query/scripts/query_logs.py \
  --query "status:500" \
  --last-minutes 15
```

可选参数：`--project`, `--logstore`, `--endpoint`, `--start`, `--end`, `--last-minutes`, `--limit`。

## 排障脚本

```bash
python skills/observability/sls/alicloud-observability-sls-log-query/scripts/troubleshoot.py \
  --group-field status \
  --last-minutes 30 \
  --limit 20
```

可选参数：`--error-query`, `--group-field`, `--limit`，以及上述时间范围参数。

## 工作流

1) 确认 Logstore 已开启索引（未开启会导致查询/分析失败）。
2) 编写查询语句，必要时追加分析语句。
3) 通过 SDK 或脚本执行查询并查看结果。
4) 用 `limit` 控制返回行数，必要时缩小时间范围。

## 参考资料

- 语法与示例：`references/query-syntax.md`
- Python SDK 初始化与查询：`references/python-sdk.md`
- 排障模板：`references/templates.md`

- 来源列表：`references/sources.md`
