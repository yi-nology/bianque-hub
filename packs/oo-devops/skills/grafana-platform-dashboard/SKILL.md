---
name: grafana-platform-dashboard
description: 为OpenShift/Kubernetes平台运维设计、重构和验证Grafana dashboards。当用户要求改进platform health dashboards、优先处理关键的tenant-impacting signals、过滤噪声（例如ArgoCD）、添加Crossplane/Keycloak health panels、以编程方式验证PromQL，或将GrafanaDashboard CR更改实时apply然后promote到GitOps时使用。
mode: on_demand
version: 1.1.0
maturity: experimental
---

# Grafana Platform Dashboard

设计platform operations dashboards，使operators首先看到tenant-impacting risk，然后深入到service-specific health而不超载。

## Quick Start

当用户要求platform dashboard更新和reliability checks时使用此skill。

1. 确认dashboard target：
```bash
oc --context <ctx> get grafanadashboard -A | rg -i '<dashboard-name-or-theme>'
```
2. 导出dashboard和JSON：
```bash
skills/grafana-platform-dashboard/scripts/grafanadashboard_roundtrip.sh export \
  --context <ctx> \
  --namespace <ns> \
  --name <grafanadashboard-name> \
  --out-dir /tmp/<workspace>
```
3. 编辑JSON并验证所有PromQL：
```bash
skills/grafana-platform-dashboard/scripts/promql_scan_thanos.sh \
  --context <ctx> \
  --dashboard-json /tmp/<workspace>/<name>.json
```
4. 安全地实时apply：
```bash
skills/grafana-platform-dashboard/scripts/grafanadashboard_roundtrip.sh apply \
  --context <ctx> \
  --namespace <ns> \
  --name <grafanadashboard-name> \
  --json /tmp/<workspace>/<name>.json
```

## Workflow

### 1) 从Platform Contracts锁定Scope

在编辑panels之前使用 [platform-contract.md](references/platform-contract.md) 中的platform contract。

1. 将L1 command view限制为critical pre-tenant-impact signals。
2. 优先使用gate-aligned components（critical CO gate、nodes、MCP、core API/etcd/ingress）。
3. 将service-specific sections（Crossplane、Keycloak）保留在L1下方。

### 2) 强制执行Information Architecture

使用 [layout-guidelines.md](references/layout-guidelines.md)：

1. L1: critical-only、immediate action、minimal panel budget。
2. L2: 按dependency domain的platform services。
3. L3: deep dives（例如future GPU dashboard），不在L1中。

### 3) 从已知Library构建Queries

使用 [promql-library.md](references/promql-library.md)：

1. 从known-good queries开始并最小化地调整labels。
2. 优先使用counts和action tables而不是decorative charts。
3. 在请求时明确过滤alert noise（例如ArgoCD/GitOps）。

### 4) 在Apply之前验证

编辑后始终运行scan script：

```bash
skills/grafana-platform-dashboard/scripts/promql_scan_thanos.sh \
  --context <ctx> \
  --dashboard-json <file.json> \
  --output <scan.tsv>
```

通过标准：所有queries报告 `success`，zero bad/parse errors。

### 5) Apply并验证Sync

仅在验证成功后apply：

```bash
skills/grafana-platform-dashboard/scripts/grafanadashboard_roundtrip.sh apply ...
oc --context <ctx> -n <ns> get grafanadashboard <name> \
  -o jsonpath='{.status.conditions[?(@.type=="DashboardSynchronized")].status}{"|"}{.status.conditions[?(@.type=="DashboardSynchronized")].reason}{"\n"}'
```

### 6) 以Operator-Focused Summary结束

报告：

1. 更改了什么（panel names和intent）。
2. 验证结果（query count和failures）。
3. Sync status和任何residual risk。
4. 下一步：将live changes提升为GitOps-managed source。

## Design Rules

1. 将critical tenant-impact predictors放在首位。
2. 每个red panel必须暗示一个action path。
3. 避免ambiguous的panel names（例如将 "platform pods" 替换为具体的namespace scope）。
4. 保持L1 low-noise；将detail移到下方或专门的dashboards。
5. 将GPU deep diagnostics保留在专门的GPU dashboard中，不要混入L1。

## References

1. [Platform Contract](references/platform-contract.md)
2. [PromQL Panel Library](references/promql-library.md)
3. [Layout Guidelines](references/layout-guidelines.md)

## Local Resources

### references/

- [references/layout-guidelines.md](references/layout-guidelines.md)
- [references/platform-contract.md](references/platform-contract.md)
- [references/promql-library.md](references/promql-library.md)

### scripts/

- `scripts/grafanadashboard_roundtrip.sh`
- `scripts/promql_scan_thanos.sh`
- `scripts/validate.sh`
