package square

import (
	"errors"
	"fmt"
	"time"

	"app/internal/logger"
	"app/internal/statemachine"
)

type squareCtx struct {
	task         *Task
	cfg          *Config
	finishReason string // 结束原因，对齐 Lua finishToday(reason)
}

// 状态流转对齐 Lua 广场_任务.lua：
//
//	start → accumulate（今日已初检）/ openDialog（未初检）
//	accumulate → openDialog（有效停留已满）/ Done（睡完一个 chunk，交还调度器）
//	openDialog → dialog
//	dialog → finish（满额）/ claim（达到 dailyCap）/ Done（未达标：初检+重置计时）
//	claim → finish
//	finish → Done（回王国主城并标记今日完成）
//
// 注意：一次 Run 最多睡一个 chunk 即返回（Lua run() 同语义），调度器是同步
// 循环，长阻塞会饿死其它任务；进度都在 State 里，下次 Run 从 start 重新推导。
func (t *Task) handlers() map[string]statemachine.Handler {
	return map[string]statemachine.Handler{
		"start": func(sm *statemachine.Machine) statemachine.Result {
			ctx := sm.Ctx.(*squareCtx)
			tk := ctx.task
			if tk.state.IsDoneToday() {
				logger.Infof("[Square] 今日已完成，跳过")
				return statemachine.Done{}
			}
			tk.state.Ensure()
			tk.pushStatus("执行中…")
			if tk.state.HasCheckedToday() {
				return statemachine.Next("accumulate")
			}
			return statemachine.Next("openDialog")
		},
		// accumulate = Lua waitAccumulationChunk
		"accumulate": func(sm *statemachine.Machine) statemachine.Result {
			ctx := sm.Ctx.(*squareCtx)
			tk := ctx.task
			if !tk.route.EnsureSquare() {
				tk.state.PauseStay()
				return statemachine.Fatal{Err: errors.New("无法进入广场")}
			}
			if tk.page.IsLeaveDialog() {
				tk.page.TapCloseDialog()
				tk.page.SleepMs(800)
				if !tk.page.IsSquare() {
					return statemachine.Fatal{Err: errors.New("关闭离开弹窗后未回到广场页")}
				}
			}
			tk.state.StartStay()
			required := ctx.cfg.staySec()
			remaining := tk.state.StayRemaining(required)
			if remaining <= 0 {
				logger.Infof("[Square] 有效停留已满 %v，打开离开弹窗检查", required)
				tk.pushStatus("检查奖励…")
				return statemachine.Next("openDialog")
			}
			tk.pushStatus(stayProgressText(tk.state, required))
			chunk := ctx.cfg.chunkSec()
			if remaining < chunk {
				chunk = remaining
			}
			tk.interruptibleSleep(chunk)
			return statemachine.Done{}
		},
		// openDialog = Lua openLeaveDialog
		"openDialog": func(sm *statemachine.Machine) statemachine.Result {
			ctx := sm.Ctx.(*squareCtx)
			tk := ctx.task
			if tk.page.IsLeaveDialog() {
				tk.state.StartStay()
				return statemachine.Next("dialog")
			}
			if !tk.route.EnsureSquare() {
				tk.state.PauseStay()
				return statemachine.Fatal{Err: errors.New("无法进入广场")}
			}
			tk.state.StartStay()
			if !tk.route.OpenLeaveDialog() {
				return statemachine.Fatal{Err: errors.New("离开广场弹窗未出现")}
			}
			return statemachine.Next("dialog")
		},
		// dialog = Lua handleLeaveDialog
		"dialog": func(sm *statemachine.Machine) statemachine.Result {
			ctx := sm.Ctx.(*squareCtx)
			tk := ctx.task
			if !tk.page.IsLeaveDialog() {
				logger.Warnf("[Square] handleLeaveDialog 调用时不在离开弹窗")
				return statemachine.Fatal{Err: errors.New("不在离开广场弹窗")}
			}
			tk.state.StartStay()
			tk.page.SleepMs(500)

			if tk.page.IsDailyRewardsMaxed() {
				ctx.finishReason = "满额标识=最大"
				return statemachine.Next("finish")
			}

			pending, total, sum, ok := tk.page.ReadRewardSum()
			if !ok {
				tk.page.SleepMs(1000)
				pending, total, sum, ok = tk.page.ReadRewardSum()
			}
			if !ok {
				logger.Warnf("[Square] 无法识别奖励数量")
				return statemachine.Fatal{Err: errors.New("无法识别奖励数量")}
			}

			cap := ctx.cfg.dailyCap()
			tk.pushStatus(fmt.Sprintf("%d+%d=%d / %d", pending, total, sum, cap))
			if sum >= cap {
				return statemachine.Next("claim")
			}

			logger.Infof("[Square] 未达领取条件 %d/%d，返回广场继续挂机", sum, cap)
			tk.page.TapCloseDialog()
			tk.page.SleepMs(800)
			tk.state.MarkCheckedToday()
			tk.state.ResetStayTimer()
			// Lua 此处递归进 waitAccumulationChunk 睡一个 chunk 再返回；
			// 交给下一轮 Run 走 accumulate，进度已在会话里。
			return statemachine.Done{}
		},
		// claim = Lua claimAndFinish
		"claim": func(sm *statemachine.Machine) statemachine.Result {
			ctx := sm.Ctx.(*squareCtx)
			tk := ctx.task
			logger.Infof("[Square] 奖励已达标，点击一次领回")
			tk.pushStatus("一次领回…")
			if !tk.page.IsLeaveDialog() && !tk.route.OpenLeaveDialog() {
				return statemachine.Fatal{Err: errors.New("无法打开离开弹窗领取奖励")}
			}
			tk.page.TapClaimAll()
			tk.page.SleepMs(1500)
			tk.page.TapUntilDialog()
			ctx.finishReason = "已领取奖励"
			return statemachine.Next("finish")
		},
		// finish = Lua finishToday
		"finish": func(sm *statemachine.Machine) statemachine.Result {
			ctx := sm.Ctx.(*squareCtx)
			tk := ctx.task
			logger.Infof("[Square] 今日广场任务结束: %s", ctx.finishReason)
			tk.pushStatus("今日已完成")
			tk.state.PauseStay()
			if !tk.route.LeaveToKingdom(30 * time.Second) {
				return statemachine.Fatal{Err: errors.New("离开广场失败")}
			}
			tk.state.MarkDoneToday()
			return statemachine.Done{}
		},
	}
}

// stayProgressText 对齐 Lua stayProgressText：有效停留进度或「可开弹窗查看奖励」。
func stayProgressText(s *State, required time.Duration) string {
	rem := s.StayRemaining(required)
	if rem <= 0 {
		return "可开弹窗查看奖励"
	}
	elapsed := max(0, required-rem)
	return fmt.Sprintf("有效停留 %ds/%ds", int64(elapsed.Seconds()), int64(required.Seconds()))
}
