---
name: n8n-instance-triage
description: n8n 实例本体分诊方法论：健康端点两级语义（/healthz vs /healthz/readiness）、起不来与反复重启、资源占用（CPU/内存）与 task runner 隔离、日志面定位——固定顺序定位与判读基准（ask-ops 只读采集面）。
mode: on_demand
version: 0.1.0
maturity: experimental
requires_mcp:            # 采集依赖的工具面（装载期对账依据；旧 bianque-tools 二进制缺这些工具时应先重建）
  - server: ask-ops
    tools: [run_readonly_command, run_readonly_commands]
---

## 触发条件

- 症状关键词：n8n 打不开、实例起不来、反复重启、n8n 卡、页面 502、CPU 100%、内存涨、日志在哪看、升级后起不来
- 组合场景：容器/调度层异常（ImagePullBackOff、节点驱逐）先走 k8s-ops；DB 本体故障先走 db-ops——本技能管「n8n 进程视角的实例健康」

## 数据来源（ask-ops 只读面）

- 健康端点：`curl -s http://127.0.0.1:5678/healthz`（可达性，**不反映 DB 状态**）、
  `curl -s http://127.0.0.1:5678/healthz/readiness`（200 仅当 **DB 已连接且 schema 迁移完成**）；
  路径可用 `N8N_ENDPOINT_HEALTH` 改（默认 healthz）；
- 容器/进程面：`docker ps`（状态与重启计数）、`docker inspect <容器>`（Env 凭证字段引用前打码）、
  `docker logs --tail 200 <容器>`；K8s 形态 `kubectl -n <ns> get pods` / `kubectl -n <ns> logs deploy/<n8n> --tail=200`；
- 资源面：`docker stats --no-stream <容器>`、`.n8n` 数据目录 du/df（磁盘水位）；
- worker 进程没有健康端点（默认关闭），需 `QUEUE_HEALTH_CHECK_ACTIVE=true` 才有 `/healthz`——
  采集不到不臆测，转配置面核对（见队列模式技能）。

## 分诊路径（固定顺序，先端点后进程再资源）

1. **健康端点两级定性**：`/healthz` 通但 `/healthz/readiness` 不通 = **依赖面**（DB 连不上或
   启动迁移未完成），转第 2 步看日志首错与 DB 侧；两级都不通 = **进程/端口面**（容器没起来、
   端口未暴露、反代错位），转第 3 步。升级后 readiness 长时间不就绪是**迁移在跑**的常态，
   先看时长与日志再定性为故障。
2. **日志首错**：容器日志尾部按**首条 ERROR/异常栈**定位（反复重启看每次重启前的最后
   50 行）：连不上 DB（连接串/凭据/网络）、`.n8n` 目录权限、端口占用是三大高频首错；
   3.0 后 `~/.n8n/binaryData` 与 `~/.n8n/storage` 并存会**启动失败**（升级残留形态）。
3. **进程/容器形态**：重启计数持续增长 = 崩溃循环，退出码与首错对齐定性；`docker ps`
   里端口映射与实际访问端口对不上是反代/暴露面问题，不是实例故障。
4. **资源面**：CPU 持续高/内存增长——先问「谁在烧」：Code 节点用户代码未隔离（1.x 未开
   task runners / internal 模式）会让用户代码直接跑在主进程里，长循环与大 payload 拖垮
   实例；判读方向是 **task runners external 化 + `EXECUTIONS_TIMEOUT*` 限时**（变更建议），
   而不是单纯加资源。磁盘面看 `.n8n` 目录与（sqlite 形态）数据库文件水位。
5. **日志面定位**：`N8N_LOG_LEVEL`（silent/error/warn/info/debug，排障用 debug）、
   `N8N_LOG_OUTPUT=console,file` 组合、文件路径 `N8N_LOG_FILE_LOCATION`
   （默认 `<n8n 目录>/logs/n8n.log`，单文件 16MB × 保留 100 个）；worker 多实例时
   `N8N_LOG_FILE_COUNT_MAX` 应显式设置（默认值会互相覆盖）。注意：**执行明细不在日志文件里**
   （存 DB 的 execution record），别在日志里找节点级输入输出。

## 判读基准

- 「实例挂了」必须落到两级端点 + 进程证据三选一：**不可达**（进程/网络面）、**可达但未就绪**
  （DB/迁移面）、**就绪但表现异常**（配置/负载面）——禁笼统下结论；
- readiness 与 healthz 的差异本身就是结论素材（差在哪一层就是哪一层的病）；
- 资源类结论必须带时间分布证据（持续 vs 突发、与执行高峰对齐与否）；
- 诊断类临时改 `N8N_LOG_LEVEL=debug` 属变更动作：进建议面、注明需重启生效与回退。

## 输出要求

- 每个结论附端点返回/日志关键行证据；变更类动作（重启实例、改 env、开 task runners、
  清数据目录）标注影响面并 requires_approval；数据不足输出「需补充采集」清单
  （如完整首错日志段、docker inspect 的挂载与 env 键名清单、部署形态说明），不臆测。
