package arena

import (
	"time"

	"app/internal/config"
	"app/internal/game/common/kingdom"
	"app/internal/guard"
	"app/internal/platform/action"
	"app/internal/platform/screen"
	"app/internal/statemachine"
	"app/internal/status"
)

// page is the interface required by Task to interact with the arena UI.
// The public *Page type implements this interface; tests inject mocks.
type page interface {
	IsLobby() bool
	WaitLobby(timeout time.Duration) bool
	ReadMedalAndTicket() (int, int, bool)
	ReadTrophyCount() (int, bool)
	FindFirstValidOpponent(cfg *config.Arena, myTrophy int) *OpponentInfo
	SwipePageLeft()
	IsFreeRefresh() bool
	TapFreeRefresh()
	ReadRefreshCountdown() (time.Duration, bool)
	BuyTicket()
	RunBattle() (string, bool)
	TapToLobby() bool
	TapOpponentSite(site action.Point)
	HasTeamSelectPage() bool
	WaitTeamSelect(timeout time.Duration) bool
	TapStartBattle()
}

// route is the interface required by Task for entering and leaving the arena.
type route interface {
	Enter() bool
	Leave() bool
}

type Task struct {
	cfg         *config.Arena
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
	cfg *config.Arena,
	det screen.Detector,
	exec action.Executor,
	feature *Feature,
	kingdomPage *kingdom.Page,
	state *State,
	guard *guard.Guard,
) *Task {
	page := NewPage(det, exec, feature)
	route := NewRoute(page, kingdomPage)
	t := &Task{
		cfg:         cfg,
		page:        page,
		route:       route,
		state:       state,
		sm:          statemachine.New(),
		kingdomPage: kingdomPage,
		guard:       guard,
	}
	registerDialogTraps(guard, exec, feature.Dialogs)
	return t
}

// registerDialogTraps 把竞技场弹窗特征注册为 Guard trap，战斗中（状态机每 tick
// 都会跑 Guard）弹窗出现时按 Confirm 区域点掉。未取色（Colors 为空）的弹窗由
// Guard.Register 跳过并告警，不会注册成永远不触发的死 trap。
func registerDialogTraps(g *guard.Guard, exec action.Executor, dialogs DialogsFeature) {
	if g == nil || exec == nil {
		return
	}
	traps := []struct {
		name string
		def  DialogDef
	}{
		{"竞技场弹窗-缺少配料", dialogs.MissingTopping},
		{"竞技场弹窗-部署更多", dialogs.DeployMore},
	}
	for _, tr := range traps {
		def := tr.def
		g.Register(tr.name, def.Identify, func() error {
			exec.Tap(action.RandomIn(def.Confirm))
			exec.Sleep(800)
			return nil
		}, 10)
	}
}

// SetShouldStop wires a stop probe (typically runtime.IsStopped) so the
// task-level state machine can exit between ticks when the script stops.
func (t *Task) SetShouldStop(fn func() bool) {
	t.shouldStop = fn
}

// SetStatusReporter 接入任务状态上报（灵动岛展示战斗次数/胜率），nil 表示不上报。
func (t *Task) SetStatusReporter(r *status.Reporter) {
	t.reporter = r
}

// pushStatus 把当前任务统计推给灵动岛；未接入上报时无操作。
func (t *Task) pushStatus() {
	if t.reporter == nil {
		return
	}
	t.reporter.Set(t.state.StatusText(t.cfg))
}

func (t *Task) Run() error {
	return t.runWithOptions(statemachine.RunOptions{
		Interval: 500 * time.Millisecond,
		Label:    "王国竞技场",
	})
}

// runWithOptions runs the arena state machine with the supplied run options.
// It is used by tests to drive the task with fast intervals.
func (t *Task) runWithOptions(opts statemachine.RunOptions) error {
	t.sm.Init("detect", statemachine.Options{
		MaxRetry:      3,
		MaxError:      3,
		Timeout:       30 * time.Minute,
		RetryInterval: 1 * time.Second,
	})
	t.sm.Ctx = &arenaCtx{task: t, cfg: t.cfg}

	if opts.Label == "" {
		opts.Label = "王国竞技场"
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
// It is intended for unit tests in package arena.
func newTask(cfg *config.Arena, p page, r route, state *State, guard *guard.Guard) *Task {
	return &Task{
		cfg:   cfg,
		page:  p,
		route: r,
		state: state,
		sm:    statemachine.New(),
		guard: guard,
	}
}
