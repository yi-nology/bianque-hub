---
name: bobagent-conventional-commits
description: 使用 Conventional Commits 规范格式化提交消息(commit message)。在创建提交、编写提交消息，或用户提及 commits、git commits、commit messages 时使用。确保提交遵循标准格式，以支持自动化工具、变更日志生成(changelog generation)与语义化版本控制(semantic versioning)。触发词：Conventional Commits、提交消息(commit message)、变更日志(changelog)、语义化版本(semantic versioning)、git 提交(git commit)。
mode: on_demand
version: 1.1.0
maturity: experimental
---

# Conventional Commits

根据 [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) 规范格式化所有提交消息。这支持自动化变更日志生成、语义化版本控制和更好的提交历史。

## 格式结构

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

## 提交类型

### 必需类型

- **`feat:`** - 一个新功能（与语义化版本控制中的 MINOR 相关）
- **`fix:`** - 一个 bug 修复（与语义化版本控制中的 PATCH 相关）

### 常见附加类型

- **`docs:`** - 仅文档变更
- **`style:`** - 代码风格变更（格式化、缺少分号等）
- **`refactor:`** - 既不修复 bug 也不添加新功能的代码重构
- **`perf:`** - 性能改进
- **`test:`** - 添加或更新测试
- **`build:`** - 构建系统或外部依赖变更
- **`ci:`** - CI/CD 配置变更
- **`chore:`** - 不修改 src 或 test 文件的其他变更
- **`revert:`** - 恢复之前的提交

## 范围

一个可选的范围提供关于代码库部分的额外上下文信息：

```
feat(parser): add ability to parse arrays
fix(auth): resolve token expiration issue
docs(readme): update installation instructions
```

## 描述

- 必须紧接在类型/范围后的冒号和空格之后
- 使用祈使语气 ("add feature" 而不是 "added feature" 或 "adds feature")
- 不要大写第一个字母
- 末尾没有句号
- 保持简洁（通常 50-72 个字符）

## 正文

- 提供额外上下文的可选较长描述
- 必须在描述后空一行开始
- 可以由多个段落组成
- 解释变更的 "what" 和 "why"，不是 "how"

## 破坏性变更

破坏性变更可以通过两种方式指示：

### 1. 在类型/范围中使用 `!`

```
feat!: send an email to the customer when a product is shipped
feat(api)!: send an email to the customer when a product is shipped
```

### 2. 使用 BREAKING CHANGE 页脚

```
feat: allow provided config object to extend other configs

BREAKING CHANGE: `extends` key in config file is now used for extending other config files
```

### 3. 两种方法

```
chore!: drop support for Node 6

BREAKING CHANGE: use JavaScript features not available in Node 6.
```

## 示例

### 简单功能

```
feat: add user authentication
```

### 带范围的功能

```
feat(auth): add OAuth2 support
```

### 带正文的 bug 修复

```
fix: prevent racing of requests

Introduce a request id and a reference to latest request. Dismiss
incoming responses other than from latest request.

Remove timeouts which were used to mitigate the racing issue but are
obsolete now.
```

### 破坏性变更

```
feat!: migrate to new API client

BREAKING CHANGE: The API client interface has changed. All methods now
return Promises instead of using callbacks.
```

### 文档更新

```
docs: correct spelling of CHANGELOG
```

### 带页脚的多段落正文

```
fix: prevent racing of requests

Introduce a request id and a reference to latest request. Dismiss
incoming responses other than from latest request.

Remove timeouts which were used to mitigate the racing issue but are
obsolete now.

Reviewed-by: Z
Refs: #123
```

## 指南

1. **始终使用类型** - 每个提交必须以类型后跟冒号和空格开始
2. **使用祈使语气** - 写作时如同完成句子 "If applied, this commit will..."
3. **具体** - 描述应清楚地传达变更了什么
4. **保持专注** - 每个提交一个逻辑变更
5. **在有帮助时使用范围** - 范围有助于在代码库中对变更进行分类
6. **记录破坏性变更** - 始终清楚地指示破坏性变更

## 语义化版本控制关联

- **`fix:`** → PATCH 版本提升 (1.0.0 → 1.0.1)
- **`feat:`** → MINOR 版本提升 (1.0.0 → 1.1.0)
- **BREAKING CHANGE** → MAJOR 版本提升 (1.0.0 → 2.0.0)

## 何时使用

将此格式用于：
- 所有 git 提交
- 提交消息生成
- Pull request 合并提交
- 当用户询问提交消息或 git 提交时

## 常见错误避免

❌ `Added new feature` (过去时，大写)
✅ `feat: add new feature` (祈使，小写)

❌ `fix: bug` (太模糊)
✅ `fix: resolve null pointer exception in user service`

❌ `feat: add feature` (冗余)
✅ `feat: add user profile page`

❌ `feat: Added OAuth support.` (过去时，句号)
✅ `feat: add OAuth support`
