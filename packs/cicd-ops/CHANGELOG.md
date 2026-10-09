# CHANGELOG

## 0.3.0 (2026-10-09)

- GitLab 流水线能力面升级（参考 GitLab 官方 CLI 插件 0.1.3 的 glab ci/api 与 preflight
  方法论，原生改写为 curl GET 只读形态——glab 二进制不在 ask-ops 白名单，不引入新工具面）：
  gitlab-ci-triage 0.2.3 批版本戳未动（内容级修复在位），本次统一升 **0.2.0**——
  分诊路径新增第 0 步「token 身份与项目可见性预检」（401/403/空清单语义三分）；
  job 卡住按 created（派活/匹配面）与 pending（runner 执行资源面）两分归因；
  流水线失败先状态语义再读日志（allow_failure/manual 红标≠阻断）；数据来源补
  MR/commit→流水线→pipelines/<pid>/jobs 明细→jobs/<jid>/trace 尾部的精确下钻链
  （`:id` 占位符与 URL 编码路径口径钉死）。
- 编目变更块三件新增（受审执行范式，包内累计五件）：`cicd-gitlab-job-retry`（作业级
  单发重跑，risk 2）、`cicd-gitlab-pipeline-retry`（流水线级重跑，risk 3）、
  `cicd-gitlab-pipeline-cancel`（叫停进行中，risk 3）——重跑粒度两分源自官方 CLI 把
  retry 钉在 job id 的口径，防「单 job 需求发流水线级」混发；verify_readonly 走既有
  curl GET 只读形态，命令依赖主机侧 GITLAB_URL/GITLAB_TOKEN（gitlab 凭证类型同源）。
- pipeline-analyst：route_keywords 补「gitlab流水线失败」（全 hub 未占用复合词；
  「gitlab流水线」被 oo-devops gitlab-employee P6 参考位占用，跨层重复不取）；
  route_desc 更新 GitLab 面语义；prompt 采集纪律钉身份预检、处方纪律接三块编目。

## 0.2.3 (2026-10-06)

- 深审修复批：argocd-sync-triage 杜撰配置名 `selfIgnoreConfig` 修正为真实机制
  `ignoreDifferences`（应用级）/`resource.customizations`（argocd-cm 级）豁免口径
  （0.1.1→0.1.2）；gitlab-ci-triage/jenkins-pipeline-triage REST URL 特殊字符
  （`?`/`|`/`[]`）整体单引号包裹示范（0.1.1→0.1.2）；CHANGELOG 0.1.2 旧条目串包
  引用（es-triage 属 obs-ops）按包内实情改写。
- 215 实弹口径前移：pipeline-analyst 输出铁律补「steps 每条=单条可执行 shell 命令」纪律
  （说明性内容归 needs_followup）。
- 215 实弹口径前移：pipeline-analyst max_iterations 12→24（三段定位采集同型预算风险）。



## 0.2.2 (2026-10-03)

- 批次九十四「诊断+处方+受审执行」接入：新增编目变更块两件——`cicd-argocd-app-sync`
  （手动同步，params: app，verify: app get；前提 diff 影响面已进结论）、
  `cicd-argocd-app-rollback`（回滚到历史 revision，params: app/revision，verify: app
  history）；argocd-sync-triage 登记 `provides_changes` 并在输出要求给 change_ref 处方
  口径（技能升 0.1.1）；pipeline-analyst 提示词处方纪律升为编目优先。

## 0.2.1 (2026-10-02)

- 0.2.0 条目数字按包内实情改写：本包为 4 技能、1 专家（原条目误抄批次合计「14 技能/六位专家」）。

## 0.2.0 (2026-10-02)

- 4 技能（jenkins-pipeline-triage/gitlab-ci-triage/harbor-triage/argocd-sync-triage）补 requires_mcp 声明
  （ask-ops：run_readonly_command/run_readonly_commands）——
  采集依赖显式化，装载期可对账（旧 bianque-tools 二进制缺工具时先重建再装包）。
- pipeline-analyst 专家新增「采集脱敏」纪律：docker inspect Env / kubectl describe 环境变量等
  凭证面字段引用前打码；Secret 材料面（kubectl get secrets、config view --raw）
  已被工具层拒收，禁绕过。

## 0.1.3 (2026-10-02)

- 采集契约升级：例检面板改用 ask-ops 批量形态 `run_readonly_commands`（≤10 条
  单会话收口、逐命令退出码、整批先审后发）。

## 0.1.2 (2026-10-02)

- 处方修正（对齐 run_readonly_command 引号感知守卫）：harbor-triage 证书探测去
  `2>/dev/null`（stderr 本就不采集，重定向会被拒），`_cat`/查询类 URL 带 `&`
  查询参数时给引号形态示范。

## 0.1.1 (2026-10-02)

- 采集契约对齐平台 ask-ops 新工具 `run_readonly_command`（受审只读命令：白名单
  + 注入防线 + 尾部截断/退出码），专家 prompt 点名调用方式。

## 0.1.0 (2026-10-02)

- 首发（oo-devops 拆分重构第一批）：jenkins-pipeline-triage / gitlab-ci-triage /
  harbor-triage / argocd-sync-triage 四技能（收割重写）
  + imports/pipeline-analyst 专家（ask-ops 只读面）。
