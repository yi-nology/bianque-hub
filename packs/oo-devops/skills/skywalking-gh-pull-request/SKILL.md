---
name: skywalking-gh-pull-request
description: 为 SkyWalking BanyanDB 创建 GitHub pull request。当用户要求创建 PR、提交更改或打开 pull request 时使用。
mode: on_demand
version: 1.1.0
maturity: experimental
---

# 为 SkyWalking BanyanDB 创建 Pull Request

## 分支规则

**始终从新的分支创建 PR。** 永远不要直接推送到 `main`。
这与主 Apache SkyWalking 仓库 (https://github.com/apache/skywalking) 的约定相同。

```bash
git checkout -b <descriptive-branch-name>
```

## 创建 PR 前的本地检查

在推送前本地运行这些检查。它们镜像 CI `check` job 和 PR-blocking tests in `.github/workflows/ci.yml`。

### 必需：代码生成和构建

```bash
make generate
make build
```

### 必需：Linting 和格式化

```bash
make lint
make check
```

`make check` 验证格式化 (gofumpt)、go mod tidy，并确保没有未提交的生成文件差异。

**为什么是这个顺序：** `build` 必须在 `lint` 之前成功，以便生成的代码存在。`check` 在 linting 后验证一致性。

**在 `make lint` 通过后：** 如果 lint 引入了任何修复（例如自动格式化、字段对齐更正），
在运行 `make check` 之前提交这些更改。不要为这些 fixup 提交更新 CHANGES.md ——
只需将所有修改的跟踪文件添加到暂存区并提交，消息如 `chore: fix lint issues`。
然后在干净的树（clean tree）上运行 `make check`。

### 必需：License Headers

```bash
make license-check
```

所有源文件必须有 Apache 2.0 license headers。

### 必需：更新 CHANGES.md

在 `CHANGES.md`（在仓库根目录）的当前开发版本部分下添加一行条目。
将其放在适当的子部分下 (`### Features`, `### Bug Fixes`, 等)。

### 单元测试

运行这些测试包。如果用户的机器有足够的核心，每个都可以并行运行，但顺序运行也可以：

```bash
make test-ci PKG=./banyand/...
make test-ci PKG=./bydbctl/...
make test-ci PKG=./pkg/...
make test-ci PKG=./fodc/...
```

CI 使用这些选项：`--vv --fail-fast --label-filter \!slow` 以及 coverage flags。
对于本地运行，使用简化形式，除非用户要求完整的 CI 同等性：

```bash
TEST_CI_OPTS="--vv --fail-fast --label-filter \!slow" make test-ci PKG=./banyand/...
```

### 集成测试

单元测试通过后运行这些：

```bash
make test-ci PKG=./test/integration/standalone/...
make test-ci PKG=./test/integration/distributed/...
```

## 本地不实用的检查

这些只在 CI 中运行 —— 不需要在本地运行：
- **e2e tests** — 需要 Docker + OAP stack (90 min timeout)
- **fodc-e2e tests** — 需要 Kind Kubernetes cluster
- **dependency-review** — GitHub-specific action
- **slow/flaky/property-repair tests** — 定时运行，不阻塞 PR

## 常见问题

- **`make lint` 因 field alignment 错误失败**: Linter 报告字段排序次优的 structs
  (例如 `fieldalignment: struct with X pointer bytes could be Y`)。用以下命令自动修复：
  ```bash
  ~/go/bin/fieldalignment -fix ./path/to/package/...
  ```
  解析 lint 输出以查找哪些包有对齐问题，在这些包上运行 `fieldalignment -fix`，
  然后重新运行 `make lint` 以确认它们已解决。

- **`make lint` 因 formatting 错误失败**: 运行 `make format` 自动修复，然后重新运行 `make lint` 确认。
- **测试超时**: 集成测试可能很慢；如果超时添加 `TEST_CI_OPTS="--timeout 60m"`
- **缺少工具**: `make check-req` 会告诉您缺少什么
- **`make lint` 或 `make check` 因 `buf: not found` 失败**: 通过运行 `make -C api generate` 自动安装 `buf`，然后重试。
- **`make build` 失败**: 检查生成的文件是否与 `make generate` 保持最新

## 创建 PR

```bash
git push -u origin <branch-name>
gh pr create --title "<title>" --body "<body>"
```

遵循标准 PR 格式，包含摘要和测试计划。
