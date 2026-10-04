# 专家级子集安装 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** `bq-markettool install --expert <专家>` 按闭包现场合成迷你包走既有插件管线，让每个专家可独立安装；hub 侧校验器把住"专家可独立安装性"契约。

**Architecture:** 包不拆、平台装载模型不动。安装器在解析出整包目录后，对每个 `--expert` 值派生迷你包（专家目录+其技能+技能 provides_changes 引用的变更块+tools server 的包内 mcp 契约+analyzers 单条目+credentials，合成 pack.yaml/provenance.json），再复用 `apiInstall`/`destInstall` 既有原子路径；插件名 `<包>.<专家目录名>`。hub validator 新增跨包技能引用与工具契约存在性两道 ERROR 门。

**Tech Stack:** Go 1.24 + gopkg.in/yaml.v3（两仓既有依赖，零新增）。

**Spec:** `docs/superpowers/specs/2026-10-05-per-expert-install-design.md`（本仓；两仓共读）

## Global Constraints

- 两仓路径：HUB=`/Users/zhangyi/my_project/bianque-hub`，PLAT=`/Users/zhangyi/my_project/bianque`。**两个独立 git 仓，分别提交。**
- `packAPIVersion = 1`（PLAT `install.go` 常量；不得改动）。
- 插件名法度 `^[a-z0-9][a-z0-9._-]{0,127}$`（PLAT `plugin_svc.go:38`）；专家目录名已全小写，`<包>.<专家>` 天然合法。
- 依赖白名单：PLAT `cmd/bq-markettool` 只许 `gopkg.in/yaml.v3`（现有 import 面：标准库 + `internal/service` + yaml）。派生逻辑放 cmd 内私有函数，**不引 `internal/agents`**（cmd 不引重包，既有纪律）。
- 注释风格：中文、说明"为什么"与平台同构关系（照 PLAT 仓现有注释密度）。
- 提交纪律：开工先 `git status`——两仓都可能有用户并发 WIP，**只 `git add` 本计划明确列出的文件**；提交前 `git pull --rebase`。
- 禁改：PLAT `experts/`、`data-e2e-*/` 目录；HUB 各包内容与版本号（本设计不改包内容，包版本不升）。
- 宿主面 server 白名单（两处口径同步）：`ask-ops` / `security-assistant` / `datasources`——包内无契约文件、装载即存在。

---

### Task 1: PLAT——readPackMeta 版本扩展 + deriveExpertPack 闭包合成

**Files:**
- Modify: `PLAT/cmd/bq-markettool/install.go`（readPackMeta 签名 3→4 返回值；cmdInstall 调用点一处）
- Create: `PLAT/cmd/bq-markettool/derive.go`
- Test: `PLAT/cmd/bq-markettool/derive_test.go`、`PLAT/cmd/bq-markettool/install_test.go`（改 TestReadPackMetaDefaults）

**Interfaces:**
- Consumes: `copyTree(src, dst string) error`（install.go:330 既有）、`readPackMeta`。
- Produces: `deriveExpertPack(packDir, expert, srcURL string) (dir, name string, cleanup func(), err error)`——Task 2 消费；`readPackMeta(packDir string) (name, version string, apiVersion int, err error)` 新签名。

- [ ] **Step 1: 改 readPackMeta 签名并更新调用点与既有测试（先红）**

`install.go` 的 `readPackMeta` 改为：

```go
func readPackMeta(packDir string) (string, string, int, error) {
	raw, err := os.ReadFile(filepath.Join(packDir, "pack.yaml"))
	if err != nil {
		return "", "", 0, err
	}
	var m struct {
		Name       string `yaml:"name"`
		Version    string `yaml:"version"`
		APIVersion int    `yaml:"api_version"`
	}
	if err := yaml.Unmarshal(raw, &m); err != nil {
		return "", "", 0, fmt.Errorf("pack.yaml 解析: %w", err)
	}
	if m.Name == "" {
		m.Name = filepath.Base(packDir)
	}
	return m.Name, m.Version, m.APIVersion, nil
}
```

`cmdInstall` 中 `name, ver, err := readPackMeta(packDir)` 改为 `name, packVer, ver, err := readPackMeta(packDir)`（`packVer` 本任务暂 `_ = packVer`，Task 2 传入派生）。`install_test.go` 的 `TestReadPackMetaDefaults` 两处调用同步改四返回值：`name, packVer, ver, err := readPackMeta(packDir)`，断言补 `packVer != "0.1.0"`（`writeMiniPack` 的 pack.yaml 固定写 `version: 0.1.0`——install_test.go:18）；第二处 `if name, _, _, err := readPackMeta(p2); ...` 不变。

Run: `cd PLAT && go test ./cmd/bq-markettool/ -run 'TestReadPackMeta|TestFindPackDir' -v`
Expected: PASS（编译过、既有断言不破）

- [ ] **Step 2: 写 derive_test.go——fixture + 闭包断言（红）**

新建 `derive_test.go`：

```go
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// writeExpertPack 整包 fixture：两专家（e1 带技能/变更/mcp 授权，e2 纯 stub）+
// 链 + 消歧 + 无关变更块 + analyzers 两条目——闭包测试基座。
func writeExpertPack(t *testing.T, root, name string) string {
	t.Helper()
	p := filepath.Join(root, name)
	mk := func(rel, content string) {
		path := filepath.Join(p, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	mk("pack.yaml", fmt.Sprintf("api_version: 1\nname: %s\nversion: 1.2.3\ndescription: 测试包\n", name))
	mk("provenance.json", `{"base_url":"https://github.com/yi-nology/bianque-hub"}`)
	mk("chain.yaml", "slug: workflow/test-chain\nname: 测试链\nroute_priority: P2\nsteps:\n  - agent: imports/e1\n")
	mk("disambiguation.yaml", "disambiguation: []\n")
	mk("changes/other.yaml", "api_version: 1\nslug: other-change\ntitle: 不相关\nrequest_type: rollback\nrisk: 2\ncommands: [\"ls\"]\n")
	mk("changes/e1-rollback.yaml", "api_version: 1\nslug: e1-rollback\ntitle: E1 回滚\nrequest_type: rollback\nrisk: 2\ncommands: [\"ls\"]\n")
	mk("mcp/testmcp.yaml", "name: testmcp\nversion: 1.0.0\ntools:\n  - name: ping\n")
	mk("skills/e1-skill/SKILL.md", "---\nname: e1-skill\ndescription: E1 技能\nversion: 0.1.0\nprovides_changes:\n  - e1-rollback\n---\n正文")
	mk("skills/shared-skill/SKILL.md", "---\nname: shared-skill\ndescription: 无人挂\nversion: 0.1.0\n---\n正文")
	mk("e1/agent.yaml", "slug: imports/e1\nname: E1 专家\nkind: llm\ntemperature: 0.2\nprompt_file: prompt.md\nskills:\n  - e1-skill\ntools:\n  - server: testmcp\n    allow: [ping]\n  - server: ask-ops\n    allow: []\nroute_priority: P2\n")
	mk("e1/prompt.md", "E1 prompt")
	mk("e2/agent.yaml", "slug: imports/e2\nname: E2 专家\nkind: stub\nroute_priority: P3\n")
	mk("analyzers.yaml", "analyzers:\n  - slug: imports/e1\n    groups: [t1]\n  - slug: imports/e2\n    groups: [t2]\n")
	return p
}

func TestDeriveExpertPackClosure(t *testing.T) {
	root := t.TempDir()
	packDir := writeExpertPack(t, root, "demo")
	dir, name, cleanup, err := deriveExpertPack(packDir, "e1", "https://github.com/yi-nology/bianque-hub")
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	if name != "demo.e1" {
		t.Fatalf("派生名应 demo.e1，得到 %s", name)
	}
	// 专家目录全量
	for _, rel := range []string{"e1/agent.yaml", "e1/prompt.md"} {
		if _, err := os.Stat(filepath.Join(dir, rel)); err != nil {
			t.Fatalf("缺 %s: %v", rel, err)
		}
	}
	// 技能闭包：只带挂载技能
	if _, err := os.Stat(filepath.Join(dir, "skills/e1-skill/SKILL.md")); err != nil {
		t.Fatal("应携带 e1-skill")
	}
	if _, err := os.Stat(filepath.Join(dir, "skills/shared-skill/SKILL.md")); err == nil {
		t.Fatal("未挂载技能不得携带")
	}
	// 变更块闭包：只带 provides_changes 引用
	if _, err := os.Stat(filepath.Join(dir, "changes/e1-rollback.yaml")); err != nil {
		t.Fatal("应携带 e1-rollback")
	}
	if _, err := os.Stat(filepath.Join(dir, "changes/other.yaml")); err == nil {
		t.Fatal("未引用变更块不得携带")
	}
	// mcp 契约只带包内有的；宿主侧 ask-ops 不应有文件
	if _, err := os.Stat(filepath.Join(dir, "mcp/testmcp.yaml")); err != nil {
		t.Fatal("应携带 testmcp 契约")
	}
	// 包级资产不带
	for _, rel := range []string{"chain.yaml", "disambiguation.yaml", "README.md", "CHANGELOG.md"} {
		if _, err := os.Stat(filepath.Join(dir, rel)); err == nil {
			t.Fatalf("%s 不得携带", rel)
		}
	}
	// analyzers 单条目
	araw, _ := os.ReadFile(filepath.Join(dir, "analyzers.yaml"))
	var an struct {
		Analyzers []struct {
			Slug string `yaml:"slug"`
		} `yaml:"analyzers"`
	}
	if err := yaml.Unmarshal(araw, &an); err != nil || len(an.Analyzers) != 1 || an.Analyzers[0].Slug != "imports/e1" {
		t.Fatalf("analyzers 应恰含 imports/e1 一条: %s %v %+v", araw, err, an.Analyzers)
	}
	// 合成 pack.yaml
	var m struct {
		Name       string `yaml:"name"`
		Version    string `yaml:"version"`
		APIVersion int    `yaml:"api_version"`
		Provides   struct {
			Experts []string `yaml:"experts"`
			Skills  []string `yaml:"skills"`
		} `yaml:"provides"`
	}
	mraw, _ := os.ReadFile(filepath.Join(dir, "pack.yaml"))
	if err := yaml.Unmarshal(mraw, &m); err != nil {
		t.Fatal(err)
	}
	if m.Name != "demo.e1" || m.Version != "1.2.3" || m.APIVersion != 1 ||
		len(m.Provides.Experts) != 1 || m.Provides.Experts[0] != "imports/e1" ||
		len(m.Provides.Skills) != 1 || m.Provides.Skills[0] != "e1-skill" {
		t.Fatalf("合成 pack.yaml 不符: %+v", m)
	}
	// provenance 血缘
	var prov struct {
		DerivedFrom struct {
			Pack    string `json:"pack"`
			Version string `json:"version"`
			Expert  string `json:"expert"`
		} `json:"derived_from"`
		BaseURL string `json:"base_url"`
	}
 praw, _ := os.ReadFile(filepath.Join(dir, "provenance.json"))
	if err := json.Unmarshal(praw, &prov); err != nil {
		t.Fatal(err)
	}
	if prov.DerivedFrom.Pack != "demo" || prov.DerivedFrom.Version != "1.2.3" ||
		prov.DerivedFrom.Expert != "e1" || prov.BaseURL != "https://github.com/yi-nology/bianque-hub" {
		t.Fatalf("provenance 不符: %+v", prov)
	}
}

func TestDeriveExpertPackStubNoOptionalAssets(t *testing.T) {
	root := t.TempDir()
	packDir := writeExpertPack(t, root, "demo")
	dir, name, cleanup, err := deriveExpertPack(packDir, "e2", "")
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	if name != "demo.e2" {
		t.Fatalf("应 demo.e2: %s", name)
	}
	// e2 无技能/无 mcp 授权/analyzers 有条目：skills、changes、mcp 不得有目录壳
	for _, rel := range []string{"skills", "changes", "mcp"} {
		if _, err := os.Stat(filepath.Join(dir, rel)); err == nil {
			t.Fatalf("e2 不应产生 %s/", rel)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "analyzers.yaml")); err != nil {
		t.Fatal("e2 的 analyzers 条目应携带")
	}
}

func TestDeriveExpertPackErrors(t *testing.T) {
	root := t.TempDir()
	packDir := writeExpertPack(t, root, "demo")
	// 未知专家
	if _, _, cleanup, err := deriveExpertPack(packDir, "ghost", ""); err == nil {
		t.Fatal("未知专家应报错")
	} else {
		cleanup()
	}
	// 跨包技能引用不可携带
	p := filepath.Join(packDir, "e3")
	os.MkdirAll(p, 0o755)
	os.WriteFile(filepath.Join(p, "agent.yaml"), []byte("slug: imports/e3\nname: E3\nkind: stub\nroute_priority: P3\nskills:\n  - not-in-pack\n"), 0o644)
	_, _, cleanup, err := deriveExpertPack(packDir, "e3", "")
	if err == nil || !strings.Contains(err.Error(), "not-in-pack") {
		t.Fatalf("跨包技能引用应报错: %v", err)
	}
	cleanup()
	// provides_changes 引用缺位不可携带
	p4 := filepath.Join(packDir, "e4")
	os.MkdirAll(p4, 0o755)
	os.WriteFile(filepath.Join(p4, "agent.yaml"), []byte("slug: imports/e4\nname: E4\nkind: stub\nroute_priority: P3\n"), 0o644)
	os.WriteFile(filepath.Join(packDir, "skills", "lonely-skill", "SKILL.md"), []byte("---\nname: lonely-skill\ndescription: d\nversion: 0.1.0\nprovides_changes:\n  - ghost-change\n---\n正文"), 0o644)
	os.WriteFile(filepath.Join(p4, "agent.yaml"), []byte("slug: imports/e4\nname: E4\nkind: stub\nroute_priority: P3\nskills:\n  - lonely-skill\n"), 0o644)
	_, _, cleanup, err = deriveExpertPack(packDir, "e4", "")
	if err == nil || !strings.Contains(err.Error(), "ghost-change") {
		t.Fatalf("缺位变更块引用应报错: %v", err)
	}
	cleanup()
}
```

（`praw` 行对齐 gofmt；写完跑 `gofmt -w derive_test.go`。）

Run: `cd PLAT && go test ./cmd/bq-markettool/ -run TestDeriveExpertPack -v`
Expected: FAIL——`deriveExpertPack` 未定义（编译错）

- [ ] **Step 3: 实现 derive.go**

```go
package main

// derive.go 专家级子集安装（bianque-hub docs/superpowers/specs/
// 2026-10-05-per-expert-install-design.md §2）：从整包按闭包现场合成迷你包——
// 专家目录 + 其技能 + 技能 provides_changes 引用的变更块 + tools server 的包内
// mcp 契约 + analyzers 单条目 + credentials。刻意不带 chain.yaml /
// disambiguation.yaml / README / CHANGELOG（包级跨专家资产，单专家语义下无意义
// 且链引用缺位专家会断 reload）。
//
// 产物走既有 apiInstall/destInstall 原子路径，插件名 <包>.<专家目录名>；
// provenance.json 记 derived_from 血缘（platform 靠该文件判 source=market 可卸载）。

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// skillRefName 技能引用取名（agent.yaml skills: 裸串 / {name, version} 双形态；
// 镜像 hub validator.skillRefName 与平台 agents deriveSkills 口径）。
func skillRefName(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case map[string]any:
		if n, ok := x["name"].(string); ok {
			return n
		}
	}
	return ""
}

// skillFrontmatter 取 SKILL.md 的 YAML frontmatter（同构复制 hub validator
// readFrontmatter：BOM/CRLF 兼容）。
func skillFrontmatter(p string) (string, error) {
	raw, err := os.ReadFile(p)
	if err != nil {
		return "", err
	}
	s := strings.ReplaceAll(string(raw), "\r\n", "\n")
	s = strings.TrimPrefix(s, "﻿")
	if !strings.HasPrefix(s, "---\n") {
		return "", fmt.Errorf("%s 缺 frontmatter 起始围栏", p)
	}
	end := strings.Index(s[4:], "\n---")
	if end < 0 {
		return "", fmt.Errorf("%s 缺 frontmatter 结束围栏", p)
	}
	return s[4 : 4+end], nil
}

type analyzerEntry struct {
	Slug   string   `yaml:"slug"`
	Groups []string `yaml:"groups"`
	Source string   `yaml:"source"`
}

type analyzerFile struct {
	Analyzers []analyzerEntry `yaml:"analyzers"`
}

// deriveExpertPack 闭包合成迷你包。dir 为临时目录（调用方必须调 cleanup）；
// name 为插件名（<包>.<专家目录名>）。
func deriveExpertPack(packDir, expert, srcURL string) (string, string, func(), error) {
	srcExpert := filepath.Join(packDir, expert)
	if _, err := os.Stat(filepath.Join(srcExpert, "agent.yaml")); err != nil {
		return "", "", nil, fmt.Errorf("包 %s 内无专家 %q（缺 agent.yaml）", filepath.Base(packDir), expert)
	}
	packName, packVersion, _, err := readPackMeta(packDir)
	if err != nil {
		return "", "", nil, err
	}
	name = packName + "." + expert

	tmp, err := os.MkdirTemp("", "bq-expert-"+expert+"-")
	if err != nil {
		return "", "", nil, err
	}
	boom := func(err error) (string, string, func(), error) {
		os.RemoveAll(tmp)
		return "", "", nil, err
	}

	// ① 专家目录全量
	if err := copyTree(srcExpert, filepath.Join(tmp, expert)); err != nil {
		return boom(err)
	}
	var agent struct {
		Slug   string   `yaml:"slug"`
		Skills []any    `yaml:"skills"`
		Tools  []struct {
			Server string `yaml:"server"`
		} `yaml:"tools"`
	}
	araw, err := os.ReadFile(filepath.Join(srcExpert, "agent.yaml"))
	if err != nil {
		return boom(err)
	}
	if err := yaml.Unmarshal(araw, &agent); err != nil {
		return boom(fmt.Errorf("%s/agent.yaml 解析: %w", srcExpert, err))
	}

	// ② 技能闭包：必须在本包（跨包引用无法携带——hub validator 硬规则 11 拦，
	// 这里是安装期最后防线）
	var skills []string
	for _, raw := range agent.Skills {
		s := skillRefName(raw)
		if s == "" {
			continue
		}
		sdir := filepath.Join(packDir, "skills", s)
		if _, err := os.Stat(filepath.Join(sdir, "SKILL.md")); err != nil {
			return boom(fmt.Errorf("专家 %s 引用技能 %q 不在包 %s 内——专家级安装无法携带（跨包技能引用请整包安装）", expert, s, packName))
		}
		if err := copyTree(sdir, filepath.Join(tmp, "skills", s)); err != nil {
			return boom(err)
		}
		skills = append(skills, s)
	}

	// ③ 变更块闭包：所携技能 frontmatter provides_changes（跨包编目同样不可携带）
	if len(skills) > 0 {
		seen := map[string]bool{}
		for _, s := range skills {
			front, err := skillFrontmatter(filepath.Join(tmp, "skills", s, "SKILL.md"))
			if err != nil {
				return boom(err)
			}
			var sf struct {
				ProvidesChanges []string `yaml:"provides_changes"`
			}
			if err := yaml.Unmarshal([]byte(front), &sf); err != nil {
				return boom(fmt.Errorf("技能 %s frontmatter 解析: %w", s, err))
			}
			for _, slug := range sf.ProvidesChanges {
				if seen[slug] {
					continue
				}
				seen[slug] = true
				src := filepath.Join(packDir, "changes", slug+".yaml")
				if _, err := os.Stat(src); err != nil {
					return boom(fmt.Errorf("技能 %s provides_changes 引用 %q 不在包 %s changes/ 内——专家级安装无法携带（跨包编目请整包安装）", s, slug, packName))
				}
				if err := copyTree(src, filepath.Join(tmp, "changes", slug+".yaml")); err != nil {
					return boom(err)
				}
			}
		}
	}

	// ④ mcp 契约：只带包内有契约文件的 server；宿主侧（ask-ops 等）无文件，跳过
	for _, t := range agent.Tools {
		if t.Server == "" {
			continue
		}
		src := filepath.Join(packDir, "mcp", t.Server+".yaml")
		if _, err := os.Stat(src); err == nil {
			if err := copyTree(src, filepath.Join(tmp, "mcp", t.Server+".yaml")); err != nil {
				return boom(err)
			}
		}
	}

	// ⑤ analyzers 单条目（装载期跨包归并、组键跨包唯一——只带本专家条目即本包视角全覆盖）
	if raw, err := os.ReadFile(filepath.Join(packDir, "analyzers.yaml")); err == nil {
		var af analyzerFile
		if err := yaml.Unmarshal(raw, &af); err != nil {
			return boom(fmt.Errorf("analyzers.yaml 解析: %w", err))
		}
		var mine []analyzerEntry
		for _, e := range af.Analyzers {
			if e.Slug == agent.Slug {
				mine = append(mine, e)
			}
		}
		if len(mine) > 0 {
			out, err := yaml.Marshal(analyzerFile{Analyzers: mine})
			if err != nil {
				return boom(err)
			}
			if err := os.WriteFile(filepath.Join(tmp, "analyzers.yaml"), out, 0o644); err != nil {
				return boom(err)
			}
		}
	}

	// ⑥ credentials 整目录（凭证类型定义面；当前无包使用，前向兼容）
	if _, err := os.Stat(filepath.Join(packDir, "credentials")); err == nil {
		if err := copyTree(filepath.Join(packDir, "credentials"), filepath.Join(tmp, "credentials")); err != nil {
			return boom(err)
		}
	}

	// ⑦ 合成 pack.yaml：name/version/api_version 承源包，provides 按实际携带
	mf := struct {
		APIVersion  int    `yaml:"api_version"`
		Name        string `yaml:"name"`
		Version     string `yaml:"version"`
		Description string `yaml:"description"`
		Provides    struct {
			Experts []string `yaml:"experts"`
			Skills  []string `yaml:"skills"`
		} `yaml:"provides"`
	}{APIVersion: packAPIVersion, Name: name, Version: packVersion,
		Description: fmt.Sprintf("自 %s v%s 子集安装（仅专家 %s）", packName, packVersion, agent.Slug)}
	mf.Provides.Experts = []string{agent.Slug}
	mf.Provides.Skills = skills
	out, err := yaml.Marshal(mf)
	if err != nil {
		return boom(err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "pack.yaml"), out, 0o644); err != nil {
		return boom(err)
	}

	// ⑧ provenance.json：base_url 承源包，derived_from 记血缘
	baseURL := srcURL
	var srcProv struct {
		BaseURL string `json:"base_url"`
	}
	if raw, err := os.ReadFile(filepath.Join(packDir, "provenance.json")); err == nil &&
		json.Unmarshal(raw, &srcProv) == nil && srcProv.BaseURL != "" {
		baseURL = srcProv.BaseURL
	}
	prov := map[string]any{
		"base_url": baseURL,
		"pack":     name,
		"edition":  "community",
		"derived_from": map[string]string{
			"pack": packName, "version": packVersion, "expert": expert,
		},
	}
	pout, err := json.MarshalIndent(prov, "", "  ")
	if err != nil {
		return boom(err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "provenance.json"), append(pout, '\n'), 0o644); err != nil {
		return boom(err)
	}
	return tmp, name, func() { os.RemoveAll(tmp) }, nil
}
```

注意：`skillFrontmatter` 用到 `strings`——import 面补 `"strings"`（上面骨架已含 `encoding/json`/`os`/`path/filepath`/yaml，按 gofmt 提示补齐）。`skills` 为 nil 时 `mf.Provides.Skills = skills` 序列化出 `skills: []`——若希望无技能时不输出该键，改为 `if len(skills) > 0 { mf.Provides.Skills = skills }`（采用后者）。

Run: `cd PLAT && gofmt -w cmd/bq-markettool/ && go test ./cmd/bq-markettool/ -v`
Expected: 全 PASS（含 Task 1 Step 1 的既有测试）

- [ ] **Step 4: 提交（PLAT 仓）**

```bash
cd PLAT && git status   # 确认无并发 WIP 混入
git add cmd/bq-markettool/derive.go cmd/bq-markettool/derive_test.go cmd/bq-markettool/install.go cmd/bq-markettool/install_test.go
git pull --rebase && git commit -m "feat(markettool): 专家级子集安装第一层——deriveExpertPack 闭包合成迷你包（技能/变更块/mcp 契约/analyzers 单条目；包级资产不带；readPackMeta 增返 version）"
```

---

### Task 2: PLAT——cmdInstall --expert 接线（预检 + api/dest 双模式）

**Files:**
- Modify: `PLAT/cmd/bq-markettool/install.go`（cmdInstall：flag、预检、派生循环；文件头注释补一行专家级模式说明）
- Test: `PLAT/cmd/bq-markettool/install_test.go`（追加四个测试）

**Interfaces:**
- Consumes: Task 1 的 `deriveExpertPack`；既有 `apiProbePack(client, base, token, name)`、`apiInstall(base, token, user, pass, packDir, name)`、`destInstall(dest, packDir, name, srcURL)`、`apiLogin`。
- Produces: CLI 口径——`--expert`（可重复/逗号分隔，须与 `--pack` 同用）；不指定时行为与现在逐字节一致。

- [ ] **Step 1: 追加测试（红）**

`install_test.go` 末尾追加：

```go
// 专家级安装（spec 2026-10-05-per-expert-install §4/§5）：--expert 须与 --pack 同用；
// 整包在位预检拒绝；dest 模式端到端落盘为 <pack>.<expert> 独立插件。
func TestCmdInstallExpertRequiresPack(t *testing.T) {
	root := t.TempDir()
	writeExpertPack(t, filepath.Join(root, "packs"), "demo")
	err := cmdInstall([]string{"--url", root, "--expert", "e1", "--dest", t.TempDir()})
	if err == nil || !strings.Contains(err.Error(), "--pack") {
		t.Fatalf("--expert 缺 --pack 应报错: %v", err)
	}
}

func TestCmdInstallExpertDest(t *testing.T) {
	root := t.TempDir()
	writeExpertPack(t, filepath.Join(root, "packs"), "demo")
	dest := t.TempDir()
	if err := cmdInstall([]string{"--url", root, "--pack", "demo", "--expert", "e1", "--dest", dest}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dest, "demo.e1", "pack.yaml")); err != nil {
		t.Fatal("应落盘 demo.e1 迷你包")
	}
	if _, err := os.Stat(filepath.Join(dest, "demo")); err == nil {
		t.Fatal("整包不得落盘")
	}
	// 升级路径：重跑同命令应成功（同名插件覆盖语义在 dest 模式为拒绝——destInstall
	// 既有口径，重复安装报错即符合预期）
	if err := cmdInstall([]string{"--url", root, "--pack", "demo", "--expert", "e1", "--dest", dest}); err == nil {
		t.Fatal("dest 重复落盘应拒绝（既有口径）")
	}
}

func TestCmdInstallExpertPrecheckDest(t *testing.T) {
	root := t.TempDir()
	writeExpertPack(t, filepath.Join(root, "packs"), "demo")
	dest := t.TempDir()
	os.MkdirAll(filepath.Join(dest, "demo"), 0o755)
	err := cmdInstall([]string{"--url", root, "--pack", "demo", "--expert", "e1", "--dest", dest})
	if err == nil || !strings.Contains(err.Error(), "互斥") {
		t.Fatalf("整包在位应预检拒绝: %v", err)
	}
}

func TestCmdInstallExpertPrecheckAPI(t *testing.T) {
	root := t.TempDir()
	writeExpertPack(t, filepath.Join(root, "packs"), "demo")
	var precheckHit bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/plugins/admin") {
			precheckHit = true
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"code":0,"data":{"plugins":[{"name":"demo","source":"market"}]}}`)
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()
	err := cmdInstall([]string{"--url", root, "--pack", "demo", "--expert", "e1", "--api", srv.URL})
	if err == nil || !strings.Contains(err.Error(), "互斥") {
		t.Fatalf("整包在位（API 登记面）应预检拒绝: %v", err)
	}
	if !precheckHit {
		t.Fatal("预检应探测登记面")
	}
}
```

import 面补 `"net/http"`、`"net/http/httptest"`（`fmt`/`strings`/`os`/`path/filepath` 若缺同补）。

Run: `cd PLAT && go test ./cmd/bq-markettool/ -run 'TestCmdInstallExpert' -v`
Expected: FAIL——cmdInstall 尚不认 --expert（flag parse 走 ExitOnError 会 os.Exit——**注意**：`--expert` 是未知 flag，ExitOnError 直接退出进程而非返回 error，测试会以 exit 崩溃呈现，同属"红"）

- [ ] **Step 2: 实现 cmdInstall 接线**

`install.go` 顶部（`packAPIVersion` 常量后）加：

```go
// expertFlags 可重复 --expert（亦容忍逗号分隔）。
type expertFlags []string

func (f *expertFlags) String() string { return strings.Join(*f, ",") }
func (f *expertFlags) Set(v string) error {
	for _, e := range strings.Split(v, ",") {
		if e = strings.TrimSpace(e); e != "" {
			*f = append(*f, e)
		}
	}
	return nil
}
```

`cmdInstall` 的 flag 面在 `dest :=` 之后加：

```go
	var experts expertFlags
	fs.Var(&experts, "expert", "专家目录名（可重复/逗号分隔；须与 --pack 同用）——按闭包合成迷你包逐个独立安装")
```

`cmdInstall` 中 `readPackMeta` 之后、`fmt.Printf("包就绪…")` 之前插入预检；`fmt.Printf` 之后替换安装分支：

```go
	if len(experts) > 0 {
		// 整包与迷你包互斥（同 slug 撞车，reloadStrict 必败）——预检登记面/落盘目录，
		// 拒绝并给出替代路径（专家级启停或先卸整包），不把原子回滚当正常流程用。
		if *api != "" {
			base := strings.TrimRight(*api, "/")
			if !strings.Contains(base, "://") {
				base = "http://" + base
			}
			tok := *token
			if tok == "" && *user != "" {
				t, err := apiLogin(base, *user, *pass)
				if err != nil {
					return err
				}
				tok = t
			}
			if _, exists, err := apiProbePack(&http.Client{Timeout: 120 * time.Second}, base, tok, *pack); err != nil {
				return err
			} else if exists {
				return fmt.Errorf("整包 %s 已在位：与专家级安装互斥（slug 撞车）。只需某专家请先卸载整包再装，或对已装整包用专家级启停（插件页 toggle）", *pack)
			}
		}
		if *dest != "" {
			if _, err := os.Stat(filepath.Join(*dest, *pack)); err == nil {
				return fmt.Errorf("整包 %s 已在 %s：与专家级安装互斥（slug 撞车）。只需某专家请先移除整包目录再装", *pack, *dest)
			}
		}
	}
	fmt.Printf("包就绪：%s（%s，来自 %s）\n", name, packDir, *srcURL)

	if len(experts) > 0 {
		for _, e := range experts {
			dir, dname, dcleanup, err := deriveExpertPack(packDir, e, *srcURL)
			if err != nil {
				return err
			}
			fmt.Printf("专家迷你包就绪：%s（自 %s 闭包派生）\n", dname, name)
			if *api != "" {
				if err := apiInstall(*api, *token, *user, *pass, dir, dname); err != nil {
					dcleanup()
					return fmt.Errorf("专家 %s 安装失败: %w", e, err)
				}
			} else {
				if err := destInstall(*dest, dir, dname, *srcURL); err != nil {
					dcleanup()
					return fmt.Errorf("专家 %s 落盘失败: %w", e, err)
				}
			}
			dcleanup()
		}
		return nil
	}
```

（`packVer` 变量在 Task 1 已就位——本任务它未被再消费也无妨，若 go vet 报 unused 则删除该返回值接收改为 `_`。注意 readPackMeta 调用点在预检**之前**保持原顺序即可。）文件头 `// install` 注释的"两种模式"段落后补一行：

```
// 专家级子集安装：--expert <目录名>（可重复，须与 --pack 同用）——按闭包合成
// 迷你包（<pack>.<expert>）走同一安装管线，独立升级/卸载/启停；与整包安装互斥。
```

Run: `cd PLAT && gofmt -w cmd/bq-markettool/ && go test ./cmd/bq-markettool/ -v`
Expected: 全 PASS

- [ ] **Step 3: 提交（PLAT 仓）**

```bash
cd PLAT && git status && git add cmd/bq-markettool/install.go cmd/bq-markettool/install_test.go
git pull --rebase && git commit -m "feat(markettool): install --expert 接线——闭包派生迷你包逐个独立安装（登记面/落盘双预检拒整包并存；不指定 --expert 行为不变）"
```

---

### Task 3: HUB——validator「专家可独立安装性」检查层

**Files:**
- Modify: `HUB/validator/main.go`（hostSideMCP 白名单；validator 结构体两字段+init；checkPack 收集 packMCP/expertRefs；checkExpertInstall 新 pass；run() 挂调用）
- Test: `HUB/validator/main_test.go`（追加 TestCheckExpertInstall）

**Interfaces:**
- Consumes: 既有 `v.skillOwner`（技能名→包）、`v.expertTools`（slug→tools servers）、`skillRefs` 局部变量、`toolServers` 局部变量。
- Produces: `checkExpertInstall()`（无入参，消费 validator 状态）；`hostSideMCP` 包级白名单。

- [ ] **Step 1: 写测试（红）**

`main_test.go` 末尾追加：

```go
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
	var errs []string
	for _, f := range findings {
		if f.level == "ERROR" {
			errs = append(errs, f.msg)
		}
	}
	joined := strings.Join(errs, "\n")
	if !strings.Contains(joined, "跨包技能") || !strings.Contains(joined, "foreign") {
		t.Fatalf("跨包技能引用应 ERROR: %s", joined)
	}
	if !strings.Contains(joined, "ghost-mcp") || !strings.Contains(joined, "宿主面") {
		t.Fatalf("无契约非宿主面 server 应 ERROR: %s", joined)
	}
	if strings.Contains(joined, "ask-ops") {
		t.Fatalf("宿主面 server 不应报错: %s", joined)
	}
}
```

Run: `cd HUB && go test ./validator/ -run TestCheckExpertInstall -v`
Expected: FAIL——两处 ERROR 未产生（`跨包技能引用应 ERROR`）

- [ ] **Step 2: 实现**

`validator/main.go`：

① 包级白名单（放 `routeGroups` 等 map 声明附近）：

```go
// hostSideMCP 平台宿主侧工具面白名单：包内无契约文件、实例装载即存在的 server。
// 专家级子集安装（spec 2026-10-05-per-expert-install）据此裁决 tools server 可携带性
// ——白名单外且无包内 mcp/ 契约 = 该专家装不出去。新增宿主面 server 与平台侧同步。
var hostSideMCP = map[string]bool{
	"ask-ops":            true, // 宿主侧受审只读命令面
	"security-assistant": true, // 安全巡检采集面（内置）
	"datasources":        true, // 平台数据源查询面（n8n/loki 等）
}
```

② validator 结构体加两字段（挨着 `expertTools` 声明处）：

```go
	// 专家独立安装性核对输入（spec 2026-10-05-per-expert-install §6）
	packMCP    map[string]map[string]bool // 包名 → 包内 mcp/ 契约 server 集
	expertRefs map[string]expertRef       // 专家 slug → 闭包引用快照
```

③ 类型（挨着 `kwFirst`）：

```go
// expertRef 专家闭包引用快照（checkExpertInstall 的核对输入）。
type expertRef struct {
	pack   string
	skills []string
	tools  []string
}
```

④ `run()` 的 validator 复合字面量加两行 init：

```go
		packMCP:      map[string]map[string]bool{},
		expertRefs:   map[string]expertRef{},
```

⑤ `checkPack` 专家循环内，`v.expertTools[d.Slug] = toolServers` 之后加：

```go
		v.expertRefs[d.Slug] = expertRef{pack: packName, skills: skillRefs[d.Slug], tools: toolServers}
```

`checkPack` 中技能目录扫描之后（`for name := range present` 未登记 WARN 块附近）加包内契约集收集：

```go
	// 包内 MCP 契约集（专家独立安装性核对：tools server 须有包内契约或属宿主面）
	if mcpDirs, _ := os.ReadDir(filepath.Join(packDir, "mcp")); mcpDirs != nil {
		set := map[string]bool{}
		for _, e := range mcpDirs {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".yaml") {
				set[strings.TrimSuffix(e.Name(), ".yaml")] = true
			}
		}
		v.packMCP[packName] = set
	}
```

⑥ 新 pass（放 `checkCrossRefs` 之后）：

```go
// checkExpertInstall 专家可独立安装性（spec 2026-10-05-per-expert-install §6）：
// 子集安装按闭包合成迷你包，闭包只含本包资产——跨包技能引用、无契约又非宿主面的
// tools server 都让该专家装不出去（安装器会拒），CI 提前拦。
func (v *validator) checkExpertInstall() {
	for slug, ref := range v.expertRefs {
		for _, s := range ref.skills {
			if owner, ok := v.skillOwner[s]; ok && owner != ref.pack {
				v.add("ERROR", ref.pack, "专家 %s 引用跨包技能 %s（属 %s）——专家级子集安装无法携带，须整包安装；如确需跨包挂载请在 provenance 说明并走整包设计", slug, s, owner)
			}
		}
		for _, srv := range ref.tools {
			if v.packMCP[ref.pack][srv] || hostSideMCP[srv] {
				continue
			}
			v.add("ERROR", ref.pack, "专家 %s tools server %q 无包内契约（mcp/%s.yaml）且不属宿主面白名单——专家级安装无法携带契约清单", slug, srv, srv)
		}
	}
}
```

⑦ `run()` 中 `v.checkCrossRefs()` 之后加一行：

```go
	v.checkExpertInstall()
```

Run: `cd HUB && gofmt -w validator/ && go test ./validator/ -v`
Expected: 全 PASS

- [ ] **Step 3: 全仓回归——现存 16 专家必须全绿**

Run: `cd HUB && go run ./validator ./packs && go run ./validator`
Expected: `0 ERROR`（WARN 数不增）。若曝出 ERROR：属现存包真缺口（预期没有），停下报告——修包内容升版本走独立提交，不混入本任务。

- [ ] **Step 4: 提交（HUB 仓）**

```bash
cd HUB && git status && git add validator/main.go validator/main_test.go
git pull --rebase && git commit -m "feat(validator): 专家可独立安装性检查层——跨包技能引用/无契约非宿主面 tools server 拦为 ERROR（宿主面白名单 ask-ops/security-assistant/datasources），配套 bq-markettool --expert 子集安装"
```

---

### Task 4: 文档同步（HUB + PLAT）

**Files:**
- Modify: `HUB/README.md`（「装进扁鹊」节，install 示例块之后）
- Modify: `HUB/CONTRIBUTING.md`（硬规则追加第 11 条）
- Modify: `HUB/knowledge/platform-matrix.md`（仅当其中载有 bq-markettool 安装说明时同步一句——先 grep 判定）
- Modify: PLAT 仓中载有 bq-markettool 安装说明的文档（先 `grep -rln "bq-markettool" PLAT/docs PLAT/README.md` 定位；通常 `docs/PACK_GUIDE.md` 或主 README）

**Interfaces:** 无代码接口；纯文档。

- [ ] **Step 1: HUB README「装进扁鹊」追加（紧跟 curl API 示例代码块之后、`已装过的包…` 段落之前）**

```markdown
只装某个专家（专家级子集安装）：

```bash
go run ./cmd/bq-markettool install \
  --url https://github.com/yi-nology/bianque-hub \
  --pack k8s-ops --expert k8s-workload-analyst \
  --api http://127.0.0.1:8900
```

`--expert` 可重复；每个专家安装为**独立插件**（`<包>.<专家>`，如
`k8s-ops.k8s-workload-analyst`），独立升级/卸载/启停。口径：与整包安装**互斥**
（同 slug 撞车，安装器预检拒绝——已装整包时如只需某专家，用插件页专家级启停，
或先卸整包）；链（chain.yaml）与跨专家消歧是包级资产，不随专家级安装落位。
```

- [ ] **Step 2: CONTRIBUTING 硬规则追加第 11 条（第 10 条之后）**

```markdown
11. **专家独立安装性**：专家默认须可独立安装（`bq-markettool install --expert`
    按闭包合成迷你包）——`agent.yaml` 的 `skills:` 引用必须在本包（跨包=CI 拦）；
    `tools[].server` 须有包内 `mcp/<server>.yaml` 契约或属宿主面白名单
    （ask-ops / security-assistant / datasources；新增宿主面 server 与平台侧同步）。
    链/消歧是包级资产，不进专家闭包。
```

- [ ] **Step 3: 知识副本与平台仓文档同步**

平台仓已定位（grep 实查）：`PLAT/docs/PACK_GUIDE.md` 与
`PLAT/docs/OO_DEVOPS_MIGRATION_RUNBOOK.md` 载有 `bq-markettool install`。改前者
（活契约文档）；**runbook 是历史迁移记录，不改**。hub 侧先判定知识副本：

```bash
grep -n "bq-markettool" HUB/knowledge/platform-matrix.md
```

`PLAT/docs/PACK_GUIDE.md` 中 bq-markettool 安装说明段落后追加：

```markdown
专家级子集安装：`install --pack <包> --expert <专家目录名>`（可重复，须与
--pack 同用）——按闭包合成迷你包（`<包>.<专家>`）走同一插件管线，独立
升级/卸载/启停。与整包安装互斥（slug 撞车，安装器预检拒绝）；链（chain.yaml）
与跨专家消歧是包级资产，不随专家级安装落位。
```

`HUB/knowledge/platform-matrix.md` 若 grep 命中安装说明，同位置补上述精简段；
未命中（该副本只载运行时事实）则不动。

Run: `cd HUB && go run ./validator && go run ./validator ./packs`
Expected: 仍 `0 ERROR`（README 校验层 `checkReadme` 不受影响；若包清单表/版本规则被误触，按其报错修文案）

- [ ] **Step 4: 提交（HUB、PLAT 各一笔）**

```bash
cd HUB && git add README.md CONTRIBUTING.md && git commit -m "docs: 专家级子集安装口径——--expert 用例、硬规则 11（独立安装性契约）"
git add knowledge/platform-matrix.md 2>/dev/null && git commit --amend --no-edit   # 仅当该文件有改动
cd PLAT && git add docs/PACK_GUIDE.md && git commit -m "docs: bq-markettool install --expert 专家级子集安装口径（PACK_GUIDE）"
```

（amend 步骤在 platform-matrix 无改动时跳过；`git status` 确认后再执行。）

---

### Task 5: 全量回归 + 推送

**Files:** 无新改动；两仓既有提交的验证与推送。

- [ ] **Step 1: HUB 回归**

```bash
cd HUB && gofmt -l . && go test ./... && go run ./validator ./packs
```
Expected: gofmt 无输出；测试全过；validator `0 ERROR`。

- [ ] **Step 2: PLAT 回归**

```bash
cd PLAT && gofmt -l cmd/bq-markettool/ && go test ./cmd/bq-markettool/... && go build ./...
```
Expected: 无输出；测试全过；构建成功。

- [ ] **Step 3: 8900 实装冒烟（实例在跑时；没跑则跳过并在总结注明）**

```bash
cd PLAT && go run ./cmd/bq-markettool install \
  --url https://github.com/yi-nology/bianque-hub \
  --pack k8s-ops --expert k8s-workload-analyst --api http://127.0.0.1:8900
# 预期：插件页出现 k8s-ops.k8s-workload-analyst；路由可达；卸载回净。
# 另验互斥：整包 k8s-ops 在位时同一命令应被预检拒绝。
```

- [ ] **Step 4: 推送（hub 443 通道惯例；先 rebase 防并发）**

```bash
cd HUB && git pull --rebase && git push
cd PLAT && git pull --rebase && git push
```

若 PLAT 仓无远端推送权限/惯例不同，停在本地提交并在总结注明。
