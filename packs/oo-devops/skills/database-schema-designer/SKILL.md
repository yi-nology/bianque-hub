---
name: database-schema-designer
description: 数据库模式设计器
mode: on_demand
version: 1.1.0
maturity: experimental
---

# 数据库模式设计器

**级别：** POWERFUL
**类别：** 工程
**领域：** 数据架构 / 后端

---

## 概述

根据需求设计关系数据库模式，并生成迁移、TypeScript/Python 类型、种子数据、RLS 策略和索引。处理多租户、软删除、审计跟踪、版本控制和多态关联。

## 核心能力

- **模式设计** — 将需求规范化为表、关系、约束
- **迁移生成** — Drizzle、Prisma、TypeORM、Alembic
- **类型生成** — TypeScript 接口、Python 数据类/Pydantic 模型
- **RLS 策略** — 多租户应用的行级安全
- **索引策略** — 复合索引、部分索引、覆盖索引
- **种子数据** — 真实测试数据生成
- **ERD 生成** — 从模式生成 Mermaid 图表

---

## 何时使用

- 设计需要数据库表的新功能时
- 审查模式的性能或规范化问题时
- 向现有模式添加多租户时
- 从 Prisma 模式生成 TypeScript 类型时
- 规划破坏性变更的模式迁移时

---

## 模式设计流程

### 步骤 1：需求 → 实体

给定需求：
> "用户可以创建项目。每个项目有任务。任务可以有标签。任务可以分配给用户。我们需要完整的审计跟踪。"

提取实体：
```
User, Project, Task, Label, TaskLabel（连接表）, TaskAssignment, AuditLog
```

### 步骤 2：识别关系

```
User 1──* Project         （所有者）
Project 1──* Task
Task *──* Label            （通过 TaskLabel）
Task *──* User            （通过 TaskAssignment）
User 1──* AuditLog
```

### 步骤 3：添加横切关注点

- 多租户：向所有租户范围的表添加 `organization_id`
- 软删除：添加 `deleted_at TIMESTAMPTZ` 而不是硬删除
- 审计跟踪：添加 `created_by`、`updated_by`、`created_at`、`updated_at`
- 版本控制：添加 `version INTEGER` 用于乐观锁定

---

## 完整模式示例（任务管理 SaaS）
→ 详见 references/full-schema-examples.md

## 行级安全（RLS）策略

```sql
-- 启用 RLS
ALTER TABLE tasks ENABLE ROW LEVEL SECURITY;
ALTER TABLE projects ENABLE ROW LEVEL SECURITY;

-- 创建应用角色
CREATE ROLE app_user;

-- 用户只能看到其组织项目中的任务
CREATE POLICY tasks_org_isolation ON tasks
  FOR ALL TO app_user
  USING (
    project_id IN (
      SELECT p.id FROM projects p
      JOIN organization_members om ON om.organization_id = p.organization_id
      WHERE om.user_id = current_setting('app.current_user_id')::text
    )
  );

-- 软删除：永远不显示已删除的记录
CREATE POLICY tasks_no_deleted ON tasks
  FOR SELECT TO app_user
  USING (deleted_at IS NULL);

-- 只有任务创建者或管理员可以删除
CREATE POLICY tasks_delete_policy ON tasks
  FOR DELETE TO app_user
  USING (
    created_by_id = current_setting('app.current_user_id')::text
    OR EXISTS (
      SELECT 1 FROM organization_members om
      JOIN projects p ON p.organization_id = om.organization_id
      WHERE p.id = tasks.project_id
        AND om.user_id = current_setting('app.current_user_id')::text
        AND om.role IN ('owner', 'admin')
    )
  );

-- 设置用户上下文（在每个请求开始时调用）
SELECT set_config('app.current_user_id', $1, true);
```

---

## 种子数据生成

```typescript
// db/seed.ts
import { faker } from '@faker-js/faker'
import { db } from './client'
import { organizations, users, projects, tasks } from './schema'
import { createId } from '@paralleldrive/cuid2'
import { hashPassword } from '../src/lib/auth'

async function seed() {
  console.log('正在填充数据库...')

  // 创建组织
  const [org] = await db.insert(organizations).values({
    id: createId(),
    name: "acme-corp",
    slug: 'acme',
    plan: 'growth',
  }).returning()

  // 创建用户
  const adminUser = await db.insert(users).values({
    id: createId(),
    email: 'admin@acme.com',
    name: "alice-admin",
    passwordHash: await hashPassword('password123'),
  }).returning().then(r => r[0])

  // 创建项目
  const projectsData = Array.from({ length: 3 }, () => ({
    id: createId(),
    organizationId: org.id,
    ownerId: adminUser.id,
    name: "fakercompanycatchphrase"
    description: faker.lorem.paragraph(),
    status: 'active' as const,
  }))

  const createdProjects = await db.insert(projects).values(projectsData).returning()

  // 为每个项目创建任务
  for (const project of createdProjects) {
    const tasksData = Array.from({ length: faker.number.int({ min: 5, max: 20 }) }, (_, i) => ({
      id: createId(),
      projectId: project.id,
      title: faker.hacker.phrase(),
      description: faker.lorem.sentences(2),
      status: faker.helpers.arrayElement(['todo', 'in_progress', 'done'] as const),
      priority: faker.helpers.arrayElement(['low', 'medium', 'high'] as const),
      position: i * 1000,
      createdById: adminUser.id,
      updatedById: adminUser.id,
    }))

    await db.insert(tasks).values(tasksData)
  }

  console.log(`✅ 已填充：1 个组织，${projectsData.length} 个项目，任务`)
}

seed().catch(console.error).finally(() => process.exit(0))
```

---

## ERD 生成（Mermaid）

```
erDiagram
    Organization ||--o{ OrganizationMember : has
    Organization ||--o{ Project : owns
    User ||--o{ OrganizationMember : joins
    User ||--o{ Task : "created by"
    Project ||--o{ Task : contains
    Task ||--o{ TaskAssignment : has
    Task ||--o{ TaskLabel : has
    Task ||--o{ Comment : has
    Task ||--o{ Attachment : has
    Label ||--o{ TaskLabel : "applied to"
    User ||--o{ TaskAssignment : assigned

    Organization {
        string id PK
        string name
        string slug
        string plan
    }

    Task {
        string id PK
        string project_id FK
        string title
        string status
        string priority
        timestamp due_date
        timestamp deleted_at
        int version
    }
```

从 Prisma 生成：
```bash
npx prisma-erd-generator
# 或：npx @dbml/cli prisma2dbml -i schema.prisma | npx dbml-to-mermaid
```

---

## 常见陷阱

- **软删除无索引** — `WHERE deleted_at IS NULL` 没有索引 = 全表扫描
- **缺少复合索引** — `WHERE org_id = ? AND status = ?` 需要复合索引
- **可变代理键** — 永远不要使用 email 或 slug 作为主键；使用 UUID/CUID
- **无默认值非空** — 向现有表添加 NOT NULL 列需要默认值或迁移计划
- **无乐观锁定** — 并发更新会相互覆盖；添加 `version` 列
- **RLS 未测试** — 始终使用非超级用户角色测试 RLS

---

## 最佳实践

1. **到处使用时间戳** — 每个表都有 `created_at`、`updated_at`
2. **可审计数据的软删除** — 使用 `deleted_at` 而不是 DELETE
3. **合规性的审计日志** — 为受监管领域记录前后 JSON
4. **UUID 或 CUID 作为主键** — 避免顺序整数泄漏
5. **索引外键** — 每个 FK 列都应该有索引
6. **部分索引** — 使用 `WHERE deleted_at IS NULL` 进行仅活动查询
7. **RLS 优于应用级过滤** — 数据库强制执行租户隔离，而不仅仅是应用代码
