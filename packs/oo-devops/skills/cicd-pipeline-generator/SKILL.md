---
name: cicd-pipeline-generator
description: 生成并配置 CI/CD 流水线文件，支持 GitHub Actions、GitLab CI、CircleCI 等平台。适用于 Node.js/Next.js 应用的自动化构建(build)、测试(test)、部署(deployment)到 Vercel、Netlify 或 AWS。触发词：CI/CD、流水线(pipeline)、GitHub Actions、GitLab CI、部署(deployment)。
mode: on_demand
version: 1.1.0
maturity: experimental
---

# CI/CD 流水线生成器

## 概述

为各种平台（GitHub Actions、GitLab CI、CircleCI、Jenkins）生成生产就绪的 CI/CD 流水线配置文件。此技能提供模板和指导，用于设置处理现代 web 应用（特别是 Node.js/Next.js 项目）的 linting、测试、构建和部署的自动化工作流。

## 核心功能

### 1. 平台选择

基于项目需求选择适当的 CI/CD 平台：

- **GitHub Actions**: 最适合具有原生集成的 GitHub 托管项目
- **GitLab CI/CD**: 理想用于具有复杂流水线需求的 GitLab 仓库
- **CircleCI**: 针对 Docker 工作流和快速构建时间优化
- **Jenkins**: 适用于自托管、高度可定制的环境

参考 `references/platform-comparison.md` 获取详细的平台比较、优缺点和用例建议。

### 2. 流水线配置生成

遵循以下原则生成流水线配置：

#### 流水线阶段

使用这些标准阶段构建流水线：

1. **安装依赖**
   - 从仓库检出代码
   - 设置运行时环境 (Node.js 版本)
   - 恢复缓存的依赖
   - 使用 `npm ci` 安装依赖
   - 缓存依赖以供将来运行

2. **Lint**
   - 运行 ESLint 进行代码质量检查
   - 运行 TypeScript 类型检查
   - 在 linting 错误时快速失败

3. **测试**
   - 执行单元测试
   - 执行集成测试
   - 生成代码覆盖率报告
   - 上传覆盖率到报告服务 (Codecov, Coveralls)

4. **构建**
   - 创建生产构建
   - 验证构建成功
   - 存储构建产物

5. **部署**
   - 部署到 staging (develop 分支)
   - 部署到 production (main 分支)
   - 运行部署后冒烟测试

#### 缓存策略

实施有效的缓存以加速构建：

```yaml
# 基于 package-lock.json 缓存 node_modules
cache:
  key: ${{ hashFiles('package-lock.json') }}
  paths:
    - node_modules/
    - .npm/
```

#### 环境变量

配置必要的环境变量：
- `NODE_ENV`: 构建时设置为 `production`
- 平台特定的 tokens: 作为 secrets 存储
- 构建时变量: 传递给构建过程

### 3. 模板使用

使用来自 `assets/` 目录的提供的模板：

**GitHub Actions 模板** (`assets/github-actions-nodejs.yml`)：
- 带有 lint、test、build、deploy 的多 job 工作流
- 多个 Node.js 版本的矩阵构建（可选）
- Vercel 部署集成
- 产物上传
- 代码覆盖率报告

**GitLab CI 模板** (`assets/gitlab-ci-nodejs.yml`)：
- 多阶段流水线
- 依赖缓存
- 手动生产部署
- 自动 staging 部署
- 覆盖率报告

使用模板：
1. 复制适当的模板文件
2. 放置在正确的位置：
   - GitHub Actions: `.github/workflows/ci.yml`
   - GitLab CI: `.gitlab-ci.yml`
3. 自定义部署目标、环境变量和分支名称
4. 将必需的 secrets 添加到平台设置

### 4. 部署配置

#### Vercel 部署

对于 GitHub Actions：
```yaml
- uses: amondnet/vercel-action@v25
  with:
    vercel-token: ${{ secrets.VERCEL_TOKEN }}
    vercel-org-id: ${{ secrets.VERCEL_ORG_ID }}
    vercel-project-id: ${{ secrets.VERCEL_PROJECT_ID }}
    vercel-args: '--prod'
```

**必需的 Secrets**：
- `VERCEL_TOKEN`: 从 Vercel 账户设置获取
- `VERCEL_ORG_ID`: 来自 Vercel 项目设置
- `VERCEL_PROJECT_ID`: 来自 Vercel 项目设置

#### Netlify 部署

```yaml
- run: |
    npm install -g netlify-cli
    netlify deploy --prod --dir=.next
  env:
    NETLIFY_AUTH_TOKEN: ${{ secrets.NETLIFY_AUTH_TOKEN }}
    NETLIFY_SITE_ID: ${{ secrets.NETLIFY_SITE_ID }}
```

#### AWS S3 + CloudFront

```yaml
- uses: aws-actions/configure-aws-credentials@v4
  with:
    aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
    aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
    aws-region: us-east-1

- run: |
    aws s3 sync .next/static s3://${{ secrets.S3_BUCKET }}/static
    aws cloudfront create-invalidation --distribution-id ${{ secrets.CF_DIST_ID }} --paths "/*"
```

### 5. 测试集成

使用适当的报告配置测试执行：

**Jest 配置**：
```yaml
- name: 运行带覆盖率的测试
  run: npm test -- --coverage --coverageReporters=text --coverageReporters=lcov

- name: 上传覆盖率
  uses: codecov/codecov-action@v4
  with:
    files: ./coverage/lcov.info
    flags: unittests
```

**快速失败策略**：
```yaml
# 先运行快速测试
jobs:
  lint:  # 约 30 秒内失败
  test:  # 约 2 分钟内失败
  build: # 约 5 分钟内失败
    needs: [lint, test]
  deploy:
    needs: [build]
```

### 6. 基于分支的工作流

为每个分支实施不同的行为：

**功能分支 / PRs**：
- 仅运行 lint + test
- 不部署
- 添加带有测试结果的 PR 评论

**Develop 分支**：
- 运行 lint + test + build
- 部署到 staging 环境
- 自动部署

**Main 分支**：
- 运行 lint + test + build
- 部署到 production
- 手动批准（可选）
- 创建发布标签

**示例**：
```yaml
deploy_staging:
  if: github.ref == 'refs/heads/develop'
  # 部署到 staging

deploy_production:
  if: github.ref == 'refs/heads/main'
  environment: production  # 需要手动批准
  # 部署到 production
```

## 工作流决策树

遵循此决策树生成适当的流水线：

1. **哪个平台？**
   - GitHub → 使用 `assets/github-actions-nodejs.yml`
   - GitLab → 使用 `assets/gitlab-ci-nodejs.yml`
   - CircleCI/Jenkins → 适配 GitHub Actions 模板
   - 不确定 → 参考 `references/platform-comparison.md`

2. **需要哪些阶段？**
   - 始终包含：Lint, Test, Build
   - 可选：安全扫描、E2E 测试、性能测试
   - 如果从 CI 部署，添加部署阶段

3. **哪个部署平台？**
   - Vercel → 使用 Vercel 部署示例
   - Netlify → 使用 Netlify CLI 方法
   - AWS → 使用 AWS Actions/CLI
   - 自定义 → 实施自定义部署脚本

4. **什么触发器？**
   - 推送到 main/develop
   - 在 pull request 上
   - 在标签创建时
   - 手动工作流分派

5. **需要什么环境变量？**
   - 平台 tokens (Vercel, Netlify, AWS)
   - 外部服务的 API keys
   - 构建时环境变量
   - 功能标志

## 最佳实践

### 安全
- 将所有 secrets 存储在平台 secret 管理中（绝不在代码中）
- 使用最小权限 tokens（尽可能只读）
- 定期轮换 secrets
- 审计 secret 访问权限
- 绝不记录 secrets（使用 `***` 掩码）

### 性能
- 积极缓存依赖
- 并行化独立的 jobs
- 使用矩阵构建进行多版本测试
- 快速失败：在慢的之前运行快速检查
- 优化 Docker 层缓存

### 可靠性
- 固定确切的 Node.js 版本（`18.x` 而不是仅 `18`）
- 提交 lockfiles (`package-lock.json`)
- 为不稳定的外部服务添加重试逻辑
- 设置合理的超时（最大 10-15 分钟）
- 对非关键步骤使用 `continue-on-error`

### 可维护性
- 为复杂逻辑添加注释
- 使用可重用工作流/模板
- 保持配置 DRY (Don't Repeat Yourself)
- 版本控制所有流水线变更
- 在 README 中记录必需的 secrets

## 常见模式

### 多环境部署
```yaml
deploy_staging:
  environment: staging
  if: github.ref == 'refs/heads/develop'

deploy_production:
  environment: production
  if: github.ref == 'refs/heads/main'
  needs: [deploy_staging]
```

### 矩阵测试
```yaml
strategy:
  matrix:
    node-version: [16.x, 18.x, 20.x]
    os: [ubuntu-latest, windows-latest]
```

### 条件步骤
```yaml
- name: Deploy
  if: github.event_name == 'push' && github.ref == 'refs/heads/main'
  run: npm run deploy
```

### 产物管理
```yaml
- name: Upload build
  uses: actions/upload-artifact@v4
  with:
    name: build-output
    path: .next/
    retention-days: 7

- name: Download build
  uses: actions/download-artifact@v4
  with:
    name: build-output
```

## 故障排除

### 流水线失败
1. 检查 action/job 日志中的错误消息
2. 验证环境变量和 secrets 是否已设置
3. 在添加到流水线之前本地测试命令
4. 检查文档中的平台特定问题

### 慢构建
1. 验证缓存是否工作（检查缓存命中/未命中日志）
2. 并行化独立的 jobs
3. 如果可用，使用更快的 runners
4. 优化依赖安装

### 部署失败
1. 验证部署 tokens 是否有效
2. 检查平台状态页面
3. 审查部署日志
4. 本地测试部署命令

## 资源

### 模板 (`assets/`)
- `github-actions-nodejs.yml`: 完整的 GitHub Actions 工作流
- `gitlab-ci-nodejs.yml`: 完整的 GitLab CI 流水线

### 参考文档 (`references/`)
- `platform-comparison.md`: CI/CD 平台、部署目标、最佳实践和常见模式的详细比较

## 示例用法

**用户请求**: "创建一个运行测试并部署到 Vercel 的 GitHub Actions 工作流"

**步骤**：
1. 复制 `assets/github-actions-nodejs.yml` 模板
2. 如果不存在，创建 `.github/workflows/` 目录
3. 保存为 `.github/workflows/ci.yml`
4. 使用 Vercel 凭据更新部署部分
5. 将 secrets 添加到 GitHub 仓库设置：
   - `VERCEL_TOKEN`
   - `VERCEL_ORG_ID`
   - `VERCEL_PROJECT_ID`
6. 提交并推送以触发工作流

**用户请求**: "设置带有 staging 和 production 环境的 GitLab CI"

**步骤**：
1. 复制 `assets/gitlab-ci-nodejs.yml` 模板
2. 在仓库根目录保存为 `.gitlab-ci.yml`
3. 配置 GitLab CI/CD 变量：
   - `VERCEL_TOKEN`
   - 其他部署凭据
4. 审查 production 的手动批准设置
5. 提交以触发流水线

## 高级配置

### Monorepo 支持
```yaml
paths:
  - 'apps/frontend/**'
  - 'packages/**'
```

### 定时运行
```yaml
on:
  schedule:
    - cron: '0 2 * * *'  # 每天凌晨 2 点
```

### 外部服务集成
```yaml
- name: Notify Slack
  uses: 8398a7/action-slack@v3
  with:
    status: ${{ job.status }}
    webhook_url: ${{ secrets.SLACK_WEBHOOK }}
```

### 安全扫描
```yaml
- name: Run security audit
  run: npm audit --audit-level=moderate

- name: Check for vulnerabilities
  uses: snyk/actions/node@master
  env:
    SNYK_TOKEN: ${{ secrets.SNYK_TOKEN }}
```
