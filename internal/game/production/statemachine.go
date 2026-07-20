package production

import (
	"app/internal/statemachine"
)

type productionCtx struct {
	task *Task
}

// handlers 任务流程步骤。骨架：detect 直接 Done，步骤名预留后续填充。
func (t *Task) handlers() map[string]statemachine.Handler {
	return map[string]statemachine.Handler{
		"detect": func(sm *statemachine.Machine) statemachine.Result {
			// TODO: 识别王国首页 / 生产界面 → navigate | collect
			return statemachine.Done{}
		},
		"navigate": func(sm *statemachine.Machine) statemachine.Result {
			ctx := sm.Ctx.(*productionCtx)
			if ctx.task.page.IsBoard() {
				return statemachine.Next("collect")
			}
			if ctx.task.route.Enter() {
				return statemachine.Next("collect")
			}
			return statemachine.Keep{}
		},
		"collect": func(sm *statemachine.Machine) statemachine.Result {
			ctx := sm.Ctx.(*productionCtx)
			if ctx.task.page.TapCollectAll() && ctx.task.state != nil {
				ctx.task.state.Collected++
				ctx.task.pushStatus()
			}
			return statemachine.Next("leave")
		},
		"leave": func(sm *statemachine.Machine) statemachine.Result {
			ctx := sm.Ctx.(*productionCtx)
			if ctx.task.route.Leave() {
				return statemachine.Done{}
			}
			return statemachine.Keep{}
		},
	}
}
