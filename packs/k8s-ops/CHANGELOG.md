# CHANGELOG

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
