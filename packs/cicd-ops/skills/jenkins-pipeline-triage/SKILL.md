---
name: jenkins-pipeline-triage
description: Jenkins 故障分诊方法论：agent 掉线、构建队列堆积与 executor 饥饿、JENKINS_HOME 磁盘、插件与凭证面、控制器负载——固定顺序定位与判读基准（ask-ops 只读采集面）。
mode: on_demand
version: 0.1.0
maturity: experimental
---

## 触发条件

- 症状关键词：构建失败、构建排队、jenkins慢、agent掉线、节点离线、构建卡住、jenkins不可用
- 组合场景：制品推送失败转 harbor-triage；目标集群部署失败转 argocd-sync-triage / k8s-ops

## 数据来源（ask-ops 只读面）

- REST GET：`<jenkins-url>/computer/api/json`（各 agent offline/idle/executors 状态）、`/queue/api/json`（排队项与阻塞原因）、`/api/json?tree=jobs[color]`（job 健康色板）；
- 失败上下文：具体 job 的 `lastFailedBuild` 控制台日志尾部（GET）；
- 主机侧：控制器进程（java）CPU/内存、JENKINS_HOME 磁盘水位（du/df）、日志尾部（jenkins.log）；
- 凭证由主机侧已配置环境注入（API token），对话不回显。

## 分诊路径（按症状类定位，固定顺序）

1. **agent 掉线**：`/computer/api/json` 逐 agent 看 offline 标志与 offlineCause → 归因三分：agent 进程/宿主死（主机侧核实，联动 os-basics）、连通性问题（JNLP/TLS 端口、反向代理超时）、磁盘满触发自动下线（agent 所在机 df）。标记「temporarilyOffline」与配置性下线不算故障。
2. **队列堆积与 executor 饥饿**：`/queue/api/json` 看 blocked/stuck 项与 why 字段 → 有空闲 executor 仍堆积=并发/标签约束不匹配（job 要的 label 无在线 agent）；executor 全忙=容量不足（扩容是变更建议）。
3. **构建失败分类**：console 日志尾部三分——**环境面**（拉代码失败：网络/凭证；依赖拉取失败：仓库不可达）、**真失败**（编译/测试断言：交业务修复，不是运维域）、**资源面**（OOM 被 kill、磁盘满写不了 workspace）。
4. **JENKINS_HOME 磁盘**：du 定位大户（workspace/builds 残留、日志堆积）→ 磁盘 >90% 时 Jenkins 会拒绝构建（保护形态）；清理旧构建是变更建议（先标注保留策略）。
5. **插件与控制器面**：失败首时间点与插件升级对齐（plugins 目录时间戳/manage 界面证据）；控制器 GC 抖动/高 CPU（主机侧）会放大所有构建延迟，属根因上游。

## 判读基准

- 「偶发 vs 持续」：结论必须带失败时间分布（首失败时间、复现窗口）；
- agent offline 修复动作（重连、重启 agent、清磁盘）是变更类建议，需标注影响（该节点上正在跑的构建）；
- 队列里 `stuck=true` 的项要看 blocked-by 关系（上游 job/资源锁），禁直接说「删队列」了事。

## 输出要求

- 每个结论附 API/日志关键行证据；变更类动作（重启 agent、清 workspace、扩 executor、插件回滚）标注影响面并 requires_approval；数据不足输出「需补充采集」清单（如完整 console 日志），不臆测。
