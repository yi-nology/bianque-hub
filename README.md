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
│   ├── cicd-ops/           # 交付链诊断包（Jenkins / GitLab CI / Harbor / ArgoCD）
│   ├── obs-ops/            # 可观测自诊包（Prometheus / Grafana / ES / Loki——监控失明场景）
│   ├── n8n-ops/            # n8n 自托管平台运维包（实例/执行/webhook/队列模式/生命周期）
│   └── chain-starter/      # 最小示例包（贡献模板：专家+技能+链逐文件拆解）
├── knowledge/              # 平台速查知识副本（platform-matrix 等，供非扁鹊运行时）
├── validator/              # 结构契约校验器（CI 与本地同款，独立无平台依赖）
```

## 装进扁鹊

扁鹊侧自带导入工具（`bq-markettool install`），两种模式：

```bash
# 推荐：经本机扁鹊实例的插件 API 安装（原子校验、失败回滚、插件页可管理）
go run ./cmd/bq-markettool install \
  --url https://github.com/yi-nology/bianque-hub \
  --pack os-basics \
  --api http://127.0.0.1:8900

# 无服务在跑时：直接落盘到 experts/（重启或控制台「重载」生效）
go run ./cmd/bq-markettool install --url <本仓> --pack os-basics --dest /path/to/experts
```

已装过的包再次导入即**升级**；`os-basics` 社区版与扁鹊内置同名包**等位可接管**
（slug/路由/技能完全一致，升级后原位替换，卸载并重启自动还原内置版）。

也可以在扁鹊控制台「插件」页上传/安装本仓包目录的 zip，或直接调 API：

```bash
curl -X POST http://127.0.0.1:8900/api/v1/plugins/install \
  -H 'Authorization: Bearer <admin-token>' -H 'Content-Type: application/json' \
  -d '{"path": "/path/to/bianque-hub/packs/os-basics", "operator": "you"}'
```

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
| [os-basics](packs/os-basics/) | 1.3.3 | 巡检/性能/内存/IO/网络/安全六域专家 + 安全四技能 + K8s 诊断（k8sgpt 桥）+ 主机快查链 |
| [k8s-ops](packs/k8s-ops/) | 0.1.4 | K8s 集群诊断：工作负载分诊/事件时间线/Helm/证书四技能 + 集群巡检链（k8sgpt + ask-ops 只读命令面）+ 编目变更块三件（回滚/续期/滚动重启） |
| [mw-ops](packs/mw-ops/) | 0.2.1 | 中间件诊断：Redis 五类分诊/热Key大Key + Kafka/RabbitMQ 积压与集群面 + 中间件例检链 |
| [db-ops](packs/db-ops/) | 0.2.2 | 数据库诊断：MySQL 连接/锁/慢查询/复制链 + PG 膨胀/WAL/复制槽 + 数据库例检链 |
| [cicd-ops](packs/cicd-ops/) | 0.2.1 | 交付链诊断：Jenkins/GitLab CI/Harbor/ArgoCD 四排障技能（构建→推送→同步三段定位） |
| [obs-ops](packs/obs-ops/) | 0.3.1 | 可观测自诊：Prometheus 采集/规则面 + Grafana 数据源/面板面 + ES 集群/分片面 + Loki 日志查询面（监控失明场景） |
| [n8n-ops](packs/n8n-ops/) | 0.2.0 | n8n 自托管平台运维：实例本体/执行面/webhook 触发面/队列模式/生命周期五技能 + 平台诊断专家（datasources n8n 查询面 + ask-ops 主机面） |
| [chain-starter](packs/chain-starter/) | 0.1.0 | 最小示例包（贡献模板，含技能钉扎演示） |

## oo-devops 迁移说明

扁鹊仓原 `experts/oo-devops`（openocta 市场归化整包：31 员工 + 189 技能，构建产物
禁手改）**已冻结，不再整包承接**。其中有诊断价值的方法论按域收割重写为本仓四包：
**mw-ops / db-ops / cicd-ops / obs-ops**（血缘与未搬运清单见各包 provenance.json 与
README 改造说明）。该域能力的后续维护只在对应领域包进行——改包内资产、升版本、
走 PR，不再经市场转换管线。

## 本地校验

```bash
go run ./validator ./packs   # CI 同款：结构契约 + 跨包唯一性 + 工具面覆盖 + 版本同步 + 脱敏扫描（自动覆盖仓库根）
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
go run ./cmd/bq-markettool seed-sync --hub <本仓路径> --repo . --packs os-basics --apply
go run ./cmd/pack-lint   # 内嵌 seed 校验（在无 experts/ 的目录跑才走内嵌源）
go build                 # seed 是 go:embed，改完必须重编译
```

## 许可

[Apache-2.0](LICENSE)
