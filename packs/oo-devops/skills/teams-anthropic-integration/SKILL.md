---
name: teams-anthropic-integration
description: 使用 @youdotcom-oss/teams-anthropic 将 Anthropic Claude 模型（Opus、Sonnet、Haiku）添加到 Microsoft Teams.ai 应用程序。可选择集成 You.com MCP 服务器以进行网络搜索和内容提取。
mode: on_demand
version: 1.1.0
maturity: experimental
---

# 使用 Anthropic Claude 构建 Teams.ai 应用

使用 `@youdotcom-oss/teams-anthropic` 将 Claude 模型（Opus、Sonnet、Haiku）添加到 Microsoft Teams.ai 应用程序。可选择集成 You.com MCP 服务器以进行网络搜索和内容提取。

## 选择你的路径

**路径 A：基本设置**（推荐用于入门）
- 在 Teams.ai 中使用 Anthropic Claude 模型
- 聊天、流式传输、函数调用
- 无额外依赖

**路径 B：带 You.com MCP**（用于网络搜索功能）
- 路径 A 中的所有内容
- 通过 You.com 进行网络搜索和内容提取
- 实时信息访问

## 决策点

**询问：你的 Teams 应用中是否需要网络搜索和内容提取？**

- **否** → 使用 **路径 A：基本设置**（更简单、更快）
- **是** → 使用 **路径 B：带 You.com MCP**

---

## 路径 A：基本设置

在你的 Teams.ai 应用中使用 Anthropic Claude 模型，无需额外依赖。

### A1. 安装包

```bash
npm install @youdotcom-oss/teams-anthropic @anthropic-ai/sdk @microsoft/teams.ai
```

### A2. 获取 Anthropic API 密钥

从 [console.anthropic.com](https://console.anthropic.com/) 获取你的 API 密钥

```bash
# 添加到 .env
ANTHROPIC_API_KEY=your-anthropic-api-key
```

### A3. 询问：新建还是现有应用？

- **新建 Teams 应用**：使用下面的完整模板
- **现有应用**：将 Claude 模型添加到现有设置

### A4. 基本模板

**对于新应用：**

```typescript
import { App } from '@microsoft/teams.apps';
import { AnthropicChatModel, AnthropicModel } from '@youdotcom-oss/teams-anthropic';

if (!process.env.ANTHROPIC_API_KEY) {
  throw new Error('ANTHROPIC_API_KEY environment variable is required');
}

const model = new AnthropicChatModel({
  model: AnthropicModel.CLAUDE_SONNET_4_5,
  apiKey: process.env.ANTHROPIC_API_KEY,
  requestOptions: {
    max_tokens: 2048,
    temperature: 0.7,
  },
});

const app = new App();

app.on('message', async ({ send, activity }) => {
  await send({ type: 'typing' });

  const response = await model.send(
    { role: 'user', content: activity.text }
  );

  if (response.content) {
    await send(response.content);
  }
});

app.start().catch(console.error);
```

**对于现有应用：**

添加到你的现有导入：
```typescript
import { AnthropicChatModel, AnthropicModel } from '@youdotcom-oss/teams-anthropic';
```

替换你现有的模型：
```typescript
const model = new AnthropicChatModel({
  model: AnthropicModel.CLAUDE_SONNET_4_5,
  apiKey: process.env.ANTHROPIC_API_KEY,
});
```

### A5. 选择你的模型

```typescript
// 最强大 - 最适合复杂任务
AnthropicModel.CLAUDE_OPUS_4_5

// 智能和速度平衡（推荐）
AnthropicModel.CLAUDE_SONNET_4_5

// 快速高效
AnthropicModel.CLAUDE_HAIKU_3_5
```

### A6. 测试基本设置

```bash
npm start
```

在 Teams 中发送消息以验证 Claude 是否响应。

---

## 路径 B：带 You.com MCP

将网络搜索和内容提取添加到你的 Claude 驱动的 Teams 应用。

### B1. 安装包

```bash
npm install @youdotcom-oss/teams-anthropic @anthropic-ai/sdk @microsoft/teams.ai @microsoft/teams.mcpclient
```

### B2. 获取 API 密钥

- **Anthropic API 密钥**：[console.anthropic.com](https://console.anthropic.com/)
- **You.com API 密钥**：[you.com/platform/api-keys](https://you.com/platform/api-keys)

```bash
# 添加到 .env
ANTHROPIC_API_KEY=your-anthropic-api-key
YDC_API_KEY=your-you-com-api-key
```

### B3. 询问：新建还是现有应用？

- **新建 Teams 应用**：使用下面的完整模板
- **现有应用**：将 MCP 添加到现有 Claude 设置

### B4. MCP 模板

**对于新应用：**

```typescript
import { App } from '@microsoft/teams.apps';
import { ChatPrompt } from '@microsoft/teams.ai';
import { ConsoleLogger } from '@microsoft/teams.common';
import { McpClientPlugin } from '@microsoft/teams.mcpclient';
import {
  AnthropicChatModel,
  AnthropicModel,
  getYouMcpConfig,
} from '@youdotcom-oss/teams-anthropic';

// 验证环境
if (!process.env.ANTHROPIC_API_KEY) {
  throw new Error('ANTHROPIC_API_KEY environment variable is required');
}

if (!process.env.YDC_API_KEY) {
  throw new Error('YDC_API_KEY environment variable is required');
}

// 配置日志记录器
const logger = new ConsoleLogger('mcp-client', { level: 'info' });

// 创建带 MCP 集成的提示
const prompt = new ChatPrompt(
  {
    instructions: 'You are a helpful assistant with access to web search and content extraction. Use these tools to provide accurate, up-to-date information.',
    model: new AnthropicChatModel({
      model: AnthropicModel.CLAUDE_SONNET_4_5,
      apiKey: process.env.ANTHROPIC_API_KEY,
      requestOptions: {
        max_tokens: 2048,
      },
    }),
  },
  [new McpClientPlugin({ logger })],
).usePlugin('mcpClient', getYouMcpConfig());

const app = new App();

app.on('message', async ({ send, activity }) => {
  await send({ type: 'typing' });

  const result = await prompt.send(activity.text);
  if (result.content) {
    await send(result.content);
  }
});

app.start().catch(console.error);
```

**对于已有 Claude 的现有应用：**

如果你已经设置了路径 A，添加 MCP 集成：

1. **安装 MCP 依赖：**
   ```bash
   npm install @microsoft/teams.mcpclient
   ```

2. **添加导入：**
   ```typescript
   import { ChatPrompt } from '@microsoft/teams.ai';
   import { ConsoleLogger } from '@microsoft/teams.common';
   import { McpClientPlugin } from '@microsoft/teams.mcpclient';
   import { getYouMcpConfig } from '@youdotcom-oss/teams-anthropic';
   ```

3. **验证 You.com API 密钥：**
   ```typescript
   if (!process.env.YDC_API_KEY) {
     throw new Error('YDC_API_KEY environment variable is required');
   }
   ```

4. **用 ChatPrompt 替换模型：**
   ```typescript
   const logger = new ConsoleLogger('mcp-client', { level: 'info' });

   const prompt = new ChatPrompt(
     {
       instructions: 'Your instructions here',
       model: new AnthropicChatModel({
         model: AnthropicModel.CLAUDE_SONNET_4_5,
         apiKey: process.env.ANTHROPIC_API_KEY,
       }),
     },
     [new McpClientPlugin({ logger })],
   ).usePlugin('mcpClient', getYouMcpConfig());
   ```

5. **使用 prompt.send() 替代 model.send()：**
   ```typescript
   const result = await prompt.send(activity.text);
   ```

### B5. 测试 MCP 集成

```bash
npm start
```

向 Claude 询问一个需要网络搜索的问题：
- "What are the latest developments in AI?"
- "Search for React documentation"
- "Extract content from https://example.com"

---

## 可用的 Claude 模型

| 模型 | 枚举 | 最适合 |
|-------|------|----------|
| Claude Opus 4.5 | `AnthropicModel.CLAUDE_OPUS_4_5` | 复杂任务，最高能力 |
| Claude Sonnet 4.5 | `AnthropicModel.CLAUDE_SONNET_4_5` | 智能和速度平衡（推荐） |
| Claude Haiku 3.5 | `AnthropicModel.CLAUDE_HAIKU_3_5` | 快速响应，高效 |
| Claude Sonnet 3.5 | `AnthropicModel.CLAUDE_SONNET_3_5` | 上一代，稳定 |

## 高级功能

### 流式响应

```typescript
const response = await model.send(
  { role: 'user', content: 'Write a short story' },
  {
    onChunk: async (delta) => {
      // 在每个 token 到达时流式传输
      process.stdout.write(delta);
    },
  }
);
```

### 函数调用

```typescript
const response = await model.send(
  { role: 'user', content: 'What is the weather in San Francisco?' },
  {
    functions: {
      get_weather: {
        description: 'Get the current weather for a location',
        parameters: {
          location: { type: 'string', description: 'City name' },
        },
        handler: async (args: { location: string }) => {
          // 你的 API 调用在这里
          return { temperature: 72, conditions: 'Sunny' };
        },
      },
    },
  }
);
```

### 对话记忆

```typescript
import { LocalMemory } from '@microsoft/teams.ai';

const memory = new LocalMemory();

// 第一条消息
await model.send(
  { role: 'user', content: 'My name is Alice' },
  { messages: memory }
);

// 第二条消息 - Claude 记得
const response = await model.send(
  { role: 'user', content: 'What is my name?' },
  { messages: memory }
);
// 响应："Your name is Alice."
```

## 验证清单

### 路径 A 清单

- [ ] 包已安装：`@youdotcom-oss/teams-anthropic`
- [ ] 环境变量已设置：`ANTHROPIC_API_KEY`
- [ ] 模型已配置为 `AnthropicChatModel`
- [ ] 已选择模型（Opus/Sonnet/Haiku）
- [ ] 应用已通过基本消息测试

### 路径 B 清单

- [ ] 已完成所有路径 A 项目
- [ ] 额外包已安装：`@microsoft/teams.mcpclient`
- [ ] 环境变量已设置：`YDC_API_KEY`
- [ ] 日志记录器已配置
- [ ] ChatPrompt 已配置为 `getYouMcpConfig()`
- [ ] 应用已通过网络搜索查询测试

## 常见问题

### 路径 A 问题

**"Cannot find module @youdotcom-oss/teams-anthropic"**
```bash
npm install @youdotcom-oss/teams-anthropic @anthropic-ai/sdk
```

**"ANTHROPIC_API_KEY environment variable is required"**
- 从 https://console.anthropic.com/ 获取密钥
- 添加到 .env：`ANTHROPIC_API_KEY=your-key-here`

**"Invalid model identifier"**
- 使用枚举：`AnthropicModel.CLAUDE_SONNET_4_5`
- 不要使用字符串：`'claude-sonnet-4-5-20250929'`

### 路径 B 问题

**"YDC_API_KEY environment variable is required"**
- 从 https://you.com/platform/api-keys 获取密钥
- 添加到 .env：`YDC_API_KEY=your-key-here`

**"MCP connection fails"**
- 在 https://you.com/platform/api-keys 验证 API 密钥是否有效
- 检查网络连接
- 查看日志记录器输出以获取详细信息

**"Cannot find module @microsoft/teams.mcpclient"**
```bash
npm install @microsoft/teams.mcpclient
```

## getYouMcpConfig() 工具

自动配置 You.com MCP 连接：
- **URL**：`https://api.you.com/mcp`
- **认证**：来自 `YDC_API_KEY` 的 Bearer 令牌
- **User-Agent**：包含包版本用于遥测

```typescript
// 选项 1：使用环境变量（推荐）
getYouMcpConfig()

// 选项 2：自定义 API 密钥
getYouMcpConfig({ apiKey: 'your-custom-key' })
```

## 资源

* **包**：https://github.com/youdotcom-oss/dx-toolkit/tree/main/packages/teams-anthropic
* **You.com MCP**：https://documentation.you.com/developer-resources/mcp-server
* **Anthropic API**：https://console.anthropic.com/
* **You.com API 密钥**：https://you.com/platform/api-keys
