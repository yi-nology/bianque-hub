---
name: docker-network-triage
description: Docker 网络面分诊方法论：端口发布两级核对（-p 映射 vs 容器内监听地址）、bridge/自定义网络语义（内置 DNS 只有自定义网络有）、network inspect 成员与 IPAM 子网冲突、宿主侧发布端口交叉核对（get_listening_ports）——固定顺序定位与判读基准（ask-ops 只读采集面）。
mode: on_demand
version: 0.1.0
maturity: experimental
requires_mcp:            # 采集依赖的工具面（装载期对账依据；旧 bianque-tools 二进制缺这些工具时应先重建）
  - server: ask-ops
    tools: [run_readonly_command, run_readonly_commands]
---

## 触发条件

- 症状关键词：容器端口不通、宿主访问不到容器、外部访问容器端口失败、容器之间不通、容器 DNS 解析失败、端口冲突、网桥异常、ip address pool 报错
- 组合场景：容器本体退出/循环先走 docker-container-triage；宿主防火墙（firewalld/ufw/安全组）
  面走 os-basics 网络域——本技能管「docker 网络模型本身」

## 数据来源（ask-ops 只读面）

- 发布面：`docker port <容器>`（容器端口→宿主地址:端口的实际映射）、
  `docker ps --format` Ports 列（0.0.0.0 vs 127.0.0.1 绑定形态一目了然）；
- 容器网络配置：`docker inspect <容器>` 的 NetworkSettings（IPAddress/Gateway/Ports/
  Networks 成员）；
- 网络面：`docker network ls`（网络清单与 driver）、
  `docker network inspect <网络>`（成员容器清单、IPAM Subnet/Gateway、options）；
- 宿主交叉核对：固定清单面 `get_listening_ports`（宿主实际监听与发布端口对账）、
  `get_network_inventory`（宿主网卡与 docker0/自定义网桥形态）；
- 应用探活：`curl -s 'http://127.0.0.1:<发布端口>/...'`（从宿主打发布端口——GET 探测，
  路径按应用语义选 / 或健康端点）。

## 分诊路径（固定顺序：先发布映射，后监听地址，再网络成员，子网冲突收尾）

1. **发布映射存在性**：`docker port <容器>` 空 = 容器根本没 -p 发布（「端口不通」的第一大主因）
   ——发布语义由容器创建时决定，补发布=重建容器（审批后自由 steps，禁处方化 run）；
2. **两级监听核对**：`docker ps` Ports 列看绑定地址——`127.0.0.1:8080->80/tcp` 形态只收
   本机回环，「外部访问不通但宿主 curl 通」的指纹；容器内应用只监听 127.0.0.1（很多镜像默认）
   时即便 `0.0.0.0:8080->80` 发布了外部也不通——容器内监听面查不到（exec 不在白名单）时用
   「宿主 curl 发布端口 vs 外部 curl」对比事实反推，结论标注证据边界；
3. **网络成员与 DNS**：容器间按名字互访需要**同一个自定义网络**（内置 DNS 127.0.0.11 只有
   自定义网络有；默认 bridge 只认 IP 不认名字）——`docker network inspect` 成员清单对账
   「两个容器在不在网络里」；默认 bridge 互访用 IP，名字解析失败先查网络归属；
4. **子网冲突**：`docker network inspect` 各网络 IPAM Subnet 与宿主内网段重叠（报
   「could not find an available, non-overlapping IPv4 address pool」或大网段环境部分容器
   异常）——自定义网段规划是建议面（network create 一律审批后动作）；
5. **收尾归因**：映射对、监听对、网络成员对仍不通 → iptables DOCKER 链/防火墙/安全组面
   （iptables 不在只读白名单——作为审批后宿主侧核验动作；宿主防火墙转 os-basics 网络域），
   结论如实标「docker 面未见异常，疑在宿主网络面」。

## 判读基准

- 「宿主通、外部不通」是三件事的指纹：127.0.0.1 绑定发布、容器内应用绑回环、宿主防火墙——
  顺序核对（第 1/2 步零成本），禁跳步直接建议关防火墙；
- 「容器互 ping 不通」先分网络归属再谈 ICMP：不同网络/默认 bridge 的名字解析失败不是网络故障，
  是架构事实——处方是「接入同一自定义网络」，不是重启容器；
- 「重启容器/重建网络」类动作全部走审批后自由 steps（docker network create/connect、容器重建），
  本技能不开任何写处方；数据不足输出「需补充采集」清单（容器与网络名、-p 原始参数、
  期望访问路径与来源），不臆测。

## 输出要求

- 每个结论附 port/Ports 列/network inspect 关键字段证据，监听地址结论必附「宿主 vs 外部」
  对比事实；网络创建/接入/重建一律标注影响面与回退路径并 requires_approval；数据不足输出
  「需补充采集」清单，不臆测。
