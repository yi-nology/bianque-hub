---
name: a6-plugin-skywalking
description: APISIX SkyWalking 集成专家 - 分布式追踪、服务拓扑、采样配置
mode: on_demand
version: 1.1.0
maturity: experimental
---

# a6-plugin-skywalking

## 概述

`skywalking` plugin 将 APISIX 与 Apache SkyWalking 集成用于
分布式追踪。它为每个请求创建 entry 和 exit spans，
通过 HTTP 将追踪信息上报到 SkyWalking OAP，并启用服务拓扑
可视化和性能分析。

## 何时使用

- 通过 SkyWalking 追踪跨微服务的请求
- 可视化服务拓扑和依赖关系图
- 分析每个路由和每个服务的延迟
- 使用 `skywalking-logger` 将追踪与日志关联

## Plugin 配置参考 (Route/Service)

| 字段 | 类型 | 必需 | 默认值 | 描述 |
|-------|------|----------|---------|-------------|
| `sample_ratio` | number | No | `1` | 采样率从 0.00001 到 1 (1 = 追踪所有请求) |

## 全局配置 (config.yaml)

在 APISIX `config.yaml` 的 `plugin_attr` 下配置：

| 字段 | 类型 | 默认值 | 描述 |
|-------|------|---------|-------------|
| `service_name` | string | `"APISIX"` | SkyWalking UI 中的服务名称 |
| `service_instance_name` | string | `"APISIX Instance Name"` | 实例名称 (使用 `$hostname` 表示动态) |
| `endpoint_addr` | string | `http://127.0.0.1:12800` | SkyWalking OAP HTTP 端点 |
| `report_interval` | integer | `3` | 上报间隔（秒） |

```yaml
plugin_attr:
  skywalking:
    service_name: api-gateway
    service_instance_name: "$hostname"
    endpoint_addr: http://skywalking-oap:12800
    report_interval: 5
```

## 分步：启用 SkyWalking 追踪

### 1. 确保 SkyWalking OAP 正在运行

```bash
# Docker 示例
docker run -d --name skywalking-oap \
  -p 12800:12800 -p 11800:11800 \
  apache/skywalking-oap-server:latest
```

### 2. 配置 APISIX 全局设置

添加到 `config.yaml`：

```yaml
plugin_attr:
  skywalking:
    service_name: my-gateway
    service_instance_name: "$hostname"
    endpoint_addr: http://skywalking-oap:12800
```

### 3. 在路由上启用

```bash
a6 route create -f - <<'EOF'
{
  "id": "traced-api",
  "uri": "/api/*",
  "plugins": {
    "skywalking": {
      "sample_ratio": 1
    }
  },
  "upstream": {
    "type": "roundrobin",
    "nodes": {
      "backend:8080": 1
    }
  }
}
EOF
```

### 4. 发送请求并查看追踪

```bash
curl http://127.0.0.1:9080/api/hello
```

在 SkyWalking UI 中查看追踪，地址为 `http://skywalking-ui:8080`。

## 常见模式

### 部分采样（生产环境）

```json
{
  "plugins": {
    "skywalking": {
      "sample_ratio": 0.1
    }
  }
}
```

追踪 10% 的请求。足以进行生产流量分析而不会产生
过多开销。

### 使用 skywalking-logger 进行追踪-日志关联

```json
{
  "plugins": {
    "skywalking": {
      "sample_ratio": 1
    },
    "skywalking-logger": {
      "endpoint_addr": "http://skywalking-oap:12800"
    }
  }
}
```

将访问日志与 SkyWalking UI 中的 trace IDs 关联，启用
从追踪到日志的点击穿透。

### 全局启用

```bash
curl "$(a6 context current -o json | jq -r .server)/apisix/admin/global_rules" \
  -X PUT \
  -H "X-API-KEY: $(a6 context current -o json | jq -r .api_key)" \
  -d '{
    "id": "skywalking-global",
    "plugins": {
      "skywalking": {
        "sample_ratio": 0.5
      }
    }
  }'
```

## Span 结构

该 plugin 为每个请求创建两个 spans：

- **entrySpan**: 从请求到达至响应完成 (component ID 6002)
- **exitSpan**: 从上游调用开始到接收到响应 (component ID 6002)

## 配置同步示例

```yaml
version: "1"
routes:
  - id: traced-api
    uri: /api/*
    plugins:
      skywalking:
        sample_ratio: 1
    upstream_id: my-upstream
upstreams:
  - id: my-upstream
    type: roundrobin
    nodes:
      "backend:8080": 1
```

## 故障排除

| 症状 | 原因 | 修复 |
|---------|-------|-----|
| SkyWalking UI 中没有追踪 | 错误的 `endpoint_addr` | 验证 OAP 在配置的地址可达 |
| 拓扑中缺少服务 | `service_name` 不匹配 | 检查 config.yaml 中的 `plugin_attr.skywalking.service_name` |
| 高开销 | 生产环境中 `sample_ratio: 1` | 对高流量路由降低至 0.01-0.1 |
| 追踪未关联 | 后端未插桩 | 在上游服务中安装 SkyWalking agent |
| Plugin 不工作 | 不在 plugins 列表中 | 确保 `skywalking` 在 config.yaml 的 `plugins` 数组中 |
