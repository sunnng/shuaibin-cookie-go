package collect

import (
	"time"

	"app/internal/config"
	"app/internal/platform/action"
	"app/internal/platform/screen"
	"app/internal/statemachine"
)

type Task struct {
	cfg        *config.Collect
	page       *Page
	route      *Route
	session    *Session
	sm         *statemachine.Machine
	shouldStop func() bool
}

func NewTask(
	cfg *config.Collect,
	det screen.Detector,
	exec action.Executor,
	feature *Feature,
	session *Session,
) *Task {
	page := NewPage(det, exec, feature)
	return &Task{
		cfg:     cfg,
		page:    page,
		route:   NewRoute(page),
		session: session,
		sm:      statemachine.New(),
	}
}

func (t *Task) SetShouldStop(fn func() bool) {
	t.shouldStop = fn
}

func (t *Task) Run() error {
	return t.runWithOptions(statemachine.RunOptions{
		Interval: 200 * time.Millisecond,
		Label:    "收集",
	})
}

func (t *Task) runWithOptions(opts statemachine.RunOptions) error {
	t.sm.Init("detect", statemachine.Options{
		MaxRetry:      3,
		MaxError:      3,
		Timeout:       5 * time.Minute,
		RetryInterval: 1 * time.Second,
	})
	t.sm.Ctx = &collectCtx{task: t, cfg: t.cfg}
	if opts.Label == "" {
		opts.Label = "收集"
	}
	if opts.ShouldStop == nil && t.shouldStop != nil {
		opts.ShouldStop = t.shouldStop
	}
	return t.sm.Run(t.handlers(), opts)
}
