---
name: nginx-tls-triage
description: Nginx TLS 面分诊方法论：握手失败的证书链缺中间证书最高频、openssl s_client 链完整性判读、证书有效期与 SAN 核对、SNI 多证书匹配、混合内容——固定顺序定位与判读基准（ask-ops 只读采集面）。
mode: on_demand
version: 0.1.0
maturity: experimental
requires_mcp:
  - server: ask-ops
    tools: [run_readonly_command, run_readonly_commands]
---

## 触发条件

- 症状关键词：https 打不开、SSL 握手失败、证书链异常、NET::ERR_CERT_AUTHORITY_INVALID、
  混合内容、证书过期告警、SNI 证书错、curl 证书报错
- 组合场景：浏览器报错但 curl 正常=SNI/客户端信任库差异；全站握手失败先看 443 在听与
  证书文件路径（转 nginx-availability-triage 第一步）

## 数据来源（ask-ops 只读面）

1. 端到端握手复现（本机直连，剥离 CDN）：
   `openssl s_client -connect 127.0.0.1:443 -servername 站点域名 </dev/null 2>/dev/null | head -30`
   ——`Verify return code: 21 (unable to verify the first certificate)`=缺中间证书
   （最高频：部署时只换了叶子证书）；`19 (self-signed)`=自签未入信任库；
2. 证书内容核对：`openssl x509 -in /etc/nginx/ssl/站点.crt -noout -subject -dates -ext subjectAltName`
   ——有效期、域名是否在 SAN（只看 CN 的老证书对现代客户端=无效）；
3. 链序核对：`openssl s_client -connect ... -showcerts | grep -c 'BEGIN CERT'`
   ——应 ≥2（叶子+中间）；只有 1 张=链缺；
4. nginx 侧证书配置：`nginx -T | grep -B 2 -A 3 'ssl_certificate'`——证书路径/每 server
   块 SNI 对应关系（默认 server 的证书会被不带 SNI 的客户端拿到）；
5. 混合内容：页面 https 打开但控制台报 mixed content——`curl -s https://站点/路径 |
   grep -o 'src="http://[^"]*"' | head`（页面内 http:// 资源引用）。

## 判读基准

- `unable to verify the first certificate`：服务器只发了叶子证书——把中间证书按序拼进
  fullchain（叶子在前中间在后；root 不需要），reload 生效（变更面，审批）；
- 域名不匹配/不在 SAN：证书与访问域名不符——多站点 SNI 配错 server 块，或证书签发时
  漏域名（重新签发，变更面）；
- expires 临期：续期后 reload（acme 场景核查定时任务与 reload hook 是否还在）；
- 混合内容：改页面资源引用为 https 或协议相对——前端面修复，nginx 侧可加
  `Content-Security-Policy: upgrade-insecure-requests` 头过渡（变更面，审批）。

## 处置面（均审批后宿主侧执行）

- 证书文件替换 → `nginx -t`（校验证书/私钥匹配：`openssl x509 -noout -modulus -in crt`
  与 `openssl rsa -noout -modulus -in key` 指纹一致）→ `nginx -s reload`；
- 变更前旧证书目录整体备份留回滚点。
