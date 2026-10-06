---
name: security-sudo-detection
description: 检测 Linux 服务器（CentOS/Ubuntu/麒麟 V10/openEuler 等）中的sudo使用情况，识别权限滥用、敏感命令执行、异常sudo失败及疑似权限提升行为，并提供风险分析与处置建议。
mode: on_demand
version: 1.0.1
maturity: stable
requires_mcp:
  - server: security-assistant
    tools:
      - collect_security_logs
      - grep_file
      - read_file
      - list_directory
      - collect_filesystem
      - collect_system_info
---



# Skill：sudo异常检测（Sudo Privilege Detection）

## 触发条件

- 症状关键词：sudo 权限滥用、敏感命令执行、异常 sudo 失败、疑似权限提升
- 组合场景：与 security-root-login-detection / security-process-detection 联动核查权限域与提权链；等保访问控制项必查

## 技能描述

负责分析 Linux 服务器（Ubuntu/CentOS/麒麟 V10/openEuler 等）的 sudo 使用情况，识别权限滥用、敏感命令执行、异常 sudo 失败及疑似权限提升行为。分析前先按 /etc/os-release 判定目标平台。

---

## 检测目标

重点检测：

- sudo 高频使用
- sudo 连续失败
- sudo 权限滥用（NOPASSWD:ALL 等）
- 敏感命令执行（useradd、passwd、visudo、systemctl、rm -rf 等）
- 异常管理员操作
- /etc/sudoers 权限异常（非 0440）

---

## MCP Tool

分析阶段调用（均为只读）：

- `collect_security_logs(log_type="sudo")` —— sudo 操作日志（已封装，检索 /var/log/secure 中 sudo 记录）
- `collect_security_logs(log_type="auth")` —— 认证日志（补充 sudo 失败记录）
- `grep_file(path="/var/log/secure", keyword="sudo")` —— 必要时直接检索 sudo 失败/错误模式
- `read_file(path="/etc/sudoers")` —— sudoers 主配置规则
- `grep_file(path="/etc/sudoers", keyword="NOPASSWD", ignore_case=True, context=2)` —— 免密授权
- `grep_file(path="/etc/sudoers", keyword="ALL", ignore_case=False)` —— 过度授权
- `list_directory(path="/etc/sudoers.d", recursive=True)` —— 列出 drop-in 规则文件
- 逐个 `read_file(path="/etc/sudoers.d/<file>")` —— 检视 drop-in 规则
- `collect_filesystem(path="/etc")` —— /etc/sudoers stat（权限应为 0440、属主 root）
- `collect_system_info` —— 系统基线

不再使用 `collect_security_configuration`；sudoers 规则通过 `read_file`/`grep_file` 读取，drop-in 通过 `list_directory` 发现，文件权限通过 `collect_filesystem` 获取。

---

## 分析流程

### Step1：采集日志与配置

- `collect_security_logs(log_type="sudo", start_time=...)` → sudo 操作记录
- `collect_security_logs(log_type="auth", keyword="sudo")` → sudo 失败/认证记录
- `read_file(path="/etc/sudoers")` → 主 sudoers 规则
- `list_directory(path="/etc/sudoers.d", recursive=True)` + 逐个 `read_file` → drop-in 规则
- `grep_file(path="/etc/sudoers", keyword="NOPASSWD", ignore_case=True)` → 免密授权
- `collect_filesystem(path="/etc")` → /etc/sudoers 权限/属主

提取：用户、执行时间、执行命令、执行结果。

---

### Step2：异常检测

统计：

- sudo 使用次数
- sudo 失败次数
- 高频 sudo 用户
- 高频失败用户

分析：

- 高频 sudo
- 连续失败
- 敏感命令执行
- 权限滥用（NOPASSWD:ALL、ALL=(ALL) ALL 滥发）
- /etc/sudoers 权限/属主异常

---

### Step3：Rule Engine 风险评分

综合：

- sudo 使用频率
- 敏感命令数量
- 执行结果
- 权限策略符合情况
- sudoers 配置宽松程度

---

### Step4：LLM 综合分析

分析：

- 是否存在权限滥用
- 是否违反最小权限原则
- 是否存在安全风险

---

### Step5：风险评估

按 Step3 评分输出风险等级（High/Medium/Low），并简述评级依据。

---

### Step6：整改建议

例如：

- 调整 sudo 权限
- 优化 sudoers 配置（移除 NOPASSWD:ALL）
- 收紧 /etc/sudoers.d drop-in
- 加强权限审计
- 遵循最小权限原则
- 修正 /etc/sudoers 权限为 0440

---

## 处置建议

可建议：

- 调整用户 sudo 权限
- 调整用户所属用户组
- 锁定异常账户
- 修改相关权限配置



---

## 处置验证

重新调用：

- `read_file(path="/etc/sudoers")`、`grep_file(path="/etc/sudoers", keyword="NOPASSWD")`
- `list_directory(path="/etc/sudoers.d", recursive=True)` + 逐个 `read_file`
- `collect_filesystem(path="/etc")`
- `collect_security_logs(log_type="sudo")`

验证：

- sudo 配置是否恢复正常；
- 权限是否符合预期；
- 是否仍存在异常 sudo 行为。

---

## 输出要求

最终输出遵循 `AGENTS.md「报告协议」` 纯 JSON schema（`requires_approval` 恒 false）；内容要素（安全事件/检测结果/异常分析/风险等级/风险依据/影响范围/整改建议/是否建议处置/处置结果/验证结果）映射至对应 JSON 字段，不作为独立输出格式。
