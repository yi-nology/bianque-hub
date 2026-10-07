---
name: gitlab-instance-triage
description: GitLab 实例本体分诊方法论：健康端点三级语义（/-/health、/-/liveness、/-/readiness 分项即地图）、全站 502 归因（反代活/上游死）、起不来与反复重启首错定位、慢面（sidekiq 积压/盘水位）——固定顺序定位与判读基准（ask-ops 只读采集面）。
mode: on_demand
version: 0.1.0
maturity: experimental
requires_mcp:            # 采集依赖的工具面（装载期对账依据；旧 bianque-tools 二进制缺这些工具时应先重建）
  - server: ask-ops
    tools: [run_readonly_command, run_readonly_commands]
---

## 触发条件

- 症状关键词：gitlab 打不开、gitlab 502、gitlab 超时、gitlab 慢、升级后起不来、服务反复重启、sidekiq 队列堆积
- 组合场景：CI 流水线/runner 作业失败走 cicd-ops；readiness 分项里 db/redis 失败，其本体深挖走 db-ops/mw-ops——本技能管「GitLab 进程视角的实例健康」

## 实例地址判定（恒为第一步，判定错了整轮采集全是空手）

- 先读注入面：环境变量 `GITLAB_URL` 非空 = 平台凭证平面注入的**实例正式地址**，本技能全部
  curl 目标一律用 `$GITLAB_URL`（API 与健康端点同源），`GITLAB_TOKEN` 作为 API 凭证搭配使用；
- `GITLAB_URL` 为空 = 实例与采集机同机共置，curl 目标回落 `https://127.0.0.1`；
- 215 实弹教训：远程实例形态不判地址、对 127.0.0.1 探健康三件（全空）再转入泛化宿主采集，
  会把整轮迭代预算烧在没有 GitLab 的主机上——地址判定先行是本技能的第一处方。

## 数据来源（ask-ops 只读面）

- 健康端点三件（无需认证，第一采集面；目标地址按上文判定，远程用 `$GITLAB_URL`，共置回落
  `https://127.0.0.1`）：`curl -sk '$TARGET/-/health'`（进程活着）、
  `/-/liveness`（进程自判死锁将重启）、`/-/readiness`（依赖分项 JSON——gitaly/db/redis/sidekiq/repositories，
  **分项即地图**）；分项速览形态：`curl -sk 'https://127.0.0.1/-/readiness' | jq -r '.services[]? | "\(.name)=\(.status)"'`；
- API 版本与 sidekiq 面（需 PAT）：`curl -s -H "PRIVATE-TOKEN: $GITLAB_TOKEN" '$GITLAB_URL/api/v4/version'`、
  `'$GITLAB_URL/api/v4/sidekiq/queue_metrics'`（管理面需 admin token，403 时标注可得性不臆测）；
- omnibus 形态：`systemctl status gitlab-runsvdir --no-pager`（runit 总控——omnibus 服务不逐个注册 systemd unit，
  看总控而不是找 gitlab-*.unit）、`journalctl -u gitlab-runsvdir --since '-2h' --no-pager | tail -100`、
  `ls -lt /var/log/gitlab/gitlab-rails/`、`tail -100 /var/log/gitlab/gitlab-rails/production_json.log`；
- docker 形态：`docker ps -a --filter name=gitlab`、`docker logs --tail 200 gitlab`、
  `docker stats --no-stream gitlab`、`docker inspect gitlab`（Env 值引用前打码）；
- 资源面：`df -h /var/opt/gitlab`、`free -m`、`uptime`、`ps aux | grep -c '[r]unsv'`（runsv 数=在管服务数，omnibus 形态）；
- **`gitlab-ctl status` / `gitlab-rake gitlab:check` 不在只读白名单**：作为审批后宿主侧动作进建议面，禁处方化。

## 分诊路径（固定顺序，先端点后日志再资源）

1. **健康端点三级定性**：`/-/health` 通而 `/-/readiness` 不通 = **依赖面**（按 readiness 分项落位：
   db 失败→db-ops、redis 失败→mw-ops、gitaly 失败→gitlab-repo-triage、repositories 失败=仓库存储面）；
   两级都不通 = **进程/端口面**（服务未起、端口未监听、反代错位），转第 3 步。升级后 readiness 长时间
   不就绪是 **background migrations 在跑**的常态，先看时长与迁移日志再定性为故障。
2. **502 特判**：502 = 反代活着、上游死——查 puma 是否监听 socket（`ls -l /var/opt/gitlab/gitlab-workhorse/sockets/`）
   与 puma 首错（`tail -100 /var/log/gitlab/gitlab-rails/puma_stderr.log`、production_json.log）；反代 200 而
   页面 502 时查 workhorse（`tail -100 /var/log/gitlab/gitlab-workhorse/current`，omnibus runit 日志在
   `/var/log/gitlab/<服务>/current`）。反代层证据：`grep -c ' 502 ' /var/log/gitlab/nginx/gitlab-access.log`。
3. **首错定位**：反复重启看**每次启动前的最后 50 行**而不是最后一条日志（最后一条常是重启中的现象不是原因）；
   db 连接拒绝、端口占用、迁移未完成、secrets 解密失败是四大高频首错（secrets 形态转 gitlab-lifecycle-ops）。
4. **慢面**：CPU/内存饱和先分「谁在烧」——sidekiq 长任务（queue_metrics 的 backlog 与 latency）、
   web 请求洪峰、数据库慢查（转 db-ops）；磁盘水位（git-data、/var/log/gitlab）是隐性慢源（日志写盘阻塞、
   页面卡顿），`df -h` 全盘速览。资源结论必须带时间分布（持续 vs 突发、与业务高峰对齐）。
5. **版本与变更窗口**：`/api/v4/version` + 近 24h 变更时间线（升级/reconfigure/扩容）——升级后首批异常
   优先对官方已知问题核对版本，而不是运行态归因。

## 判读基准

- 「实例挂了」必须落到三级证据之一：**不可达**（进程/网络面）、**可达未就绪**（依赖/迁移面）、
  **就绪但异常**（负载/配置面）——禁笼统下结论；
- readiness 与 health 的差异本身就是结论素材（差在哪一层就是哪一层的病）；
- 502 不能只看反代层：定位到 puma/workhorse 才算落位；反代 access log 的 502 计数是影响面证据；
- sidekiq 积压判读分「队列堵」（特定队列 backlog 增长，对应任务类型阻塞）与「全局堵」（latency 全面上涨，
  sidekiq 进程/资源面），处置方向不同。

## 输出要求

- 每个结论附端点返回/日志关键行证据；变更类动作（重启 puma/sidekiq、gitlab-ctl reconfigure、扩容、
  清缓存）标注影响面并 requires_approval；`gitlab-ctl`/`gitlab-rake` 类核验与处置动作处方为单条
  宿主侧命令进 recommendation.steps；数据不足输出「需补充采集」清单（如部署形态、实例版本、
  readiness 完整 JSON、首错日志段），不臆测。
