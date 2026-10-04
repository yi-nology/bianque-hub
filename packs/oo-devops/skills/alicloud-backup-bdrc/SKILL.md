---
name: alicloud-backup-bdrc
description: 通过 OpenAPI/SDK 管理阿里云备份与灾备中心 (BDRC) 资源。支持列出资源、创建或更新配置、查询状态与故障排查。触发词：BDRC、备份与灾备(backup and disaster recovery)、OpenAPI 管理、配置更新(configuration update)。
mode: on_demand
version: 1.1.0
maturity: experimental
---

Category: service

# 备份与灾备中心

使用阿里云 OpenAPI (RPC) 配合官方 SDK 或 OpenAPI Explorer 管理备份与灾备中心的资源。

## 工作流

1) 确认地域、资源标识符和所需操作。
2) 发现 API 列表和所需参数 (参见参考资料)。
3) 使用 SDK 或 OpenAPI Explorer 调用 API。
4) 使用 describe/list API 验证结果。

## AccessKey 优先级 (必须遵守)

1) 环境变量: `ALICLOUD_ACCESS_KEY_ID` / `ALICLOUD_ACCESS_KEY_SECRET` / `ALICLOUD_REGION_ID`
地域策略: `ALICLOUD_REGION_ID` 为可选默认值。如果未设置，请为任务选择最合理的地域；如果不确定，请询问用户。
2) 共享配置文件: `~/.alibabacloud/credentials`

## API 发现

- 产品代码: `BDRC`
- 默认 API 版本: `2023-08-08`
- 使用 OpenAPI 元数据端点列出 API 和获取模式定义 (参见参考资料)。

## 高频操作模式

1) 清单/列表: 优先使用 `List*` / `Describe*` API 获取当前资源。
2) 变更/配置: 优先使用 `Create*` / `Update*` / `Modify*` / `Set*` API 进行修改。
3) 状态/故障排查: 优先使用 `Get*` / `Query*` / `Describe*Status` API 进行诊断。

## 最小可执行快速入门

在调用业务 API 之前，先使用元数据优先发现：

```bash
python scripts/list_openapi_meta_apis.py
```

可选覆盖参数:

```bash
python scripts/list_openapi_meta_apis.py --product-code <ProductCode> --version <Version>
```

该脚本将 API 清单产物写入 skill 输出目录下。

## 输出策略

如果需要保存响应或生成的产物，请写入：
`output/alicloud-backup-bdrc/`

## 参考资料

- 来源: `references/sources.md`
