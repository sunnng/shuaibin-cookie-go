package ui

import (
	"sync"

	"app/internal/logger"
)

type ScriptState int

const (
	StateIdle ScriptState = iota
	StateRunning
	StatePaused
)

// SessionHooks.OnStart 返回：阻塞的 run（在 goroutine 中调用）、以及 pause/resume/stop 钩子。
type SessionHooks struct {
	OnStart func() (run func() error, pause, resume, stop func())
	OnExit  func()
}

type SessionController struct {
	mu     sync.Mutex
	state  ScriptState
	hooks  SessionHooks
	pause  func()
	resume func()
	stop   func()
	gen    int
}

func NewSessionController(hooks SessionHooks) *SessionController {
	return &SessionController{hooks: hooks}
}

func (c *SessionController) State() ScriptState {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.state
}

func (c *SessionController) Start() {
	c.mu.Lock()
	if c.state != StateIdle {
		c.mu.Unlock()
		return
	}
	if c.hooks.OnStart == nil {
		c.mu.Unlock()
		return
	}
	run, pause, resume, stop := c.hooks.OnStart()
	c.pause, c.resume, c.stop = pause, resume, stop
	c.state = StateRunning
	c.gen++
	myGen := c.gen
	c.mu.Unlock()

	go func() {
		var err error
		if run != nil {
			err = run()
		}
		if err != nil {
			logger.Errorf("script run error: %v", err)
		}
		c.mu.Lock()
		if c.gen == myGen {
			c.state = StateIdle
			c.pause, c.resume, c.stop = nil, nil, nil
		}
		c.mu.Unlock()
	}()
}

func (c *SessionController) Pause() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.state != StateRunning {
		return
	}
	if c.pause != nil {
		c.pause()
	}
	c.state = StatePaused
}

func (c *SessionController) Resume() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.state != StatePaused {
		return
	}
	if c.resume != nil {
		c.resume()
	}
	c.state = StateRunning
}

func (c *SessionController) Stop() {
	c.mu.Lock()
	stop := c.stop
	if c.state == StateIdle {
		c.mu.Unlock()
		return
	}
	c.mu.Unlock()
	if stop != nil {
		stop()
	}
	// Do not flip to Idle here: wait for the run goroutine to exit so a
	// subsequent Start cannot spawn a second runtime while the old one still
	// owns the device (e.g. mid state-machine tick).
}

func (c *SessionController) Exit() {
	c.Stop()
	if c.hooks.OnExit != nil {
		c.hooks.OnExit()
	}
}
