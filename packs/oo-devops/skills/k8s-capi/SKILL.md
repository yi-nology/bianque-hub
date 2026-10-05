---
name: k8s-capi
description: 用于配置、扩展和升级Kubernetes集群的Cluster API生命周期管理。在管理集群基础设施或多集群操作时使用。
mode: on_demand
version: 1.1.0
maturity: experimental
---

# Cluster API 生命周期管理

使用kubectl-mcp-server的Cluster API工具管理Kubernetes集群（11个工具）。

## 检查安装

```python
capi_detect_tool()
```

## 列出集群

```python
# 列出所有CAPI集群
capi_clusters_list_tool(namespace="default")

# 显示：
# - 集群名称
# - 阶段（Provisioning, Provisioned, Deleting）
# - 基础设施就绪
# - 控制平面就绪
```

## 获取集群详情

```python
capi_cluster_get_tool(name="my-cluster", namespace="default")

# 显示：
# - 规格（控制平面、基础设施）
# - 状态（阶段、条件）
# - 网络配置
```

## 获取集群Kubeconfig

```python
# 获取工作负载集群的kubeconfig
capi_cluster_kubeconfig_tool(name="my-cluster", namespace="default")

# 返回用于访问集群的kubeconfig
```

## Machines

### 列出Machines

```python
capi_machines_list_tool(namespace="default")

# 显示：
# - Machine名称
# - 集群
# - 阶段（Running, Provisioning, Failed）
# - Provider ID
# - 版本
```

### 获取Machine详情

```python
capi_machine_get_tool(name="my-cluster-md-0-xxx", namespace="default")
```

## Machine Deployments

### 列出Machine Deployments

```python
capi_machinedeployments_list_tool(namespace="default")

# 显示：
# - Deployment名称
# - 集群
# - 副本数（ready/total）
# - 版本
```

### 扩展Machine Deployment

```python
# 扩展工作节点
capi_machinedeployment_scale_tool(
    name="my-cluster-md-0",
    namespace="default",
    replicas=5
)
```

## Machine Sets

```python
capi_machinesets_list_tool(namespace="default")
```

## Machine Health Checks

```python
capi_machinehealthchecks_list_tool(namespace="default")

# Health checks自动修复不健康的machines
```

## Cluster Classes

```python
# 列出集群模板
capi_clusterclasses_list_tool(namespace="default")

# ClusterClasses定义可复用的集群配置
```

## 创建集群

```python
kubectl_apply(manifest="""
apiVersion: cluster.x-k8s.io/v1beta1
kind: Cluster
metadata:
  name: my-cluster
  namespace: default
spec:
  clusterNetwork:
    pods:
      cidrBlocks:
      - 198.51.100.0/16
    services:
      cidrBlocks:
      - 198.51.100.0/12
  controlPlaneRef:
    apiVersion: controlplane.cluster.x-k8s.io/v1beta1
    kind: KubeadmControlPlane
    name: my-cluster-control-plane
  infrastructureRef:
    apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
    kind: AWSCluster
    name: my-cluster
""")
```

## 创建Machine Deployment

```python
kubectl_apply(manifest="""
apiVersion: cluster.x-k8s.io/v1beta1
kind: MachineDeployment
metadata:
  name: my-cluster-md-0
  namespace: default
spec:
  clusterName: my-cluster
  replicas: 3
  selector:
    matchLabels:
      cluster.x-k8s.io/cluster-name: my-cluster
  template:
    spec:
      clusterName: my-cluster
      version: v1.28.0
      bootstrap:
        configRef:
          apiVersion: bootstrap.cluster.x-k8s.io/v1beta1
          kind: KubeadmConfigTemplate
          name: my-cluster-md-0
      infrastructureRef:
        apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
        kind: AWSMachineTemplate
        name: my-cluster-md-0
""")
```

## 集群生命周期工作流

### 配置新集群
```python
1. kubectl_apply(cluster_manifest)
2. capi_clusters_list_tool(namespace)  # 等待Provisioned
3. capi_cluster_kubeconfig_tool(name, namespace)  # 获取访问权限
```

### 扩展Workers
```python
1. capi_machinedeployments_list_tool(namespace)
2. capi_machinedeployment_scale_tool(name, namespace, replicas)
3. capi_machines_list_tool(namespace)  # 监控
```

### 升级集群
```python
1. # 更新控制平面版本
2. # 更新machine deployment版本
3. capi_machines_list_tool(namespace)  # 监控滚动更新
```

## 故障排除

### 集群卡在Provisioning

```python
1. capi_cluster_get_tool(name, namespace)  # 检查条件
2. capi_machines_list_tool(namespace)  # 检查machine状态
3. get_events(namespace)  # 检查事件
4. # 检查基础设施提供商日志
```

### Machine失败

```python
1. capi_machine_get_tool(name, namespace)
2. get_events(namespace)
3. # 常见问题：
   # - 云提供商配额
   # - 无效的machine template
   # - 网络问题
```

## 相关技能

- [k8s-multicluster](../k8s-multicluster/SKILL.md) - 多集群操作
- [k8s-operations](../k8s-operations/SKILL.md) - kubectl操作
