---
name: security-root-login-detection
description: 检测 Linux 服务器（CentOS/Ubuntu/麒麟 V10/openEuler 等）中Root账户登录行为，识别Root远程登录、异常时间登录、未知来源IP登录及违反安全策略的登录风险，并提供风险分析与处置建议。
mode: on_demand
version: 1.0.1
maturity: stable
requires_mcp:
  - server: security-assistant
    tools:
      - collect_security_logs
      - grep_file
      - read_file
      - collect_filesystem
      - collect_service_info
      - collect_system_info
      - list_directory
---



# Skill：Root登录检测（Root Login Detection）

## 触发条件

- 症状关键词：Root 登录、远程登录、异常时间登录、未知来源 IP 登录、违反安全策略登录
- 组合场景：与 security-ssh-attack / security-sudo-detection 联动核查登录链；等保身份鉴别项必查

## 技能描述

负责分析 Linux 服务器（Ubuntu/CentOS/麒麟 V10/openEuler 等）的 Root 账户登录情况，识别 Root 远程登录、异常时间登录、未知来源 IP 登录及违反安全策略的登录行为，并完成风险评估。分析前先按 /etc/os-release 判定目标平台。

---

## 检测目标

重点检测：

- Root SSH 登录
- Root 远程登录
- Root 登录成功
- 未知来源登录
- 公网来源登录
- 非工作时间登录
- Root 登录配置风险（PermitRootLogin）

---

## MCP Tool

分析阶段应调用（均为只读）：

- `collect_security_logs(log_type="ssh")` —— SSH 认证日志（journalctl _COMM=sshd，已封装）
- `collect_security_logs(log_type="auth")` —— /var/log/secure 认证日志（补充）
- `grep_file(path="/etc/ssh/sshd_config", keyword="PermitRootLogin", ignore_case=True, context=3)` —— Root 登录策略
- `read_file(path="/etc/ssh/sshd_config")` —— SSH 完整配置
- `collect_filesystem(path="/etc/ssh")` —— sshd_config stat（权限/属主）
- `collect_service_info(service_name="sshd.service")` —— sshd 运行状态/启用
- `collect_system_info` —— 系统基线

不再使用 `collect_security_configuration`；SSH 配置内容通过 `grep_file`/`read_file` 获取，配置文件权限通过 `collect_filesystem` 获取，服务状态通过 `collect_service_info` 获取。

---

## 分析流程

### Step1：采集数据

- `collect_security_logs(log_type="ssh", start_time=..., keyword="root")` → Root 登录事件
- `collect_security_logs(log_type="auth", keyword="root")` → /var/log/secure 中 Root 记录
- `grep_file(path="/etc/ssh/sshd_config", keyword="PermitRootLogin")` → 是否允许 Root 远程登录
- `read_file(path="/etc/ssh/sshd_config")` → 完整 SSH 配置（LoginGraceTime、MaxAuthTries 等）
- `collect_filesystem(path="/etc/ssh")` → sshd_config 权限/属主
- `collect_service_info(service_name="sshd.service")` → sshd 是否 active/enabled
- **安全模块上下文核查（按目标平台）**：麒麟 V10 为 KYSEC 安全架构（security-switch 执行控制）——`list_directory` 探测 /etc/kylin-security、/etc/security 等目录，`read_file`/`grep_file` 检视 security-switch 相关配置，识别 strict/custom/selinux 等安全模式；CentOS/openEuler 以 `getenforce` 取 SELinux 状态；Ubuntu 以 `aa-status` 取 AppArmor 状态。受管控（如 KYSEC strict）时认证与写操作受安全模块管控，Root 登录判定需结合该上下文；无法确认模块状态时标注"需人工核查"

提取：登录时间、登录来源 IP、登录方式（密码/密钥）、登录结果（成功/失败）。

---

### Step2：异常检测

统计：

- Root 登录次数
- 登录成功次数
- 登录失败次数
- 来源 IP
- 登录时间

识别：

- Root 远程登录（PermitRootLogin=yes/prohibit-password 且有远程登录记录）——**先结合目标平台安全模块状态**：受管控（如 KYSEC strict 模式）时 Root 登录受安全模块管控，评分可适度降级并标注管控来源（如"KYSEC 严格模式管控中"）；无法确认安全模块状态时标注"需人工核查"
- 未知来源登录
- 异常时间登录
- 公网登录

---

### Step3：Rule Engine 风险评分

综合以下因素：

- PermitRootLogin 配置
- 登录来源
- 登录方式
- 登录时间
- 登录频率
- 目标平台安全模块状态（SELinux/KYSEC/AppArmor 管控情况）

---

### Step4：LLM 综合分析

分析：

- 是否违反安全策略
- 是否存在入侵风险
- 是否建议立即整改

---

### Step5：风险评估

按 Step3 评分输出风险等级（High/Medium/Low），并简述评级依据。

---

### Step6：整改建议

例如：

- 禁止 Root SSH 登录（PermitRootLogin no）
- 修改 SSH 配置
- 启用密钥认证（禁用密码）
- 限制异常来源访问

---

## 处置建议

当风险等级为 High 时，可建议：

- 调整 Root 登录相关配置
- 限制 Root 远程登录
- 封禁异常来源 IP
- 调整网络访问控制策略



---

## 处置验证

重新调用：

- `grep_file(path="/etc/ssh/sshd_config", keyword="PermitRootLogin")` → 复核配置
- `collect_security_logs(log_type="ssh", keyword="root")` → 复核登录记录

验证：

- Root 登录配置是否符合要求；
- 异常登录是否消失；
- 网络访问控制是否生效。

---

## 输出要求

最终输出遵循平台统一报告 schema（纯 JSON，字段语义见专家 prompt 的输出协议段）；内容要素（安全事件/检测结果/异常分析/风险等级/风险依据/影响范围/整改建议/是否建议处置/处置结果/验证结果）映射至对应 JSON 字段，不作为独立输出格式。
