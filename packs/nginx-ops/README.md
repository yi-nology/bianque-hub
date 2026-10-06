# nginx-ops

Nginx 自托管运维包（社区版）：纯宿主形态诊断——统一 ask-ops 只读采集面
（`nginx -t/-T`、curl 本地端点、ss/lsof、systemctl/journalctl、error/access log tail），
reload/restart/证书替换类变更动作一律审批后宿主侧执行（不在只读白名单）。

## 组成

- **专家**：`imports/nginx-platform-analyst`（四面分诊：可用性/路由重写/TLS/生命周期）
- **技能**：`nginx-availability-triage`（502/504 upstream 三分与容量）、
  `nginx-route-triage`（location 匹配序/root-alias/proxy_pass 尾斜杠/循环）、
  `nginx-tls-triage`（证书链/s_client 判读/SNI/混合内容）、
  `nginx-lifecycle-ops`（-t 纪律/reload vs restart/USR2 平滑升级/logrotate USR1）

## 范围边界

K8s Ingress 形态、openresty/lua 扩展面、其他反代（HAProxy/Traefik/Caddy）不在本包；
不依赖凭证平面（纯宿主 ask-ops 面即全能力）。

## 方法论来源

原生编写：Nginx 官方文档（nginx.org/en/docs：http 反代语义、ssl 模块、进程模型、
信号控制）+ 社区共识排障流程整理重写；命令均按 ask-ops 只读白名单校形。
