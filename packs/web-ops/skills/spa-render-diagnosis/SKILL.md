---
name: spa-render-diagnosis
description: SPA 白屏/渲染异常分型方法论：渲染采集→DOM 状态探查（eval 受审）→四分型判读（前端路由/资源失败/接口失败/样式遮蔽）——每型有判据与下一步，禁止"打不开"一句话收场（browser-ops 工具面）。
mode: on_demand
version: 0.2.0
maturity: experimental
requires_mcp:
  - server: browser-ops
    tools: [browser_open, browser_content, browser_eval, browser_screenshot]
---

## 触发条件

- 页面白屏、SPA 路由页渲染异常、"页面打开了但什么都没有"；
- browser_content 正文为空/只有壳文本（`<div id="app">` 类容器无内容）。

## 数据来源

`requires_mcp` 声明 browser-ops 四工具。**前提**：browser-ops 是站点 conf
`mcp.servers` 显式启用面（缺省硬关）——工具不在场时不空转重试，输出启用指引
（conf 增加 browser-ops server 并重启）并如实报告不可用。其中 `browser_eval`
是执行档：本技能的探查表达式固定为只读 DOM 查询（JSON.stringify 包裹），不做
任何写操作；该固定探查属只读判读采集，不受专家「执行档须用户明确要求」纪律
阻断（专家 prompt 已开对应例外口径）。

## 方法论（固定顺序）

1. **采集**：`browser_open`（60s 预算）→ `browser_content` 拿当前正文与标题；
2. **DOM 探查**（正文空/壳文本时）：`browser_eval` 依次查——
   - `document.getElementById('app')||document.getElementById('root')` 的
     `innerHTML.length`（容器在但空=前端未挂载）；
   - `document.scripts.length` 与 `performance.getEntriesByType('resource').length`
     （0=资源完全没加载，指向网络/网关层）；
   - `document.body.innerText.length`（有 DOM 无可见文本=样式遮蔽/字体缺失）；
3. **分型判读**（按探查结果对号，每型给判据）：
   - **前端路由型**：容器存在且空 + URL 含深链路径 → 前端路由未兜住该路径或
     静态资源 404，下一步核对静态资源部署与路由 fallback；
   - **资源失败型**：resource 条目为 0 或异常少 → JS/CSS 未送达，下一步核对
     CDN/网关与静态资源路径；
   - **接口失败型**：容器挂载但列表空 + 页面依赖的 XHR 域名与页面不同源 → 后端
     接口/CORS，下一步转到接口巡检（非本技能面）；
   - **样式遮蔽型**：有 DOM 有文本但截图空白 → CSS/字体，低置信标记；
4. **取证固化**：`browser_screenshot` 全页；把分型、判据命中项、截图三件对应。

## 输出要求

- 分型结论 + 命中判据逐条列出（探查表达式的返回值原样引用）；
- 下一步动作按型给出，接口失败型明确移交边界（本技能只判到"接口层"为止）；
- `browser_eval` 的每次执行在输出中留表达式原文。
