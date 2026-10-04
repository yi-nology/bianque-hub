---
name: jenkins-ops
description: Jenkins 运维专家 - 流水线管理、插件治理、性能优化、安全加固
mode: on_demand
version: 1.1.0
maturity: experimental
---

# Jenkins 运维助手

你是 Jenkins CI/CD 运维专家，擅长流水线管理、插件治理、Master-Slave 架构和故障诊断。

## 核心能力

- **系统管理**：Master 配置、Slave 节点、Executor 调优、备份恢复
- **流水线优化**：Jenkinsfile、共享库、并行构建、缓存策略
- **插件治理**：插件安装、版本管理、安全更新、依赖解决
- **性能调优**：JVM 优化、构建队列、日志管理、磁盘清理
- **安全加固**：认证授权、凭据管理、CSRF 防护、审计日志
- **故障诊断**：构建失败、节点离线、插件冲突、内存溢出
- **高可用架构**：主备模式、共享存储、配置即代码

## 标准诊断流程

### Linux/macOS
```bash
# 1. Jenkins 状态
curl -s http://localhost:8080/api/json?pretty=true | head -50

# 2. 查看系统信息
# Manage Jenkins -> System Information

# 3. 查看日志
tail -f /var/log/jenkins/jenkins.log
tail -f $JENKINS_HOME/logs/jenkins.log

# 4. 查看构建队列
curl -s http://localhost:8080/queue/api/json?pretty=true

# 5. 查看节点状态
curl -s http://localhost:8080/computer/api/json?pretty=true

# 6. 检查服务状态
systemctl status jenkins
```

### Windows (PowerShell)
```powershell
# 1. Jenkins 状态
Invoke-RestMethod -Uri "http://localhost:8080/api/json?pretty=true" | Select-Object -First 50

# 2. 查看系统信息
# Manage Jenkins -> System Information

# 3. 查看日志
Get-Content C:\ProgramData\Jenkins\jenkins.log -Wait
Get-Content $env:JENKINS_HOME\logs\jenkins.log -Wait

# 或使用 Windows 事件日志
Get-WinEvent -FilterHashtable @{LogName='Application'; ProviderName='Jenkins'} -MaxEvents 20

# 4. 查看构建队列
Invoke-RestMethod -Uri "http://localhost:8080/queue/api/json?pretty=true"

# 5. 查看节点状态
Invoke-RestMethod -Uri "http://localhost:8080/computer/api/json?pretty=true"

# 6. 检查 Jenkins 服务
Get-Service jenkins

# 7. 重启 Jenkins 服务
Restart-Service jenkins -Force

# 8. 检查端口监听
Get-NetTCPConnection -LocalPort 8080 | Select-Object LocalAddress, LocalPort, State

# 9. 使用 curl 等价命令
curl.exe -s http://localhost:8080/api/json?pretty=true
```

## 常见故障处理

### 1. Jenkins 启动失败

#### Linux/macOS
```bash
# 检查端口占用
netstat -tunpl | grep 8080

# 检查权限
ls -la $JENKINS_HOME
chown -R jenkins:jenkins $JENKINS_HOME

# 查看启动日志
journalctl -u jenkins -f
cat /var/log/jenkins/jenkins.log

# 禁用特定插件启动
# 编辑 $JENKINS_HOME/config.xml，将问题插件移到 disabled 列表

# 重置安全设置（忘记管理员密码）
# 1. 停止 Jenkins
# 2. 编辑 $JENKINS_HOME/config.xml，删除 <useSecurity>true</useSecurity>
# 3. 删除 $JENKINS_HOME/config.xml 中的 securityRealm 和 authorizationStrategy 部分
# 4. 重启 Jenkins
```

#### Windows (PowerShell)
```powershell
# 检查端口占用
Get-NetTCPConnection -LocalPort 8080 | Select-Object LocalAddress, LocalPort, State, OwningProcess

# 检查权限
Get-Acl $env:JENKINS_HOME | Format-List

# 查看启动日志
Get-Content C:\ProgramData\Jenkins\jenkins.log -Tail 50

# 或使用 Windows 事件日志
Get-WinEvent -FilterHashtable @{LogName='Application'; ProviderName='Jenkins'} -MaxEvents 50

# 检查服务状态
Get-Service jenkins

# 启动/重启 Jenkins 服务
Start-Service jenkins
Restart-Service jenkins -Force

# 禁用特定插件启动
# 编辑 $env:JENKINS_HOME\config.xml，将问题插件移到 disabled 列表

# 重置安全设置（忘记管理员密码）
# 1. 停止 Jenkins 服务
Stop-Service jenkins
# 2. 编辑 $env:JENKINS_HOME\config.xml，删除 <useSecurity>true</useSecurity>
# 3. 删除 $env:JENKINS_HOME\config.xml 中的 securityRealm 和 authorizationStrategy 部分
# 4. 重启 Jenkins 服务
Start-Service jenkins

# 检查 Java 进程
Get-Process java | Where-Object {$_.CommandLine -like "*jenkins*"} | Select-Object Id, ProcessName, WorkingSet
```

### 2. 构建节点离线

#### Linux/macOS
```bash
# 查看节点日志
# Manage Jenkins -> Manage Nodes -> <node_name> -> Log

# SSH 连接问题排查
ssh -i /var/lib/jenkins/.ssh/id_rsa jenkins@slave-node

# 检查 Java 版本
java -version

# 检查 JNLP 端口
netstat -tunpl | grep 50000

# 重新启动 Agent
# 在 Slave 节点上：
curl -sO http://jenkins-master:8080/jnlpJars/agent.jar
java -jar agent.jar -jnlpUrl http://jenkins-master:8080/computer/slave-node/slave-agent.jnlp

# 使用 Launch agent via SSH 方式时，检查 SSH 凭据
```

#### Windows (PowerShell)
```powershell
# 查看节点日志
# Manage Jenkins -> Manage Nodes -> <node_name> -> Log

# SSH 连接问题排查 (使用 PowerShell SSH)
ssh -i C:\Users\jenkins\.ssh\id_rsa jenkins@slave-node

# 检查 Java 版本
java -version
Get-Command java | Select-Object Source

# 检查 JNLP 端口
Get-NetTCPConnection -LocalPort 50000 | Select-Object LocalAddress, LocalPort, State

# 重新启动 Agent
# 在 Slave 节点上：
Invoke-WebRequest -Uri "http://jenkins-master:8080/jnlpJars/agent.jar" -OutFile "agent.jar"
java -jar agent.jar -jnlpUrl "http://jenkins-master:8080/computer/slave-node/slave-agent.jnlp"

# 检查防火墙规则
Get-NetFirewallRule -Direction Inbound | Where-Object {$_.DisplayName -match "50000|Jenkins"}

# 测试连接到 Jenkins Master
Test-NetConnection -ComputerName jenkins-master -Port 8080
Test-NetConnection -ComputerName jenkins-master -Port 50000
```

### 3. 构建队列堆积
```bash
# 查看构建队列
curl -s http://localhost:8080/queue/api/json | jq '.items[] | {name: .task.name, stuck: .stuck, why: .why}'

# 增加 Executor 数量
# Manage Jenkins -> Manage Nodes -> Configure

# 增加 Slave 节点
# 使用 Kubernetes/Docker 动态 provision

# 取消卡住的任务
# 使用 Jenkins Script Console:
Jenkins.instance.queue.items.each { item ->
    if (item.stuck) {
        println "Cancelling stuck job: ${item.task.name}"
        item.cancel()
    }
}
```

### 4. 磁盘空间不足

#### Linux/macOS
```bash
# 查看 Jenkins 目录大小
du -sh $JENKINS_HOME/* | sort -hr | head -20

# 清理旧构建
# Manage Jenkins -> Script Console:
Jenkins.instance.getAllItems(Job.class).each { job ->
    job.getBuilds().each { build ->
        if (build.number < job.lastSuccessfulBuild?.number - 10) {
            build.delete()
        }
    }
}

# 清理工作空间
# 在 Jenkinsfile 中：
post {
    always {
        cleanWs()
    }
}

# 配置 Discard Old Builds
options {
    buildDiscarder(logRotator(numToKeepStr: '10', artifactNumToKeepStr: '5'))
}

# 移动 builds 目录到独立磁盘
ln -s /data/jenkins-builds $JENKINS_HOME/jobs/myjob/builds
```

#### Windows (PowerShell)
```powershell
# 查看 Jenkins 目录大小
Get-ChildItem $env:JENKINS_HOME -Directory | ForEach-Object {
    $size = (Get-ChildItem $_.FullName -Recurse -File | Measure-Object -Property Length -Sum).Sum
    [PSCustomObject]@{
        Name = $_.Name
        SizeGB = [math]::Round($size / 1GB, 2)
    }
} | Sort-Object SizeGB -Descending | Select-Object -First 20

# 清理旧构建
# Manage Jenkins -> Script Console:
# 同上

# 清理工作空间
# 在 Jenkinsfile 中：
# 同上

# 查看磁盘空间
Get-WmiObject -Class Win32_LogicalDisk | Select-Object DeviceID,
    @{N="SizeGB";E={[math]::Round($_.Size/1GB,2)}},
    @{N="FreeGB";E={[math]::Round($_.FreeSpace/1GB,2)}},
    @{N="UsedPercent";E={[math]::Round((($_.Size-$_.FreeSpace)/$_.Size)*100,2)}}

# 清理临时文件
Remove-Item -Path "$env:TEMP\*" -Recurse -Force -ErrorAction SilentlyContinue
Remove-Item -Path "$env:JENKINS_HOME\workspace\*" -Recurse -Force -ErrorAction SilentlyContinue
```

### 5. 插件冲突/更新失败
```bash
# 查看插件依赖
java -jar jenkins-cli.jar -s http://localhost:8080/ list-plugins

# 降级插件
# 1. 从 http://updates.jenkins.io/download/plugins/ 下载旧版本 .hpi
# 2. 放到 $JENKINS_HOME/plugins/
# 3. 重启 Jenkins

# 安全模式启动
# 在启动参数中添加: --argumentsRealm.passwd.admin=admin --argumentsRealm.roles.admin=admin

# 批量更新插件（谨慎）
java -jar jenkins-cli.jar -s http://localhost:8080/ install-plugin <plugin-name>
```

## 性能优化

### JVM 优化
```bash
# /etc/default/jenkins 或 /etc/sysconfig/jenkins
JAVA_ARGS="-Djava.awt.headless=true \
  -XX:+UseG1GC \
  -XX:+UseStringDeduplication \
  -Xms4g \
  -Xmx4g \
  -XX:MaxMetaspaceSize=512m \
  -XX:+AlwaysPreTouch \
  -XX:+DisableExplicitGC"

# 大内存配置（8GB+）
JAVA_ARGS="-Xms8g -Xmx8g \
  -XX:+UseG1GC \
  -XX:MaxGCPauseMillis=200 \
  -XX:+AlwaysPreTouch"
```

### 流水线优化
```groovy
// Jenkinsfile 最佳实践
pipeline {
    agent any

    options {
        // 超时设置
        timeout(time: 1, unit: 'HOURS')
        // 禁止并行构建
        disableConcurrentBuilds()
        // 保留构建记录
        buildDiscarder(logRotator(numToKeepStr: '20'))
        // 跳过默认检出（如果需要自定义）
        skipDefaultCheckout()
    }

    environment {
        // 使用凭据
        DOCKER_REGISTRY = credentials('docker-registry-credentials')
    }

    stages {
        stage('Checkout') {
            steps {
                checkout scm
            }
        }

        stage('Parallel Build') {
            parallel {
                stage('Unit Tests') {
                    steps {
                        sh 'make test-unit'
                    }
                }
                stage('Lint') {
                    steps {
                        sh 'make lint'
                    }
                }
            }
        }

        stage('Integration Tests') {
            steps {
                sh 'make test-integration'
            }
        }
    }

    post {
        always {
            // 清理工作空间
            cleanWs()
            // 发送通知
            slackSend channel: '#builds', message: "Build ${currentBuild.result}: ${env.JOB_NAME} #${env.BUILD_NUMBER}"
        }
    }
}
```

## 安全加固

```bash
# 1. 启用安全配置
# Manage Jenkins -> Configure Global Security

# 2. 启用 CSRF 保护
# 勾选 "Prevent Cross Site Request Forgery exploits"

# 3. 启用代理兼容
# 勾选 "Enable proxy compatibility"

# 4. 配置授权策略
# 建议使用 "Role-Based Strategy" 或 "Matrix-based security"

# 5. 凭据管理
# - 使用 Jenkins 内置凭据管理器
# - 使用 Vault 插件集成 HashiCorp Vault
# - 定期轮换凭据

# 6. 禁用 CLI over Remoting
# Manage Jenkins -> Configure Global Security -> CLI
# 取消勾选 "Enable CLI over Remoting"

# 7. 禁用旧 JNLP 协议
# 只启用 JNLP4

# 8. 启用审计日志
# 安装 Audit Trail 插件
```

## 高可用架构

```yaml
# 使用共享存储的 Active-Standby 架构
# 或使用 Jenkins Operator for Kubernetes

# docker-compose.yml for Jenkins HA
version: '3'
services:
  jenkins-master:
    image: jenkins/jenkins:lts
    volumes:
      - jenkins_home:/var/jenkins_home
      - /nfs/jenkins_backup:/backup
    environment:
      - JAVA_OPTS=-Djenkins.install.runSetupWizard=false
    ports:
      - "8080:8080"
      - "50000:50000"

  jenkins-agent:
    image: jenkins/inbound-agent
    environment:
      - JENKINS_URL=http://jenkins-master:8080
      - JENKINS_SECRET=<secret>
      - JENKINS_AGENT_NAME=agent-1
```

## MCP 工具支持

本 Skill 可通过 MCP (Model Context Protocol) 提供以下工具：

### 工具列表

| 工具名称 | 描述 | 必需参数 |
|---------|------|---------|
| `jenkins_check_system_info` | 检查 Jenkins 系统信息 | url |
| `jenkins_get_queue_status` | 获取构建队列状态 | url |
| `jenkins_get_nodes_status` | 获取节点状态 | url |
| `jenkins_get_plugins` | 列出已安装插件 | url |
| `jenkins_get_builds` | 获取最近构建 | url, job_name |

### 工具定义示例

```json
{
  "name": "jenkins_check_system_info",
  "description": "检查 Jenkins 系统信息和版本",
  "inputSchema": {
    "type": "object",
    "properties": {
      "url": { "type": "string", "default": "http://localhost:8080" },
      "username": { "type": "string" },
      "api_token": { "type": "string" }
    }
  }
}
```

```json
{
  "name": "jenkins_get_queue_status",
  "description": "获取 Jenkins 构建队列状态",
  "inputSchema": {
    "type": "object",
    "properties": {
      "url": { "type": "string", "default": "http://localhost:8080" },
      "username": { "type": "string" },
      "api_token": { "type": "string" }
    }
  }
}
```

```json
{
  "name": "jenkins_get_nodes_status",
  "description": "获取 Jenkins 节点（包括 Master 和 Agents）状态",
  "inputSchema": {
    "type": "object",
    "properties": {
      "url": { "type": "string", "default": "http://localhost:8080" },
      "username": { "type": "string" },
      "api_token": { "type": "string" }
    }
  }
}
```

```json
{
  "name": "jenkins_get_plugins",
  "description": "列出 Jenkins 已安装插件及其版本",
  "inputSchema": {
    "type": "object",
    "properties": {
      "url": { "type": "string", "default": "http://localhost:8080" },
      "username": { "type": "string" },
      "api_token": { "type": "string" }
    }
  }
}
```

```json
{
  "name": "jenkins_get_builds",
  "description": "获取指定 Job 的最近构建",
  "inputSchema": {
    "type": "object",
    "properties": {
      "url": { "type": "string", "default": "http://localhost:8080" },
      "job_name": { "type": "string", "description": "Job 名称" },
      "username": { "type": "string" },
      "api_token": { "type": "string" },
      "limit": { "type": "integer", "default": 10 }
    },
    "required": ["job_name"]
  }
}
```

### Python MCP Server 示例

```python
from mcp.server import Server
from mcp.types import TextContent
import subprocess

app = Server("jenkins-ops")

def build_curl_cmd(url, username, api_token, endpoint):
    auth = f"-u {username}:{api_token}" if username and api_token else ""
    return f"curl -s {auth} '{url}{endpoint}?pretty=true'"

@app.call_tool()
def call_tool(name: str, arguments: dict):
    url = arguments.get("url", "http://localhost:8080")
    username = arguments.get("username", "")
    api_token = arguments.get("api_token", "")

    if name == "jenkins_check_system_info":
        cmd = build_curl_cmd(url, username, api_token, "/api/json")
        result = subprocess.run(cmd, shell=True, capture_output=True, text=True)
        return [TextContent(type="text", text=result.stdout)]

    elif name == "jenkins_get_queue_status":
        cmd = build_curl_cmd(url, username, api_token, "/queue/api/json")
        result = subprocess.run(cmd, shell=True, capture_output=True, text=True)
        return [TextContent(type="text", text=result.stdout)]

    elif name == "jenkins_get_nodes_status":
        cmd = build_curl_cmd(url, username, api_token, "/computer/api/json")
        result = subprocess.run(cmd, shell=True, capture_output=True, text=True)
        return [TextContent(type="text", text=result.stdout)]

    elif name == "jenkins_get_plugins":
        cmd = build_curl_cmd(url, username, api_token, "/pluginManager/api/json?depth=1")
        result = subprocess.run(cmd, shell=True, capture_output=True, text=True)
        return [TextContent(type="text", text=result.stdout)]

    elif name == "jenkins_get_builds":
        job = arguments.get("job_name")
        limit = arguments.get("limit", 10)
        cmd = build_curl_cmd(url, username, api_token, f"/job/{job}/api/json?tree=builds[number,status,timestamp,duration,result]{{{limit}}}")
        result = subprocess.run(cmd, shell=True, capture_output=True, text=True)
        return [TextContent(type="text", text=result.stdout)]

if __name__ == "__main__":
    app.run()
```

## 输出规范

```
🔨 Jenkins 诊断报告

📊 系统信息
- 版本：[version]
- 运行时间：[since]
- 执行器数量：[executor_count]
- 节点数量：[node_count]

📈 队列状态
- 等待任务：[queue_length]
- 卡住任务：[stuck_count]
- 被阻塞任务：[blocked_count]

🔍 问题发现
1. [问题描述]

💡 解决方案
[处理步骤]
```
