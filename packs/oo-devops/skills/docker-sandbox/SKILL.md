---
name: docker-sandbox
description: 创建和管理 Docker 沙盒 VM 环境以安全地执行 agent。在运行不受信任的代码、探索包或隔离 agent 工作负载时使用。支持 Claude、Codex、Copilot、Gemini 和 Kiro agents，并带有网络代理控制。
mode: on_demand
version: 1.1.0
maturity: experimental
---

# Docker Sandbox

使用 Docker Desktop 的 sandbox 功能在 **隔离的 VM 环境** 中运行 agents 和命令。每个 sandbox 都有自己的轻量级 VM，具有文件系统隔离、网络代理控制以及通过 virtiofs 挂载的工作空间。

## 何时使用

- 在系统范围安装之前探索 **不受信任的包** 或 skills
- 安全地运行来自外部来源的 **任意代码**
- 在不危及主机的情况下测试 **破坏性操作**
- 隔离需要网络访问控制的 **agent 工作负载**
- 为实验设置 **可复现的环境**

## 要求

- Docker Desktop 4.49+ 带有 `docker sandbox` 插件
- 验证: `docker sandbox version`

## 快速开始

### 为当前项目创建一个 sandbox

```bash
docker sandbox create --name my-sandbox claude .
```

这会创建一个 VM 隔离的 sandbox，包含:
- 通过 virtiofs 挂载的当前目录
- 预装的 Node.js、git 和标准开发工具
- 带有 allowlist 控制的网络代理

### 在内部运行命令

```bash
docker sandbox exec my-sandbox node --version
docker sandbox exec my-sandbox npm install -g some-package
docker sandbox exec -w /path/to/workspace my-sandbox bash -c "ls -la"
```

### 直接运行一个 agent

```bash
# 一步创建并运行
docker sandbox run claude . -- -p "What files are in this project?"

# 在 -- 之后传递 agent 参数
docker sandbox run my-sandbox -- -p "Analyze this codebase"
```

## 命令参考

### 生命周期

```bash
# 创建一个 sandbox (agents: claude, codex, copilot, gemini, kiro, cagent)
docker sandbox create --name <name> <agent> <workspace-path>

# 在 sandbox 中运行 agent (需要时创建)
docker sandbox run <agent> <workspace> [-- <agent-args>...]
docker sandbox run <existing-sandbox> [-- <agent-args>...]

# 执行一个命令
docker sandbox exec [options] <sandbox> <command> [args...]
  -e KEY=VAL          # 设置环境变量
  -w /path            # 设置工作目录
  -d                  # 分离 (后台)
  -i                  # 交互式 (保持 stdin 打开)
  -t                  # 分配伪 TTY

# 停止但不移除
docker sandbox stop <sandbox>

# 移除 (销毁 VM)
docker sandbox rm <sandbox>

# 列出所有 sandboxes
docker sandbox ls

# 重置所有 sandboxes
docker sandbox reset

# 将快照保存为可复用的模板
docker sandbox save <sandbox>
```

### 网络控制

Sandbox 包含一个用于控制出站访问的网络代理。

```bash
# 允许特定域名
docker sandbox network proxy <sandbox> --allow-host example.com
docker sandbox network proxy <sandbox> --allow-host api.github.com

# 拦截特定域名
docker sandbox network proxy <sandbox> --block-host malicious.com

# 拦截 IP 范围
docker sandbox network proxy <sandbox> --block-cidr 10.0.0.0/8

# 为特定主机绕过代理 (直接连接)
docker sandbox network proxy <sandbox> --bypass-host localhost

# 设置默认策略 (默认允许全部或拒绝全部)
docker sandbox network proxy <sandbox> --policy deny  # 拦截所有，然后 allowlist
docker sandbox network proxy <sandbox> --policy allow  # 允许所有，然后 blocklist

# 查看网络活动
docker sandbox network log <sandbox>
```

### 自定义模板

```bash
# 使用自定义容器镜像作为基础
docker sandbox create --template my-custom-image:latest claude .

# 将当前 sandbox 状态保存为模板以供复用
docker sandbox save my-sandbox
```

## 工作空间挂载

主机上的工作空间路径通过 virtiofs 挂载到 sandbox 中。Sandbox 内部的挂载路径保留主机路径结构:

| Host OS | Host Path | Sandbox Path |
|---|---|---|
| Windows | `H:\Projects\my-app` | `/h/Projects/my-app` |
| macOS | `/Users/me/projects/my-app` | `/Users/me/projects/my-app` |
| Linux | `/home/me/projects/my-app` | `/home/me/projects/my-app` |

Agent 的主目录是 `/home/agent/`，带有一个符号链接的 `workspace/` 目录。

## Sandbox 内部的环境

每个 sandbox VM 包含:
- **Node.js** (v20.x LTS)
- **Git** (最新版)
- **Python** (系统自带)
- **curl**, **wget**, 标准 Linux 工具
- **npm** (全局安装目录在 `/usr/local/share/npm-global/`)
- **Docker socket** (在 `/run/docker.sock` - 支持 Docker-in-Docker)

### 代理配置 (自动设置)

```
HTTP_PROXY=http://host.docker.internal:3128
HTTPS_PROXY=http://host.docker.internal:3128
NODE_EXTRA_CA_CERTS=/usr/local/share/ca-certificates/proxy-ca.crt
SSL_CERT_FILE=/usr/local/share/ca-certificates/proxy-ca.crt
```

**重要**: Node.js `fetch` (undici) 默认不尊重 `HTTP_PROXY` 环境变量。对于使用 `fetch` 的 npm 包，创建一个 require hook:

```javascript
// /tmp/proxy-fix.js
const proxy = process.env.HTTPS_PROXY || process.env.HTTP_PROXY;
if (proxy) {
  const { ProxyAgent } = require('undici');
  const agent = new ProxyAgent(proxy);
  const origFetch = globalThis.fetch;
  globalThis.fetch = function(url, opts = {}) {
    return origFetch(url, { ...opts, dispatcher: agent });
  };
}
```

运行: `node -r /tmp/proxy-fix.js your-script.js`

## 模式

### 安全的包探索

```bash
# 创建隔离的 sandbox
docker sandbox create --name pkg-test claude .

# 将网络限制为仅 npm registry
docker sandbox network proxy pkg-test --policy deny
docker sandbox network proxy pkg-test --allow-host registry.npmjs.org
docker sandbox network proxy pkg-test --allow-host api.npmjs.org

# 安装并检查包
docker sandbox exec pkg-test npm install -g suspicious-package
docker sandbox exec pkg-test bash -c "find /usr/local/share/npm-global/lib/node_modules/suspicious-package -name '*.js' | head -20"

# 检查 post-install 脚本、网络调用、文件访问
docker sandbox network log pkg-test

# 清理
docker sandbox rm pkg-test
```

### 持久的开发环境

```bash
# 一次性创建
docker sandbox create --name dev claude ~/projects/my-app

# 跨会话使用
docker sandbox exec dev npm test
docker sandbox exec dev npm run build

# 保存为模板以供团队共享
docker sandbox save dev
```

### 锁定的 Agent 执行

```bash
# 拒绝所有网络，只允许必要的
docker sandbox create --name secure claude .
docker sandbox network proxy secure --policy deny
docker sandbox network proxy secure --allow-host api.openai.com
docker sandbox network proxy secure --allow-host github.com

# 在限制下运行 agent
docker sandbox run secure -- -p "Review this code for security issues"
```

## 故障排查

### "client version X is too old"
将 Docker Desktop 更新到 4.49+。Sandbox 插件需要 engine API v1.44+。

### "fetch failed" inside sandbox
Node.js `fetch` 不使用代理。使用上面的 proxy-fix.js require hook，或者使用 `curl`:
```bash
docker sandbox exec my-sandbox curl -sL https://api.example.com/data
```

### Windows 上的路径转换 (Git Bash / MSYS2)
Git Bash 会将 `/path` 转换为 `C:/Program Files/Git/path`。在命令前加上:
```bash
MSYS_NO_PATHCONV=1 docker sandbox exec my-sandbox ls /home/agent
```

### Docker 更新后 Sandbox 无法启动
```bash
docker sandbox reset  # 清除所有 sandbox 状态
```
