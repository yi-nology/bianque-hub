---
name: analyzing-kubernetes-audit-logs
description: Kubernetes 安全分析专家 - 审计日志解析、威胁检测、入侵调查、RBAC 安全审计
mode: on_demand
version: 1.1.0
maturity: experimental
---

# 分析 Kubernetes 审计日志


## 何时使用

- 当调查需要分析 kubernetes 审计日志的安全事件时
- 当为此领域构建检测规则或威胁狩猎查询时
- 当 SOC 分析师需要此类分析的结构化程序时
- 当验证相关攻击技术的安全监控覆盖范围时

## 先决条件

- 熟悉容器安全概念和工具
- 访问测试或实验室环境以安全执行
- 安装了所需依赖项的 Python 3.8+
- 任何测试活动的适当授权

## 使用说明

解析 Kubernetes 审计日志文件（JSON 行格式）以检测安全相关的
事件，包括未经授权的访问、权限提升和数据泄露。

```python
import json

with open("/var/log/kubernetes/audit.log") as f:
    for line in f:
        event = json.loads(line)
        verb = event.get("verb")
        resource = event.get("objectRef", {}).get("resource")
        user = event.get("user", {}).get("username")
        if verb == "create" and resource == "pods/exec":
            print(f"Pod exec by {user}")
```

要检测的关键事件：
1. pods/exec 和 pods/attach（进入容器 shell）
2. secrets 访问（get/list/watch）
3. clusterrolebindings 创建（RBAC 提升）
4. 特权 pod 创建
5. 匿名或 system:unauthenticated 访问

## 示例

```python
# 检测 secret 枚举
if verb in ("get", "list") and resource == "secrets":
    print(f"Secret access: {user} -> {event['objectRef'].get('name')}")
```
