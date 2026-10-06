---
name: gitlab-repo-triage
description: GitLab 仓库推拉面分诊方法论：Git 智能协议探测状态码语义（401/403/404/500）、push 403 三分（分支保护/push rules/权限）、gitaly 与 hooks 面、git-data 与 LFS 盘水位、SSH 通道——固定顺序定位与判读基准（ask-ops 只读采集面）。
mode: on_demand
version: 0.1.0
maturity: experimental
requires_mcp:            # 采集依赖的工具面（装载期对账依据；旧 bianque-tools 二进制缺这些工具时应先重建）
  - server: ask-ops
    tools: [run_readonly_command, run_readonly_commands]
---

## 触发条件

- 症状关键词：git push 失败、push 被拒、gitlab 克隆失败、克隆慢、仓库打不开、大仓库超时、gitaly 异常
- 组合场景：CI 拉代码失败可能是本仓库面也可能是 runner 网络面（runner 侧走 cicd-ops）；push 后
  webhook 不生效走对应集成面；盘满的根因与容量处置联动 gitlab-instance-triage

## 数据来源（ask-ops 只读面）

- Git 智能协议探测（匿名即可，状态码即证据）：`curl -s -w '\n%{http_code}\n' 'http://127.0.0.1/<group>/<project>.git/info/refs?service=git-upload-pack'`
  ——401=私有仓库未认证（正常形态）、403=权限拒绝、404=仓库不存在或仓库盘异常、500=gitaly/hooks 崩；
- API 面（PAT）：`curl -s -H "PRIVATE-TOKEN: $GITLAB_TOKEN" 'http://127.0.0.1/api/v4/projects/<id>'`
  （repository_storage 字段=仓库所在存储分片；路径形态用 `<group>%2F<project>` URL 编码）、
  `'.../api/v4/projects/<id>/protected_branches'`、`'.../api/v4/projects/<id>/push_rules'`（实例/组级规则 403 时标注）；
- gitaly 面：`tail -100 /var/log/gitlab/gitaly/gitaly.log`；仓库实体（hashed 存储）：
  `ls -l /var/opt/gitlab/git-data/repositories/@hashed/<哈希前2>/<哈希>/<project>.git`、
  钩子面 `ls -l /var/opt/gitlab/git-data/repositories/@hashed/<哈希前2>/<哈希>/<project>.git/custom_hooks`；
- 存储水位：`df -h /var/opt/gitlab/git-data`、`du -sh /var/opt/gitlab/gitlab-rails/shared/lfs-objects`；
- SSH 通道：`ss -tlnp | grep ':22 '`（sshd 监听面）；gitlab-shell 的 authorized_keys 由 GitLab 托管——
  **手工改 authorized_keys 是事故源**，只观测不修改；
- **`gitlab-rake`（fsck/storage:verify 类核验）不在只读白名单**：作为审批后宿主侧动作进建议面，禁处方化。

## 分诊路径（固定顺序，先分通道再落位）

1. **先分通道**：HTTP(S) push/pull、SSH pull、Web/API 读仓——三通道其一异常其二是定位信息：
   仅 SSH 失败→sshd/gitlab-shell/authorized_keys 面（本机 `ss` 监听 + `/var/log/gitlab/gitlab-shell/gitlab-shell.log`）；
   仅 HTTP 失败→workhorse/反代/TLS 面（转 gitlab-instance-triage 502 分支）；全失败→gitaly/仓库盘（第 3 步）。
2. **push 403 三分**：protected branch（分支保护，API protected_branches 落证）、push rules
   （提交信息/文件名规则拒，push_rules 落证）、角色权限不足（maintainer 以下推保护分支）——
   以 API 与服务端日志证据落位，不猜；用户侧报错文案与 API 证据对齐后再下结论。
3. **push 500 → gitaly 与 hooks**：`tail -100 /var/log/gitlab/gitaly/gitaly.log` 找对应仓库路径的
   错误行；custom_hooks/post-receive 自定义钩子崩溃是高频根因（脚本缺陷或执行环境缺失）；
   存储盘满（`df -h /var/opt/gitlab/git-data`）是 500 的另一高频根因——git-data 满导致只读降级/写入失败。
4. **clone 慢/超时三分**：大仓库 + LFS（`du -sh lfs-objects` 体积证据）、存储盘 IO（`iostat -x 2 5` 采样带次数，
   持续高 util 落证）、gitaly 慢 RPC（gitaly.log 中慢查询行）——大仓库问题优先建议浅克隆/部分克隆
   （客户端侧建议），服务端不轻易动。
5. **404 面**：项目私有 + 匿名探测（正常形态）、hashed 路径缺失（仓库盘损坏或存储迁移中）、URL 大小写/
   路径错。hashed 路径对不上时先核对 API 的 repository_storage 与多存储分片，不臆测路径拼错。

## 判读基准

- 「推送失败」结论必须给出「通道 + 状态码/日志行」双要素；本机智能协议探测通而用户 push 不通=客户端侧/
  网络侧，不当服务端故障报；
- 401 对私有仓库是正常形态不是故障；403/404/500 的语义分开，混报是误诊；
- 盘满定级 CRIT：git-data 满影响推拉与仓库完整性，backups 满影响备份链（转 gitlab-lifecycle-ops），
  /var/log 满影响日志写盘——一个 df 三个面；
- hooks 面只观测存在性与报错行，不代改脚本内容；外部钩子逻辑归业务方。

## 输出要求

- 每个结论附状态码/日志关键行证据；变更类动作（改分支保护、关 push rule、扩容 git-data、
  存储迁移、gitlab-rake fsck 类核验）标注影响面并 requires_approval、处方为单条宿主侧命令进
  recommendation.steps；数据不足输出「需补充采集」清单（如完整 gitaly 错误行、push 报错原文、
  项目 API 元数据），不臆测。
