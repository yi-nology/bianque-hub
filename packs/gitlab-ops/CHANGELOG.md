## 0.3.0（2026-10-07）

- **实例形态分诊前置**（215 远程实例实弹教训）：`gitlab-instance-triage` 技能与专家提示词
  增「实例地址判定恒为第一步」——先读凭证平面注入的 `GITLAB_URL`，非空=远程形态一律走
  API 面（`$GITLAB_URL/-/readiness` 分项即地图 + `/api/v4/*`），空=共置形态才回落
  `127.0.0.1` 宿主采集；此前实例地址硬编码 127.0.0.1，远程实例整轮采集烧在无 GitLab 的
  主机上（健康三件全空+泛化宿主面耗尽迭代预算，零 GitLab 结论）。

## 0.2.1 (2026-10-07)

- `credentials/gitlab.yaml` 增 `query_auth`（header/PRIVATE-TOKEN/token）：http_query
  消费面查询自动带鉴权——此前包分发类型一律匿名（公开项目 200 与鉴权端点 401 混报、
  membership=true 恒空表的实弹根因）。

## 0.2.0（2026-10-07）

- **mcp/gitlab 只读桥声明**：吃主仓批次一百八十七 MCP 治理面——包声明 server 在控制台
  MCP 页激活（Popconfirm「激活=开洞」确认）+ 按 server 凭证绑定（gitlab:* / gitlab:name），
  即热生效，无需改 conf 重启。桥二进制站点自备（env BQ_GITLAB_MCP_BIN，未部署=启动
  自动剔除不空转）；工具面契约待站点实测后补录（不同桥工具名差异大，不编造）。

# CHANGELOG

## 0.1.1 (2026-10-06)

- 新增 `credentials/gitlab.yaml` 包分发凭证类型：PRIVATE-TOKEN 头 + `/api/v4/user`
  鉴权探测（200=凭证有效/401=token 无效，依赖平台探测 headers 能力）+
  GITLAB_URL/GITLAB_TOKEN env 注入。技能正文「PAT 走主机侧凭证注入」的约定自此有
  平台供给路径；站点 conf 在消费方 MCP server `credentials:` 声明加 `gitlab:*` 后
  自动注入（装包≠开洞）。录入走控制台凭证页或「用一句话创建」（类型注册表随包热装）。

## 0.1.0 (2026-10-06)

- 首发（原生编写，非 oo-devops 血缘整搬；参考库 gitlab 系技能作素材对账）：
  gitlab-instance-triage / gitlab-repo-triage / gitlab-auth-triage /
  gitlab-lifecycle-ops 四技能 + imports/gitlab-platform-analyst 专家（ask-ops
  只读面，与 cicd-ops 流水线面路由错位）。
- 方法论基于 GitLab 官方文档（docs.gitlab.com：omnibus/Docker 安装形态、
  health/readiness/liveness 端点、备份与恢复、升级路径停靠点、secrets 管理）
  整理；CI 流水线与 runner 作业面明确归 cicd-ops，不做二义入口。
- 采集面：curl GET 健康三件与 API（PAT 走主机侧凭证注入，不收集不回显）、
  systemctl/journalctl、docker 只读子命令、gitlab-runner 只读守卫
  （list/status/verify/version/check，需含该守卫的 bianque-tools）；
  gitlab-ctl/gitlab-rake/gitlab-rails/gitlab-psql/docker exec 不在只读白名单，
  相关核验与处置一律审批后宿主侧动作（处方走 recommendation.steps）。
- v0.1.0 未编目变更块：全部处置走自由 steps + requires_approval，编目后续
  按「受审执行」范式按需增补。
