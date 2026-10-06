---
name: nginx-availability-triage
description: Nginx 可用性面分诊方法论：502 与 504 的 upstream 三分（连接拒=后端挂/超时=后端慢/无可用 upstream=worker 或配置面）、worker 打满与连接数上限、本地回环复现剥离客户端网络面——固定顺序定位与判读基准（ask-ops 只读采集面）。
mode: on_demand
version: 0.1.0
maturity: experimental
requires_mcp:
  - server: ask-ops
    tools: [run_readonly_command, run_readonly_commands]
---

## 触发条件

- 症状关键词：nginx 502、nginx 504、网站打不开、连接被拒绝、upstream 超时、偶发 5xx
- 组合场景：全站 502 先查 nginx 进程与端口（本技能第一步），部分路径 502 转
  nginx-route-triage 核对对应 location 的 upstream；伴随证书告警转 nginx-tls-triage

## 数据来源（ask-ops 只读面，固定顺序）

1. 进程与端口：`systemctl status nginx`（active? worker 数?）、`ss -ltnp | grep -E ':(80|443)\b'`
   ——端口没人听=配置 listen 错或 master 挂；端口在听=往后走；
2. 本地回环复现（剥离客户端网络/CDN 面）：
   `curl -s -o /dev/null -w '%{http_code} %{time_total}\n' http://127.0.0.1/目标路径`
   ——本地也 502=upstream 面；本地 200=客户端到 nginx 之间（反代/防火墙/CDN）；
3. error log 定性（最近段即现场）：`tail -50 /var/log/nginx/error.log`
   ——`connect() failed (111: Connection refused)`=后端进程挂；
   `(110: Connection timed out)`=后端慢/满；`worker_connections are not enough`=打满；
   `no live upstreams`=全部后端被标记失败（max_fails 窗口内）；
4. 后端本体：`ss -tnp | grep <后端端口> | head`（连接堆积?）、后端进程存活与自身日志；
5. 容量面：`ss -s`（总连接/各态）、`ps aux | grep nginx | wc -l` 对照 `worker_processes`/
   `worker_connections` 乘积——接近系统 fd 上限（`cat /proc/sys/fs/file-max`）即容量顶。

## 判读基准

- 502 + Connection refused：后端挂/端口错——修后端或改 upstream 端口（变更面，审批）；
- 504 + upstream timed out：后端慢——看后端（应用慢查询/GC），nginx 侧只能调
  proxy_read_timeout 缓解（治标，标注清楚）；
- `no live upstreams` 突发恢复：max_fails+fail_timeout 把后端临时拉黑——后端间歇抖动，
  查后端重启记录；
- worker 打满：并发顶配——调 worker_processes=auto/worker_connections（变更面，审批）
  并先核内存（每 worker 常驻数十 MB 量级）。

## 处置面（均审批后宿主侧执行）

- 配置修正走 `nginx -t` → `nginx -s reload`（reload 平滑不断连；restart 仅最后手段）；
- 变更前 `nginx -T > /tmp/nginx-T-$(date +%s).dump` 留回滚点。
