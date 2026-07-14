package statemachine

import (
	"errors"
	"fmt"
	"time"

	"app/internal/logger"
)

// ErrStopped is returned when RunOptions.ShouldStop becomes true.
var ErrStopped = errors.New("stopped")

type Result interface{ result() }

type Keep struct{}
type Retry struct{}
type Done struct{}
type Next string
type Fatal struct{ Err error }

func (Keep) result()  {}
func (Retry) result() {}
func (Done) result()  {}
func (Next) result()  {}
func (Fatal) result() {}

type Handler func(sm *Machine) Result

type Machine struct {
	Current string
	Ctx     any

	currentState  string
	retries       int
	errors        int
	maxRetry      int
	maxError      int
	timeout       time.Duration
	retryInterval time.Duration
	startTime     time.Time
	ticks         int
}

type Options struct {
	MaxRetry      int
	MaxError      int
	Timeout       time.Duration
	RetryInterval time.Duration
}

type RunOptions struct {
	Interval   time.Duration
	Guard      func() bool
	ShouldStop func() bool
	Label      string
}

func New() *Machine { return &Machine{} }

func (m *Machine) Init(firstState string, opts Options) {
	m.currentState = firstState
	m.Current = firstState
	m.retries = 0
	m.errors = 0
	m.maxRetry = opts.MaxRetry
	m.maxError = opts.MaxError
	m.timeout = opts.Timeout
	m.retryInterval = opts.RetryInterval
	m.startTime = time.Now()
	m.ticks = 0
}

func (m *Machine) Run(handlers map[string]Handler, runOpts RunOptions) error {
	interval := runOpts.Interval
	if interval <= 0 {
		interval = 500 * time.Millisecond
	}
	label := runOpts.Label
	if label == "" {
		label = "statemachine"
	}

	if m.startTime.IsZero() {
		return fmt.Errorf("statemachine [%s] not initialized", label)
	}

	logger.Infof("[StateMachine] [%s] start state=%s maxRetry=%d maxError=%d timeout=%v",
		label, m.currentState, m.maxRetry, m.maxError, m.timeout)

	for {
		if runOpts.ShouldStop != nil && runOpts.ShouldStop() {
			logger.Infof("[StateMachine] [%s] stopped at state=%s", label, m.currentState)
			return fmt.Errorf("statemachine [%s]: %w", label, ErrStopped)
		}

		m.ticks++
		if m.timeout > 0 && time.Since(m.startTime) > m.timeout {
			return fmt.Errorf("statemachine [%s] timeout after %v", label, m.timeout)
		}

		if runOpts.Guard != nil && runOpts.Guard() {
			if err := m.sleepInterruptible(interval, runOpts, false); err != nil {
				return err
			}
			continue
		}

		handler, ok := handlers[m.currentState]
		if !ok {
			return fmt.Errorf("statemachine [%s] unknown state: %s", label, m.currentState)
		}

		logger.Debugf("[StateMachine] [%s] tick#%d state=%s retry=%d err=%d",
			label, m.ticks, m.currentState, m.retries, m.errors)

		ret := handler(m)
		retried := false

		switch r := ret.(type) {
		case Done:
			logger.Infof("[StateMachine] [%s] done after %d ticks", label, m.ticks)
			m.Current = m.currentState
			return nil
		case Keep:
			// stay
		case Retry:
			m.retries++
			if m.retries > m.maxRetry {
				return fmt.Errorf("statemachine [%s] state %s retry exceeded (%d/%d)",
					label, m.currentState, m.retries, m.maxRetry)
			}
			logger.Infof("[StateMachine] [%s] retry %d/%d", label, m.retries, m.maxRetry)
			retried = true
		case Next:
			m.currentState = string(r)
			m.Current = m.currentState
			m.retries = 0
			m.errors = 0
			logger.Infof("[StateMachine] [%s] -> %s", label, m.currentState)
		case Fatal:
			return fmt.Errorf("statemachine [%s] fatal in state %s: %w", label, m.currentState, r.Err)
		default:
			m.errors++
			if m.errors > m.maxError {
				return fmt.Errorf("statemachine [%s] error limit exceeded", label)
			}
			logger.Warnf("[StateMachine] [%s] unexpected result type, error count %d/%d",
				label, m.errors, m.maxError)
			retried = true
		}

		sleep := interval
		if retried && m.retryInterval > 0 {
			sleep = m.retryInterval
		}

		if err := m.sleepInterruptible(sleep, runOpts, runOpts.Guard != nil); err != nil {
			return err
		}
	}
}

func (m *Machine) sleepInterruptible(d time.Duration, runOpts RunOptions, breakOnGuard bool) error {
	label := runOpts.Label
	if label == "" {
		label = "statemachine"
	}
	deadline := time.Now().Add(d)
	step := 500 * time.Millisecond
	if d > 0 && d < step {
		step = d
	}
	for time.Now().Before(deadline) {
		if runOpts.ShouldStop != nil && runOpts.ShouldStop() {
			logger.Infof("[StateMachine] [%s] stopped during sleep at state=%s", label, m.currentState)
			return fmt.Errorf("statemachine [%s]: %w", label, ErrStopped)
		}
		if breakOnGuard && runOpts.Guard != nil && runOpts.Guard() {
			return nil
		}
		remaining := time.Until(deadline)
		if remaining <= 0 {
			break
		}
		if remaining > step {
			remaining = step
		}
		time.Sleep(remaining)
	}
	return nil
}
