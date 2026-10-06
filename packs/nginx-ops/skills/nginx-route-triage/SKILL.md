---
name: nginx-route-triage
description: Nginx 路由与重写面分诊方法论：404 的 location 匹配序与 root/alias 语义差、重定向循环的反代头与 TLS 终止点形态、try_files 兜底项、proxy_pass 尾斜杠语义——固定顺序定位与判读基准（ask-ops 只读采集面）。
mode: on_demand
version: 0.1.0
maturity: experimental
requires_mcp:
  - server: ask-ops
    tools: [run_readonly_command, run_readonly_commands]
---

## 触发条件

- 症状关键词：nginx 404、页面 404 但文件存在、重定向循环、ERR_TOO_MANY_REDIRECTS、
  location 不生效、rewrite 不生效、try_files、代理路径丢失、proxy_pass 尾斜杠
- 组合场景：新增路径 404 先看该 server 块是否 reload 过（`nginx -T` 是运行时态，与磁盘
  文件 diff 即知）；整站循环转 nginx-tls-triage（http/https 互跳是头号形态）

## 数据来源（ask-ops 只读面）

1. 运行时配置全量（含 include 展开，与磁盘态 diff）：
   `nginx -T 2>/dev/null | grep -n -A 8 'location'`、`nginx -T > /tmp/nginx-T.dump` 后
   `diff /tmp/nginx-T.dump /etc/nginx/nginx.conf`——运行中配置≠磁盘配置是高频冤案；
2. 匹配复现：`curl -s -o /dev/null -w '%{http_code} %{redirect_url}\n' -H 'Host: 站点域名'
   http://127.0.0.1/路径`——带 Host 复现（server_name 分流），看 status 与 Location 头；
3. 匹配序核对（nginx 语义单源）：精确 `= /path` > `^~` 前缀 > 正则（按书写序）> 普通前缀
   最长匹配——正则互相遮蔽与书写顺序相关是 404 高频根因；
4. root vs alias：`location /static/ { alias /data/; }`（alias 尾斜杠替换整段前缀）与
   `root`（拼接）语义差——路径多一层/少一层目录即此因；
5. 重定向循环：`curl -sIL http://127.0.0.1/路径 | grep -E 'HTTP|Location'` 看跳转链——
   http→https→http 循环=反代头（`proxy_set_header X-Forwarded-Proto $scheme`）缺失或
   上层 LB 已终止 TLS 而 nginx 又强制跳转。

## 判读基准

- 文件在磁盘但 404：root/alias 拼接路径错——用 `nginx -T` 里的最终路径手拼核对；
- 代理路径丢失（后端收到多/少一段前缀）：proxy_pass 带尾斜杠=替换 location 前缀、
  不带=原样透传——二选一语义混用是最高频；
- try_files `$uri $uri/ /index.php?$args`：最后兜底项决定 404 还是进框架前端控制器——
  兜底项写错=所有缺失路径都进错误入口；
- 循环跳转：跳转链里 http 与 https 交替=终止点与跳转逻辑不一致，修反代头或去重跳转层
  （变更面，审批）。

## 处置面（均审批后宿主侧执行）

- 修正 location/rewrite/proxy_pass → `nginx -t` → `nginx -s reload`；变更前 `-T` dump 留档。
