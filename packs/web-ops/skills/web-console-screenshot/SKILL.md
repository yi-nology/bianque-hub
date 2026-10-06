---
name: web-console-screenshot
description: 页面/控制台截图取证方法论：无头 Chromium 打开目标、确认渲染、全页与视口双档截图、tab 回收——取证三件套（截图+正文摘录+最终 URL）的固定产出形态（browser-ops 只读档）。
mode: on_demand
version: 0.2.0
maturity: experimental
requires_mcp:
  - server: browser-ops
    tools: [browser_open, browser_content, browser_screenshot, browser_tabs, browser_close]
---

## 触发条件

- 用户要求：截图取证、页面留档、报单附件、把"现在页面长什么样"固化下来；
- 巡检结论需要视觉证据补强时（正文判读完，截图固化）。

## 数据来源

`requires_mcp` 声明 browser-ops 只读档五工具；本技能不使用执行档（click/type/eval
不进本技能步骤）。**前提**：browser-ops 是站点 conf `mcp.servers` 显式启用面
（缺省硬关）——工具不在场时不空转重试，输出启用指引（conf 增加 browser-ops
server 并重启）并如实报告不可用。

## 方法论（固定顺序）

1. **打开**：`browser_open` 目标 URL（timeout_seconds 按站点快慢 30-60s）；记录
   返回的 `tab_id` 与**重定向后最终 URL**——取证以最终 URL 为准；
2. **渲染确认**：`browser_content`（max_chars 800）确认非空、识别页面类型；
   登录墙/错误页如实记入结论，不算取证失败；
3. **截图**：默认 `browser_screenshot` 全页（quality 75）；仅当用户要"当前视口"
   时 `viewport_only=true`；记录 `bytes`——异常小的截图（<10KB）多为空白页，
   回到渲染确认复判；
4. **回收**：`browser_close` 该 tab（取证物料已在结果里，不占 tab 池）。

## 输出要求

- 三件套缺一不可：截图（base64+bytes）、正文摘录（≤300 字）、最终 URL；
- 结论一句话：页面状态（正常/登录墙/错误页/空白）+ 与用户诉求的关联；
- 严禁只给截图不给正文判读——截图没有判读不算取证。
