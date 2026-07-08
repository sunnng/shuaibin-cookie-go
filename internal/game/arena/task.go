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

type Task struct {
	cfg         *config.Arena
	page        *Page
	route       *Route
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
	t.sm.Init("detect", statemachine.Options{
		MaxRetry:      3,
		MaxError:      3,
		Timeout:       30 * time.Minute,
		RetryInterval: 1 * time.Second,
	})
	t.sm.Ctx = &arenaCtx{task: t, cfg: t.cfg}

	runOpts := statemachine.RunOptions{
		Interval: 500 * time.Millisecond,
		Label:    "王国竞技场",
	}
	if t.guard != nil {
		runOpts.Guard = t.guard.Check
	}
	return t.sm.Run(t.handlers(), runOpts)
}
