package ui

import (
	"sync"
)

type ScriptState int

const (
	StateIdle ScriptState = iota
	StateRunning
	StatePaused
)

// ScriptHooks.OnStart 返回：阻塞的 run（在 goroutine 中调用）、以及 pause/resume/stop 钩子。
type ScriptHooks struct {
	OnStart func() (run func() error, pause, resume, stop func())
	OnExit  func()
}

// LogErrorf 框架内部错误日志钩子（如脚本 run 返回错误）。默认丢弃；
// 应用可替换为自身日志器（如 main 中 ui.LogErrorf = logger.Errorf）。
// 包级变量非并发安全，应在启动早期赋值一次。
var LogErrorf = func(format string, args ...any) {}

type ScriptController struct {
	mu     sync.Mutex
	state  ScriptState
	hooks  ScriptHooks
	pause  func()
	resume func()
	stop   func()
	gen    int
}

func NewScriptController(hooks ScriptHooks) *ScriptController {
	return &ScriptController{hooks: hooks}
}

func (c *ScriptController) State() ScriptState {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.state
}

func (c *ScriptController) Start() {
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
			LogErrorf("script run error: %v", err)
		}
		c.mu.Lock()
		if c.gen == myGen {
			c.state = StateIdle
			c.pause, c.resume, c.stop = nil, nil, nil
		}
		c.mu.Unlock()
	}()
}

func (c *ScriptController) Pause() {
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

func (c *ScriptController) Resume() {
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

func (c *ScriptController) Stop() {
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

func (c *ScriptController) Exit() {
	c.Stop()
	if c.hooks.OnExit != nil {
		c.hooks.OnExit()
	}
}
