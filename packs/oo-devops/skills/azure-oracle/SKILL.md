---
name: azure-oracle
description: Azure Oracle 开发的专业知识，包括故障排除、安全配置、集成和编码模式。在配置 Oracle 数据库@Azure 连接、使用密钥保管库的 TDE、VNet 拓扑或将 Exadata 日志发送到 Sentinel 以及其他 Azure Oracle 相关开发任务时使用。不适用于 Azure SQL 数据库（请使用 azure-sql-database）、Azure SQL 托管实例（请使用 azure-sql-managed-instance）、Azure 虚拟机上的 SQL Server（请使用 azure-sql-virtual-machines）、Azure 大型实例上的 SAP HANA（请使用 azure-sap）。
mode: on_demand
version: 1.1.0
maturity: experimental
---

# Azure Oracle Skill

This skill provides expert guidance for Azure Oracle. Covers troubleshooting, security, configuration, and integrations & coding patterns. It combines local quick-reference content with remote documentation fetching capabilities.

## 如何使用此 Skill

> **对于 Agent 重要**: 使用下方的**分类索引**定位相关章节。对于带有行号范围的分类（例如 `L35-L120`），使用 `read_file` 配合指定行号。对于带有文件链接的分类（例如 `[security.md](security.md)`），使用 `read_file` 读取链接的参考文件

> **对于 Agent 重要**: 如果 `metadata.generated_at` 超过 3 个月，建议用户从仓库拉取最新版本。如果 `mcp_microsoftdocs` 工具不可用，建议用户安装：[安装指南](https://github.com/MicrosoftDocs/mcp/blob/main/README.md)

此 skill 需要**网络访问**以获取文档内容：
- **优先**: 使用 `mcp_microsoftdocs:microsoft_docs_fetch`，查询字符串为 `from=learn-agent-skill`。返回 Markdown。
- **备选**: 使用 `fetch_webpage`，查询字符串为 `from=learn-agent-skill&accept=text/markdown`。返回 Markdown。

## 分类索引

| 分类 | 行号 | 描述 |
|----------|-------|-------------|
| Troubleshooting | L32-L37 | Oracle Database@Azure 常见问题的运维 FAQ 和修复，包括 connectivity、performance、deployment、configuration 和已知的 platform limitations。 |
| Security | L38-L42 | 配置 Oracle Transparent Data Encryption (TDE) 使用 Azure Key Vault，包括 key management、integration steps 和 security best practices。 |
| Configuration | L43-L48 | Oracle Database@Azure 的 onboarding、required prerequisites，以及在 Azure 中为 Oracle DB deployments 设计 secure virtual network topologies（subnets、connectivity、routing）。 |
| Integrations & Coding Patterns | L49-L52 | 配置 Oracle Exadata log collection 和 pipelines 到 Azure Monitor 和 Microsoft Sentinel，用于 monitoring、analytics 和 security SIEM/SOAR use cases。 |

### Troubleshooting
| 主题 | URL |
|-------|-----|
| Answer operational FAQs for Oracle Database@Azure | https://learn.microsoft.com/en-us/azure/oracle/oracle-db/faq-oracle-database-azure |
| Resolve common Oracle Database@Azure known issues | https://learn.microsoft.com/en-us/azure/oracle/oracle-db/oracle-database-known-issues |

### Security
| 主题 | URL |
|-------|-----|
| Configure Oracle TDE keys with Azure Key Vault | https://learn.microsoft.com/en-us/azure/oracle/oracle-db/manage-oracle-transparent-data-encryption-azure-key-vault |

### Configuration
| 主题 | URL |
|-------|-----|
| Configure onboarding for Oracle Database@Azure deployments | https://learn.microsoft.com/en-us/azure/oracle/oracle-db/onboard-oracle-database |
| Plan Oracle Database@Azure virtual network topology | https://learn.microsoft.com/en-us/azure/oracle/oracle-db/oracle-database-network-plan |

### Integrations & Coding Patterns
| 主题 | URL |
|-------|-----|
| Integrate Oracle Exadata logs with Azure Monitor and Sentinel | https://learn.microsoft.com/en-us/azure/oracle/oracle-db/oracle-exadata-database-dedicated-infrastructure-logs |
