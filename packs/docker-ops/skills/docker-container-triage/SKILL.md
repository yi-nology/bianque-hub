---
name: docker-container-triage
description: Docker 容器运行面分诊方法论：状态全景（ps -a 重启计数）、退出码判读表（0/1/125/126/127/137/143 与 OOMKilled 双义拆分）、重启循环时间线（events + logs 首错）、OOMKilled 内存上限证据链、活容器资源水位（stats/top）——固定顺序定位与判读基准（ask-ops 只读采集面 + 编目变更块受审执行）。
mode: on_demand
version: 0.1.0
maturity: experimental
requires_mcp:            # 采集依赖的工具面（装载期对账依据；旧 bianque-tools 二进制缺这些工具时应先重建）
  - server: ask-ops
    tools: [run_readonly_command, run_readonly_commands]
provides_changes:        # 编目变更块（受审执行）：本技能方法论覆盖的处方编目
  - docker-container-restart   # 容器重启（params: container）——根因处置到位后的重启处方优先引用编目
  - docker-resource-update     # 内存上限在线调整（params: container/memory）——OOMKilled 缓解处方优先引用编目
---

## 触发条件

- 症状关键词：容器退出/挂了、容器反复重启/重启循环、容器起不来、容器 OOM/被杀、容器卡死无响应、CPU/内存占用高
- 组合场景：引擎本体问题先走 docker-daemon-triage；应用语义故障（n8n/GitLab 自身报错）转各自
  领域包——本技能管「容器进程视角」的生命周期与健康

## 数据来源（ask-ops 只读面）

- 全景面：`docker ps -a`（状态列 exited/restarting/up + Names；`--format` 自选列控量）、
  `docker inspect <容器>`（State 三键 ExitCode/OOMKilled/Error + RestartCount + StartedAt/FinishedAt）；
- 日志面：`docker logs --tail 200 <容器>`（-f 会被拒；空输出 = 应用没写 stdout，
  转宿主文件 `docker inspect --format '{{.LogPath}}'` 后 ls -l 看轮转与体积）；
- 资源面：`docker stats --no-stream`（无 --no-stream 会被拒）、`docker top <容器>`（进程树）；
- 时间线：`docker events --since 30m --until 0s`（die/oom/kill/restart 事件序列；--until 必带）、
  `journalctl -k --since "1h ago" | grep -i -E "oom|killed process"`（宿主内核 OOM 面；
  受 kernel.dmesg_restrict 限制取不到时如实标注）；
- 配置面：`docker inspect` 的 HostConfig（Memory/MemorySwap/RestartPolicy/OomScoreAdj）、
  Healthcheck 段（test/interval/retries 与 Health.Log 失败序列）。

## 分诊路径（固定顺序：先全景后状态键，退出容器查三键，循环查时间线，OOM 查证据链）

1. **全景定位**：`docker ps -a` 锁定目标容器与状态形态——exited（退过）、restarting（循环中）、
   created 未起（起不来）、up 但卡死（活而不健）——形态决定后续入口；
2. **退出容器查 State 三键**：ExitCode 按判读表归因（0 正常 / 1 应用错误 / 125 docker 自身 /
   126 不可执行 / 127 命令不存在 / 137 被 SIGKILL / 143 被 SIGTERM）+ `Error` 字段原文 +
   `logs --tail` 尾部首错——**137 必须看 OOMKilled 拆双义**：true = 内存上限杀（转第 4 步），
   false = 人为 kill -9 或内核 OOM（转时间线对齐）；entrypoint 缺执行位/解释器错是 126/127 的常客；
3. **重启循环查时间线**：`docker events --since --until` 的 die→restart 序列 + RestartCount +
   每次 logs 尾部首错——「每次起几秒就死」= 应用启动即崩（配置/依赖缺）；
   healthcheck 失败连坐重启（inspect Health.Log 失败序列 vs 应用日志）要如实区分「应用死」与「探活误杀」；
4. **OOMKilled 证据链**：inspect 的 OOMKilled:true + HostConfig.Memory/MemorySwap（上限值）+
   stats --no-stream 当前水位 + journalctl -k OOM 行——四点齐才下「内存上限不足」结论；
   根因处置（应用内存泄漏查应用面）到位后，缓解处方走编目 `docker-resource-update`
   （params: container/memory，内存上限在线调高）；
5. **活容器资源面**：stats --no-stream 排序找 CPU/内存尖峰容器 + docker top 看进程树异常
   （僵尸/失控子进程）——「活而不健」的容器按应用面转域，容器面只给资源事实。

## 判读基准

- 退出码 137 双义是本技能最高频误诊点：OOMKilled:false 的人为 kill 与 OOMKilled:true 的上限杀
  处置完全不同——禁见 137 就建议加内存；
- RestartPolicy=always 的容器退出后由引擎拉起，「退出又活了」不等于自愈——看 RestartCount 与
  事件密度；手动 `docker run --rm` 的临时容器退出即删，误判「容器消失」要先看 run 形态；
- **受审执行处方**：根因处置到位后的容器重启优先引用编目变更块——
  `recommendation.change_ref: docker-container-restart` + `change_params: {container: "<容器名>"}`
  （不给自由 steps，命令本体由编目模板固定；前置纪律同 n8n 包：先有根因结论，重启只是让处置
  生效，未查因重启只会复现）；OOMKilled 缓解优先引用 `docker-resource-update` +
  `change_params: {container: "<容器名>", memory: "<2g 等 docker 量化>"}`（rollback 依据诊断
  报告里的原 HostConfig 值）；变更后 `docker ps --filter` / `docker inspect` 自动验证。
  编目未覆盖的动作（stop/rm/重建容器、改 restart policy）仍走自由 steps + requires_approval
  （审批卡标「未编目」）；编目未安装的站点照常走自由 steps，语义不变；
- 数据不足输出「需补充采集」清单（容器名/ID、run 或 compose 形态、完整 State 段、
  应用日志尾部），不臆测。

## 输出要求

- 每个结论附 State 三键/事件序列/stats 水位证据，退出码结论必附判读表归因；重启、重建、
  资源调整一律标注影响面与回退路径并 requires_approval；数据不足输出「需补充采集」清单，不臆测。
