# CHANGELOG

## 0.1.4 (2026-10-03)

- 批次九十四「诊断+处方+受审执行」接入（第三波）：新增编目变更块三件——`k8s-helm-rollback`
  （release 应急回滚，params: release/revision/namespace，verify: helm history）、
  `k8s-kubeadm-certs-renew`（证书逐张续期，params: certs 清单，verify: check-expiration；
  控制面静态 Pod 重启不进编目走自由 steps）、`k8s-rollout-restart`（滚动重启，
  params: resource/namespace，verify: kubectl get）；三技能 frontmatter 登记
  `provides_changes` 并在输出要求给 change_ref 处方口径（编目优先、自由 steps 兜底；
  helm-ops/triage 0.1.1→0.1.2、workload-triage 0.1.0→0.1.1）；专家提示词处方铁律
  升为编目优先。配套：bianque-tools ask-ops 只读白名单同批补 helm/kubeadm 守卫
  （本包 0.1.3 已声明的只读面此前实际被白名单拒收降级）。

## 0.1.3 (2026-10-02)

- 工具面补全（0.1.2 遗漏）：路由词 helm排查/k8s证书巡检 已开，但专家只授了 k8sgpt——
  增授 ask-ops 只读命令面；k8s-helm-ops/k8s-certs-ops 补 requires_mcp 声明
  （ask-ops run_readonly_command，技能升 0.1.1，certs 补数据来源段）；提示词新增
  「宿主侧只读命令面」纪律（白名单外被拒不绕过、降级方法论输出）。

## 0.1.2 (2026-10-02)

- 工作负载专家新增「采集脱敏」纪律（与四领域包同批）：kubectl describe 输出的
  环境变量等凭证面字段引用前打码；Secret 材料面已被 run_readonly_command 守卫
  拒收，禁绕过。

## 0.1.1 (2026-10-01)

- 同源工具契约落位：analyst 提示词+workload-triage 技能补 explain:false 契约与零发现口径纪律。

## 0.1.0 (2026-10-01)

- 首发（rd1/106 集群实弹驱动）：k8s-workload-triage/k8s-events-timeline 原创两技能
  + k8s-helm-ops/k8s-certs-ops（oo-devops 改造去工具耦合）+ imports/k8s-workload-analyst
  专家 + workflow/k8s-cluster-audit 集群巡检链（跨包引用 k8s-health，技能钉扎演示）
  + mcp/k8sgpt.yaml 0.4.39 契约清单。
