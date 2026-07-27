package biscuit

import (
	"app/internal/logger"
	"app/internal/statemachine"
)

// biscuitCtx 单次运行的上下文。resetHandled/sameHandled 对齐 Lua 的
// isConfirmResetDialog/isConfirmSameDialog：每个弹窗本轮只处理一次
// （点过"今日不再显示"后当天不会再弹）。
type biscuitCtx struct {
	task         *Task
	cfg          *Config
	resetHandled bool
	sameHandled  bool
}

// 洗脆饼流程是单状态循环（对齐 Lua task.lua 的 while 循环）：
// 读词条 → 毕业判定 → 未毕业则点洗炼并处理两个确认弹窗，直到毕业或达上限。
func (t *Task) handlers() map[string]statemachine.Handler {
	return map[string]statemachine.Handler{
		"roll": func(sm *statemachine.Machine) statemachine.Result {
			ctx := sm.Ctx.(*biscuitCtx)
			task := ctx.task
			cfg := ctx.cfg

			if task.state.IsGraduated() {
				task.pushStatus()
				return statemachine.Done{}
			}
			if cfg.MaxRolls <= 0 || task.state.Rolls >= cfg.MaxRolls {
				logger.Infof("[洗脆饼] 已达上限 %d 次，未毕业", task.state.Rolls)
				task.pushStatus()
				return statemachine.Done{}
			}

			task.state.Rolls++
			task.pushStatus()

			effects := task.page.ReadEffects()
			if ok, msg := check(effects, cfg.Targets, cfg.SumRules); ok {
				task.state.MarkGraduated()
				cfg.Enabled = false // 对齐 Lua UserConfig.set("biscuit", {enabled=false})
				logger.Infof("[洗脆饼] %s", msg)
				task.pushStatus()
				return statemachine.Done{}
			}

			task.page.TapReroll()
			if !ctx.resetHandled && task.page.ConfirmResetDialog() {
				ctx.resetHandled = true
			}
			if !ctx.sameHandled && task.page.ConfirmSameDialog() {
				ctx.sameHandled = true
			}
			return statemachine.Keep{}
		},
	}
}
