package market

import (
	"time"

	"app/internal/game/common/kingdom"
	"app/internal/guard"
	"app/internal/platform/action"
	"app/internal/platform/screen"
	"app/internal/statemachine"
	"app/internal/status"
)

// page is the interface required by Task to interact with the market UI.
// The public *Page type implements this interface; tests inject mocks.
type page interface {
	IsCurrent() bool
	WaitCurrent(timeout time.Duration) bool
	EnsureItemTab() bool
	IsFreeRefresh() bool
	TapRefresh()
	ReadRestockSeconds() (int, string, bool)
	StockKeys() []string
	PurchaseWishlist(items []string) PurchaseStats
}

// route is the interface required by Task for entering and leaving the market.
type route interface {
	Enter() bool
	Leave() bool
}

type Task struct {
	cfg         *Config
	page        page
	route       route
	state       *State
	sm          *statemachine.Machine
	kingdomPage *kingdom.Page
	guard       *guard.Guard
	shouldStop  func() bool
	reporter    *status.Reporter
}

func NewTask(
	cfg *Config,
	det screen.Detector,
	exec action.Executor,
	feature *Feature,
	kingdomPage *kingdom.Page,
	state *State,
	guard *guard.Guard,
) *Task {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	page := NewPage(det, exec, feature)
	route := NewRoute(page, kingdomPage, exec)
	return &Task{
		cfg:         cfg,
		page:        page,
		route:       route,
		state:       state,
		sm:          statemachine.New(),
		kingdomPage: kingdomPage,
		guard:       guard,
	}
}

// SetShouldStop wires a stop probe (typically runtime.IsStopped) so the
// task-level state machine can exit between ticks when the script stops.
func (t *Task) SetShouldStop(fn func() bool) {
	t.shouldStop = fn
}

// SetStatusReporter 接入任务状态上报（灵动岛展示扫货阶段/统计），nil 表示不上报。
func (t *Task) SetStatusReporter(r *status.Reporter) {
	t.reporter = r
}

// pushStatus 推一行状态给灵动岛；未接入上报时无操作。
func (t *Task) pushStatus(text string) {
	if t.reporter == nil {
		return
	}
	t.reporter.Set(text)
}

func (t *Task) Run() error {
	return t.runWithOptions(statemachine.RunOptions{
		Interval: 500 * time.Millisecond,
		Label:    "海滩交易所",
	})
}

// runWithOptions runs the market state machine with the supplied run options.
// It is used by tests to drive the task with fast intervals.
func (t *Task) runWithOptions(opts statemachine.RunOptions) error {
	t.sm.Init("detect", statemachine.Options{
		MaxRetry:      3,
		MaxError:      3,
		Timeout:       30 * time.Minute,
		RetryInterval: 1 * time.Second,
	})
	t.sm.Ctx = &marketCtx{task: t, cfg: t.cfg}

	if opts.Label == "" {
		opts.Label = "海滩交易所"
	}
	if t.guard != nil {
		opts.Guard = t.guard.Check
	}
	if opts.ShouldStop == nil && t.shouldStop != nil {
		opts.ShouldStop = t.shouldStop
	}
	return t.sm.Run(t.handlers(), opts)
}

// newTask constructs a Task with injected page, route and state.
// It is intended for unit tests in package market.
func newTask(cfg *Config, p page, r route, state *State, guard *guard.Guard) *Task {
	return &Task{
		cfg:   cfg,
		page:  p,
		route: r,
		state: state,
		sm:    statemachine.New(),
		guard: guard,
	}
}
