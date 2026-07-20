package runtime

import (
	"sync"
	"time"

	"app/internal/guard"
	"app/internal/logger"
	"app/internal/scheduler"
)

type RuntimeConfig struct {
	GuardInterval time.Duration
	IdleDelay     time.Duration
	StepDelay     time.Duration
	StopOnError   bool
}

type Options struct {
	Scheduler  *scheduler.Scheduler
	Guard      *guard.Guard
	RuntimeCfg RuntimeConfig
}

type Runtime struct {
	scheduler *scheduler.Scheduler
	guard     *guard.Guard
	stopCh    chan struct{}
	stopOnce  sync.Once
	cfg       RuntimeConfig

	pauseMu sync.Mutex
	pauseCh chan struct{} // non-nil => paused; close to resume
}

func New(opts Options) *Runtime {
	cfg := opts.RuntimeCfg
	if cfg.GuardInterval <= 0 {
		cfg.GuardInterval = 500 * time.Millisecond
	}
	if cfg.IdleDelay <= 0 {
		cfg.IdleDelay = 30 * time.Second
	}
	if cfg.StepDelay <= 0 {
		cfg.StepDelay = 5 * time.Second
	}
	return &Runtime{
		scheduler: opts.Scheduler,
		guard:     opts.Guard,
		cfg:       cfg,
		stopCh:    make(chan struct{}),
	}
}

func (r *Runtime) Stop() {
	r.stopOnce.Do(func() { close(r.stopCh) })
}

// IsStopped reports whether Stop has been called.
func (r *Runtime) IsStopped() bool {
	select {
	case <-r.stopCh:
		return true
	default:
		return false
	}
}

func (r *Runtime) Pause() {
	r.pauseMu.Lock()
	defer r.pauseMu.Unlock()
	if r.pauseCh == nil {
		r.pauseCh = make(chan struct{})
	}
}

func (r *Runtime) Resume() {
	r.pauseMu.Lock()
	defer r.pauseMu.Unlock()
	if r.pauseCh != nil {
		close(r.pauseCh)
		r.pauseCh = nil
	}
}

func (r *Runtime) waitIfPaused() {
	for {
		r.pauseMu.Lock()
		ch := r.pauseCh
		r.pauseMu.Unlock()
		if ch == nil {
			return
		}
		select {
		case <-r.stopCh:
			return
		case <-ch:
			// resumed
		}
	}
}

func (r *Runtime) Run() error {
	if r.scheduler == nil {
		return nil
	}

	logger.Infof("[Runtime] start guardInterval=%v idleDelay=%v stepDelay=%v stopOnError=%v",
		r.cfg.GuardInterval, r.cfg.IdleDelay, r.cfg.StepDelay, r.cfg.StopOnError)

	round := 0
	for {
		select {
		case <-r.stopCh:
			logger.Infof("[Runtime] stopped")
			return nil
		default:
		}

		r.waitIfPaused()

		select {
		case <-r.stopCh:
			logger.Infof("[Runtime] stopped")
			return nil
		default:
		}

		round++
		logger.Debugf("[Runtime] round #%d", round)

		if r.guard != nil {
			r.guard.Check()
		}

		hasWork, err := r.scheduler.Run(r.cfg.StopOnError)
		if err != nil {
			logger.Errorf("[Runtime] scheduler error: %v", err)
			return err
		}

		if !hasWork {
			wait, label := r.scheduler.MinIdleWait()
			if wait > 0 {
				logger.Infof("[Runtime] idle wait %v | %s", wait, label)
				r.sleepWithGuard(minDuration(wait, r.cfg.IdleDelay))
			} else {
				logger.Infof("[Runtime] idle sleep %v", r.cfg.IdleDelay)
				r.sleepWithGuard(r.cfg.IdleDelay)
			}
		} else {
			logger.Debugf("[Runtime] step sleep %v", r.cfg.StepDelay)
			r.sleepWithGuard(r.cfg.StepDelay)
		}
	}
}

func (r *Runtime) sleepWithGuard(d time.Duration) {
	step := r.cfg.GuardInterval
	if step <= 0 {
		step = 100 * time.Millisecond
	}
	left := d
	for left > 0 {
		if r.guard != nil {
			r.guard.Check()
		}
		select {
		case <-r.stopCh:
			return
		default:
		}
		chunk := step
		if left < chunk {
			chunk = left
		}
		select {
		case <-r.stopCh:
			return
		case <-time.After(chunk):
		}
		left -= chunk
	}
}

func minDuration(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}
