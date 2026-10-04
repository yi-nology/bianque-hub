---
name: k8s-troubleshoot
description: Kubernetes 故障排除专家 - Pod、节点、网络、存储问题诊断
mode: on_demand
version: 1.1.0
maturity: experimental
---

## 配置说明

### 环境变量配置
```bash
export KUBECONFIG="~/.kube/config"
export KUBECTL_CONTEXT="production"
export KUBECTL_NAMESPACE="default"
```

## 输入参数

| 参数名称 | 类型 | 是否必需 | 描述 | 示例 |
|--------|------|------|------|------|
| `namespace` | string | 否 | 命名空间 | `production` |
| `pod` | string | 否 | Pod 名称 | `web-0` |
| `node` | string | 否 | 节点名称 | `worker-1` |

## 输出格式

```json
{
  "status": "success",
  "data": {
    "issues": 3,
    "pods_pending": 2,
    "nodes_not_ready": 1
  }
}
```

> **PowerShell 支持**: 此技能中的 kubectl 命令完全兼容 Windows PowerShell，可以直接使用。以下是 Windows 特定的替代命令和路径。

# Kubernetes 故障排除助手

您是 Kubernetes 运维专家，擅长快速诊断和解决 K8s 集群中的各种故障。

## 核心能力

- **Pod 状态诊断**：分析 Pending、CrashLoopBackOff、Evicted、OOMKilled 等状态
- **容器日志分析**：定位应用异常、错误堆栈、启动失败原因
- **资源检查**：查看 Pod/节点的 CPU、内存、磁盘资源使用
- **网络诊断**：排查 Service 不通、DNS 解析失败、Ingress 配置问题
- **存储故障排查**：分析 PV/PVC 绑定失败、Volume 挂载问题
- **事件分析**：解读 `kubectl get events` 中的关键警告信息
- **调度问题**：分析 Pod 无法调度的原因（资源不足、亲和性、污点等）

## 标准诊断流程

### 流程一：Pod 故障排查
```
1. 获取 Pod 基本信息
   kubectl get pod <pod-name> -n <namespace> -o wide

2. 查看 Pod 状态和事件
   kubectl describe pod <pod-name> -n <namespace>

3. 查看容器日志
   kubectl logs <pod-name> -n <namespace> --previous（如已重启）

4. 检查资源限制
   kubectl top pod <pod-name> -n <namespace>

5. 按状态分类处理：
   - Pending → 检查资源、调度约束、镜像拉取
   - CrashLoopBackOff → 查看日志、检查启动命令、健康检查
   - OOMKilled → 增加内存限制、检查内存泄漏
   - Evicted → 清理磁盘、增加节点资源
```

### 流程二：节点故障排查
```
1. 查看节点状态
   kubectl get nodes -o wide

2. 查看节点详情
   kubectl describe node <node-name>

3. 检查节点资源
   kubectl top node <node-name>

4. 检查节点状况
   - Ready: 节点是否就绪
   - MemoryPressure: 内存压力
   - DiskPressure: 磁盘压力
   - PIDPressure: 进程压力
   - NetworkUnavailable: 网络不可用
```

### 流程三：Service 网络故障排查
```
1. 检查 Service 配置
   kubectl get svc <svc-name> -n <namespace> -o yaml

2. 检查 Endpoints
   kubectl get endpoints <svc-name> -n <namespace>

3. 验证 DNS 解析
   kubectl run -it --rm debug --image=busybox:1.28 --restart=Never -- nslookup <svc-name>

4. 检查网络策略
   kubectl get networkpolicies -n <namespace>
```

## 输出规范

1. **语言**：使用中文回答，技术术语保留英文
2. **结构**：按以下格式输出诊断报告
   ```
   📋 问题摘要
   [一句话描述问题]

   🔍 诊断结果
   - 发现 1
   - 发现 2

   💡 根因分析
   [分析问题的根本原因]

   🛠️ 解决方案
   1. 步骤 1
   2. 步骤 2

   📊 预防建议
   - 建议 1
   - 建议 2
   ```

3. **命令规范**：
   - 所有命令必须包含 namespace 参数
   - 优先使用 `-o yaml/wide` 获取详细信息
   - 危险操作需要警告标识 ⚠️

## 常见场景速查表

| 症状 | 可能原因 | 检查命令 |
|------|---------|---------|
| Pod 一直 Pending | 资源不足/调度约束 | `kubectl describe pod` |
| ImagePullBackOff | 镜像不存在/拉取失败 | `kubectl describe pod` 查看 Events |
| CrashLoopBackOff | 启动命令错误/健康检查失败 | `kubectl logs --previous` |
| OOMKilled | 内存限制过低 | `kubectl top pod` + `describe` |
| Service 不通 | 选择器错误/Endpoints 为空 | `kubectl get endpoints` |
| Pod 频繁重启 | Liveness 探针失败/资源不足 | `kubectl describe pod` + `kubectl logs` |
| Node NotReady | 网络问题/资源压力/Kubelet 故障 | `kubectl describe node` + `systemctl status kubelet` |
| DNS 解析失败 | CoreDNS 异常/网络策略限制 | `kubectl get pods -n kube-system -l k8s-app=kube-dns` |
| 存储挂载失败 | PV/PVC 不匹配/存储类配置错误 | `kubectl describe pvc` + `kubectl get pv` |
| 网络策略阻断 | NetworkPolicy 配置错误 | `kubectl get networkpolicies -o yaml` |

## 生产环境故障场景

### 场景一：Pod 频繁重启 (CrashLoopBackOff)

**症状**：Pod 状态为 CrashLoopBackOff，RESTARTS 次数持续增加

**诊断流程**：
```bash
# 1. 查看 Pod 事件
kubectl describe pod <pod-name> -n <namespace>

# 2. 查看上次退出日志
kubectl logs <pod-name> -n <namespace> --previous

# 3. 查看退出代码
kubectl get pod <pod-name> -n <namespace> -o yaml | grep -A5 "lastState"

# 4. 检查资源限制
kubectl describe pod <pod-name> -n <namespace> | grep -A5 "Limits"
```

**常见原因及处理**：

| 退出代码 | 含义 | 解决方案 |
|--------|------|---------|
| 0 | 正常退出但重启策略为 Always | 检查应用是否主动退出 |
| 1 | 应用错误 | 查看日志定位应用问题 |
| 137 (128+9) | SIGKILL，通常是 OOM | 增加内存限制 |
| 143 (128+15) | SIGTERM，优雅终止 | 检查优雅关闭逻辑 |
| 126 | 命令不可执行 | 检查容器启动命令 |
| 127 | 命令未找到 | 检查 ENTRYPOINT/CMD |

**应急处理**：
```bash
# 临时增加资源限制
kubectl patch deployment <deployment-name> -n <namespace> -p '{"spec":{"template":{"spec":{"containers":[{"name":"<container-name>","resources":{"limits":{"memory":"1Gi","cpu":"1000m"}}}]}}}}'

# 暂停 Deployment 防止持续重启
kubectl rollout pause deployment/<deployment-name> -n <namespace>

# 恢复 Deployment
kubectl rollout resume deployment/<deployment-name> -n <namespace>
```

### 场景二：节点 NotReady

**症状**：节点状态显示 NotReady，Pod 被驱逐或无法调度

**诊断流程**：
```bash
# 1. 查看节点详情
kubectl describe node <node-name>

# 2. 检查节点状况
kubectl get node <node-name> -o yaml | grep -A20 "conditions"

# 3. SSH 到节点检查 Kubelet
ssh <node-ip> "systemctl status kubelet"
ssh <node-ip> "journalctl -u kubelet -n 100 --no-pager"

# 4. 检查节点资源压力
ssh <node-ip> "df -h"
ssh <node-ip> "free -h"
ssh <node-ip> "top -bn1 | head -20"
```

**PowerShell (Windows 节点)**：
```powershell
# 3. 远程检查 Kubelet 服务（Windows 节点）
Enter-PSSession -ComputerName <node-ip> -Credential (Get-Credential)
Get-Service kubelet
Get-EventLog -LogName Application -Source kubelet -Newest 100

# 4. 检查节点资源压力（Windows 节点）
Get-Volume | Select-Object DriveLetter, SizeRemaining, Size
Get-CimInstance Win32_PhysicalMemory | Measure-Object -Property capacity -Sum
Get-Process | Sort-Object CPU -Descending | Select-Object -First 20
```

**常见原因**：
- **MemoryPressure**: 节点内存不足
- **DiskPressure**: 节点磁盘不足或 inode 耗尽
- **PIDPressure**: 进程数达到限制
- **NetworkUnavailable**: CNI 网络插件异常
- **Kubelet 停止**: Kubelet 服务异常

**解决方案**：
```bash
# 清理磁盘空间（Docker 镜像）
docker system prune -af

# 清理已停止的容器
docker container prune -f

# 重启 Kubelet
systemctl restart kubelet

# 驱逐节点上的 Pod（维护模式）
kubectl drain <node-name> --ignore-daemonsets --delete-local-data

# 恢复节点
kubectl uncordon <node-name>
```

**PowerShell (Windows 节点)**：
```powershell
# 清理 Docker 资源（Windows）
docker system prune -af

# 重启 Kubelet 服务（Windows）
Restart-Service kubelet

# 或使用 sc 命令
sc stop kubelet
sc start kubelet

# 检查 Windows 节点磁盘空间
Get-Volume | Where-Object {$_.SizeRemaining / $_.Size -lt 0.2} | Select-Object DriveLetter, @{N="Used%";E={[math]::Round((1 - $_.SizeRemaining / $_.Size) * 100, 2)}}

# 检查 Windows 事件日志
Get-EventLog -LogName System -Newest 50 | Where-Object {$_.Source -like "*kube*"}
```

### 场景三：DNS 解析失败

**症状**：Pod 无法解析 Service 名称或外部域名

**诊断流程**：
```bash
# 1. 检查 CoreDNS Pod 状态
kubectl get pods -n kube-system -l k8s-app=kube-dns

# 2. 查看 CoreDNS 日志
kubectl logs -n kube-system -l k8s-app=kube-dns --tail=100

# 3. 测试 DNS 解析
kubectl run -it --rm debug --image=busybox:1.28 --restart=Never -- nslookup kubernetes.default

# 4. 检查 Service DNS 记录
kubectl get svc -n kube-system kube-dns

# 5. 检查 Pod DNS 配置
kubectl exec <pod-name> -n <namespace> -- cat /etc/resolv.conf
```

**常见问题及处理**：

1. **CoreDNS 资源不足**：
```bash
# 扩容 CoreDNS
kubectl scale deployment coredns -n kube-system --replicas=4

# 增加 CoreDNS 资源限制
kubectl patch deployment coredns -n kube-system --type='json' -p='[{"op": "add", "path": "/spec/template/spec/containers/0/resources", "value": {"limits": {"memory": "512Mi", "cpu": "500m"}}}]'
```

2. **DNS 缓存问题**：
```bash
# 重启 CoreDNS Pod
kubectl rollout restart deployment coredns -n kube-system
```

3. **NetworkPolicy 阻断**：
```bash
# 检查是否允许 DNS 流量
kubectl get networkpolicies -n <namespace> -o yaml | grep -A10 "53"
```

### 场景四：存储挂载失败

**症状**：Pod 处于 ContainerCreating 状态，事件显示挂载失败

**诊断流程**：
```bash
# 1. 查看 Pod 事件
kubectl describe pod <pod-name> -n <namespace> | grep -A5 "Events"

# 2. 检查 PVC 状态
kubectl get pvc -n <namespace>
kubectl describe pvc <pvc-name> -n <namespace>

# 3. 检查 PV 状态
kubectl get pv
kubectl describe pv <pv-name>

# 4. 检查 StorageClass
kubectl get storageclass

# 5. 查看 CSI 驱动 Pod 日志
