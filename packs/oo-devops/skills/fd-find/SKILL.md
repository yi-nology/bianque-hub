---
name: fd-find
description: '''find'' 的快速且用户友好的替代方案 - 语法简单、默认智能、尊重 gitignore。'
mode: on_demand
version: 1.1.0
maturity: experimental
---

# fd - 快速文件查找器

`find` 的用户友好替代方案，具有智能默认值。

## 快速开始

### 基础搜索
```bash
# 按名称查找文件
fd pattern

# 在指定目录中查找
fd pattern /path/to/dir

# 不区分大小写
fd -i pattern
```

### 常见模式
```bash
# 查找所有 Python 文件
fd -e py

# 查找多个扩展名
fd -e py -e js -e ts

# 仅查找目录
fd -t d pattern

# 仅查找文件
fd -t f pattern

# 查找符号链接
fd -t l
```

## 高级用法

### 过滤
```bash
# 排除模式
fd pattern -E "node_modules" -E "*.min.js"

# 包含隐藏文件
fd -H pattern

# 包含被忽略的文件 (.gitignore)
fd -I pattern

# 搜索所有（隐藏 + 被忽略）
fd -H -I pattern

# 最大深度
fd pattern -d 3
```

### 执行
```bash
# 对结果执行命令
fd -e jpg -x convert {} {.}.png

# 并行执行
fd -e md -x wc -l

# 与 xargs 配合使用
fd -e log -0 | xargs -0 rm
```

### 正则模式
```bash
# 完整正则搜索
fd '^test.*\.js$'

# 匹配完整路径
fd --full-path 'src/.*/test'

# Glob 模式
fd -g "*.{js,ts}"
```

## 基于时间的过滤
```bash
# 最近一天内修改
fd --changed-within 1d

# 在指定日期之前修改
fd --changed-before 2024-01-01

# 最近创建
fd --changed-within 1h
```

## 基于大小的过滤
```bash
# 大于 10MB 的文件
fd --size +10m

# 小于 1KB 的文件
fd --size -1k

# 特定大小范围
fd --size +100k --size -10m
```

## 输出格式化
```bash
# 绝对路径
fd --absolute-path

# 列表格式（类似 ls -l）
fd --list-details

# 空分隔符（用于 xargs）
fd -0 pattern

# 颜色始终/从不/自动
fd --color always pattern
```

## 常见用例

**查找并删除旧文件：**
```bash
fd --changed-before 30d -t f -x rm {}
```

**查找大文件：**
```bash
fd --size +100m --list-details
```

**将所有 PDF 复制到目录：**
```bash
fd -e pdf -x cp {} /target/dir/
```

**统计所有 Python 文件的行数：**
```bash
fd -e py -x wc -l | awk '{sum+=$1} END {print sum}'
```

**查找损坏的符号链接：**
```bash
fd -t l -x test -e {} \; -print
```

**在特定时间窗口内搜索：**
```bash
fd --changed-within 2d --changed-before 1d
```

## 与其他工具集成

**与 ripgrep 配合使用：**
```bash
fd -e js | xargs rg "pattern"
```

**与 fzf（模糊查找器）配合使用：**
```bash
vim $(fd -t f | fzf)
```

**与 bat（cat 替代方案）配合使用：**
```bash
fd -e md | xargs bat
```

## 性能提示

- `fd` 通常比 `find` 快得多
- 默认尊重 `.gitignore`（使用 `-I` 禁用）
- 自动使用并行遍历
- 智能大小写：小写 = 不区分大小写，任何大写 = 区分大小写

## 提示

- 使用 `-t` 进行类型过滤（f=文件，d=目录，l=符号链接，x=可执行文件）
- `-e` 用于扩展名比 `-g "*.ext"` 更简单
- `-x` 命令中的 `{}` 表示找到的路径
- `{.}` 去掉扩展名
- `{/}` 获取基本名称，`{//}` 获取目录

## 文档

GitHub: https://github.com/sharkdp/fd
Man page: `man fd`
