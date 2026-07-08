package bot

import (
	"image"
	"time"

	"app/internal/action"
	"app/internal/screen"
	"app/internal/utils"
)

type State interface {
	Name() string
	Detect(ctx *Context) bool
	Act(ctx *Context) error
	Next(ctx *Context) State
	Recover(ctx *Context) error
}

type Context struct {
	Config     *Config
	Detector   screen.Detector
	Executor   action.Executor
	Current    State
	LastState  State
	EnteredAt  time.Time
	RetryCount int
	Screenshot *image.NRGBA
}

func (c *Context) ResetRetry() {
	c.RetryCount = 0
}

func (c *Context) Log(format string, args ...any) {
	utils.Infof(format, args...)
}
