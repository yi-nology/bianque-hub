# CHANGELOG

## 0.1.3 (2026-11-17)

- 链模板升 v3 图模型（`version: 3` + nodes/edges），模板变量改用显式 v3 写法：
  `{{sys.input}}` / `{{sys.host}}`，上一步产出按节点 id 引用
  （`{{n1.output.summary}}` / `{{n1.output.followups}}`）。v2 兼容占位
  `{{input}}` / `{{host}}` / `{{summary}}` / `{{followups}}` 退役。

## 0.1.2 (2026-10-06)

- 模板专家 prompt 示范平台统一报告 schema：starter-analyst 原自造 `conclusion`
  单字段输出，改为 conclusion/confidence/recommendation.steps 三件套并显式钉
  `requires_approval`/`decision` 治理语义（只读分析恒 false/none，变更建议
  true/pending_approval）——复制起点即正确口径，贡献者照抄不再产出非协议报告。

## 0.1.1 (2026-10-04)

- 链模板变量教学增强：第一步点名 `{{host}}`（目标机），第二步点名 `{{summary}}`
  （上一步结论摘要）与 `{{followups}}`（上一步待跟进事项——引擎结构化接力指令
  Followup{By/What/Why}，无则空串）；chain.yaml 头注补全变量词表。与平台批次
  一百二十七注入面升级（artifacts 分级摘要视图+接力指令）对齐。

## 0.1.0 (2026-09-30)

- 首发：最小示例包（imports/starter-analyst 专家 + starter-method 技能 +
  workflow/starter-chain 两步链，含技能钉扎演示）。
