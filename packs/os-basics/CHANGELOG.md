# CHANGELOG

## 1.8.1 (2026-10-10)

- **k8s-health 补 k8s 专属镜像拉取词面（bianque 批次二百六十四追加一，用户纠正）**：
  1.8.0 评审曾以「镜像拉取失败已存 docker-ops@P3、加入 k8s-health@P2 会跨包抢赢」为由
  剔除该词——用户指出 **docker-ops 与 k8s-health 是两个独立域**（Docker Engine 镜像拉取
  vs K8s Pod 镜像拉取是不同场景，docker-analyst route_desc 自己就声明「K8s 编排容器
  pod/crashloop/驱逐 走 k8s-ops」）。正解不是剔除、也不是塞裸词（裸词「镜像拉取失败」
  进 P2 反会劫持 docker 的 P3 裸词 symptom，validator 跨层重复告警），而是给 k8s-health
  **自己的 k8s 专属复合词**：route_keywords 补 `ErrImagePull`/`ImagePullBackOff`（k8s
  规范状态串，与基线 crashloop/OOMKilled 同模式）+ `pod镜像拉取失败`（蒸馏「镜像拉取失败」
  8 hits 的 pod 复合形态）。三词均 k8s 专属、不与 docker-ops 的 docker 前缀复合词
  （docker镜像拉取失败/docker镜像拉不下来）或裸词撞，各域独立各管各的。
- **为何进 route_keywords 而非 symptoms**：k8s-health 是 P2，symptoms 仅 P6 专家经确定性
  兜底层生效（1.8.0 架构校准），故 k8s 镜像拉取词须进 route_keywords（P2 关键词层）才
  确定性路由；ErrImagePull/ImagePullBackOff 为长拉丁串不触发 pack-lint 裸泛词警告。
- index.json 经 index-gen 再生（os-basics 1.8.1），validator os-basics scope 0 error，
  README 版本行同步。

## 1.8.0 (2026-10-10)

- **运维场景识别词面增强（bianque 批次二百六十四，红帽知识库蒸馏）**：从红帽公开文档
  语料（HuggingFace `mtpti5iD/redhat-docs_dataset`，55,741 条 docs.redhat.com 结构化
  文档，CC-BY-SA-4.0）经 DGX Spark 容器化三阶段蒸馏（确定性筛选 4,950 候选 →
  TensorFold Qwen3.8-Flash-Next 判域提炼中文症状词，bad_json=0/err=0 → 聚合 9,860
  独立词），**人工保守评审**收敛为 148 词面增量 + 9 条消歧规则并入六域专家：
  - io-analysis +23（症状 磁盘空间不足/挂载失败/文件系统损坏/文件系统只读/存储不可用…；
    关键词 GFS2/XFS/ext4/NFS/VDO/multipath/iSCSI/CephFS/OSD…）
  - memory-analysis +7（内存占用高/内存访问慢/内存超分…；KSM/TLB miss）
  - network-analysis +29（端口不通/路由不通/防火墙规则不生效/负载均衡不生效/DNS解析失败…；
    nmcli/NetworkManager/VLAN/firewalld/MTU/DHCP/Keepalived/iptables/bonding/VRRP…）
  - perf-tuning +25（CPU占用高/系统卡顿/响应慢/性能下降/CPU争抢/性能抖动…；
    PCP/Performance Co-Pilot/tuned/SystemTap/cgroups/numactl/turbostat…）
  - security-assistant +40（认证失败/登录失败/无法访问文件/证书不受信任/SELinux拦截…；
    IdM/SSSD/Kerberos/LDAP/SCAP/RBAC/CVE/oscap/LUKS/keytab/Clevis/Tang/PKCS#11…）
  - k8s-health +13（pod状态异常/节点失联/网络策略不生效/pod一直重启…；etcd/StorageClass/
    NetworkPolicy/ConfigMap/StatefulSet——遵守本专家「禁裸泛词只收复合限定词」纪律，
    剔除 CRD/CSI/cordon/drain 裸缩写裸动词；镜像拉取失败已存 docker-ops@P3 剔除防跨包劫持）
  - system-patrol +11（时间不同步/时钟漂移/服务状态异常/驱动加载失败；kdump/vmcore/
    crashkernel/NTP/chronyd/sysctl/GRUB）
- **9 条跨域消歧规则**（disambiguation.yaml）：内存不足/iowait/磁盘IO慢/NUMA/权限不足/
  权限被拒绝/集群状态异常/firewalld/PVC——按上下文分流（如「内存不足+性能」归 perf-tuning、
  默认归 memory-analysis），防批次二百六十二式词面劫持。
- **评审纪律（吸取本仓两条实弹教训）**：①批次二百四十七裁决「不再投机堆词」——堆词面
  对语义层接管率零位移，真杠杆=消歧规则+route_desc 例句；故本批以**外科式补确定性
  症状/关键词层真缺口 + 消歧**为主，9,860→149 保守收敛（非 bulk 灌词），子串/过泛风险词
  （延迟高⊂网络延迟高、服务不可用/资源不足/升级失败 过泛、裸 perf 子串）已剔除。
  ②跨域歧义词一律进消歧不塞单域 symptoms。③route_desc 症状样例句富化（批次247 定档的
  另一真杠杆）与蒸馏暴露的 **RHEL HA/Pacemaker 孤儿词族**（pcs/Pacemaker/stonith/fencing/
  集群脑裂——os-basics 现无专家承接）挂账后续批，需产品拍板是否新增 ha-cluster 域。
- 蒸馏产物与评审件（curated-additions.yaml/ambiguity.md/symptom-candidates.yaml）存
  bianque 侧 `~/my_project/redhat-kb/distill/`；OpenShift 产品专属词（RHACS/OLM/
  MachineConfig/ClusterLogging…）不属本包（主机/OS+K8s 社区版），延后未来 openshift 包。
- **架构校准（bianque 侧 routecheck 内嵌 seed 实证）**：确定性路由中 `symptoms` 仅对
  **P6 专家**（io/memory/network）经症状兜底层生效；perf(P2)/security(P0)/k8s(P2)/
  patrol(P1)/log(P2) 的 symptoms 不进确定性兜底，仅供**在线 LLM 语义层**富化专家描述子
  （与 security P0 基线 symptoms 同一既有设计，非本批引入）。故本批**确定性生效面**=
  route_keywords（全优先级关键词层）+ P6 三域 symptoms + 9 条消歧规则（消歧层先于
  关键词层，实测「内存不足导致系统卡顿」正确归 perf 而非被「卡顿」P1 劫持）。
- **验证**：bianque golden 路由回归 39/39（新增 7 条批次264 用例锁消歧默认/上下文分支+
  P6 症状兜底+route_keyword 直派）；pack-lint 内嵌源 exit 0（短拉丁词 XFS/NFS/MTU/NTP/OSD
  带 advisory 裸泛词警告，与存量 LVM 同类，概念题由「什么是X」守卫拦）；hub validator
  os-basics scope 0 error；go build + agents/scheduler 包绿。
- index.json 经 index-gen 再生（os-basics 1.8.0），validator 0 error，README 版本行同步。

## 1.7.0 (2026-10-10)

- **三域诊断专家最小权限收敛**（bianque 批次二百六十三 工具权限标注审计轮）：
  io/memory/network-analysis 的 `tools.allow: []`（=ask-ops 全部 14 工具，含破坏性
  run_approved_command/run_approved_commands）改显式 12 工具只读白名单（十结构化
  采集器 + probe_host + run_readonly_command(s) 白名单只读 CLI 对）。修正三处不合理：
  ①授权面与 prompt 能力声明矛盾——三份 prompt 首行都承诺「你没有任何处置工具——
  处置建议只随报告提交审批，绝不自行执行」，allow 面却发了破坏性工具（运行时策略门
  会拒绝 LLM 自主变异调用，但授予本身污染工具清单、浪费 ReAct 迭代并诱使模型试探）；
  ②`allow: []` 使专家工具集含变更类，agentrun 重试守卫无法整体放行（RetryAfterMutation
  被迫 false，bianque 实弹 sess-1010-bbnbxb56 轮 429 限流后拒重试环节 failed 的促成
  因素之一，v0.14.6 豁免缝治标、本笔授权收敛治本）；③意图/能力面失真——只读诊断
  专家的工具清单是路由与派工的能力事实源，超授即误描述。
- 三专家 prompt Step1 采集面同步补 `run_readonly_command(s)` 声明（域内补证示例：
  df -i/mount/lsof、slabtop/vmstat、ss/ip route——此前实弹中模型自发使用该工具
  取证有效但 prompt 未声明，授权面与 prompt 各说各话）；io-analysis 补 `get_logs`
  （ENOSPC/挂载错误的应用与内核日志痕迹，Docker 被测体实弹 order-service ENOSPC
  案例即此路径）。
- 审计记录：平台工具注解面（48+8 工具 readOnly/destructive 全标注）与运行时治理链
  （变异 OpToolCall 任何模式拒绝/变更唯一通道=审批方案+一次性凭据/灾难黑名单/
  受审命令 vetting/L4 双确认）核验健壮，本笔为内容侧授权面收敛；hub 其余 46 个
  `allow: []` 专家文件（mw/db/k8s/docker/nginx/web/obs/cicd/gitlab/n8n/oo-devops 等
  非 Linux 主机域）同模式挂账待后续批清扫。

## 1.6.0 (2026-10-10)

- 「报错」消歧规则补排查/性能诉求前置分支（bianque Docker Ubuntu 被测体实弹
  sess-1010-bbnbxb56）：原规则「报错+日志词」直接分流 workflow/log-analysis——用户
  输入「订单服务一直报错，日志里全是写文件失败和超时，机器整体也感觉很卡，帮我全面
  排查」被日志词劫持单派日志专家，CPU 压力/分区写满/僵尸进程/内存大户四重故障全漏，
  报告以 low 置信「需转其他采集面复核」收场。新前置分支：命中 排查/全面/诊断/根因/
  故障/卡顿/很卡/太卡/卡死/变慢/很慢/太慢/性能/负载 任一词即归 workflow/fault-diagnosis
  （「日志里全是报错」是证据位置描述，不改变全面排查诉求；故障链内含日志域专家，
  纯日志分析诉求——无排查/性能词——仍走原日志分支不受影响）。

## 1.5.1 (2026-10-10)

- seed 漂移回流补丁（1.5.0 回灌时发现）：io/memory/k8s-health/network 四专家 agent.yaml
  补 `spawnable_by: [platform/generalist]` spawn 白名单声明——bianque 批次二百零九
  （GE 二期 dispatch_agent 派工）当时直接改了平台仓 seed 快照未上回流 hub，seed-sync
  回灌会把该声明冲掉（通用顾问对四只读诊断域的派工能力回退）。hub 为内容唯一编辑点，
  本笔逐字节对齐 seed 现值完成回流；perf-tuning/security-assistant 不在派工白名单
  （变更倾向/复核面），不声明。

## 1.5.0 (2026-10-10)

- 报告可视化增强批（bianque 批次二百六十联动，用户拍板「尽可能多用图表时间线，不要一片
  文字」）：①六专家（io/memory/k8s-health/network/perf-tuning/security）数据可视化节
  统一补 **timeline 图型**指引——离散事件时序（IO 异常演进/OOM 事件/集群 Pod 异常演进/
  链路异常/负载劣化/攻击痕迹）与折线图分工明示（line 画数值采样趋势，timeline 画事件
  先后），示例含 time 字面/status ok|warn|down/3–12 事件升序/只收证据支持事件；
  ②六专家输出铁律补 conclusion/summary **结构化表达**纪律——首行一句话判定、多要点
  markdown 列表、枚举对比 GFM 表格（表格必须前置空行另起一行行首，内联同行不渲染），
  数据只放真实采集值。协议侧 timeline 容错解码与前端 TimelineChart 渲染已在 bianque
  批次二百六十落地（"timeline" 含子串 "line"，协议判定先于 line；「时序/timeseries」
  仍归折线不抢）。

## 1.4.2 (2026-10-07)

- 215 八会话实弹扫描修复批：①报告协议 steps 硬钉补齐——memory/io/network/security/k8s-health
  五专家 prompt 输出铁律统一加「steps[] 每项=单条可执行命令字符串，说明放 needs_followup」
  （实弹 sess-1007-mrydrs8w：memory 独立会话 steps 写成散文被 schema 白名单拒收，用户拿到
  错误串而会话仍 succeeded）；②k8s 集群健康路由劫持修复——system-patrol 裸词「健康」拆为
  系统健康/健康检查/健康巡检/健康体检（P1 裸词把「检查K8s集群健康状态」劫去全域巡检、锚点
  被回写改道重扫 40min，sess-1007-zea572dx），k8s-health 补 k8s集群健康/k8s集群状态 复合词
  （「k8s」3 rune 过不了短拉丁词闸，原词表 k8s集群诊断 覆盖不了健康构式）。
## 1.4.1 (2026-10-07)

- 部署形态分诊前置（215 实弹 sess-1007-hxr6k25b 教训：localhost=扁鹊容器（Alpine
  最小镜像）被当"这台机器"审计，容器识别靠模型自觉而非制度）：平台适配矩阵
  （hub knowledge/ 单源 + seed-sync 三份同步）新增「部署形态分诊」节——容器证据
  （.dockerenv/containerenv/overlay/PID1）→ 容器口径（主机级项标「容器内不可验证，
  需宿主/编排层复核」不判不符合；kernel/swap/内存注明归属宿主）；security-assistant
  prompt Step0 改「先分诊后采集」，判读口径补容器目标与工具能力降级两条（read 家族
  缺 `file` 环境同路径失败一次即转 list_directory/collect 面，不重试）。配套：扁鹊
  仓 tools/servers/securityx 预检对 file 命令缺失（本地 ENOENT/远端 rc 127）降级——
  head -c 8192 探头 NUL 扫描 + 整读内容复扫兜底（本仓不阻塞，随扁鹊发布生效）。

## 1.4.0 (2026-10-06)

- 深审修复批：`k8s-node-diagnosis` 按仓内五段式补齐（数据来源/判读基准/输出要求，
  0.2.0→0.3.0）并去掉未知 frontmatter 键 `provides`、补 `maturity`；安全四技能
  收割残留清理——悬空互引（security-c2-detection/linux-persistence/account-permission
  → 包内真实技能）、`AGENTS.md「报告协议」`死指针→平台统一报告 schema、自创
  「用户确认后执行处置」审批语义→平台审批口径、requires_mcp 补声明正文实际
  使用的 collect_network_status/list_directory/read_file；`changes/sysctl-kv-tune`
  持久化改按 key 幂等 upsert（整文件覆盖会抹掉同站点首次调参）；四专家补
  route_desc；k8s-health prompt 显式钉 requires_approval/decision 语义、agent 注释
  指针更新。（注：曾试随包带 mcp/k8sgpt.yaml 契约，215 实测 reload 报「tool_manifest
  同名冲突」——k8s-ops 已带同名契约，平台按包名字典序取 k8s-ops，双份纯噪音，故撤出；
  hub 侧对应 WARN 由「契约单源在 k8s-ops」口径收口。）

## 1.3.3 (2026-10-03)

- verify_readonly 接入平台自动执行（批次九十五二阶段）：sysctl-kv-tune 补 `sysctl -n {{key}}` 变更后验证命令（RenderVerify 参数代换与 commands 同口径）——首个带自动验证的编目变更块。

## 1.3.2 (2026-10-02)

- 批次九十四「诊断+处方+受审执行」首批：新增编目变更块 `sysctl-kv-tune`（sysctl 运行时+持久化调参，params: key/value）；perf-tuning 输出铁律补 change_ref 处方口径（sysctl 类调参优先编目、自由 steps 兜底）。

## 1.3.1 (2026-10-01)

- k8s-health 提示词补 k8sgpt 工具契约：analyze 一律 explain:false（explain:true 走哑后端必失败、有发现的扫描整体报错——rd1/106 实弹 9 条发现被吞）；零发现必须带过滤器口径；系统 ConfigMap 未引用降权；事件计数标注时效。

## 1.3.0 (2026-09-30)

- 社区版首发（bianque-hub 上游化）：本包成为内置同名核心包的上游，slug/路由/技能
  与内置版完全等位，升级即原位接管；平台速查段（platform-matrix）按契约引用平台
  `_shared/` 共享层，参考副本随仓 knowledge/ 分发供非扁鹊运行时使用。
- 新增 chain.yaml：workflow/host-quick-audit 主机快查链（安全基线 + 内存进程
  事实两步；P3，路由词避开全域巡检域词）。

## 1.2.0 (2026-09-24)

- 合并 k8s-pack：specialists/k8s-health（P2，k8sgpt 分析桥）+ k8s-node-diagnosis 技能并入本包，
  插件页收敛为「内置核心（os-basics）+ 市场包（oo-devops）」两层；slug 不变，调度链绑定
  （scheduler/flows.go 巡检扇出）不受影响。原 k8s-pack 1.0.0~1.1.0 历史并入本包线。
- analyzers.yaml 合并 sysprobe 域映射与 k8sgpt 检查源声明；kubeconfig 凭证前提仅 K8s 域消费。

## 1.1.1 (2026-09-24)

- 门面定位化：本包为主机/OS 域唯一入口（oo-devops 主机巡检/资源监控类技能已让位收敛至此）。

## 1.1.0 (2026-09-22)

- 统一包格式：补 provides/changelog；安全域 4 技能（ssh 爆破/-root/sudo/进程）与 6 域专家。
