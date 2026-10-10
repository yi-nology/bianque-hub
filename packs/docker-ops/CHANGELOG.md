# CHANGELOG

## 0.2.0 (2026-10-10)

- **docker-analyst 工具授权最小权限收敛**（bianque 批次二百六十四，除 oo-devops 外全 hub 清扫）：
  `ask-ops allow: []`（=全部 14 工具，含破坏性 run_approved_command(s)）改显式 12 工具
  只读白名单（十结构化采集器+probe_host+run_readonly 只读 CLI 对）。修正三处不合理：
  ①授权面与 prompt 能力声明矛盾（分析专家=只读采集判读域，prompt 零变更工具引用，
  处置建议只随报告进审批）；②`allow: []` 使工具集含变更类，agentrun 重试守卫整步骤
  收紧（RetryAfterMutation=false，429 限流拒重试环节 failed 实弹在案，bianque 侧
  agentkit v0.14.6 豁免缝治标、授权收敛治本）；③能力/路由事实源失真——工具清单是
  派工与路由的能力面，超授即误描述。域 CLI 采集面（run_readonly 白名单内 mysql/psql/
  redis-cli/kafka/docker/kubectl 只读子命令等）完整保留，零能力损失。

## 0.1.0（2026-10-07）

- 首发：五技能 + 一专家——引擎本体（docker-daemon-triage）、容器运行面（docker-container-triage）、镜像与磁盘面（docker-image-triage）、网络面（docker-network-triage）、Compose 编排面（docker-compose-ops）；专家 docker-analyst（P3 system，max_iterations 24，报告协议三硬约束钉扎）。方法论基于 Docker 官方文档原生编写，oo-devops 参考库 docker-employee/docker-ops/docker-sandbox 仅作素材对账（env 注入/MCP 示例/PowerShell/sandbox 插件面未搬）。
- 编目变更块三件（受审执行范式，宿主引擎域首批可读写）：`docker-container-restart`（risk 2）/ `docker-resource-update`（risk 3，OOMKilled 后内存上限在线调整）/ `docker-image-prune`（risk 3，悬空镜像回收）；技能 frontmatter `provides_changes` 接线，专家 prompt 钉 `change_ref + change_params` 编目优先口径。
- 采集面全部落在 ask-ops dockerGuard 只读白名单内（ps -a/logs --tail/inspect/stats --no-stream/top/events --since+--until/port/history/diff/images/network|volume|system 子面）；docker compose 子命令未入白名单——compose 面经容器 label（com.docker.compose.project）+ compose 文件 cat 采集；docker exec/pull/update/prune 一律审批后宿主侧动作，禁处方化。
- 路由词按 215 实弹三课配置：复合限定词（docker容器退出/docker镜像拉取失败…）+ 中性实例态词（docker状态/docker健康/docker版本检查…，防「检查」类消歧劫航）+ 空格变体（docker 容器/docker 引擎…，防 KeywordHit 空格打断）；symptoms 全部 docker 限定，避开 k8s-health（P2 OOMKilled/pod重启）与 cicd-ops（P3 镜像推送失败）。
