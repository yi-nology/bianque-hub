---
name: security-ssh-attack
description: 检测 Linux 服务器（CentOS/Ubuntu/麒麟 V10/openEuler 等）中的SSH暴力破解、密码猜测、凭据填充等攻击行为，识别异常登录模式及持续攻击，并提供风险分析与处置建议。
mode: on_demand
version: 1.0.1
maturity: stable
requires_mcp:
  - server: security-assistant
    tools:
      - collect_security_logs
      - grep_file
      - collect_service_info
      - collect_system_info
      - collect_network_status
      - list_directory
      - read_file
---



# Skill：SSH攻击检测（SSH Attack Detection）

## 触发条件

- 症状关键词：SSH 暴力破解、密码猜测、凭据填充、异常登录失败、持续 SSH 攻击
- 组合场景：与 security-root-login-detection / security-sudo-detection 联动核查登录链；等保入侵防范项必查

## 技能描述

负责检测 Linux 服务器（Ubuntu/CentOS/麒麟 V10/openEuler 等）是否存在 SSH 暴力破解、密码猜测、凭据填充等攻击行为。分析前先按 /etc/os-release 判定目标平台。

通过分析 SSH 认证日志，识别异常登录模式、攻击来源及持续攻击行为，结合规则引擎与大模型完成风险分析；处置建议只随诊断报告提交平台审批（requires_approval 语义），技能自身不执行任何变更动作。

---

## 检测目标

重点关注以下安全事件：

- SSH 暴力破解
- 密码猜测攻击
- 凭据填充攻击
- 高频登录失败
- 单个 IP 攻击多个账户
- 多个 IP 攻击同一账户
- 登录成功前存在大量失败尝试
- 持续性 SSH 攻击

---

## MCP Tool

分析阶段应调用（均为只读）：

- `collect_security_logs(log_type="ssh")` —— SSH 认证日志（journalctl _COMM=sshd，已封装，支持时间范围+关键字）
- `collect_security_logs(log_type="auth")` —— /var/log/secure 认证日志（补充/回退）
- `grep_file(path="/var/log/secure", keyword="Failed password")` —— 必要时直接在 secure 日志检索特定模式
- `grep_file(path="/etc/ssh/sshd_config", keyword="MaxAuthTries", ignore_case=True)` / `keyword="PermitRootLogin"` / `keyword="PasswordAuthentication"` —— SSH 加固策略
- `collect_service_info(service_name="sshd.service")` —— sshd 运行状态/启用
- `collect_system_info` —— 系统基线

不再使用 `collect_security_configuration`；SSH 配置通过 `grep_file` 检索，服务状态通过 `collect_service_info` 获取。

若数据不足，应继续采集，不得直接生成分析结论。

---

## 分析流程

### Step1：采集日志

调用 `collect_security_logs(log_type="ssh", start_time=..., end_time=..., keyword="Failed")` 获取指定时间范围内 SSH 认证日志。

重点获取：登录时间、来源 IP、用户名、登录结果、登录方式。

如需定位特定模式，`grep_file(path="/var/log/secure", keyword="Invalid user")` / `keyword="Failed password for root")`。

如需要，`grep_file(path="/etc/ssh/sshd_config", keyword="MaxAuthTries")` 获取 SSH 加固配置。

**安全模块上下文核查（按目标平台）**：麒麟 V10 为 KYSEC 安全架构（security-switch 执行控制）——`list_directory` 探测 /etc/kylin-security、/etc/security 等目录，`read_file`/`grep_file` 检视 security-switch 相关配置，识别 strict/custom/selinux 等安全模式；CentOS/openEuler 以 `getenforce` 取 SELinux 状态；Ubuntu 以 `aa-status` 取 AppArmor 状态。受管控（如 KYSEC strict）时认证与写操作受安全模块管控，攻击判定需结合该上下文；无法确认模块状态时标注"需人工核查"。

---

### Step2：异常检测

统计：

- 登录失败次数
- 每个来源 IP 的失败次数
- 每个用户失败次数
- 登录失败频率
- 攻击持续时间

识别：

- 高频失败登录
- 多账户攻击
- 多来源攻击
- 攻击成功前的大量失败记录——**先结合目标平台安全模块状态**：受管控（如 KYSEC strict 模式）时认证行为可能由安全模块策略产生，评分可适度降级并标注管控来源（如"KYSEC 严格模式管控中"）；无法确认安全模块状态时标注"需人工核查"

---

### Step3：Rule Engine 风险评分

结合以下因素进行评分：

- 登录失败次数
- 攻击持续时间
- 登录失败频率
- 是否针对 Root
- 是否涉及多个账户
- 是否存在成功登录
- 来源 IP 风险
- 目标平台安全模块状态（SELinux/KYSEC/AppArmor 管控情况）

---

### Step4：LLM 综合分析

结合日志及 Rule Engine 评分，对攻击行为进行分析，包括：

- 攻击模式
- 攻击来源
- 攻击影响
- 是否属于持续攻击
- 是否建议立即处置

---

### Step5：风险评估

按 Step3 评分输出风险等级（High/Medium/Low），并简述评级依据。

---

### Step6：整改建议

根据分析结果提出建议，例如：

- 启用 Fail2Ban
- 修改 SSH 登录策略（MaxAuthTries、LoginGraceTime）
- 禁止 Root SSH 登录（PermitRootLogin no）
- 使用密钥认证（禁用密码）
- 加强密码策略
- 限制登录失败次数

---

## 处置建议

当风险等级为 High 时，可建议：

- 将恶意来源 IP 加入黑名单
- 阻断异常来源的网络访问
- 根据安全策略调整访问控制规则



---

## 处置验证

完成处置后，应重新调用：

- `collect_network_status`
- `collect_security_logs(log_type="ssh", keyword="Failed")`

验证：

- 恶意来源是否已被阻断；
- SSH 服务是否正常；
- 是否仍存在攻击行为。

---

## 输出要求

最终输出遵循平台统一报告 schema（纯 JSON，字段语义见专家 prompt 的输出协议段）；下述内容要素映射至对应 JSON 字段，不作为独立输出格式。仅检测与建议、不含处置执行时 `requires_approval` 为 false；报告含处置建议（封禁 IP/停账号等）时随建议置 true 交平台审批。

最终输出应包括：

- 安全事件
- 检测结果
- 异常分析
- 风险等级
- 风险依据
- 影响范围
- 整改建议
- 是否建议处置
- 处置结果（如执行）
- 验证结果（如执行）
