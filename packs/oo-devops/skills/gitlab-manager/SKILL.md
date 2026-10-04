---
name: gitlab-manager
description: 通过 API 管理 GitLab 仓库(repository)、合并请求(merge request, MR)和问题(issue)。支持创建仓库、审查 MR 代码和跟踪问题。触发词：GitLab 仓库(repository)、合并请求(merge request)、问题跟踪(issue tracking)
mode: on_demand
version: 1.1.0
maturity: experimental
---

# GitLab Manager

此技能允许通过 API 与 GitLab.com 互动。

## 前提条件

- **GITLAB_TOKEN**: 必须在环境中设置具有 `api` 范围的个人访问令牌。

## 用法

使用提供的 Node.js 脚本与 GitLab 互动。

### 脚本位置
`scripts/gitlab_api.js`

### 命令

#### 1. 创建仓库
在 GitLab 中创建一个新项目。
```bash
./scripts/gitlab_api.js create_repo "<name>" "<description>" "<visibility>"
# 可见性: private (默认), public, internal
```

#### 2. 列出合并请求
列出特定项目的 MR。
```bash
./scripts/gitlab_api.js list_mrs "<project_path>" "[state]"
# 项目路径: 例如, "jorgermp/my-repo" (将自动进行 URL 编码)
# 状态: opened (默认), closed, merged, all
```

#### 3. 对合并请求发表评论
向特定 MR 添加评论 (note)。用于代码审查。
```bash
./scripts/gitlab_api.js comment_mr "<project_path>" <mr_iid> "<comment_body>"
```

#### 4. 创建问题
打开一个新问题。
```bash
./scripts/gitlab_api.js create_issue "<project_path>" "<title>" "<description>"
```

## 示例

**创建私有仓库:**
```bash
GITLAB_TOKEN=... ./scripts/gitlab_api.js create_repo "new-tool" "A cool new tool" "private"
```

**审查 MR:**
```bash
# 首先列出以查找 ID
GITLAB_TOKEN=... ./scripts/gitlab_api.js list_mrs "jorgermp/my-tool" "opened"
# 然后评论
GITLAB_TOKEN=... ./scripts/gitlab_api.js comment_mr "jorgermp/my-tool" 1 "Great work, but check indentation."
```
