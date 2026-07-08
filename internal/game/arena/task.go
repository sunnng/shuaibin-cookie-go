package arena

import (
	"time"

	"app/internal/config"
	"app/internal/game/common/kingdom"
	"app/internal/guard"
	"app/internal/platform/action"
	"app/internal/platform/screen"
	"app/internal/statemachine"
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
	session     *Session
	sm          *statemachine.Machine
	kingdomPage *kingdom.Page
	guard       *guard.Guard
}

func NewTask(
	cfg *config.Arena,
	det screen.Detector,
	exec action.Executor,
	feature *Feature,
	kingdomPage *kingdom.Page,
	kingdomFeature *kingdom.Feature,
	session *Session,
	guard *guard.Guard,
) *Task {
	_ = kingdomFeature
	page := NewPage(det, exec, feature)
	route := NewRoute(page, kingdomPage)
	return &Task{
		cfg:         cfg,
		page:        page,
		route:       route,
		session:     session,
		sm:          statemachine.New(),
		kingdomPage: kingdomPage,
		guard:       guard,
	}
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
	return t.sm.Run(t.handlers(), opts)
}

// newTask constructs a Task with injected page, route and session.
// It is intended for unit tests in package arena.
func newTask(cfg *config.Arena, p page, r route, session *Session, guard *guard.Guard) *Task {
	return &Task{
		cfg:     cfg,
		page:    p,
		route:   r,
		session: session,
		sm:      statemachine.New(),
		guard:   guard,
	}
}
