## 0.2.0 (2026-10-10)

- **nginx-platform-analyst 工具授权最小权限收敛**（bianque 批次二百六十四，除 oo-devops 外全 hub 清扫）：
  `ask-ops allow: []`（=全部 14 工具，含破坏性 run_approved_command(s)）改显式 12 工具
  只读白名单（十结构化采集器+probe_host+run_readonly 只读 CLI 对）。修正三处不合理：
  ①授权面与 prompt 能力声明矛盾（分析专家=只读采集判读域，prompt 零变更工具引用，
  处置建议只随报告进审批）；②`allow: []` 使工具集含变更类，agentrun 重试守卫整步骤
  收紧（RetryAfterMutation=false，429 限流拒重试环节 failed 实弹在案，bianque 侧
  agentkit v0.14.6 豁免缝治标、授权收敛治本）；③能力/路由事实源失真——工具清单是
  派工与路由的能力面，超授即误描述。域 CLI 采集面（run_readonly 白名单内 mysql/psql/
  redis-cli/kafka/docker/kubectl 只读子命令等）完整保留，零能力损失。

## 0.1.0（2026-10-07）

- 首版：可用性面（502/504/upstream 三分/worker 打满）、路由重写面（404/循环/
  root-alias/proxy_pass 尾斜杠）、TLS 面（证书链/握手/SNI/混合内容）、生命周期面
  （-t 前置纪律/reload 语义/USR2 平滑升级/logrotate USR1）四技能 + Nginx 平台诊断
  专家（ask-ops 只读采集面；reload/restart/证书替换类动作恒审批后宿主侧执行）。
