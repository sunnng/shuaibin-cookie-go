package runtime

import (
	"sync"
	"time"

	"app/internal/guard"
	"app/internal/hud"
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
	HUD        *hud.HUD
	Logger     *logger.Logger
	RuntimeCfg RuntimeConfig
}

type Runtime struct {
	scheduler  *scheduler.Scheduler
	guard      *guard.Guard
	hud        *hud.HUD
	logger     *logger.Logger
	registerFn func()
	stopCh     chan struct{}
	stopOnce   sync.Once
	cfg        RuntimeConfig
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
		hud:       opts.HUD,
		logger:    opts.Logger,
		cfg:       cfg,
		stopCh:    make(chan struct{}),
	}
}

func (r *Runtime) Register(fn func()) {
	r.registerFn = fn
}

func (r *Runtime) Stop() {
	r.stopOnce.Do(func() { close(r.stopCh) })
}

func (r *Runtime) Run() error {
	if r.scheduler == nil {
		return nil
	}
	r.scheduler.Clear()
	if r.hud != nil {
		r.hud.SetTask("runtime", "running")
	}
	if r.registerFn != nil {
		r.registerFn()
	}

	r.infof("[Runtime] start guardInterval=%v idleDelay=%v stepDelay=%v stopOnError=%v",
		r.cfg.GuardInterval, r.cfg.IdleDelay, r.cfg.StepDelay, r.cfg.StopOnError)

	round := 0
	for {
		select {
		case <-r.stopCh:
			r.infof("[Runtime] stopped")
			return nil
		default:
		}

		round++
		r.debugf("[Runtime] round #%d", round)

		if r.guard != nil {
			r.guard.Check()
		}

		hasWork, err := r.scheduler.Run(r.cfg.StopOnError)
		if err != nil {
			r.errorf("[Runtime] scheduler error: %v", err)
			return err
		}

		if !hasWork {
			wait, label := r.scheduler.MaxIdleWait()
			if wait > 0 {
				if r.hud != nil {
					r.hud.SetWait(label)
				}
				r.infof("[Runtime] idle wait %v | %s", wait, label)
				r.sleepWithGuard(minDuration(wait, r.cfg.IdleDelay))
			} else {
				if r.hud != nil {
					r.hud.SetIdle()
				}
				r.infof("[Runtime] idle sleep %v", r.cfg.IdleDelay)
				r.sleepWithGuard(r.cfg.IdleDelay)
			}
		} else {
			r.debugf("[Runtime] step sleep %v", r.cfg.StepDelay)
			r.sleepWithGuard(r.cfg.StepDelay)
		}
	}
}

func (r *Runtime) infof(format string, args ...any) {
	if r.logger != nil {
		r.logger.Infof(format, args...)
	} else {
		logger.Infof(format, args...)
	}
}

func (r *Runtime) debugf(format string, args ...any) {
	if r.logger != nil {
		r.logger.Debugf(format, args...)
	} else {
		logger.Debugf(format, args...)
	}
}

func (r *Runtime) errorf(format string, args ...any) {
	if r.logger != nil {
		r.logger.Errorf(format, args...)
	} else {
		logger.Errorf(format, args...)
	}
}

func (r *Runtime) sleepWithGuard(d time.Duration) {
	if r.guard != nil {
		r.guard.SleepWithInterval(d, r.cfg.GuardInterval)
	} else {
		time.Sleep(d)
	}
}

func minDuration(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}
