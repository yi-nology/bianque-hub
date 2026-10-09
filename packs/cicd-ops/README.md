# cicd-ops：交付链诊断包（社区版）

Jenkins / GitLab CI / Harbor / ArgoCD 四技能 + 一专家。全部方法论面向**只读采集 +
诊断判读**，重跑/GC/sync/轮证书一律走平台审批面——其中 GitLab 重跑/取消与 ArgoCD
sync/rollback 已接编目变更块（命令模板钉死、审批卡可回退说明），其余处置走自由 steps。

| 资产 | 说明 |
|---|---|
| `skills/jenkins-pipeline-triage` | agent 掉线、队列堆积、JENKINS_HOME 磁盘、插件面分诊（oo-devops jenkins-ops 重写） |
| `skills/gitlab-ci-triage` | token 身份预检、job 卡 created/pending 两分、runner 掉线、流水线状态语义与 MR→流水线→trace 下钻、配额与存储面（oo-devops gitlab-ops 重写；0.2.0 起参考 GitLab 官方 CLI glab 方法论） |
| `skills/harbor-triage` | 拉取失败四归因、GC 窗口误删、磁盘水位、证书面（oo-devops harbor-ops 重写） |
| `skills/argocd-sync-triage` | SyncFailed/OutOfSync/Degraded 三态定位、漂移归因（oo-devops argocd-ops 重写） |
| `pipeline-analyst/` | 交付链分析专家（ask-ops 只读面） |

## 与其他包的错位

- **k8s-ops** 管 K8s 运行态（工作负载/事件/Helm release 执行面）；本包管**交付面**
  （构建→推送→同步）。镜像拉取失败从 K8s 侧症状先走 k8s-ops（凭证/镜像名/网络
  三分），落到制品库本体再进本包 harbor-triage；
- 部署目标资源不健康（同步已成功）是运行态问题，本包只给联动线索不代诊。

本包无链：交付链故障是事件驱动排障（构建失败/推送失败/同步失败），无例检语义，
不设链入口，走专家路由即可。

## 工具面

无专用 MCP 依赖：各平台 REST GET / CLI 只读命令（argocd、gitlab-runner、openssl
探测等）经平台 `ask-ops` 采集面在目标运维主机执行，凭证走主机侧登录态，对话不回显。
GitLab 流水线面的处方命令依赖主机侧 `GITLAB_URL`/`GITLAB_TOKEN` env（与 gitlab-ops
包的 gitlab 凭证类型同源，站点未录入时 curl 如实报错）。`glab` 二进制不在只读白名单——
方法论以 curl GET 原生形态承载，装包不引入 CLI 依赖。

## 安装

```bash
go run ./cmd/bq-markettool install --url <本仓> --pack cicd-ops --api <扁鹊实例>
```

## 改造说明

源自 openocta 收割的 oo-devops 市场包。取交付链域四技能收割重写：剥离原版环境
变量注入与 MCP 假设，统一扁鹊契约。gitlab-manager/azure-devops 等宽面技能族与
代码审查类（SecureCodeGuardian 等）暂未搬运，按需在后续版本承接。
