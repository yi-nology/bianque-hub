---
name: awscli
description: 使用 AWS CLI 管理 AWS Lightsail 和 EC2 实例。
mode: on_demand
version: 1.1.0
maturity: experimental
---

# AWS CLI Control Skill

此技能管理 AWS Lightsail 实例。

## 要求

- 主机上已安装 AWS CLI
- 已配置 AWS 凭证（IAM 用户或角色）
- 环境变量：

  - AWS_REGION
  - ALLOWED_INSTANCES

  ## 环境变量

此技能需要以下环境变量：

- AWS_REGION（例如，ap-southeast-1）
- ALLOWED_INSTANCES（逗号分隔列表）

示例：

AWS_REGION=ap-southeast-1
ALLOWED_INSTANCES=Ubuntu,Binami

## 可用操作

### 1. 列出实例

action: "list"

示例：
{
"action": "list"
}

---

### 2. 重启实例

action: "reboot"  
instance: "<instance-name>"

示例：
{
"action": "reboot",
"instance": "Ubuntu-1"
}

---

### 3. 启动实例

action: "start"  
instance: "<instance-name>"

---

### 4. 停止实例

action: "stop"  
instance: "<instance-name>"

---

## 注意

- 仅使用结构化 JSON 输入。
- 不要生成 AWS CLI 命令。
- 实例名称必须与现有 Lightsail 实例完全匹配。
