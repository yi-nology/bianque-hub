package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRegisterKeyword(t *testing.T) {
	kw := map[string]kwFirst{}

	// 首登无发现
	if level, _, dup := registerKeyword(kw, "慢查询", "db-ops", "P3"); dup || level != "" {
		t.Fatalf("首登不应有发现：level=%q dup=%v", level, dup)
	}
	// 同包重复放行
	if level, _, dup := registerKeyword(kw, "慢查询", "db-ops", "P2"); dup || level != "" {
		t.Fatalf("同包重复应放行：level=%q dup=%v", level, dup)
	}
	// 跨包同优先级 = 真歧义（平台装载会失败）
	level, msg, dup := registerKeyword(kw, "慢查询", "mw-ops", "P3")
	if !dup || level != "ERROR" || !strings.Contains(msg, "真歧义") {
		t.Fatalf("跨包同优先级应 ERROR：level=%q msg=%q", level, msg)
	}
	// 跨层 = 警告（高优先级胜出）
	level, msg, dup = registerKeyword(kw, "慢查询", "obs-ops", "P4")
	if !dup || level != "WARN" || !strings.Contains(msg, "跨层") {
		t.Fatalf("跨层应 WARN：level=%q msg=%q", level, msg)
	}
}

func TestReadFrontmatter(t *testing.T) {
	dir := t.TempDir()
	write := func(name, content string) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		return p
	}
	// LF 正常解析
	f, b, err := readFrontmatter(write("lf.md", "---\nname: x\n---\nbody"))
	if err != nil || !strings.Contains(f, "name: x") || !strings.Contains(b, "body") {
		t.Fatalf("LF 解析失败：front=%q body=%q err=%v", f, b, err)
	}
	// CRLF（Windows 编辑器）不误报缺围栏
	if _, _, err := readFrontmatter(write("crlf.md", "---\r\nname: x\r\n---\r\nbody")); err != nil {
		t.Fatalf("CRLF 应兼容：%v", err)
	}
	// BOM + CRLF
	if _, _, err := readFrontmatter(write("bom.md", "\ufeff---\r\nname: x\r\n---\r\nbody")); err != nil {
		t.Fatalf("BOM 应兼容：%v", err)
	}
	if _, _, err := readFrontmatter(write("nostart.md", "name: x\n---\nbody")); err == nil {
		t.Fatal("缺起始围栏应报错")
	}
	if _, _, err := readFrontmatter(write("noend.md", "---\nname: x\nbody")); err == nil {
		t.Fatal("缺结束围栏应报错")
	}
}

func TestSkillRefName(t *testing.T) {
	if got := skillRefName("redis-triage"); got != "redis-triage" {
		t.Fatalf("裸串形态：%q", got)
	}
	if got := skillRefName(map[string]any{"name": "redis-triage", "allow": []any{}}); got != "redis-triage" {
		t.Fatalf("对象形态：%q", got)
	}
	if got := skillRefName(map[string]any{"server": "ask-ops"}); got != "" {
		t.Fatalf("无 name 的对象应返回空：%q", got)
	}
	if got := skillRefName(42); got != "" {
		t.Fatalf("未知类型应返回空：%q", got)
	}
}

func TestScanSensitive(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		"a.md":           "内网 IP 192.168.1.10 出现",
		"b.txt":          "AKIAABCDEFGHIJKLMNOP",
		"c.go":           "// 192.168.1.10 非 scanned 类型",
		".git/config.md": "10.0.0.1 隐藏目录应跳过",
	}
	for name, content := range files {
		p := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	out := scanSensitive(dir)
	if len(out) != 2 {
		t.Fatalf("应命中 2 处（.md 私网 IP + .txt AKIA），.go 与 .git 不计，实得：%v", out)
	}
	// 同一目录重复传入不双报
	if out := scanSensitive(dir, dir); len(out) != 2 {
		t.Fatalf("重复根不应双报，实得 %d 处", len(out))
	}
}

// TestRunSinglePackAndEmptyContainer 回归两件事：
// 1. 空容器目录必须报错——曾经的假绿（chain-starter 模板自检 0 项通过）；
// 2. 单包目录（参数自身含 pack.yaml）必须真的校验到内容。
func TestRunSinglePackAndEmptyContainer(t *testing.T) {
	empty := t.TempDir()
	findings, errN, _ := run([]string{empty})
	if errN == 0 || !hasMsg(findings, "未发现任何包") {
		t.Fatalf("空容器应 ERROR（防假绿）：errN=%d findings=%v", errN, findings)
	}

	root := t.TempDir()
	pack := filepath.Join(root, "demo")
	writeMinimalPack(t, pack)

	findings, errN, _ = run([]string{pack})
	if errN != 0 {
		t.Fatalf("最小合法包应 0 error：%v", errorMessages(findings))
	}

	// 破坏契约门：单包模式必须拦下（此前该形态静默校验 0 项）
	manifest := readFileT(t, filepath.Join(pack, "pack.yaml"))
	writeFileT(t, filepath.Join(pack, "pack.yaml"), strings.Replace(manifest, "api_version: 1", "api_version: 2", 1))
	findings, errN, _ = run([]string{pack})
	if errN == 0 || !hasMsg(findings, "api_version") {
		t.Fatalf("单包模式应校验 pack.yaml 契约门：errN=%d", errN)
	}
}

// TestRunToolCoverageCrossCheck 回归 obs-ops 0.3.0 loki-triage 事故形态：
// 技能声明了 requires_mcp server，挂载它的专家 tools 未授权 → ERROR。
func TestRunToolCoverageCrossCheck(t *testing.T) {
	root := t.TempDir()
	pack := filepath.Join(root, "bad")
	writeMinimalPack(t, pack)
	skill := filepath.Join(pack, "skills", "demo-skill", "SKILL.md")
	s := readFileT(t, skill)
	s = strings.Replace(s, "maturity: experimental",
		"maturity: experimental\nrequires_mcp:\n  - server: datasources\n    tools: [x_query]", 1)
	writeFileT(t, skill, s)

	findings, errN, _ := run([]string{pack})
	if errN == 0 || !hasMsg(findings, "requires_mcp") {
		t.Fatalf("工具面覆盖缺口应 ERROR：%v", errorMessages(findings))
	}
}

// TestRunVersionSync 校验 CHANGELOG 头部与 pack.yaml 版本同步门。
func TestRunVersionSync(t *testing.T) {
	root := t.TempDir()
	pack := filepath.Join(root, "vers")
	writeMinimalPack(t, pack)
	// 升 pack.yaml 不升 CHANGELOG → ERROR
	manifest := readFileT(t, filepath.Join(pack, "pack.yaml"))
	writeFileT(t, filepath.Join(pack, "pack.yaml"), strings.Replace(manifest, "version: 1.0.0", "version: 1.1.0", 1))
	findings, errN, _ := run([]string{pack})
	if errN == 0 || !hasMsg(findings, "头部版本") {
		t.Fatalf("CHANGELOG 头版本漂移应 ERROR：%v", errorMessages(findings))
	}
}

// ---- 测试脚手架 ----

// writeMinimalPack 落一个全绿的最小合法包（覆盖全部校验层的最小形态）。
func writeMinimalPack(t *testing.T, dir string) {
	t.Helper()
	must := func(rel, content string) {
		p := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	must("pack.yaml", "api_version: 1\nname: demo\nversion: 1.0.0\ndescription: 测试包\nprovides:\n  experts: [imports/demo-expert]\n  skills: [demo-skill]\nchangelog: CHANGELOG.md\n")
	must("CHANGELOG.md", "# CHANGELOG\n\n## 1.0.0 (2026-10-02)\n\n- 首发。\n")
	must("provenance.json", `{"base_url": "https://example.com", "pack": "demo"}`)
	must("demo-expert/agent.yaml", "slug: imports/demo-expert\nname: 测试专家\nlayer: expert\nkind: llm\nprompt_file: prompt.md\nskills: [demo-skill]\ntools:\n  - server: ask-ops\n    allow: []\nroute_priority: P3\nroute_group: tooling\nroute_keywords: [测试词]\nroute_desc: 测试专家路由描述\n")
	must("demo-expert/prompt.md", "你是测试专家。\n")
	must("skills/demo-skill/SKILL.md", "---\nname: demo-skill\ndescription: 测试技能\nmode: on_demand\nversion: 1.0.0\nmaturity: experimental\n---\n\n## 触发条件\n\n- 测试场景\n")
}

func readFileT(t *testing.T, p string) string {
	t.Helper()
	raw, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func writeFileT(t *testing.T, p, content string) {
	t.Helper()
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func hasMsg(findings []finding, sub string) bool {
	for _, f := range findings {
		if strings.Contains(f.msg, sub) {
			return true
		}
	}
	return false
}

func errorMessages(findings []finding) []string {
	var out []string
	for _, f := range findings {
		if f.level == "ERROR" {
			out = append(out, f.msg)
		}
	}
	return out
}

// TestRunChangeBlocks 编目变更块校验层（批次九十四「受审执行」）：
// 合法块通过；模板带命令替换/最小契约缺失 ERROR；provides_changes 引用闭包
// （引用不存在 slug ERROR、引用存在块 0 error）。
func TestRunChangeBlocks(t *testing.T) {
	root := t.TempDir()
	pack := filepath.Join(root, "demo")
	writeMinimalPack(t, pack)
	must := func(rel, content string) {
		p := filepath.Join(pack, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	must("changes/good.yaml", "api_version: 1\nslug: demo-restart\ntitle: 重启服务\nrequest_type: restart\nrisk: 2\nrollback: 再启一次\ncommands:\n  - \"systemctl restart demo\"\n")

	// 合法块 + 技能引用存在 slug → 0 error
	skill := filepath.Join(pack, "skills", "demo-skill", "SKILL.md")
	s := readFileT(t, skill)
	writeFileT(t, skill, strings.Replace(s, "maturity: experimental",
		"maturity: experimental\nprovides_changes: [demo-restart]", 1))
	findings, errN, _ := run([]string{pack})
	if errN != 0 {
		t.Fatalf("合法变更块+引用应 0 error：%v", errorMessages(findings))
	}

	// 模板带命令替换 → ERROR
	must("changes/tainted.yaml", "api_version: 1\nslug: demo-taint\ntitle: 坏模板\nrequest_type: restart\nrisk: 2\nrollback: x\ncommands:\n  - \"echo `id`\"\n")
	findings, errN, _ = run([]string{pack})
	if errN == 0 || !hasMsg(findings, "命令模板非法") {
		t.Fatalf("命令替换模板应 ERROR：%v", errorMessages(findings))
	}
	_ = os.Remove(filepath.Join(pack, "changes", "tainted.yaml"))

	// 最小契约缺失（缺 request_type）→ ERROR
	must("changes/minimal.yaml", "api_version: 1\nslug: demo-min\ntitle: 缺字段\nrisk: 2\ncommands: [a]\n")
	findings, errN, _ = run([]string{pack})
	if errN == 0 || !hasMsg(findings, "最小契约") {
		t.Fatalf("缺 request_type 应 ERROR：%v", errorMessages(findings))
	}
	_ = os.Remove(filepath.Join(pack, "changes", "minimal.yaml"))

	// 引用不存在 slug → ERROR（编目引用闭包）
	writeFileT(t, skill, strings.Replace(s, "maturity: experimental",
		"maturity: experimental\nprovides_changes: [no-such-change]", 1))
	findings, errN, _ = run([]string{pack})
	if errN == 0 || !hasMsg(findings, "provides_changes 引用") {
		t.Fatalf("引用不存在变更块应 ERROR：%v", errorMessages(findings))
	}
}

// 专家可独立安装性（spec 2026-10-05-per-expert-install §6）：跨包技能引用与
// 无契约非宿主面 tools server 都是 ERROR；宿主面白名单放行。
func TestCheckExpertInstall(t *testing.T) {
	root := t.TempDir()
	write := func(rel, content string) {
		p := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("demo/pack.yaml", "api_version: 1\nname: demo\nversion: 1.0.0\ndescription: 测试\nprovides:\n  experts:\n    - imports/e1\n")
	write("demo/provenance.json", `{"base_url":"https://github.com/yi-nology/bianque-hub"}`)
	write("demo/CHANGELOG.md", "## 1.0.0\n- init\n")
	write("demo/e1/agent.yaml", "slug: imports/e1\nname: E1\nkind: llm\nprompt_file: prompt.md\nroute_priority: P2\nroute_keywords: [k1]\nskills:\n  - mine\n  - foreign\ntools:\n  - server: ghost-mcp\n    allow: []\n  - server: ask-ops\n    allow: []\n")
	write("demo/e1/prompt.md", "p")
	write("demo/skills/mine/SKILL.md", "---\nname: mine\ndescription: d\nversion: 0.1.0\n---\nbody\n")
	write("other/pack.yaml", "api_version: 1\nname: other\nversion: 1.0.0\ndescription: 测试\nprovides:\n  skills: [foreign]\n")
	write("other/provenance.json", `{"base_url":"https://github.com/yi-nology/bianque-hub"}`)
	write("other/CHANGELOG.md", "## 1.0.0\n- init\n")
	write("other/skills/foreign/SKILL.md", "---\nname: foreign\ndescription: d\nversion: 0.1.0\n---\nbody\n")

	findings, _, _ := run([]string{root})
	var errs, warns []string
	for _, f := range findings {
		if f.level == "ERROR" {
			errs = append(errs, f.msg)
		} else {
			warns = append(warns, f.msg)
		}
	}
	joined := strings.Join(errs, "\n")
	if !strings.Contains(joined, "跨包技能") || !strings.Contains(joined, "foreign") {
		t.Fatalf("跨包技能引用应 ERROR: %s", joined)
	}
	if strings.Contains(joined, "ghost-mcp") {
		t.Fatalf("契约缺席是 WARN 不是 ERROR: %s", joined)
	}
	wjoined := strings.Join(warns, "\n")
	if !strings.Contains(wjoined, "ghost-mcp") || !strings.Contains(wjoined, "宿主面") {
		t.Fatalf("无契约非宿主面 server 应 WARN: %s", wjoined)
	}
	if strings.Contains(wjoined+"\n"+joined, "ask-ops") {
		t.Fatalf("宿主面 server 不应报错: %s", wjoined)
	}
}

// TestRunIndexConsistency index.json 对账门（批次一百三十九）：缺=WARN、漂移/多录/
// 漏包=ERROR。索引放 packs 容器的父目录（=真实 hub 仓根位置）。
func TestRunIndexConsistency(t *testing.T) {
	hub := t.TempDir()
	packs := filepath.Join(hub, "packs")
	writeMinimalPack(t, filepath.Join(packs, "demo"))
	indexPath := filepath.Join(hub, "index.json")
	writeIndex := func(entries string) {
		writeFileT(t, indexPath, `{"generated_at":"2026-10-05","packs":[`+entries+`]}`)
	}

	// 缺索引：WARN 不 ERROR（渐进采用；检索侧自动降级现场扫描）。
	// 断言只看 index 维度 findings——run 末尾的 scanSensitive(".") 依赖测试进程
	// cwd，环境告警不归本测试管。
	findings, errN, _ := run([]string{packs})
	if errN != 0 || !hasMsg(findings, "index.json 缺失") {
		t.Fatalf("缺索引应 WARN 提示且无 ERROR：%v", errorMessages(findings))
	}

	// 一致索引：index 维度全静默
	writeIndex(`{"name":"demo","version":"1.0.0"}`)
	findings, errN, _ = run([]string{packs})
	if errN != 0 || hasMsg(findings, "index") {
		t.Fatalf("一致索引应全静默：errN=%d %v", errN, errorMessages(findings))
	}

	// 版本漂移：ERROR
	writeIndex(`{"name":"demo","version":"9.9.9"}`)
	findings, errN, _ = run([]string{packs})
	if errN == 0 || !hasMsg(findings, "版本漂移") {
		t.Fatalf("版本漂移应 ERROR：%v", errorMessages(findings))
	}

	// 多录 + 漏包：双 ERROR
	writeIndex(`{"name":"ghost","version":"1.0.0"}`)
	findings, errN, _ = run([]string{packs})
	if errN < 2 || !hasMsg(findings, "多录") || !hasMsg(findings, "漏包") {
		t.Fatalf("多录与漏包应双 ERROR：%v", errorMessages(findings))
	}
}
