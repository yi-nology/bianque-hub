---
name: alicloud-security-cloudfw
description: 通过 OpenAPI/SDK 管理 Alibaba Cloud Cloud Firewall (Cloudfw)。用于列出资源、创建或更新配置、查询状态以及该产品的故障排查工作流。
mode: on_demand
version: 1.1.0
maturity: experimental
---

Category: service

# Cloud Firewall

使用 Alibaba Cloud OpenAPI (RPC) 配合官方 SDK 或 OpenAPI Explorer 管理 Cloud Firewall 的资源。

## 工作流

1) 确认 region、resource identifiers 和期望的操作。
2) 发现 API 列表和必需参数（见 references）。
3) 使用 SDK 或 OpenAPI Explorer 调用 API。
4) 使用 describe/list API 验证结果。

## AccessKey 优先级（必须遵循）

1) 环境变量：`ALICLOUD_ACCESS_KEY_ID` / `ALICLOUD_ACCESS_KEY_SECRET` / `ALICLOUD_REGION_ID`
Region policy：`ALICLOUD_REGION_ID` 是可选默认值。若未设置，则为任务决定最合理的 region；若不清楚，询问用户。
2) 共享配置文件：`~/.alibabacloud/credentials`

## API 发现

- Product code: `Cloudfw`
- 默认 API version: `2017-12-07`
- 使用 OpenAPI metadata endpoints 列出 API 并获取 schema（见 references）。

## 高频操作模式

1) 清单/列表：优先使用 `List*` / `Describe*` API 获取当前资源。
2) 变更/配置：优先使用 `Create*` / `Update*` / `Modify*` / `Set*` API 进行变更。
3) 状态/故障排查：优先使用 `Get*` / `Query*` / `Describe*Status` API 进行诊断。

## 最小可执行快速开始

在调用业务 API 前先进行 metadata-first 发现：

```bash
python scripts/list_openapi_meta_apis.py
```

可选覆盖：

```bash
python scripts/list_openapi_meta_apis.py --product-code <ProductCode> --version <Version>
```

该脚本将 API 清单产物写入 skill 输出目录下。

## 输出策略

如需保存响应或生成的产物，写入：
`output/alicloud-security-cloudfw/`

## References

- Sources: `references/sources.md`
