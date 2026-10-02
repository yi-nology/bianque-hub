---
name: n8n-lifecycle-ops
description: n8n 生命周期运维方法论：加密密钥不可丢失语义与轮换纪律、备份三件套完整性核对、升级路径与 2.x/3.x 破坏性变更基线、SQLite→Postgres 迁移口径、license 排障——固定顺序定位与判读基准（ask-ops 只读采集面）。
mode: on_demand
version: 0.1.0
maturity: experimental
requires_mcp:            # 采集依赖的工具面（装载期对账依据；旧 bianque-tools 二进制缺这些工具时应先重建）
  - server: ask-ops
    tools: [run_readonly_command, run_readonly_commands]
---

## 触发条件

- 症状关键词：n8n 升级、跨版本升级、备份、恢复、迁移数据库、加密密钥、credentials 打不开、license 报错、激活失败
- 组合场景：升级后的执行/webhook/队列异常分别先走对应分诊技能定故障位，本技能管**版本与数据生命周期本身**（要不要升、怎么备份、密钥与 license 面）

## 数据来源（ask-ops 只读面）

- 版本面：`curl -s http://127.0.0.1:5678/healthz`（部分版本回版本串）、`docker ps`
  （镜像 tag 是最可靠版本证据）、`docker inspect <容器>`（Image 字段）；
- 数据面：`.n8n` 目录结构 du（config/数据库文件/logs/binaryData 或 storage）、
  Postgres 形态的库体积走 db-ops 口径；
- 配置面：`docker inspect <容器>` 的 env 键名清单（`DB_TYPE`/`N8N_ENCRYPTION_KEY`
  只核有无与多进程一致性，**值永不进报告**）；
- 注意：`n8n export/import/license:info` 等 CLI 经 `docker exec` 执行，**不在只读白名单**，
  一律作为审批后的宿主侧动作进建议面，禁处方化。

## 方法论（固定顺序，按事由选段）

1. **加密密钥纪律**（一切备份/迁移/多进程问题的基石）：`N8N_ENCRYPTION_KEY` 缺省时
   首次启动随机生成并落在 `.n8n/config`——**丢失/换掉后库内 credentials 全部无法解密、
   不可恢复**（官方口径）。三条铁律：外置化显式设置（queue mode 下所有进程必须同值）；
   备份必含此密钥；轮换是单向操作且必须先全量备份。「credentials 全报错但工作流还在」
   是密钥被换过的典型指纹。
2. **备份完整性核对**（三件套缺一不可）：① `.n8n` 目录（含 config 内密钥；sqlite 形态
   还含数据库本身，**必须先停实例再拷贝**否则备份不一致）；② Postgres 用库自身工具
   （pg_dump 类，走 db-ops 口径）；③ 外部二进制存储与自定义节点目录。判读基准：CLI
   导出（workflow/credentials）**不含**用户/角色、执行历史、variables、实例设置——
   只有导出文件 ≠ 可完整恢复。恢复顺序固定：停实例 → 还原 `.n8n` → 还原 DB → 还原
   env → 启动。
3. **升级路径**：Docker 形态 `docker compose pull && docker compose down && docker compose up -d`
   （生产 pin 具体版本 tag，禁 latest 漂移）；schema 迁移在新版启动时**自动执行**
   （`/healthz/readiness` 就绪即迁移完成的语义），升级窗口长的实例先看迁移进度再判故障。
   破坏性变更基线：**2.0**——MySQL/MariaDB 移除（先迁移 DB 再升级）、task runners 转
   默认、`update:workflow` CLI 废弃为 `publish:workflow`；**3.0**——npm/npx 装法移除
   （只支持 Docker）、`~/.n8n/binaryData` 更名 `~/.n8n/storage`（并存=启动失败）、
   internal task runner 废弃（需 external sidecar，`n8nio/runners` 镜像版本必须与主实例
   一致）。跨大版本升级前先读 release notes 破坏性变更段并在测试实例演练（建议面）。
4. **SQLite→Postgres 迁移口径**（社区标准流程）：停实例备份 `.n8n` → 记录密钥 →
   `n8n export:workflow/credentials`（审批后宿主侧）→ 配 `DB_TYPE=postgresdb` +
   `DB_POSTGRESDB_*` 首次启动自动建表 → import 回灌 → **新实例必须同密钥**否则
   credentials 不可解。执行历史不随 export 走（最痛的取舍，提前告知用户）。
5. **license 排障**：报「license is not valid / key already in use」按序核——
   ① Docker 容器 machine-id 每次重启漂移（固定映射 machine-id 文件，建议面）；
   ② 自托管付费 license 需**每日访问 n8n license 服务器**，出网被拦会报无效（网络侧
   证据）；③ key 已绑其它实例（`n8n license:clear` 后重激活，审批动作）。功能边界
   先查清楚再报「功能不可用」：SSO/log streaming/multi-main/外部密钥/外部二进制存储是
   付费面，queue mode 与 metrics 免费版可用——别把 license 边界误诊为故障。

## 判读基准

- 升级类故障先问「跨了几个版本、破坏性变更段读过没」再定位——跨大版本升级的中途态
  是第一嫌疑；
- 「备份能不能救」的结论必须对照三件套完整性给缺口清单，禁只说「建议备份」；
- 密钥类结论只依据「有无/是否一致」，任何情况下不在报告引用密钥值；
- 变更窗口证据（升级/迁移/换密钥的时间线）优先于运行态归因。

## 输出要求

- 每个结论附版本 tag/目录结构/env 键名核对证据；一切升级、迁移、轮换、license 操作
  标注影响面与回退路径并 requires_approval；数据不足输出「需补充采集」清单（如当前
  版本与目标版本号、备份现状三件套清单、出网连通性证据），不臆测。
