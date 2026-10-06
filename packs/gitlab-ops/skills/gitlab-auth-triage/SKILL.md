---
name: gitlab-auth-triage
description: GitLab 认证与权限面分诊方法论：PAT 有效性判定（401 vs 403 语义）、交互登录 422 与重定向循环的反代头形态、LDAP 三分（bind 失败/搜索过滤/组映射）、账号锁定面——固定顺序定位与判读基准（ask-ops 只读采集面）。
mode: on_demand
version: 0.1.0
maturity: experimental
requires_mcp:            # 采集依赖的工具面（装载期对账依据；旧 bianque-tools 二进制缺这些工具时应先重建）
  - server: ask-ops
    tools: [run_readonly_command, run_readonly_commands]
---

## 触发条件

- 症状关键词：gitlab 登录不上、登录 422、登录重定向循环、gitlab LDAP 排障、token 失效、401/403 异常、账号被锁
- 组合场景：全员登录失败先查 license 过期与 secrets（转 gitlab-lifecycle-ops）；反代/证书问题可能伪装成
  登录故障——422/CSRF 与重定向循环第一嫌疑是反代头

## 数据来源（ask-ops 只读面）

- PAT 有效性判定：`curl -s -w '\n%{http_code}\n' -H "PRIVATE-TOKEN: $GITLAB_TOKEN" 'http://127.0.0.1/api/v4/user'`
  ——200=token 有效、401=无效/过期/吊销、403=有效但权限不足（如需 admin 的端点）；
- API 日志面（每次请求的 status 在这）：`grep '"status":401' /var/log/gitlab/gitlab-rails/api_json.log | tail -20`、
  `grep '"status":422' /var/log/gitlab/gitlab-rails/api_json.log | tail -10`（422 集中出现=CSRF/session 面）；
- 应用日志（登录与 LDAP 错误在这）：`grep -i ldap /var/log/gitlab/gitlab-rails/application.log | tail -20`、
  `tail -100 /var/log/gitlab/gitlab-rails/production_json.log`（sessions controller 的 5xx 与重定向面）；
- LDAP 配置面（**选择性取键，密码键不取**）：`grep -E 'ldap_(enabled|host|port|method|uid|bind_dn|base|group_base|allow_username_or_email_login)' /etc/gitlab/gitlab.rb | grep -v '^#'`；
- 用户面（admin PAT）：`curl -s -H "PRIVATE-TOKEN: $GITLAB_TOKEN" 'http://127.0.0.1/api/v4/users?username=<name>'`
  （state 字段：active/blocked/banned——state 即结论素材）；全员封禁形态 `'.../api/v4/users?blocked=true'`；
- **`gitlab-rake gitlab:ldap:check` / `gitlab-rails console` 不在只读白名单**：作为审批后宿主侧动作进建议面，禁处方化。

## 分诊路径（固定顺序，先分「谁登不上」再落位）

1. **先分人群与通道**：API token（401/403 面，第 2 步）vs 交互登录（session/LDAP 面，第 3-4 步）vs
   单个用户 vs 全员——全员失败第一嫌疑是变更窗口（升级/反代/LDAP 服务端变更时间线），第二是
   license 过期与 secrets 异常（转 gitlab-lifecycle-ops），先时间线后配置。
2. **token 401 三查**：token 过期/吊销（换新 token 重试对照）、实例 secrets 变更导致存量 token 解密失败
   （联动 gitlab-lifecycle-ops 的 secrets 事故形态）、时钟漂移（过期时间判定错位）——`/api/v4/user`
   状态码是判定锚点，不猜 token 状态。
3. **交互登录 422/重定向循环**：反代未转发 Host / X-Forwarded-Proto（CSRF 校验失败的标准形态）——
   这是**反代侧**问题，GitLab 本体可能无病；核对反代配置与 422 集中出现的时间点（变更窗口对齐）；
   Cookie SameSite/secure 属性与外部 URL 协议不一致也是同族根因。
4. **LDAP 三分**（application.log 证据落位）：bind 失败（'Invalid Credentials'——bind_dn/bind 密码问题，
   服务端侧改密/禁用也是同形态）、搜索过滤不到用户（base/uid 过滤配置面——配置 grep 落证）、
   登录成功但组权限不对（group_base/group 过滤面——用户能进但权限缺是这条线）；网络不可达
   （connection refused/timeout）与凭证错误是两种病，日志文案区分。
5. **账号锁面**：连续失败锁定与 admin 手工 block——users API 的 state 字段落证；unlock 属变更动作。

## 判读基准

- 401/403 语义分开：401=未认证（token/凭证问题）、403=已认证权限不足（授权/封禁/license 问题）——混报是误诊；
- 「全员性登录故障」必须先给变更窗口时间线证据再给配置结论；无变更证据的全员故障，license/secrets 是前置排除项；
- LDAP bind 凭证错误与网络不可达是两种病（日志文案区分），处置方向完全不同（改密码 vs 查网络/防火墙）；
- 报告中**不出现任何凭证值**：bind 密码、PAT 值、口令哈希一律打码——结论只依据「有无/状态码/时间戳」；
  gitlab.rb 的密码类键不 grep 不引用。

## 输出要求

- 每个结论附状态码/日志关键行证据；变更类动作（解锁账号、轮换 token、改 LDAP 配置 + reconfigure、
  重置 root 密码的 rails console 操作）标注影响面并 requires_approval、处方为单条宿主侧命令进
  recommendation.steps；数据不足输出「需补充采集」清单（如 api_json.log 完整 401/422 段、LDAP
  配置键值、反代配置块），不臆测。
