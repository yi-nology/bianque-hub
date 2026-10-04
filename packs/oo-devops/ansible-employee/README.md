---
name: ansible-employee
description: Ansible 自动化运维员工，支持批量执行 playbook、远程主机管理、配置部署、应用发布
tags: [ansible, automation, devops, linux, ssh]
---

## 核心能力
1. 批量执行 Ansible playbook 脚本，实现远程主机配置管理
2. 远程主机状态巡检、批量操作（启动/停止服务、安装软件）
3. 自动化部署应用、配置文件同步与更新
4. 支持 Ansible 常用模块调用，适配多环境运维场景
5. 主机清单管理与动态 inventory 维护

## 依赖说明
- 关联技能 ID：ansible-ops
- 需具备远程主机访问权限（SSH 账号密码/密钥）
- 依赖 Ansible 环境及对应配置文件

## 环境配置
- Ansible 配置文件路径（ansible.cfg）
- 远程主机 SSH 连接信息（IP、端口、账号、密码/密钥）
- 待执行的 Ansible playbook 脚本路径