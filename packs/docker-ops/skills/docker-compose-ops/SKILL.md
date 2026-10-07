---
name: docker-compose-ops
description: Docker Compose 编排面分诊方法论：compose CLI 不在只读白名单的替代采集法（容器 label com.docker.compose.project/service 过滤 + compose 文件 cat 对账）、依赖顺序故障（depends_on 与 healthcheck condition）、配置漂移对账（compose 文件 vs 运行态 inspect）、down/up 的数据面红线——固定顺序定位与判读基准（ask-ops 只读采集面）。
mode: on_demand
version: 0.1.0
maturity: experimental
requires_mcp:            # 采集依赖的工具面（装载期对账依据；旧 bianque-tools 二进制缺这些工具时应先重建）
  - server: ask-ops
    tools: [run_readonly_command, run_readonly_commands]
---

## 触发条件

- 症状关键词：docker compose 报错/起不来、compose 某服务退出、up 之后服务数量不对、依赖服务没就绪、改了 compose 不生效、compose 升级后异常、服务重建失败
- 组合场景：单容器退出/循环先走 docker-container-triage；n8n compose 形态的应用语义走 n8n-ops
  ——本技能管「compose 编排层」的共性问题（工程清单/依赖/漂移）

## 数据来源（ask-ops 只读面）

- **采集原则：`docker compose` 子命令不在只读白名单**——一切 compose 状态经原生 docker 面
  + **label 过滤**等价采集：`docker ps -a --filter label=com.docker.compose.project=<项目>`
  （工程容器全景：服务名/状态/Names）、`docker inspect <容器>` 的
  Config.Labels（com.docker.compose.project/service/working_dir/config_files 定位工程与清单文件）、
  `docker logs --tail 200 <容器>`（服务日志）；
- 清单文件：`cat <working_dir>/<config_files>`（compose 文件本身——services/depends_on/
  healthcheck/volumes 段是判读基准；路径从 label 拿，不猜）；
- 网络与卷：`docker network ls` + `docker network inspect <project>_default`（工程网络成员）、
  `docker volume ls`（工程卷清单）；
- 差异对账素材：`docker inspect` 的 Config.Env/Image/Mounts vs compose 文件 environment/image/
  volumes 段——漂移检测的两端事实。

## 分诊路径（固定顺序：先工程全景，后逐服务，依赖顺序与漂移各一段，红线收尾）

1. **工程全景**：从任一容器 label 定位 project 与 compose 文件路径 →
   `docker ps -a --filter label=com.docker.compose.project=<p>` 对照 compose 文件 services 段
   ——「少服务」（没起）/「多服务」（旧容器残留）/状态异常逐个标记；
2. **起不来的服务**：该容器 inspect State 三键 + logs 尾部首错（判读表见 container-triage）
   ——compose 层的特有指纹：报「service X depends on undefined service」= 清单语法坏
   （cat 文件核 yaml）；「container is unhealthy」= 依赖的 healthcheck 没过；
3. **依赖顺序故障**：compose 文件 depends_on + condition: service_healthy 段核对——被卡服务
   的日志是**依赖方的**（自己还没起）；依赖服务健康检查定义（healthcheck test 语义）与
   依赖方等待的 condition 是否同构，` unhealthy` 连坐是 compose 面最高频误诊点；
4. **配置漂移**：compose 文件 vs `docker inspect` 运行态对账（image tag/env 键值/mounts 路径）
   ——「改了 compose 不生效」的指纹：`up -d` 不重建未变更容器，改 env 后没 recreate；
   compose 文件路径与 label config_files 不一致 = 多份清单并存（改错文件）；
5. **红线收尾**：compose 变更动作（pull/down/up --build/scale）一律审批后宿主侧自由 steps——
   **`down` 删容器与网络但不删卷，`down -v` 连卷删（数据面红线，处方必须显式警示）**；
   处方按依赖序给全链命令，每条单条命令。

## 判读基准

- 服务名 ≠ 容器名：compose 容器名 `<project>-<service>-<序号>`（或自定义 container_name）
  ——编目重启块、日志命令的 container 参数一律取 `docker ps` 的真实 Names；
- 「compose 起不来」先分三层：清单语法坏（cat 即见）/ 依赖健康检查不过（依赖方日志）/
  服务自身崩（State 三键）——三层各有证据点，禁拿「重启 compose」当万能药；
- 项目多份 compose 文件（override/多环境）形态下，label config_files 是**实际生效清单**，
  用户口述的文件不是——漂移结论以 label 为准；
- 数据不足输出「需补充采集」清单（project 名、compose 文件全文、目标服务名、期望与实际
  服务清单），不臆测。

## 输出要求

- 每个结论附 label 过滤清单/compose 文件段/inspect 对账差异证据；pull/down/up/recreate 一律
  标注影响面与回退路径（down 与 down -v 的数据差异必须写进 needs_followup）并 requires_approval；
  数据不足输出「需补充采集」清单，不臆测。
