---
name: nginx-lifecycle-ops
description: Nginx 生命周期面方法论：nginx -t 前置纪律、reload 与 restart 的进程语义差（master/worker 继承与连接排空）、-t 过但 reload 失败的冤案形态、logrotate 与 USR1、平滑升级二进制——固定顺序与判读基准（ask-ops 只读采集面；变更动作一律审批后执行）。
mode: on_demand
version: 0.1.0
maturity: experimental
requires_mcp:
  - server: ask-ops
    tools: [run_readonly_command, run_readonly_commands]
---

## 触发条件

- 症状关键词：nginx reload 失败、nginx -t 失败、配置改了不生效、日志没切、nginx 升级、
  二进制替换、logrotate 报错
- 组合场景：改配置后 502/404 先核对运行时态（`nginx -T` 与磁盘 diff——reload 没执行
  或执行失败是高频冤案）；日志文件句柄不释放=轮转后缺 USR1

## 数据来源（ask-ops 只读面）

1. 配置校验：`nginx -t`（恒第一步——语法/包含文件/证书路径全查）；带路径变体
   `nginx -t -c /path/to/nginx.conf`（多实例场景）；
2. 运行时态 vs 磁盘态：`nginx -T > /tmp/nginx-T-now.dump` 与磁盘配置 diff——
   不一致=改完没 reload 或 reload 失败；
3. 进程语义核对：`ps -o pid,ppid,cmd -C nginx`（master 单个+worker 群，worker 的
   PPID=master）——reload 语义=master 重读配置、新 worker 起旧 worker 排空退出
   （不断连接）；`systemctl status nginx` 的信号历史；
4. error log：`tail -20 /var/log/nginx/error.log`——reload 失败的真实原因在这
   （-t 过但 reload 失败常见于：运行中 master 用的二进制/配置路径与 -t 的不同实例）；
5. 日志轮转面：`cat /etc/logrotate.d/nginx`、`ls -l /var/log/nginx/`（有无
   access.log.1 与 nginx 进程仍握旧文件句柄——`lsof -p <master_pid> | grep log`）。

## 判读基准

- `-t` 过但 reload 不生效：多 nginx 实例/容器内外配置不同源——核对运行中 master 的
  `nginx -V` 编译前缀与 -T 输出是否同一份配置（`ls -l /proc/<master_pid>/exe`）；
- reload 后旧 worker 长存：旧 worker 在排空长连接（keepalive_timeout 内正常），
  超时仍存=有连接卡死（查其服务的 client）；
- 日志不切：logrotate 用了 copytruncate 则无需信号；rename+create 形态必须
  `nginx -s reopen`（USR1）——轮转后旧句柄继续写=缺这一步；
- 平滑升级二进制：替换二进制→`kill -USR2 <master>`（新 master 双跑）→旧 master
  `WINCH`/`QUIT` 收 worker——任一步失败回滚=旧二进制原路径还在（升级前备份）。

## 处置面（均审批后宿主侧执行）

- 一切配置变更：`nginx -t` → `nginx -s reload`；变更前 `nginx -T` dump 落盘留回滚点；
- 二进制升级/回滚：备份旧二进制→按 USR2/WINCH/QUIT 序贯——全程 error log 观察。
