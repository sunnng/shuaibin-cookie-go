package survey

import (
	"time"

	"app/internal/game/common/kingdom"
	"app/internal/game/mine"
	"app/internal/guard"
	"app/internal/logger"
	"app/internal/platform/action"
	"app/internal/platform/screen"
	"app/internal/statemachine"
)

// Config 矿山勘查配置（对应 Lua config.lua USER.mine 的勘查字段）。
type Config struct {
	Enabled     bool // surveyEnabled，默认 true
	TargetFloor int  // targetFloor，目标层数，默认 6
	FarGap      int  // farGap，近距阈值：|目标-当前|<=farGap 时原地轮询，默认 2
	OCRPollSec  int  // ocrPollSec，近距轮询 OCR 间隔秒，默认 60
	FarWaitSec  int  // farWaitSec，远距回城等待秒，默认 600
}

func DefaultConfig() *Config {
	return &Config{
		Enabled:     true,
		TargetFloor: 6,
		FarGap:      2,
		OCRPollSec:  60,
		FarWaitSec:  600,
	}
}

// page is the interface required by Task to interact with the survey UI.
// The public *Page type implements this interface; tests inject mocks.
type page interface {
	IsDomain() bool
	IsRunning() bool
	Setup() bool
	StopVenture() bool
	GetCurrentFloor() (int, bool)
	TapBackBtn()
}

// homePage 是 Task 需要的矿山首页窄接口（*mine.Page 实现）。
type homePage interface {
	IsCurrent() bool
	WaitCurrent(timeout time.Duration) bool
	TapVenture()
}

// route 是 Task 需要的共享路由窄接口（*mine.Route 实现）。
type route interface {
	KingdomHomeToMineHome() bool
	MineHomeToKingdom() bool
}

// kingdomPage 是 Task 需要的王国页窄接口（*kingdom.Page 实现）。
// 与 arena 传具体类型不同，这里收窄成接口以便 detect/navigate 分支可注入 mock。
type kingdomPage interface {
	IsKingdomHome() bool
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

// Run 执行任务。远距等待未到期时本轮直接跳过（Lua resolveInitialState 的双保险；
// 正常情况下调度 CheckReady 已挡住）。
func (t *Task) Run() error {
	if remain := t.state.RestoreProgress(); remain > 0 {
		logger.Infof("[矿山勘查] 远距等待中，剩余 %v，本轮跳过", remain)
		return nil
	}
	return t.runWithOptions(statemachine.RunOptions{
		Interval: 500 * time.Millisecond,
		Label:    "矿山勘查",
	})
}

// resolveInitialState 按当前页面决定起点（Lua resolveInitialState 第 2、3 步）。
func (t *Task) resolveInitialState() string {
	if t.page.IsDomain() {
		return "prepare"
	}
	if t.home.IsCurrent() || t.kingdom.IsKingdomHome() {
		return "navigate"
	}
	return "detect"
}

// runWithOptions runs the survey state machine with the supplied run options.
// It is used by tests to drive the task with fast intervals.
func (t *Task) runWithOptions(opts statemachine.RunOptions) error {
	t.sm.Init(t.resolveInitialState(), statemachine.Options{
		MaxRetry:      3,
		MaxError:      3,
		Timeout:       30 * time.Minute,
		RetryInterval: 1 * time.Second,
	})
	t.sm.Ctx = &surveyCtx{task: t, cfg: t.cfg}

	if opts.Label == "" {
		opts.Label = "矿山勘查"
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
// It is intended for unit tests in package survey.
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
