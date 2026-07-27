package jelly

import (
	"time"

	"app/internal/game/common/kingdom"
	"app/internal/game/mine"
	"app/internal/guard"
	"app/internal/platform/action"
	"app/internal/platform/screen"
	"app/internal/statemachine"
)

// Config 解除洋菜冻配置（对应 Lua config.lua USER.mine 的洋菜冻字段）。
type Config struct {
	Enabled     bool // jellyEnabled，默认 false
	IntervalSec int  // jellyIntervalSec，完成后冷却间隔秒，默认 3600
}

func DefaultConfig() *Config {
	return &Config{
		Enabled:     false,
		IntervalSec: 3600,
	}
}

// page is the interface required by Task to interact with the jelly UI.
// The public *Page type implements this interface; tests inject mocks.
type page interface {
	IsJellyPage() bool
	WaitJellyPage(timeout time.Duration) bool
	CanClaimAll() bool
	TapClaimAll()
	TapSettleUntilJellyPage() bool
	TapBack()
	FindConfigBtn() (screen.Point, bool)
	TapConfigBtn(pt screen.Point)
	WaitConfigPage(timeout time.Duration) bool
	CanChoose() bool
	TapChoose()
	TapConfigBack()
	ReadRemainTime() (time.Duration, bool)
}

// homePage 是 Task 需要的矿山首页窄接口（*mine.Page 实现）。
type homePage interface {
	IsCurrent() bool
	WaitCurrent(timeout time.Duration) bool
	TapJellyEntry()
	TapBack()
}

// route 是 Task 需要的共享路由窄接口（*mine.Route 实现）。
type route interface {
	KingdomHomeToMineHome() bool
}

// kingdomPage 是 Task 需要的王国页窄接口（*kingdom.Page 实现）。
type kingdomPage interface {
	IsKingdomHome() bool
	WaitHome(timeout time.Duration) bool
}

type Task struct {
	cfg      *Config
	page     page
	home     homePage
	route    route
	kingdom  kingdomPage
	state    *State
	sm       *statemachine.Machine
	guard    *guard.Guard
	stopFunc func() bool
}

func NewTask(
	cfg *Config,
	det screen.Detector,
	exec action.Executor,
	feature *Feature,
	home *mine.Page,
	route *mine.Route,
	kingdomPage *kingdom.Page,
	state *State,
	g *guard.Guard,
) *Task {
	return &Task{
		cfg:     cfg,
		page:    NewPage(det, exec, feature),
		home:    home,
		route:   route,
		kingdom: kingdomPage,
		state:   state,
		sm:      statemachine.New(),
		guard:   g,
	}
}

// SetShouldStop wires a stop probe (typically runtime.IsStopped) so the
// task-level state machine can exit between ticks when the script stops.
func (t *Task) SetShouldStop(fn func() bool) {
	t.stopFunc = fn
}

func (t *Task) Run() error {
	return t.runWithOptions(statemachine.RunOptions{
		Interval: 500 * time.Millisecond,
		Label:    "解除洋菜冻",
	})
}

// runWithOptions runs the jelly state machine with the supplied run options.
// It is used by tests to drive the task with fast intervals.
func (t *Task) runWithOptions(opts statemachine.RunOptions) error {
	t.sm.Init("detect", statemachine.Options{
		MaxRetry:      3,
		MaxError:      3,
		Timeout:       30 * time.Minute,
		RetryInterval: 1 * time.Second,
	})
	t.sm.Ctx = &jellyCtx{task: t, cfg: t.cfg}

	if opts.Label == "" {
		opts.Label = "解除洋菜冻"
	}
	if t.guard != nil {
		opts.Guard = t.guard.Check
	}
	if opts.ShouldStop == nil && t.stopFunc != nil {
		opts.ShouldStop = t.stopFunc
	}
	return t.sm.Run(t.handlers(), opts)
}

// newTask constructs a Task with injected page, home, route, kingdom and state.
// It is intended for unit tests in package jelly.
func newTask(cfg *Config, p page, h homePage, r route, kp kingdomPage, state *State, g *guard.Guard) *Task {
	return &Task{
		cfg:     cfg,
		page:    p,
		home:    h,
		route:   r,
		kingdom: kp,
		state:   state,
		sm:      statemachine.New(),
		guard:   g,
	}
}
