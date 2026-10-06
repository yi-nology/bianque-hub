---
name: zabbix
description: Zabbix 运维专家 - 监控配置、告警管理、模板定制、性能优化、分布式部署
mode: on_demand
version: 1.1.1
maturity: experimental
---

# Zabbix 自动化技能

## 环境变量配置

本技能使用以下环境变量来配置 Zabbix 服务器连接信息，请在执行前设置这些环境变量, 在执行过程中需要加载这些环境变量，直接使用环境变量中你配置的值去连接Zabbix服务：

| 环境变量名 | 描述 |
|--------|------|
| **ZABBIX_URL** | Zabbix 服务器 URL |
| **ZABBIX_USER** | 登录用户名 |
| **ZABBIX_PASSWORD** | 登录密码 |
| **ZABBIX_TOKEN** | API Token（可选，优先使用） |

> **注意**：在 OpenOcta 数字员工配置中，这些变量可通过 MCP 配置或环境变量注入。

**初始化代码（所有操作前直接使用）**：

```python
import os
from zabbix_utils import ZabbixAPI
import os

# 使用变量配置 Zabbix 连接
# 优先从环境变量读取，否则使用默认值（请替换为实际值）
ZABBIX_URL = os.environ.get("ZABBIX_URL", "{{ZABBIX_URL}}")
ZABBIX_USER = os.environ.get("ZABBIX_USER", "{{ZABBIX_USER}}")
ZABBIX_PASSWORD = os.environ.get("ZABBIX_PASSWORD", "{{ZABBIX_PASSWORD}}")

api = ZabbixAPI(url=ZABBIX_URL)
api.login(user=ZABBIX_USER, password=ZABBIX_PASSWORD)

# 验证连接
print(f"✅ 已连接到 Zabbix: {api.api_version()}")
```

> **安全提示**：生产环境建议使用 API Token 替代用户名/密码，并将敏感信息存储在环境变量或密钥管理系统中。

---

## 概述

本技能提供通过 API 和官方 Python 库 `zabbix_utils` 自动化 Zabbix 监控操作的指导。

## 快速开始

### 安装

```bash
pip install zabbix-utils --break-system-packages
```

### 认证

```python
import os
from zabbix_utils import ZabbixAPI
import os

# 从环境变量读取配置（推荐方式）
ZABBIX_URL = os.environ.get("ZABBIX_URL", "{{ZABBIX_URL}}")
ZABBIX_USER = os.environ.get("ZABBIX_USER", "{{ZABBIX_USER}}")
ZABBIX_PASSWORD = os.environ.get("ZABBIX_PASSWORD", "{{ZABBIX_PASSWORD}}")

# 选项 1: 使用用户名/密码认证
api = ZabbixAPI(url=ZABBIX_URL)
api.login(user=ZABBIX_USER, password=ZABBIX_PASSWORD)

# 选项 2: API token (Zabbix 5.4+, 如已配置推荐优先使用)
# ZABBIX_TOKEN = os.environ.get("ZABBIX_TOKEN", "{{ZABBIX_TOKEN}}")
# api = ZabbixAPI(url=ZABBIX_URL)
# api.login(token=ZABBIX_TOKEN)

# 验证连接
print(api.api_version())
```

### 环境变量模式（推荐用于生产环境）

```python
import os
import os
from zabbix_utils import ZabbixAPI

# 使用环境变量读取配置
ZABBIX_URL = os.environ.get("ZABBIX_URL")
ZABBIX_USER = os.environ.get("ZABBIX_USER")
ZABBIX_PASSWORD = os.environ.get("ZABBIX_PASSWORD")

if not all([ZABBIX_URL, ZABBIX_USER, ZABBIX_PASSWORD]):
    raise ValueError("请设置 ZABBIX_URL, ZABBIX_USER, ZABBIX_PASSWORD 环境变量")

api = ZabbixAPI(url=ZABBIX_URL)
api.login(user=ZABBIX_USER, password=ZABBIX_PASSWORD)
```

## 核心 API 方法

所有 API 遵循模式: `api.<object>.<method>()`，方法包括: `get`、`create`、`update`、`delete`。

### Host 操作

```python
# 获取 hosts
hosts = api.host.get(output=["hostid", "host", "name"],
                     selectInterfaces=["ip"])

# 创建 host
api.host.create(
    host="server01",
    groups=[{"groupid": "2"}],  # Linux 服务器
    interfaces=[{
        "type": 1,  # 1=agent, 2=SNMP, 3=IPMI, 4=JMX
        "main": 1,
        "useip": 1,
        "ip": "<主机IP>",
        "dns": "",
        "port": "10050"
    }],
    templates=[{"templateid": "10001"}]
)

# 更新 host
api.host.update(hostid="10084", status=0)  # 0=启用, 1=禁用

# 删除 host
api.host.delete("10084")
```

### Template 操作

```python
# 获取 templates
templates = api.template.get(output=["templateid", "host", "name"],
                             selectHosts=["hostid", "name"])

# 将 template 关联到 host
api.host.update(hostid="10084",
                templates=[{"templateid": "10001"}])

# 从 XML 导入 template
with open("template.xml") as f:
    api.configuration.import_(
        source=f.read(),
        format="xml",
        rules={
            "templates": {"createMissing": True, "updateExisting": True},
            "items": {"createMissing": True, "updateExisting": True},
            "triggers": {"createMissing": True, "updateExisting": True}
        }
    )
```

### Item 操作

```python
# 获取 items
items = api.item.get(hostids="10084",
                     output=["itemid", "name", "key_"],
                     search={"key_": "system.cpu"})

# 创建 item
api.item.create(
    name="CPU 负载",
    key_="system.cpu.load[percpu,avg1]",
    hostid="10084",
    type=0,  # 0=Zabbix agent
    value_type=0,  # 0=float, 3=integer, 4=text
    delay="30s",
    interfaceid="1"
)
```

### Trigger 操作

```python
# 获取 triggers
triggers = api.trigger.get(hostids="10084",
                          output=["triggerid", "description", "priority"],
                          selectFunctions="extend")

# 创建 trigger
api.trigger.create(
    description="{HOST.NAME} 的 CPU 过高",
    expression="last(/server01/system.cpu.load[percpu,avg1])>5",
    priority=3  # 0=未分类, 1=信息, 2=警告, 3=一般严重, 4=严重, 5=灾难
)
```

### Host Group 操作

```python
# 获取 groups
groups = api.hostgroup.get(output=["groupid", "name"])

# 创建 group
api.hostgroup.create(name="生产环境/服务器")

# 添加 hosts 到 group
api.hostgroup.massadd(groups=[{"groupid": "5"}],
                      hosts=[{"hostid": "10084"}])
```

### 维护窗口 (Maintenance Windows)

```python
import time

# 创建维护窗口
api.maintenance.create(
    name="服务器维护",
    active_since=int(time.time()),
    active_till=int(time.time()) + 3600,  # 1 小时
    hostids=["10084"],
    timeperiods=[{
        "timeperiod_type": 0,  # 一次性
        "period": 3600
    }]
)
```

### 事件和问题 (Events and Problems)

```python
# 获取当前问题 (Problems)
problems = api.problem.get(output=["eventid", "name", "severity"],
                          recent=True)

# 获取事件 (Events)
events = api.event.get(hostids="10084",
                       time_from=int(time.time()) - 86400,
                       output="extend")
```

### 历史数据 (History Data)

```python
# 获取历史数据 (value_type 必须与 item 的 value_type 匹配)
# 0=float, 1=character, 2=log, 3=integer, 4=text
history = api.history.get(
    itemids="28269",
    history=0,  # float
    time_from=int(time.time()) - 3600,
    output="extend",
    sortfield="clock",
    sortorder="DESC"
)
```

## Zabbix Sender (Trapper Items)

```python
from zabbix_utils import Sender

sender = Sender(server="ZABBIX_URL", port=10051)

# 发送单个值
response = sender.send_value("hostname", "trap.key", "value123")
print(response)  # {"processed": 1, "failed": 0, "total": 1}

# 发送多个值
from zabbix_utils import ItemValue
values = [
    ItemValue("host1", "key1", "value1"),
    ItemValue("host2", "key2", 42),
]
response = sender.send(values)
```

## Zabbix Getter (Agent 查询)

```python
from zabbix_utils import Getter

agent = Getter(host="ZABBIX_URL", port=10050)
response = agent.get("system.uname")
print(response.value)
```

## 常见模式

### 从 CSV 批量创建 Hosts

```python
import csv
import os
import os
from zabbix_utils import ZabbixAPI

# 从环境变量读取配置
ZABBIX_URL = os.environ.get("ZABBIX_URL", "{{ZABBIX_URL}}")
ZABBIX_USER = os.environ.get("ZABBIX_USER", "{{ZABBIX_USER}}")
ZABBIX_PASSWORD = os.environ.get("ZABBIX_PASSWORD", "{{ZABBIX_PASSWORD}}")

api = ZabbixAPI(url=ZABBIX_URL)
api.login(user=ZABBIX_USER, password=ZABBIX_PASSWORD)

with open("hosts.csv") as f:
    for row in csv.DictReader(f):
        try:
            api.host.create(
                host=row["hostname"],
                groups=[{"groupid": row["groupid"]}],
                interfaces=[{
                    "type": 1, "main": 1, "useip": 1,
                    "ip": row["ip"], "dns": "", "port": "10050"
                }]
            )
            print(f"已创建: {row['hostname']}")
        except Exception as e:
            print(f"创建失败 {row['hostname']}: {e}")
```

### 查找未关联 Template 的 Hosts

```python
# 获取所有 hosts
all_hosts = api.host.get(output=["hostid", "host"],
                         selectParentTemplates=["templateid"])

# 筛选未关联指定 template 的 hosts
template_id = "10001"
hosts_without = [h for h in all_hosts
                 if not any(t["templateid"] == template_id
                           for t in h.get("parentTemplates", []))]
```

### 按模式禁用 Triggers

```python
triggers = api.trigger.get(
    search={"description": "test"},
    output=["triggerid"]
)
for t in triggers:
    api.trigger.update(triggerid=t["triggerid"], status=1)  # 1=禁用
```

## Item 类型参考

| 类型 | 值 | 描述 |
|------|-------|-------------|
| Zabbix agent | 0 | 被动检查 |
| Zabbix trapper | 2 | 被动，通过 sender 推送数据 |
| Simple check | 3 | ICMP、TCP 等 |
| Zabbix internal | 5 | 服务端内部指标 |
| Zabbix agent (active) | 7 | Agent 主动发起 |
| HTTP agent | 19 | HTTP/REST API 监控 |
| Dependent item | 18 | 派生自主 item |
| Script | 21 | 自定义脚本 |

## 值类型参考 (Value Types)

| 类型 | 值 | 描述 |
|------|-------|-------------|
| Float | 0 | 数值（浮点） |
| Character | 1 | 字符字符串 |
| Log | 2 | 日志文件 |
| Unsigned | 3 | 数值（整数） |
| Text | 4 | 文本 |

## Trigger 严重性参考 (Severity)

| 严重性 | 值 | 颜色 |
|----------|-------|-------|
| 未分类 (Not classified) | 0 | 灰色 |
| 信息 (Information) | 1 | 浅蓝色 |
| 警告 (Warning) | 2 | 黄色 |
| 一般严重 (Average) | 3 | 橙色 |
| 严重 (High) | 4 | 浅红色 |
| 灾难 (Disaster) | 5 | 红色 |

## 错误处理

```python
import os
from zabbix_utils import ZabbixAPI
from zabbix_utils.exceptions import APIRequestError

try:
    api.host.create(host="duplicate_host", groups=[{"groupid": "2"}])
except APIRequestError as e:
    print(f"API 错误: {e.message}")
    print(f"错误码: {e.code}")
```

## 调试

```python
import logging
logging.basicConfig(level=logging.DEBUG)
# 现在所有 API 调用都会被记录到日志
```

## 脚本参考

查看 `scripts/` 目录获取即用的自动化脚本：

- `zabbix-bulk-hosts.py` - 从 CSV 批量管理 hosts
- `zabbix-maintenance.py` - 创建/管理维护窗口
- `zabbix-export.py` - 导出 hosts/templates 为 JSON/XML

## 最佳实践

1. **使用 API token** - 尽可能替代用户名/密码
2. **限制输出字段** - 始终指定 `output=["field1", "field2"]` 而非 `output="extend"`
3. **使用搜索/过滤** - 绝不要在 Python 中获取所有对象后再过滤
4. **处理分页** - 大数据集可能需要 `limit` 和 `offset`
5. **批量操作** - 使用 `massadd`、`massupdate` 进行批量修改
6. **错误处理** - 始终用 try/except 包裹 API 调用
7. **幂等性** - 创建前检查对象是否已存在

---

## 场景化操作指南

### 场景 1：主机在线状态检查与离线排查

**场景描述**：查看所有监控主机的在线状态，列出离线超过 10 分钟的主机并排查原因。

**实现代码**：

```python
import os
from zabbix_utils import ZabbixAPI
from datetime import datetime, timedelta
import time
import os

# 从环境变量读取配置
ZABBIX_URL = os.environ.get("ZABBIX_URL", "{{ZABBIX_URL}}")
ZABBIX_USER = os.environ.get("ZABBIX_USER", "{{ZABBIX_USER}}")
ZABBIX_PASSWORD = os.environ.get("ZABBIX_PASSWORD", "{{ZABBIX_PASSWORD}}")

api = ZabbixAPI(url=ZABBIX_URL)
api.login(user=ZABBIX_USER, password=ZABBIX_PASSWORD)

def check_host_status():
    """检查主机在线状态"""

    # 获取所有主机及其状态信息
    hosts = api.host.get(
        output=["hostid", "host", "name", "status", "available"],
        selectInterfaces=["ip", "dns", "port", "type"],
        selectItems=["itemid", "key_", "lastclock", "lastvalue"],
        filter={"status": "0"}  # 只查询启用的主机
    )

    # 获取当前问题（离线相关问题）
    problems = api.problem.get(
        output=["eventid", "name", "severity", "clock"],
        recent=True,
        search={"name": "unavailable"}
    )

    # 计算10分钟前的时间戳
    ten_minutes_ago = int(time.time()) - 600

    offline_hosts = []

    for host in hosts:
        # 判断主机可用性状态
        # available: 0=未知, 1=可用, 2=不可用
        available = host.get("available", "0")

        # 获取ZBX agent最后报告时间
        last_check = None
        for item in host.get("items", []):
            if item.get("key_") == "agent.ping" or "zbx" in item.get("key_", "").lower():
                last_check = int(item.get("lastclock", 0))
                break

        # 判断离线状态
        is_offline = False
        offline_duration = 0

        if available == "2":  # 明确标记为不可用
            is_offline = True
            if last_check:
                offline_duration = int(time.time()) - last_check
        elif last_check and last_check < ten_minutes_ago:
            is_offline = True
            offline_duration = int(time.time()) - last_check

        if is_offline and offline_duration > 600:  # 超过10分钟
            offline_hosts.append({
                "hostid": host["hostid"],
                "hostname": host["host"],
                "name": host["name"],
                "ip": host.get("interfaces", [{}])[0].get("ip", "N/A"),
                "available_status": available,
                "last_check": datetime.fromtimestamp(last_check).strftime("%Y-%m-%d %H:%M:%S") if last_check else "未知",
                "offline_minutes": offline_duration // 60,
                "offline_seconds": offline_duration
            })

    # 输出结果
    print("=" * 80)
    print(f"主机状态检查报告 - {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
    print("=" * 80)
    print(f"\n总主机数: {len(hosts)}")
    print(f"离线超过10分钟: {len(offline_hosts)} 台\n")

    if offline_hosts:
        print("-" * 80)
        print(f"{'主机名':<20} {'IP地址':<15} {'离线时长':<12} {'最后检查':<20}")
        print("-" * 80)

        for h in offline_hosts:
            print(f"{h['hostname']:<20} {h['ip']:<15} {h['offline_minutes']}分钟{'':<6} {h['last_check']:<20}")

        print("-" * 80)

        # 排查建议
        print("\n🔍 离线原因排查建议：")
        print("\n1. 网络连通性检查：")
        for h in offline_hosts:
            print(f"   ping {h['ip']} # 检查主机 {h['hostname']} 网络")

        print("\n2. Zabbix Agent服务检查：")
        for h in offline_hosts:
            print(f"   ssh root@{h['ip']} 'systemctl status zabbix-agent' # 检查 {h['hostname']}")

        print("\n3. 防火墙检查：")
        print("   # 检查 10050 端口是否开放")
        for h in offline_hosts:
            print(f"   telnet {h['ip']} 10050")

        print("\n4. 日志检查：")
        print("   # 在离线主机上执行")
        print("   tail -n 100 /var/log/zabbix/zabbix_agentd.log")
    else:
        print("✅ 所有主机均正常在线")

    return offline_hosts

# 执行检查
offline_hosts = check_host_status()
```

**快速排查命令**：

```bash
# 1. 批量检查主机ZBX可用性
# 在Zabbix服务器上执行
zabbix_get -s <host_ip> -k agent.ping

# 2. 检查Zabbix Server日志
tail -f /var/log/zabbix/zabbix_server.log | grep "cannot send list of active checks"

# 3. 检查主机名配置（必须一致）
# 在客户端执行
hostname
# 与Zabbix配置的Host name对比
```

---

### 场景 2：添加MySQL监控模板和CPU告警规则

**场景描述**：为 ZABBIX_URL 服务器添加 MySQL 监控模板，并配置 CPU 使用率超过 80% 时的告警规则。

**实现代码**：

```python
import os
from zabbix_utils import ZabbixAPI
import os

# 从环境变量读取配置
ZABBIX_URL = os.environ.get("ZABBIX_URL", "{{ZABBIX_URL}}")
ZABBIX_USER = os.environ.get("ZABBIX_USER", "{{ZABBIX_USER}}")
ZABBIX_PASSWORD = os.environ.get("ZABBIX_PASSWORD", "{{ZABBIX_PASSWORD}}")

api = ZabbixAPI(url=ZABBIX_URL)
api.login(user=ZABBIX_USER, password=ZABBIX_PASSWORD)

def add_mysql_monitoring(host_ip="{{ZABBIX_SERVER_IP}}"):
    """为指定服务器添加MySQL监控和CPU告警"""

    # 步骤1：查找或创建主机
    hosts = api.host.get(
        filter={"ip": host_ip},
        output=["hostid", "host", "name"]
    )

    if hosts:
        hostid = hosts[0]["hostid"]
        hostname = hosts[0]["name"]
        print(f"✅ 找到现有主机: {hostname} (ID: {hostid})")
    else:
        # 创建新主机
        print(f"⚠️ 未找到IP为 {host_ip} 的主机，需要先创建主机")
        return None

    # 步骤2：获取MySQL监控模板ID
    # 常用MySQL模板名称参考：
    # - "Template DB MySQL" (官方模板)
    # - "Template App MySQL"
    # - "MySQL by Zabbix agent"
    mysql_templates = api.template.get(
        output=["templateid", "host", "name"],
        search={"name": "MySQL"}
    )

    print(f"\n📋 找到的MySQL模板:")
    for t in mysql_templates:
        print(f"   - {t['name']} (ID: {t['templateid']})")

    # 选择合适的MySQL模板ID
    mysql_template_id = None
    for t in mysql_templates:
        if "MySQL" in t["name"] and "agent" in t["name"].lower():
            mysql_template_id = t["templateid"]
            break

    if not mysql_template_id and mysql_templates:
        mysql_template_id = mysql_templates[0]["templateid"]

    # 步骤3：关联MySQL模板到主机
    if mysql_template_id:
        # 获取当前已关联的模板
        current_host = api.host.get(
            hostids=hostid,
            selectParentTemplates=["templateid", "name"]
        )[0]

        existing_template_ids = [t["templateid"] for t in current_host.get("parentTemplates", [])]

        if mysql_template_id not in existing_template_ids:
            existing_template_ids.append(mysql_template_id)

            api.host.update(
                hostid=hostid,
                templates=[{"templateid": tid} for tid in existing_template_ids]
            )
            print(f"\n✅ 已关联MySQL模板 (ID: {mysql_template_id}) 到主机 {hostname}")
        else:
            print(f"\nℹ️ MySQL模板已关联，无需重复操作")
    else:
        print("\n⚠️ 未找到MySQL模板，请确认模板已导入")

    # 步骤4：创建CPU使用率超过80%的告警规则
    # 首先检查是否已存在类似的CPU告警
    existing_triggers = api.trigger.get(
        hostids=hostid,
        search={"description": "CPU"},
        output=["triggerid", "description", "expression"]
    )

    cpu_trigger_exists = any("80" in t.get("description", "") or "80" in t.get("expression", "")
                           for t in existing_triggers)

    if cpu_trigger_exists:
        print(f"\nℹ️ CPU使用率告警规则已存在")
    else:
        # 需要找到CPU item的key
        # 查找系统CPU相关的item
        cpu_items = api.item.get(
            hostids=hostid,
            search={"key_": "system.cpu"},
            output=["itemid", "name", "key_"]
        )

        # 创建CPU告警Trigger
        try:
            # 使用标准的CPU使用率监控项
            trigger_result = api.trigger.create(
                description=f"{hostname}: CPU使用率超过80%",
                expression=f"last(/{hostname}/system.cpu.util[,user])>80 or last(/{hostname}/system.cpu.util)>80",
                priority=3,  # 一般严重
                status=0     # 启用
            )
            print(f"\n✅ 已创建CPU告警规则: CPU使用率超过80% (Trigger ID: {trigger_result['triggerids'][0]})")
        except Exception as e:
            print(f"\n⚠️ 创建CPU告警规则失败: {e}")
            print("   可能需要先确保主机有CPU监控Item")

    # 步骤5：验证配置
    print("\n" + "=" * 60)
    print("配置验证")
    print("=" * 60)

    # 检查关联的模板
    updated_host = api.host.get(
        hostids=hostid,
        selectParentTemplates=["templateid", "name"]
    )[0]

    print(f"\n已关联模板:")
    for t in updated_host.get("parentTemplates", []):
        print(f"   - {t['name']}")

    # 检查告警规则
    triggers = api.trigger.get(
        hostids=hostid,
        output=["triggerid", "description", "priority", "status"],
        selectItems=["key_"]
    )

    print(f"\n已配置告警规则:")
    for t in triggers:
        priority_name = ["未分类", "信息", "警告", "一般严重", "严重", "灾难"][int(t["priority"])]
        status_str = "启用" if t["status"] == "0" else "禁用"
        print(f"   - {t['description']} [严重性: {priority_name}, 状态: {status_str}]")

    return hostid

# 执行配置
hostid = add_mysql_monitoring("ZABBIX_URL")

# 补充：MySQL监控前置条件检查清单
print("\n" + "=" * 60)
print("MySQL监控前置条件检查清单")
print("=" * 60)
checklist = """
在 ZABBIX_URL 服务器上需要完成以下配置：

1. 安装Zabbix Agent（如未安装）
   # rpm -Uvh https://repo.zabbix.com/zabbix/6.0/rhel/8/x86_64/zabbix-release-6.0-4.el8.noarch.rpm
   # dnf install zabbix-agent2

2. 配置MySQL监控用户
   CREATE USER 'zabbix'@'localhost' IDENTIFIED BY 'password';
   GRANT USAGE,REPLICATION CLIENT,PROCESS,SHOW DATABASES,SHOW VIEW ON *.* TO 'zabbix'@'localhost';
   FLUSH PRIVILEGES;

3. 创建MySQL配置文件
   # vim /var/lib/zabbix/.my.cnf
   [client]
   user=zabbix
   password=password
   host=localhost

4. 确保Zabbix Agent有权限读取
   # chown zabbix:zabbix /var/lib/zabbix/.my.cnf
   # chmod 600 /var/lib/zabbix/.my.cnf

5. 重启Zabbix Agent
   # systemctl restart zabbix-agent2
"""
print(checklist)
```

**常用模板ID参考**：

| 模板名称 | 用途 | 常用Key |
|---------|------|--------|
| Template DB MySQL | MySQL基础监控 | mysql.ping, mysql.status[
Uptime] |
| Template OS Linux | Linux系统监控 | system.cpu.util, vm.memory.size[
available] |
| Template App HTTP Service | HTTP服务监控 | net.tcp.service[http] |
| Template ICMP Ping | 网络连通性 | icmpping |

---

### 场景 3：生成告警统计报告

**场景描述**：生成过去24小时的Zabbix告警统计报告（按告警级别、主机、频率排序），并给出运维优化建议。

**实现代码**：

```python
import os
from zabbix_utils import ZabbixAPI
from datetime import datetime, timedelta
from collections import defaultdict
import json
import os

# 从环境变量读取配置
ZABBIX_URL = os.environ.get("ZABBIX_URL", "{{ZABBIX_URL}}")
ZABBIX_USER = os.environ.get("ZABBIX_USER", "{{ZABBIX_USER}}")
ZABBIX_PASSWORD = os.environ.get("ZABBIX_PASSWORD", "{{ZABBIX_PASSWORD}}")

api = ZabbixAPI(url=ZABBIX_URL)
api.login(user=ZABBIX_USER, password=ZABBIX_PASSWORD)

def generate_alert_report(hours=24):
    """生成告警统计报告"""

    # 计算时间范围
    time_from = int((datetime.now() - timedelta(hours=hours)).timestamp())
    time_till = int(datetime.now().timestamp())

    print(f"📊 Zabbix 告警统计报告")
    print(f"时间范围: {datetime.fromtimestamp(time_from)} 至 {datetime.now()}")
    print("=" * 80)

    # 1. 获取过去24小时的所有事件
    events = api.event.get(
        output=["eventid", "source", "object", "objectid", "clock", "value", "severity", "name"],
        time_from=time_from,
        time_till=time_till,
        source=0,  # 触发器事件
        value=1,   # 问题状态
        sortfield="clock",
        sortorder="DESC"
    )

    if not events:
        print("✅ 过去24小时内无告警事件")
        return

    print(f"\n总计告警事件: {len(events)} 个\n")

    # 2. 按告警级别统计
    severity_stats = defaultdict(int)
    severity_names = {
        "0": "未分类",
        "1": "信息",
        "2": "警告",
        "3": "一般严重",
        "4": "严重",
        "5": "灾难"
    }

    for event in events:
        severity = str(event.get("severity", "0"))
        severity_stats[severity] += 1

    print("-" * 80)
    print("按告警级别统计")
    print("-" * 80)
    print(f"{'级别':<10} {'数量':<8} {'占比':<10}")
    print("-" * 80)

    for severity in ["5", "4", "3", "2", "1", "0"]:
        count = severity_stats[severity]
        if count > 0:
            percentage = (count / len(events)) * 100
            print(f"{severity_names[severity]:<10} {count:<8} {percentage:.1f}%")

    print("-" * 80)

    # 3. 按主机统计
    host_stats = defaultdict(lambda: {"count": 0, "severities": defaultdict(int), "triggers": set()})

    # 获取所有相关触发器信息
    trigger_ids = list(set(e["objectid"] for e in events))
    triggers_info = {}

    # 分批获取触发器信息（避免请求过大）
    for i in range(0, len(trigger_ids), 50):
        batch = trigger_ids[i:i+50]
        trigger_batch = api.trigger.get(
            triggerids=batch,
            output=["triggerid", "description", "priority", "hosts"],
            selectHosts=["hostid", "host", "name"]
        )
        for t in trigger_batch:
            triggers_info[t["triggerid"]] = t

    for event in events:
        trigger_id = event["objectid"]
        trigger = triggers_info.get(trigger_id, {})
        hosts = trigger.get("hosts", [])

        for host in hosts:
            hostid = host["hostid"]
            hostname = host.get("name", host.get("host", "Unknown"))
            host_stats[hostname]["count"] += 1
            host_stats[hostname]["severities"][str(event.get("severity", "0"))] += 1
            host_stats[hostname]["triggers"].add(trigger.get("description", "Unknown"))

    # 按告警数量排序
    sorted_hosts = sorted(host_stats.items(), key=lambda x: x[1]["count"], reverse=True)

    print("\n" + "-" * 80)
    print("按主机统计（Top 10）")
    print("-" * 80)
    print(f"{'主机名':<25} {'告警数':<8} {'严重':<6} {'一般':<6} {'警告':<6}")
    print("-" * 80)

    for hostname, stats in sorted_hosts[:10]:
        severities = stats["severities"]
        print(f"{hostname:<25} {stats['count']:<8} {severities.get('4',0)+severities.get('5',0):<6} {severities.get('3',0):<6} {severities.get('2',0):<6}")

    print("-" * 80)

    # 4. 按触发器频率统计
    trigger_stats = defaultdict(int)
    trigger_severity = {}

    for event in events:
        trigger_id = event["objectid"]
        trigger = triggers_info.get(trigger_id, {})
        desc = trigger.get("description", "Unknown")
        trigger_stats[desc] += 1
        trigger_severity[desc] = event.get("severity", "0")

    # 按频率排序
    sorted_triggers = sorted(trigger_stats.items(), key=lambda x: x[1], reverse=True)

    print("\n" + "-" * 80)
    print("高频告警 Top 10")
    print("-" * 80)
    print(f"{'触发器':<50} {'次数':<6} {'级别':<8}")
    print("-" * 80)

    for desc, count in sorted_triggers[:10]:
        severity = trigger_severity.get(desc, "0")
        severity_name = severity_names.get(severity, "未知")
        short_desc = desc[:47] + "..." if len(desc) > 50 else desc
        print(f"{short_desc:<50} {count:<6} {severity_name:<8}")

    print("-" * 80)

    # 5. 生成运维优化建议
    print("\n" + "=" * 80)
    print("💡 运维优化建议")
    print("=" * 80)

    suggestions = []

    # 基于严重告警的建议
    if severity_stats["5"] > 0 or severity_stats["4"] > 0:
        suggestions.append({
            "priority": "高",
            "title": "严重告警处理",
            "content": f"过去24小时有 {severity_stats['5'] + severity_stats['4']} 个严重/灾难级告警，建议立即排查根因",
            "action": "检查系统稳定性，可能需要升级硬件或优化应用"
        })

    # 基于高频告警的建议
    if sorted_triggers and sorted_triggers[0][1] > 10:
        top_trigger = sorted_triggers[0]
        suggestions.append({
            "priority": "高",
            "title": "频发告警优化",
            "content": f"告警 '{top_trigger[0][:40]}...' 在过去24小时触发 {top_trigger[1]} 次",
            "action": "建议调整阈值或从根本上解决问题，避免告警疲劳"
        })

    # 基于主机告警分布的建议
    if sorted_hosts and sorted_hosts[0][1]["count"] > len(events) * 0.3:
        problem_host = sorted_hosts[0]
        suggestions.append({
            "priority": "中",
            "title": "重点主机关注",
            "content": f"主机 '{problem_host[0]}' 产生 {problem_host[1]['count']} 个告警，占总量的 {(problem_host[1]['count']/len(events)*100):.1f}%",
            "action": "建议对该主机进行深度健康检查"
        })

    # 基于警告级别过多的建议
    if severity_stats["2"] > len(events) * 0.5:
        suggestions.append({
            "priority": "中",
            "title": "告警级别调整",
            "content": f"警告级别告警占比 {(severity_stats['2']/len(events)*100):.1f}%，可能存在阈值设置过于敏感",
            "action": "建议审查并适当提高部分监控项的告警阈值"
        })

    # 通用建议
    suggestions.append({
        "priority": "低",
        "title": "监控覆盖检查",
        "content": "定期检查是否有新增业务未纳入监控",
        "action": "建立监控配置审核流程，确保新业务上线时同步配置监控"
    })

    for i, sug in enumerate(suggestions, 1):
        print(f"\n{i}. [优先级: {sug['priority']}] {sug['title']}")
        print(f"   问题: {sug['content']}")
        print(f"   建议: {sug['action']}")

    # 6. 导出JSON报告
    report = {
        "generated_at": datetime.now().isoformat(),
        "time_range": {"from": time_from, "till": time_till},
        "total_events": len(events),
        "severity_distribution": dict(severity_stats),
        "top_hosts": [{"host": h, "count": s["count"]} for h, s in sorted_hosts[:10]],
        "top_triggers": [{"trigger": t, "count": c} for t, c in sorted_triggers[:10]],
        "suggestions": suggestions
    }

    # 保存报告
    report_filename = f"zabbix_alert_report_{datetime.now().strftime('%Y%m%d_%H%M%S')}.json"
    with open(report_filename, "w", encoding="utf-8") as f:
        json.dump(report, f, ensure_ascii=False, indent=2)

    print(f"\n✅ 详细报告已保存至: {report_filename}")

    return report

# 生成报告
report = generate_alert_report(24)
```

---

### 场景 4：查看已配置的告警规则

**场景描述**：查看Zabbix中已经配置的所有告警规则（触发器）。

**实现代码**：

```python
import os
from zabbix_utils import ZabbixAPI

api = ZabbixAPI(url="ZABBIX_URL")
api.login(user=os.environ["ZABBIX_USER"], password=os.environ["ZABBIX_PASSWORD"])   # 禁硬编码默认口令（Admin/zabbix 是出厂默认，示例照抄=弱凭证）

def list_all_triggers(hostgroup_name=None, hostname=None, severity=None, status=None):
    """
    查看所有告警规则（触发器）

    参数:
        hostgroup_name: 按主机组筛选，如 "Linux servers"
        hostname: 按主机名筛选，如 "127.0.0.1"
        severity: 按严重级别筛选，如 "4" (严重)
        status: 按状态筛选，如 "0" (启用) 或 "1" (禁用)
    """

    # 构建筛选条件
    params = {
        "output": ["triggerid", "description", "priority", "status", "value", "lastchange", "comments"],
        "selectHosts": ["hostid", "host", "name"],
        "selectItems": ["itemid", "name", "key_", "lastvalue"],
        "expandDescription": True,
        "sortfield": "priority",
        "sortorder": "DESC",
        "limit": 1000
    }

    # 按主机组筛选
    if hostgroup_name:
        hostgroups = api.hostgroup.get(
            filter={"name": hostgroup_name},
            output=["groupid"]
        )
        if hostgroups:
            params["groupids"] = [hg["groupid"] for hg in hostgroups]

    # 按主机名筛选
    if hostname:
        hosts = api.host.get(
            filter={"host": hostname},
            output=["hostid"]
        )
        if hosts:
            params["hostids"] = [h["hostid"] for h in hosts]

    # 按严重级别筛选
    if severity is not None:
        params["filter"] = params.get("filter", {})
        params["filter"]["priority"] = severity

    # 按状态筛选
    if status is not None:
        params["filter"] = params.get("filter", {})
        params["filter"]["status"] = status

    # 获取触发器
    triggers = api.trigger.get(**params)

    # 严重性名称映射
    severity_names = {
        "0": ("未分类", "⚪"),
        "1": ("信息", "🔵"),
        "2": ("警告", "🟡"),
        "3": ("一般严重", "🟠"),
        "4": ("严重", "🔴"),
        "5": ("灾难", "⛔")
    }

    # 状态映射
    status_names = {"0": "启用", "1": "禁用"}
    trigger_status = {"0": "正常", "1": "告警中"}

    print("=" * 100)
    print(f"Zabbix 告警规则列表 - 总计 {len(triggers)} 条")
    print("=" * 100)

    # 按严重性分组显示
    for sev in ["5", "4", "3", "2", "1", "0"]:
        sev_triggers = [t for t in triggers if t.get("priority") == sev]
        if not sev_triggers:
            continue

        sev_name, sev_icon = severity_names[sev]
        print(f"\n{sev_icon} {sev_name}级别 ({len(sev_triggers)} 条)")
        print("-" * 100)

        for t in sev_triggers:
            hosts = t.get("hosts", [])
            host_names = ", ".join([h.get("name", h.get("host", "Unknown")) for h in hosts]) if hosts else "N/A"

            status_str = status_names.get(t.get("status", "0"), "未知")
            value_str = trigger_status.get(t.get("value", "0"), "未知")

            print(f"\n  ID: {t['triggerid']}")
            print(f"  名称: {t['description']}")
            print(f"  关联主机: {host_names}")
            print(f"  状态: {status_str} | 当前: {value_str}")

            # 显示关联的监控项
            items = t.get("items", [])
            if items:
                item_keys = ", ".join([i.get("key_", "N/A") for i in items[:3]])
                print(f"  监控项: {item_keys}")

    print("\n" + "=" * 100)

    # 统计信息
    print("\n📊 统计摘要:")
    print(f"  总计: {len(triggers)} 条触发器")
    print(f"  启用: {sum(1 for t in triggers if t.get('status') == '0')} 条")
    print(f"  禁用: {sum(1 for t in triggers if t.get('status') == '1')} 条")
    print(f"  当前告警中: {sum(1 for t in triggers if t.get('value') == '1')} 条")

    # 按严重性统计
    print("\n  按严重性分布:")
    for sev in ["5", "4", "3", "2", "1", "0"]:
        count = len([t for t in triggers if t.get("priority") == sev])
        if count > 0:
            sev_name, sev_icon = severity_names[sev]
            print(f"    {sev_icon} {sev_name}: {count} 条")

    return triggers

# 使用示例
print("\n【示例1】查看所有启用的告警规则")
triggers = list_all_triggers(status="0")

print("\n\n【示例2】查看特定主机的告警规则")
triggers = list_all_triggers(hostname="ZABBIX_URL")

print("\n\n【示例3】查看严重级别的告警规则")
triggers = list_all_triggers(severity="4")

print("\n\n【示例4】按主机组查看告警规则")
triggers = list_all_triggers(hostgroup_name="Linux servers")

# 快速查询命令
quick_queries = """

📌 快速查询命令参考

# 查看特定主机的所有触发器
api.trigger.get(
    hostids=["hostid_here"],
    output=["description", "priority", "status"]
)

# 查看当前处于告警状态的触发器
api.trigger.get(
    output=["description", "priority"],
    filter={"value": "1"}  # 1表示当前处于问题状态
)

# 查看禁用的触发器
api.trigger.get(
    output=["description", "priority"],
    filter={"status": "1"}
)

# 搜索特定关键字的触发器
api.trigger.get(
    output=["description", "priority"],
    search={"description": "CPU"}
)
"""
print(quick_queries)
```

**命令行快速查看方式**：

```bash
# 1. 使用 zabbix_get 查看特定主机的触发器
# 需要先在Zabbix服务器上配置

# 2. 直接在Zabbix服务器数据库查询（高级）
mysql -uzabbix -p zabbix -e "
SELECT
    h.host,
    t.description,
    CASE t.priority
        WHEN 0 THEN '未分类'
        WHEN 1 THEN '信息'
        WHEN 2 THEN '警告'
        WHEN 3 THEN '一般严重'
        WHEN 4 THEN '严重'
        WHEN 5 THEN '灾难'
    END as severity,
    CASE t.status
        WHEN 0 THEN '启用'
        WHEN 1 THEN '禁用'
    END as status
FROM triggers t
JOIN functions f ON t.triggerid = f.triggerid
JOIN items i ON f.itemid = i.itemid
JOIN hosts h ON i.hostid = h.hostid
WHERE t.flags = 0
ORDER BY t.priority DESC, h.host
LIMIT 50;
"
```
