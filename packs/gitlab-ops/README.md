# gitlab-ops：GitLab 自托管平台运维包（社区版）

当 GitLab 自己病了——全站 502、仓库推不上去、全员登录不上、备份链断了、
升级半路卡死——本包接管诊断。四技能 + 一专家，全部只读采集。

| 资产 | 说明 |
|---|---|
| `skills/gitlab-instance-triage` | 实例本体分诊：健康端点三级语义（/-/health / /-/liveness / /-/readiness 分项即地图）、502 归因（反代活/上游死）、首错定位、慢面（sidekiq 积压/盘水位） |
| `skills/gitlab-repo-triage` | 仓库推拉面分诊：Git 智能协议探测状态码语义（401/403/404/500）、push 403 三分（分支保护/push rules/权限）、gitaly 与 hooks 面、git-data/LFS 盘水位、SSH 通道 |
| `skills/gitlab-auth-triage` | 认证与权限面分诊：PAT 有效性判定（401 vs 403 语义）、LDAP 三分（bind/过滤/组映射）、422 与重定向循环的反代头形态、账号锁定面 |
| `skills/gitlab-lifecycle-ops` | 生命周期运维：备份三件套可用性判定（tar+secrets 分离+容量）、secrets 丢失事故形态、升级路径停靠点红线、docker 形态回滚与 DB 兼容红线、license 面 |
| `gitlab-platform-analyst/` | GitLab 平台诊断专家（ask-ops 只读面） |

## 路由错位（重要）

- **CI 流水线失败、runner 作业/注册面**（构建失败、job 卡 pending、runner 掉线）→
  cicd-ops 交付链专家（gitlab-ci-triage 已覆盖），本包不做二义入口；
- **bundled PostgreSQL / Redis 本体深挖** → db-ops / mw-ops——本包只判「GitLab 视角
  的依赖面」（readiness 分项失败落位）；
- **代码审查/MR/issue 使用侧问题**（.gitlab-ci.yml 怎么写、权限怎么申请）→ 不走本包；
- **泛化 GitLab DevOps 咨询** → oo-devops 参考库 gitlab-employee（P6 低优先，本包
  P3 平台本体面优先承接）。
- 入口词全部带 gitlab 实例/仓库/登录/备份升级等复合限定词，与 cicd-ops（构建/流水线/
  runner）、n8n-ops（实例打不开/加密密钥丢失）语义相邻域刻意错开。

## 工具面

单层采集（ask-ops 受审只读命令面）：

- **健康三件**（无需认证）：`/-/health`（进程活着）、`/-/liveness`、`/-/readiness`
  （依赖分项 JSON：gitaly/db/redis/sidekiq/repositories——分项就是实例面的地图）；
- **GitLab API**（curl GET）：`/api/v4/version`、projects/runners/users/license 等；
  PAT 经主机侧凭证注入（采集机环境变量或客户端配置），对话不收集不回显；管理面
  端点需 admin token，403 时标注可得性不臆测；
- **主机侧**：`systemctl status gitlab-runsvdir`（omnibus runit 总控）、
  `journalctl --since`、`ls/cat/grep/tail`（/var/log/gitlab 分服务日志）、
  `df/du/stat`（盘水位与 secrets 时间戳）；
- **docker 形态**：`docker ps/logs --tail/inspect/stats --no-stream`（Env 值引用前打码）；
- **runner 主机**：`gitlab-runner list/status/verify/version/check`（工具面只读守卫，
  需含 gitlab-runner 守卫的 bianque-tools）。

**不在只读白名单**（一律审批后宿主侧动作进 `recommendation.steps`，禁处方化）：
`gitlab-ctl`（status/tail/reconfigure）、`gitlab-rake`（gitlab:check / gitlab:ldap:check /
backup:restore）、`gitlab-rails console`、`gitlab-psql`、`docker exec`。

本包无链：GitLab 故障是事件驱动排障（实例/仓库/认证/生命周期四面），无例检语义，
走专家路由即可。v0.1.0 未编目变更块，处置全走自由 steps（`requires_approval` 恒 true）。

## MCP 治理面（v0.2.0 起）

包内 `mcp/gitlab.yaml` 声明只读 GitLab MCP 桥（吃主仓批次一百八十七治理面）：

1. 站点自备桥二进制并 export `BQ_GITLAB_MCP_BIN=<路径>`（未部署=启动自动剔除不空转）；
2. 控制台 MCP 页对该 server 点「激活」（确认语义=开洞，热生效）；
3. 同页「绑定凭证」选 `gitlab:*`（全部启用实例）或 `gitlab:<实例名>`——凭证平面物化
   注入 GITLAB_URL/GITLAB_TOKEN，改绑定即热重物化。

未激活/未装桥时专家仍走 ask-ops 只读宿主面（本包主体方法论不依赖 MCP 桥）。
