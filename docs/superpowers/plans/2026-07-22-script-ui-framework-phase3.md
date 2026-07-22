# 脚本 UI 框架 Phase 3（应用侧重接与 internal/ui 退役）实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 把应用（main.go）从旧 `internal/ui` 切到新框架 `ui`（ADR-0002/0003，Phase 1/2 已交付），用字段描述符替代 binding.go 的四点同步，删除整个 `internal/ui` 包。

**Architecture:** 纯接线与删除，无新框架能力。main.go 是组合根：构造 arena 的 `ui.Task` 描述符（Fields 的 get/set 闭包直连 `config.Config`）、Nav（框架提供的 `TaskListPage`/`SystemPage`）、Controller hooks（OnStart 内 `ui.ApplyAll` 回写后 `buildRuntime`）。持久化格式、设备路径、store 键全部保持不变——设备上已有的 ui.json/store.json 无缝兼容。

**Tech Stack:** Go 1.25，无新依赖。

**分支：** `feat/script-ui-framework-phase3`（从 main 切出）。

---

## 现状核查（已调研，计划据此编写）

- `main.go` 是当前唯一 import `app/internal/ui` 的文件（已 grep 确认）。
- 旧 store 键（binding.go）：`arena_enabled` / `arena_max_battles` / `arena_auto_buy_count` / `arena_trophy_diff`；面板键 `panel_nav`（`tasks|system`）/ `panel_cat` / `panel_selected`。**全部原样保留**。
- 旧设备路径（internal/ui/options.go）：`/sdcard/shuaibin-cookie/ui.json`、`/sdcard/shuaibin-cookie/store.json`。新框架不再持有路径常量 → 由 main.go 接手。
- `ui/store.go` 与 `internal/ui/store.go` 序列化逐字节一致（仅一行注释差）——格式兼容已确认。
- `config.Arena.MaxBattles` 是 `*int`（nil = 不限）；UI 层 0 = 不限（旧 ApplyToConfig 的映射）。
- `ui.Task.Category` 是自由字符串、直接作为 chip 展示文本（`categoryLabel` 只兼容旧 id）。新描述符直接用 `"日常"`。
- `ui/controller.go:6` import `app/internal/logger`——框架反向依赖应用，违反 ADR-0002 依赖方向，Phase 3 顺手切除。
- `status.Reporter` 的 `Text() string` 天然实现 `ui.StatusSource`，无需改动。
- 旧描述符能力覆盖核对：arena 详情 = 启用 checkbox + 三个数字框 + 摘要行「上限 X · 购买 Y · 奖杯差 Z」——新 `Form` 自动渲染 + `Task.Summary` 全覆盖，无需 `RenderDetail` 逃生门。旧数字框的 hint（"0=不限" 等）在字段描述符中无对应——v1 接受损失，记录在案。
- main.go 现行流程中 `ui.SeedFromConfig(uiStore, cfg)` 的启动前播种由框架首帧 `LoadConfig → shell.Seed()` 取代（Field.Seed 仅填缺失键，文件值优先，语义等价）。

---

### Task 1: 框架日志钩子（切除 ui → internal/logger 反向依赖）

**Files:**
- Modify: `ui/controller.go`
- Test: `ui/controller_test.go`（追加一个用例）

**Interfaces:**
- Produces: `ui.LogErrorf`（包级钩子变量）

- [ ] **Step 1: 写失败测试**

`ui/controller_test.go` 追加：

```go
func TestScriptControllerRunErrorGoesToLogHook(t *testing.T) {
	old := LogErrorf
	defer func() { LogErrorf = old }()
	var got string
	LogErrorf = func(format string, args ...any) { got = fmt.Sprintf(format, args...) }

	ctrl := NewScriptController(ScriptHooks{
		OnStart: func() (func() error, func(), func(), func()) {
			return func() error { return errors.New("boom") }, func() {}, func() {}, func() {}
		},
	})
	ctrl.Start()
	deadline := time.Now().Add(500 * time.Millisecond)
	for ctrl.State() != StateIdle && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	if !strings.Contains(got, "boom") {
		t.Fatalf("LogErrorf got %q", got)
	}
}
```

Run: `go test ./ui/ -run TestScriptControllerRunError -v`
Expected: FAIL（undefined: LogErrorf）

- [ ] **Step 2: 改 ui/controller.go**

删除 `app/internal/logger` import；`Start` 的 run goroutine 错误处理改为：

```go
// LogErrorf 框架内部错误日志钩子（如脚本 run 返回错误）。默认丢弃；
// 应用可替换为自身日志器（如 main 中 ui.LogErrorf = logger.Errorf）。
// 包级变量非并发安全，应在启动早期赋值一次。
var LogErrorf = func(format string, args ...any) {}
```

`if err != nil { LogErrorf("script run error: %v", err) }`。

- [ ] **Step 3: 验证**

Run: `gofmt -l ui/`、`go vet ./ui/...`、`go build ./...`、`go test ./...`、`GOOS=android GOARCH=arm64 CGO_ENABLED=0 go build ./...`
Expected: 全部通过；`grep -rn "internal/" ui/` 无结果（框架零应用依赖，ADR-0002 收口）

- [ ] **Step 4: 提交**

```bash
git add ui/controller.go ui/controller_test.go
git commit -m "refactor(ui): 日志经 LogErrorf 钩子，切除框架对 internal/logger 依赖

Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>"
```

---

### Task 2: arena UI 描述符（tagless，TDD）

**Files:**
- Create: `ui_arena.go`（package main）
- Test: `ui_arena_test.go`（package main）

**Interfaces:**
- Consumes: `ui.Task`、`ui.Bool/Number`、`ui.SeedAll/ApplyAll`、`config.Config`
- Produces: `arenaTaskDescriptor(cfg *config.Config) ui.Task`

- [ ] **Step 1: 写失败测试**

`ui_arena_test.go`：

```go
package main

import (
	"testing"

	"app/internal/config"
	"app/ui"
)

func TestArenaDescriptorSeedApplyRoundTrip(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Tasks.Arena.Enabled = true
	cfg.Tasks.Arena.AutoBuyCount = 2
	cfg.Tasks.Arena.TrophyDiff = 50
	maxB := 8
	cfg.Tasks.Arena.MaxBattles = &maxB

	task := arenaTaskDescriptor(cfg)
	tasks := []ui.Task{task}

	s := ui.NewStore()
	ui.SeedAll(s, tasks)
	if !s.GetBool("arena_enabled") || s.GetFloat("arena_max_battles") != 8 ||
		s.GetFloat("arena_auto_buy_count") != 2 || s.GetFloat("arena_trophy_diff") != 50 {
		t.Fatalf("seed: %v %v %v %v", s.GetBool("arena_enabled"),
			s.GetFloat("arena_max_battles"), s.GetFloat("arena_auto_buy_count"),
			s.GetFloat("arena_trophy_diff"))
	}

	// 面板改值 → Apply 回写 cfg；上限 0 → MaxBattles 归 nil（不限）
	s.SetBool("arena_enabled", false)
	s.SetFloat("arena_max_battles", 0)
	s.SetFloat("arena_auto_buy_count", 3)
	ui.ApplyAll(s, tasks)
	a := cfg.Tasks.Arena
	if a.Enabled || a.MaxBattles != nil || a.AutoBuyCount != 3 || a.TrophyDiff != 50 {
		t.Fatalf("apply: %+v", a)
	}
}

func TestArenaDescriptorSummary(t *testing.T) {
	cfg := config.DefaultConfig()
	task := arenaTaskDescriptor(cfg)
	s := ui.NewStore()
	s.SetFloat("arena_max_battles", 10)
	s.SetFloat("arena_auto_buy_count", 1)
	s.SetFloat("arena_trophy_diff", 30)
	if got := task.Summary(s); got != "上限 10 · 购买 1 · 奖杯差 30" {
		t.Fatalf("summary=%q", got)
	}
	s.SetFloat("arena_max_battles", 0)
	if got := task.Summary(s); got != "上限 不限 · 购买 1 · 奖杯差 30" {
		t.Fatalf("summary unlimited=%q", got)
	}
}
```

Run: `go test . -run TestArenaDescriptor -v`
Expected: FAIL（undefined: arenaTaskDescriptor）

- [ ] **Step 2: 写 ui_arena.go**

```go
package main

import (
	"fmt"

	"app/internal/config"
	"app/ui"
)

// arenaTaskDescriptor 王国竞技场的面板描述符（ADR-0002）：字段键与旧
// internal/ui/binding.go 完全一致，设备上已有的 ui.json 无缝兼容。
// 详情页由框架 Form 按 Fields 自动渲染（无 RenderDetail 逃生门需求）。
func arenaTaskDescriptor(cfg *config.Config) ui.Task {
	a := &cfg.Tasks.Arena
	return ui.Task{
		ID:         "arena",
		Title:      "王国竞技场",
		Category:   "日常",
		EnabledKey: "arena_enabled",
		Fields: []ui.Field{
			ui.Bool("arena_enabled", "启用",
				func() bool { return a.Enabled },
				func(v bool) { a.Enabled = v }),
			ui.Number("arena_max_battles", "每日战斗上限", 0, 999, 1,
				func() int {
					if a.MaxBattles == nil {
						return 0 // 0 = 不限
					}
					return *a.MaxBattles
				},
				func(v int) {
					if v > 0 {
						a.MaxBattles = &v
					} else {
						a.MaxBattles = nil
					}
				}),
			ui.Number("arena_auto_buy_count", "自动购买次数", 0, 999, 1,
				func() int { return a.AutoBuyCount },
				func(v int) { a.AutoBuyCount = v }),
			ui.Number("arena_trophy_diff", "奖杯差阈值", 0, 999, 1,
				func() int { return a.TrophyDiff },
				func(v int) { a.TrophyDiff = v }),
		},
		Summary: func(s *ui.Store) string {
			max := int(s.GetFloat("arena_max_battles"))
			maxLabel := "不限"
			if max > 0 {
				maxLabel = fmt.Sprintf("%d", max)
			}
			return fmt.Sprintf("上限 %s · 购买 %d · 奖杯差 %d",
				maxLabel, int(s.GetFloat("arena_auto_buy_count")), int(s.GetFloat("arena_trophy_diff")))
		},
	}
}
```

注意：`ui.Bool`/`ui.Number` 的实际签名以 `ui/field.go` 为准（实现前先读；`Number(key, label string, min, max, step float64, get func() int, set func(int))`）。闭包写法 `func(v int) { a.MaxBattles = &v }` 中 v 为参数、每调用独立，取址安全。

- [ ] **Step 3: 跑测试确认通过**

Run: `go test . -run TestArenaDescriptor -v`
Expected: PASS

- [ ] **Step 4: 提交**

```bash
git add ui_arena.go ui_arena_test.go
git commit -m "feat(main): arena 面板描述符（字段键与旧 binding 一致）

Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>"
```

---

### Task 3: main.go 重接 ui.RunShell

**Files:**
- Modify: `main.go`
- Create: `ui/taskpage_stub.go`、`ui/system_stub.go`（`//go:build !android || !cgo`）

> **计划修正（执行时发现）**：`ui.TaskListPage()`/`ui.SystemPage()` 仅存在于 `android && cgo` 文件；无构建标签的 main.go 引用它们会导致桌面编译失败。需配套两个桌面符号桩（与 `ui/run_stub.go` 同标签拆分约定，返回空渲染函数）。

**Interfaces:**
- Consumes: Task 1/2 全部产出、`ui.RunShell`、`ui.TaskListPage()`、`ui.SystemPage()`、`ui.NavEntry`、`ui.ScriptHooks`
- Produces: 应用唯一入口的新接线

- [ ] **Step 1: 重写 main.go**

完整替换为：

```go
package main

import (
	"os"
	"time"

	"app/internal/config"
	"app/internal/game/arena"
	"app/internal/game/common/kingdom"
	"app/internal/guard"
	"app/internal/logger"
	"app/internal/platform/action"
	"app/internal/platform/screen"
	"app/internal/runtime"
	"app/internal/scheduler"
	"app/internal/status"
	"app/internal/store"
	"app/ui"
)

// 设备上的持久化路径（原 internal/ui/options.go，全项目唯一来源）。
const (
	defaultDataDir    = "/sdcard/shuaibin-cookie"
	defaultConfigPath = defaultDataDir + "/ui.json"
	defaultStorePath  = defaultDataDir + "/store.json"
)

func main() {
	logger.SetLevel(logger.LevelInfo)
	logger.Infof("shuaibin cookie run kingdom start...")

	cfg, err := config.LoadConfig("config.json")
	if err != nil {
		logger.Errorf("failed to load config: %v", err)
		return
	}

	ui.LogErrorf = logger.Errorf
	uiStore := ui.NewStore()
	statusReporter := status.New()
	tasks := []ui.Task{arenaTaskDescriptor(cfg)}

	ctrl := ui.NewScriptController(ui.ScriptHooks{
		OnStart: func() (run func() error, pause, resume, stop func()) {
			ui.ApplyAll(uiStore, tasks) // 面板值回写 cfg 后重建运行时
			rt := buildRuntime(cfg, statusReporter)
			return rt.Run, rt.Pause, rt.Resume, rt.Stop
		},
		OnExit: func() {
			time.Sleep(500 * time.Millisecond)
			os.Exit(0)
		},
	})

	ui.RunShell(ui.ShellOptions{
		Title:       "帅宾 Cookie",
		Tasks:       tasks,
		Nav: []ui.NavEntry{
			{ID: "tasks", Title: "任务", Render: ui.TaskListPage()},
			{ID: "system", Title: "系统", Render: ui.SystemPage()},
		},
		Store:            uiStore,
		Controller:       ctrl,
		Status:           statusReporter,
		ConfigPath:       defaultConfigPath,
		DataStorePath:    defaultStorePath,
		OpenPanelOnStart: true,
	})
}

func buildRuntime(cfg *config.Config, statusReporter *status.Reporter) *runtime.Runtime {
	det := screen.NewDetector(0)
	exec := action.NewExecutor(0)
	s := store.New(defaultStorePath)
	g := guard.New(det)
	sched := scheduler.New()

	kingdomFeature := kingdom.DefaultFeature()
	kingdomPage := kingdom.NewPage(det, exec, kingdomFeature)
	arenaFeature := arena.DefaultFeature()
	arenaState := arena.NewState(s)
	arenaTask := arena.NewTask(
		&cfg.Tasks.Arena,
		det, exec, arenaFeature, kingdomPage, arenaState, g,
	)

	sched.Build(scheduler.TaskOpts{
		Name: "王国竞技场",
		CheckEnabled: func() bool {
			return cfg.Tasks.Arena.Enabled
		},
		CheckReady: func() (bool, time.Duration) {
			if arenaState.IsReachMaxBattles(&cfg.Tasks.Arena) {
				return false, 0
			}
			remain := arenaState.TimeUntilRefresh()
			if remain > 0 {
				return false, remain
			}
			return true, 0
		},
		WaitHUD: func(remain time.Duration) string {
			return "免费刷新等待"
		},
		Action: arenaTask.Run,
	})

	rt := runtime.New(runtime.Options{
		Scheduler: sched,
		Guard:     g,
		RuntimeCfg: runtime.RuntimeConfig{
			GuardInterval: 500 * time.Millisecond,
			IdleDelay:     30 * time.Second,
			StepDelay:     5 * time.Second,
			StopOnError:   false,
		},
	})
	arenaTask.SetShouldStop(rt.IsStopped)
	arenaTask.SetStatusReporter(statusReporter)
	return rt
}
```

变化点（相对旧 main.go，逐项核对）：
1. import `app/ui` 替换 `app/internal/ui`。
2. 路径常量本地化（值不变）；`buildRuntime` 里 `store.New(ui.DefaultStorePath)` → `store.New(defaultStorePath)`。
3. 删启动前 `ui.SeedFromConfig`——框架首帧 `LoadConfig → shell.Seed()` 取代（Field.Seed 仅填缺失键）。
4. OnStart 内 `ui.ApplyToConfig(uiStore, cfg)` → `ui.ApplyAll(uiStore, tasks)`（描述符回写）。
5. `Reseed` 选项消失——SystemPage 清缓存后调 `shell.Seed()`（Field.Seed 等价物）。
6. Nav 描述符化挂载框架页；nav id `tasks`/`system` 与旧值一致（ui.json 的 panel_nav 兼容）。
7. `ui.LogErrorf = logger.Errorf`（Task 1 钩子接线）。
8. `Status: statusReporter`——`*status.Reporter` 实现 `ui.StatusSource`，直传。

- [ ] **Step 2: 验证**

Run: `gofmt -l . | grep -v AutoGo`（仅确认 main.go/ui_arena.go 不在列）、`go vet ./...`、`go build ./...`、`go test ./...`、`GOOS=android GOARCH=arm64 CGO_ENABLED=0 go build ./...`
Expected: 全部通过（桌面走 `ui/run_stub.go`：Seed 后直接 Start——与旧 stub 行为一致，仅编译验证，不运行）

- [ ] **Step 3: 提交**

```bash
git add main.go
git commit -m "refactor(main): 切换到 ui 框架 RunShell（描述符接线）

Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>"
```

---

### Task 4: 删除 internal/ui

**Files:**
- Delete: `internal/ui/`（全部 16 个文件）

- [ ] **Step 1: 确认零引用后删除**

```bash
grep -rln "app/internal/ui" --include="*.go" . | grep -v AutoGo   # 预期：无输出
git rm -r internal/ui
```

- [ ] **Step 2: 验证（本计划的收口门禁）**

Run: `go vet ./...`、`go build ./...`、`go test ./... -count=1`、`GOOS=android GOARCH=arm64 CGO_ENABLED=0 go build ./...`
Expected: 全部通过

- [ ] **Step 3: 提交**

```bash
git commit -m "chore(ui): 删除 internal/ui（框架迁移完成）

Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>"
```

---

### Task 5: 文档同步

**Files:**
- Modify: `CLAUDE.md`、`AGENTS.md`、`docs/开发手册.md`

- [ ] **Step 1: CLAUDE.md 与 AGENTS.md（两文件同步，改同样内容）**

1. Project Structure：
   - `main.go` 行改为：`# Entry: config → ui.RunShell（描述符接线）；OnStart 内 ApplyAll 回写后 buildRuntime`
   - 删除 `│   └── ui/                      # ImGui shell: ...`（internal/ui）整行
   - `ui/` 行保持现状（已是 Phase 2 描述）
2. Key Architectural Notes 的「UI shell & config binding」条重写为描述符流程：
   - `ui.RunShell(ShellOptions)`：Tasks（字段描述符）+ Nav（TaskListPage/SystemPage）+ Controller + StatusSource
   - 配置流：config.json → Field.Seed（首帧 LoadConfig 后仅填缺失键）→ 面板改 Store → OnStart `ui.ApplyAll` 回写 cfg → buildRuntime
   - **加用户设置 = 一个 Field 描述符**（取代"四处同键"）；键与设备 ui.json 持久化兼容由描述符键决定
   - 遮挡策略（开面板自动暂停/关面板恢复、岛卡片 6s 收起、识别避开岛区）迁至 `ui.Shell`/绘制层，措辞保留
   - 状态上报：StatusSource 窄接口 + status.Reporter，措辞保留
3. 「Adding a Game Task」第 3 步改为：在 main 的组合根加 `ui.Task` 描述符（Fields 用 `ui.Bool/Number/Text`，键决定 ui.json 持久化兼容）；第 4 步注册 Build 不变。
4. 全文 grep `internal/ui`、`binding.go`、`SeedFromConfig`、`ApplyToConfig`、`cookie_panel`、`四处`/`same key` 相关段落，逐处改写或删除。

- [ ] **Step 2: docs/开发手册.md**

grep `internal/ui`、`binding.go`、`SeedFromConfig`、`ApplyToConfig`、`UI_创建` 的段落（重点 §6 UI 绑定与 §13 坑、§15 检查单），改写为描述符流程：加设置 = Field 描述符一处声明；面板页 = NavEntry 挂载；自定义详情 = `Task.RenderDetail`。保留 §6.4 状态上报机制描述（接口名换 StatusSource）。

- [ ] **Step 3: 验证**

Run: `grep -rn "internal/ui\|SeedFromConfig\|ApplyToConfig\|binding.go" CLAUDE.md AGENTS.md docs/开发手册.md`
Expected: 无残留（历史计划/ADR 文档不在此列，不动）

- [ ] **Step 4: 提交**

```bash
git add CLAUDE.md AGENTS.md docs/开发手册.md
git commit -m "docs: 应用侧切换到 ui 框架描述符接线（internal/ui 退役）

Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>"
```

---

## 收尾

- 全分支终审（fable）：main..HEAD review package，重点核对——设备兼容性声明（键/路径/格式）、描述符闭包正确性（*int 映射）、删除彻底性、文档一致性。
- 合并 main（fast-forward，沿 Phase 1/2 惯例）。
- **设备验证（人工程序，无法在 CI 做）**：插件 NDK 构建一次 APK 上机——首帧 LoadConfig/Seed、面板 Form 渲染与回写、开面板自动暂停、岛状态展示、清缓存后 Seed。这是 Phase 1-3 全部 android 文件第一次真实编译，风险集中在 imgui 标识符拼写（已经过约 130 处人工核验）。

## Self-Review 记录

- **兼容性**：store 键、nav id（tasks/system）、设备路径、ui.json 序列化格式全部不变；MaxBattles 的 *int↔0 映射与旧 ApplyToConfig 逐项一致。
- **有意行为差**：① 旧数字框 hint（"0=不限"）无描述符对应，v1 损失；② 桌面 stub 依旧直接 Start（不可在本机运行验证 UI）；③ main.go 不再启动前 Seed（框架首帧 Seed，语义等价）。
- **范围纪律**：不加新框架能力；不顺手重构 buildRuntime/scheduler；开发手册只改与 internal/ui 直接矛盾的段落。
