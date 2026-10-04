---
name: alicloud-ai-search-dashvector
description: 使用 Python SDK 通过 DashVector 构建向量检索。用于在 Claude Code/Codex 中创建 collection、upsert 文档以及运行带过滤条件的相似性搜索。
mode: on_demand
version: 1.1.0
maturity: experimental
---

Category: provider

# DashVector Vector Search

使用 DashVector 管理 collection 并执行向量相似性搜索，支持可选过滤条件和稀疏向量。

## 前置条件

- 安装 SDK（建议在 venv 中安装以避免 PEP 668 限制）：

```bash
python3 -m venv .venv
. .venv/bin/activate
python -m pip install dashvector
```
- 通过环境变量提供凭证和 endpoint：
  - `DASHVECTOR_API_KEY`
  - `DASHVECTOR_ENDPOINT`（cluster endpoint）

## 规范化操作

### 创建 collection
- `name` (str)
- `dimension` (int)
- `metric` (str: `cosine` | `dotproduct` | `euclidean`)
- `fields_schema`（可选的 field 类型 dict）

### Upsert 文档
- `docs` 列表，元素为 `{id, vector, fields}` 或 tuples
- 支持 `sparse_vector` 和多向量 collection

### 查询文档
- `vector` 或 `id`（二选一必填；若均为空，则仅应用 filter）
- `topk` (int)
- `filter`（类 SQL 的 where 子句）
- `output_fields`（field 名称列表）
- `include_vector` (bool)

## 快速开始（Python SDK）

```python
import os
import dashvector
from dashvector import Doc

client = dashvector.Client(
    api_key=os.getenv("DASHVECTOR_API_KEY"),
    endpoint=os.getenv("DASHVECTOR_ENDPOINT"),
)

# 1) 创建 collection
ret = client.create(
    name="docs",
    dimension=768,
    metric="cosine",
    fields_schema={"title": str, "source": str, "chunk": int},
)
assert ret

# 2) Upsert 文档
collection = client.get(name="docs")
ret = collection.upsert(
    [
        Doc(id="1", vector=[0.01] * 768, fields={"title": "Intro", "source": "kb", "chunk": 0}),
        Doc(id="2", vector=[0.02] * 768, fields={"title": "FAQ", "source": "kb", "chunk": 1}),
    ]
)
assert ret

# 3) 查询
ret = collection.query(
    vector=[0.01] * 768,
    topk=5,
    filter="source = 'kb' AND chunk >= 0",
    output_fields=["title", "source", "chunk"],
    include_vector=False,
)
for doc in ret:
    print(doc.id, doc.fields)
```

## 脚本快速开始

```bash
python skills/ai/search/alicloud-ai-search-dashvector/scripts/quickstart.py
```

环境变量：

- `DASHVECTOR_API_KEY`
- `DASHVECTOR_ENDPOINT`
- `DASHVECTOR_COLLECTION`（可选）
- `DASHVECTOR_DIMENSION`（可选）

可选参数：`--collection`、`--dimension`、`--topk`、`--filter`。

## Claude Code/Codex 注意事项

- 优先使用 `upsert` 实现幂等写入。
- 保持 `dimension` 与 embedding model 输出维度一致。
- 使用 filters 强制租户或数据集范围。
- 如果使用稀疏向量，在 upsert/query 时传入 `sparse_vector={token_id: weight, ...}`。

## 错误处理

- 401/403：无效的 `DASHVECTOR_API_KEY`
- 400：无效的 collection schema 或 dimension 不匹配
- 429/5xx：指数退避重试

## References

- DashVector Python SDK: `Client.create`、`Collection.upsert`、`Collection.query`

- Source list: `references/sources.md`
