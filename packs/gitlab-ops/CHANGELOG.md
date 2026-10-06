# CHANGELOG

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
