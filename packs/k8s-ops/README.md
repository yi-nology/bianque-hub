# k8s-ops：Kubernetes 集群诊断包（社区版）

四技能 + 一专家 + 一链 + 工具面契约清单。与内置 os-basics 的 K8s 域（k8s-health
专家 + k8s-node-diagnosis 技能，节点侧）**错位互补**：本包补工作负载层与专项面。

| 资产 | 说明 |
|---|---|
| `skills/k8s-workload-triage` | CrashLoop/OOMKilled/Pending/ImagePull/Evicted 五类分诊（原创，k8sgpt 工具映射） |
| `skills/k8s-events-timeline` | 事件时间线归因：因果起点定位、Warning/Normal 分级（原创） |
| `skills/k8s-helm-ops` | Helm release 排查（oo-devops/k8s-helm 改造，去 MCP 耦合） |
| `skills/k8s-certs-ops` | 集群 kubeadm 证书巡检 + cert-manager 状态机（oo-devops/k8s-certs 改造） |
| `k8s-workload-analyst/` | 工作负载深诊专家（k8sgpt 7 工具只读面） |
| `chain.yaml` | `workflow/k8s-cluster-audit` 集群巡检链（全景→深挖两步，技能钉扎演示） |
| `mcp/k8sgpt.yaml` | k8sgpt 0.4.39 契约清单（12 工具事实记录，血缘/对账面） |

## 为什么需要链入口（设计动机）

意图路由对「检查」类词消歧到主机域巡检，K8s 诊断诉求会被劫去 SSH 采集；专家显式
绑定（expert_slug）也可能被出域症状词推翻（rd1/106 实弹在案，2026-10-01）。链入口
带自有关键词（k8s巡检链/集群巡检链等复合词）**直达两步流程**，不经意图路由——
这是 K8s 例检当前最可靠的入口形态。

## 前置（与 os-basics K8s 域一致）

> 链 `workflow/k8s-cluster-audit` 第一步引用跨包专家 `specialists/k8s-health`
> （平台内置或 os-basics 镜像）做全景采集——单装本包的站点须确认该专家在场，
> 否则链入口不可用（专家路由不受影响）。

1. k8sgpt 0.4.39 sidecar 挂目标集群 kubeconfig（见 os-basics README「K8s 集群诊断域」）；
2. 扁鹊 `mcp.servers.k8sgpt` url 模式指向 sidecar MCP 口；
3. 安装本包：`bq-markettool install --url <本仓> --pack k8s-ops --api <实例>`。

## 改造说明

helm/certs 两技能源自 openocta 收割的 oo-devops 包（原版绑定未供给的
kubectl-mcp-server 工具面）——本版重写为通用 CLI 方法论并在 provenance 留痕；
oo 的 k8s-manage（965 行多集群运维大全）体量与「诊断」定位不符，未搬运。
