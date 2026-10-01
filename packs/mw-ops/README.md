# mw-ops：中间件诊断包（社区版）

缓存（Redis）与消息队列（Kafka/RabbitMQ）域的双专家 + 四技能 + 一链。全部方法论
面向**只读采集 + 诊断判读**，变更动作一律走平台审批面。

| 资产 | 说明 |
|---|---|
| `skills/redis-triage` | 五类分诊：内存满/驱逐、延迟毛刺、雪崩击穿穿透、主从中断、连接打满（oo-devops redis-ops/redis-expert/redis-inspect 三技能合并重写） |
| `skills/redis-hotkey-bigkey` | 热Key/大Key 定位三段式：全量抽样→定点体检→行为交叉（原创） |
| `skills/kafka-triage` | 消费积压三因子、分区 ISR、存储水位、消息丢失/重复成因链（oo-devops kafka-ops 重写） |
| `skills/rabbitmq-triage` | 队列堆积三态、连接限流、水位告警、脑裂一致性（oo-devops rabbitmq-ops 重写） |
| `cache-analyst/` | 缓存诊断专家（ask-ops 只读面） |
| `mq-analyst/` | 消息队列分析专家（ask-ops 只读面） |
| `chain.yaml` | `workflow/mw-quick-audit` 中间件例检链（缓存→队列两步，技能钉扎） |

## 与其他包的错位

- 主机/OS 层（CPU/内存/IO/网络）归 **os-basics**；K8s 集群层归 **k8s-ops**——中间件
  实例跑在 K8s 上时，先经 k8s-ops 排除工作负载层再进本包；
- 本包不碰数据库（**db-ops**）、交付链（**cicd-ops**）与监控栈自诊（**obs-ops**）。

## 工具面

无专用 MCP 依赖：redis-cli / kafka-*.sh / rabbitmqctl 等只读 CLI 经平台 `ask-ops`
采集面在目标运维主机执行，凭证走主机侧已配置客户端，对话不回显。

## 安装

```bash
go run ./cmd/bq-markettool install --url <本仓> --pack mw-ops --api <扁鹊实例>
```

## 改造说明

源自 openocta 收割的 oo-devops 市场包（31 员工/189 技能整包，构建产物禁手改）。
本包取中间件域高价值技能收割重写：剥离原版环境变量注入与未供给的 MCP 假设，
统一扁鹊契约（四段式技能、输出协议、审批纪律）。zookeeper/memcached/consul 等
薄壳资产暂未搬运，按需在后续版本承接。
