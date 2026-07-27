package starlight

import (
	"errors"
	"time"

	"app/internal/logger"
	"app/internal/statemachine"
)

// starlightCtx 状态机上下文。
type starlightCtx struct {
	task *Task
}

// handlers 对应 Lua 繁星岛_任务.lua 的流程：
// check → detect → navigate → openManual → enterIsland → returnFromIsland
// → openTask → claimTask → finish。
func (t *Task) handlers() map[string]statemachine.Handler {
	return map[string]statemachine.Handler{
		"check": func(sm *statemachine.Machine) statemachine.Result {
			ctx := sm.Ctx.(*starlightCtx)
			if ctx.task.state.IsDoneToday() {
				logger.Infof("[梦幻繁星岛任务] 今日已执行，跳过")
				return statemachine.Done{}
			}
			return statemachine.Next("detect")
		},
		"detect": func(sm *statemachine.Machine) statemachine.Result {
			ctx := sm.Ctx.(*starlightCtx)
			p := ctx.task.page
			switch {
			case p.IsHomePage():
				logger.Infof("[梦幻繁星岛任务] [detect] 当前在梦幻繁星岛首页")
				return statemachine.Next("openManual")
			case p.IsManualPage():
				logger.Infof("[梦幻繁星岛任务] [detect] 当前在航海手册页")
				return statemachine.Next("enterIsland")
			case p.IsVanillaIslandPage():
				logger.Infof("[梦幻繁星岛任务] [detect] 当前在纯香草小岛页")
				return statemachine.Next("returnFromIsland")
			case p.IsTaskPage():
				logger.Infof("[梦幻繁星岛任务] [detect] 当前在任务页")
				return statemachine.Next("claimTask")
			default:
				logger.Infof("[梦幻繁星岛任务] [detect] 不在已知页面，尝试导航")
				return statemachine.Next("navigate")
			}
		},
		"navigate": func(sm *statemachine.Machine) statemachine.Result {
			ctx := sm.Ctx.(*starlightCtx)
			ctx.task.pushStatus("导航到活动…")
			if ctx.task.route.EnsureHome() {
				return statemachine.Next("openManual")
			}
			logger.Warnf("[梦幻繁星岛任务] [navigate] 导航到梦幻繁星岛首页失败")
			return statemachine.Fatal{Err: errors.New("导航到梦幻繁星岛首页失败")}
		},
		"openManual": func(sm *statemachine.Machine) statemachine.Result {
			ctx := sm.Ctx.(*starlightCtx)
			ctx.task.pushStatus("打开航海手册…")
			if !ctx.task.page.TapSailingManual() {
				return statemachine.Retry{}
			}
			if ctx.task.page.WaitManualPage(10 * time.Second) {
				return statemachine.Next("enterIsland")
			}
			logger.Warnf("[梦幻繁星岛任务] [openManual] 等待航海手册页超时")
			return statemachine.Retry{}
		},
		"enterIsland": func(sm *statemachine.Machine) statemachine.Result {
			ctx := sm.Ctx.(*starlightCtx)
			ctx.task.pushStatus("进入纯香草小岛…")
			if !ctx.task.page.TapLoginIsland() {
				return statemachine.Retry{}
			}
			if ctx.task.page.WaitVanillaIslandPage(10 * time.Second) {
				return statemachine.Next("returnFromIsland")
			}
			logger.Warnf("[梦幻繁星岛任务] [enterIsland] 等待纯香草小岛页超时")
			return statemachine.Retry{}
		},
		"returnFromIsland": func(sm *statemachine.Machine) statemachine.Result {
			ctx := sm.Ctx.(*starlightCtx)
			ctx.task.pushStatus("返回首页…")
			if !ctx.task.page.TapBackFromVanilla() {
				return statemachine.Retry{}
			}
			if ctx.task.page.WaitHomePage(10 * time.Second) {
				return statemachine.Next("openTask")
			}
			logger.Warnf("[梦幻繁星岛任务] [returnFromIsland] 等待首页超时")
			return statemachine.Retry{}
		},
		"openTask": func(sm *statemachine.Machine) statemachine.Result {
			ctx := sm.Ctx.(*starlightCtx)
			ctx.task.pushStatus("打开任务页…")
			if !ctx.task.page.TapTaskBtn() {
				return statemachine.Retry{}
			}
			if ctx.task.page.WaitTaskPage(10 * time.Second) {
				return statemachine.Next("claimTask")
			}
			logger.Warnf("[梦幻繁星岛任务] [openTask] 等待任务页超时")
			return statemachine.Retry{}
		},
		"claimTask": func(sm *statemachine.Machine) statemachine.Result {
			ctx := sm.Ctx.(*starlightCtx)
			ctx.task.pushStatus("领取任务奖励…")
			p := ctx.task.page
			if pt, ok := p.FindClaimableBtn(); ok {
				logger.Infof("[梦幻繁星岛任务] [claimTask] 发现可领奖按钮 (%d,%d)", pt.X, pt.Y)
				p.TapClaimableBtn(pt)
				var check func() bool
				if ctx.task.guard != nil {
					check = ctx.task.guard.Check
				}
				p.SettleAfterClaim(check)
				p.DismissRewardPopupIfNeeded()
			} else {
				logger.Infof("[梦幻繁星岛任务] [claimTask] 无可领奖按钮")
			}
			ctx.task.state.MarkDoneToday()
			return statemachine.Next("finish")
		},
		"finish": func(sm *statemachine.Machine) statemachine.Result {
			ctx := sm.Ctx.(*starlightCtx)
			ctx.task.pushStatus("返回首页…")
			if !ctx.task.page.TapBackFromTask() {
				return statemachine.Retry{}
			}
			if !ctx.task.page.WaitHomePage(10 * time.Second) {
				logger.Warnf("[梦幻繁星岛任务] [finish] 等待首页超时")
				return statemachine.Retry{}
			}

			ctx.task.pushStatus("返回王国…")
			if !ctx.task.page.TapBackToKingdom() {
				return statemachine.Retry{}
			}
			if ctx.task.kingdom != nil && ctx.task.kingdom.WaitHome(10*time.Second) {
				logger.Infof("[梦幻繁星岛任务] 任务完成")
				return statemachine.Done{}
			}
			logger.Warnf("[梦幻繁星岛任务] [finish] 等待王国首页超时")
			return statemachine.Retry{}
		},
	}
}
