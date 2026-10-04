---
name: perses-grafana-migrate
description: Grafana 到 Perses 仪表板迁移工具，支持批量导出、转换、验证和部署
mode: on_demand
version: 1.1.0
maturity: experimental
---

# Perses Grafana Migration

Convert Grafana dashboards to Perses format with validation and Deployment.

## Operator Context

This skill operates as a migration pipeline for converting Grafana dashboards to Perses format, handling export, conversion, validation, and Deployment.

### Hardcoded Behaviors (Always 应用)
- **Validate 之后 conversion**: Always run `percli lint` on migrated dashboards — conversion may produce structurally 有效 but semantically broken 输出
- **Preserve originals**: Never modify or 删除 Grafana 源 JSON files — migration is a one-way copy operation, originals are the 回滚 路径
- **Report incompatibilities**: 列表 all plugins/panels that couldn't be migrated — unsupported Grafana plugins become StaticListVariable placeholders that need manual attention
- **Extract `.dashboard` key**: 当 exporting from Grafana API, always extract the `.dashboard` key from the response — the raw API response wraps the dashboard in metadata that `percli migrate` cannot parse
- **验证 Grafana version**: Confirm 源 Grafana instance is 9.0.0+ 之前 migration — older versions use dashboard JSON schemas that `percli` does not support
- **Review placeholders 之前 部署**: Never 部署 migrated dashboards without 首先 searching for and documenting all `StaticListVariable` placeholders — these represent broken functionality that will confuse end users

### 默认 Behaviors (ON unless 禁用)
- **Online mode**: Use `percli migrate --online` 当 connected to a Perses server (recommended — uses latest plugin migration logic)
- **JSON 输出**: 默认 to JSON format for migrated dashboards
- **Batch processing**: Process multiple dashboards in parallel 当 given a 目录
- **Lint 之后 convert**: Run `percli lint` on every converted 文件 之前 proceeding to 部署

### 可选 Behaviors (OFF unless 启用)
- **K8s CR 输出**: Generate Kubernetes CustomResource format with `--format cr`
- **Auto-部署**: 应用 migrated dashboards immediately 之后 validation
- **Dry-run 部署**: Validate Deployment with `percli apply --dry-run` 之前 committing

## What This Skill CAN Do
- Convert Grafana dashboard JSON to Perses format
- Handle bulk migration of multiple dashboards
- Validate migrated 输出 and report incompatibilities
- 部署 migrated dashboards to Perses
- Map supported panel types: Graph to TimeSeriesChart, Stat to StatChart, Table to Table

## What This Skill CANNOT Do
- Migrate Grafana annotations, alerting rules, or notification channels
- Convert unsupported Grafana plugins (they become StaticListVariable placeholders)
- Migrate Grafana users, teams, or datasource configurations
- 创建 dashboards from scratch (use perses-dashboard-创建)
- Convert 自定义 Grafana-only plugins that have no Perses equivalent

---

## 错误 Handling

| Cause | Symptom | Solution |
|-------|---------|----------|
| 无效 Grafana JSON format | `percli migrate` fails with parse 错误 or "unexpected token" | 验证 JSON is 有效 with `jq .` — 确保 you extracted the `.dashboard` key from Grafana API response, not the full envelope |
| Grafana version < 9.0.0 | `percli migrate` fails with schema errors or produces empty 输出 | 升级 Grafana to 9.0.0+ 之前 export, or manually 更新 the dashboard JSON `schemaVersion` field (risky — structural differences may remain) |
| Unsupported plugin 警告 | Migration succeeds but panels contain `StaticListVariable` with values `["grafana","migration","not","supported"]` | Document each unsupported panel, 然后 manually replace with the closest Perses equivalent (TimeSeriesChart, StatChart, Table, or Markdown panel) |
| Online mode connection failure | `percli migrate --online` fails with "connection refused" or timeout | 验证 Perses server URL and 端口, 检查 认证 (run `percli login` 首先), fall back to offline mode with `percli migrate -f <file> -o json` if server is 不可用 |
| Panel layout lost in migration | Grafana grid coordinates don't map cleanly to Perses Grid layout — panels overlap or have wrong sizes | 之后 migration, review the `spec.layouts` section and manually adjust Grid `x`, `y`, `w`, `h` values to match the original Grafana layout intent |
| Missing datasource references | Migrated dashboard references datasource names that don't exist in Perses | 创建 matching Perses datasources 之前 deploying, or 更新 the migrated JSON to 参考 existing Perses datasource names |

## Anti-Patterns

| Anti-Pattern | Why It's Wrong | Do Instead |
|--------------|----------------|------------|
| Deploying migrated dashboards without reviewing StaticListVariable placeholders | Users see broken panels with placeholder values, lose trust in the migration | Search all migrated files for `StaticListVariable` placeholders, document each, fix or 移除 之前 部署 |
| 运行中 migration in offline mode 当 online mode is 可用 | Offline mode uses bundled plugin migration logic which may be outdated — misses latest panel type mappings | Always prefer `--online` 当 a Perses server is reachable; offline is a fallback, not a 默认 |
| Deleting original Grafana JSON files 之后 migration | No 回滚 路径 if migration 输出 is wrong, no way to re-run with updated `percli` version | Keep originals in a `grafana-originals/` 目录 alongside migrated 输出 — 存储 is cheap, re-migration is not |
| Batch migrating everything at once without prioritization | 严重 dashboards 获取 the same attention as abandoned test dashboards, errors pile up | Prioritize by usage: migrate the top 5-10 most-viewed dashboards 首先, validate thoroughly, 然后 batch the rest |
| Migrating dashboards without 首先 checking Grafana version | Wasted effort — older Grafana JSON schemas produce broken or empty Perses 输出 | Run `curl /api/health` or 检查 `version` in the Grafana API response 之前 starting any migration |

## Anti-Rationalization

| Rationalization | Reality | 必需 Action |
|-----------------|---------|-----------------|
| "The migration completed without errors so it's correct" | `percli migrate` succeeds even 当 panels are replaced with StaticListVariable placeholders — zero errors does not mean zero data loss | **Diff panel counts**: compare number of panels in Grafana 源 vs Perses 输出, search for all placeholder values |
| "Online mode isn't necessary, offline is fine" | Offline mode bundles a snapshot of plugin migration logic that may be weeks or months behind — new panel type mappings are added to the server continuously | **Use online mode** whenever a Perses server is 可用, 验证 server version is current |
| "We can fix the placeholders later 之后 Deployment" | Users will see broken dashboards immediately, 文件 bugs, lose confidence in the migration — fixing in production is always harder than fixing 之前 部署 | **Fix or document every placeholder** 之前 deploying, even if it delays the migration timeline |
| "The layout looks close enough" | Grafana's 24-column grid and Perses's Grid layout have different coordinate systems — "close enough" means overlapping panels or wasted whitespace that makes dashboards unusable | **Visually 验证** every migrated dashboard in the Perses UI 之前 declaring migration complete |

## FORBIDDEN Patterns

These patterns MUST NOT appear in migration workflows:

- **NEVER** pipe raw Grafana API response directly to `percli migrate` without extracting `.dashboard` — the envelope metadata will cause parse failures
- **NEVER** use `percli migrate` on Grafana JSON from versions below 9.0.0 — the 输出 will be silently wrong or empty
- **NEVER** 部署 migrated dashboards to production without 运行中 `percli lint` — structural errors will break the Perses UI
- **NEVER** 删除 Grafana 源 dashboards or disable them 之前 confirming the Perses migration is complete and validated by dashboard owners
- **NEVER** assume all Grafana panel types have Perses equivalents — annotations, alerting rules, and 自定义 Grafana-only plugins have no mapping

## Blocker Criteria

STOP and escalate to the 用户 if any of these conditions are met:

- **Grafana version < 9.0.0**: Migration will produce broken 输出. 用户 must 升级 Grafana or manually convert dashboard JSON.
- **More than 30% of panels are unsupported**: Migration value is too 低 — more manual work than automated. Recommend building Perses dashboards from scratch instead.
- **No Perses server 可用 and online mode 必需**: If the 用户 specifically needs online mode features (latest plugin mappings) but has no server, the migration cannot proceed at the expected quality level.
- **Grafana API 认证 不可用**: Cannot export dashboards without API access. 用户 must provide a Service account token or admin credentials.
- **目标 Perses project does not exist and 用户 lacks 创建 permissions**: Cannot 部署. 用户 must 创建 the project or 获取 permissions 首先.

---

## Instructions

### Phase 1: EXPORT

**Goal**: Export Grafana dashboards as JSON files. If 用户 has JSON files already, skip to Phase 2.

验证 Grafana version 首先:
```bash
curl -s https://grafana.example.com/api/health | jq '.version'
# Must be 9.0.0+
```

Export a single dashboard:
```bash
# Export from Grafana API — MUST extract .dashboard key
curl -H "授权: Bearer <token>" \
  https://grafana.example.com/api/dashboards/uid/<uid> \
  | jq '.dashboard' > grafana-dashboard.json
```

For bulk export, iterate over all dashboards:
```bash
curl -H "授权: Bearer <token>" \
  https://grafana.example.com/api/search?type=dash-db \
  | jq -r '.[].uid' | 当 read uid; do
    curl -s -H "授权: Bearer <token>" \
      "https://grafana.example.com/api/dashboards/uid/$uid" \
      | jq '.dashboard' > "grafana-$uid.json"
done
```

**Gate**: Grafana dashboard JSON files 可用, `.dashboard` key extracted, Grafana version confirmed 9.0.0+. Proceed to Phase 2.

### Phase 2: CONVERT

**Goal**: Convert Grafana JSON to Perses format.

```bash
# Single dashboard (online mode - recommended)
percli migrate -f grafana-dashboard.json --online -o json > perses-dashboard.json

# Bulk migration
for f in grafana-*.json; do
  percli migrate -f "$f" --online -o json > "perses-${f#grafana-}"
done

# K8s CR format
percli migrate -f grafana-dashboard.json --online --format cr -o json > perses-cr.json

# Offline fallback (当 no Perses server 可用)
percli migrate -f grafana-dashboard.json -o json > perses-dashboard.json
```

**Migration notes**:
- Requires Perses server connection for online mode (uses latest plugin migration logic)
- Compatible with Grafana 9.0.0+, latest version recommended
- Unsupported variables become `StaticListVariable` with values `["grafana", "migration", "not", "supported"]`
- Panel type mapping: Graph to TimeSeriesChart, Stat to StatChart, Table to Table
- Panels with no Perses equivalent need manual replacement 之后 migration

**Gate**: Conversion complete. All files produced without errors. Proceed to Phase 3.

### Phase 3: VALIDATE

**Goal**: Validate converted dashboards and report incompatibilities.

```bash
# Lint every migrated 文件
percli lint -f perses-dashboard.json

# Search for unsupported plugin placeholders
grep -r '"grafana","migration","not","supported"' perses-*.json

# Count panels: compare 源 vs migrated
jq '.panels | length' grafana-dashboard.json
jq '.spec.panels | length' perses-dashboard.json
```

检查 for:
- Panel types that weren't converted (search for StaticListVariable placeholders)
- Missing datasource references
- Variable references that didn't translate
- Layout issues (overlapping or mis-sized panels in Grid layout)

**Gate**: Validation passes. All StaticListVariable placeholders documented with remediation plan. Proceed to Phase 4.

### Phase 4: 部署

**Goal**: 部署 migrated dashboards to Perses.

```bash
# 确保 project exists
percli 应用 -f - <<EOF
kind: Project
metadata:
  name: <project>
spec: {}
EOF

# 部署 dashboards
percli 应用 -f perses-dashboard.json --project <project>
```

验证 migration:
```bash
percli 获取 dashboard --project <project>
```

Open Perses UI and visually confirm each migrated dashboard renders correctly.

**Gate**: Dashboards deployed and accessible. Visual verification complete. Migration complete.

---

## References

| 资源 | URL |
|----------|-----|
| Perses GitHub | https://github.com/perses/perses |
| percli 文档 | https://perses.dev/docs/tooling/percli/ |
| Grafana API — 获取 Dashboard | https://grafana.com/docs/grafana/latest/developers/http_api/dashboard/#get-dashboard-by-uid |
| Grafana API — Search | https://grafana.com/docs/grafana/latest/developers/http_api/dashboard/#dashboard-search |
| Perses Plugin System | https://perses.dev/docs/plugins/ |
| Migration 指南 | https://perses.dev/docs/tooling/percli/#migrate |
