# os-basics：主机/OS + K8s 集群诊断包（社区版）

两域合一：**主机/OS**（巡检、性能、内存、IO、网络、安全）+ **Kubernetes 集群诊断**
（k8s-health 专家 + k8sgpt 分析桥）。

本包是扁鹊（bianque）内置同名核心包的**社区上游版**：slug/路由/技能与内置版完全
等位，安装后原位接管内置包（upgrade 接管，uninstall 后重启自动还原内置版）。**本仓
是唯一编辑点**——扁鹊仓库的 `internal/agents/seed/os-basics` 只是本包的同步快照
（维护者经 `bq-markettool seed-sync` 单向同步，直接改 seed 会在下次同步被覆盖），
欢迎 PR 更新专家方法论、技能与诊断链。

## 与其他智能体的兼容性

- `skills/*/SKILL.md` 遵循开放技能格式（`name`+`description` 为标准字段；
  `mode`/`version`/`maturity`/`requires_mcp` 为扁鹊扩展键，其他运行时按需忽略）——
  任何支持 Agent Skills 的智能体可直接挂载 `skills/` 目录作参考方法论使用。
- 专家（`*/agent.yaml`+`prompt.md`）与 `chain.yaml` 为扁鹊包契约格式，面向扁鹊
  平台导入；`prompt.md` 的输出协议依赖扁鹊报告 schema，其他运行时可参考但不可直搬。
- 平台速查段（`_shared/prompts/platform-matrix.md`，中文 Linux 发行版识别矩阵）
  按契约由扁鹊平台共享层分发；非扁鹊读者可在仓内 [knowledge/](../../knowledge/) 取参考副本。
- 技能声明的 `requires_mcp`（如 k8sgpt、security-assistant 工具面）需要扁鹊的
  MCP 工具底座才有完整自动化能力；无工具面时技能退化为人工核查方法论。

## 主机/OS 域

- 专家：workflow/system-patrol（巡检引擎 P1）、workflow/perf-tuning、
  specialists/{memory,io,network}-analysis、specialists/security-assistant
- 技能：安全域 4 项（ssh 爆破 / root 登录 / sudo / 进程）+ k8s 节点诊断 1 项
- analyzers.yaml：域 slug → sysprobe 检查器组映射；k8s 域 source: k8sgpt
- chain.yaml（1.3.0 新增）：workflow/host-quick-audit 主机快查链（安全基线 +
  内存进程事实两步，与全域巡检错位）

## K8s 集群诊断域

> k8sgpt 0.4.39 实测：**MCP 走 `--mcp-http` 的 streamable HTTP**（stdio 形态在
> 同版本会被 serve 基座的 metrics/api 端口绑定拖死），扁鹊侧以 **url 模式**挂载。
> 本包白名单按 0.4.39 实弹工具名终稿。

### 1. 部署 k8sgpt sidecar（实例级前置）

```bash
brew install k8sgpt            # 或官方二进制
# MCP serve 不需要真 AI 后端（平台侧 LLM 负责 explain）——挂哑后端过启动检查：
k8sgpt auth add --backend localai --model dummy --baseurl http://127.0.0.1:9
k8sgpt auth default -p localai
# sidecar 常驻（launchd/systemd/nohup 均可）；--kubeconfig 指向集群凭证：
k8sgpt serve --mcp -b localai --mcp-http \
  --mcp-port 18089 -p 18080 --metrics-port 18081 \
  --kubeconfig /path/to/kubeconfig
# 就绪探针：POST http://127.0.0.1:18089/mcp（streamable HTTP MCP）
```

端口注意：serve 基座**必然绑定** metrics/api 端口——多实例部署用不同
`-p/--metrics-port/--mcp-port` 区分，避免互踩。

### 2. 挂载扁鹊（etc/config-*.yaml → mcp.servers）

```yaml
mcp:
  servers:
    k8sgpt:
      enabled: true
      url: "http://127.0.0.1:18089/mcp"   # streamable HTTP MCP（无 command）
      timeout: 2m
```

### 3. 版本与风险

- 工具名以 k8sgpt 0.4.39 实测为准：`analyze / cluster-info / get-logs /
  get-resource / list-events / list-namespaces / list-resources / list-filters /
  add-filters / remove-filters / list-integrations / config`（12 个；包白名单
  只放诊断面子集）。
- k8sgpt 迭代快：analyze 输出 schema 漂移时，平台侧映射器单测会红——升级 sidecar
  前先确认工具名与输出结构未变。
