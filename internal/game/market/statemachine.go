package market

import (
	"errors"
	"time"

	"app/internal/logger"
	"app/internal/statemachine"
)

type marketCtx struct {
	task *Task
	cfg  *Config
}

// 状态流：detect → navigate → prepare → refresh → purchase → schedule → leave。
// 对齐 Lua 交易所_任务.lua Task.run：进场 → 确认道具交易所页 → 免费刷新/
// 冷却推迟（首轮强制除外）→ 扫货 → 按页面补货倒计时写调度 → 离场。
func (t *Task) handlers() map[string]statemachine.Handler {
	return map[string]statemachine.Handler{
		"detect": func(sm *statemachine.Machine) statemachine.Result {
			ctx := sm.Ctx.(*marketCtx)
			if ctx.task.page.IsCurrent() {
				return statemachine.Next("prepare")
			}
			if ctx.task.kingdomPage != nil && (ctx.task.kingdomPage.IsKingdomHome() || ctx.task.kingdomPage.IsEventPage()) {
				return statemachine.Next("navigate")
			}
			return statemachine.Fatal{Err: errors.New("无法识别当前页面")}
		},
		"navigate": func(sm *statemachine.Machine) statemachine.Result {
			ctx := sm.Ctx.(*marketCtx)
			if ctx.task.page.IsCurrent() {
				return statemachine.Next("prepare")
			}
			ctx.task.pushStatus("海滩交易所 进入中…")
			if ctx.task.route.Enter() {
				return statemachine.Next("prepare")
			}
			return statemachine.Keep{}
		},
		"prepare": func(sm *statemachine.Machine) statemachine.Result {
			ctx := sm.Ctx.(*marketCtx)
			if ctx.task.page.EnsureItemTab() {
				return statemachine.Next("refresh")
			}
			logger.Warnf("[Market] 未能确认道具交易所页")
			return statemachine.Retry{}
		},
		"refresh": func(sm *statemachine.Machine) statemachine.Result {
			ctx := sm.Ctx.(*marketCtx)
			ctx.task.pushStatus("海滩交易所 检查刷新…")
			forceFirstRun := ctx.task.state.ConsumeStartupBypass()
			if ctx.task.page.IsFreeRefresh() {
				logger.Infof("[Market] 可免费刷新，先刷新")
				ctx.task.page.TapRefresh()
				return statemachine.Next("purchase")
			}
			remain, raw, ok := ctx.task.page.ReadRestockSeconds()
			if ok && remain > 0 {
				if forceFirstRun {
					logger.Infof("[Market] 首轮强制扫货，忽略页面补货倒计时: %s", raw)
				} else {
					logger.Infof("[Market] 当前冷却中，以 OCR 为准推迟: %s", raw)
					ctx.task.state.ScheduleAfterRestock(time.Duration(remain)*time.Second, ctx.cfg.BufferSec())
					return statemachine.Next("leave")
				}
			}
			return statemachine.Next("purchase")
		},
		"purchase": func(sm *statemachine.Machine) statemachine.Result {
			ctx := sm.Ctx.(*marketCtx)
			items := ctx.cfg.Items
			if len(items) == 0 {
				items = ctx.task.page.StockKeys()
			}
			logger.Infof("[Market] 开始扫货: %v", items)
			ctx.task.pushStatus("海滩交易所 扫货中…")
			stats := ctx.task.page.PurchaseWishlist(items)
			ctx.task.state.Purchased = stats.Purchased
			ctx.task.state.SoldOut = stats.SoldOut
			ctx.task.state.Shortage = stats.Shortage
			ctx.task.state.Failed = stats.Failed
			logger.Infof("[Market] 扫货结束 purchased=%d soldOut=%d shortage=%d failed=%d",
				stats.Purchased, stats.SoldOut, stats.Shortage, stats.Failed)
			ctx.task.pushStatus(ctx.task.state.StatusText())
			return statemachine.Next("schedule")
		},
		"schedule": func(sm *statemachine.Machine) statemachine.Result {
			ctx := sm.Ctx.(*marketCtx)
			ctx.task.pushStatus("海滩交易所 读取补货…")
			restockSec, raw, ok := ctx.task.page.ReadRestockSeconds()
			switch {
			case ok && restockSec > 0:
				ctx.task.state.ScheduleAfterRestock(time.Duration(restockSec)*time.Second, ctx.cfg.BufferSec())
			case ok:
				logger.Infof("[Market] 刷新按钮仍为免费刷新，不写等待")
			default:
				logger.Warnf("[Market] 未能读取补货时间 raw=%q", raw)
			}
			ctx.task.pushStatus(ctx.task.state.StatusText()) // 收尾回到统计文案
			return statemachine.Next("leave")
		},
		"leave": func(sm *statemachine.Machine) statemachine.Result {
			ctx := sm.Ctx.(*marketCtx)
			if !ctx.task.route.Leave() {
				return statemachine.Fatal{Err: errors.New("离开交易所失败")}
			}
			return statemachine.Done{}
		},
	}
}
