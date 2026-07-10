package scheduler

import (
	"time"

	"app/internal/logger"
)

type TaskOpts struct {
	Name         string
	CheckEnabled func() bool
	CheckReady   func() (bool, time.Duration)
	WaitHUD      func(remain time.Duration) string
	Action       func() error
}

func (s *Scheduler) Build(opts TaskOpts) {
	if opts.CheckReady != nil {
		s.AddIdleProvider(opts.Name, func() (time.Duration, string) {
			if opts.CheckEnabled != nil && !opts.CheckEnabled() {
				return 0, ""
			}
			ready, remain := opts.CheckReady()
			if ready || remain <= 0 {
				return 0, ""
			}
			label := opts.Name
			if opts.WaitHUD != nil {
				if msg := opts.WaitHUD(remain); msg != "" {
					label = msg
				}
			}
			return remain, label
		})
	}

	condition := func() bool {
		if opts.CheckEnabled != nil && !opts.CheckEnabled() {
			return false
		}

		if opts.CheckReady != nil {
			ready, remain := opts.CheckReady()
			if !ready {
				if opts.WaitHUD != nil && remain > 0 {
					logger.Infof("[TaskBuilder] %s waiting: %s", opts.Name, opts.WaitHUD(remain))
				}
				return false
			}
		}

		return true
	}

	s.Add(opts.Name, condition, opts.Action)
}
