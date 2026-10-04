---
name: skywalking-compile
description: 构建 SkyWalking OAP 服务器，运行 javadoc 检查，并验证 checkstyle。用于在提交 PR 前验证更改。
mode: on_demand
version: 1.1.0
maturity: experimental
---

# 编译与验证

构建项目并运行与 CI 匹配的静态检查。

## 前置条件

- JDK 11, 17, or 21 (LTS 版本)
- Maven 3.6+ (使用 `./mvnw` wrapper)

## Maven profiles

- `backend` (默认): 构建 OAP 服务器模块
- `ui` (默认): 构建 web 应用
- `dist` (默认): 创建分发包
- `all`: 构建所有内容包括子模块初始化

## 按参数的命令

### `all` 或无参数 — 完整 CI 构建

```bash
./mvnw clean flatten:flatten install javadoc:javadoc -B -q -Pall \
  -Dmaven.test.skip \
  -Dcheckstyle.skip \
  -Dgpg.skip
```

### `backend` — 仅后端（更快）

```bash
./mvnw clean flatten:flatten package -Pbackend,dist -Dmaven.test.skip
```

### `javadoc` — 仅 javadoc 检查

Javadoc 需要 delombok 输出，所以必须先运行 `install`：

```bash
./mvnw clean flatten:flatten install javadoc:javadoc -B -q -Pall \
  -Dmaven.test.skip \
  -Dcheckstyle.skip \
  -Dgpg.skip
```

单独运行 `javadoc:javadoc` 而不带 `install` 会遗漏错误，因为 `${delombok.output.dir}` 不会被填充。

### `checkstyle` — 仅 checkstyle

```bash
./mvnw -B -q clean flatten:flatten checkstyle:check
```

### 模块名称 — 单模块构建

```bash
./mvnw clean flatten:flatten package -pl oap-server/analyzer/<module-name> -Dmaven.test.skip
```

## 阅读 javadoc 输出

Maven 在所有 javadoc 输出前加上 `[ERROR]`，但实际严重级别在
行号之后的消息中。只有包含 `error:` 的行会导致构建失败；包含 `warning:` 的行不会。

```
[ERROR] Foo.java:42: error: bad use of '>'        ← 实际错误（必须修复）
[ERROR] Foo.java:50: warning: no @param for <T>   ← 警告（不会导致构建失败）
```

### 常见 javadoc 错误

| 错误 | 原因 | 修复 |
|-------|-------|-----|
| `bad use of '>'` | javadoc HTML 中的裸 `>` (例如 `<pre>` 块中的 `->`) | 使用 `{@code ->}` 或 `-&gt;` |
| heading out of sequence | 标题级别跳过预期的层级 | 参见下面的标题规则 |
| reference not found | `{@link Foo#bar()}` 签名错误 | 匹配确切的参数类型: `{@link Foo#bar(ArgType)}` |

### Javadoc 标题规则 (JDK 13+)

JDK 13 引入了严格的标题验证。JDK 11 不强制执行，但 JDK 17/21/25 会。
为向前兼容，正确编写标题：

| Javadoc 位置 | 起始标题级别 |
|---|---|
| Class, interface, enum, package, module | `<h2>` |
| Constructor, method, field | `<h4>` |
| Standalone HTML files (`doc-files/`) | `<h1>` |

生成的 javadoc 页面使用 `<h1>` 表示类名，`<h3>` 表示成员部分 (Methods, Fields 等)，
所以类级子章节必须使用 `<h2>`，方法级子章节必须使用 `<h4>` 以维持正确的嵌套。

## CI 参考

CI 在 Linux 上使用 JDK 11。`dist-tar` job 运行：

```bash
./mvnw clean flatten:flatten install javadoc:javadoc -B -q -Pall \
  -Dmaven.test.skip \
  -Dcheckstyle.skip \
  -Dgpg.skip
```

`code-style` job 运行：

```bash
./mvnw -B -q clean flatten:flatten checkstyle:check
```
