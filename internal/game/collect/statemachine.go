package collect

import (
	"app/internal/config"
	"app/internal/logger"
	"app/internal/statemachine"
)

type collectCtx struct {
	task *Task
	cfg  *config.Collect
}

func (t *Task) handlers() map[string]statemachine.Handler {
	return map[string]statemachine.Handler{
		"detect": func(sm *statemachine.Machine) statemachine.Result {
			logger.Infof("[Collect] skeleton detect → done (implement page flow here)")
			return statemachine.Done{}
		},
	}
}
