---
name: skywalking-run-e2e
description: 本地运行 SkyWalking E2E 测试
mode: on_demand
version: 1.1.0
maturity: experimental
---

# 运行 SkyWalking E2E 测试

使用 `skywalking-infra-e2e` 运行 E2E 测试用例。用户提供测试用例路径（例如 `simple/jdk`, `storage/banyandb`, `alarm`）。

## 前置条件

所有工具需要安装 **Go**。检查 `.github/workflows/` 以获取 CI 中使用的确切 `e2e` commit。

### e2e CLI

从 [apache/skywalking-infra-e2e](https://github.com/apache/skywalking-infra-e2e) 构建，由 CI 中的 commit 固定：

```bash
# 安装固定的 commit
go install github.com/apache/skywalking-infra-e2e/cmd/e2e@<commit-id>

# 或者本地克隆和构建（调试 e2e 工具本身时有用）
git clone https://github.com/apache/skywalking-infra-e2e.git
cd skywalking-infra-e2e
git checkout <commit-id>
make build
# 二进制文件在 bin/e2e —— 添加到 PATH 或复制到 $GOPATH/bin
```

### swctl, yq, 和其他工具

E2E 测试用例运行预安装步骤（参见每个 `e2e.yaml` 中的 `setup.steps`），将工具安装到 `/tmp/skywalking-infra-e2e/bin`。
本地运行时，您需要在 PATH 上安装这些工具。

**swctl** — SkyWalking CLI，用于在验证案例中查询 OAP 的 GraphQL API。
固定在 `test/e2e-v2/script/env` 中的 `SW_CTL_COMMIT`：

```bash
# 选项 1：使用安装脚本（与 CI 相同）
bash test/e2e-v2/script/prepare/setup-e2e-shell/install.sh swctl
export PATH=/tmp/skywalking-infra-e2e/bin:$PATH

# 选项 2：从源码构建
go install github.com/apache/skywalking-cli/cmd/swctl@<SW_CTL_COMMIT>
```

**yq** — YAML 处理器，用于验证案例：

```bash
# 选项 1：使用安装脚本
bash test/e2e-v2/script/prepare/setup-e2e-shell/install.sh yq
export PATH=/tmp/skywalking-infra-e2e/bin:$PATH

# 选项 2：brew install yq (macOS)
```

**其他工具**（仅特定测试用例需要）：

| 工具 | 安装脚本 | 使用者 |
|------|---------------|---------|
| `kubectl` | `install.sh kubectl` | Kubernetes-based tests |
| `helm` | `install.sh helm` | Helm chart tests |
| `istioctl` | `install.sh istioctl` | Istio/service mesh tests |
| `etcdctl` | `install.sh etcdctl` | etcd cluster tests |

所有安装脚本位于 `test/e2e-v2/script/prepare/setup-e2e-shell/`。

## 步骤

### 1. 确定测试用例

将用户的参数解析为 `test/e2e-v2/cases/` 下的完整路径。如果模糊，列出匹配的目录并询问。

```bash
ls test/e2e-v2/cases/<argument>/e2e.yaml
```

### 2. 检查是否需要重建

将源文件时间戳与上次构建进行比较：

```bash
# OAP 服务器自上次构建以来的更改
find oap-server apm-protocol -type f \( \
  -name "*.java" -o -name "*.yaml" -o -name "*.yml" -o \
  -name "*.json" -o -name "*.xml" -o -name "*.properties" -o \
  -name "*.proto" \
\) -newer dist/apache-skywalking-apm-bin.tar.gz 2>/dev/null | head -5

# 测试服务自上次构建以来的更改
find test/e2e-v2/java-test-service -type f \( \
  -name "*.java" -o -name "*.xml" -o -name "*.yaml" -o -name "*.yml" \
\) -newer test/e2e-v2/java-test-service/e2e-service-provider/target/*.jar 2>/dev/null | head -5
```

如果找到文件，警告用户并建议在运行前重建。

### 3. 如果需要则重建（仅用户确认后）

```bash
# 重建 OAP
./mvnw clean flatten:flatten package -Pall -Dmaven.test.skip && make docker

# 重建测试服务
./mvnw -f test/e2e-v2/java-test-service/pom.xml clean flatten:flatten package
```

### 4. 运行 E2E 测试

设置所需的环境变量并运行：

```bash
export SW_AGENT_JDK_VERSION=8
e2e run -c test/e2e-v2/cases/<case-path>/e2e.yaml
```

### 5. 如果测试失败

不要立即运行 cleanup。而是：

1. 检查容器日志：
   ```bash
   docker compose -f test/e2e-v2/cases/<case-path>/docker-compose.yml logs oap
   docker compose -f test/e2e-v2/cases/<case-path>/docker-compose.yml logs provider
   ```

2. 单独运行验证（调查后可以重试）：
   ```bash
   e2e verify -c test/e2e-v2/cases/<case-path>/e2e.yaml
   ```

3. 仅在调试完成后 cleanup：
   ```bash
   e2e cleanup -c test/e2e-v2/cases/<case-path>/e2e.yaml
   ```

## 常见测试用例

| 简写 | 路径 |
|-----------|------|
| `simple/jdk` | `test/e2e-v2/cases/simple/jdk/` |
| `storage/banyandb` | `test/e2e-v2/cases/storage/banyandb/` |
| `storage/elasticsearch` | `test/e2e-v2/cases/storage/elasticsearch/` |
| `alarm` | `test/e2e-v2/cases/alarm/` |
| `log` | `test/e2e-v2/cases/log/` |
| `profiling/trace` | `test/e2e-v2/cases/profiling/trace/` |
