package bot

import (
	"time"

	"app/internal/platform/action"
	"app/internal/utils"
)

type Machine struct {
	ctx              *Context
	registry         *Registry
	unknownCount     int
	recoveryAttempts int
	running          bool
}

func NewMachine(ctx *Context, registry *Registry) *Machine {
	return &Machine{
		ctx:      ctx,
		registry: registry,
	}
}

func (m *Machine) Run() error {
	m.running = true
	for m.running {
		start := time.Now()
		m.tick()
		elapsed := time.Since(start)
		sleepMs := m.ctx.Config.TickIntervalMs - int(elapsed.Milliseconds())
		if sleepMs > 0 {
			m.ctx.Executor.Sleep(sleepMs)
		}
	}
	return nil
}

func (m *Machine) tick() {
	ctx := m.ctx
	current := ctx.Current

	if current == nil {
		found := m.registry.Find(ctx)
		if found == nil {
			m.handleUnknown()
			return
		}
		m.transition(found)
		return
	}

	if !current.Detect(ctx) {
		m.unknownCount++
		if m.unknownCount >= ctx.Config.MaxUnknownRetries {
			m.handleUnknown()
		}
		return
	}
	m.unknownCount = 0

	if time.Since(ctx.EnteredAt).Seconds() > float64(ctx.Config.MaxStateDurationSec) {
		utils.Errorf("state %s stuck, triggering recovery", current.Name())
		m.recover(current)
		return
	}

	if err := current.Act(ctx); err != nil {
		utils.Errorf("state %s act error: %v", current.Name(), err)
		return
	}

	next := current.Next(ctx)
	if next != nil && next != current {
		m.transition(next)
	}
}

func (m *Machine) transition(s State) {
	fromName := "nil"
	if m.ctx.Current != nil {
		m.ctx.LastState = m.ctx.Current
		fromName = m.ctx.Current.Name()
	}
	utils.PrintStateTransition(fromName, s.Name())
	m.ctx.Current = s
	m.ctx.EnteredAt = time.Now()
	m.ctx.RetryCount = 0
	m.recoveryAttempts = 0
}

func (m *Machine) handleUnknown() {
	ctx := m.ctx
	found := m.registry.Find(ctx)
	if found != nil {
		m.transition(found)
		return
	}

	m.unknownCount = 0
	m.recoveryAttempts++
	if m.recoveryAttempts > ctx.Config.MaxRecoveryAttempts {
		utils.Errorf("max recovery attempts exceeded, entering low power wait")
		ctx.Executor.Sleep(ctx.Config.LowPowerWaitSec * 1000)
		m.recoveryAttempts = 0
		return
	}

	if ctx.Current != nil {
		m.recover(ctx.Current)
	} else {
		_ = ctx.Executor.Home()
		ctx.Executor.Sleep(1000)
	}
}

func (m *Machine) recover(s State) {
	utils.Infof("recovering from state %s", s.Name())
	if err := s.Recover(m.ctx); err == nil {
		return
	}
	action.TapBackMultiple(m.ctx.Executor, 3)
	_ = m.ctx.Executor.Home()
	m.ctx.Executor.Sleep(1000)
}
