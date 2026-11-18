// 本文件由批607 validator 重构拆出（纯代码搬移，判据逐字未动）。
package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

// checkReadme 核对 README 包清单版本表与各 pack.yaml 同步（手工表格是版本纪律
// 最常漂移之处）。约定在仓库根运行；找不到 README.md 时降级 WARN 跳过。
func (v *validator) checkReadme() {
	if len(v.packVersions) == 0 {
		return
	}
	raw, err := os.ReadFile("README.md")
	if err != nil {
		v.add("WARN", "README", "未在当前目录找到 README.md，跳过包清单版本核对（应在仓库根运行）")
		return
	}
	table := map[string]string{}
	for _, m := range readmeRowRe.FindAllStringSubmatch(string(raw), -1) {
		table[m[1]] = m[2]
	}
	if len(table) == 0 {
		v.add("WARN", "README", "README 未识别到包清单表行（| [包名](packs/包名/) | 版本 |），跳过版本核对")
		return
	}
	names := make([]string, 0, len(v.packVersions))
	for name := range v.packVersions {
		names = append(names, name)
	}
	slices.Sort(names)
	for _, name := range names {
		ver := v.packVersions[name]
		rv, ok := table[name]
		if !ok {
			v.add("WARN", name, "README 包清单缺 %s 行", name)
			continue
		}
		if rv != ver {
			v.add("ERROR", name, "README 包清单版本 %s 与 pack.yaml %s 不一致（发版须同步 README 表；README 清单表是 hub 仓级资产，校验内嵌快照请用 --mode=embedded）", rv, ver)
		}
	}
}

// checkIndex index.json 对账（批次一百三十九）：索引是 bq-markettool search/info 的
// 机器可读事实源（Claude Code marketplace.json 同构物）。缺失=WARN（渐进采用，检索
// 侧自动降级现场扫描）；在而漂移=ERROR——多录/漏包/版本不符逼 CI 重新
// `bq-markettool index-gen` 后入库，安装正确性不依赖索引（install 仍现扫目录）。
// 包清单自本容器目录现扫（不依赖跨根累积的 packVersions——多根调用互不污染）。
//
// 索引是仓级资产，恒取 packs 容器的父目录（仓根）。packsRoot 先 Clean 再 Dir：
// 带尾斜杠的 "./packs/" 不归一时 Dir("./packs/")="packs"，索引被找成
// packs/index.json——读不到即走 IsNotExist 分支只 WARN 便 return，ERROR 级
// 对账（漏包/多录/版本漂移）整段被静默跳过，退出码 0 假绿（issue #3）。

// checkIndex index.json 对账（批次一百三十九）：索引是 bq-markettool search/info 的
// 机器可读事实源（Claude Code marketplace.json 同构物）。缺失=WARN（渐进采用，检索
// 侧自动降级现场扫描）；在而漂移=ERROR——多录/漏包/版本不符逼 CI 重新
// `bq-markettool index-gen` 后入库，安装正确性不依赖索引（install 仍现扫目录）。
// 包清单自本容器目录现扫（不依赖跨根累积的 packVersions——多根调用互不污染）。
//
// 索引是仓级资产，恒取 packs 容器的父目录（仓根）。packsRoot 先 Clean 再 Dir：
// 带尾斜杠的 "./packs/" 不归一时 Dir("./packs/")="packs"，索引被找成
// packs/index.json——读不到即走 IsNotExist 分支只 WARN 便 return，ERROR 级
// 对账（漏包/多录/版本漂移）整段被静默跳过，退出码 0 假绿（issue #3）。
func (v *validator) checkIndex(packsRoot string) {
	raw, err := os.ReadFile(filepath.Join(filepath.Dir(filepath.Clean(packsRoot)), "index.json"))
	if err != nil {
		if os.IsNotExist(err) {
			v.add("WARN", "index", "index.json 缺失（仓根）——跑 `bq-markettool index-gen -hub <本仓>` 生成并提交（检索面走索引，缺省降级现场扫描；校验内嵌快照请用 --mode=embedded）")
			return
		}
		v.add("ERROR", "index", "读取 index.json: %v", err)
		return
	}
	var idx struct {
		Packs []struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"packs"`
	}
	if err := json.Unmarshal(raw, &idx); err != nil {
		v.add("ERROR", "index", "index.json 解析失败: %v", err)
		return
	}
	onDisk := map[string]string{}
	entries, err := os.ReadDir(packsRoot)
	if err != nil {
		return // run 的容器发现已报，此处不双报
	}
	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") || strings.HasPrefix(e.Name(), "_") {
			continue
		}
		rawManifest, err := os.ReadFile(filepath.Join(packsRoot, e.Name(), "pack.yaml"))
		if err != nil {
			continue // 非包目录不校验（与 run 同口径）
		}
		var m packManifest
		if yaml.Unmarshal(rawManifest, &m) != nil || m.Name == "" {
			continue
		}
		onDisk[m.Name] = m.Version
	}
	seen := map[string]bool{}
	for _, p := range idx.Packs {
		if seen[p.Name] {
			v.add("ERROR", "index", "index.json 重复条目 %s", p.Name)
			continue
		}
		seen[p.Name] = true
		ver, ok := onDisk[p.Name]
		if !ok {
			v.add("ERROR", "index", "index.json 多录 %s（packs/ 无此包——重跑 index-gen 刷新）", p.Name)
			continue
		}
		if ver != "" && ver != p.Version {
			v.add("ERROR", p.Name, "index.json 版本漂移：索引 %s / pack.yaml %s（重跑 index-gen 刷新）", p.Version, ver)
		}
	}
	for name := range onDisk {
		if !seen[name] {
			v.add("ERROR", "index", "index.json 漏包 %s（重跑 index-gen 刷新）", name)
		}
	}
}

// dangerousPatterns 发布侧内容扫描（批次一百四十，G4b；ClawHub「SKILL.md 变安装器」
// 投毒案对标）：技能/专家正文是喂给 LLM 的执行面指令。分档依据实弹校准——
// `curl|sh` 在运维指引正文有合法形态（helm/tidb 等工具安装，oo-devops 现存 4 处）
// → 正文 WARN；**changes/ 变更块是受审执行通道的实际执行面 → 任何命中升 ERROR**
// （声明层零容忍）。启发式有边界：只拦明文形态，变形/编码载荷靠平台执行面治理
// （受审命令+审批门）兜底——扫描是内容审的第一道网，不是替代执行面治理。
var dangerousPatterns = []struct {
	re    *regexp.Regexp
	level string
	why   string
}{
	{regexp.MustCompile(`(?i)(curl|wget)\b[^|\n]{0,200}\|\s*(sudo\s+)?(ba|z|da|k)?sh\b`), "WARN", "pipe-to-shell 安装形态（curl|sh——投毒标准载荷；正文指引供复核）"},
	{regexp.MustCompile(`(?im)^\s*rm\s+-rf\s+/(?:\s|$|\*)`), "WARN", "根目录递归删除形态（rm -rf /；示例/反例文档供复核）"},
	{regexp.MustCompile(`(?i)(ignore|disregard|override)\s+(all\s+|any\s+)?(previous|prior|above)\s+(instructions|rules|prompts?)`), "WARN", "英文提示注入短语（ignore previous instructions 形态）"},
	{regexp.MustCompile(`忽略(之前|上面|以上|先前)(的)?(所有)?(系统|历史)?指令|无视(之前|系统|上面)(的)?指令`), "WARN", "中文提示注入短语（忽略之前指令形态）"},
	{regexp.MustCompile(`(?i)(cat|less|more|head|tail)\s+[^|\n;]*(id_rsa|\.aws/credentials|\.kube/config|/etc/shadow|\.ssh/)`), "WARN", "凭证文件直读建议（应走平台凭证面 requires_credentials）"},
}

// scanDangerous 对 packs 文本做危险内容扫描（walk 与 scanSensitive 同口径：跳过
// ./_ 前缀目录、scanExts 文件类型、跨根去重）。changes/ 目录命中一律升 ERROR。
