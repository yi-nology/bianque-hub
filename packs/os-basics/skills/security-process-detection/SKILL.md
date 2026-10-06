---
name: security-process-detection
description: 检测 Linux 服务器（CentOS/Ubuntu/麒麟 V10/openEuler 等）中的疑似恶意进程、异常运行行为、疑似挖矿程序及后门行为，并提供风险分析与处置建议。
mode: on_demand
version: 1.0.1
maturity: stable
requires_mcp:
  - server: security-assistant
    tools:
      - collect_process_info
      - collect_network_status
      - collect_filesystem
      - list_directory
      - read_file
      - collect_security_logs
      - tail_file
---



# Skill10：恶意进程检测（Suspicious Process Detection）

## 触发条件

- 症状关键词：恶意进程、挖矿程序、异常进程、僵尸进程、可疑可执行路径、父子进程攻击链
- 组合场景：与 security-ssh-attack 联动核查攻击来源；入侵响应时配合 security-root-login-detection / security-sudo-detection 查权限滥用链

## Skill职责

负责分析 Linux 服务器（Ubuntu/CentOS/麒麟 V10/openEuler 等）当前运行进程，结合进程属性、资源占用、启动路径、父子进程关系、网络连接等信息，识别疑似恶意进程、疑似挖矿程序、疑似后门行为及其他异常运行行为。分析前先按 /etc/os-release 判定目标平台。

本 Skill 基于规则引擎（Rule Engine）与大模型（LLM）联合分析，不依赖病毒特征库，不得直接认定某进程为木马程序，而应依据检测证据进行综合判断。

---

## MCP Tool

优先调用以下工具（均为只读）：

- `collect_process_info` —— 进程列表（pid/ppid/user/pcpu/pmem/stat/comm/args/executable）、僵尸进程、负载；可按 process_name 过滤、top_n 限制
- `collect_network_status` —— 关联网络连接（ss 输出含 process 字段，可按 PID 对位）
- `collect_filesystem(path="<可执行文件目录>")` —— 对可疑可执行文件做 stat（权限/属主/时间）与目录列表
- `list_directory(path="<可执行文件目录>", recursive=True, depth=2)` —— 列出可疑目录（/tmp、/dev/shm、/var/tmp、用户目录隐藏文件）
- `read_file` —— 必要时读取可疑脚本/配置文件内容
- `collect_security_logs(log_type="all", keyword="<comm>")` —— 在日志中检索可疑进程痕迹
- `tail_file(path="/var/log/messages", lines=200)` —— 必要时查看系统日志末尾，定位可疑进程相关的近期事件

如需处置：处置建议。

---

## 执行流程

### Step1：采集进程信息

调用 `collect_process_info` 获取：

- 当前运行进程
- PID / PPID / User / CPU / Memory / State(stat) / comm / args(完整命令行) / executable
- 僵尸进程
- 系统负载

必要时调用 `collect_network_status` 获取进程对应的网络连接（按 ss 的 process 字段对位 PID），识别外联目标与高危端口。

对可疑可执行路径（如 /tmp/.xxx、/dev/shm/xxx），调用：

- `list_directory(path="<目录>", recursive=True, depth=2)` 列出目录内容
- `collect_filesystem(path="<目录>")` 获取文件 stat（权限/属主/修改时间）
- `read_file(path="<脚本路径>")` 读取脚本内容（若是脚本）

必要时 `collect_security_logs(log_type="all", keyword="<comm或路径>")` 检索日志痕迹。

---

### Step2：识别异常进程

#### （1）资源异常

- CPU 长期高占用（疑似挖矿）
- Memory 异常占用

#### （2）启动路径异常

- /tmp、/var/tmp、/dev/shm、用户目录隐藏文件

#### （3）异常进程名称

伪装为系统进程名（systemd、sshd、kworker、dbus、[kworker]）但 executable 路径明显异常。

#### （4）Deleted Executable

executable 显示 (deleted) 属于高风险行为。

**说明**：`/proc/PID/exe` 符号链接的 (deleted) 状态不在当前只读 Tool 直接采集范围，应基于 `collect_process_info` 的 executable 字段推断；如需精确确认，标注"需人工执行 `ls -l /proc/PID/exe` 核查"。

#### （5）父子进程异常

典型攻击链：bash → curl → bash → python → nc。

#### （6）异常网络通信

未知外联、高危端口、持续连接、矿池连接（常见 3333/4444/5555/14444/8080 端口外联）。

---

### Step3：综合行为分析

结合以下维度综合分析：

- 资源占用
- 启动路径
- 父子关系
- 网络连接
- 文件状态
- 命令行参数

判断：是否疑似挖矿程序 / 是否疑似后门行为 / 是否存在可疑行为。

不得仅依据单一指标直接判定恶意程序。

---

### Step4：风险评估

低风险：

- 单一异常指标
- 可解释行为

中风险：

- 多项异常同时出现
- 疑似恶意行为

高风险：

满足多个高危规则，例如：

- Deleted Executable
- 高 CPU
- 外联异常 IP
- 异常启动目录
- 攻击链明显

---

### Step5：输出报告

最终输出遵循平台统一报告 schema（纯 JSON，字段语义见专家 prompt 的输出协议段）；下述内容要素映射至对应 JSON 字段，不作为独立输出格式。

输出：

- 异常进程列表
- 关键行为
- 风险原因
- 综合判断
- 整改建议

---

### Step6：处置建议

若风险等级为 High，可建议：终止或隔离疑似恶意进程。



---

## 分析约束

不得直接输出：已确认木马 / 已确认病毒 / 已确认后门，除非具有充分证据。

优先使用：疑似恶意进程 / 疑似挖矿程序 / 疑似后门行为 / 存在可疑行为，建议进一步排查。

最终分析必须说明：判断依据、命中的规则、使用的数据来源。
