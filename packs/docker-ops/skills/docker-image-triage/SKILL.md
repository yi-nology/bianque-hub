---
name: docker-image-triage
description: Docker 镜像与磁盘面分诊方法论：system df 四类水位定位、悬空镜像与层膨胀判读（images/history）、拉取失败三分（网络不可达/认证失效/限流 429）与镜像加速器核对、磁盘回收编目纪律（prune 梯度：dangling 可编目、-a 级深清走审批）——固定顺序定位与判读基准（ask-ops 只读采集面 + 编目变更块受审执行）。
mode: on_demand
version: 0.1.0
maturity: experimental
requires_mcp:            # 采集依赖的工具面（装载期对账依据；旧 bianque-tools 二进制缺这些工具时应先重建）
  - server: ask-ops
    tools: [run_readonly_command, run_readonly_commands]
provides_changes:        # 编目变更块（受审执行）：本技能方法论覆盖的处方编目
  - docker-image-prune   # 悬空镜像回收（无参数）——磁盘水位恢复处方优先引用编目
---

## 触发条件

- 症状关键词：docker 磁盘占满/空间不足、no space left on device、镜像拉取失败/超时、dockerhub 限流、镜像仓库登录失败、悬空镜像清理、构建缓存膨胀
- 组合场景：容器写日志打满盘走 docker-container-triage 日志面；宿主本体磁盘面（非 /var/lib/docker）走 os-basics——本技能管「镜像/层/卷/构建缓存的 docker 存储」与拉取通道

## 数据来源（ask-ops 只读面）

- 水位面：`docker system df`（Images/Containers/Volumes/Build Cache 四类总量与可回收量）、
  `docker system df -v`（逐镜像/容器/卷明细，输出大先 df 后 -v）、
  `df -h /var/lib/docker`（宿主落盘点）、`du -sh /var/lib/docker/*`（分目录对账）；
- 镜像面：`docker images`（dangling `<none>` 判读 + 体积）、`docker history <镜像>`（逐层体积，
  层膨胀定位）、`docker images -f dangling=true`（悬空清单）；
- 拉取通道：`curl -s 'https://registry-1.docker.io/v2/'`（401=网络通（registry 语义，正常形态）、
  超时/连接失败=不可达）、`curl -s 'https://auth.docker.io/token'`（认证通道）、
  `cat /etc/docker/daemon.json` 的 registry-mirrors 键（加速器有无与指向）；
  **docker pull 不在只读面**——「复现拉取」一律审批后宿主侧动作；
- 认证面：`ls ~/.docker/config.json` + 选择性读——`auths` 段**值禁读**（base64 即明文），
  只判「有无对应 registry 的登录态」。

## 分诊路径（固定顺序：先水位后对账，镜像分层，拉取三分，回收按梯度）

1. **四类水位**：`docker system df` 定位大头在 Images/Containers/Volumes/Build Cache 哪一类
   ——再决定走本技能哪段（Images → 第 2/4 步；Volumes → 卷面（volume ls/inspect 逐个对账
   挂载引用）；Build Cache → builder 面审批后清）；df 可回收量与 `du /var/lib/docker` 对不上
   = 有非 docker 目录落在同盘或元数据漂移，如实区分；
2. **镜像分层**：`docker images` 悬空（`<none>`）与同 tag 多条（旧层滞留）清单 +
   `docker history` 层体积——「镜像越滚越大」通常是最胖层（日志/缓存被写进层），history 给层号证据；
3. **拉取失败三分**：①网络不可达（curl registry 超时——代理/出网面，转宿主网络域）；
   ②认证失效（config.json 无对应 registry 登录态或曾 401/unauthorized——私有仓库必判此段）；
   ③限流 429（toomanyrequests——匿名拉取配额，登录态/加速器是建议面处方）；
   错误原文以用户侧报错为准（只读面不能 pull 复现），三分先于「换镜像加速器」的万能建议；
4. **磁盘回收按梯度**（红线：先清单后删除，运行中容器的镜像与被引用卷永不在「清理」语义内）：
   悬空镜像回收走编目 `docker-image-prune`（等价 `docker image prune -f`，仅删无 tag 引用层）；
   `-a` 级深清（docker system/image prune -a，删全部未使用镜像）、未引用卷删除、构建缓存清理
   **不入编目**——「未使用」的站点语义各异，一律自由 steps + requires_approval（审批卡标「未编目」），
   处方前必须附「将被删除对象清单」（images -a 最小集/卷引用对账）；
5. **收尾核验**：回收后 `docker system df` 复测水位（编目块 verify_readonly 自动带）。

## 判读基准

- 「no space left」但 system df 四类都不大：查 docker 日志驱动文件（containers 下 *-json.log，
  container-triage 日志面）与 journal 磁盘占用——容器写盘的债常不在镜像账上；
- 悬空镜像是「构建残留的中间层」，删了不影响任何运行中容器与有 tag 镜像——这是唯一可编目
  的自动回收面；「prune -a 也安全」是误传：它删的是所有**未被容器引用**的镜像（含备用回滚镜像），
  回滚镜像被删=事故，必须人工审；
- dockerhub 限流按请求方 IP 计（出网 NAT 后多机共享配额）——单机重试无解时如实说明形态；
- 数据不足输出「需补充采集」清单（df -v 明细、目标镜像与 tag、daemon.json mirrors 段、
  报错原文截图/文本），不臆测。

## 输出要求

- 每个结论附 system df 数值/images 清单/curl 状态码证据；一切回收/删除动作标注梯度与影响面并
  requires_approval（悬空回收优先 `recommendation.change_ref: docker-image-prune` 引用编目；
  变更后 `docker system df` 自动验证）；数据不足输出「需补充采集」清单，不臆测。
