---
name: k8s-dev-deploy
description: Kurtosis 开发镜像构建与部署 - 多架构 Docker 镜像构建、推送至 Docker Hub 并部署到 Kubernetes 集群进行本地测试
mode: on_demand
version: 1.1.0
maturity: experimental
---

# K8s Dev 部署

从源代码构建并部署 Kurtosis 到 Kubernetes 集群进行测试，而不制作发布。

## 概述

必须构建和推送三个容器镜像：
- `engine` — 来自 `engine/server/Dockerfile`
- `core` (APIC) — 来自 `core/server/Dockerfile`
- `files-artifacts-expander` — 来自 `core/files_artifacts_expander/Dockerfile`

CLI 二进制文件也在本地重建。

所有镜像共享相同的标签，该标签在 `kurtosis_version/kurtosis_version.go` 中设置为 `KurtosisVersion`。此版本被编译到 engine 二进制文件中，并在运行时使用以拉取 core 和 files-artifacts-expander 镜像。

## 步骤

### 1. 确定 Docker Hub 用户名并生成唯一标签

```bash
DOCKER_USER=$(docker info 2>/dev/null | grep Username | awk '{print $2}')
TAG="$(git rev-parse --short HEAD)-$(date +%s)"
```

### 2. 更新镜像引用以使用用户的 Docker Hub

编辑这三个常量以将 `kurtosistech` 替换为 `$DOCKER_USER`：

| 文件 | Constant | 默认值 |
|------|----------|--------------|
| `engine/launcher/engine_server_launcher/engine_server_launcher.go` | `containerImage` | `kurtosistech/engine` |
| `core/launcher/api_container_launcher/api_container_launcher.go` | `containerImage` | `kurtosistech/core` |
| `core/server/api_container/server/startosis_engine/kurtosis_types/service_config/service_config.go` | `filesArtifactsExpanderImage` | `kurtosistech/files-artifacts-expander` |

### 3. 设置版本标签

编辑 `kurtosis_version/kurtosis_version.go` 并将 `KurtosisVersion` 设置为生成的 `$TAG`。

**重要**: 始终使用唯一标签（包含时间戳）以避免 Kubernetes 节点镜像缓存问题。带有 `imagePullPolicy: IfNotPresent` 的节点不会重新拉取相同标签的镜像。

### 4. 构建所有二进制文件（并行）

```bash
# Engine
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o engine/server/build/kurtosis-engine.arm64 engine/server/engine/main.go
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o engine/server/build/kurtosis-engine.amd64 engine/server/engine/main.go

# Core (APIC)
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o core/server/build/api-container.arm64 core/server/api_container/main.go
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o core/server/build/api-container.amd64 core/server/api_container/main.go

# Files Artifacts Expander
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o core/files_artifacts_expander/build/files-artifacts-expander.arm64 core/files_artifacts_expander/main.go
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o core/files_artifacts_expander/build/files-artifacts-expander.amd64 core/files_artifacts_expander/main.go

# CLI (本地使用)
go build -o /tmp/kurtosis ./cli/cli/
```

### 5. 推送多架构 Docker 镜像（并行）

确保 buildx builder 存在（如需要，使用 `docker buildx create --name kurtosis-builder --use` 创建）。

```bash
docker buildx build --platform linux/amd64,linux/arm64 --builder kurtosis-builder --no-cache -t $DOCKER_USER/engine:$TAG --push engine/server/
docker buildx build --platform linux/amd64,linux/arm64 --builder kurtosis-builder --no-cache -t $DOCKER_USER/core:$TAG --push core/server/
docker buildx build --platform linux/amd64,linux/arm64 --builder kurtosis-builder --no-cache -t $DOCKER_USER/files-artifacts-expander:$TAG --push core/files_artifacts_expander/
```

**重要**: 始终使用 `--no-cache` 以防止 buildx 缓存旧二进制文件。

### 6. 部署到集群

```bash
/tmp/kurtosis engine stop
# 清理任何剩余的 namespaces
kubectl get ns | grep kurtosis | awk '{print $1}' | xargs -r kubectl delete ns
/tmp/kurtosis engine start
```

### 7. 启动 gateway 并测试

```bash
/tmp/kurtosis gateway &
/tmp/kurtosis run github.com/ethpandaops/ethereum-package
/tmp/kurtosis clean -a
```

## 迭代

当进行进一步的代码更改时，每次使用**新的唯一标签**从步骤 3 开始重复。永远不要重用标签 — k8s 节点缓存镜像，在相同标签下不会拉取更新。

## Common issues

- **ImagePullBackOff**: 标签在 Docker Hub 上不存在。验证推送是否成功。
- **尽管推送但旧代码仍在运行**: k8s 节点在相同标签下缓存了镜像。使用新的基于时间戳的标签。
- **Engine start 挂起**: Logs collector DaemonSet pods 在有 taint/不健康的节点上失败。检查 `kubectl get pods -A | grep kurtosis`。
- **Clean 挂起**: fluentbit 清理进程尝试在有 taint 的节点上创建 cleanup pods。尽力修复应该能处理这个问题。
