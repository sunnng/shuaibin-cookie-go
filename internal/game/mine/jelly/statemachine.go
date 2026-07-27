package jelly

import (
	"errors"
	"time"

	"app/internal/logger"
	"app/internal/statemachine"
)

type jellyCtx struct {
	task *Task
	cfg  *Config

	remain time.Duration // readRemainTime 识别到的剩余时间（<=0 表示未识别）
}

func (t *Task) handlers() map[string]statemachine.Handler {
	return map[string]statemachine.Handler{
		"detect": func(sm *statemachine.Machine) statemachine.Result {
			ctx := sm.Ctx.(*jellyCtx)
			if ctx.task.page.IsJellyPage() {
				return statemachine.Next("processPage")
			}
			if ctx.task.home.IsCurrent() {
				return statemachine.Next("enterJelly")
			}
			if ctx.task.kingdom.IsKingdomHome() {
				return statemachine.Next("navigate")
			}
			return statemachine.Fatal{Err: errors.New("解除洋菜冻[detect] 页面识别失败")}
		},
		"navigate": func(sm *statemachine.Machine) statemachine.Result {
			ctx := sm.Ctx.(*jellyCtx)
			logger.Infof("[解除洋菜冻] [navigate] 王国首页 → 矿山首页")
			if ctx.task.route.KingdomHomeToMineHome() {
				return statemachine.Next("enterJelly")
			}
			return statemachine.Fatal{Err: errors.New("解除洋菜冻[navigate] 导航失败")}
		},
		"enterJelly": func(sm *statemachine.Machine) statemachine.Result {
			ctx := sm.Ctx.(*jellyCtx)
			logger.Infof("[解除洋菜冻] [enterJelly] 矿山首页 → 解除洋菜冻页面")
			ctx.task.home.TapJellyEntry()
			if ctx.task.page.WaitJellyPage(30 * time.Second) {
				return statemachine.Next("processPage")
			}
			return statemachine.Retry{}
		},
		"processPage": func(sm *statemachine.Machine) statemachine.Result {
			ctx := sm.Ctx.(*jellyCtx)

			// 1. 可全部领取则先领取并结算
			if ctx.task.page.CanClaimAll() {
				logger.Infof("[解除洋菜冻] [processPage] 检测到可全部领取")
				ctx.task.page.TapClaimAll()
				if !ctx.task.page.TapSettleUntilJellyPage() {
					logger.Warnf("[解除洋菜冻] [processPage] 点击 settleBtn 后页面未恢复")
					return statemachine.Retry{}
				}
			}

			// 2. OCR 查找「配置」按钮
			if pt, ok := ctx.task.page.FindConfigBtn(); ok {
				logger.Infof("[解除洋菜冻] [processPage] 找到配置按钮，进入配置界面")
				ctx.task.page.TapConfigBtn(pt)
				if !ctx.task.page.WaitConfigPage(2 * time.Second) {
					logger.Infof("[解除洋菜冻] [processPage] 未进入配置页，无可选择洋菜冻，结束任务")
					ctx.remain = 0
					return statemachine.Next("returnHome")
				}
				return statemachine.Next("configJelly")
			}

			// 3. 无配置按钮：识别剩余时间
			logger.Infof("[解除洋菜冻] [processPage] 未找到配置按钮，准备识别剩余时间")
			remain, ok := ctx.task.page.ReadRemainTime()
			if !ok {
				ctx.remain = 0
			} else {
				ctx.remain = remain
			}
			return statemachine.Next("returnHome")
		},
		"configJelly": func(sm *statemachine.Machine) statemachine.Result {
			ctx := sm.Ctx.(*jellyCtx)
			if ctx.task.page.CanChoose() {
				logger.Infof("[解除洋菜冻] [configJelly] 可选择，点击选择按钮")
				ctx.task.page.TapChoose()
				if ctx.task.page.WaitJellyPage(30 * time.Second) {
					return statemachine.Next("processPage")
				}
				return statemachine.Retry{}
			}

			// 不可选择：返回解除洋菜冻页面，再走返回链结束
			logger.Infof("[解除洋菜冻] [configJelly] 不可选择，返回解除洋菜冻页面")
			ctx.remain = 0
			ctx.task.page.TapConfigBack()
			if !ctx.task.page.WaitJellyPage(30 * time.Second) {
				return statemachine.Retry{}
			}
			return statemachine.Next("returnHome")
		},
		"returnHome": func(sm *statemachine.Machine) statemachine.Result {
			ctx := sm.Ctx.(*jellyCtx)
			logger.Infof("[解除洋菜冻] [returnHome] 返回王国首页")

			// 统一记录冷却，避免任务立即被再次调度
			if ctx.remain > 0 {
				ctx.task.state.EnterWait(ctx.remain)
			} else {
				ctx.task.state.EnterWait(ctx.task.intervalOrDefault())
			}

			if ctx.task.page.IsJellyPage() {
				ctx.task.page.TapBack()
				if !ctx.task.home.WaitCurrent(30 * time.Second) {
					logger.Warnf("[解除洋菜冻] [returnHome] 返回矿山首页超时")
					return statemachine.Retry{}
				}
			}

			if ctx.task.home.IsCurrent() {
				ctx.task.home.TapBack()
				if !ctx.task.kingdom.WaitHome(30 * time.Second) {
					logger.Warnf("[解除洋菜冻] [returnHome] 返回王国首页超时")
					return statemachine.Retry{}
				}
			}

			if ctx.task.kingdom.IsKingdomHome() {
				logger.Infof("[解除洋菜冻] [returnHome] 已回到王国首页")
				return statemachine.Done{}
			}

			logger.Warnf("[解除洋菜冻] [returnHome] 未知页面，无法返回")
			return statemachine.Retry{}
		},
	}
}

// intervalOrDefault 冷却间隔（Lua resolveWaitSec：配置无效时默认 3600）。
func (t *Task) intervalOrDefault() time.Duration {
	sec := t.cfg.IntervalSec
	if sec <= 0 {
		sec = 3600
	}
	return time.Duration(sec) * time.Second
}
