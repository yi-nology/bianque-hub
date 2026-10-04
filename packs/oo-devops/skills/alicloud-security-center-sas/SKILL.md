---
name: alicloud-security-center-sas
description: 通过 OpenAPI/SDK 管理阿里云安全中心 (Sas) 资源。支持列出资源、创建或更新配置、查询状态与故障排查。触发词：Sas、安全中心(security center)、OpenAPI 管理、配置更新(configuration update)、安全监控(security monitoring)。
mode: on_demand
version: 1.1.0
maturity: experimental
---

Category: service

# 安全中心

使用阿里云 OpenAPI (RPC) 配合官方 SDK 或 OpenAPI Explorer 管理安全中心的资源。

## 工作流程

1) 确认区域、资源标识符和所需操作。
2) 发现 API 列表和所需参数（参见参考资料）。
3) 使用 SDK 或 OpenAPI Explorer 调用 API。
4) 使用 describe/list API 验证结果。

## AccessKey 优先级（必须遵循）

1) 环境变量：`ALICLOUD_ACCESS_KEY_ID` / `ALICLOUD_ACCESS_KEY_SECRET` / `ALICLOUD_REGION_ID`
区域策略：`ALICLOUD_REGION_ID` 是可选默认值。如果未设置，为任务决定最合理的区域；如果不清楚，询问用户。
2) 共享配置文件：`~/.alibabacloud/credentials`

## API 发现

- 产品代码：`Sas`
- 默认 API 版本：`2018-12-03`
- 使用 OpenAPI 元数据端点列出 API 并获取模式（参见参考资料）。

## 高频操作模式

1) 清单/列出：优先使用 `List*` / `Describe*` API 获取当前资源。
2) 变更/配置：优先使用 `Create*` / `Update*` / `Modify*` / `Set*` API 进行变更。
3) 状态/故障排除：优先使用 `Get*` / `Query*` / `Describe*Status` API 进行诊断。

## 最小可执行快速开始

在调用业务 API 之前使用元数据优先发现：

```bash
python scripts/list_openapi_meta_apis.py
```

可选覆盖：

```bash
python scripts/list_openapi_meta_apis.py --product-code <ProductCode> --version <Version>
```

脚本将 API 清单产物写入技能输出目录下。

## 输出策略

如果你需要保存响应或生成的产物，将它们写入：
`output/alicloud-security-center-sas/`

## 参考资料

- 来源：`references/sources.md`
