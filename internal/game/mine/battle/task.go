package battle

import (
	"time"

	"app/internal/game/common/kingdom"
	"app/internal/game/mine"
	"app/internal/guard"
	"app/internal/platform/action"
	"app/internal/platform/screen"
	"app/internal/statemachine"
)

// Config 矿山战斗配置（对应 Lua config.lua USER.mine 的战斗字段）。
type Config struct {
	Enabled     bool     // battleEnabled，默认 false
	IntervalSec int      // battleIntervalSec，战斗检测间隔秒，默认 21600（6 小时）
	SoulStones  []string // battleSoulStones，目标灵魂石名称（同特征库键名，中文）
}

func DefaultConfig() *Config {
	return &Config{
		Enabled:     false,
		IntervalSec: 21600,
		SoulStones:  []string{"妖精王", "莓果", "雷神武将"},
	}
}

// page is the interface required by Task to interact with the battle UI.
// The public *Page type implements this interface; tests inject mocks.
type page interface {
	IsBattlePage() bool
	WaitBattlePage(timeout time.Duration) bool
	TapBackBtn()
	FindQuickBattleButton() (screen.Point, bool)
	TapQuickBattleButton(pt screen.Point)
	WaitQuickBattleDialog(timeout time.Duration) bool
	WaitQuickBattleDialogGone(timeout time.Duration) bool
	ReadClockCount() (int, int, bool)
	TapQuickBattleConfirm()
	TapQuickBattleCancel()
	TapSettleUntilBattlePage() bool
	FindBattleCards() []screen.Point
	TapBattleCard(pt screen.Point)
	RecognizeSoulStoneType(targets map[string]bool) string
	SwipeUpAndCheckLastPage() bool
}

// homePage 是 Task 需要的矿山首页窄接口（*mine.Page 实现）。
type homePage interface {
	IsCurrent() bool
	WaitCurrent(timeout time.Duration) bool
	TapBattle()
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

// Run 执行任务。与 Lua 一致：运行前先记录本次战斗时间（即使失败也推开下次调度）。
func (t *Task) Run() error {
	t.state.RecordBattle()
	return t.runWithOptions(statemachine.RunOptions{
		Interval: 500 * time.Millisecond,
		Label:    "矿山战斗",
	})
}

// runWithOptions runs the battle state machine with the supplied run options.
// It is used by tests to drive the task with fast intervals.
func (t *Task) runWithOptions(opts statemachine.RunOptions) error {
	t.sm.Init("detect", statemachine.Options{
		MaxRetry:      3,
		MaxError:      3,
		Timeout:       30 * time.Minute,
		RetryInterval: 1 * time.Second,
	})
	t.sm.Ctx = &battleCtx{task: t, cfg: t.cfg}

	if opts.Label == "" {
		opts.Label = "矿山战斗"
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
// It is intended for unit tests in package battle.
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

// resolveTargetSoulStones 用户勾选的灵魂石名称集合（Lua resolveTargetSoulStones）。
func resolveTargetSoulStones(cfg *Config) map[string]bool {
	out := map[string]bool{}
	if cfg == nil {
		return out
	}
	for _, name := range cfg.SoulStones {
		out[name] = true
	}
	return out
}
