---
name: gitbackup
description: 创建 OpenClaw workspace repository 的本地 Git bundle 备份。在 Telegram 中运行 /gitbackup 或当用户要求将 Git history/refs 备份到本地文件时使用。
mode: on_demand
version: 1.1.0
maturity: experimental
---

# Git 备份（本地 bundle）

## 概述
创建 workspace repo 的自包含 Git bundle 并将其存储在 `/root/.openclaw/backups` 中。

## 快速开始
运行打包脚本：
```bash
bash /root/.openclaw/workspace/skills/gitbackup/scripts/git-backup.sh
```

## 输出
打印 bundle 路径和大小。bundle 文件名包含 UTC 时间戳。

## 注意事项
- 如果 workspace 不是 Git repo，以明确的错误退出。
- 不要删除旧的 bundles。
