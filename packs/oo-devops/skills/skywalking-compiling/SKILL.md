---
name: skywalking-compiling
description: 编译和构建 SkyWalking BanyanDB 项目。当用户要求编译、构建或为此项目生成代码时使用。
mode: on_demand
version: 1.1.0
maturity: experimental
---

# 编译 SkyWalking BanyanDB

按照以下步骤编译项目。

## 步骤 1：生成 protobuf 和 mock 代码

首先运行 `make generate`。这会从 `.proto` 定义重新生成 `.pb.go` 文件，
并通过 mockgen 生成 mock 文件。

每当以下情况时，此步骤是**必需的**：
- Proto 文件 (`.proto`) 已被添加或修改
- Go 接口 (mockgen 使用的) 已更改

如果 generate 失败，检查：
- `protoc` 和 Go protobuf plugins 已安装
- Proto 文件语法有效
- 接口签名匹配 mock 期望

## 步骤 2：构建所有二进制文件

运行 `make build` 编译所有项目组件：ui, banyand, bydbctl, mcp, fodc/agent, fodc/proxy。

二进制文件输出到每个组件的 `build/bin/dev/` 目录。

## 故障排除

- **.pb.go 中缺少字段**: Proto 文件已更新但未运行 `make generate`。先运行它。
- **Mock 生成失败**: 接口已更改但 mock 未重新生成。`make generate` 修复此问题。
- **导入错误**: 检查导入别名是否匹配 CLAUDE.md 约定 (例如 `commonv1`, `databasev1`)。

## 其他有用的 targets

- `make clean` — 清理所有构建产物
- `make clean-build` — 仅清理构建二进制文件
- `make lint` — 运行 linters
- `make test` — 运行测试
