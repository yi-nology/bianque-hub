---
name: volcengine-database-rds
description: 操作 Volcengine RDS 实例和数据库工作流。当用户需要配置指导、连接检查、性能故障排除或备份/恢复流程时使用。
mode: on_demand
version: 1.1.1
maturity: experimental
---

# volcengine-database-rds

以安全的操作顺序处理 RDS 任务：检查、验证、更改、确认。

## 执行清单

1. 确认引擎类型、区域和实例标识符。
2. 检查连接性、安全规则和参数组。
3. 执行目标操作（查询、调优、备份、恢复）。
4. 返回状态、指标和下一步建议操作。

## 安全规则

- 高风险更改前优先使用快照。
- 在应用更新前显示参数漂移。
- 将只读诊断与写操作分离。

## 参考

- `references/sources.md`


> ⚠️ **收割占位（stub）**：本技能收割时仅落了元信息，正文方法论待补全；
> 使用前以下游官方文档为准，勿按此空壳行事。同类完整件见同域技能（如 loki-ops）。
