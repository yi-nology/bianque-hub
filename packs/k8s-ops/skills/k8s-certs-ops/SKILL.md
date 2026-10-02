---
name: k8s-certs-ops
description: K8s 证书双面排查方法论：集群面（kubeadm apiserver/etcd 证书到期巡检）与应用面（cert-manager Certificate/Order/Challenge 状态机）。宿主侧只读 CLI 采集（ask-ops 受审通道），只读优先。
mode: on_demand
version: 0.1.1
maturity: experimental
requires_mcp:            # 采集依赖的工具面（装载期对账依据；旧 bianque-tools 二进制缺这些工具时应先重建）
  - server: ask-ops
    tools: [run_readonly_command]
---

> 改造自 oo-devops/k8s-certs（openocta 收割）：原技能绑定 kubectl-mcp-server 的
> cert-manager 工具（该 MCP 面未随包供给）——本版去工具耦合改写为通用方法论。

## 触发条件

- 症状关键词：k8s 证书过期、apiserver 证书、cert-manager 排查、TLS 证书失败、Order 卡住
- 组合场景：安全巡检证书项（os-basics 安全域）在 K8s 集群语境下的深化

## 数据来源（ask-ops 只读面）

- 正文处方的宿主侧命令（`kubeadm certs check-expiration`、`kubectl get/describe issuer,clusterissuer,certificate,order,challenge` 等只读子命令）统一经 ask-ops `run_readonly_command` 受审执行；被白名单拒收时降级为方法论输出（标注需人工执行）；
- 证书文件本体与 Secret 材料面（`kubectl get secrets`、`config view --raw`）已被工具层拒收，不要尝试绕过。

## 集群面：kubeadm 证书巡检

1. `kubeadm certs check-expiration`——输出剩余天数表；**<90 天列观察清单，<30 天必须处置**；
2. 到期处置：`kubeadm certs renew all` 后**必须逐个重启控制面静态 Pod**（kube-apiserver/etcd/controller-manager/scheduler——manifest 变更触发自动重启，确认重启完成才算续期生效）；
3. 常见坑：手工签发的证书（非 kubeadm 管理）不在 check 列表——apiserver 报 `x509: certificate has expired` 但 kubeadm 显示正常时，查 apiserver 实际加载的 cert 文件（静态 Pod manifest 的 --tls-cert-file）。

## 应用面：cert-manager 状态机排查

1. **Certificate**：`READY=False` 时看 conditions——`Ordering=True` 卡签发、`Issueing` 循环=签了又不认；
2. **Order→Challenge 链**：Certificate 的事件指向 Order；Challenge 卡 dns-01/http-01 验证——dns-01 查 DNS provider 凭证与传播（TXT 记录 _acme-challenge），http-01 查 ingress 暴露与 solver 路由；
3. **Issuer 面**：`kubectl get issuer,clusterissuer` READY 状态；ACME 账号注册失败（Rate limit/网络）时所有 Order 连带失败；
4. **时间窗**：Let's Encrypt 生产环境有失败重试退避（1h/6h/24h）——反复重签被限速时先查为什么上一轮失败，别盲目 delete certificate 重触发。

## 判读基准

- 「证书过期」三义分清：**集群控制面证书**（ apiserver 不可用级故障）、**应用 ingress 证书**（浏览器告警/业务中断）、**webhook 证书**（特定操作报 x509，如 cert-manager 自身 webhook）——处置路径完全不同；
- NBF/开始时间与当前时间差大的新证书报 not-yet-valid，先查节点时钟同步。

## 输出要求

- 证书类结论必须给「哪一面、剩余天数/状态、处置路径」三要素；重续/重签类动作标注影响面（控制面重启/ingress 中断窗口）。
