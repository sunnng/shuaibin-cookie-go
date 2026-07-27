package starlight

import (
	"time"

	"app/internal/game/common/kingdom"
	"app/internal/guard"
	"app/internal/platform/action"
	"app/internal/platform/screen"
	"app/internal/statemachine"
	"app/internal/status"
)

// Config 梦幻繁星岛任务配置，对应 Lua config.lua 的 USER.starlight 段。
type Config struct {
	Enabled bool // 默认 false（Lua USER.starlight.enabled = false）
}

func DefaultConfig() *Config {
	return &Config{Enabled: false}
}

// page 是 Task 依赖的页面窄接口；公开的 *Page 实现它，测试注入 mock。
type page interface {
	IsHomePage() bool
	WaitHomePage(timeout time.Duration) bool
	TapSailingManual() bool
	TapTaskBtn() bool
	TapBackToKingdom() bool
	IsManualPage() bool
	WaitManualPage(timeout time.Duration) bool
	TapLoginIsland() bool
	IsVanillaIslandPage() bool
	WaitVanillaIslandPage(timeout time.Duration) bool
	TapBackFromVanilla() bool
	IsTaskPage() bool
	WaitTaskPage(timeout time.Duration) bool
	TapBackFromTask() bool
	FindClaimableBtn() (screen.Point, bool)
	TapClaimableBtn(pt screen.Point)
	SettleAfterClaim(check func() bool)
	DismissRewardPopupIfNeeded()
}

// route 是 Task 依赖的导航窄接口；公开的 *Route 实现它，测试注入 mock。
type route interface {
	IsStarlightHome() bool
	EnsureHome() bool
}

// kingdomHome 是 finish 态回王国首页所需的窄接口；*kingdom.Page 实现它。
type kingdomHome interface {
	WaitHome(timeout time.Duration) bool
}

type Task struct {
	cfg        *Config
	page       page
	route      route
	state      *State
	sm         *statemachine.Machine
	kingdom    kingdomHome
	guard      *guard.Guard
	shouldStop func() bool
	reporter   *status.Reporter
}

func NewTask(
	cfg *Config,
	det screen.Detector,
	exec action.Executor,
	feature *Feature,
	kingdomPage *kingdom.Page,
	state *State,
	g *guard.Guard,
) *Task {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	page := NewPage(det, exec, feature)
	route := NewRoute(page, kingdomPage)
	return &Task{
		cfg:     cfg,
		page:    page,
		route:   route,
		state:   state,
		sm:      statemachine.New(),
		kingdom: kingdomPage,
		guard:   g,
	}
}

// SetShouldStop 接入停止探针（通常是 runtime.IsStopped），脚本停止时
// 任务级状态机可在 tick 间退出。
func (t *Task) SetShouldStop(fn func() bool) {
	t.shouldStop = fn
}

// SetStatusReporter 接入任务状态上报（灵动岛展示当前步骤），nil 表示不上报。
func (t *Task) SetStatusReporter(r *status.Reporter) {
	t.reporter = r
}

// pushStatus 把当前步骤文本推给灵动岛；未接入上报时无操作。
func (t *Task) pushStatus(text string) {
	if t.reporter == nil {
		return
	}
	t.reporter.Set("梦幻繁星岛 " + text)
}

func (t *Task) Run() error {
	return t.runWithOptions(statemachine.RunOptions{
		Interval: 500 * time.Millisecond,
		Label:    "梦幻繁星岛",
	})
}

// runWithOptions 以指定的运行参数驱动状态机；测试用短间隔调用。
func (t *Task) runWithOptions(opts statemachine.RunOptions) error {
	t.sm.Init("check", statemachine.Options{
		MaxRetry:      3,
		MaxError:      3,
		Timeout:       3 * time.Minute, // Lua timeout = 180s
		RetryInterval: 1 * time.Second,
	})
	t.sm.Ctx = &starlightCtx{task: t}

	if opts.Label == "" {
		opts.Label = "梦幻繁星岛"
	}
	if t.guard != nil {
		opts.Guard = t.guard.Check
	}
	if opts.ShouldStop == nil && t.shouldStop != nil {
		opts.ShouldStop = t.shouldStop
	}
	err := t.sm.Run(t.handlers(), opts)
	if err != nil {
		t.pushStatus("失败")
	} else {
		t.pushStatus("今日已完成")
	}
	return err
}

// newTask 用注入的 page/route/state 构造 Task，供包内单测使用。
func newTask(cfg *Config, p page, r route, k kingdomHome, state *State, g *guard.Guard) *Task {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &Task{
		cfg:     cfg,
		page:    p,
		route:   r,
		state:   state,
		sm:      statemachine.New(),
		kingdom: k,
		guard:   g,
	}
}
