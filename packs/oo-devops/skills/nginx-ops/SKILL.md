---
name: nginx-ops
description: Nginx 运维专家 - 配置优化、性能调优、负载均衡、故障排查
mode: on_demand
version: 1.1.0
maturity: experimental
---

## 配置说明

### 环境变量配置
```bash
export NGINX_CONF_PATH="/etc/nginx/nginx.conf"
export NGINX_SITES_ENABLED="/etc/nginx/sites-enabled"
```

## 输入参数

| 参数名 | 类型 | 必填 | 描述 | 示例 |
|--------|------|------|------|------|
| `site_name` | string | 否 | 站点名称 | `example.com` |
| `action` | string | 否 | 操作 | `reload`, `test` |

## 输出格式

```json
{
  "status": "success",
  "data": {
    "config_test": "successful",
    "active_connections": 42
  }
}
```

> **PowerShell 支持**: Nginx for Windows 可用，但功能有所限制。以下提供 Windows 特定的路径和命令。

# Nginx 运维助手

你是 Nginx 运维专家，擅长 Web 服务器配置优化、负载均衡架构、SSL 证书管理和故障排查。

## 核心能力

- **配置管理**：虚拟主机、反向代理、动静分离、URL 重写
- **性能优化**：Worker 进程优化、连接池、Gzip 压缩、缓存策略
- **负载均衡**：Upstream 配置、健康检查、会话保持、熔断降级
- **安全防护**：WAF 规则、限流限速、IP 黑名单、Bot 防护
- **SSL/TLS**：证书配置、HTTPS 优化、HTTP/2、OCSP Stapling
- **日志分析**：访问日志、错误日志、实时分析、性能统计
- **故障诊断**：502/504 错误、高 CPU、连接泄漏、配置校验

## 标准诊断流程

```bash
# 1. 版本和配置检查 (Linux)
nginx -v
nginx -V  # 查看编译参数
nginx -t  # 检查配置语法

# 2. 进程状态 (Linux)
ps aux | grep nginx
systemctl status nginx

# 3. 连接状态 (Linux)
ss -antp | grep nginx
netstat -antp | grep nginx

# 4. 日志查看 (Linux)
tail -f /var/log/nginx/error.log
tail -f /var/log/nginx/access.log

# 5. 性能统计
curl http://localhost/nginx_status  # 需开启 stub_status
```

**PowerShell (Windows)**:
```powershell
# 1. 版本和配置检查 (Windows)
nginx -v
nginx -V
nginx -t  # 默认配置文件在 C:\nginx\conf\nginx.conf

# 2. 进程状态 (Windows)
Get-Process | Where-Object {$_.Name -like "*nginx*"}
Get-Service nginx  # 如果注册为服务

# 3. 连接状态 (Windows)
Get-NetTCPConnection | Where-Object {$_.OwningProcess -in (Get-Process nginx).Id} | Select-Object LocalAddress, LocalPort, State

# 4. 日志查看 (Windows)
Get-Content C:\nginx\logs\error.log -Tail 50 -Wait
Get-Content C:\nginx\logs\access.log -Tail 50 -Wait

# 5. 性能统计
curl http://localhost/nginx_status
```

## 常见故障处理

### 1. 502 Bad Gateway
```bash
# 检查后端服务
systemctl status php-fpm
systemctl status gunicorn

# 检查后端端口
ss -antp | grep :9000

# 检查 Upstream 配置
cat /etc/nginx/conf.d/upstream.conf | grep -A 5 "upstream"

# 检查文件权限
ls -la /var/run/php/

# 增加超时时间
proxy_connect_timeout 60s;
proxy_send_timeout 60s;
proxy_read_timeout 60s;
```

### 2. 504 Gateway Timeout
```bash
# 增加超时配置
proxy_read_timeout 300s;
fastcgi_read_timeout 300s;

# 检查后端处理时间
tail -f /var/log/php-fpm/www-slow.log

# 优化 PHP 执行时间
sed -i 's/max_execution_time = 30/max_execution_time = 300/' /etc/php.ini
```

### 3. 高 CPU 占用
```bash
# 找到高 CPU Worker (Linux)
ps aux | grep nginx | sort -k3 -rn | head

# 检查访问日志，定位攻击或异常请求 (Linux)
cat /var/log/nginx/access.log | awk '{print $7}' | sort | uniq -c | sort -rn | head -20

# 启用限流
limit_req_zone $binary_remote_addr zone=one:10m rate=10r/s;
limit_req zone=one burst=20 nodelay;

# 检查是否有死循环重定向
curl -I -L http://localhost 2>&1 | grep Location
```

**PowerShell (Windows)**:
```powershell
# 找到高 CPU Worker (Windows)
Get-Process | Where-Object {$_.Name -like "*nginx*"} | Sort-Object CPU -Descending | Select-Object Name, CPU, Id

# 检查访问日志，定位攻击或异常请求 (Windows)
Get-Content C:\nginx\logs\access.log | ForEach-Object { ($_ -split " ")[6] } | Group-Object | Sort-Object Count -Descending | Select-Object -First 20

# 启用限流 (nginx.conf)
# limit_req_zone $binary_remote_addr zone=one:10m rate=10r/s;
# limit_req zone=one burst=20 nodelay;

# 检查是否有死循环重定向
Invoke-WebRequest -Uri http://localhost -MaximumRedirection 5 -UseBasicParsing
```

### 4. 内存不足
```bash
# 调整 Worker 进程数
worker_processes auto;  # 或设置为 CPU 核心数
worker_connections 4096;

# 启用高效的文件传输
sendfile on;
tcp_nopush on;
tcp_nodelay on;
```

### 5. 大量 499/444 错误

**症状**：访问日志中出现大量 499（客户端关闭连接）或 444（Nginx 拒绝连接）错误

**诊断流程**：
```bash
# 1. 统计 499 错误
awk '$9 == 499 {print $0}' /var/log/nginx/access.log | head -20

# 2. 查看对应时间点的错误日志
grep "upstream timed out" /var/log/nginx/error.log

# 3. 分析请求时间分布
awk -F'rt=' '{print $2}' /var/log/nginx/access.log | awk '{print $1}' | sort -n | uniq -c | tail -20
```

**常见原因及处理**：

1. **后端响应慢**：
```nginx
# 增加超时时间
proxy_read_timeout 120s;
proxy_connect_timeout 60s;
proxy_send_timeout 60s;

# 启用缓冲
proxy_buffering on;
proxy_buffer_size 4k;
proxy_buffers 8 4k;
proxy_busy_buffers_size 8k;
```

2. **客户端超时关闭**：
```nginx
# 优化 keepalive
keepalive_timeout 65;
keepalive_requests 1000;
```

### 6. SSL/TLS 握手失败

**症状**：HTTPS 请求失败，错误日志显示 SSL 相关错误

**诊断流程**：
```bash
# 1. 检查证书有效期
openssl x509 -in /etc/nginx/ssl/cert.pem -noout -dates

# 2. 检查证书链完整性
openssl x509 -in /etc/nginx/ssl/cert.pem -noout -text | grep -A2 "Issuer"

# 3. 测试 SSL 连接
openssl s_client -connect localhost:443 -servername example.com

# 4. 检查支持的协议和加密套件
nmap --script ssl-enum-ciphers -p 443 localhost
```

**处理方案**：
```nginx
# 更新 SSL 配置
ssl_certificate /etc/nginx/ssl/cert.pem;
ssl_certificate_key /etc/nginx/ssl/key.pem;
ssl_trusted_certificate /etc/nginx/ssl/chain.pem;

# 协议和加密套件
ssl_protocols TLSv1.2 TLSv1.3;
ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384;
ssl_prefer_server_ciphers off;

# 会话缓存
ssl_session_cache shared:SSL:10m;
ssl_session_timeout 1d;
ssl_session_tickets off;

# OCSP Stapling
ssl_stapling on;
ssl_stapling_verify on;
resolver 8.8.8.8 8.8.4.4 valid=300s;
resolver_timeout 5s;
```

### 7. 文件描述符耗尽

**症状**：错误日志显示 "too many open files"

**诊断流程**：
```bash
# 1. 查看当前限制
ulimit -n
cat /proc/$(pgrep nginx | head -1)/limits | grep "Max open files"

# 2. 查看 Nginx 打开的文件数
ls /proc/$(pgrep nginx | head -1)/fd | wc -l

# 3. 查看连接数
ss -ant | grep :80 | wc -l
ss -ant | grep :443 | wc -l
```

### 8. 高并发场景优化

**症状**：高并发时 Nginx 响应慢、连接数暴涨、CPU负载高

**诊断流程**：
```bash
# 1. 查看当前连接状态
ss -ant | awk '{print $1}' | sort | uniq -c

# 2. 查看 Nginx 连接统计
curl http://localhost/nginx_status

# 3. 查看系统连接数
netstat -n | awk '/^tcp/ {++S[$NF]} END {for(a in S) print a, S[a]}'

# 4. 检查系统限制
cat /proc/sys/net/core/somaxconn
cat /proc/sys/net/netfilter/nf_conntrack_max
```

**高并发优化配置**：

1. **系统内核优化**：
```bash
# /etc/sysctl.conf

# 连接跟踪优化
net.netfilter.nf_conntrack_max = 1000000
net.netfilter.nf_conntrack_tcp_timeout_established = 600
net.netfilter.nf_conntrack_tcp_timeout_time_wait = 30

# TCP 连接优化
net.ipv4.tcp_max_tw_buckets = 6000
net.ipv4.tcp_sack = 1
net.ipv4.tcp_window_scaling = 1
net.ipv4.tcp_rmem = 4096 87380 4194304
net.ipv4.tcp_wmem = 4096 16384 4194304
net.core.wmem_default = 8388608
net.core.rmem_default = 8388608
net.core.rmem_max = 16777216
net.core.wmem_max = 16777216
net.core.netdev_max_backlog = 65536
net.ipv4.tcp_max_orphans = 3276800
net.ipv4.tcp_max_syn_backlog = 65536
net.ipv4.tcp_timestamps = 0
net.ipv4.tcp_synack_retries = 2
net.ipv4.tcp_syn_retries = 2
net.ipv4.tcp_tw_reuse = 1
net.ipv4.tcp_mem = 94500000 915000000 927000000
net.ipv4.tcp_fin_timeout = 30
net.ipv4.tcp_keepalive_time = 1200
net.ipv4.ip_local_port_range = 1024 65535

# 文件描述符
fs.file-max = 655360

# 应用配置
sysctl -p
```

2. **Nginx 高并发配置**：
```nginx
# nginx.conf
user nginx;
worker_processes auto;
worker_rlimit_nofile 65535;

error_log /var/log/nginx/error.log warn;
pid /var/run/nginx.pid;

events {
    use epoll;
    worker_connections 65535;
    multi_accept on;
    accept_mutex off;
}

http {
    include /etc/nginx/mime.types;
    default_type application/octet-stream;

    # 日志格式
    log_format main '$remote_addr - $remote_user [$time_local] "$request" '
                    '$status $body_bytes_sent "$http_referer" '
                    '"$http_user_agent" "$http_x_forwarded_for" '
                    'rt=$request_time uct="$upstream_connect_time" '
                    'uht="$upstream_header_time" urt="$upstream_response_time"';

    access_log /var/log/nginx/access.log main buffer=32k flush=5s;

    # 性能优化
    sendfile on;
    tcp_nopush on;
    tcp_nodelay on;
    keepalive_timeout 60;
    keepalive_requests 10000;

    # 文件上传限制
    client_max_body_size 100m;
    client_body_buffer_size 128k;
    client_header_buffer_size 4k;
    large_client_header_buffers 4 8k;

    # 高效文件传输
    sendfile_max_chunk 512k;

    # 开启 gzip
    gzip on;
    gzip_vary on;
    gzip_proxied any;
    gzip_comp_level 6;
    gzip_buffers 16 8k;
    gzip_http_version 1.1;
    gzip_types text/plain text/css application/json application/javascript text/xml application/xml application/xml+rss text/javascript;

    # 限流配置
    limit_req_zone $binary_remote_addr zone=req_limit:10m rate=100r/s;
    limit_conn_zone $binary_remote_addr zone=conn_limit:10m;

    # 上游连接池
    upstream backend {
        least_conn;
        server 10.0.0.1:8080 weight=5 max_fails=3 fail_timeout=30s;
        server 10.0.0.2:8080 weight=5 max_fails=3 fail_timeout=30s;
        keepalive 300;
        keepalive_requests 1000;
        keepalive_timeout 60s;
    }

    server {
        listen 80 backlog=65535;
        server_name example.com;

        # 限流
        limit_req zone=req_limit burst=200 nodelay;
        limit_conn conn_limit 50;

        location / {
            proxy_pass http://backend;
            proxy_http_version 1.1;
            proxy_set_header Connection "";
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;

            # 超时配置
            proxy_connect_timeout 5s;
            proxy_send_timeout 60s;
            proxy_read_timeout 60s;

            # 缓冲配置
            proxy_buffering on;
            proxy_buffer_size 128k;
            proxy_buffers 4 256k;
            proxy_busy_buffers_size 256k;
            proxy_temp_file_write_size 256k;

            # 错误处理
            proxy_next_upstream error timeout invalid_header http_500 http_502 http_503 http_504;
            proxy_next_upstream_tries 2;
        }
    }
}
```

3. **长连接优化（Keepalive）**：
```nginx
# 与客户端的长连接
keepalive_timeout 60s;
keepalive_requests 10000;
keepalive_disable msie6;

# 与后端的长连接
upstream backend {
    server 10.0.0.1:8080;
    server 10.0.0.2:8080;
    keepalive 300;  # 保持300个空闲连接
}

server {
    location / {
        proxy_pass http://backend;
        proxy_http_version 1.1;           # 必须使用 HTTP/1.1
        proxy_set_header Connection "";   # 清空 Connection 头
    }
}
```

### 9. Location 匹配规则详解

**匹配优先级**（从高到低）：
1. `=` 精确匹配
2. `^~` 前缀匹配（匹配后停止搜索）
3. `~` 和 `~*` 正则匹配（区分/不区分大小写）
4. 普通前缀匹配
5. `/` 通用匹配

**Location 配置示例**：

```nginx
server {
    listen 80;
    server_name example.com;

    # 1. 精确匹配 (=)
    location = / {
        root /var/www/html;
        index index.html;
    }

    # 2. 精确匹配特定文件
    location = /favicon.ico {
        root /var/www/static;
        access_log off;
        log_not_found off;
    }

    # 3. 前缀匹配 (^~)
    location ^~ /static/ {
        root /var/www/static;
        expires 30d;
        access_log off;
    }

    # 4. 正则匹配 (~ 区分大小写)
    location ~ \.(gif|jpg|jpeg|png|css|js)$ {
        root /var/www/images;
        expires 30d;
    }

    # 5. 正则匹配 (~* 不区分大小写)
    location ~* \.(html|htm)$ {
        root /var/www/html;
        add_header Cache-Control "no-cache";
    }

    # 6. 排除特定路径
    location !~ ^/api/ {
        root /var/www/public;
    }

    # 7. 匹配后停止正则匹配
    location ^~ /api/ {
        proxy_pass http://api_backend;
    }

    # 8. 通用匹配
    location / {
        proxy_pass http://default_backend;
    }

    # 9. 带变量的 location
    location ~ ^/user/(?<user_id>\d+)$ {
        proxy_pass http://user_service/users/$user_id;
    }

    # 10. 嵌套 location
    location /app/ {
        alias /var/www/app/;

        location ~ \.php$ {
            fastcgi_pass php_backend;
        }
    }
}
```

**Location 修饰符对比**：

| 修饰符 | 说明 | 示例 |
|--------|------|------|
| `=` | 精确匹配 | `location = /exact` 只匹配 /exact |
| `^~` | 前缀匹配，匹配后停止搜索 | `location ^~ /static/` 匹配 /static/... |
| `~` | 正则匹配，区分大小写 | `location ~ \.php$` 匹配 .php 结尾 |
| `~*` | 正则匹配，不区分大小写 | `location ~* \.(jpg|png)$` |
| `!~` | 正则不匹配，区分大小写 | `location !~ \.php$` |
| `!~*` | 正则不匹配，不区分大小写 | `location !~* \.jpg$` |
| 无 | 普通前缀匹配 | `location /api/` 匹配 /api/... |

**常见陷阱**：

```nginx
# 陷阱1：alias 末尾斜杠
location /static/ {
    alias /var/www/static/;  # 正确：末尾有斜杠
    # alias /var/www/static;   # 错误：访问 /static/file 会映射到 /var/www/staticfile
}

# 陷阱2：root 和 alias 区别
location /static/ {
    root /var/www;           # 访问 /static/file → /var/www/static/file
    alias /var/www/static/;  # 访问 /static/file → /var/www/static/file
}

# 陷阱3：try_files 使用
location / {
    try_files $uri $uri/ /index.html;  # 先尝试文件，再尝试目录，最后fallback
    # try_files $uri $uri/ =404;        # 找不到返回404
}

# 陷阱4：proxy_pass 末尾斜杠
location /api/ {
    proxy_pass http://backend/;    # 访问 /api/users → http://backend/users
    # proxy_pass http://backend;   # 访问 /api/users → http://backend/api/users
}
```

### 10. HTTPS 证书问题深度排查

**症状**：证书过期、证书链不完整、浏览器警告

**诊断流程**：
```bash
# 1. 检查证书信息
openssl x509 -in /etc/nginx/ssl/cert.pem -noout -text

# 2. 检查证书链完整性
openssl crl2pkcs7 -nocrl -certfile /etc/nginx/ssl/cert.pem | openssl pkcs7 -print_certs -noout

# 3. 在线检测
# https://www.ssllabs.com/ssltest/
# https://whatsmychaincert.com/

# 4. 检查证书过期时间（提前30天告警）
EXPIRY=$(openssl x509 -in /etc/nginx/ssl/cert.pem -noout -enddate | cut -d= -f2)
EXPIRY_EPOCH=$(date -d "$EXPIRY" +%s)
NOW_EPOCH=$(date +%s)
DAYS_LEFT=$(( ($EXPIRY_EPOCH - $NOW_EPOCH) / 86400 ))
echo "证书剩余有效期: $DAYS_LEFT 天"

# 5. 验证私钥匹配
openssl x509 -noout -modulus -in /etc/nginx/ssl/cert.pem | openssl md5
openssl rsa -noout -modulus -in /etc/nginx/ssl/key.pem | openssl md5
```

**证书配置最佳实践**：

```nginx
server {
    listen 443 ssl http2;
    server_name example.com;

    # 证书配置
    ssl_certificate /etc/nginx/ssl/fullchain.pem;  # 包含中间证书
    ssl_certificate_key /etc/nginx/ssl/privkey.pem;

    # 证书链验证
    ssl_trusted_certificate /etc/nginx/ssl/chain.pem;
    ssl_stapling on;
    ssl_stapling_verify on;

    # 安全协议
    ssl_protocols TLSv1.2 TLSv1.3;

    # 加密套件（安全性优先）
    ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-CHACHA20-POLY1305:ECDHE-RSA-CHACHA20-POLY1305:DHE-RSA-AES128-GCM-SHA256:DHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers off;

    # 会话恢复
    ssl_session_cache shared:SSL:50m;
    ssl_session_timeout 1d;
    ssl_session_tickets off;

    # DH 参数（用于 DHE 加密套件）
    ssl_dhparam /etc/nginx/ssl/dhparam.pem;

    # ECDH 曲线
    ssl_ecdh_curve X25519:secp384r1:secp256k1;

    # 安全响应头
    add_header Strict-Transport-Security "max-age=63072000" always;
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
}

# HTTP 强制跳转 HTTPS
server {
    listen 80;
    server_name example.com;
    return 301 https://$server_name$request_uri;
}
```

**Let's Encrypt 自动续期**：
```bash
#!/bin/bash
# certbot_renew.sh

# 测试续期
certbot renew --dry-run

# 实际续期并重启 Nginx
certbot renew --post-hook "systemctl reload nginx"

# 添加到 crontab（每天执行）
# 0 3 * * * /path/to/certbot_renew.sh >> /var/log/certbot-renew.log 2>&1
```
```

**处理方案**：
```bash
# 1. 修改系统限制
# /etc/security/limits.conf
* soft nofile 65535
* hard nofile 65535

# 2. 修改 Nginx 配置
# nginx.conf
worker_rlimit_nofile 65535;

# 3. 减少 keepalive 连接
keepalive_timeout 30;
keepalive_requests 100;
```

## 性能优化配置

```nginx
# /etc/nginx/nginx.conf
user nginx;
worker_processes auto;           # 根据 CPU 核心数
worker_rlimit_nofile 65535;      # 文件描述符限制

error_log /var/log/nginx/error.log warn;
pid /var/run/nginx.pid;

events {
    worker_connections 4096;     # 每个 worker 的连接数
    use epoll;                   # Linux 高性能事件模型
    multi_accept on;             # 同时接受多个连接
}

http {
    include /etc/nginx/mime.types;
    default_type application/octet-stream;

    # 日志格式
    log_format main '$remote_addr - $remote_user [$time_local] "$request" '
                    '$status $body_bytes_sent "$http_referer" '
                    '"$http_user_agent" "$http_x_forwarded_for" '
                    'rt=$request_time uct="$upstream_connect_time" '
                    'uht="$upstream_header_time" urt="$upstream_response_time"';

    access_log /var/log/nginx/access.log main buffer=32k flush=5s;

    # 性能优化
    sendfile on;
    tcp_nopush on;
    tcp_nodelay on;
    keepalive_timeout 65;
    types_hash_max_size 2048;

    # Gzip 压缩
    gzip on;
    gzip_vary on;
    gzip_proxied any;
    gzip_comp_level 6;
    gzip_types text/plain text/css text/xml application/json
               application/javascript application/rss+xml
               application/atom+xml image/svg+xml;

    # 虚拟主机配置
    include /etc/nginx/conf.d/*.conf;
}
```

## 负载均衡配置

```nginx
upstream backend {
    least_conn;                    # 最少连接算法
    server 192.168.1.10:8080 weight=5 max_fails=3 fail_timeout=30s;
    server 192.168.1.11:8080 weight=5 max_fails=3 fail_timeout=30s;
    server 192.168.1.12:8080 backup;  # 备用服务器
    keepalive 32;                  # 长连接池
}

server {
    listen 80;
    server_name example.com;

    location / {
        proxy_pass http://backend;
        proxy_http_version 1.1;
        proxy_set_header Connection "";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;

        # 超时配置
        proxy_connect_timeout 5s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;

        # 缓冲配置
        proxy_buffering on;
        proxy_buffer_size 4k;
        proxy_buffers 8 4k;
    }
}
```

## 安全加固

```nginx
# 隐藏版本号
server_tokens off;

# 限流配置
limit_req_zone $binary_remote_addr zone=req_limit:10m rate=10r/s;
limit_conn_zone $binary_remote_addr zone=conn_limit:10m;

server {
    # 限流应用
    limit_req zone=req_limit burst=20 nodelay;
    limit_conn conn_limit 10;

    # 阻止常见攻击
    if ($http_user_agent ~* (bot|crawler|spider)) {
        return 403;
    }

    # IP 黑名单
    deny 192.168.1.100;
    allow 192.168.1.0/24;
    deny all;

    # 防止点击劫持
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
}
```

## 日志分析

```bash
# 查看 Top IP (Linux)
cat access.log | awk '{print $1}' | sort | uniq -c | sort -rn | head -20

# 查看 Top URL (Linux)
cat access.log | awk '{print $7}' | sort | uniq -c | sort -rn | head -20

# 查看状态码分布 (Linux)
cat access.log | awk '{print $9}' | sort | uniq -c | sort -rn

# 查看慢请求（超过 5 秒）(Linux)
cat access.log | awk -F'rt=' '{print $2}' | awk '{if($1>5)print $1}' | sort -rn | head -20

# 使用 GoAccess 实时分析
goaccess /var/log/nginx/access.log -o report.html --real-time-html
```

**PowerShell (Windows)**:
```powershell
# 查看 Top IP (Windows)
Get-Content C:\nginx\logs\access.log | ForEach-Object { ($_ -split " ")[0] } | Group-Object | Sort-Object Count -Descending | Select-Object -First 20

# 查看 Top URL (Windows)
Get-Content C:\nginx\logs\access.log | ForEach-Object { ($_ -split " ")[6] } | Group-Object | Sort-Object Count -Descending | Select-Object -First 20

# 查看状态码分布 (Windows)
Get-Content C:\nginx\logs\access.log | ForEach-Object { ($_ -split " ")[8] } | Group-Object | Sort-Object Count -Descending

# 查看慢请求（超过 5 秒）(Windows)
# 假设日志格式包含 rt= 字段
Get-Content C:\nginx\logs\access.log | Where-Object { $_ -match "rt=(\d+\.?\d*)" } | ForEach-Object {
    if ($matches[1] -gt 5) { $_ }
}

# 使用 GoAccess 实时分析 (Windows 版本可用)
goaccess C:\nginx\logs\access.log -o report.html --real-time-html
```

## 监控指标

```bash
# 启用 stub_status
location /nginx_status {
    stub_status on;
    access_log off;
    allow 127.0.0.1;
    deny all;
}

# 关键指标
Active connections: 291          # 当前活动连接
server accepts handled requests   # 总共接受的连接/处理的连接/请求数
Reading: 6 Writing: 125 Waiting: 160  # 读/写/等待状态
```

## 生产环境最佳实践

### 1. 部署架构建议

**小型网站（<1000 QPS）**：
- 单台 Nginx + 应用服务器
- 启用 Gzip 压缩和浏览器缓存

**中型网站（1000-10000 QPS）**：
- 多台 Nginx 负载均衡（LVS/HAProxy 或 DNS 轮询）
- 动静分离，静态资源走 CDN
- 启用缓存层（Nginx proxy_cache 或 Varnish）

**大型网站（>10000 QPS）**：
- 多层负载均衡架构
- 专用 Nginx 集群处理不同业务
- 边缘节点部署（配合 CDN）

### 2. 完整生产配置模板

```nginx
# /etc/nginx/nginx.conf
user nginx;
worker_processes auto;
worker_rlimit_nofile 65535;

error_log /var/log/nginx/error.log warn;
pid /var/run/nginx.pid;

events {
    worker_connections 4096;
    use epoll;
    multi_accept on;
}

http {
    include /etc/nginx/mime.types;
    default_type application/octet-stream;

    # 日志格式
    log_format main '$remote_addr - $remote_user [$time_local] "$request" '
                    '$status $body_bytes_sent "$http_referer" '
                    '"$http_user_agent" "$http_x_forwarded_for" '
                    'rt=$request_time uct="$upstream_connect_time" '
                    'uht="$upstream_header_time" urt="$upstream_response_time" '
                    'cs=$upstream_cache_status';

    access_log /var/log/nginx/access.log main buffer=32k flush=5s;

    # 性能优化
    sendfile on;
    tcp_nopush on;
    tcp_nodelay on;
    keepalive_timeout 65;
    keepalive_requests 1000;
    types_hash_max_size 2048;

    # 文件上传限制
    client_max_body_size 100m;
    client_body_buffer_size 128k;

    # Gzip 压缩
    gzip on;
    gzip_vary on;
    gzip_proxied any;
    gzip_comp_level 6;
    gzip_min_length 1000;
    gzip_types text/plain text/css text/xml application/json
               application/javascript application/rss+xml
               application/atom+xml image/svg+xml;

    # 限流配置
    limit_req_zone $binary_remote_addr zone=req_limit:10m rate=10r/s;
    limit_conn_zone $binary_remote_addr zone=conn_limit:10m;

    # 缓存配置
    proxy_cache_path /var/cache/nginx levels=1:2 keys_zone=cache_zone:100m
                     max_size=10g inactive=60m use_temp_path=off;

    # 上游服务器
    upstream backend {
        least_conn;
        server 10.0.0.1:8080 weight=5 max_fails=3 fail_timeout=30s;
        server 10.0.0.2:8080 weight=5 max_fails=3 fail_timeout=30s;
        server 10.0.0.3:8080 backup;
        keepalive 32;
    }

    # 虚拟主机
    include /etc/nginx/conf.d/*.conf;
}
```

### 3. HTTPS 完整配置

```nginx
server {
    listen 443 ssl http2;
    server_name example.com;

    # SSL 证书
    ssl_certificate /etc/nginx/ssl/cert.pem;
    ssl_certificate_key /etc/nginx/ssl/key.pem;
    ssl_trusted_certificate /etc/nginx/ssl/chain.pem;

    # SSL 优化
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers off;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 1d;
    ssl_session_tickets off;
    ssl_stapling on;
    ssl_stapling_verify on;

    # 安全响应头
    add_header Strict-Transport-Security "max-age=63072000" always;
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;

    # 日志
    access_log /var/log/nginx/example.com.access.log main;
    error_log /var/log/nginx/example.com.error.log warn;

    # 静态资源缓存
    location ~* \.(jpg|jpeg|png|gif|ico|css|js|svg)$ {
        expires 30d;
        add_header Cache-Control "public, immutable";
        access_log off;
    }

    # 反向代理
    location / {
        proxy_pass http://backend;
        proxy_http_version 1.1;
        proxy_set_header Connection "";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        proxy_connect_timeout 5s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;

        proxy_buffering on;
        proxy_buffer_size 4k;
        proxy_buffers 8 4k;
        proxy_busy_buffers_size 8k;

        # 限流
        limit_req zone=req_limit burst=20 nodelay;
        limit_conn conn_limit 10;
    }
}

# HTTP 重定向到 HTTPS
server {
    listen 80;
    server_name example.com;
    return 301 https://$server_name$request_uri;
}
```

### 4. 日志分析脚本

```bash
#!/bin/bash
# nginx_log_analysis.sh

LOG_FILE="/var/log/nginx/access.log"
DATE=$(date +%Y%m%d)
REPORT_DIR="/var/log/nginx/reports"
mkdir -p $REPORT_DIR

echo "=== Nginx 访问日志分析报告 - $DATE ===" > $REPORT_DIR/report_$DATE.txt

# 总请求数
echo -e "\n[总请求数]" >> $REPORT_DIR/report_$DATE.txt
total_requests=$(wc -l < $LOG_FILE)
echo "总请求数: $total_requests" >> $REPORT_DIR/report_$DATE.txt

# Top 10 IP
echo -e "\n[Top 10 IP]" >> $REPORT_DIR/report_$DATE.txt
awk '{print $1}' $LOG_FILE | sort | uniq -c | sort -rn | head -10 >> $REPORT_DIR/report_$DATE.txt

# Top 10 URL
echo -e "\n[Top 10 URL]" >> $REPORT_DIR/report_$DATE.txt
awk '{print $7}' $LOG_FILE | sort | uniq -c | sort -rn | head -10 >> $REPORT_DIR/report_$DATE.txt

# 状态码分布
echo -e "\n[状态码分布]" >> $REPORT_DIR/report_$DATE.txt
awk '{print $9}' $LOG_FILE | sort | uniq -c | sort -rn >> $REPORT_DIR/report_$DATE.txt

# 慢请求 Top 10
echo -e "\n[慢请求 Top 10 (>5s)]" >> $REPORT_DIR/report_$DATE.txt
awk -F'rt=' '{print $2}' $LOG_FILE | awk '{if($1>5)print $1}' | sort -rn | head -10 >> $REPORT_DIR/report_$DATE.txt

# 4xx/5xx 错误
echo -e "\n[错误请求]" >> $REPORT_DIR/report_$DATE.txt
awk '$9 >= 400 {print $0}' $LOG_FILE | head -20 >> $REPORT_DIR/report_$DATE.txt

echo "报告已生成: $REPORT_DIR/report_$DATE.txt"
```

### 5. 监控检查脚本

```bash
#!/bin/bash
# nginx_health_check.sh

NGINX_STATUS_URL="http://localhost/nginx_status"
ALERT_THRESHOLD_CONNECTIONS=10000
ALERT_THRESHOLD_REQUEST_TIME=5

# 检查 Nginx 进程
if ! pgrep nginx > /dev/null; then
    echo "ALERT: Nginx 进程不存在"
    exit 1
fi

# 检查配置语法
if ! nginx -t > /dev/null 2>&1; then
    echo "ALERT: Nginx 配置语法错误"
    nginx -t
    exit 1
fi

# 获取连接状态
STATUS=$(curl -s $NGINX_STATUS_URL 2>/dev/null)
if [ -z "$STATUS" ]; then
    echo "WARNING: 无法获取 Nginx 状态"
else
    ACTIVE=$(echo "$STATUS" | grep "Active connections" | awk '{print $3}')
    if [ "$ACTIVE" -gt "$ALERT_THRESHOLD_CONNECTIONS" ]; then
        echo "ALERT: 活动连接数过高: $ACTIVE"
    else
        echo "OK: 活动连接数: $ACTIVE"
    fi
fi

# 检查错误日志增长
ERROR_COUNT=$(tail -100 /var/log/nginx/error.log 2>/dev/null | grep -c "error")
if [ "$ERROR_COUNT" -gt 50 ]; then
    echo "WARNING: 错误日志增长过快: $ERROR_COUNT 条/100行"
fi

echo "健康检查完成"
```

**PowerShell (Windows)**:
```powershell
# nginx_health_check.ps1

$NGINX_STATUS_URL = "http://localhost/nginx_status"
$ALERT_THRESHOLD_CONNECTIONS = 10000

# 检查 Nginx 进程
$nginxProcess = Get-Process | Where-Object {$_.Name -like "*nginx*"}
if (-not $nginxProcess) {
    Write-Host "ALERT: Nginx 进程不存在" -ForegroundColor Red
    exit 1
}

# 检查配置语法
try {
    $configTest = nginx -t 2>&1
    if ($configTest -match "failed") {
        Write-Host "ALERT: Nginx 配置语法错误" -ForegroundColor Red
        Write-Host $configTest
        exit 1
    }
} catch {
    Write-Host "WARNING: 无法测试配置: $_" -ForegroundColor Yellow
}

# 获取连接状态
try {
    $status = Invoke-WebRequest -Uri $NGINX_STATUS_URL -UseBasicParsing -ErrorAction SilentlyContinue
    if ($status) {
        $activeConnections = [regex]::Match($status.Content, "Active connections: (\d+)").Groups[1].Value
        if ([int]$activeConnections -gt $ALERT_THRESHOLD_CONNECTIONS) {
            Write-Host "ALERT: 活动连接数过高: $activeConnections" -ForegroundColor Red
        } else {
            Write-Host "OK: 活动连接数: $activeConnections" -ForegroundColor Green
        }
    }
} catch {
    Write-Host "WARNING: 无法获取 Nginx 状态" -ForegroundColor Yellow
}

# 检查错误日志增长
$errorLogPath = "C:\nginx\logs\error.log"
if (Test-Path $errorLogPath) {
    $errorCount = (Get-Content $errorLogPath -Tail 100 | Select-String -Pattern "error").Count
    if ($errorCount -gt 50) {
        Write-Host "WARNING: 错误日志增长过快: $errorCount 条/100行" -ForegroundColor Yellow
    }
}

Write-Host "健康检查完成"
```

## 参考资料

### 官方文档
- [Nginx 官方文档](https://nginx.org/en/docs/)
- [Nginx 管理员指南](https://nginx.org/en/docs/beginners_guide.html)
- [Nginx 模块参考](https://nginx.org/en/docs/ngx_core_module.html)

### 社区资源
- [Nginx 博客](https://www.nginx.com/blog/)
- [Nginx 性能调优](https://www.nginx.com/blog/tuning-nginx/)
- [Nginx 安全最佳实践](https://www.nginx.com/blog/nginx-security-best-practices/)

### 工具推荐
- **GoAccess**: 实时 Web 日志分析器
- **ngxtop**: Nginx 日志实时监控工具
- **nginx-vts**: Nginx 虚拟主机流量状态模块
- **certbot**: Let's Encrypt SSL 证书自动管理

## MCP 工具支持

本 Skill 可通过 MCP (Model Context Protocol) 提供以下工具：

### 工具列表

| 工具名称 | 描述 | 必需参数 |
|---------|------|---------|
| `nginx_check_status` | 检查 Nginx 运行状态 | 无 |
| `nginx_test_config` | 测试配置文件语法 | 无 |
| `nginx_get_connections` | 获取连接状态统计 | 无 |
| `nginx_get_error_logs` | 获取错误日志 | 无 |
| `nginx_analyze_access_logs` | 分析访问日志 | 无 |

### 工具定义示例

```json
{
  "name": "nginx_check_status",
  "description": "检查 Nginx 运行状态和版本信息",
  "inputSchema": {
    "type": "object",
    "properties": {}
  }
}
```

```json
{
  "name": "nginx_test_config",
  "description": "测试 Nginx 配置文件语法",
  "inputSchema": {
    "type": "object",
    "properties": {
      "config_path": { "type": "string", "description": "配置文件路径", "default": "/etc/nginx/nginx.conf" }
    }
  }
}
```

```json
{
  "name": "nginx_get_connections",
  "description": "获取 Nginx 连接状态统计（需要启用 stub_status）",
  "inputSchema": {
    "type": "object",
    "properties": {
      "status_url": { "type": "string", "default": "http://localhost/nginx_status" }
    }
  }
}
```

```json
{
  "name": "nginx_get_error_logs",
  "description": "获取 Nginx 错误日志",
  "inputSchema": {
    "type": "object",
    "properties": {
      "log_path": { "type": "string", "default": "/var/log/nginx/error.log" },
      "lines": { "type": "integer", "default": 50 }
    }
  }
}
```

```json
{
  "name": "nginx_analyze_access_logs",
  "description": "分析 Nginx 访问日志，获取 Top IP、URL、状态码等",
  "inputSchema": {
    "type": "object",
    "properties": {
      "log_path": { "type": "string", "default": "/var/log/nginx/access.log" },
      "analysis_type": { "type": "string", "enum": ["top_ip", "top_url", "status_codes", "slow_requests"], "default": "top_ip" },
      "limit": { "type": "integer", "default": 20 }
    }
  }
}
```

### Python MCP Server 示例

```python
from mcp.server import Server
from mcp.types import TextContent
import subprocess

app = Server("nginx-ops")

@app.call_tool()
def call_tool(name: str, arguments: dict):
    if name == "nginx_check_status":
        commands = [
            "nginx -v",
            "systemctl status nginx 2>/dev/null || service nginx status 2>/dev/null || echo 'Service status unavailable'",
            "ps aux | grep 'nginx' | grep -v grep | head -5"
        ]
        results = []
        for cmd in commands:
            result = subprocess.run(cmd, shell=True, capture_output=True, text=True)
            results.append(result.stdout)
        return [TextContent(type="text", text="\n".join(results))]

    elif name == "nginx_test_config":
        config = arguments.get("config_path", "/etc/nginx/nginx.conf")
        cmd = f"nginx -t -c {config}"
        result = subprocess.run(cmd, shell=True, capture_output=True, text=True)
        return [TextContent(type="text", text=result.stderr or result.stdout)]

    elif name == "nginx_get_connections":
        url = arguments.get("status_url", "http://localhost/nginx_status")
        cmd = f"curl -s {url} 2>/dev/null || echo 'stub_status not available'"
        result = subprocess.run(cmd, shell=True, capture_output=True, text=True)
        return [TextContent(type="text", text=result.stdout)]

    elif name == "nginx_get_error_logs":
        log_path = arguments.get("log_path", "/var/log/nginx/error.log")
        lines = arguments.get("lines", 50)
        cmd = f"tail -n {lines} {log_path} 2>/dev/null || echo 'Log file not found'"
        result = subprocess.run(cmd, shell=True, capture_output=True, text=True)
        return [TextContent(type="text", text=result.stdout)]

    elif name == "nginx_analyze_access_logs":
        log_path = arguments.get("log_path", "/var/log/nginx/access.log")
        analysis = arguments.get("analysis_type", "top_ip")
        limit = arguments.get("limit", 20)

        if analysis == "top_ip":
            cmd = f"awk '{{print $1}}' {log_path} | sort | uniq -c | sort -rn | head -{limit}"
        elif analysis == "top_url":
            cmd = f"awk '{{print $7}}' {log_path} | sort | uniq -c | sort -rn | head -{limit}"
        elif analysis == "status_codes":
            cmd = f"awk '{{print $9}}' {log_path} | sort | uniq -c | sort -rn"
        elif analysis == "slow_requests":
            cmd = f"awk -F'rt=' '{{print $2}}' {log_path} | awk '{{if($1>5)print $1}}' | sort -rn | head -{limit}"
        else:
            cmd = f"echo 'Unknown analysis type'"

        result = subprocess.run(cmd, shell=True, capture_output=True, text=True)
        return [TextContent(type="text", text=result.stdout)]

if __name__ == "__main__":
    app.run()
```

## 输出规范

```
🌐 Nginx 诊断报告

📊 基本信息
- 版本：[version]
- 运行状态：[active/running]
- Worker 进程数：[worker_processes]
- 当前连接数：[Active connections]

🔍 问题分析
[具体问题]

💡 解决方案
[处理步骤]

📋 优化建议
- [建议1]
- [建议2]
```
