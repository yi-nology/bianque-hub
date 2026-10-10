# CHANGELOG

## 0.2.0 (2026-10-10)

- **k8s-workload-analyst 工具授权最小权限收敛**（bianque 批次二百六十四，除 oo-devops 外全 hub 清扫）：
  `ask-ops allow: []`（=全部 14 工具，含破坏性 run_approved_command(s)）改显式 12 工具
  只读白名单（十结构化采集器+probe_host+run_readonly 只读 CLI 对）。修正三处不合理：
  ①授权面与 prompt 能力声明矛盾（分析专家=只读采集判读域，prompt 零变更工具引用，
  处置建议只随报告进审批）；②`allow: []` 使工具集含变更类，agentrun 重试守卫整步骤
  收紧（RetryAfterMutation=false，429 限流拒重试环节 failed 实弹在案，bianque 侧
  agentkit v0.14.6 豁免缝治标、授权收敛治本）；③能力/路由事实源失真——工具清单是
  派工与路由的能力面，超授即误描述。域 CLI 采集面（run_readonly 白名单内 mysql/psql/
  redis-cli/kafka/docker/kubectl 只读子命令等）完整保留，零能力损失。

## 0.1.5 (2026-10-06)

- 深审修复批：`k8s-certs-ops` 续期处方 change_params 口径注解明确（string_list 值逗号分隔、
  平台按值逐张展开——经平台 Render 语义核实，模板本身无误）；
  静态 Pod 重启机制表述修正（续期不改 manifest、不会自动重载，须显式重启）；
  `k8s-helm-ops` 回滚决策步收口到编目口径（去掉内联 `helm rollback` 命令字面）；
  `k8s-events-timeline` 补数据来源段 + dmesg 佐证归 ask-ops 只读面（版本 0.1.0→0.1.1）；
  `k8s-workload-triage` requires_mcp 补 analyze（工具契约段约束的正是它，0.1.1→0.1.2）；
  README 前置段补链对 `specialists/k8s-health` 的跨包依赖声明。
- 215 实弹口径前移：k8s-workload-analyst 输出铁律补「steps 每条=单条可执行 shell 命令」
  纪律（说明性内容归 needs_followup）。
- 215 实弹修复：k8s-workload-analyst max_iterations 14→28（实测 k8sgpt 全景扫描+
  逐发现 get-resource/get-logs 深挖把 14 轮打爆 → exceeds max iterations 整会话失败
  兜底；52 次会话级 LLM 调用为证）。



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
