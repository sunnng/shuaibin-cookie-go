package scheduler

import (
	"time"

	"app/internal/logger"
)

type TaskOpts struct {
	Name               string
	ConfigKey          string
	CheckEnabled       func() bool
	CanResume          func() bool
	CheckReady         func() (bool, time.Duration)
	WaitHUD            func(remain time.Duration) string
	OnNotReady         func(remain time.Duration)
	Precondition       func() bool
	OnPreconditionFail func()
	Prepare            func() error
	Action             func() error
}

func (s *Scheduler) Build(opts TaskOpts) {
	condition := func() bool {
		if opts.CheckEnabled != nil {
			if !opts.CheckEnabled() {
				return false
			}
		}

		if opts.Precondition != nil && !opts.Precondition() {
			logger.Infof("[TaskBuilder] %s precondition failed", opts.Name)
			if opts.OnPreconditionFail != nil {
				opts.OnPreconditionFail()
			}
			return false
		}

		if opts.CanResume != nil && opts.CanResume() {
			return true
		}

		if opts.CheckReady != nil {
			ready, remain := opts.CheckReady()
			if !ready {
				if opts.WaitHUD != nil && remain > 0 {
					logger.Infof("[TaskBuilder] %s waiting: %s", opts.Name, opts.WaitHUD(remain))
				}
				if opts.OnNotReady != nil {
					opts.OnNotReady(remain)
				}
				return false
			}
		}

		return true
	}

	action := func() error {
		if opts.Prepare != nil {
			if err := opts.Prepare(); err != nil {
				return err
			}
		}
		return opts.Action()
	}

	s.Add(opts.Name, condition, action)
}
