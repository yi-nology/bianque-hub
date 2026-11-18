// disclosure.go —— 凭证前提披露面判据（批607 披露面普及轮）。
//
// requires_credentials 是**警告级**声明（缺失只提示不阻塞装包），但它回答的是
// 「装这个包前管理员该准备什么」——消费了专用凭证面却不声明，披露面就有洞：
// 站点装包后采集全空、排障先撞「凭证未配」。声明面靠 policy，不靠自觉：
// 本判据把「可机判」的 server→凭证类型映射钉住，漏声明 CI 先拦（WARN 级，
// 不拦发布——与 requires_credentials 的警告级语义一致）。
//
// 映射范围刻意窄：只登记**可机判**的专用面（datasources→datasources、
// k8sgpt→kubeconfig）。ssh 等默认采集面不计（os-basics 口径：主机域零凭证
// 依赖——ask-ops 对目标机的采集凭证是实例采集配置，不是包级前提）。
// 新增专用面时在此登记映射 + 对应包补声明，两处成对。
package main

// serverCredentialPrereq server → 凭证类型前提映射（可机判子集）。
// 单源镜像：主仓 internal/credentials/defs/（kubeconfig.yaml/datasources.yaml）
// 与 conf 挂载面 mcp.servers 词表。改任一侧须对表。
var serverCredentialPrereq = map[string]string{
	"datasources": "datasources", // datasources 采集面消费 datasources 类型凭证
	"k8sgpt":      "kubeconfig",  // k8sgpt sidecar 挂目标集群 kubeconfig
}

// checkPackCredentialDisclosure 包内专家 tools 消费了映射内 server 而未声明
// 对应凭证类型前提 ⇒ WARN（披露面缺口，不拦发布）。
func (v *validator) checkPackCredentialDisclosure(packName string, declared []string, servers map[string]bool, add func(level, format string, a ...any)) {
	set := map[string]bool{}
	for _, d := range declared {
		set[d] = true
	}
	for server := range servers {
		cred, ok := serverCredentialPrereq[server]
		if !ok || set[cred] {
			continue
		}
		add("WARN", "专家 tools 引用 %s server 但 pack.yaml 未声明 requires_credentials: [%s]（披露面缺口：装包后采集全空、排障先撞凭证未配；批607 判据）", server, cred)
	}
}
