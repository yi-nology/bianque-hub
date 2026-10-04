# 专家级子集安装（派生迷你包）设计

- 日期：2026-10-05
- 状态：已评审通过（会话内设计门禁），待拆实施计划
- 涉及仓：bianque-hub（本仓，内容与校验）、bianque（平台仓，安装器）
- 路线决策：B「子集安装」——包不拆，安装时按闭包现场合成迷你包（备选 A 拆包 / C 登记面学专家粒度已否决）

## 1. 背景与目标

**现状**：扁鹊的安装单元是整包。`bq-markettool install --pack <名>` 与插件 API
`/api/v1/plugins/install` 把整个包目录装进 `experts/`，`bq_plugin` 登记面按包记
升级/卸载/启停。专家只是包内含 `agent.yaml` 的子目录（保留目录 `skills/ mcp/
changes/ credentials/` 除外），装载期自动识别。平台已有**专家级启停**
（`bq_expert_toggle`：Disabled + SkillsOverride），但那只是已装文件的开关。

**问题**：用户无法只装某一个专家。全仓 16 专家中 os-basics 独占 7 个，
"只要一个专家却拖进整包"的痛点集中在该包。

**目标**：安装粒度下沉到专家——`--expert` 指定专家，安装器按闭包合成迷你包，
走既有插件管线原子安装；包在 hub 侧仍是唯一编辑/版本单元，结构零变动。

**非目标**：

- 不改 `plugin_svc` 装载/登记模型，不动 DB schema。
- 不拆包、不建专家级 seed（seed 仍是 os-basics 整包快照；专家级安装是市场面行为）。
- 不做链级安装：`chain.yaml` 是包级多专家资产，刻意不进迷你包。
- 控制台 zip 上传路径不做专家选择（zip 装的仍是整包；本设计只覆盖 bq-markettool）。

## 2. 闭包规则（每专家）

从源包 `P` 拷贝进迷你包的资产（`<专家>` 取目录名 basename）：

| 资产 | 取法 |
|---|---|
| 专家目录 | `P/<专家>/` 全量（agent.yaml、prompt.md、ui.yaml，含后续新增的同目录可选资产） |
| 技能 | `agent.yaml skills:` 列表 → `P/skills/<名>/`（多数 os-basics 专家无技能，闭包自然为空） |
| 变更块 | 所携技能 SKILL.md frontmatter `provides_changes` 引用的 slug → `P/changes/<slug>.yaml` |
| MCP 契约 | `agent.yaml tools[].server` 在 `P/mcp/` 有同名 `<server>.yaml` → 拷贝；宿主侧 server（ask-ops 等，包内无契约文件）跳过 |
| analyzers | `P/analyzers.yaml` 中 `slug` 等于该专家 slug 的条目 → 合成单条目 `analyzers.yaml`；无条目则不带该文件 |
| credentials | `P/credentials/` 整目录（若有；当前无包使用，前向兼容） |

**刻意不带**：

- `chain.yaml`：包级资产，步骤引用可能缺位的专家（含跨包
  `specialists/<slug>`），带上会断 reload；单专家安装语义下链无意义。
- `disambiguation.yaml`：跨专家路由仲裁，单专家安装无仲裁对象。
- `README.md` / `CHANGELOG.md`：包级文档。

**analyzers 语义核对**（已查实）：装载期各包 `analyzers.yaml` 跨包归并、组键
跨包唯一；缺文件=零映射优雅降级。单专家迷你包携带且仅携带自己的条目，
既是该包视角的全覆盖，又不会与同源其他迷你包撞组键。

## 3. 合成 manifest 与血缘

- `pack.yaml`：
  - `name: <包>.<专家目录名>`（如 `os-basics.system-patrol`；过平台
    `pluginNameRe` `^[a-z0-9][a-z0-9._-]{0,127}$`，专家目录名已全小写）。
  - `api_version`、`version` 承源包（升级=重跑同命令重派生）。
  - `provides` 按实际携带重写（experts=[slug]，skills=所携技能）。
  - `description` 自动生成：`自 <包> v<版本> 子集安装（仅专家 <slug>）`。
- `provenance.json`：重写，`base_url` 承源包，追加
  `derived_from: {pack, version, expert}`。platform 判 `source=market`（可卸载）
  靠该文件存在，顺便留下血缘供排查。

## 4. CLI 口径与生命周期

```
go run ./cmd/bq-markettool install \
  --url https://github.com/yi-nology/bianque-hub \
  --pack k8s-ops --expert k8s-workload-analyst \
  --api http://127.0.0.1:8900
```

- `--expert` 可重复；每个专家各自合成迷你包、各自安装为独立插件
  （独立升级/卸载/启停生命周期——与"每个专家独立安装"的目标语义一致）。
  `--expert` 必须与 `--pack` 同用，且所列专家须同属该包（`--pack` 单值）。
- 不带 `--expert` 时行为完全不变（整包安装），向后兼容。
- `--dest` 落盘模式同样支持：先派生再落盘，语义与 API 模式一致。
- 升级：重跑同一命令，同名插件走既有 upgrade 路径
  （fingerprint=合成 pack.yaml sha256，照旧）。
- 卸载：删迷你包；其上 `bq_expert_toggle` 残留由既有 reload 漂移清理收口。

## 5. 冲突面

- **整包与迷你包互斥**：同 slug 并存（如 `k8s-ops` 与 `k8s-ops.k8s-workload-analyst`）
  会在 reloadStrict 撞 slug 失败。安装器加预检（复用 `apiProbePack` 探登记面）：
  源包已在位则拒绝安装并提示——"整包已在位：若只需某专家，用专家级启停
  （toggle）；或先卸载整包再装专家级"。
- **同源多迷你包并存**：允许（slug 各异）。若两者路由关键词同优先级相撞，
  reload 原子回滚（现状语义），安装器把 reload 错误透传给用户。
- **MCP 契约清单跨包同名**：契约清单是覆盖合并语义（`_shared/mcp/` 基线 +
  各包 `mcp/`），同源迷你包携带相同内容文件属良性；激活仍归 conf 面，不在此层。

## 6. 改动落点

### 平台仓（bianque）

- `cmd/bq-markettool/install.go`：
  - flag 新增 `--expert`（StringSlice，须与 `--pack` 同用）。
  - 新增 `deriveExpertPack(packDir, expert, srcURL) (string, error)`：按 §2 闭包
    合成到临时目录（SKILL.md frontmatter 解析出 provides_changes；不含则空）。
  - 合成产物接既有 `apiInstall` / `destInstall`；临时目录清理。
  - 预检：`--expert` 时先探源包是否在位（§5）。
  - `install_test.go` 补用例（§7）。
- `plugin_svc` 不动；文档（platform-matrix 等）同步一句专家级安装口径。

### hub 仓（本仓）

- `validator`：新增"专家可独立安装性"检查层——每专家 `agent.yaml` 的 `skills:`
  引用闭包存在、`tools[].server` 契约存在或属宿主面（复用既有工具面覆盖口径，
  补专家视角缺口）；违规 error。
- `README.md`「装进扁鹊」补 `--expert` 示例与独立安装口径（整包互斥、链不随行）。
- `CONTRIBUTING.md`：一句"专家默认须可独立安装（闭包自洽），校验器把关"。

### seed

不涉及（seed 只含 os-basics 整包快照；专家级安装是市场面行为，无需 seed-sync）。

## 7. 测试计划

`install_test.go` 闭包单测基准：

1. `k8s-ops/k8s-workload-analyst`：1 专家 + 4 技能 + 其 provides_changes 引用的
   变更块（非全量 changes/）+ mcp/k8sgpt.yaml + 合成 pack.yaml/provenance；
   不含 chain.yaml。
2. `os-basics/security-assistant`：4 技能 + analyzers 单条目（slug
   specialists/security-assistant）+ 0 mcp（宿主侧 security-assistant server）。
3. `os-basics/system-patrol`：0 技能 + analyzers 单条目。

负例：专家名不存在；`--expert` 无 `--pack`；源包已在位的预检拒绝。

端到端：8900 实装——装 `k8s-ops.k8s-workload-analyst`，插件页可见、路由可达、
卸载回净；再验整包在位时预检拒绝路径。

## 8. 兼容与回滚

- 不带 `--expert` 的既有命令行为逐字节不变。
- 迷你包卸载即回净（`_trash` 可手工找回）；安装失败由既有原子回滚兜底
  （reload 拒收则目录撤出，旧视图持续服务）。
- hub 侧无内容变更，包版本不升；validator 新检查层对现存 16 专家应全绿
  （若曝出真缺口，属包内容修复，随本设计一并收口）。
