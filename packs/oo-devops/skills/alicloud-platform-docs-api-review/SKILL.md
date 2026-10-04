---
name: alicloud-platform-docs-api-review
description: 按产品名称自动审查最新阿里云产品文档与 OpenAPI 文档，输出带证据与评分的优先级改进建议。用于审计产品文档质量、API 文档质量，或获取可操作的文档/API 优化建议。触发词：文档审查(document review)、API 审查(API review)、文档质量(document quality)、优化建议(optimization suggestions)、阿里云文档。
mode: on_demand
version: 1.1.0
maturity: experimental
---

# 阿里云产品文档 + API 文档审查器

当用户提供一个产品名称并要求进行端到端的文档/API 质量审查时使用此技能。

## 此技能的作用

1) 从最新的 OpenAPI 元数据中解析产品。
2) 获取默认版本的最新 API 文档。
3) 从官方产品页面发现产品/帮助文档链接。
4) 生成一份结构化审查报告，包含：
- 评分
- 证据
- 优先化建议 (P0/P1/P2)

## 工作流程

运行捆绑脚本：

```bash
python skills/platform/docs/alicloud-platform-docs-api-review/scripts/review_product_docs_and_api.py --product "<产品名或产品代码>"
```

示例：

```bash
python skills/platform/docs/alicloud-platform-docs-api-review/scripts/review_product_docs_and_api.py --product "ECS"
```

## 输出策略

所有生成的产物必须写入：

`output/alicloud-platform-docs-api-review/`

每次运行，脚本会创建：

- `review_evidence.json`
- `review_report.md`

## 报告指导

当回答用户时：

1) 首先说明解析出的产品 + 版本。
2) 总结评分和前 3 个问题。
3) 列出 P0/P1/P2 建议及具体行动。
4) 提供报告中使用的来源链接。

## 参考资料

- 审查标准：`references/review-rubric.md`
