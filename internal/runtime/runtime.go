package runtime

import (
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
	registerFn func()
	stopCh     chan struct{}
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
		cfg:       cfg,
		stopCh:    make(chan struct{}),
	}
}

func (r *Runtime) Register(fn func()) {
	r.registerFn = fn
}

func (r *Runtime) Stop() {
	close(r.stopCh)
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
			wait, label := r.scheduler.MaxIdleWait()
			if wait > 0 {
				if r.hud != nil {
					r.hud.SetWait(label)
				}
				logger.Infof("[Runtime] idle wait %v | %s", wait, label)
				r.sleepWithGuard(minDuration(wait, r.cfg.IdleDelay))
			} else {
				if r.hud != nil {
					r.hud.SetIdle()
				}
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
	if r.guard != nil {
		r.guard.Sleep(d)
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
