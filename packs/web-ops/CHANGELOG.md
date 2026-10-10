## 0.3.0 (2026-10-10)

- **web-inspector 工具授权显式化**（bianque 批次二百六十四 最小权限清扫，除 oo-devops 外全 hub 收敛）：
  `browser-ops allow: []`（=全量授予语义）改显式 8 工具白名单——只读档五工具
  （open/content/screenshot/tabs/close）+执行档三工具（click/type/eval）。执行档是
  技能钉扎的设计能力（spa-render-diagnosis 的 eval 只读三连等），prompt 已带执行档
  纪律（必须说明必要性/严禁 eval 绕只读档），browser-ops server 治理缺省硬关
  （装包≠开洞）；显式列出防「allow 空=全量」在新工具注册时静默扩权。

# Changelog

## 0.2.0 - 2026-10-06

### Changed

- **专家 slug 改 `imports/web-inspector`**（原 `specialists/` 前缀是平台命名空间，
  社区新领域包一律 `imports/` 防撞平台域——首发窗口期改名，成本最低）。
- 专家 prompt 与两技能数据来源段补 browser-ops「conf 缺省关、显式启用」前提与
  工具缺席降级口径（不空转重试，输出启用指引后收口）——obs-ops 0.3.0
  loki-triage「装得上用不了」事故形态的包侧预防。

### Fixed

- 专家执行档纪律与 spa-render-diagnosis 白屏探查互斥：prompt 为「技能钉扎的固定
  只读 DOM 探查表达式」开明确例外口径（仍逐次留痕），状态改变型交互纪律不变。
- **输出协议对齐（215 实测修复）**：prompt 输出要求钉平台统一报告 JSON 骨架
  （agent_name/domain/symptom「检测对象：」前缀/conclusion/evidence 三键/
  confidence/decision/summary）——实弹会话暴露原散文式输出要求产出非协议 JSON，
  被协议门拦截重试后整轮作废兜底到模板专家；只读定位 requires_approval=false、
  decision=none 与全仓口径拉齐。

## 0.1.0 - 2026-10-05

### Added

- 首版：`specialists/web-inspector` 网页巡检取证专家 +
  `web-console-screenshot` / `spa-render-diagnosis` 两技能。
- 依赖 browser-ops 工具面（bianque-tools 批次一百四十八起）；conf 缺省关、
  显式启用（缺省硬关纪律）。
