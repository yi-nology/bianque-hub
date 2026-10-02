---
name: harbor-triage
description: Harbor 制品库分诊方法论：拉取失败（存储/GC 窗口）、磁盘水位与垃圾、证书过期、复制任务失败、组件健康五面——固定顺序定位与判读基准（ask-ops 只读采集面）。
mode: on_demand
version: 0.1.0
maturity: experimental
---

## 触发条件

- 症状关键词：镜像推送失败、拉取不到镜像、制品库排障、harbor慢、harbor不可用、复制失败、仓库证书告警
- 组合场景：K8s 侧 ImagePullBackOff 先经 k8s-ops 分诊（凭证/网络三分）后再进本技能——本技能管制品库本体

## 数据来源（ask-ops 只读面）

- Harbor REST GET：`/api/v2.0/health`（组件健康）、`/api/v2.0/replication/executions?sort=-start_time`（复制任务最近态）；
- 主机侧：`docker ps`（core/jobservice/registry/redis/db 容器状态）、存储目录 du/df、registry 日志尾部；
- 证书面：`echo | openssl s_client -connect <harbor-host>:443 | openssl x509 -noout -enddate`（只读探测剩余有效期；stderr 不采集，无需也无法重定向）；
- 凭证由主机侧已配置环境注入，对话不回显。

## 分诊路径（按症状类定位，固定顺序）

1. **组件健康**：`/api/v2.0/health` 一发定性：任一组件 unhealthy 先定位该组件容器/日志；整体健康但拉取仍失败进第 2 步（问题在存储/网络/凭证面，不在服务面）。
2. **拉取失败三分**：**404/blob unknown**=存储层缺块（GC 误删嫌疑，去第 4 步核对最近 GC/删除事件；manifest 在 blob 不在的形态是铁证）、**401/403**=凭证/项目权限（K8s侧 imagePullSecrets 或机器人账号面）、**timeout/连接拒绝**=网络与 LB 面（联动主机网络域）。
3. **磁盘水位**：du 存储目录（registry blob 大户）、df 对总量 → 水位逼近上限时推送先失败（写入面比读取面先受害）；清理与 GC 是变更建议（GC 期间库只读，影响面必须写明）。
4. **GC 与删除面**：核对 GC 执行记录时间窗（维护日志/API 证据）——**GC 进行中或异常中断后的窗口期，blob 被标记删除会造成批量 unknown**；确认时间对齐后才下「GC 误删」结论，恢复属变更类方案（备份/重推镜像）。
5. **复制与证书面**：复制 executions 最近失败记录看 error 摘要（对端不可达/认证失败/带宽限速）；证书剩余天数 <30 天列 WARN（轮换是变更建议，需与 k8s-ops 集群侧 imagePullSecrets 联动更新）。

## 判读基准

- 拉取失败先分「服务面/存储面/凭证面/网络面」四归因，禁笼统说「Harbor 挂了」；
- 偶发推送失败与网络抖动对齐（时间分布），持续失败才指向容量/配置；
- 复制任务延迟在带宽约束内属正常形态，失败重试成功的不构成异常结论。

## 输出要求

- 每个结论附 API/日志关键行证据；变更类动作（GC、清项目、轮证书、重推镜像）标注影响面并 requires_approval；数据不足输出「需补充采集」清单（如 GC 执行记录、LB 日志），不臆测。
