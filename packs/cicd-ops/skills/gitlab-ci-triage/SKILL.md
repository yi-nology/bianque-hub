---
name: gitlab-ci-triage
description: GitLab CI 分诊方法论：runner 掉线与并发饱和、job 卡 created/pending、注册令牌与配额面、流水线级失败分类、对象存储与日志面——固定顺序定位与判读基准（ask-ops 只读采集面）。
mode: on_demand
version: 0.1.0
maturity: experimental
requires_mcp:            # 采集依赖的工具面（装载期对账依据；旧 bianque-tools 二进制缺这些工具时应先重建）
  - server: ask-ops
    tools: [run_readonly_command, run_readonly_commands]
---

## 触发条件

- 症状关键词：runner掉线、流水线卡住、job一直pending、gitlabci排障、构建失败、runner未接活
- 组合场景：镜像推送失败转 harbor-triage；部署阶段失败转 argocd-sync-triage / k8s-ops

## 数据来源（ask-ops 只读面）

- GitLab REST GET（凭证走已配置 token；URL 含 `?`/`|` 特殊字符一律整体单引号包裹，防 shell 解析）：`'/api/v4/runners?scope=online'`、`'/api/v4/runners?scope=offline'`、`/api/v4/projects/<id>/pipelines`、`/api/v4/projects/<id>/jobs`（status 分布）、job 日志尾部；
- runner 主机侧：`gitlab-runner list` / `gitlab-runner status`、config.toml 的 concurrent 限制、runner 进程与宿主资源；
- GitLab 侧健康：`/api/v4/version`（版本面）、sidekiq 队列延迟证据（管理面，通常只有平台侧可见，无证据时标注）。

## 分诊路径（按症状类定位，固定顺序）

1. **job 卡 created/pending**：`'/api/v4/runners?scope=online'` 确认可用 runner → 三分：**无匹配 runner**（job tag 与 runner tags 不交，或 runner 未分配到该项目/组）、**并发饱和**（runner 全忙，concurrent 配置口径）、**runner 下线**（offline 清单，去第 2 步）。tag 不匹配是最常见的「pending 永动机」根因。
2. **runner 掉线**：runner 主机 `gitlab-runner status` + 进程/容器状态 → 归因三分：进程死（重启是变更建议）、认证失效（注册令牌被轮换/解除注册，重新注册是变更）、宿主资源不足（联动 os-basics）。批量 runner 同时掉线优先查共同上游（网络/注册中心/GitLab 侧变更）。
3. **流水线级失败分类**：失败 job 日志尾部三分——**环境面**（clone 失败：Git 通道/网络；依赖拉取失败）、**真失败**（测试/编译断言，交业务修复）、**资源面**（runner 磁盘满、job 超时 trace 被截断标志）。
4. **配额与令牌面**：项目 CI 分钟数/配额耗尽时 job 直接失败（错误文案证据）；共享 runner vs 自建 runner 归属核对（配额只影响共享面）。
5. **日志与对象存储面**：job 日志缺失/加载失败指向对象存储侧（GitLab 外部依赖），此时构建本身可能正常——归因到存储面，不要误判构建失败。

## 判读基准

- pending 结论必须给出「等待时长 + runner 可用性证据」双要素；短时排队是正常并发形态；
- runner offline 判据是注册心跳超时口径（通常数分钟），瞬时抖动不构成结论；
- CI 分钟数配额只约束共享 runner：自建 runner 不受影响，核对归属后再下结论。

## 输出要求

- 每个结论附 API/日志关键行证据；变更类动作（重启 runner、重新注册、调 concurrent、清配额）标注影响面并 requires_approval；数据不足输出「需补充采集」清单（如 job 完整日志），不臆测。
