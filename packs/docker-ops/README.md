# docker-ops：Docker Engine 运维包（社区版）

当 Docker 自己病了——daemon 起不来、容器退出/重启循环/OOM、镜像拉不下来、
磁盘被镜像和层吃满、端口映射不通、compose 服务起不来——本包接管诊断。
五技能 + 一专家；采集全走只读面，处置走编目变更块（受审执行）。

| 资产 | 说明 |
|---|---|
| `skills/docker-daemon-triage` | 引擎本体分诊：控制面可达性三级语义（CLI 报错/服务态/sock 面）、journalctl 首错定位、daemon.json 核对（语法坏/存储驱动漂移=「镜像容器全消失」/live-restore）、重启后不自启与 restart policy |
| `skills/docker-container-triage` | 容器运行面分诊：State 三键 + 退出码判读表（137 双义拆分）、重启循环事件时间线、OOMKilled 四点证据链、stats/top 资源水位；编目块：容器重启、内存上限在线调整 |
| `skills/docker-image-triage` | 镜像与磁盘面分诊：system df 四类水位、悬空镜像与 history 层膨胀、拉取失败三分（不可达/认证/限流）、回收梯度纪律；编目块：悬空镜像回收 |
| `skills/docker-network-triage` | 网络面分诊：端口发布两级核对（-p 映射 vs 容器内监听 127.0.0.1 指纹）、自定义网络内置 DNS 语义、network inspect 成员与 IPAM 子网冲突 |
| `skills/docker-compose-ops` | Compose 编排面分诊：compose CLI 不在只读面的 label 等价采集法、工程全景清点、depends_on/healthcheck 依赖连坐、compose 文件 vs 运行态漂移对账、down/-v 数据红线 |
| `docker-analyst/` | Docker 引擎诊断专家（P3 system，ask-ops 只读面 + 编目变更块受审执行） |

## 路由错位（重要）

- **K8s 编排的容器**（pod crashloop、驱逐、调度失败）→ k8s-ops（P2），本包不做二义入口；
- **宿主机本体**（CPU/内存/磁盘/内核的通用资源面）→ os-basics 系统巡检链——本包只判
  「Docker 视角」的资源面；215 形态分诊纪律（采集机自身可能是容器）前置；
- **应用平台本体**（n8n/GitLab/Nginx 自身业务故障）→ 各自领域包——本包管引擎与容器
  运行面共性；**Harbor 镜像仓库服务面** → cicd-ops；
- **镜像构建优化**（Dockerfile 写法、Buildx 多平台）是咨询面，不做诊断结论。
- 入口词全部带 docker 复合限定（docker容器退出/docker镜像拉取失败/docker compose报错…）
  + 中性实例态词（docker状态/docker健康/docker版本检查…）+ 空格变体（docker 容器/docker
  引擎…），与 k8s-ops（k8s 域词）、os-basics（巡检/健康体检）、cicd-ops（镜像推送失败）
  语义相邻域刻意错开。

## 工具面（只读采集 + 受审执行）

单层采集（ask-ops 受审只读命令面，dockerGuard 白名单为准）：

- **引擎态**：`docker version/info`、`systemctl status docker`、`journalctl -u docker --since`、
  `docker events --since X --until Y`（--until 必带）；
- **容器面**：`docker ps -a`、`logs --tail`、`inspect`、`stats --no-stream`、`top`、`port`、`diff`；
- **存储面**：`images`、`history`、`volume ls|inspect`、`system df`、`df/du /var/lib/docker`；
- **网络面**：`network ls|inspect`、固定清单面 `get_listening_ports` 交叉核对；
- **拉取通道**：`curl -s 'https://registry-1.docker.io/v2/'`（401=通，registry 语义）。

**不在只读白名单**（审批后宿主侧动作，禁处方化）：`docker exec/run/pull/cp`、
`docker compose` 全部子命令、`iptables/nsenter`、一切 `update/prune`/`system prune`。

**编目变更块**（受审执行：报告 `change_ref + change_params` → 平台审批 →
`run_approved_command` 渲染执行 → verify_readonly 自动验证）：

| slug | 触发场景 | risk |
|---|---|---|
| `docker-container-restart` | 僵死恢复/根因处置生效（params: container） | 2 |
| `docker-resource-update` | OOMKilled 后内存上限在线调整（params: container/memory） | 3 |
| `docker-image-prune` | 悬空镜像回收、磁盘水位恢复（无参数） | 3 |

编目未覆盖的变更（daemon.json 修改、容器重建、compose pull/down/up、prune -a 级深清、
卷删除）一律自由 steps + `requires_approval`（审批卡标「未编目」）；未安装编目的站点
照常走自由 steps，语义不变。

## 血缘与未搬运清单

方法论基于 Docker 官方文档原生编写；oo-devops 参考库 `docker-employee` 与
`docker-ops`/`docker-sandbox` 技能仅作素材对账（env 注入/MCP 示例/PowerShell 变体/
Docker Desktop sandbox 插件面未搬）。未搬运清单见 `provenance.json`：Swarm 集群编面、
Buildx 多平台构建、镜像安全扫描工具面（trivy/scout 装机自备）、cAdvisor/fluentd 部署面、
containerd/cri 运行时面、rootless/userns-remap 深度加固——按需后续 minor 承接。
