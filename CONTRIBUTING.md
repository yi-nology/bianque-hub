# 贡献指南

欢迎为扁鹊贡献**技能**（方法论）、**专家**（领域角色）与**诊断链**（编排流水线）。
本仓只收**内容资产**；平台能力（调度引擎、审批治理、合议、报告协议等）归扁鹊
平台仓，不在此收。

## 快速起步：复制 chain-starter

[`packs/chain-starter/`](packs/chain-starter/) 是逐文件拆解的最小模板：

1. 复制整个目录为 `packs/<你的包名>/`（小写中划线）；
2. 改 `pack.yaml`（包名/版本/描述/provides）、`provenance.json`（pack 字段）；
3. 换掉专家（slug 用 `imports/` 前缀）与技能（四段式正文）；
4. 需要编排就写 `chain.yaml`；
5. 本地校验全绿再提 PR：

```bash
go run ./validator ./packs
```

## 契约速查

### pack.yaml（包清单）

```yaml
api_version: 1              # 契约版本门（与平台 PackAPIVersion 一致，勿动）
name: your-pack             # 必须与目录名一致
version: 1.0.0              # SemVer
description: 一句话人话描述
provides:
  experts: [imports/your-expert]   # 与包内实际资产对齐（CI 交叉核对）
  skills: [your-skill]
changelog: CHANGELOG.md
```

### agent.yaml（专家）

```yaml
slug: imports/your-expert   # imports/ 前缀防撞平台命名空间（workflow/specialists/platform/ 为平台域）
name: 人话名
layer: expert               # primary|flow|expert|support|governance（包内一般用 expert）
kind: llm                   # engine|llm|stub（engine 仅平台已实现 slug，包侧勿新增）
prompt_file: prompt.md      # kind=llm 必填（专家目录内）
skills: [your-skill]        # 挂载技能（全局唯一名，可跨包引用）
route_priority: P4          # P0-P6；P0 红线是专家域，包侧从 P2 起步
route_group: tooling        # 平台 route_groups 白名单：system/storage/network/middleware/infra/security/offline/tooling/knowledge/status_query/monitoring/kubernetes
route_keywords: [你的词]     # 路由词纪律见下
route_desc: 一句话路由描述（LLM 语义面）
```

### SKILL.md（技能，开放格式）

```markdown
---
name: your-skill            # 与目录名一致，全局唯一
description: 一句话描述（触发面靠它）
mode: on_demand             # static（全文注入系统提示）| on_demand（渐进披露，推荐）
version: 1.0.0              # SemVer
maturity: experimental      # experimental|stable|frozen|deprecated
requires_mcp:               # 可选：声明的工具面
  - server: some-mcp
    tools: [tool-a, tool-b]
---
## 触发条件
## 数据来源        （只允许使用已声明工具）
## 方法论          （固定顺序步骤）
## 输出要求
```

### 工具面约定（ask-ops 两层）

- **固定清单面**（十工具：inventory 四件 + 性能快照/日志/进程/端口/目录体积/探测）：
  适合主机层事实问答。声明方式 `tools: [{server: ask-ops}]`（allow 留空=全部）；
  问答型 support 角色请像内置 `platform/ask-ops` 一样显式钉 allow，不领受命令面。
- **受审只读命令面**（`run_readonly_command`）：领域专家跑域 CLI 的通道——白名单
  受审（redis 只读命令族 / kafka-\*.sh --describe / mysql -e 'SELECT…' / curl GET /
  docker+kubectl 只读子命令等），引号外控制字符、写形态旗标、白名单外命令一律
  拒绝；stdout-only、尾部截断 400 行/32KB、退出码在 exit_code。
- 技能处方里的命令必须与该守卫兼容：URL 带 `&` 等特殊字符要写成引号形态；
  密码类凭证走主机侧客户端配置（裸 `-p`/`-W` 会挂起被拒）；sudo 仅支持
  `-u <user>` 目标切换；变更动作永远写进 `recommendation.steps` 走审批，不处方化。

### chain.yaml（诊断链，每包最多一条）

```yaml
slug: workflow/your-chain   # 必须 workflow/ 前缀且全局唯一
name: 人话名
route_priority: P3          # 必填，禁 P0
route_group: system
route_keywords: [链专属词]
steps:
  - agent: imports/your-expert
    instruction: "指令模板，{{input}} 占位用户输入"
    # skill: your-skill     # 可选：技能钉扎（该步强制注入技能全文）
```

治理环节（方案→风险评估→审批→执行→验证）与结论合议由平台自动衔接，
**包侧不要写** approval/execute 类治理节点。

## 硬规则（CI 拦截，PR 前自检）

1. `api_version` 必须 = 平台契约版本（当前 1）；
2. 专家 slug / 技能名 / 链 slug 全局唯一；包名与目录名一致；
3. 链步骤只能引用**叶子专家**（禁 engine、禁 `workflow/` 前缀目标）；
4. `route_group` 必须在平台白名单；`kind=llm` 必有 `prompt_file`；
5. **脱敏**：内网域名/内部工具名/密钥形态/私网 IP 一律拦截；
6. **路由词纪律**：新包路由词避开已占用的域词——os-basics（巡检/体检/调优/瓶颈/
   安全/等保及六域症状词）、k8s-ops（工作负载诊断/pod诊断/helm排查等）、mw-ops
   （缓存诊断/消息积压/主从延迟等）、db-ops（慢查询/锁等待等）、cicd-ops（构建失败/
   流水线排障等）、obs-ops（监控失明/采集断点/面板无数据等）；平台查询面词
   （prometheus/监控指标/指标查询）归内置监控入口。错位竞争会让两个入口互相劫持
   ——一律用复合限定词。CI 已机械拦截**跨包同优先级的 route_keywords/symptoms
   重复**（真歧义，平台装载会失败）；与平台内置域词的冲突 CI 看不到，仍靠本条
   人工把关；
7. 版本纪律：任何内容变更都升 `version`（SemVer：修文案 patch、加条目 minor、
   破坏契约 major）并在包内 CHANGELOG.md 记一行。

## review 约定

- 技能正文质量看三件事：触发条件是否可判定、方法论是否可复现（步骤固定顺序）、
  输出要求是否可校验；
- 专家 prompt 不得绕过平台输出协议（`requires_approval`/`decision` 语义不许自创）；
- 引用平台 `_shared/` 共享段（如 platform-matrix）合法；包内私有段放专家目录里。
