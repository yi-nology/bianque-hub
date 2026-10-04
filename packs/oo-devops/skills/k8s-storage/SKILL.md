---
name: k8s-storage
description: Kubernetes 存储管理，包括 PVCs、storage classes 和 persistent volumes。在配置存储、管理 volumes 或排查存储问题时使用。
mode: on_demand
version: 1.1.0
maturity: experimental
---

# Kubernetes 存储

使用 kubectl-mcp-server 的存储工具管理 Kubernetes 存储。

## 何时应用

在以下情况下使用：
- 用户提到："PVC"、"PV"、"storage class"、"volume"、"disk"、"storage"
- 操作：配置存储、挂载 volumes、扩展存储
- 关键词："persist"、"data"、"backup storage"、"volume claim"

## 优先级规则

| 优先级 | 规则 | 影响 | 工具 |
|----------|------|--------|-------|
| 1 | PVC 前验证 storage class 存在 | 严重 | `get_storage_classes` |
| 2 | Pod 部署前检查 PVC 状态 | 高 | `describe_pvc` |
| 3 | 检查多 Pod 访问的 access modes | 中 | `get_pvcs` |
| 4 | 监控 PV 回收策略 | 低 | `get_persistent_volumes` |

## 快速参考

| 任务 | 工具 | 示例 |
|------|------|---------|
| 列出 PVCs | `get_pvcs` | `get_pvcs(namespace)` |
| PVC 详情 | `describe_pvc` | `describe_pvc(name, namespace)` |
| Storage classes | `get_storage_classes` | `get_storage_classes()` |
| 列出 PVs | `get_persistent_volumes` | `get_persistent_volumes()` |

## Persistent Volume Claims (PVCs)

```python
get_pvcs(namespace="default")

describe_pvc(name="my-pvc", namespace="default")

kubectl_apply(manifest="""
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: my-pvc
  namespace: default
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
  storageClassName: standard
""")

kubectl_delete(resource_type="pvc", name="my-pvc", namespace="default")
```

## Storage Classes

```python
get_storage_classes()

kubectl_apply(manifest="""
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: fast-ssd
provisioner: kubernetes.io/gce-pd
parameters:
  type: pd-ssd
reclaimPolicy: Delete
volumeBindingMode: WaitForFirstConsumer
""")
```

## Persistent Volumes

```python
get_persistent_volumes()

describe_persistent_volume(name="pv-001")
```

## Volume Snapshots

```python
kubectl_apply(manifest="""
apiVersion: snapshot.storage.k8s.io/v1
kind: VolumeSnapshot
metadata:
  name: my-snapshot
  namespace: default
spec:
  volumeSnapshotClassName: csi-snapclass
  source:
    persistentVolumeClaimName: my-pvc
""")

kubectl_apply(manifest="""
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: restored-pvc
spec:
  dataSource:
    name: my-snapshot
    kind: VolumeSnapshot
    apiGroup: snapshot.storage.k8s.io
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
""")
```

## 存储故障排除

```python
describe_pvc(name="my-pvc", namespace="default")

get_events(namespace="default")
describe_pod(name="my-pod", namespace="default")
```

## 相关技能

- [k8s-backup](../k8s-backup/SKILL.md) - Velero 备份/恢复
- [k8s-operations](../k8s-operations/SKILL.md) - kubectl apply/patch
