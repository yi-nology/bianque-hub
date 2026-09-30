# chain-starter：最小示例包

一个专家 + 一项技能 + 一条两步诊断链——**贡献者以此为模板起步**。本包刻意保持
最小：没有工具面、没有凭证依赖、链只走两步。复制本包目录改造成你的领域包即可。

## 逐文件拆解

| 文件 | 作用 | 你要改什么 |
|---|---|---|
| `pack.yaml` | 包清单（`api_version` 契约门 + `provides` 资产登记） | 包名/版本/描述；provides 与实际资产对齐 |
| `provenance.json` | 来源标记（装进扁鹊后显示为市场包） | `pack` 字段改成你的包名 |
| `starter-analyst/agent.yaml` | 专家声明（slug/kind/路由词/技能挂载） | slug 用 `imports/` 前缀防撞平台命名空间；路由词避开内置包域词 |
| `starter-analyst/prompt.md` | 专家系统提示（kind=llm 必填） | 换成你的领域职责与输出协议 |
| `skills/starter-method/SKILL.md` | 技能（开放 SKILL.md 格式） | frontmatter 契约 + 正文四段式（触发条件/数据来源/方法论/输出要求） |
| `chain.yaml` | 诊断链（线性步骤 + 技能钉扎演示） | slug 必须 `workflow/` 前缀；步骤只能引用叶子专家；priority 禁 P0 |

## 三条硬规则（CI 会拦）

1. `pack.yaml` 的 `api_version` 必须等于扁鹊平台契约版本（当前 1）；
2. `chain.yaml` 的 slug 必须 `workflow/` 前缀、跨仓唯一，步骤禁引用 `workflow/`
   前缀专家（治理环节由平台自动衔接，包侧不用管）；
3. 技能 `SKILL.md` frontmatter 为合法 YAML，`name` 全局唯一、与目录名一致。

## 本地自检

```bash
go run ./validator ./packs/chain-starter   # 结构契约校验（CI 同款）
```

装入扁鹊实测（在扁鹊仓库侧）：

```bash
go run ./cmd/bq-markettool install \
  --url https://github.com/yi-nology/bianque-hub \
  --pack chain-starter \
  --api http://127.0.0.1:8900
```

## CHANGELOG

见 [CHANGELOG.md](CHANGELOG.md)。
