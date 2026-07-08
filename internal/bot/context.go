package bot

import (
	"image"
	"time"

	"app/internal/platform/action"
	"app/internal/config"
	"app/internal/platform/screen"
	"app/internal/utils"
)

type Context struct {
	Config     *config.Config
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
