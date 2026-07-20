package production

import (
	"time"

	"app/internal/game/common/kingdom"
	"app/internal/guard"
	"app/internal/platform/action"
	"app/internal/platform/screen"
	"app/internal/statemachine"
	"app/internal/status"
)

// page 是 Task 依赖的页内能力；*Page 实现它，单测可注入 mock。
type page interface {
	IsBoard() bool
	WaitBoard(timeout time.Duration) bool
	TapCollectAll() bool
}

// route 是 Task 依赖的进出导航。
type route interface {
	Enter() bool
	Leave() bool
}

// Task 王国生产任务：调度器 Action 入口。
// 配置/面板/main 注册待特征与流程实现后再接（手册：未实现任务不加配置）。
type Task struct {
	page        page
	route       route
	state       *State
	sm          *statemachine.Machine
	kingdomPage *kingdom.Page
	guard       *guard.Guard
	shouldStop  func() bool
	reporter    *status.Reporter
}

// NewTask 装配页面、路线与任务状态；会尝试注册 Dialogs 陷阱（未取色则跳过）。
func NewTask(
	det screen.Detector,
	exec action.Executor,
	feature *Feature,
	kingdomPage *kingdom.Page,
	state *State,
	g *guard.Guard,
) *Task {
	if feature == nil {
		feature = DefaultFeature()
	}
	p := NewPage(det, exec, feature)
	r := NewRoute(p, kingdomPage)
	t := &Task{
		page:        p,
		route:       r,
		state:       state,
		sm:          statemachine.New(),
		kingdomPage: kingdomPage,
		guard:       g,
	}
	registerDialogTraps(g, exec, feature.Dialogs)
	return t
}

func registerDialogTraps(g *guard.Guard, exec action.Executor, dialogs DialogsFeature) {
	if g == nil || exec == nil {
		return
	}
	_ = dialogs
	// 有 DialogDef 时在此 Register；骨架阶段无弹窗。
}

// SetShouldStop 接入脚本停止探测（通常 runtime.IsStopped）。
func (t *Task) SetShouldStop(fn func() bool) {
	t.shouldStop = fn
}

// SetStatusReporter 接入任务状态上报（灵动岛）。
func (t *Task) SetStatusReporter(r *status.Reporter) {
	t.reporter = r
}

func (t *Task) pushStatus() {
	if t.reporter == nil || t.state == nil {
		return
	}
	t.reporter.Set(t.state.StatusText())
}

// Run 驱动任务流程状态机；骨架以 Done 收束，避免误注册后空转。
func (t *Task) Run() error {
	return t.runWithOptions(statemachine.RunOptions{
		Interval: 500 * time.Millisecond,
		Label:    "王国生产",
	})
}

func (t *Task) runWithOptions(opts statemachine.RunOptions) error {
	t.sm.Init("detect", statemachine.Options{
		MaxRetry:      3,
		MaxError:      3,
		Timeout:       10 * time.Minute,
		RetryInterval: 1 * time.Second,
	})
	t.sm.Ctx = &productionCtx{task: t}

	if opts.Label == "" {
		opts.Label = "王国生产"
	}
	if t.guard != nil {
		opts.Guard = t.guard.Check
	}
	if opts.ShouldStop == nil && t.shouldStop != nil {
		opts.ShouldStop = t.shouldStop
	}
	t.pushStatus()
	return t.sm.Run(t.handlers(), opts)
}

// newTask 供单测注入 page/route/state。
func newTask(p page, r route, state *State, g *guard.Guard) *Task {
	return &Task{
		page:    p,
		route:   r,
		state:   state,
		sm:      statemachine.New(),
		guard:   g,
	}
}
