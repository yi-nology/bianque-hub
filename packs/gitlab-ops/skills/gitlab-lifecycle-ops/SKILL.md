---
name: gitlab-lifecycle-ops
description: GitLab 生命周期运维方法论：备份三件套可用性判定（tar 产物+secrets 分离+容量）、secrets 丢失事故形态、升级路径停靠点红线与回滚的 DB 兼容红线、docker 形态版本证据、license 面——固定顺序定位与判读基准（ask-ops 只读采集面）。
mode: on_demand
version: 0.1.0
maturity: experimental
requires_mcp:            # 采集依赖的工具面（装载期对账依据；旧 bianque-tools 二进制缺这些工具时应先重建）
  - server: ask-ops
    tools: [run_readonly_command, run_readonly_commands]
---

## 触发条件

- 症状关键词：gitlab 备份、备份失败、恢复演练、gitlab 升级、升级失败回滚、gitlab 密钥丢失、gitlab-secrets、500 全站且 CI 变量解不开、license 过期
- 组合场景：盘满导致备份失败先过盘面（gitlab-instance-triage / gitlab-repo-triage 的 df 证据链）；
  升级后 readiness 不就绪的迁移判定联动 gitlab-instance-triage

## 数据来源（ask-ops 只读面）

- 备份产物面：`ls -lt /var/opt/gitlab/backups | head -10`（omnibus 缺省备份目录；产物命名
  `<时间戳>_<gitlab_version>_<时间戳>_backup.tar`——**文件名自带版本号**=恢复兼容性证据）、
  `df -h /var/opt/gitlab/backups`；备份计划面：`systemctl list-timers`、`grep -r backup /etc/cron.d`（omnibus crond 形态；
  文件不存在时 grep 报错属正常，不是采集失败）；
- secrets 面（**存在性与时间戳，禁读内容**）：`ls -l /etc/gitlab/gitlab-secrets.json`、
  `stat -c '%y %s %n' /etc/gitlab/gitlab-secrets.json`、`stat -c '%y %n' /etc/gitlab/gitlab.rb`（reconfigure 时间线对照）；
- 版本面：`curl -s -H "PRIVATE-TOKEN: $GITLAB_TOKEN" 'http://127.0.0.1/api/v4/version'`——升级路径
  停靠点核对的锚点（官方 upgrade path 文档）；
- docker 形态：`docker inspect --format '{{.Config.Image}}' gitlab`（镜像 tag=版本证据）、
  `docker images | head -20`（回滚 tag 在不在）；
- 备份配置面（选择性取键）：`grep -E 'backup_(path|keep_time|upload)' /etc/gitlab/gitlab.rb | grep -v '^#'`；
- license 面（admin PAT）：`curl -s -H "PRIVATE-TOKEN: $GITLAB_TOKEN" 'http://127.0.0.1/api/v4/license'`
  （403 时标注可得性不臆测）；
- **`gitlab-rake gitlab:backup:create/restore` / `gitlab-ctl reconfigure` / `gitlab-psql` 不在只读白名单**：
  备份、恢复、升级执行动作一律审批后宿主侧，禁处方化。

## 分诊路径（固定顺序）

1. **「备份没跑」三分**：计划面没配（cron/timer 空——计划面证据）、计划在但失败（backups 目录无新产物 +
   盘水位/权限证据）、在跑但产物不可用（第 2 步）——先计划面后产物面。
2. **「备份不可用」三查**：只有 tar 没有 gitlab-secrets.json 独立备份（**secrets 与 tar 分离是铁律**——secrets
   丢了 tar 恢复出来也是 500 全站）；tar 中断残片（文件名无时间戳后缀、0 字节、无最新 mtime）；跨版本恢复
   红线（恢复目标版本必须 ≥ 备份文件名里的版本且沿升级路径停靠点，**低版本备份不能直接进高版本主**
   的相邻跳变核对）。
3. **secrets 丢失事故形态**：全站 500 + 日志大量解密错误 + CI 变量/JWT/token 类功能失灵——三证据齐即
   定性；处置是恢复 secrets 独立备份，**恢复后存量 token/CI 变量/JWT 签名材料需批量重置**（变更建议列清单）。
   secrets 的 mtime 突变（对照 gitlab.rb reconfigure 时间线）本身就是事故线索。
4. **升级面**：`/api/v4/version` 对照官方 upgrade path 核对当前→目标的**停靠点序列**（跨多版本必须逐
   停靠点升，跳版本升级是数据事故第一来源）；升级前三件：备份完成、停机窗口、迁移队列——升级后
   background migrations 未完成不强行回滚；**回滚红线：新版 DB 已迁移后不可原地降级，只能恢复备份**
   （docker 形态回 tag 也受此约束——先核对 DB 迁移是否已跑）。
5. **license 面**：过期形态（部分功能关停/推送受限，实例仍可用）与失效形态（全站降级只读）分开——
   以 license API 与页面行为双证据定性；renew 属变更动作。

## 判读基准

- 「备份可用」必须三件套齐：tar 产物（版本号+时间戳双要素）+ secrets 独立异地副本 + 恢复目标容量核对
  （df）——缺一即 WARN/CRIT 定级，不给出「备份没问题」的笼统结论；
- 备份文件名里的版本号是恢复兼容性的第一证据，先读文件名再谈恢复；
- 升级红线三条：**无备份不升级、跳停靠点不升级、DB 已迁移不回退**——三条都来自官方升级路径语义，
  处方任何升级动作前先核这三条；
- secrets 结论只写「有无/时间戳/是否分离备份」，文件内容禁入报告。

## 输出要求

- 每个结论附产物清单/时间戳/版本号证据；恢复、升级、重置 token、license renew 全部进
  recommendation.steps 且 requires_approval 恒 true（升级类处方按停靠点逐条给单条命令，不合并成
  复合命令）；数据不足输出「需补充采集」清单（如备份产物完整清单、secrets 副本所在位置、
  升级窗口与迁移队列状态），不臆测。
