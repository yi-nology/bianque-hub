---
name: alicloud-compute-fc-agentrun
description: 通过 OpenAPI 管理 Function Compute AgentRun 资源（runtime、sandbox、model、memory、credentials）。用于创建 runtime/endpoint、查询状态以及排查 AgentRun 工作流问题。
mode: on_demand
version: 1.1.0
maturity: experimental
---

Category: service

# Function Compute AgentRun (OpenAPI)

使用 AgentRun OpenAPI (ROA) 管理 runtime、sandbox、model service、memory 和 credentials。

## 前置条件

- 通过 RAM 用户获取 AccessKey（最小权限原则）。
- 选择正确的区域端点（见 `references/endpoints.md`）。如果不确定，为任务选择最合理的区域或询问用户。
- 使用 OpenAPI Explorer 或官方 SDK 避免手动签名（ROA 需要 SignatureV1）。

## 工作流

1) 选择区域端点 (`agentrun.cn-<region>.aliyuncs.com`)。
2) 创建 runtime → 发布版本 → 创建 runtime endpoint。
3) 如需要，创建 sandbox/template。
4) 根据需要配置 credentials 和 model services。
5) 查询资源以进行故障排查。

## API 分组

完整 API 列表和分组见 `references/api_overview.md`。

## 脚本快速开始

```bash
python skills/compute/fc/alicloud-compute-fc-agentrun/scripts/quickstart.py
```

环境变量：

- `AGENTRUN_ENDPOINT`
- `ALICLOUD_ACCESS_KEY_ID`
- `ALICLOUD_ACCESS_KEY_SECRET`
- `OUTPUT_DIR` (optional)

## Runtime 流脚本

```bash
AGENTRUN_RUNTIME_NAME="my-runtime" \
AGENTRUN_RUNTIME_ENDPOINT_NAME="my-runtime-endpoint" \
python skills/compute/fc/alicloud-compute-fc-agentrun/scripts/runtime_flow.py
```

环境变量：

- `AGENTRUN_ENDPOINT`
- `ALICLOUD_ACCESS_KEY_ID`
- `ALICLOUD_ACCESS_KEY_SECRET`
- `AGENTRUN_RUNTIME_NAME`
- `AGENTRUN_RUNTIME_ENDPOINT_NAME`
- `AGENTRUN_RUNTIME_DESC` (optional)
- `OUTPUT_DIR` (optional)

## 清理脚本

```bash
AGENTRUN_RUNTIME_ID="runtime-id" \
AGENTRUN_RUNTIME_ENDPOINT_ID="endpoint-id" \
python skills/compute/fc/alicloud-compute-fc-agentrun/scripts/cleanup_runtime.py
```

环境变量：

- `AGENTRUN_ENDPOINT`
- `ALICLOUD_ACCESS_KEY_ID`
- `ALICLOUD_ACCESS_KEY_SECRET`
- `AGENTRUN_RUNTIME_ID`
- `AGENTRUN_RUNTIME_ENDPOINT_ID`
- `OUTPUT_DIR` (optional)

## SDK 说明

SDK 获取指导见 `references/sdk.md`。

## 输出策略

如果存储任何生成的文件或响应，写入：
`output/compute-fc-agentrun/`。

## 参考

- API 概览和操作列表：`references/api_overview.md`
- 区域端点：`references/endpoints.md`
- SDK 指导：`references/sdk.md`

- 来源列表：`references/sources.md`
