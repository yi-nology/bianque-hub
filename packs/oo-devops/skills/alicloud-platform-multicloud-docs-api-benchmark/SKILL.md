---
name: alicloud-platform-multicloud-docs-api-benchmark
description: 在阿里云、AWS、Azure、GCP、腾讯云、火山引擎和华为云之间对类似产品的文档和 API 文档进行基准测试。给定一个产品关键词，自动发现最新的官方文档/API 链接，一致地评分质量，并输出详细的优先级改进建议。
mode: on_demand
version: 1.1.0
maturity: experimental
---

# 多云产品文档/API 基准测试

当用户需要跨云文档/API 对比类似产品时使用此技能。

## 支持的云

- Alibaba Cloud
- AWS
- Azure
- GCP
- Tencent Cloud
- Volcano Engine
- Huawei Cloud

## 数据源策略

- `L0` (最高)：用户通过 `--<provider>-links` 固定的官方链接
- `L1`：机器可读的官方元数据/来源
  - GCP: Discovery API
  - AWS: API Models repository
  - Azure: REST API Specs repository
- `L2`：官方域名约束的网页发现回退
- `L3`：发现不足（低置信度）

## 工作流

运行基准测试脚本：

```bash
python skills/platform/docs/alicloud-platform-multicloud-docs-api-benchmark/scripts/benchmark_multicloud_docs_api.py --product "<产品关键词>"
```

示例：

```bash
python skills/platform/docs/alicloud-platform-multicloud-docs-api-benchmark/scripts/benchmark_multicloud_docs_api.py --product "serverless"
```

LLM 平台横评示例（百炼/Bedrock/Azure OpenAI/Vertex AI/混元/方舟/盘古）：

```bash
python skills/platform/docs/alicloud-platform-multicloud-docs-api-benchmark/scripts/benchmark_multicloud_docs_api.py --product "百炼" --preset "llm-platform"
```

如不传 `--preset`，脚本会根据关键词自动尝试匹配预设。

评分权重可通过 profile 切换（见 `references/scoring.json`）：

```bash
python skills/platform/docs/alicloud-platform-multicloud-docs-api-benchmark/scripts/benchmark_multicloud_docs_api.py --product "百炼" --preset "llm-platform" --scoring-profile "llm-platform"
```

## 可选：固定权威链接

自动发现可能遗漏页面。对于更严格的对比，手动传入官方链接：

```bash
python skills/platform/docs/alicloud-platform-multicloud-docs-api-benchmark/scripts/benchmark_multicloud_docs_api.py \
  --product "object storage" \
  --aws-links "https://docs.aws.amazon.com/AmazonS3/latest/userguide/Welcome.html" \
  --azure-links "https://learn.microsoft.com/azure/storage/blobs/"
```

可用的手动标志：

- `--alicloud-links`
- `--aws-links`
- `--azure-links`
- `--gcp-links`
- `--tencent-links`
- `--volcengine-links`
- `--huawei-links`

每个标志接受逗号分隔的 URL。

## 输出策略

所有产物必须写入：

`output/alicloud-platform-multicloud-docs-api-benchmark/`

每次运行：

- `benchmark_evidence.json`
- `benchmark_report.md`

## 报告指导

回答用户时：

1) 展示所有提供商的评分排名。
2) 突出主要差距（P0/P1/P2）和具体的修复措施。
3) 如果发现置信度低，请用户提供固定链接并重新运行。

## 参考

- Rubric: `references/review-rubric.md`
