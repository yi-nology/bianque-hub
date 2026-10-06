# oo-devops（OpenOcta 生态运维技能包）

OpenOcta 生态归化包：31 位数字员工（专家）+ 189 项运维技能 + 员工 MCP server 包声明
（缺省硬关，装包≠开洞；激活员工 MCP = 站点 conf 加同名条目 enabled: true）。

## 形态与安装（2026-10-05 A 案拍板：hub 托管源，不进 seed-sync 白名单）

- 本目录是**内容唯一编辑点**（原主仓 market/mirror+convert 收割管线已随 A 案退役，
  内容改动直接编辑本目录并走 hub 评审）。
- **不随 bianque 二进制缺省分发**：seed-sync 白名单（缺省 os-basics）不含本包——
  保持缺省足迹，与「装包≠开洞」治理哲学一致。
- 安装：打包本目录为 zip（顶层=包内容）→ 控制台「安装 bundle」路径安装（原子激活，
  校验失败自动回滚）；升级同路径。215 现网 2026-09-22 core 装机件曾为本包同源产物——
  其后整包已从现网卸载、由四领域包（cicd/db/mw/obs）承接，本包仅作参考库与源。
- 校验：hub validator（仓根 `go run ./validator ./packs`，CI 同款；pack-lint 旧管线已随
  A 案退役）。

## 生成溯源

内容由主仓 harvest/convert 管线自 OpenOcta 注册表归化生成后一次性迁入（2026-10-05，
迁出时实测 31 专家/189 技能/7 MCP 声明）；管线退役后本包即终态源。
