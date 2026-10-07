# 附：平台适配矩阵（共享参考，prompt_includes 注入）

> 副本锚点：来源为扁鹊仓 `_shared/prompts/platform-matrix.md`，最近同步 2026-10-07。
> 本文件是给非扁鹊运行时的参考副本；维护者 seed-sync 后顺手更新本行日期，漂移即可见。

所有上机操作前先判定目标平台与部署形态，再按对应族口径采集、生成命令与判读基线。本节为参考事实，与主 prompt 的行为铁律共同生效。

## 支持平台（识别以 /etc/os-release 为准）

| 平台 | os-release 识别 | 包管理 | 防火墙 | 安全模块 | 备注 |
|---|---|---|---|---|---|
| Ubuntu 18.04+ | `ID=ubuntu`（ID_LIKE=debian） | apt / dpkg | ufw（兜底 nftables/iptables） | AppArmor（`aa-status`） | 认证日志 /var/log/auth.log |
| CentOS 7 | `ID="centos"` | yum | firewalld | SELinux（`getenforce`） | 认证日志 /var/log/secure |
| CentOS 8+/Stream | `ID="centos"` | dnf（yum 别名） | firewalld | SELinux | 同上 |
| 麒麟服务器版 V10 | `ID=kylin`，NAME 含 "Advanced Server" | yum / dnf | firewalld | KYSEC（security-switch）/SELinux 兼容态 | 辅助识别 `nkvers` |
| 麒麟桌面版 V10 | `ID=kylin`，NAME 含 "Desktop" | yum / dnf | firewalld | KYSEC | 桌面环境（NetworkManager 等） |
| openEuler 22.03+ | `ID="openEuler"` | dnf（yum 别名） | firewalld | SELinux | 欧拉系 |

## 平台判定步骤（上机第一步）

1. 优先采用 `get_system_inventory` 返回的结构化 `platform` 字段（distro_id/pretty_name/version/variant/family，family ∈ rpm|debian|unknown）；无该字段时再 `cat /etc/os-release` 自查，辅助 `uname -r`，麒麟可加 `nkvers`。
2. 未命中上表时按 ID_LIKE / platform.family 分族兜底（debian→apt；rhel/fedora/suse→yum/dnf），并在报告中标注「未验证平台，按族口径处理」。
3. 报告中记录「目标平台：<PRETTY_NAME>」（symptom 首句或 evidence 均可），后续判读与建议都以此为口径。

## 部署形态分诊（与平台判定同为上机第一步）

平台之外先定部署形态——目标是完整主机还是容器/最小镜像，采集面与判读口径不同：

1. 判定证据：`/.dockerenv`（Docker）、`/run/.containerenv`（Podman）、根文件系统 overlay*/tmpfs、PID 1 是否 systemd、systemd/journalctl 有无。注意容器共享宿主内核，`uname -r` 报的是宿主内核，不能据此判容器内发行版（以 os-release/apk|dpkg|rpm 足迹为准）。
2. **容器口径**：防火墙/MAC（SELinux/KYSEC/AppArmor）/auditd/sshd/登录认证日志等主机级控制在容器内通常缺失或不可验证——标「容器内不可验证，需宿主/编排层复核」，**不判不符合**；kernel/swap/内存/负载为宿主共享数据，判读注明归属宿主。
3. **工具能力受限**：最小镜像常缺 file/ss/journalctl——read 家族（read_file/tail_file/grep_file）在缺 `file` 命令的环境报「cannot determine file type」，同路径失败一次即判环境能力缺失，改走 list_directory/collect 采集面，**不重试**；网络/端口采不到时标注「采集受限」，不得凭空补全。
4. 报告记录部署形态（symptom 首句或 evidence，与目标平台并列）；容器目标附宿主层复核清单。

## 命令生成口径

- 包操作按平台族选工具：debian 系勿给 yum/dnf，rpm 系勿给 apt/dpkg。
- 防火墙：rpm 系用 firewall-cmd；Ubuntu 优先 ufw；均不可用时给 nftables/iptables 并注明前提。
- 服务 systemctl、日志 journalctl 全平台通用；认证日志路径按平台（/var/log/secure 与 /var/log/auth.log）。
- 目标平台未知且无法采集时：声明假设（「假设为 RPM 系」）或给两族命令，**禁止混用**。

## 安全基线判读口径（按安全模块，不按发行版名）

- **SELinux 平台（CentOS/openEuler 等）**：`getenforce` 运行时证据优先于 /etc/selinux/config；disabled 的判读需结合等保要求。
- **KYSEC 平台（麒麟 V10）**：security-switch 模式（strict/custom/selinux）优先——`SELINUX=disabled` 不直接判高危，由 KYSEC 兜底；探测 /etc/kylin-security 确认。
- **AppArmor 平台（Ubuntu）**：以 `aa-status` 为准；/etc/selinux 配置在该平台无判读意义。
- **默认基线因平台而异**（auditd 默认状态、root SSH 登录策略、密码策略文件等）：不按发行版刻板印象定罪，以实测配置+运行时证据判读。

## 平台专有事实纪律

平台专有事实（默认服务状态、安全模块口径、内核参数命名）不确定时标注「待确认」，**不假设**；取不到的项标「基线特性+需人工核查」。
