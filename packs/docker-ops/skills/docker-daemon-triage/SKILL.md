---
name: docker-daemon-triage
description: Docker 引擎本体分诊方法论：控制面可达性三级语义（CLI 报错/服务态/sock 面）、daemon 起不来首错定位（journalctl -u docker）、daemon.json 配置面核对（语法坏/存储驱动漂移/live-restore）、重启后容器不自启与 restart policy——固定顺序定位与判读基准（ask-ops 只读采集面）。
mode: on_demand
version: 0.1.0
maturity: experimental
requires_mcp:            # 采集依赖的工具面（装载期对账依据；旧 bianque-tools 二进制缺这些工具时应先重建）
  - server: ask-ops
    tools: [run_readonly_command, run_readonly_commands]
---

## 触发条件

- 症状关键词：docker 命令报 Cannot connect to the Docker daemon、dockerd 起不来、docker 反复退出、docker 升级后异常、storage driver 报错、重启宿主后容器没了/没自启、daemon.json 改完起不来
- 组合场景：容器退出/重启循环先走 docker-container-triage（引擎活着才有容器面）；宿主磁盘满连带引擎写失败先看 docker-image-triage 磁盘段——本技能管「引擎进程与配置本身」

## 数据来源（ask-ops 只读面）

- 控制面：`docker version`（Client/Server 分段——Server 段空 = 控制面不可达）、
  `docker info`（Storage Driver、Cgroup、Live Restore、Containers 计数）；
- 服务态：`systemctl status docker`（active/exit/restart-loop 与主 PID）、
  `systemctl is-enabled docker`（开机自启面）；
- 首错定位：`journalctl -u docker --since "2h ago" | tail -100`（引擎日志；升级窗口拉长 --since）；
- 配置面：`cat /etc/docker/daemon.json`（选择性读取：storage-driver/registry-mirrors/
  log-opts/live-restore 是诊断键，凭证类键只判有无）、`ls -l /var/run/docker.sock`（权限面）；
- 事件面：`docker events --since 30m --until 0s`（引擎活着时的事件流；--until 必带否则被拒）。

## 分诊路径（固定顺序：先控制面后服务态再首错，配置核对收尾）

1. **控制面可达性**：`docker version`——Server 段正常则引擎活着，本技能多半不是病灶
   （转容器面/镜像面）；报 Cannot connect 时看错误形态：「permission denied … docker.sock」
   = 权限面（采集用户不在 docker 组或 sock 挂载缺失），「Cannot connect」无权限字样 =
   引擎未起或 sock 路径非默认（DOCKER_HOST 形态核对）；
2. **服务态**：`systemctl status docker` + `systemctl is-enabled docker`——inactive = 引擎没起；
   activating(restart) = 反复崩溃循环，转第 3 步拿首错；enabled 缺失解释「重启宿主后容器没了」的一半；
3. **首错定位**（journalctl 最早错误行是引擎面的地图，常见指纹）：daemon.json 解析失败
   （json 语法坏/未知键 → 「unable to configure the Docker daemon with file」）；
   网桥/子网冲突（「failed to create bridge / could not find an available, non-overlapping
   IPv4 address pool」——转网络面子网段核对）；磁盘写失败（「no space left」→ 磁盘面）；
   存储驱动不兼容（升级/换内核后「error initializing graphdriver」）；
4. **配置面核对**：daemon.json 与 `docker info` 对账——Storage Driver 字段与 json 声明一致？
   声明了没生效 = 配置未加载（改完没 restart）；**换 storage-driver 需清空 /var/lib/docker，
   误改驱动声明会让既有镜像容器「全部消失」**——「docker 升级/改配置后镜像容器全没了」先查
   这一条（数据通常还在旧 driver 目录，恢复原声明即回）；
5. **容器自启面**：`docker inspect <容器> --format '{{.HostConfig.RestartPolicy.Name}} {{.State.StartedAt}}'`
   ——重启宿主后「消失」的容器：restart policy 为 no（默认）的容器本就不自启；
   live-restore 开启时引擎重启不连坐容器，关闭时引擎重启会重启全部容器——两种形态的行为差异要如实区分。

## 判读基准

- 「Cannot connect to the Docker daemon」≠ 引擎死：权限/sock 路径/采集机自身在容器内都报这个——
  先分级再归因，禁一步跳「dockerd 挂了，建议重启」；
- 引擎反复崩溃的第一嫌疑是**变更窗口**（daemon.json 手改、yum/apt 升级 docker、内核升级）——
  journalctl 首错时间点对齐变更时间线；
- daemon.json 建议修改一律审批后宿主侧动作（本技能不开任何写处方）；改配置→restart docker 的
  组合必须先给「引擎重启对在跑容器的影响面」结论（live-restore 开/关行为不同），写进 needs_followup；
- 数据不足输出「需补充采集」清单（sock 路径与权限、docker 版本、daemon.json 全文诊断键、
  变更时间线），不臆测。

## 输出要求

- 每个结论附 journalctl 首错行/daemon.json 键值/docker info 字段证据；重启引擎、改 daemon.json、
  换存储驱动一律标注影响面与回退路径并 requires_approval；数据不足输出「需补充采集」清单，不臆测。
