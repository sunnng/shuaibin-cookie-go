# 脚本 UI 框架 Phase 1（纯逻辑核心） Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在本模块新建顶层公开包 `ui/`（import 路径 `app/ui`），交付脚本 UI 框架的全部纯逻辑核心：Store、ScriptController、字段/任务描述符、组件状态（Ctx）、主题数据、缩放、Shell 实例——全部无构建标签、全部可本地单测。

**Architecture:** 依据 ADR-0002（描述符接缝、实例化 Shell、应用→框架单向依赖）与 ADR-0003（函数组件模型：Props 结构体 + 路径显式键托管状态）。本阶段只建无标签纯逻辑文件；`internal/ui` 保持原样继续服役（复制而非移动，Phase 3 才删除）；任何文件不得 import `github.com/Dasongzi1366/AutoGo/imgui`（绘制是 Phase 2 的事）。

**Tech Stack:** Go 1.25.0（模块 `app`，泛型可用）、标准库 only。

## Global Constraints

- 新代码全部在仓库根 `ui/` 目录，包名 `ui`，import 路径 `app/ui`。**不得改动 `internal/ui` 的任何文件。**
- `ui/` 下所有文件**无构建标签**，且不得 import `github.com/Dasongzi1366/AutoGo/imgui`（允许 `app/internal/logger`，它是同模块叶子工具）。
- 术语遵守 CONTEXT.md：标识符禁用 Module/Session；机制名允许 runtime/Store/Shell。
- 先建分支：`git checkout -b feat/script-ui-framework`（当前在 main）。
- 所有命令在仓库根 `c:\Users\1\Desktop\pr\shuaibin-cookie-go` 执行；测试命令 `go test ./ui/...`，全量验证 `go build ./... && go test ./...`。
- gofmt 只作用于新目录：`gofmt -w ui/`（本仓库 CRLF 存量文件勿碰）。
- 提交信息格式：`feat(ui): <中文摘要>`，结尾附 `Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>`。
- Phase 2（android 绘制层）与 Phase 3（应用接线、删 internal/ui）不在本计划内，各自另行成计划。

---

### Task 1: 包骨架 + Store 与 ClearPanelCache 迁移

**Files:**
- Create: `ui/doc.go`
- Create: `ui/store.go`（内容 = `internal/ui/store.go` 逐字复制）
- Create: `ui/panel_cache.go`（内容 = `internal/ui/panel_logic.go` 逐字复制）
- Test: `ui/store_test.go`

**Interfaces:**
- Consumes: 无
- Produces: `Store`、`NewStore()`、`GetBool/SetBool/GetString/SetString/GetFloat/SetFloat/HasKey/ToJSON/SaveConfig/LoadConfig/Clear`、`ClearPanelCache(store *Store, configPath, dataStorePath string, reseed func(*Store)) error`

- [ ] **Step 1: 写失败测试**

`ui/store_test.go`（在 `internal/ui/store_test.go` 基础上把 `KeyArenaEnabled` 换成字面量，框架包不得引用应用键）：

```go
package ui

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStoreClear(t *testing.T) {
	s := NewStore()
	s.SetBool("a", true)
	s.SetFloat("b", 1.5)
	s.Clear()
	if s.HasKey("a") || s.HasKey("b") {
		t.Fatalf("Clear should remove all keys")
	}
}

func TestStoreSaveLoadRoundtrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ui.json")
	s := NewStore()
	s.SetBool("flag", true)
	s.SetFloat("num", 42)
	s.SetString("name", "arena")
	if err := s.SaveConfig(path); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}
	s2 := NewStore()
	if err := s2.LoadConfig(path); err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if !s2.GetBool("flag") || s2.GetFloat("num") != 42 || s2.GetString("name") != "arena" {
		t.Fatalf("roundtrip mismatch: %#v", s2)
	}
}

func TestClearPanelCache(t *testing.T) {
	dir := t.TempDir()
	uiPath := filepath.Join(dir, "ui.json")
	kvPath := filepath.Join(dir, "store.json")
	if err := os.WriteFile(uiPath, []byte(`{"arena_enabled":true}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(kvPath, []byte(`{"k":1}`), 0o644); err != nil {
		t.Fatal(err)
	}

	s := NewStore()
	s.SetBool("arena_enabled", true)
	reseeded := false
	if err := ClearPanelCache(s, uiPath, kvPath, func(st *Store) {
		reseeded = true
		st.SetBool("arena_enabled", false)
	}); err != nil {
		t.Fatalf("ClearPanelCache: %v", err)
	}
	if !reseeded {
		t.Fatal("expected reseed")
	}
	if s.GetBool("arena_enabled") {
		t.Fatal("reseed should set arena_enabled false")
	}
	if _, err := os.Stat(uiPath); !os.IsNotExist(err) {
		t.Fatalf("ui.json should be removed, err=%v", err)
	}
	if _, err := os.Stat(kvPath); !os.IsNotExist(err) {
		t.Fatalf("store.json should be removed, err=%v", err)
	}
}
```

- [ ] **Step 2: 跑测试确认失败**

Run: `go test ./ui/...`
Expected: FAIL（`ui/store_test.go:9:8: undefined: NewStore` 等编译错误）

- [ ] **Step 3: 写实现**

`ui/doc.go`:

```go
// Package ui 是脚本 UI 框架（ADR-0002/0003）：灵动岛、配置面板、配置绑定、
// 函数组件模型与托管状态。应用以描述符（Task/Field/NavEntry）与框架接缝，
// 框架不感知具体游戏与配置类型，依赖方向只能 应用→框架。
//
// 本包所有无构建标签文件为纯逻辑，可本地测试；绘制层在 *_android.go
// （//go:build android && cgo），只做薄绘制。
package ui
```

`ui/store.go`：逐字复制 `internal/ui/store.go`（包名同为 `ui`，无需任何改动）。

`ui/panel_cache.go`：逐字复制 `internal/ui/panel_logic.go`（函数 `ClearPanelCache`，包名同为 `ui`）。

- [ ] **Step 4: 跑测试确认通过**

Run: `go test ./ui/... -v`
Expected: PASS（3 个测试全过）

- [ ] **Step 5: 提交**

```bash
git checkout -b feat/script-ui-framework
git add ui/
git commit -m "feat(ui): 框架包骨架与 Store/ClearPanelCache 迁移

Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>"
```

---

### Task 2: ScriptController 迁移

**Files:**
- Create: `ui/controller.go`（内容 = `internal/ui/controller.go` 逐字复制）
- Test: `ui/controller_test.go`（内容 = `internal/ui/controller_test.go` 逐字复制）

**Interfaces:**
- Consumes: `app/internal/logger`
- Produces: `ScriptState`（`StateIdle/StateRunning/StatePaused`）、`ScriptHooks{OnStart, OnExit}`、`ScriptController`、`NewScriptController(hooks)`、`State()/Start()/Pause()/Resume()/Stop()/Exit()`

- [ ] **Step 1: 复制测试**

`ui/controller_test.go` = `internal/ui/controller_test.go` 逐字复制（无应用依赖，3 个测试：`TestScriptControllerStateTransitions`、`TestScriptControllerRunEndReturnsIdle`、`TestScriptControllerStopBlocksStartUntilRunEnds`）。

- [ ] **Step 2: 跑测试确认失败**

Run: `go test ./ui/ -run TestScriptController -v`
Expected: FAIL（undefined: NewScriptController）

- [ ] **Step 3: 复制实现**

`ui/controller.go` = `internal/ui/controller.go` 逐字复制（保留 `app/internal/logger` import——同模块叶子工具，框架可用）。

- [ ] **Step 4: 跑测试确认通过**

Run: `go test ./ui/ -v`
Expected: PASS（含 Task 1 的 3 个，共 6 个）

- [ ] **Step 5: 提交**

```bash
git add ui/controller.go ui/controller_test.go
git commit -m "feat(ui): ScriptController 迁入框架包

Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>"
```

---

### Task 3: Field 字段描述符

**Files:**
- Create: `ui/field.go`
- Test: `ui/field_test.go`

**Interfaces:**
- Consumes: `Store`（Task 1）
- Produces:
  - `WidgetKind`（`WidgetCheckbox`/`WidgetNumberInput`/`WidgetTextInput`），`func (k WidgetKind) String() string`
  - `NumberConstraints struct{ Min, Max, Step float64 }`
  - `Field` 接口：`Key() string`、`Label() string`、`Widget() WidgetKind`、`Constraints() NumberConstraints`、`Seed(*Store)`、`Apply(*Store)`
  - 构造函数：`Bool(key, label string, get func() bool, set func(bool)) Field`、`Number(key, label string, min, max, step float64, get func() int, set func(int)) Field`、`Text(key, label string, get func() string, set func(string)) Field`

- [ ] **Step 1: 写失败测试**

`ui/field_test.go`:

```go
package ui

import "testing"

func TestBoolFieldSeedAndApply(t *testing.T) {
	backing := true
	f := Bool("arena_enabled", "启用", func() bool { return backing }, func(v bool) { backing = v })

	s := NewStore()
	f.Seed(s)
	if !s.GetBool("arena_enabled") {
		t.Fatal("seed should write default when key missing")
	}

	s.SetBool("arena_enabled", false)
	f.Seed(s)
	if s.GetBool("arena_enabled") {
		t.Fatal("seed must not overwrite existing key")
	}

	f.Apply(s)
	if backing != false {
		t.Fatalf("apply should write store value back, backing=%v", backing)
	}
	if f.Widget() != WidgetCheckbox {
		t.Fatalf("widget=%v want checkbox", f.Widget())
	}
}

func TestNumberFieldStoresFloatAndAppliesInt(t *testing.T) {
	backing := 3
	f := Number("arena_max_battles", "战斗上限", 0, 99, 1,
		func() int { return backing }, func(v int) { backing = v })

	s := NewStore()
	f.Seed(s)
	if got := s.GetFloat("arena_max_battles"); got != 3 {
		t.Fatalf("seed stored %v want 3", got)
	}

	s.SetFloat("arena_max_battles", 7.9)
	f.Apply(s)
	if backing != 7 {
		t.Fatalf("apply truncated float to int, backing=%d want 7", backing)
	}

	c := f.Constraints()
	if c.Min != 0 || c.Max != 99 || c.Step != 1 {
		t.Fatalf("constraints=%+v", c)
	}
	if f.Widget() != WidgetNumberInput {
		t.Fatalf("widget=%v want number", f.Widget())
	}
}

func TestTextFieldAndNilStoreSafety(t *testing.T) {
	backing := "x"
	f := Text("note", "备注", func() string { return backing }, func(v string) { backing = v })
	s := NewStore()
	f.Seed(s)
	if s.GetString("note") != "x" {
		t.Fatal("seed string")
	}
	s.SetString("note", "y")
	f.Apply(s)
	if backing != "y" {
		t.Fatalf("apply string, backing=%q", backing)
	}
	if f.Widget() != WidgetTextInput {
		t.Fatalf("widget=%v want text", f.Widget())
	}

	// nil Store 不得 panic
	f.Seed(nil)
	f.Apply(nil)
}

func TestWidgetKindString(t *testing.T) {
	if WidgetCheckbox.String() != "checkbox" || WidgetNumberInput.String() != "number" || WidgetTextInput.String() != "text" {
		t.Fatal("widget kind strings")
	}
}
```

- [ ] **Step 2: 跑测试确认失败**

Run: `go test ./ui/ -run "TestBoolField|TestNumberField|TestTextField|TestWidgetKind" -v`
Expected: FAIL（undefined: Bool / Number / Text / WidgetCheckbox）

- [ ] **Step 3: 写实现**

`ui/field.go`:

```go
package ui

// WidgetKind 字段在面板中使用的控件形态。
type WidgetKind int

const (
	WidgetCheckbox WidgetKind = iota
	WidgetNumberInput
	WidgetTextInput
)

func (k WidgetKind) String() string {
	switch k {
	case WidgetCheckbox:
		return "checkbox"
	case WidgetNumberInput:
		return "number"
	case WidgetTextInput:
		return "text"
	default:
		return "unknown"
	}
}

// NumberConstraints 数字字段的取值约束，供 Form 渲染 NumberInput 使用。
type NumberConstraints struct {
	Min, Max, Step float64
}

// Field 一项配置的唯一声明处（CONTEXT.md「字段」）：种子写入、回写配置、
// 面板渲染都由它推导。实现不可变，构造后安全共享。
type Field interface {
	Key() string
	Label() string
	Widget() WidgetKind
	Constraints() NumberConstraints
	// Seed 在 store 缺少该键时写入应用配置的当前值。
	Seed(*Store)
	// Apply 把 store 值写回应用配置。
	Apply(*Store)
}

type field[T any] struct {
	key    string
	label  string
	widget WidgetKind
	cons   NumberConstraints
	get    func() T
	set    func(T)
	sget   func(*Store, string) T
	sset   func(*Store, string, T)
}

func (f field[T]) Key() string                  { return f.key }
func (f field[T]) Label() string                { return f.label }
func (f field[T]) Widget() WidgetKind           { return f.widget }
func (f field[T]) Constraints() NumberConstraints { return f.cons }

func (f field[T]) Seed(s *Store) {
	if s == nil || f.get == nil {
		return
	}
	if !s.HasKey(f.key) {
		f.sset(s, f.key, f.get())
	}
}

func (f field[T]) Apply(s *Store) {
	if s == nil || f.set == nil {
		return
	}
	f.set(f.sget(s, f.key))
}

// Bool 声明一个布尔字段（面板渲染为复选框）。
func Bool(key, label string, get func() bool, set func(bool)) Field {
	return field[bool]{
		key: key, label: label, widget: WidgetCheckbox, get: get, set: set,
		sget: func(s *Store, k string) bool { return s.GetBool(k) },
		sset: func(s *Store, k string, v bool) { s.SetBool(k, v) },
	}
}

// Number 声明一个整数字段（面板渲染为步进数字输入框；store 中以 float64 存放，
// 读写时与 int 互转）。min/max/step 仅供渲染层约束，Seed/Apply 不做钳制。
func Number(key, label string, min, max, step float64, get func() int, set func(int)) Field {
	return field[int]{
		key: key, label: label, widget: WidgetNumberInput,
		cons: NumberConstraints{Min: min, Max: max, Step: step}, get: get, set: set,
		sget: func(s *Store, k string) int { return int(s.GetFloat(k)) },
		sset: func(s *Store, k string, v int) { s.SetFloat(k, float64(v)) },
	}
}

// Text 声明一个字符串字段（面板渲染为文本输入框）。
func Text(key, label string, get func() string, set func(string)) Field {
	return field[string]{
		key: key, label: label, widget: WidgetTextInput, get: get, set: set,
		sget: func(s *Store, k string) string { return s.GetString(k) },
		sset: func(s *Store, k string, v string) { s.SetString(k, v) },
	}
}
```

- [ ] **Step 4: 跑测试确认通过**

Run: `go test ./ui/ -v`
Expected: PASS（共 10 个）

- [ ] **Step 5: 提交**

```bash
git add ui/field.go ui/field_test.go
git commit -m "feat(ui): 字段描述符（Bool/Number/Text）与 Seed/Apply 推导

Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>"
```

---

### Task 4: Task 描述符与列表逻辑

**Files:**
- Create: `ui/task.go`
- Test: `ui/task_test.go`

**Interfaces:**
- Consumes: `Field`（Task 3）、`Store`（Task 1）、`Ctx`（Task 5 定义 `RenderDetail func(*Ctx)` 的参数类型——本任务只引用类型名，Task 5 落地后编译；**实施顺序保证 Task 5 的 `Ctx` 已存在，或将本任务排在 Task 5 之后**）
- Produces:
  - `Task struct{ ID, Title, Category, EnabledKey string; Fields []Field; Summary func(*Store) string; RenderDetail func(*Ctx) }`
  - `Categories(tasks []Task) []string`（首次出现顺序去重，跳过空串）
  - `FilterByCategory(tasks []Task, cat string) []Task`、`FindTask(tasks []Task, id string) (Task, bool)`、`CountEnabled(store *Store, tasks []Task) (enabled, total int)`
  - `SeedAll(store *Store, tasks []Task)`、`ApplyAll(store *Store, tasks []Task)`
  - 面板偏好键：`KeyPanelNav = "panel_nav"`、`KeyPanelCat = "panel_cat"`、`KeyPanelSelected = "panel_selected"`、`PanelCatAll = "all"`
  - `SeedPanelDefaults(store *Store, tasks []Task, navIDs []string)`

> **实施注意**：`Task.RenderDetail` 引用 `*Ctx`（Task 5）。执行时**先完成 Task 5 再做本任务**，或在本任务内临时不含 `RenderDetail` 字段、Task 5 时补上——推荐前者（调整执行顺序为 1→2→3→5→4→6→7→8→9）。

- [ ] **Step 1: 写失败测试**

`ui/task_test.go`:

```go
package ui

import "testing"

func testTasks() []Task {
	return []Task{
		{ID: "arena", Title: "王国竞技场", Category: "日常", EnabledKey: "arena_enabled"},
		{ID: "tower", Title: "混沌塔", Category: "日常", EnabledKey: "tower_enabled"},
		{ID: "raid", Title: "讨伐", Category: "活动", EnabledKey: "raid_enabled"},
		{ID: "nocat", Title: "未分类", Category: "", EnabledKey: "nocat_enabled"},
	}
}

func TestCategoriesDedupKeepsOrder(t *testing.T) {
	got := Categories(testTasks())
	if len(got) != 2 || got[0] != "日常" || got[1] != "活动" {
		t.Fatalf("Categories=%v", got)
	}
	if got := Categories(nil); len(got) != 0 {
		t.Fatalf("Categories(nil)=%v", got)
	}
}

func TestFilterByCategory(t *testing.T) {
	tasks := testTasks()
	if got := FilterByCategory(tasks, PanelCatAll); len(got) != 4 {
		t.Fatalf("all len=%d want 4", len(got))
	}
	if got := FilterByCategory(tasks, ""); len(got) != 4 {
		t.Fatalf("empty cat len=%d want 4", len(got))
	}
	if got := FilterByCategory(tasks, "日常"); len(got) != 2 {
		t.Fatalf("日常 len=%d want 2", len(got))
	}
	if got := FilterByCategory(tasks, "维护"); len(got) != 0 {
		t.Fatalf("维护 len=%d want 0", len(got))
	}
}

func TestFindTaskAndCountEnabled(t *testing.T) {
	tasks := testTasks()
	m, ok := FindTask(tasks, "arena")
	if !ok || m.Title != "王国竞技场" {
		t.Fatalf("FindTask arena: %#v ok=%v", m, ok)
	}
	if _, ok := FindTask(tasks, "nope"); ok {
		t.Fatal("expected missing")
	}

	store := NewStore()
	store.SetBool("arena_enabled", true)
	store.SetBool("raid_enabled", true)
	en, total := CountEnabled(store, tasks)
	if en != 2 || total != 4 {
		t.Fatalf("CountEnabled=%d/%d want 2/4", en, total)
	}
	if en, total := CountEnabled(nil, tasks); en != 0 || total != 4 {
		t.Fatalf("CountEnabled nil store=%d/%d", en, total)
	}
}

func TestSeedAllAndApplyAll(t *testing.T) {
	enabled := false
	maxBattles := 10
	tasks := []Task{
		{ID: "arena", Fields: []Field{
			Bool("arena_enabled", "启用", func() bool { return enabled }, func(v bool) { enabled = v }),
			Number("arena_max_battles", "上限", 0, 99, 1, func() int { return maxBattles }, func(v int) { maxBattles = v }),
		}},
	}
	s := NewStore()
	SeedAll(s, tasks)
	if s.GetFloat("arena_max_battles") != 10 {
		t.Fatal("SeedAll should seed number field")
	}
	s.SetBool("arena_enabled", true)
	s.SetFloat("arena_max_battles", 20)
	ApplyAll(s, tasks)
	if !enabled || maxBattles != 20 {
		t.Fatalf("ApplyAll: enabled=%v max=%d", enabled, maxBattles)
	}
}

func TestSeedPanelDefaults(t *testing.T) {
	store := NewStore()
	SeedPanelDefaults(store, testTasks(), []string{"tasks", "system"})
	if store.GetString(KeyPanelNav) != "tasks" {
		t.Fatalf("nav=%q", store.GetString(KeyPanelNav))
	}
	if store.GetString(KeyPanelCat) != PanelCatAll {
		t.Fatalf("cat=%q", store.GetString(KeyPanelCat))
	}
	if store.GetString(KeyPanelSelected) != "arena" {
		t.Fatalf("selected=%q", store.GetString(KeyPanelSelected))
	}

	store.SetString(KeyPanelNav, "system")
	SeedPanelDefaults(store, testTasks(), []string{"tasks", "system"})
	if store.GetString(KeyPanelNav) != "system" {
		t.Fatal("should keep existing nav")
	}

	empty := NewStore()
	SeedPanelDefaults(empty, nil, nil)
	if empty.HasKey(KeyPanelNav) || empty.HasKey(KeyPanelSelected) {
		t.Fatal("no tasks/nav -> no defaults")
	}
	SeedPanelDefaults(nil, testTasks(), nil) // nil store 不得 panic
}
```

- [ ] **Step 2: 跑测试确认失败**

Run: `go test ./ui/ -run "TestCategories|TestFilter|TestFindTask|TestSeedAll|TestSeedPanelDefaults" -v`
Expected: FAIL（undefined: Categories / PanelCatAll / SeedAll ...）

- [ ] **Step 3: 写实现**

`ui/task.go`:

```go
package ui

// Task 任务描述符（CONTEXT.md「描述符」）：应用启动时显式构造并交给框架，
// 框架据此驱动面板列表、分类 chips、配置绑定与详情页渲染。
// RenderDetail 为 nil 时详情页按 Fields 自动渲染 Form；非 nil 时为自定义
// section 逃生门（ADR-0003），内部可自行组合组件（包括复用 Form）。
type Task struct {
	ID         string
	Title      string
	Category   string // 自由字符串，即 chip 展示文本；空串不参与 chips
	EnabledKey string
	Fields     []Field
	// Summary 列表摘要行（如「已战斗 12 场」），可为 nil。
	Summary func(*Store) string
	// RenderDetail 自定义详情渲染（绘制层，Phase 2 起使用），可为 nil。
	RenderDetail func(*Ctx)
}

// Categories 按首次出现顺序返回去重后的非空分类。
func Categories(tasks []Task) []string {
	seen := map[string]bool{}
	var out []string
	for _, t := range tasks {
		if t.Category == "" || seen[t.Category] {
			continue
		}
		seen[t.Category] = true
		out = append(out, t.Category)
	}
	return out
}

// 面板偏好键（随 ui.json 落盘；非业务配置）。
const (
	KeyPanelNav      = "panel_nav"
	KeyPanelCat      = "panel_cat"
	KeyPanelSelected = "panel_selected"

	PanelCatAll = "all"
)

// FilterByCategory cat 为空或 PanelCatAll 时返回全部。
func FilterByCategory(tasks []Task, cat string) []Task {
	if cat == "" || cat == PanelCatAll {
		return append([]Task(nil), tasks...)
	}
	out := make([]Task, 0, len(tasks))
	for _, t := range tasks {
		if t.Category == cat {
			out = append(out, t)
		}
	}
	return out
}

// FindTask 按 ID 查找。
func FindTask(tasks []Task, id string) (Task, bool) {
	for _, t := range tasks {
		if t.ID == id {
			return t, true
		}
	}
	return Task{}, false
}

// CountEnabled 统计 EnabledKey 为 true 的任务数。
func CountEnabled(store *Store, tasks []Task) (enabled, total int) {
	total = len(tasks)
	if store == nil {
		return 0, total
	}
	for _, t := range tasks {
		if t.EnabledKey != "" && store.GetBool(t.EnabledKey) {
			enabled++
		}
	}
	return enabled, total
}

// SeedAll 对全部任务的全部字段执行 Seed（仅填缺失键）。
func SeedAll(store *Store, tasks []Task) {
	if store == nil {
		return
	}
	for _, t := range tasks {
		for _, f := range t.Fields {
			f.Seed(store)
		}
	}
}

// ApplyAll 对全部任务的全部字段执行 Apply（写回应用配置）。
func ApplyAll(store *Store, tasks []Task) {
	if store == nil {
		return
	}
	for _, t := range tasks {
		for _, f := range t.Fields {
			f.Apply(store)
		}
	}
}

// SeedPanelDefaults 填充面板偏好默认值（仅缺失键）：导航取 navIDs[0]、
// 分类取全部、选中取首个任务。无任务/无导航则不写对应键。
func SeedPanelDefaults(store *Store, tasks []Task, navIDs []string) {
	if store == nil {
		return
	}
	if !store.HasKey(KeyPanelNav) && len(navIDs) > 0 {
		store.SetString(KeyPanelNav, navIDs[0])
	}
	if !store.HasKey(KeyPanelCat) {
		store.SetString(KeyPanelCat, PanelCatAll)
	}
	if (!store.HasKey(KeyPanelSelected) || store.GetString(KeyPanelSelected) == "") && len(tasks) > 0 {
		store.SetString(KeyPanelSelected, tasks[0].ID)
	}
}
```

- [ ] **Step 4: 跑测试确认通过**

Run: `go test ./ui/ -v`
Expected: PASS（含 Task 5 的测试，数量视执行顺序而定）

- [ ] **Step 5: 提交**

```bash
git add ui/task.go ui/task_test.go
git commit -m "feat(ui): 任务描述符与列表/分类/面板偏好逻辑

Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>"
```

---

### Task 5: Ctx 与组件状态（路径 + 显式键）

**Files:**
- Create: `ui/context.go`
- Test: `ui/context_test.go`

**Interfaces:**
- Consumes: `Store`（Task 1）
- Produces:
  - `Ctx struct{ Store *Store; Scale float64 }`（含未导出 `path []string`、`states map[string]any`）
  - `NewCtx(store *Store, scale float64) *Ctx`（scale ≤ 0 归一为 1）
  - `(c *Ctx) Push(id string)`、`(c *Ctx) Pop()`、`(c *Ctx) S(base float64) float64`
  - `State[T any](c *Ctx, key string, initial T) *T`（ADR-0003：路径 + 显式键寻址）

- [ ] **Step 1: 写失败测试**

`ui/context_test.go`:

```go
package ui

import "testing"

func TestStatePersistsAcrossFrames(t *testing.T) {
	c := NewCtx(NewStore(), 1)
	c.Push("form")
	p1 := State(c, "draft", "init")
	*p1 = "edited"
	c.Pop()

	// 模拟下一帧：同一路径同一键拿到同一份状态
	c.Push("form")
	p2 := State(c, "draft", "init")
	c.Pop()
	if p2 != p1 || *p2 != "edited" {
		t.Fatalf("state should persist across frames, got %q", *p2)
	}
}

func TestStateIsolatedByComponentPath(t *testing.T) {
	c := NewCtx(NewStore(), 1)

	c.Push("panel")
	c.Push("formA")
	a := State(c, "draft", 0)
	*a = 1
	c.Pop()
	c.Push("formB")
	b := State(c, "draft", 0)
	*b = 2
	c.Pop()
	c.Pop()

	if *a != 1 || *b != 2 {
		t.Fatalf("same component twice must be isolated: a=%d b=%d", *a, *b)
	}

	// 键的类型不同属调用方错误；同键同型返回同一指针
	c.Push("panel")
	c.Push("formA")
	a2 := State(c, "draft", 99)
	c.Pop()
	c.Pop()
	if a2 != a {
		t.Fatal("same path+key must return same pointer")
	}
}

func TestStateConditionalRenderingSafe(t *testing.T) {
	c := NewCtx(NewStore(), 1)
	// 第一帧渲染了可选区块
	c.Push("root")
	State(c, "a", 1)
	c.Push("optional")
	State(c, "b", 2)
	c.Pop()
	c.Pop()
	// 第二帧不渲染可选区块，已有状态不受影响
	c.Push("root")
	if got := *State(c, "a", 0); got != 1 {
		t.Fatalf("a=%d want 1", got)
	}
	c.Pop()
}

func TestCtxScaleAndPopSafety(t *testing.T) {
	c := NewCtx(NewStore(), 1.5)
	if got := c.S(100); got != 150 {
		t.Fatalf("S(100)=%v want 150", got)
	}
	c.Pop() // 空路径 Pop 不得 panic
	if got := NewCtx(NewStore(), 0).Scale; got != 1 {
		t.Fatalf("scale<=0 should normalize to 1, got %v", got)
	}
}
```

- [ ] **Step 2: 跑测试确认失败**

Run: `go test ./ui/ -run "TestState|TestCtxScale" -v`
Expected: FAIL（undefined: NewCtx / State）

- [ ] **Step 3: 写实现**

`ui/context.go`:

```go
package ui

import "strings"

// Ctx 组件帧上下文：每帧贯穿组件树的句柄，携带 Store、缩放系数与当前
// 组件嵌套路径。组件状态经 State 按「路径 + 显式键」托管（ADR-0003）。
// Ctx 只在 UI goroutine 使用，非并发安全。
type Ctx struct {
	Store *Store
	Scale float64

	path   []string
	states map[string]any
}

// NewCtx 创建帧上下文；scale <= 0 归一为 1。
func NewCtx(store *Store, scale float64) *Ctx {
	if scale <= 0 {
		scale = 1
	}
	return &Ctx{Store: store, Scale: scale, states: map[string]any{}}
}

// Push 进入子组件作用域（组件函数惯例：Push(id) + defer Pop()）。
func (c *Ctx) Push(id string) { c.path = append(c.path, id) }

// Pop 离开当前组件作用域；空路径调用安全。
func (c *Ctx) Pop() {
	if n := len(c.path); n > 0 {
		c.path = c.path[:n-1]
	}
}

// S 把基准分辨率（1600×900）下的尺寸换算为设备尺寸。
func (c *Ctx) S(base float64) float64 { return base * c.Scale }

func (c *Ctx) scope() string { return strings.Join(c.path, "/") }

// State 返回当前组件实例内 key 对应的托管状态指针；首次访问写入 initial。
// 同一路径 + 同一键跨帧返回同一指针；不同组件实例（路径不同）各自隔离。
// 规则：同一组件实例内键唯一；条件渲染自由（ADR-0003）。
func State[T any](c *Ctx, key string, initial T) *T {
	full := c.scope() + "\x00" + key
	if v, ok := c.states[full]; ok {
		return v.(*T)
	}
	v := new(T)
	*v = initial
	c.states[full] = v
	return v
}
```

- [ ] **Step 4: 跑测试确认通过**

Run: `go test ./ui/ -v`
Expected: PASS

- [ ] **Step 5: 提交**

```bash
git add ui/context.go ui/context_test.go
git commit -m "feat(ui): Ctx 与路径+显式键托管组件状态

Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>"
```

---

### Task 6: Theme 主题数据

**Files:**
- Create: `ui/theme.go`
- Test: `ui/theme_test.go`

**Interfaces:**
- Consumes: 无（不得 import imgui——颜色用自有 `Color` 类型，Phase 2 的 android 文件负责换算成 `imgui.Vec4`）
- Produces: `Color struct{ R, G, B, A float32 }`、`Hex(s string) Color`（`#rrggbb` 或 `#rrggbbaa`，非法输入返回不透明黑）、`Theme` 结构体、`DefaultTheme() Theme`（QQ 蓝）

- [ ] **Step 1: 写失败测试**

`ui/theme_test.go`:

```go
package ui

import "testing"

func TestHexParsesRGBA(t *testing.T) {
	c := Hex("#2f8fd0ff")
	want := Color{R: 0x2f / 255.0, G: 0x8f / 255.0, B: 0xd0 / 255.0, A: 1}
	if c != want {
		t.Fatalf("Hex=#%+v want %+v", c, want)
	}
}

func TestHexDefaultsAlphaAndRejectsBadInput(t *testing.T) {
	if c := Hex("#ffffff"); c.A != 1 || c.R != 1 {
		t.Fatalf("6-digit hex should default alpha=1: %+v", c)
	}
	for _, bad := range []string{"", "#12345", "zzzzzz", "#1234567890"} {
		if c := Hex(bad); c != (Color{0, 0, 0, 1}) {
			t.Fatalf("Hex(%q)=%+v want opaque black", bad, c)
		}
	}
}

func TestDefaultThemeIsQQBlue(t *testing.T) {
	th := DefaultTheme()
	if th.Accent != Hex("#2f8fd0ff") {
		t.Fatalf("accent=%+v", th.Accent)
	}
	if th.TitleTop != Hex("#5aa9e6ff") || th.TitleBottom != Hex("#2f7fc4ff") {
		t.Fatal("title gradient colors")
	}
	if th.RailBg != Hex("#cfe4f7ff") {
		t.Fatal("rail bg")
	}
	if th.White != (Color{1, 1, 1, 1}) {
		t.Fatal("white")
	}
	if th.Rounding != 4 {
		t.Fatalf("rounding=%v want 4", th.Rounding)
	}
	// 主题可作零值比较（ShellOptions 缺省判断依赖这一点）
	if th == (Theme{}) {
		t.Fatal("default theme must differ from zero value")
	}
}
```

- [ ] **Step 2: 跑测试确认失败**

Run: `go test ./ui/ -run "TestHex|TestDefaultTheme" -v`
Expected: FAIL（undefined: Hex / DefaultTheme）

- [ ] **Step 3: 写实现**

`ui/theme.go`（色值取自 `internal/ui/theme_android.go`，原样保留 QQ 蓝）:

```go
package ui

import (
	"strconv"
	"strings"
)

// Color RGBA 颜色，分量 0..1。框架自有类型，避免纯逻辑层依赖 imgui；
// android 绘制层负责换算为 imgui.Vec4。
type Color struct {
	R, G, B, A float32
}

// Hex 解析 "#rrggbb" 或 "#rrggbbaa"；非法输入返回不透明黑。
func Hex(s string) Color {
	black := Color{0, 0, 0, 1}
	s = strings.TrimPrefix(s, "#")
	if len(s) != 6 && len(s) != 8 {
		return black
	}
	v, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return black
	}
	if len(s) == 6 {
		return Color{
			R: float32(v>>16&0xff) / 255,
			G: float32(v>>8&0xff) / 255,
			B: float32(v&0xff) / 255,
			A: 1,
		}
	}
	return Color{
		R: float32(v>>24&0xff) / 255,
		G: float32(v>>16&0xff) / 255,
		B: float32(v>>8&0xff) / 255,
		A: float32(v&0xff) / 255,
	}
}

// Theme 框架视觉令牌（CONTEXT.md「主题」）：整套替换或覆盖个别令牌，
// 无运行时切换。全部为可比较字段，零值表示「未指定」。
type Theme struct {
	WindowBg, ChildBg, PopupBg, Border     Color
	FrameBg, FrameHover, FrameActive       Color
	Button, ButtonHover, ButtonActive      Color
	Header, HeaderHover, HeaderActive      Color
	Text, TextDisabled, Accent             Color
	TitleBg, TitleBgActive                 Color
	TitleTop, TitleBottom, RailBg, White   Color
	Rounding                               float32
}

// DefaultTheme QQ 风浅蓝默认主题（沿用 internal/ui 的 QQ 蓝色值）。
func DefaultTheme() Theme {
	return Theme{
		WindowBg:      Hex("#e9f2fbff"),
		ChildBg:       Hex("#f7fbffff"),
		PopupBg:       Hex("#f2f8feff"),
		Border:        Hex("#9cc3e5ff"),
		FrameBg:       Hex("#ffffffff"),
		FrameHover:    Hex("#e3f0fbff"),
		FrameActive:   Hex("#cde6faff"),
		Button:        Hex("#dcebfaff"),
		ButtonHover:   Hex("#bcdcf7ff"),
		ButtonActive:  Hex("#8fc3efff"),
		Header:        Hex("#dcebfaff"),
		HeaderHover:   Hex("#bcdcf7ff"),
		HeaderActive:  Hex("#8fc3efff"),
		Text:          Hex("#1f3a52ff"),
		TextDisabled:  Hex("#7a8fa3ff"),
		Accent:        Hex("#2f8fd0ff"),
		TitleBg:       Hex("#3d8fd1ff"),
		TitleBgActive: Hex("#3d8fd1ff"),
		TitleTop:      Hex("#5aa9e6ff"),
		TitleBottom:   Hex("#2f7fc4ff"),
		RailBg:        Hex("#cfe4f7ff"),
		White:         Color{1, 1, 1, 1},
		Rounding:      4,
	}
}
```

- [ ] **Step 4: 跑测试确认通过**

Run: `go test ./ui/ -v`
Expected: PASS

- [ ] **Step 5: 提交**

```bash
git add ui/theme.go ui/theme_test.go
git commit -m "feat(ui): 主题令牌数据与 QQ 蓝默认主题

Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>"
```

---

### Task 7: 基准分辨率与缩放

**Files:**
- Create: `ui/scale.go`
- Test: `ui/scale_test.go`

**Interfaces:**
- Consumes: 无
- Produces: `DefaultBaseWidth = 1600`、`DefaultBaseHeight = 900`、`ComputeScale(displayW, displayH, baseW, baseH int) float64`（宽度驱动：`displayW/baseW`；`displayW<=0` 或 `baseW<=0` 返回 1；displayH/baseH 为签名预留，当前不参与计算）

- [ ] **Step 1: 写失败测试**

`ui/scale_test.go`:

```go
package ui

import "testing"

func TestComputeScaleWidthDriven(t *testing.T) {
	cases := []struct {
		dw, dh, bw, bh int
		want           float64
	}{
		{1600, 900, 1600, 900, 1.0},
		{2400, 1080, 1600, 900, 1.5},
		{800, 450, 1600, 900, 0.5},
		{0, 0, 1600, 900, 1.0},   // 无效显示尺寸回退 1
		{1600, 900, 0, 0, 1.0},   // 无效基准回退 1
		{-100, 900, 1600, 900, 1.0},
	}
	for _, c := range cases {
		if got := ComputeScale(c.dw, c.dh, c.bw, c.bh); got != c.want {
			t.Errorf("ComputeScale(%d,%d,%d,%d)=%v want %v", c.dw, c.dh, c.bw, c.bh, got, c.want)
		}
	}
}
```

- [ ] **Step 2: 跑测试确认失败**

Run: `go test ./ui/ -run TestComputeScale -v`
Expected: FAIL（undefined: ComputeScale）

- [ ] **Step 3: 写实现**

`ui/scale.go`:

```go
package ui

// 基准分辨率（CONTEXT.md）：UI 布局常量与游戏特征常量同写于 1600×900。
const (
	DefaultBaseWidth  = 1600
	DefaultBaseHeight = 900
)

// ComputeScale 计算 UI 布局缩放系数：宽度驱动 displayW/baseW，
// 启动时算一次，帧内经 Ctx.S 统一换算。无效输入回退 1。
// displayH/baseH 为策略调整预留，当前不参与计算。
func ComputeScale(displayW, displayH, baseW, baseH int) float64 {
	if displayW <= 0 || baseW <= 0 {
		return 1
	}
	return float64(displayW) / float64(baseW)
}
```

- [ ] **Step 4: 跑测试确认通过**

Run: `go test ./ui/ -v`
Expected: PASS

- [ ] **Step 5: 提交**

```bash
git add ui/scale.go ui/scale_test.go
git commit -m "feat(ui): 基准分辨率与宽度驱动缩放

Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>"
```

---

### Task 8: ShellOptions、NavEntry 与 Shell 实例（自动暂停核心）

**Files:**
- Create: `ui/options.go`
- Create: `ui/shell.go`
- Test: `ui/shell_test.go`

**Interfaces:**
- Consumes: `Store`（T1）、`ScriptController`/`ScriptHooks`（T2）、`Task`/`SeedAll`/`ApplyAll`/`SeedPanelDefaults`（T4）、`Theme`/`DefaultTheme`（T6）、`DefaultBaseWidth/Height`（T7）
- Produces:
  - `StatusSource interface{ Text() string }`
  - `NavEntry struct{ ID, Title string; Render func(*Ctx) }`
  - `ShellOptions struct{ Title string; Tasks []Task; Nav []NavEntry; Store *Store; Controller *ScriptController; Status StatusSource; Theme Theme; ConfigPath, DataStorePath string; OpenPanelOnStart bool; BaseWidth, BaseHeight int }`
  - `Shell`、`NewShell(opts ShellOptions) *Shell`
  - 方法：`Store() *Store`、`Theme() Theme`、`Tasks() []Task`、`Nav() []NavEntry`、`Seed()`、`Apply()`、`OpenPanel()`、`ClosePanel()`、`PanelOpen() bool`、`ToggleMinimized()`、`Minimized() bool`、`AutoPaused() bool`、`PauseResume()`、`StartStop() error`、`StatusText() string`

- [ ] **Step 1: 写失败测试**

`ui/shell_test.go`:

```go
package ui

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// blockingHooks 返回一对阻塞 run 的 hooks：run 直到 stop 被调用才返回。
func blockingHooks() (ScriptHooks, chan struct{}) {
	stopCh := make(chan struct{})
	hooks := ScriptHooks{
		OnStart: func() (func() error, func(), func(), func()) {
			run := func() error {
				<-stopCh
				return nil
			}
			return run, func() {}, func() {}, func() {}
		},
		OnExit: func() {},
	}
	return hooks, stopCh
}

func waitState(t *testing.T, c *ScriptController, want ScriptState) {
	t.Helper()
	deadline := time.Now().Add(500 * time.Millisecond)
	for c.State() != want && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	if c.State() != want {
		t.Fatalf("state=%v want %v", c.State(), want)
	}
}

func TestShellDefaults(t *testing.T) {
	s := NewShell(ShellOptions{})
	if s.Store() == nil {
		t.Fatal("Store should default to NewStore")
	}
	if s.Theme() != DefaultTheme() {
		t.Fatal("zero Theme should default to DefaultTheme")
	}
	if s.PanelOpen() {
		t.Fatal("panel closed by default")
	}
	if s.Minimized() || s.AutoPaused() {
		t.Fatal("minimized/autoPaused default false")
	}
}

func TestShellOpenPanelAutoPausesAndCloseResumes(t *testing.T) {
	hooks, stopCh := blockingHooks()
	defer close(stopCh)
	ctrl := NewScriptController(hooks)
	s := NewShell(ShellOptions{Controller: ctrl})

	ctrl.Start()
	waitState(t, ctrl, StateRunning)

	s.OpenPanel()
	if !s.PanelOpen() || ctrl.State() != StatePaused || !s.AutoPaused() {
		t.Fatalf("open panel should auto-pause: open=%v state=%v auto=%v",
			s.PanelOpen(), ctrl.State(), s.AutoPaused())
	}

	s.ClosePanel()
	if s.PanelOpen() || ctrl.State() != StateRunning || s.AutoPaused() {
		t.Fatalf("close panel should resume: open=%v state=%v auto=%v",
			s.PanelOpen(), ctrl.State(), s.AutoPaused())
	}
}

func TestShellManualResumeOverridesAutoPause(t *testing.T) {
	hooks, stopCh := blockingHooks()
	defer close(stopCh)
	ctrl := NewScriptController(hooks)
	s := NewShell(ShellOptions{Controller: ctrl})

	ctrl.Start()
	waitState(t, ctrl, StateRunning)
	s.OpenPanel()

	s.PauseResume() // 手动恢复：清除 autoPaused
	if ctrl.State() != StateRunning || s.AutoPaused() {
		t.Fatalf("manual resume: state=%v auto=%v", ctrl.State(), s.AutoPaused())
	}
	s.ClosePanel() // 不得二次动作
	if ctrl.State() != StateRunning {
		t.Fatalf("close after manual resume must not touch controller: %v", ctrl.State())
	}
}

func TestShellStartWhilePanelOpenAutoPausesAndSaves(t *testing.T) {
	hooks, stopCh := blockingHooks()
	defer close(stopCh)
	ctrl := NewScriptController(hooks)
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "ui.json")

	s := NewShell(ShellOptions{Controller: ctrl, ConfigPath: cfgPath, OpenPanelOnStart: true})
	if !s.PanelOpen() {
		t.Fatal("OpenPanelOnStart should open panel")
	}
	s.Store().SetBool("k", true)

	if err := s.StartStop(); err != nil {
		t.Fatalf("StartStop: %v", err)
	}
	waitState(t, ctrl, StatePaused) // 启动后面板仍开 -> 自动暂停
	if !s.AutoPaused() {
		t.Fatal("start while panel open should auto-pause")
	}
	if _, err := os.Stat(cfgPath); err != nil {
		t.Fatalf("start should save config: %v", err)
	}

	if err := s.StartStop(); err != nil { // 停止
		t.Fatalf("StartStop stop: %v", err)
	}
}

func TestShellStatusTextOnlyWhenRunning(t *testing.T) {
	hooks, stopCh := blockingHooks()
	defer close(stopCh)
	ctrl := NewScriptController(hooks)
	s := NewShell(ShellOptions{Controller: ctrl, Status: fakeStatus{"战斗 12 场"}})
	if s.StatusText() != "" {
		t.Fatal("idle -> empty status")
	}
	ctrl.Start()
	waitState(t, ctrl, StateRunning)
	if s.StatusText() != "战斗 12 场" {
		t.Fatalf("status=%q", s.StatusText())
	}
	if NewShell(ShellOptions{Controller: ctrl}).StatusText() != "" {
		t.Fatal("nil StatusSource -> empty")
	}
}

type fakeStatus struct{ text string }

func (f fakeStatus) Text() string { return f.text }

func TestShellSeedAppliesTasksAndPanelDefaults(t *testing.T) {
	enabled := true
	s := NewShell(ShellOptions{
		Tasks: []Task{{
			ID: "arena", Title: "王国竞技场", Category: "日常", EnabledKey: "arena_enabled",
			Fields: []Field{Bool("arena_enabled", "启用", func() bool { return enabled }, func(v bool) { enabled = v })},
		}},
		Nav: []NavEntry{{ID: "tasks", Title: "任务"}, {ID: "system", Title: "系统"}},
	})
	s.Seed()
	if !s.Store().GetBool("arena_enabled") {
		t.Fatal("Seed should seed task fields")
	}
	if s.Store().GetString(KeyPanelNav) != "tasks" || s.Store().GetString(KeyPanelSelected) != "arena" {
		t.Fatal("Seed should seed panel defaults")
	}
	s.Store().SetBool("arena_enabled", false)
	s.Apply()
	if enabled {
		t.Fatal("Apply should write back to app config")
	}
}
```

- [ ] **Step 2: 跑测试确认失败**

Run: `go test ./ui/ -run TestShell -v`
Expected: FAIL（undefined: NewShell / ShellOptions）

- [ ] **Step 3: 写实现**

`ui/options.go`:

```go
package ui

// StatusSource 任务状态文本来源（ADR-0002 窄接口）：应用的 status.Reporter
// 天然实现它；框架不拥有该机制，游戏代码因此无需 import 框架。
type StatusSource interface {
	Text() string
}

// NavEntry 面板左栏导航条目（ADR-0002 导航全描述符化）。Render 为绘制层
// 函数（Phase 2 起使用）；任务列表页是框架提供的可复用组件，应用把它作为
// 一个条目挂载。
type NavEntry struct {
	ID     string
	Title  string
	Render func(*Ctx)
}

// ShellOptions Shell 的全部外部输入：描述符、持久化路径、生命周期钩子。
// 零值字段取默认：Store→NewStore、Theme→DefaultTheme、BaseWidth/Height→1600×900。
type ShellOptions struct {
	Title      string
	Tasks      []Task
	Nav        []NavEntry
	Store      *Store
	Controller *ScriptController
	// Status 任务状态来源；非 nil 且脚本运行中时，灵动岛展示该文本。
	Status StatusSource
	Theme  Theme
	// ConfigPath UI 配置持久化路径（启动脚本时落盘）；空则不落盘。
	ConfigPath string
	// DataStorePath 业务 KV 路径；清除缓存时一并删除。
	DataStorePath    string
	OpenPanelOnStart bool
	BaseWidth        int
	BaseHeight       int
}
```

`ui/shell.go`:

```go
package ui

// Shell 框架运行时实例（CONTEXT.md 术语见 ADR-0002）：持有全部 UI 状态
// （面板可见性、最小化、自动暂停标记），无包级可变状态；桌面测试可起多个
// 实例互不干扰。绘制层（Phase 2）以它为状态后端。
type Shell struct {
	opts ShellOptions

	store *Store
	ctrl  *ScriptController

	panelOpen  bool
	minimized  bool
	autoPaused bool
}

// NewShell 构造 Shell 并归一默认值。OpenPanelOnStart 时面板初始打开
// （此时脚本必为空闲，不触发自动暂停）。
func NewShell(opts ShellOptions) *Shell {
	if opts.Store == nil {
		opts.Store = NewStore()
	}
	if opts.Controller == nil {
		opts.Controller = NewScriptController(ScriptHooks{})
	}
	if opts.Theme == (Theme{}) {
		opts.Theme = DefaultTheme()
	}
	if opts.BaseWidth <= 0 {
		opts.BaseWidth = DefaultBaseWidth
	}
	if opts.BaseHeight <= 0 {
		opts.BaseHeight = DefaultBaseHeight
	}
	return &Shell{
		opts:      opts,
		store:     opts.Store,
		ctrl:      opts.Controller,
		panelOpen: opts.OpenPanelOnStart,
	}
}

func (s *Shell) Store() *Store      { return s.store }
func (s *Shell) Theme() Theme       { return s.opts.Theme }
func (s *Shell) Tasks() []Task      { return s.opts.Tasks }
func (s *Shell) Nav() []NavEntry    { return s.opts.Nav }
func (s *Shell) PanelOpen() bool    { return s.panelOpen }
func (s *Shell) Minimized() bool    { return s.minimized }
func (s *Shell) AutoPaused() bool   { return s.autoPaused }

func (s *Shell) ToggleMinimized() { s.minimized = !s.minimized }

// Seed 用任务字段默认值与面板偏好默认值填充 Store（仅填缺失键）。
// 应用通常在首帧（LoadConfig 之后）调用。
func (s *Shell) Seed() {
	SeedAll(s.store, s.opts.Tasks)
	navIDs := make([]string, 0, len(s.opts.Nav))
	for _, n := range s.opts.Nav {
		navIDs = append(navIDs, n.ID)
	}
	SeedPanelDefaults(s.store, s.opts.Tasks, navIDs)
}

// Apply 把 Store 中的配置写回应用配置。应用在 ScriptHooks.OnStart 内调用。
func (s *Shell) Apply() {
	ApplyAll(s.store, s.opts.Tasks)
}

// OpenPanel 打开配置面板；脚本运行中时自动暂停（遮挡策略：面板遮挡画面
// 期间不识别）。
func (s *Shell) OpenPanel() {
	s.panelOpen = true
	if s.ctrl.State() == StateRunning {
		s.ctrl.Pause()
		s.autoPaused = true
	}
}

// ClosePanel 关闭配置面板；若为自动暂停则恢复运行。手动暂停/恢复后
// （PauseResume 已清除 autoPaused）关闭面板不触碰控制器。
func (s *Shell) ClosePanel() {
	s.panelOpen = false
	if s.autoPaused {
		s.ctrl.Resume()
		s.autoPaused = false
	}
}

// PauseResume 手动暂停/继续；清除自动暂停标记（手动操作优先于遮挡策略）。
func (s *Shell) PauseResume() {
	switch s.ctrl.State() {
	case StateRunning:
		s.ctrl.Pause()
	case StatePaused:
		s.ctrl.Resume()
	}
	s.autoPaused = false
}

// StartStop 启动或停止脚本。启动时先把配置落盘（ConfigPath 非空），
// 若面板仍打开则启动后立即自动暂停。
func (s *Shell) StartStop() error {
	if s.ctrl.State() == StateIdle {
		if s.opts.ConfigPath != "" {
			if err := s.store.SaveConfig(s.opts.ConfigPath); err != nil {
				return err
			}
		}
		s.ctrl.Start()
		if s.panelOpen && s.ctrl.State() == StateRunning {
			s.ctrl.Pause()
			s.autoPaused = true
		}
		return nil
	}
	s.ctrl.Stop()
	return nil
}

// StatusText 运行中且配置了状态来源时返回任务状态文本，否则空串
// （UI 回退默认文案）。
func (s *Shell) StatusText() string {
	if s.opts.Status == nil || s.ctrl.State() != StateRunning {
		return ""
	}
	return s.opts.Status.Text()
}
```

- [ ] **Step 4: 跑测试确认通过**

Run: `go test ./ui/ -v`
Expected: PASS（全部）

- [ ] **Step 5: 提交**

```bash
git add ui/options.go ui/shell.go ui/shell_test.go
git commit -m "feat(ui): Shell 实例与遮挡自动暂停核心

Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>"
```

---

### Task 9: 组件 Props 结构体

**Files:**
- Create: `ui/props.go`
- Test: `ui/props_test.go`

**Interfaces:**
- Consumes: `Field`（T3）
- Produces（纯数据，渲染在 Phase 2 的 android 组件函数中）:
  - `ButtonKind`（`ButtonPrimary`/`ButtonSecondary`/`ButtonDanger`）
  - `ButtonProps{ Label string; Kind ButtonKind; Width, Height float64; Disabled bool; OnClick func() }`
  - `CheckboxProps{ Label string; Checked bool; OnChange func(bool) }`
  - `NumberInputProps{ Label string; Value, Min, Max, Step, Width float64; OnChange func(float64) }`，方法 `Clamp(v float64) float64`（`Max > Min` 时钳入 [Min,Max]；`Step <= 0` 视为 1——Clamp 不含步进吸附）
  - `InputProps{ Label, Hint, Value string; Width float64; OnChange func(string) }`
  - `DropdownProps{ Label string; Options []string; Selected int; OnChange func(int) }`
  - `TabsProps{ Items []string; Selected int; OnChange func(int) }`
  - `CollapsibleProps{ Label string; Open bool; OnToggle func(bool) }`
  - `FormProps{ Store *Store; Fields []Field }`
  - （`ImageProps` 依赖 imgui 纹理类型，留待 Phase 2 在 android 文件中定义）

- [ ] **Step 1: 写失败测试**

`ui/props_test.go`:

```go
package ui

import "testing"

func TestNumberInputClamp(t *testing.T) {
	p := NumberInputProps{Min: 0, Max: 99, Step: 1}
	if got := p.Clamp(120); got != 99 {
		t.Fatalf("Clamp(120)=%v want 99", got)
	}
	if got := p.Clamp(-5); got != 0 {
		t.Fatalf("Clamp(-5)=%v want 0", got)
	}
	if got := p.Clamp(42); got != 42 {
		t.Fatalf("Clamp(42)=%v want 42", got)
	}

	unbounded := NumberInputProps{Min: 0, Max: 0} // Max <= Min 表示不钳上界
	if got := unbounded.Clamp(12345); got != 12345 {
		t.Fatalf("unbounded Clamp=%v want 12345", got)
	}
	if got := unbounded.Clamp(-3); got != -3 {
		t.Fatalf("unbounded Clamp(-3)=%v want -3", got)
	}
}

func TestFormPropsCarriesFields(t *testing.T) {
	backing := false
	fp := FormProps{
		Store:  NewStore(),
		Fields: []Field{Bool("k", "开关", func() bool { return backing }, func(v bool) { backing = v })},
	}
	if len(fp.Fields) != 1 || fp.Fields[0].Key() != "k" {
		t.Fatalf("FormProps fields: %+v", fp.Fields)
	}
}
```

- [ ] **Step 2: 跑测试确认失败**

Run: `go test ./ui/ -run "TestNumberInputClamp|TestFormProps" -v`
Expected: FAIL（undefined: NumberInputProps / FormProps）

- [ ] **Step 3: 写实现**

`ui/props.go`:

```go
package ui

// 组件 Props（ADR-0003）：每组件一个结构体，唯一输入，含数据与回调。
// 本文件仅为纯数据定义；渲染函数在 Phase 2 的 android 组件中。

type ButtonKind int

const (
	ButtonPrimary ButtonKind = iota
	ButtonSecondary
	ButtonDanger
)

type ButtonProps struct {
	Label           string
	Kind            ButtonKind
	Width, Height   float64 // 基准分辨率尺寸；0 表示按内容自适应
	Disabled        bool
	OnClick         func()
}

type CheckboxProps struct {
	Label   string
	Checked bool
	OnChange func(bool)
}

type NumberInputProps struct {
	Label         string
	Value         float64
	Min, Max, Step float64 // Max <= Min 表示不钳上界
	Width         float64
	OnChange      func(float64)
}

// Clamp 把 v 钳入 [Min, Max]（Max > Min 时）；不含步进吸附。
func (p NumberInputProps) Clamp(v float64) float64 {
	if p.Max > p.Min {
		if v < p.Min {
			return p.Min
		}
		if v > p.Max {
			return p.Max
		}
	}
	return v
}

type InputProps struct {
	Label, Hint, Value string
	Width              float64
	OnChange           func(string)
}

type DropdownProps struct {
	Label    string
	Options  []string
	Selected int
	OnChange func(int)
}

type TabsProps struct {
	Items    []string
	Selected int
	OnChange func(int)
}

type CollapsibleProps struct {
	Label    string
	Open     bool
	OnToggle func(bool)
}

// FormProps Form 组件输入：按 Fields 自动排版，值直连 Store 读写。
type FormProps struct {
	Store  *Store
	Fields []Field
}
```

- [ ] **Step 4: 跑测试确认通过**

Run: `go test ./ui/ -v`
Expected: PASS（全部）

- [ ] **Step 5: 提交**

```bash
git add ui/props.go ui/props_test.go
git commit -m "feat(ui): 组件 Props 结构体（按钮/输入/下拉/标签页/表单）

Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>"
```

---

### Task 10: 全量验证与文档同步

**Files:**
- Modify: `CLAUDE.md`（Project Structure 一节）
- Modify: `AGENTS.md`（与 CLAUDE.md 同处，保持同步）

- [ ] **Step 1: 格式化与静态检查**

Run: `gofmt -w ui/ && go vet ./ui/...`
Expected: 无输出

- [ ] **Step 2: 全量构建与测试（确认 internal/ui 与既有包零影响）**

Run: `go build ./... && go test ./...`
Expected: 全部 PASS

- [ ] **Step 3: 文档同步**

`CLAUDE.md` 的 Project Structure 代码块中，`├── internal/` 之前插入一行：

```
├── ui/                        # 脚本 UI 框架（ADR-0002/0003）：纯逻辑无标签可测，android 绘制薄层（Phase 1 进行中）
```

`AGENTS.md` 同处做相同修改（两文件保持同步）。

- [ ] **Step 4: 提交**

```bash
git add CLAUDE.md AGENTS.md ui/
git commit -m "docs: CLAUDE/AGENTS 增补 ui 框架包说明

Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>"
```

---

## Self-Review 记录

- **Spec 覆盖**：ADR-0002 的 Phase 1 部分（描述符模型、字段绑定、Shell 实例化、主题数据、缩放、窄接口 StatusSource、纯逻辑可测）→ Task 1–9 全覆盖；ADR-0003 的纯数据部分（Ctx/State、Props）→ Task 5/9。绘制层、岛/面板 android 实现、应用接线属 Phase 2/3，已显式排除。
- **执行顺序注意**：Task 4 的 `Task.RenderDetail func(*Ctx)` 依赖 Task 5 的 `Ctx`——按 1→2→3→5→4→6→7→8→9 顺序执行。
- **类型一致性**：`Field` 接口方法集（Key/Label/Widget/Constraints/Seed/Apply）在 Task 3/4/9 间一致；`SeedPanelDefaults(store, tasks, navIDs)` 三参签名在 Task 4 定义、Task 8 消费，一致。
