---
name: azure-devops
description: 列出 Azure DevOps 项目、仓库和分支，创建 Pull Request，管理工作项，检查构建状态。适用于操作 Azure DevOps 资源、查看 PR 状态、查询项目结构或自动化 DevOps 工作流。
mode: on_demand
version: 1.1.0
maturity: experimental
---

# Azure DevOps Skill

列出项目、仓库、分支。创建 pull request。管理工作项。检查构建状态。

## 运行前检查有效配置，如果值缺失则询问用户！

**必需：**
- `AZURE_DEVOPS_PAT`: Personal Access Token
- `AZURE_DEVOPS_ORG`: Organization name

**如果 `~/.openclaw/openclaw.json` 中缺少值，代理应该：**
1. **询问** 用户缺失的 PAT 和/或组织名称
2. 将它们存储在 `~/.openclaw/openclaw.json` 的 `skills.entries["azure-devops"]` 下

### 示例配置

```json5
{
  skills: {
    entries: {
      "azure-devops": {
        apiKey: "YOUR_PERSONAL_ACCESS_TOKEN",  // AZURE_DEVOPS_PAT
        env: {
          AZURE_DEVOPS_ORG: "YourOrganizationName"
        }
      }
    }
  }
}
```

## 命令

### 列出项目

```bash
curl -s -u ":${AZURE_DEVOPS_PAT}" \
  "https://dev.azure.com/${AZURE_DEVOPS_ORG}/_apis/projects?api-version=7.1" \
  | jq -r '.value[] | "\(.name) - \(.description // "No description")"'
```

### 列出项目中的仓库

```bash
PROJECT="YourProject"
curl -s -u ":${AZURE_DEVOPS_PAT}" \
  "https://dev.azure.com/${AZURE_DEVOPS_ORG}/${PROJECT}/_apis/git/repositories?api-version=7.1" \
  | jq -r '.value[] | "\(.name) - \(.webUrl)"'
```

### 列出仓库中的分支

```bash
PROJECT="YourProject"
REPO="YourRepo"
curl -s -u ":${AZURE_DEVOPS_PAT}" \
  "https://dev.azure.com/${AZURE_DEVOPS_ORG}/${PROJECT}/_apis/git/repositories/${REPO}/refs?filter=heads/&api-version=7.1" \
  | jq -r '.value[] | .name | sub("refs/heads/"; "")'
```

### 创建 Pull Request

```bash
PROJECT="YourProject"
REPO_ID="repo-id-here"
SOURCE_BRANCH="feature/my-branch"
TARGET_BRANCH="main"
TITLE="PR Title"
DESCRIPTION="PR Description"

curl -s -u ":${AZURE_DEVOPS_PAT}" \
  -H "Content-Type: application/json" \
  -X POST \
  -d '{
    "sourceRefName": "refs/heads/'"${SOURCE_BRANCH}"'",
    "targetRefName": "refs/heads/'"${TARGET_BRANCH}"'",
    "title": "'"${TITLE}"'",
    "description": "'"${DESCRIPTION}"'"
  }' \
  "https://dev.azure.com/${AZURE_DEVOPS_ORG}/${PROJECT}/_apis/git/repositories/${REPO_ID}/pullrequests?api-version=7.1"
```

### 获取仓库 ID

```bash
PROJECT="YourProject"
REPO_NAME="YourRepo"
curl -s -u ":${AZURE_DEVOPS_PAT}" \
  "https://dev.azure.com/${AZURE_DEVOPS_ORG}/${PROJECT}/_apis/git/repositories/${REPO_NAME}?api-version=7.1" \
  | jq -r '.id'
```

### 列出 Pull Requests

```bash
PROJECT="YourProject"
REPO_ID="repo-id"
curl -s -u ":${AZURE_DEVOPS_PAT}" \
  "https://dev.azure.com/${AZURE_DEVOPS_ORG}/${PROJECT}/_apis/git/repositories/${REPO_ID}/pullrequests?api-version=7.1" \
  | jq -r '.value[] | "#\(.pullRequestId): \(.title) [\(.sourceRefName | sub("refs/heads/"; ""))] -> [\(.targetRefName | sub("refs/heads/"; ""))] - \(.createdBy.displayName)"'
```

## 注意

- Base URL: `https://dev.azure.com/${AZURE_DEVOPS_ORG}`
- API Version: `7.1`
- Auth: Basic Auth with empty username and PAT as password
- Never log or expose the PAT in responses
- Documentation: https://learn.microsoft.com/en-us/rest/api/azure/devops/
