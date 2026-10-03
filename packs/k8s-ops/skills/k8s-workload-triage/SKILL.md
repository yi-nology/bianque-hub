---
name: k8s-workload-triage
description: K8s 工作负载异常分诊方法论：CrashLoopBackOff / OOMKilled / Pending / ImagePullBackOff / Evicted 五类异常的定位路径与判读基准（k8sgpt 工具面映射）。
mode: on_demand
version: 0.1.1
maturity: experimental
requires_mcp:
  - server: k8sgpt
    tools: [get-resource, get-logs, list-events, list-namespaces]
provides_changes:        # 编目变更块（批次九十四「受审执行」）：本技能方法论覆盖的处方编目
  - k8s-rollout-restart  # 工作负载滚动重启（params: resource/namespace）——处方优先引用编目而非自由 steps
---

## 触发条件

- 症状关键词：CrashLoopBackOff、OOMKilled、pod 重启、拉不起来、Pending、ImagePullBackOff、Evicted、工作负载异常
- 组合场景：节点压力传导（os-basics/k8s-node-diagnosis 联动）；Helm release 部署后异常（k8s-helm-ops 联动）

## 工具契约

- `analyze` 一律 `explain:false`——explain 模式把发现送往解释后端，平台部署的哑后端
  必拒连：**有发现的扫描整体报错、发现全丢**（rd1/106 实弹在案）；解释是宿主 LLM 职责。
- 零发现结论必须带过滤器口径；系统命名空间（kube-system/kube-public）ConfigMap
  「未被引用」是正常形态，不进结论主体。

## 分诊路径（按异常类定位，固定顺序）

1. **CrashLoopBackOff**：`get-logs`（previous=true 取崩溃前输出）→ 进程启动失败原因三查：配置错误/依赖不可达/权限；退出码 137=SIGKILL（内存），1=应用错误，0=立即退出（命令形态错）。
2. **OOMKilled**：`get-resource`（Pod）对比 limits.memory 与 lastState.terminated.exitCode=137；区分「limit 过低」与「真实内存泄漏」（重启间隔递减=泄漏特征）。
3. **Pending**：`list-events` 查调度事件三类——资源不足（Insufficient cpu/memory）、亲和/污点不匹配、PVC 未绑定；`get-resource`（Node）看 allocatable 余量。
4. **ImagePullBackOff**：events 中 registry 凭证（401）/镜像名（404）/网络（timeout）三分；私有仓库查 imagePullSecrets 是否挂。
5. **Evicted**：events 查驱逐原因（node.kubernetes.io/memory-pressure 等条件），归因到节点层（联动 os-basics/k8s-node-diagnosis）。

## 判读基准

- 重启次数 restartCount 需结合 pod 启动时长折算频率（30 分钟重启 10 次 ≠ 3 天重启 10 次）；
- 结论必须区分「工作负载层」（镜像/配置/资源声明）与「节点层」（资源压力/条件异常）归因——两层处置方向完全不同；
- 单 pod 异常先查是否 deploy/rs 滚动残留（ownerReferences 链），避免对已弃用副本做无效诊断。

## 输出要求

- 每个结论附 events/logs 关键行证据；无法定位时明确说「需节点层诊断」而不是猜测。

**受审执行处方（批次九十四）**：配置错误修复上线后、内存泄漏应急处置等「重启工作负载让它重读配置/清内存」场景，处方优先引用编目变更块——报告 `recommendation.change_ref: k8s-rollout-restart` + `change_params: {resource: "<type/name 形式，如 deployment/api-server>", namespace: "<ns>"}`（不给自由 steps，命令本体由编目模板固定；变更后 `kubectl get` 自动验证 READY 副本回升）；注意前置纪律：**重启前必须已有根因结论**（重启只是让处置生效的手段，不是止血本身——CrashLoop 未查因就重启只会再崩）。`kubectl delete pod`（驱散单副本）、scale 扩缩容等其余动作仍走自由 steps + requires_approval（审批卡标「未编目」）。编目未安装的站点照常走自由 steps，语义不变。
