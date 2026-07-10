# Floating Ball UI Shell Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 把 ImGui 悬浮球接入 `internal/ui`，用统一 UI 壳驱动面板+球，并通过 Controller 控制 runtime 的开始/暂停/停止/退出。

**Architecture:** 单一 `imgui.Run` 循环（`RunShell`）每帧绘制悬浮球，并按需绘制配置面板；脚本在 goroutine 中跑。`ui.Controller` 解耦 UI 与 bot；`runtime` 在每轮开头支持 Pause 挂起。

**Tech Stack:** Go、AutoGo `imgui`/`device`、现有 `internal/runtime` / `internal/ui` / `internal/scheduler`

**Spec:** `docs/superpowers/specs/2026-07-10-floating-ball-design.md`

## Global Constraints

- Android UI 文件使用 `//go:build android && cgo`
- 非 Android stub 必须让 `go test ./...` / `go build ./...` 通过
- 不热更新运行中配置；不隐藏球再唤起；不改 `Android.toml` 系统悬浮球
- 退出进程只走球「关闭」→ `Exit`；关面板标题栏 X 只关面板
- Pause 粒度：轮次边界挂起，不打断当前 scheduler step
- 提交前勿擅自 `git commit`（除非用户明确要求）

## File Structure

| 文件 | 职责 |
|------|------|
| `internal/runtime/runtime.go` | 增加 Pause/Resume；Run 循环开头挂起 |
| `internal/runtime/runtime_test.go` | Pause/Resume/Stop 打断 Pause |
| `internal/ui/controller.go` | `ScriptState` + `Controller` 接口 + 可测 `SessionController` |
| `internal/ui/controller_test.go` | 状态转换与重复 Start |
| `internal/ui/ball_android.go` | 悬浮球绘制/拖动/吸附/展开（含设置按钮） |
| `internal/ui/shell_android.go` | `ShellOptions` + `RunShell` |
| `internal/ui/panel_android.go` | 面板改为可嵌入壳；去掉阻塞式独占 Run |
| `internal/ui/panel_stub.go` | `RunShell` stub → `Controller.Start()` |
| `main.go` | `BotController` 接线 + `RunShell` |

---

### Task 1: Runtime Pause / Resume

**Files:**
- Modify: `internal/runtime/runtime.go`
- Modify: `internal/runtime/runtime_test.go`

**Interfaces:**
- Produces: `(*Runtime).Pause()`, `(*Runtime).Resume()`；`Run` 在每轮开头调用内部 `waitIfPaused`
- Consumes: 现有 `Stop()` / `stopCh`

- [ ] **Step 1: 写失败测试 `TestRuntimePauseBlocksScheduling`**

在 `runtime_test.go` 追加（若 `mockDetector` 缺方法导致编译失败，按当前 `Detector` 接口补全 stub 方法）：

```go
func TestRuntimePauseBlocksScheduling(t *testing.T) {
	var mu sync.Mutex
	calls := 0

	s := scheduler.New()
	s.Add("tick", func() bool { return true }, func() error {
		mu.Lock()
		calls++
		mu.Unlock()
		return nil
	})

	rt := New(Options{Scheduler: s, RuntimeCfg: RuntimeConfig{
		GuardInterval: 10 * time.Millisecond,
		IdleDelay:     10 * time.Millisecond,
		StepDelay:     20 * time.Millisecond,
	}})

	done := make(chan struct{})
	go func() {
		_ = rt.Run()
		close(done)
	}()

	time.Sleep(60 * time.Millisecond)
	rt.Pause()
	time.Sleep(40 * time.Millisecond)

	mu.Lock()
	pausedCalls := calls
	mu.Unlock()

	time.Sleep(80 * time.Millisecond)

	mu.Lock()
	after := calls
	mu.Unlock()
	if after != pausedCalls {
		t.Fatalf("expected no new calls while paused, before=%d after=%d", pausedCalls, after)
	}

	rt.Resume()
	time.Sleep(80 * time.Millisecond)

	mu.Lock()
	resumed := calls
	mu.Unlock()
	if resumed <= pausedCalls {
		t.Fatalf("expected calls to resume, paused=%d resumed=%d", pausedCalls, resumed)
	}

	rt.Stop()
	<-done
}

func TestRuntimeStopUnblocksPause(t *testing.T) {
	s := scheduler.New()
	rt := New(Options{Scheduler: s, RuntimeCfg: RuntimeConfig{
		GuardInterval: 10 * time.Millisecond,
		IdleDelay:     50 * time.Millisecond,
		StepDelay:     20 * time.Millisecond,
	}})

	done := make(chan struct{})
	go func() {
		_ = rt.Run()
		close(done)
	}()

	time.Sleep(30 * time.Millisecond)
	rt.Pause()

	select {
	case <-done:
		t.Fatal("Run returned while paused without Stop")
	case <-time.After(50 * time.Millisecond):
	}

	rt.Stop()
	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Stop did not unblock paused Run")
	}
}
```

并在文件 import 中加入 `"sync"`（若尚未导入）。

- [ ] **Step 2: 跑测试确认失败**

Run: `go test ./internal/runtime -run "TestRuntimePause|TestRuntimeStopUnblocks" -v`

Expected: FAIL（`Pause`/`Resume` undefined）

- [ ] **Step 3: 实现 Pause/Resume**

在 `Runtime` 结构体增加字段，并实现方法。在 `New` 里初始化：

```go
type Runtime struct {
	scheduler *scheduler.Scheduler
	guard     *guard.Guard
	logger    *logger.Logger
	stopCh    chan struct{}
	stopOnce  sync.Once
	cfg       RuntimeConfig

	pauseMu sync.Mutex
	pauseCh chan struct{} // non-nil => paused; close to resume
}

func (r *Runtime) Pause() {
	r.pauseMu.Lock()
	defer r.pauseMu.Unlock()
	if r.pauseCh == nil {
		r.pauseCh = make(chan struct{})
	}
}

func (r *Runtime) Resume() {
	r.pauseMu.Lock()
	defer r.pauseMu.Unlock()
	if r.pauseCh != nil {
		close(r.pauseCh)
		r.pauseCh = nil
	}
}

func (r *Runtime) waitIfPaused() {
	for {
		r.pauseMu.Lock()
		ch := r.pauseCh
		r.pauseMu.Unlock()
		if ch == nil {
			return
		}
		select {
		case <-r.stopCh:
			return
		case <-ch:
			// resumed
		}
	}
}
```

在 `Run` 的 `for` 循环里，`select` 检查 `stopCh` 之后、`guard.Check` 之前调用 `r.waitIfPaused()`；若 `waitIfPaused` 因 stop 返回，下一轮 `select` 会退出。更稳妥：`waitIfPaused` 后立刻再检查 stop：

```go
for {
	select {
	case <-r.stopCh:
		r.infof("[Runtime] stopped")
		return nil
	default:
	}

	r.waitIfPaused()

	select {
	case <-r.stopCh:
		r.infof("[Runtime] stopped")
		return nil
	default:
	}

	// ... existing guard / scheduler / sleep ...
}
```

`Stop` 不变（close `stopCh`）。注意：若处于 Pause，`waitIfPaused` 的 `select` 会收到 `stopCh` 并返回。

- [ ] **Step 4: 跑测试确认通过**

Run: `go test ./internal/runtime -v`

Expected: PASS（含原有测试）

---

### Task 2: Controller 接口 + SessionController

**Files:**
- Create: `internal/ui/controller.go`
- Create: `internal/ui/controller_test.go`

**Interfaces:**
- Produces:
  - `type ScriptState int`：`StateIdle`, `StateRunning`, `StatePaused`
  - `type Controller interface { State() ScriptState; Start(); Pause(); Resume(); Stop(); Exit() }`
  - `type SessionHooks struct { OnStart func() (run func() error, pause, resume, stop func()); OnExit func() }`
  - `type SessionController struct` 实现 `Controller`
- Consumes: 无

- [ ] **Step 1: 写失败测试**

`controller_test.go`：

```go
package ui

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestSessionControllerStateTransitions(t *testing.T) {
	var started atomic.Int32
	var paused atomic.Int32
	var resumed atomic.Int32
	var stopped atomic.Int32

	c := NewSessionController(SessionHooks{
		OnStart: func() (func() error, func(), func(), func()) {
			started.Add(1)
			run := func() error {
				time.Sleep(200 * time.Millisecond)
				return nil
			}
			return run, func() { paused.Add(1) }, func() { resumed.Add(1) }, func() { stopped.Add(1) }
		},
		OnExit: func() {},
	})

	if c.State() != StateIdle {
		t.Fatalf("want Idle")
	}
	c.Start()
	time.Sleep(20 * time.Millisecond)
	if c.State() != StateRunning {
		t.Fatalf("want Running, got %v", c.State())
	}
	c.Start() // duplicate
	if started.Load() != 1 {
		t.Fatalf("duplicate Start should be ignored, started=%d", started.Load())
	}

	c.Pause()
	if c.State() != StatePaused || paused.Load() != 1 {
		t.Fatalf("want Paused")
	}
	c.Resume()
	if c.State() != StateRunning || resumed.Load() != 1 {
		t.Fatalf("want Running after resume")
	}
	c.Stop()
	time.Sleep(20 * time.Millisecond)
	if c.State() != StateIdle || stopped.Load() != 1 {
		t.Fatalf("want Idle after stop")
	}
}

func TestSessionControllerRunEndReturnsIdle(t *testing.T) {
	c := NewSessionController(SessionHooks{
		OnStart: func() (func() error, func(), func(), func()) {
			run := func() error { return nil }
			return run, func() {}, func() {}, func() {}
		},
		OnExit: func() {},
	})
	c.Start()
	time.Sleep(50 * time.Millisecond)
	if c.State() != StateIdle {
		t.Fatalf("want Idle after run ends, got %v", c.State())
	}
}
```

- [ ] **Step 2: 跑测试确认失败**

Run: `go test ./internal/ui -run TestSessionController -v`

Expected: FAIL（undefined）

- [ ] **Step 3: 实现 `controller.go`**

```go
package ui

import (
	"sync"

	"app/internal/logger"
)

type ScriptState int

const (
	StateIdle ScriptState = iota
	StateRunning
	StatePaused
)

type Controller interface {
	State() ScriptState
	Start()
	Pause()
	Resume()
	Stop()
	Exit()
}

// SessionHooks.OnStart 返回：阻塞的 run（在 goroutine 中调用）、以及 pause/resume/stop 钩子。
type SessionHooks struct {
	OnStart func() (run func() error, pause, resume, stop func())
	OnExit  func()
}

type SessionController struct {
	mu     sync.Mutex
	state  ScriptState
	hooks  SessionHooks
	pause  func()
	resume func()
	stop   func()
}

func NewSessionController(hooks SessionHooks) *SessionController {
	return &SessionController{hooks: hooks}
}

func (c *SessionController) State() ScriptState {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.state
}

func (c *SessionController) Start() {
	c.mu.Lock()
	if c.state != StateIdle {
		c.mu.Unlock()
		return
	}
	if c.hooks.OnStart == nil {
		c.mu.Unlock()
		return
	}
	run, pause, resume, stop := c.hooks.OnStart()
	c.pause, c.resume, c.stop = pause, resume, stop
	c.state = StateRunning
	c.mu.Unlock()

	go func() {
		var err error
		if run != nil {
			err = run()
		}
		if err != nil {
			logger.Errorf("script run error: %v", err)
		}
		c.mu.Lock()
		c.state = StateIdle
		c.pause, c.resume, c.stop = nil, nil, nil
		c.mu.Unlock()
	}()
}

func (c *SessionController) Pause() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.state != StateRunning {
		return
	}
	if c.pause != nil {
		c.pause()
	}
	c.state = StatePaused
}

func (c *SessionController) Resume() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.state != StatePaused {
		return
	}
	if c.resume != nil {
		c.resume()
	}
	c.state = StateRunning
}

func (c *SessionController) Stop() {
	c.mu.Lock()
	stop := c.stop
	if c.state == StateIdle {
		c.mu.Unlock()
		return
	}
	c.mu.Unlock()
	if stop != nil {
		stop()
	}
	// state -> Idle 由 run goroutine 结束时设置；若 stop 后 run 立即返回即可。
	// 若 stop 同步结束 run，下面再保险置位：
	c.mu.Lock()
	c.state = StateIdle
	c.pause, c.resume, c.stop = nil, nil, nil
	c.mu.Unlock()
}

func (c *SessionController) Exit() {
	c.Stop()
	if c.hooks.OnExit != nil {
		c.hooks.OnExit()
	}
}
```

注意：`Stop` 与 run 结束竞态——测试里 `Stop` 后允许短暂等待。若 `Stop` 先置 Idle、run 稍后又写 Idle，可接受。避免 run 结束把已重新 Start 的状态打回 Idle：用 generation 计数：

```go
// 在 SessionController 增加 gen int
// Start 时：c.gen++; my := c.gen; 然后 go func() { ...; c.mu.Lock(); if c.gen == my { c.state = Idle; ... }; c.mu.Unlock() }
// Stop 时同样只清理当前 gen
```

实现时必须带 `gen`，防止 Stop 后立刻再 Start 被旧 goroutine 覆盖。

- [ ] **Step 4: 跑测试确认通过**

Run: `go test ./internal/ui -run TestSessionController -v`

Expected: PASS

---

### Task 3: 重构面板 API + Stub RunShell

**Files:**
- Modify: `internal/ui/panel_android.go`
- Modify: `internal/ui/panel_stub.go`

**Interfaces:**
- Produces: `ShellOptions`（两端都有）、`RunShell(opts ShellOptions)`；android 上 `drawConfigPanel(...)` 可被壳调用
- Consumes: `Controller`（Task 2）
- 保留 `PanelOptions` / `RunPanel` 作为薄封装转调 `RunShell`（兼容），或删除并由 main 改用 `RunShell`——本计划采用：**定义 `ShellOptions` + `RunShell`，`RunPanel` 转调以减少一次性改动面**

- [ ] **Step 1: 更新 stub `panel_stub.go`**

完整替换为：

```go
//go:build !android || !cgo

package ui

type PanelOptions struct {
	Title        string
	ConfigPath   string
	CountdownSec float64
	Store        *Store
	Render       func(store *Store)
	OnRun        func(store *Store)
	OnClose      func(store *Store)
}

type ShellOptions struct {
	Title            string
	ConfigPath       string
	CountdownSec     float64
	Store            *Store
	Render           func(store *Store)
	Controller       Controller
	OpenPanelOnStart bool
}

func RunPanel(opts PanelOptions) {
	if opts.Store == nil {
		opts.Store = NewStore()
	}
	if opts.OnRun != nil {
		opts.OnRun(opts.Store)
	}
}

func RunShell(opts ShellOptions) {
	if opts.Store == nil {
		opts.Store = NewStore()
	}
	if opts.Controller != nil {
		opts.Controller.Start()
	}
}

func DefaultCookiePanel(store *Store) {}
```

- [ ] **Step 2: 改 android `panel_android.go`——抽出面板绘制，RunPanel 转调 RunShell**

将 `PanelOptions` 保留；新增与 stub 相同的 `ShellOptions`（android 文件里也要有类型定义，或把 `ShellOptions`/`PanelOptions` 挪到无 build tag 的 `options.go`——**推荐新建 `internal/ui/options.go` 放两个 Options 结构体**，两端 `RunShell`/`RunPanel` 只留函数）。

创建 `internal/ui/options.go`（无 build tag）：

```go
package ui

type PanelOptions struct {
	Title        string
	ConfigPath   string
	CountdownSec float64
	Store        *Store
	Render       func(store *Store)
	OnRun        func(store *Store)
	OnClose      func(store *Store)
}

type ShellOptions struct {
	Title            string
	ConfigPath       string
	CountdownSec     float64
	Store            *Store
	Render           func(store *Store)
	Controller       Controller
	OpenPanelOnStart bool
}
```

从 `panel_android.go` / `panel_stub.go` **删除**重复的 struct 定义。

`panel_android.go` 暂改为：

```go
//go:build android && cgo

package ui

import (
	"github.com/Dasongzi1366/AutoGo/device"
	"github.com/Dasongzi1366/AutoGo/imgui"
)

// RunPanel 兼容入口：无 Controller 时用 OnRun 包一层 SessionController 不合适；
// 保留行为：直接 RunShell 且 Controller 为 nil 时仅显示面板逻辑——
// 本任务先让 RunPanel 转调旧行为的最小壳；Task 5 用完整 RunShell 替换。
func RunPanel(opts PanelOptions) {
	// 过渡：仍用旧独占循环，直到 Task 5。
	runPanelLegacy(opts)
}

func runPanelLegacy(opts PanelOptions) {
	// 把当前 RunPanel 函数体移到这里（原样）
}
```

本 Task 目标：Options 去重 + stub `RunShell` 可编译。完整壳在 Task 5。

- [ ] **Step 3: 验证编译与测试**

Run: `go test ./internal/ui ./internal/runtime -count=1`

Expected: PASS

---

### Task 4: 悬浮球（Android）

**Files:**
- Create: `internal/ui/ball_android.go`
- Reference: `c:\Users\1\Downloads\imgui悬浮球.go`

**Interfaces:**
- Consumes: `Controller.State()`（颜色）；`BallCallbacks`
- Produces: `type BallCallbacks struct{ OnSettings, OnStartStop, OnPauseResume, OnClose func() }`；`type FloatingBall struct`；`NewFloatingBall() *FloatingBall`；`(*FloatingBall).Draw(cb BallCallbacks, state ScriptState)`

- [ ] **Step 1: 创建 `ball_android.go`，移植示例并改造成类型化 API**

要点（相对示例的差异，必须全部落地）：

1. `package ui`，`//go:build android && cgo`
2. 去掉 `main`/`init` 全局 `ball`；改为 `FloatingBall` 实例字段
3. 展开菜单 **5 个按钮**（右贴边向左）：Logo(0)、关闭(1)、暂停(2)、开始/停止(3)、**设置(4)**
4. 窗口宽度按 `buttonSpacing*4` 计算（原为 `*3`）
5. `Draw(cb BallCallbacks, state ScriptState)`：用 `state` 上色（Idle 蓝 / Running 绿 / Paused 黄 `Vec4{1,0.85,0.2,0.95}`），点击走回调而非改内部 `IsRunning`
6. 球内部可保留 `IsExpanded` 等 UI 状态；**不要**在球内保存 `IsRunning`/`IsPaused` 业务状态
7. `checkButtonClick`：
   - 0 → 收起
   - 1 → `cb.OnClose`
   - 2 → `cb.OnPauseResume`
   - 3 → `cb.OnStartStop`
   - 4 → `cb.OnSettings`
8. 屏幕尺寸：首次 `Draw` 时 `device.GetDisplayInfo(0)`，与示例相同
9. 缓动函数、图标绘制从示例原样迁入（可私有函数）

`BallCallbacks` 与 `FloatingBall` 导出部分：

```go
type BallCallbacks struct {
	OnSettings    func()
	OnStartStop   func()
	OnPauseResume func()
	OnClose       func()
}

func NewFloatingBall() *FloatingBall {
	return &FloatingBall{Radius: 40, IsOnRightSide: true}
}

func (ball *FloatingBall) Draw(cb BallCallbacks, state ScriptState) {
	// updateAnimations + drawInteractionWindow...
}
```

设置按钮颜色可用：`imgui.Vec4{X: 0.4, Y: 0.5, Z: 0.9, W: 1.0}`；图标用简单齿轮近似（两个圆 + 中心点）或小矩形组合即可。

- [ ] **Step 2: 桌面侧确认无 android 文件不参与编译**

Run: `go test ./internal/ui -count=1`

Expected: PASS（不编译 `ball_android.go`）

---

### Task 5: RunShell 完整实现（Android）

**Files:**
- Create: `internal/ui/shell_android.go`
- Modify: `internal/ui/panel_android.go`（删除 legacy 独占循环，改为 `drawConfigPanel`）

**Interfaces:**
- Consumes: `FloatingBall`, `Controller`, `DefaultCookiePanel`, `Store`
- Produces: `RunShell(opts ShellOptions)` 阻塞（`imgui.Run` + `select {}`）

- [ ] **Step 1: 实现 `shell_android.go`**

```go
//go:build android && cgo

package ui

import (
	"github.com/Dasongzi1366/AutoGo/imgui"
)

func RunShell(opts ShellOptions) {
	if opts.Store == nil {
		opts.Store = NewStore()
	}
	if opts.Title == "" {
		opts.Title = "帅宾 Cookie"
	}
	if opts.ConfigPath == "" {
		opts.ConfigPath = "/sdcard/shuaibin-cookie/ui.json"
	}
	if opts.CountdownSec <= 0 {
		opts.CountdownSec = 300
	}
	if opts.Render == nil {
		opts.Render = DefaultCookiePanel
	}
	openPanel := opts.OpenPanelOnStart
	// 零值 false 时：若调用方想默认打开，main 传 OpenPanelOnStart: true（spec 默认 true）

	_ = imgui.Init()
	ball := NewFloatingBall()
	loaded := false

	titleBgActive := HexToVec4("#686868")
	imgui.PushStyleColorVec4(imgui.ColTitleBgActive, titleBgActive)

	imgui.Run(func() {
		if !loaded {
			loaded = true
			_ = opts.Store.LoadConfig(opts.ConfigPath)
		}

		state := StateIdle
		if opts.Controller != nil {
			state = opts.Controller.State()
		}

		ball.Draw(BallCallbacks{
			OnSettings: func() { openPanel = true },
			OnStartStop: func() {
				if opts.Controller == nil {
					return
				}
				switch opts.Controller.State() {
				case StateIdle:
					_ = opts.Store.SaveConfig(opts.ConfigPath)
					opts.Controller.Start()
				case StateRunning, StatePaused:
					opts.Controller.Stop()
				}
			},
			OnPauseResume: func() {
				if opts.Controller == nil {
					return
				}
				switch opts.Controller.State() {
				case StateRunning:
					opts.Controller.Pause()
				case StatePaused:
					opts.Controller.Resume()
				}
			},
			OnClose: func() {
				if opts.Controller != nil {
					opts.Controller.Exit()
				}
			},
		}, state)

		if openPanel {
			stillOpen := drawConfigPanel(opts, &openPanel)
			if !stillOpen {
				openPanel = false
			}
		}
	})
	select {}
}

func RunPanel(opts PanelOptions) {
	// 兼容：无 Controller 时用 OnRun 包一层
	var ctrl Controller
	if opts.OnRun != nil {
		ctrl = NewSessionController(SessionHooks{
			OnStart: func() (func() error, func(), func(), func()) {
				run := func() error {
					opts.OnRun(opts.Store)
					return nil
				}
				return run, func() {}, func() {}, func() {}
			},
			OnExit: func() {},
		})
	}
	RunShell(ShellOptions{
		Title:            opts.Title,
		ConfigPath:       opts.ConfigPath,
		CountdownSec:     opts.CountdownSec,
		Store:            opts.Store,
		Render:           opts.Render,
		Controller:       ctrl,
		OpenPanelOnStart: true,
	})
}
```

- [ ] **Step 2: 把原面板绘制抽成 `drawConfigPanel`**

在 `panel_android.go`：删除 `runPanelLegacy`；实现：

```go
// drawConfigPanel 绘制配置窗。返回 false 表示用户关掉了窗口。
// 点击「运行脚本」：保存配置、Controller.Start、open=false（不退出进程）。
func drawConfigPanel(opts ShellOptions, open *bool) bool {
	// 使用与原 RunPanel 相同的尺寸/位置/样式
	// BeginV(opts.Title, open, flags)
	// 加时按钮 + 倒计时「运行脚本」：
	//   callback: SaveConfig; if Controller != nil { Controller.Start() }; *open = false
	// opts.Render(opts.Store)
	// 若 *open 变为 false 且并非 run：仅关面板
	return *open
}
```

具体布局代码从现有 `RunPanel` 函数体迁移，把 `OnRun` 换成 `opts.Controller.Start()`，把最终 `select {}` / `imgui.Run` 外壳删掉。

- [ ] **Step 3: 桌面测试**

Run: `go test ./... -count=1`

Expected: PASS

---

### Task 6: main.go 接线 BotController

**Files:**
- Modify: `main.go`

**Interfaces:**
- Consumes: `ui.NewSessionController`, `ui.RunShell`, `runtime.Pause/Resume/Stop`, 现有 `runBot` 组装逻辑
- Produces: 可运行的 Android 入口

- [ ] **Step 1: 改写 `main.go`**

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
	"app/internal/store"
	"app/internal/ui"
)

func main() {
	logger.SetLevel(logger.LevelInfo)
	logger.Infof("superbin cookie run kingdom start...")

	cfg, err := config.LoadConfig("config.json")
	if err != nil {
		logger.Errorf("failed to load config: %v", err)
		return
	}

	uiStore := ui.NewStore()
	ui.SeedFromConfig(uiStore, cfg)

	var currentRT *runtime.Runtime

	ctrl := ui.NewSessionController(ui.SessionHooks{
		OnStart: func() (run func() error, pause, resume, stop func()) {
			ui.ApplyToConfig(uiStore, cfg)
			rt := buildRuntime(cfg)
			currentRT = rt
			return func() error {
					err := rt.Run()
					currentRT = nil
					return err
				},
				rt.Pause,
				rt.Resume,
				rt.Stop
		},
		OnExit: func() {
			time.Sleep(500 * time.Millisecond)
			os.Exit(0)
		},
	})

	ui.RunShell(ui.ShellOptions{
		Title:            "Superbin Cookie",
		ConfigPath:       "/sdcard/shuaibin-cookie/ui.json",
		Store:            uiStore,
		Controller:       ctrl,
		OpenPanelOnStart: true,
	})
}

func buildRuntime(cfg *config.Config) *runtime.Runtime {
	det := screen.NewDetector(0)
	exec := action.NewExecutor(0)
	s := store.New("data/store.json")
	g := guard.New(det)
	sched := scheduler.New()

	kingdomFeature := kingdom.DefaultFeature()
	kingdomPage := kingdom.NewPage(det, exec, kingdomFeature)
	arenaFeature := arena.DefaultFeature()
	arenaSession := arena.NewSession(s)
	arenaTask := arena.NewTask(
		&cfg.Modules.Arena,
		det, exec, arenaFeature, kingdomPage, arenaSession, g,
	)

	sched.Build(scheduler.TaskOpts{
		Name: "王国竞技场",
		CheckEnabled: func() bool {
			return cfg.Modules.Arena.Enabled
		},
		CheckReady: func() (bool, time.Duration) {
			if arenaSession.IsReachMaxBattles(&cfg.Modules.Arena) {
				return false, 0
			}
			remain := arenaSession.TimeUntilRefresh()
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

	return runtime.New(runtime.Options{
		Scheduler: sched,
		Guard:     g,
		RuntimeCfg: runtime.RuntimeConfig{
			GuardInterval: 500 * time.Millisecond,
			IdleDelay:     30 * time.Second,
			StepDelay:     5 * time.Second,
			StopOnError:   false,
		},
	})
}
```

删除旧的 `runBot`；`currentRT` 若 unused 可删（hooks 已直接闭包 `rt`）。

- [ ] **Step 2: 全量测试与编译**

Run:

```bash
go test ./... -count=1
go build ./...
```

Expected: 全部 PASS / 无错误

---

### Task 7: 对照 Spec 验收清单（桌面 + 文档）

**Files:**
- 无代码改动（除非发现缺口）

- [ ] **Step 1: Spec 覆盖核对**

对照 `2026-07-10-floating-ball-design.md`：

| Spec 项 | Task |
|---------|------|
| Runtime Pause/Resume | 1 |
| Controller 状态机 | 2 |
| 悬浮球 + 设置按钮 | 4 |
| RunShell 统一循环 | 5 |
| 面板双入口 Start | 5+6 |
| Exit → os.Exit | 6 |
| stub 桌面可测 | 3+6 |
| 不热更新 / 不改系统球 | 遵守 |

- [ ] **Step 2: 真机手测（人工）**

按 spec §8 八条执行；问题记回 issue/下轮修复。

---

## Self-Review (plan author)

1. **Spec coverage:** 目标、组件、按钮语义、Pause 粒度、错误处理、测试、双入口均有对应 Task；热更新/藏球明确排除。  
2. **Placeholders:** 无 TBD；Task 4 因示例过长用「差异清单 + API」而非全文粘贴，实现时打开参考文件移植。  
3. **Type consistency:** `ScriptState` / `Controller` / `SessionHooks` / `ShellOptions` / `BallCallbacks` 在 Task 2–6 命名一致；`OpenPanelOnStart` 由 main 显式传 `true`。
