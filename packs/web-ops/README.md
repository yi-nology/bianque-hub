# web-ops — 网页巡检取证包

无头 Chromium 渲染级巡检：SPA 正文抽取、页面截图取证、白屏/渲染异常分型。
纯 HTTP 采集（datasources/a2ax）看不见的"浏览器里才发生的事"，是这个包的职责面。

## 前置条件（缺一不可）

1. **bianque-tools 含 browser-ops server**（批次一百四十八起随主二进制工具面分发；
   `bianque-tools manifest browser-ops` 可验）；
2. 部署机已装 Chromium（包管理器安装），或 `BQ_BROWSER_PATH` 指向既有二进制——
   工具面**不做运行时下载**；
3. conf `mcp.servers` 启用（缺省关，显式开闸）：

```yaml
mcp:
  servers:
    browser-ops:
      enabled: true
      command: ["${BQ_TOOLS_BIN}", "browser-ops"]
      timeout: 2m
```

## 内容

| 件 | 说明 |
|---|---|
| `imports/web-inspector` | 网页巡检取证专家：只读档采集→判读→取证固化三段式；交互动作仅受审使用 |
| `web-console-screenshot` | 截图取证方法论：三件套（截图+正文摘录+最终 URL）固定产出 |
| `spa-render-diagnosis` | SPA 白屏四分型（前端路由/资源失败/接口失败/样式遮蔽）：判据+探查表达式+下一步 |

## 安装

```bash
go run ./cmd/bq-markettool install --url https://github.com/yi-nology/bianque-hub \
  --pack web-ops --api http://<实例>:8900
```

装包后专家/技能即入路由面：说「网页巡检」「页面取证」「spa白屏」类诉求即达
（P4 复合词，与内置词表零碰撞）。装包≠开洞——browser-ops server 仍需 conf 显式
启用（缺省硬关纪律）。

## 工具分档

只读档（open/content/screenshot/tabs/close）为专家默认采集面；执行档
（click/type/eval）改变页面状态或执行 JS，走扁鹊审批/policy 闸，专家提示词层面
要求逐次留痕表达式与理由。
