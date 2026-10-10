# bianque-hub

[扁鹊（bianque）](https://github.com/yi-nology)的**内容资产上游仓**：运维诊断的
**技能 / 专家 / 诊断链**以领域包（pack）形式在这里维护，社区通过 PR 持续更新，
扁鹊平台一条命令导入。

```
├── packs/                  # 可安装的领域包（每个子目录 = 一个包）
│   ├── os-basics/          # 主机/OS + K8s 诊断包（社区版，内置同名包的上游）
│   ├── k8s-ops/            # Kubernetes 集群诊断包（工作负载/事件/Helm/证书）
│   ├── mw-ops/             # 中间件诊断包（Redis 缓存 / Kafka / RabbitMQ）
│   ├── db-ops/             # 数据库诊断包（MySQL / PostgreSQL）
│   ├── web-ops/            # 网页巡检取证包（无头 Chromium 渲染采集——SPA 正文/截图/白屏分型）
│   ├── cicd-ops/           # 交付链诊断包（Jenkins / GitLab CI / Harbor / ArgoCD）
│   ├── obs-ops/            # 可观测自诊包（Prometheus / Grafana / ES / Loki——监控失明场景）
│   ├── n8n-ops/            # n8n 自托管平台运维包（实例/执行/webhook/队列模式/生命周期）
│   ├── docker-ops/         # Docker Engine 运维包（引擎/容器/镜像磁盘/网络/Compose 编排）
│   └── chain-starter/      # 最小示例包（贡献模板：专家+技能+链逐文件拆解）
├── index.json              # 包索引（机器可读清单：名称/版本/专家/构成计数；
│                           #   bq-markettool index-gen 维护，validator 对账）
├── knowledge/              # 平台速查知识副本（platform-matrix 等，供非扁鹊运行时）
├── validator/              # 结构契约校验器（CI 与本地同款，独立无平台依赖）
```

## 装进扁鹊

### 一条命令装包

扁鹊实例在本机运行时，只需要包名——来源缺省即本仓（官方 hub），
目标缺省自动探测 `http://127.0.0.1:8900` 走 API 模式（原子校验、
失败回滚、插件页可管理）：

```bash
bianque markettool install os-basics
```

工具随扁鹊二进制分发（部署机无需 Go 工具链）；在扁鹊仓内开发也可用独立壳
`go run ./cmd/bq-markettool install os-basics`。本机探测不到实例且未指定
目标时**不会静默落盘**——命令会列出 API/落盘两种模式供你选。

### 显式形态

```bash
# 远端/非缺省端口实例
bianque markettool install os-basics --api http://<扁鹊主机>:8900 --user admin --pass <口令>

# 包不在服务器同机（容器化部署）：zip 按值直传，免拷文件
bianque markettool install os-basics --api http://127.0.0.1:8900 --upload

# 无服务在跑：直接落盘到 experts/（重启或控制台「重载」生效）
bianque markettool install os-basics --dest /path/to/experts
#   已存在要覆盖升级：加 --force（旧版自动移入 <dest>/_trash/，免手删）

# 装本地 fork 的分支
bianque markettool install os-basics --url https://github.com/you/bianque-hub --ref my-fix
```

### 环境变量常驻（CI/常用免带参）

`BQMT_HUB_URL`（=`--url`）、`BQMT_API`（=`--api`）、`BQMT_TOKEN`（=`--token`）、
`BQMT_USER`+`BQMT_PASS`（=`--user`/`--pass`）：仅在对应 flag 未给出时生效，
凭据不落命令行历史。`bianque markettool`（不带参）看全部子命令与缺省规则。

已装过的包再次导入即**升级**；`os-basics` 社区版与扁鹊内置同名包**等位可接管**
（slug/路由/技能完全一致，升级后原位替换，卸载并重启自动还原内置版）。

也可以在扁鹊控制台「插件」页上传/安装本仓包目录的 zip，或直接调 API：

```bash
curl -X POST http://127.0.0.1:8900/api/v1/plugins/install \
  -H 'Authorization: Bearer <admin-token>' -H 'Content-Type: application/json' \
  -d '{"path": "/path/to/bianque-hub/packs/os-basics", "operator": "you"}'
```

只装某个专家（专家级子集安装）：

```bash
bianque markettool install k8s-ops --expert k8s-workload-analyst
```

`--expert` 可重复；每个专家安装为**独立插件**（`<包>.<专家>`，如
`k8s-ops.k8s-workload-analyst`），独立升级/卸载/启停。装机前可先看清单：
`bianque markettool experts k8s-ops`（列专家+闭包摘要+可安装性）；
`--expert-all` 一键把整包所有专家各自装成独立插件；`pack-expert` 出迷你包 zip
（控制台「插件」页上传路径同享）。插件页对迷你包显示血缘徽标
（`derived_from`：源包/版本/专家；源包升级后未重派生提示 `derived_stale`）。
口径：与整包安装**互斥**（双向预检拒绝——已装整包时如只需某专家，用插件页
专家级启停，或先卸整包；反之迷你包在位装整包同样预检拒绝）；链（chain.yaml）
与跨专家消歧是包级资产，不随专家级安装落位。

### 卸载与启停

```bash
bianque markettool uninstall os-basics               # 卸载：目录入 experts/_trash/ 可找回，core 内置包治理拒卸
bianque markettool disable k8s-ops                   # 包级禁用（加载期整包跳过）；enable 还原
bianque markettool uninstall k8s-ops.k8s-workload-analyst   # 迷你包按名同样可卸/可停
```

`bianque markettool list` 速览本机装了啥（版本/启停/专家/血缘/漂移）。卸载/启停与
安装共用同一套目标与凭据口径（`--api`/`BQMT_API` 自动发现、`BQMT_TOKEN`/
`BQMT_USER`+`BQMT_PASS` 兜底）。服务没起的机器清落盘残留：
`uninstall <名> --dest /path/to/experts`（目录同样移入 `_trash/`，不注销登记）。

## 索引与检索

hub 仓根的 `index.json` 是**机器可读包索引**（对标 Claude Code 插件市场的
marketplace.json 形态），`bq-markettool` 的检索面消费它（`--url` 缺省=官方 hub，
下列命令在任意目录可直接跑）：

```bash
# 检索（有索引用索引；缺索引自动降级现场扫描=旧克隆兼容）
bianque markettool search k8s

# 单包详情（构成/专家闭包摘要+可安装性/凭证前提）
bianque markettool info k8s-ops

# 落后包一键批量升级（对标 npm update；缺省自动探测本机实例）
bianque markettool update --user admin --pass <口令>

# 已装包 vs hub 版本对照（--api 读插件登记面 / --dest 扫落盘目录；迷你包按血缘对源包版本）
bianque markettool outdated --user admin --pass <口令>

# 检索/对照面（search/info/experts/outdated/list）支持 --json：schema 版本化输出，
# 脚本与智能体直接消费（--json 下人类提示全静默，stdout 即纯 JSON）
bianque markettool search --json k8s | jq '.results[].name'
```

- **索引维护**：包内容变更后由维护者 `bq-markettool index-gen --hub <本仓>` 全量
  刷新提交；`validator` 对账门禁校验索引与 `packs/` 一致（漏包/多录/版本漂移=
  ERROR）。外部 PR 只需手改索引里对应包的 `version` 一行即可过门，全量刷新由
  维护者收口。
- **版本落账**：install 会把 `resolved_ref`/`resolved_commit`/`pack_version` 写进
  provenance.json——装完可回答「装的是哪个 commit」，`outdated` 与血缘漂移提示
  消费该账。安装正确性不依赖索引（install 仍现场扫目录）。

## 给其他智能体用

- `packs/*/skills/*/SKILL.md` 遵循开放技能格式（`name`+`description` 标准字段；
  `mode`/`version`/`maturity`/`requires_mcp` 为扁鹊扩展键，其他运行时可忽略）。
  任何支持 Agent Skills 的智能体（Claude Code / ZCode / opencode 等）可直接把
  `skills/` 目录挂为技能参考库。
- 专家 `prompt.md` 与 `chain.yaml` 面向扁鹊平台契约，其他运行时可参考方法论但
  不可直搬（输出协议依赖扁鹊报告 schema）。
- `knowledge/` 是平台无关的速查知识（如中文 Linux 发行版识别矩阵），可直接取用。

## 包清单

| 包 | 版本 | 内容 |
|---|---|---|
| [os-basics](packs/os-basics/) | 1.8.1 | 巡检/性能/内存/IO/网络/安全六域专家 + 安全四技能 + K8s 诊断（k8sgpt 桥）+ 主机快查链 |
| [k8s-ops](packs/k8s-ops/) | 0.2.0 | K8s 集群诊断：工作负载分诊/事件时间线/Helm/证书四技能 + 集群巡检链（k8sgpt + ask-ops 只读命令面）+ 编目变更块三件（回滚/续期/滚动重启） |
| [mw-ops](packs/mw-ops/) | 0.3.0 | 中间件诊断：Redis 五类分诊/热Key大Key + Kafka/RabbitMQ 积压与集群面 + 中间件例检链 + 编目变更块三件（UNLINK/清队列/位点重置） |
| [db-ops](packs/db-ops/) | 0.3.0 | 数据库诊断：MySQL 连接/锁/慢查询/复制链 + PG 膨胀/WAL/复制槽 + 数据库例检链 |
| [web-ops](packs/web-ops/) | 0.3.0 | 网页巡检取证：SPA 渲染正文/全页截图取证/白屏四分型（browser-ops 工具面，conf 缺省关） |
| [cicd-ops](packs/cicd-ops/) | 0.4.0 | 交付链诊断：Jenkins/GitLab CI/Harbor/ArgoCD 四排障技能（构建→推送→同步三段定位；GitLab 面含身份预检/状态语义/trace 下钻，参考官方 CLI glab 方法论）+ 编目变更块五件（argocd sync/rollback + GitLab 作业级/流水线级重跑与取消） |
| [obs-ops](packs/obs-ops/) | 0.4.0 | 可观测自诊：Prometheus 采集/规则面 + Grafana 数据源/面板面 + ES 集群/分片面 + Loki 日志查询面（监控失明场景）+ 编目变更块两件（flood 解锁/分配恢复） |
| [gitlab-ops](packs/gitlab-ops/) | 0.4.0 | GitLab 自托管平台运维：实例本体/仓库推拉/认证权限/生命周期四技能 + 平台诊断专家（ask-ops 只读面）+ gitlab 凭证类型与 mcp/gitlab 只读桥声明（控制台治理面激活+凭证绑定）；CI 流水线与 runner 面归 cicd-ops |
| [nginx-ops](packs/nginx-ops/) | 0.2.0 | Nginx 自托管运维：可用性/路由重写/TLS/生命周期四技能 + 平台诊断专家（ask-ops 只读面，reload 类变更恒审批后执行）；纯宿主形态不依赖凭证平面；K8s Ingress 与 openresty/lua 不在范围 |
| [n8n-ops](packs/n8n-ops/) | 0.3.0 | n8n 自托管平台运维：实例本体/执行面/webhook 触发面/队列模式/生命周期五技能 + 平台诊断专家（datasources n8n 查询面 + ask-ops 主机面）+ 编目变更块一件（容器重启） |
| [docker-ops](packs/docker-ops/) | 0.2.0 | Docker Engine 运维：引擎本体/容器运行面（退出码/OOMKilled）/镜像磁盘面（拉取失败/回收）/网络面/Compose 编排面五技能 + 平台诊断专家（ask-ops 只读面）+ 编目变更块三件（容器重启/内存上限在线调整/悬空镜像回收）；K8s 编排容器归 k8s-ops |
| [chain-starter](packs/chain-starter/) | 0.1.2 | 最小示例包（贡献模板，含技能钉扎演示） |
| [oo-devops](packs/oo-devops/) | 1.2.0 | OpenOcta 归化参考库（31 员工 + 189 技能 + 7 MCP 声明）——**不进缺省分发**：方法论已按域收割进四领域包，本包作参考库维护（A 案：hub 为内容唯一编辑点） |

## oo-devops 迁移说明

扁鹊仓原 `experts/oo-devops`（openocta 市场归化整包：31 员工 + 189 技能）**不再整包
承接**。其中有诊断价值的方法论按域收割重写为本仓四包：**mw-ops / db-ops / cicd-ops /
obs-ops**（血缘与未搬运清单见各包 provenance.json 与 README 改造说明）。该域能力的
日常维护在对应领域包进行；oo-devops 本体按 2026-10-05 A 案转为 hub 托管参考库——
内容唯一编辑点即 `packs/oo-devops/`（市场转换管线已退役，改动直接编辑并走 PR），
不进 seed-sync 缺省分发，整包安装走控制台 bundle 路径。

## 本地校验

```bash
go run ./validator ./packs   # CI 同款：结构契约 + 跨包唯一性 + 工具面覆盖 + 版本同步 + index.json 对账 + 危险内容扫描 + 脱敏扫描（自动覆盖仓库根）
```

校验单个包可直达包目录：`go run ./validator ./packs/chain-starter`。

## 参与贡献

新增/修改技能、专家、诊断链请读 [CONTRIBUTING.md](CONTRIBUTING.md)——复制
`chain-starter` 起步，校验器全绿即可提 PR。

## 维护者：同步进扁鹊内置 seed

本仓是包内容**唯一编辑点**；扁鹊仓的 `internal/agents/seed/` 只是随二进制分发的
新机引导快照（直接改 seed 会在下次同步被覆盖）。PR 合并后由维护者执行：

```bash
cd <扁鹊仓>
go run ./cmd/bq-markettool seed-sync --repo . --packs os-basics --apply
# （--hub 可缺省：在扁鹊仓内自动识别同级 bianque-hub，或常设 BQMT_HUB_DIR）
go run ./cmd/pack-lint   # 内嵌 seed 校验（在无 experts/ 的目录跑才走内嵌源）
go build                 # seed 是 go:embed，改完必须重编译
```

## 许可

[Apache-2.0](LICENSE)
