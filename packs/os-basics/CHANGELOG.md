# CHANGELOG

## 1.3.0 (2026-09-30)

- 社区版首发（bianque-hub 上游化）：本包成为内置同名核心包的上游，slug/路由/技能
  与内置版完全等位，升级即原位接管；平台速查段（platform-matrix）按契约引用平台
  `_shared/` 共享层，参考副本随仓 knowledge/ 分发供非扁鹊运行时使用。
- 新增 chain.yaml：workflow/host-quick-audit 主机快查链（安全基线 + 内存进程
  事实两步；P3，路由词避开全域巡检域词）。

## 1.2.0 (2026-09-24)

- 合并 k8s-pack：specialists/k8s-health（P2，k8sgpt 分析桥）+ k8s-node-diagnosis 技能并入本包，
  插件页收敛为「内置核心（os-basics）+ 市场包（oo-devops）」两层；slug 不变，调度链绑定
  （scheduler/flows.go 巡检扇出）不受影响。原 k8s-pack 1.0.0~1.1.0 历史并入本包线。
- analyzers.yaml 合并 sysprobe 域映射与 k8sgpt 检查源声明；kubeconfig 凭证前提仅 K8s 域消费。

## 1.1.1 (2026-09-24)

- 门面定位化：本包为主机/OS 域唯一入口（oo-devops 主机巡检/资源监控类技能已让位收敛至此）。

## 1.1.0 (2026-09-22)

- 统一包格式：补 provides/changelog；安全域 4 技能（ssh 爆破/-root/sudo/进程）与 6 域专家。
