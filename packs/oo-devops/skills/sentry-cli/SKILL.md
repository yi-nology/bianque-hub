---
name: sentry-cli
description: 通过 sentry-cli 进行 Sentry.io 错误监控。在处理 Sentry 发布、source maps、dSYMs、事件或问题管理时使用。涵盖认证、发布工作流、部署跟踪和调试文件上传。
mode: on_demand
version: 1.1.1
maturity: experimental
---

# Sentry CLI

与 Sentry.io 交互以进行错误监控、发布管理和调试工件上传。

## 安装

```bash
# macOS
brew install sentry-cli

# npm (跨平台)
npm install -g @sentry/cli

# 直接下载
# 供应链纪律：不直灌管道安装。包管理器优先（brew install sentry-cli /
# npm install -g @sentry/cli / apt 装 sentry-cli 官方源）；确需安装脚本时两步式：
curl -sL -o /tmp/sentry-cli-install.sh https://sentry.io/get-cli/
less /tmp/sentry-cli-install.sh   # 过目后再 bash 执行
```

## 认证

```bash
# 交互式登录（打开浏览器）
sentry-cli login

# 或直接设置令牌
export SENTRY_AUTH_TOKEN="sntrys_..."

# 验证
sentry-cli info
```

将令牌存储在 `.sentryclirc` 或环境中：
```ini
[auth]
token=sntrys_...

[defaults]
org=my-org
project=my-project
```

## 发布

### 创建和完成

```bash
# 创建发布（通常是 git SHA 或版本）
sentry-cli releases new "$VERSION"

# 关联提交（将错误链接到提交）
sentry-cli releases set-commits "$VERSION" --auto

# 部署时完成
sentry-cli releases finalize "$VERSION"

# CI 的一行命令
sentry-cli releases new "$VERSION" --finalize
```

### 部署

```bash
# 将发布标记为部署到环境
sentry-cli releases deploys "$VERSION" new -e production
sentry-cli releases deploys "$VERSION" new -e staging
```

### 列出发布

```bash
sentry-cli releases list
sentry-cli releases info "$VERSION"
```

## Source Maps

上传 source maps 以进行 JavaScript 错误反混淆：

```bash
# 上传所有 .js 和 .map 文件
sentry-cli sourcemaps upload ./dist --release="$VERSION"

# 使用 URL 前缀（匹配你的部署路径）
sentry-cli sourcemaps upload ./dist \
  --release="$VERSION" \
  --url-prefix="~/static/js"

# 上传前验证
sentry-cli sourcemaps explain ./dist/main.js.map
```

### 注入 Debug IDs（推荐）

```bash
# 将 debug IDs 注入源文件（现代方法）
sentry-cli sourcemaps inject ./dist
sentry-cli sourcemaps upload ./dist --release="$VERSION"
```

## 调试文件 (iOS/Android)

### dSYMs (iOS)

```bash
# 从 Xcode 归档上传 dSYMs
sentry-cli debug-files upload --include-sources path/to/dSYMs

# 从派生数据
sentry-cli debug-files upload ~/Library/Developer/Xcode/DerivedData/*/Build/Products/*/*.app.dSYM
```

### ProGuard (Android)

```bash
sentry-cli upload-proguard mapping.txt --uuid="$UUID"
```

### 检查调试文件

```bash
sentry-cli debug-files check path/to/file
sentry-cli debug-files list
```

## 事件和问题

### 发送测试事件

```bash
sentry-cli send-event -m "Test error message"
sentry-cli send-event -m "Error" --logfile /var/log/app.log
```

### 查询问题

```bash
# 列出未解决的问题
sentry-cli issues list

# 解决问题
sentry-cli issues resolve ISSUE_ID

# 静音/忽略
sentry-cli issues mute ISSUE_ID
```

## 监控器 (Cron)

```bash
# 包装 cron 作业
sentry-cli monitors run my-cron-monitor -- /path/to/script.sh

# 手动签到
sentry-cli monitors check-in my-monitor --status ok
sentry-cli monitors check-in my-monitor --status error
```

## CI/CD 集成

### GitHub Actions

```yaml
- name: Create Sentry Release
  uses: getsentry/action-release@v1
  env:
    SENTRY_AUTH_TOKEN: ${{ secrets.SENTRY_AUTH_TOKEN }}
    SENTRY_ORG: my-org
    SENTRY_PROJECT: my-project
  with:
    environment: production
    sourcemaps: ./dist
```

### 通用 CI

```bash
export SENTRY_AUTH_TOKEN="$SENTRY_TOKEN"
export SENTRY_ORG="my-org"
export SENTRY_PROJECT="my-project"
VERSION=$(sentry-cli releases propose-version)

sentry-cli releases new "$VERSION" --finalize
sentry-cli releases set-commits "$VERSION" --auto
sentry-cli sourcemaps upload ./dist --release="$VERSION"
sentry-cli releases deploys "$VERSION" new -e production
```

## 常见标志

| Flag | Description |
|------|-------------|
| `-o, --org` | Organization slug |
| `-p, --project` | Project slug |
| `--auth-token` | 覆盖认证令牌 |
| `--log-level` | debug/info/warn/error |
| `--quiet` | 抑制输出 |

## 故障排除

```bash
# 检查配置
sentry-cli info

# 调试上传问题
sentry-cli --log-level=debug sourcemaps upload ./dist

# 验证 source map
sentry-cli sourcemaps explain ./dist/main.js.map

# 检查连接性
sentry-cli send-event -m "test" --log-level=debug
```
