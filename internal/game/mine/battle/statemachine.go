package battle

import (
	"errors"
	"time"

	"app/internal/logger"
	"app/internal/platform/screen"
	"app/internal/statemachine"
)

type battleCtx struct {
	task *Task
	cfg  *Config

	quickBattlePoint *screen.Point
}

func (t *Task) handlers() map[string]statemachine.Handler {
	return map[string]statemachine.Handler{
		"detect": func(sm *statemachine.Machine) statemachine.Result {
			ctx := sm.Ctx.(*battleCtx)
			if ctx.task.page.IsBattlePage() {
				return statemachine.Next("battleLoop")
			}
			if ctx.task.home.IsCurrent() || ctx.task.kingdom.IsKingdomHome() {
				return statemachine.Next("navigate")
			}
			return statemachine.Fatal{Err: errors.New("矿山战斗[detect] 页面识别失败")}
		},
		"navigate": func(sm *statemachine.Machine) statemachine.Result {
			ctx := sm.Ctx.(*battleCtx)
			if ctx.task.kingdom.IsKingdomHome() {
				logger.Infof("[矿山战斗] [navigate] 王国首页 → 矿山首页")
				if !ctx.task.route.KingdomHomeToMineHome() {
					return statemachine.Retry{}
				}
				return statemachine.Keep{}
			}
			if ctx.task.home.IsCurrent() {
				logger.Infof("[矿山战斗] [navigate] 矿山首页 → 战斗页")
				ctx.task.home.TapBattle()
				if ctx.task.page.WaitBattlePage(30 * time.Second) {
					return statemachine.Next("battleLoop")
				}
				return statemachine.Retry{}
			}
			if ctx.task.page.IsBattlePage() {
				return statemachine.Next("battleLoop")
			}
			return statemachine.Fatal{Err: errors.New("矿山战斗[navigate] 当前页面未知")}
		},
		"battleLoop":  battleLoop,
		"quickBattle": quickBattle,
		"exit":        exitTask,
	}
}

func battleLoop(sm *statemachine.Machine) statemachine.Result {
	ctx := sm.Ctx.(*battleCtx)
	if !ctx.task.page.IsBattlePage() {
		logger.Warnf("[矿山战斗] [battleLoop] 当前不在战斗页")
		return statemachine.Retry{}
	}

	targets := resolveTargetSoulStones(ctx.cfg)

	// 1. 优先处理当前页可见的快转
	if pt, ok := ctx.task.page.FindQuickBattleButton(); ok {
		logger.Infof("[矿山战斗] [battleLoop] 发现快转按钮")
		if name := ctx.task.page.RecognizeSoulStoneType(targets); name != "" {
			logger.Infof("[矿山战斗] [battleLoop] 快转灵魂石匹配: %s", name)
			pt := pt
			ctx.quickBattlePoint = &pt
			return statemachine.Next("quickBattle")
		}
		logger.Infof("[矿山战斗] [battleLoop] 快转灵魂石不匹配，扫描战斗卡")
		return scanAndIterateCards(sm, targets)
	}

	// 2. 无快转按钮，扫描战斗卡
	logger.Infof("[矿山战斗] [battleLoop] 无快转按钮，扫描战斗卡")
	return scanAndIterateCards(sm, targets)
}

// scanAndIterateCards 战斗卡扫描与迭代（Lua scanAndIterateCards）。
func scanAndIterateCards(sm *statemachine.Machine, targets map[string]bool) statemachine.Result {
	ctx := sm.Ctx.(*battleCtx)
	cards := ctx.task.page.FindBattleCards()
	logger.Infof("[矿山战斗] [battleLoop] 战斗卡数量=%d", len(cards))

	if len(cards) == 1 {
		logger.Infof("[矿山战斗] [battleLoop] 仅1张战斗卡，退出")
		return statemachine.Next("exit")
	}

	if len(cards) > 1 {
		for i := 1; i < len(cards); i++ {
			card := cards[i]
			ctx.task.page.TapBattleCard(card)
			if name := ctx.task.page.RecognizeSoulStoneType(targets); name != "" {
				logger.Infof("[矿山战斗] [battleLoop] 灵魂石匹配: %s", name)
				if pt, ok := ctx.task.page.FindQuickBattleButton(); ok {
					pt := pt
					ctx.quickBattlePoint = &pt
					return statemachine.Next("quickBattle")
				}
				logger.Warnf("[矿山战斗] [battleLoop] 灵魂石匹配但快转按钮消失，继续迭代")
			}
		}
	}

	if len(cards) >= 5 {
		logger.Infof("[矿山战斗] [battleLoop] 战斗卡≥5，执行翻页检查")
		if ctx.task.page.SwipeUpAndCheckLastPage() {
			logger.Infof("[矿山战斗] [battleLoop] 已到末页，退出")
			return statemachine.Next("exit")
		}
		logger.Infof("[矿山战斗] [battleLoop] 未到末页，重新扫描战斗卡")
		return statemachine.Next("battleLoop")
	}

	logger.Infof("[矿山战斗] [battleLoop] 战斗卡<5且无可操作项，退出")
	return statemachine.Next("exit")
}

func quickBattle(sm *statemachine.Machine) statemachine.Result {
	ctx := sm.Ctx.(*battleCtx)
	if ctx.quickBattlePoint == nil {
		logger.Warnf("[矿山战斗] [quickBattle] 缺少快转按钮坐标")
		return statemachine.Next("exit")
	}

	ctx.task.page.TapQuickBattleButton(*ctx.quickBattlePoint)
	if !ctx.task.page.WaitQuickBattleDialog(10 * time.Second) {
		logger.Warnf("[矿山战斗] [quickBattle] 快转弹窗未出现")
		return statemachine.Retry{}
	}

	used, owned, ok := ctx.task.page.ReadClockCount()
	logger.Infof("[矿山战斗] [quickBattle] 发条 used=%d owned=%d ok=%v", used, owned, ok)
	if !ok {
		logger.Warnf("[矿山战斗] [quickBattle] 发条数量读取失败，取消快转")
		ctx.task.page.TapQuickBattleCancel()
		ctx.task.page.WaitQuickBattleDialogGone(5 * time.Second)
		return statemachine.Next("exit")
	}
	if used > owned {
		logger.Infof("[矿山战斗] [quickBattle] 发条不足，取消快转")
		ctx.task.page.TapQuickBattleCancel()
		ctx.task.page.WaitQuickBattleDialogGone(5 * time.Second)
		return statemachine.Next("exit")
	}

	logger.Infof("[矿山战斗] [quickBattle] 发条充足，确认快转")
	ctx.task.page.TapQuickBattleConfirm()
	if ctx.task.page.TapSettleUntilBattlePage() {
		logger.Infof("[矿山战斗] [quickBattle] 快转结算完成 → battleLoop")
		return statemachine.Next("battleLoop")
	}
	logger.Warnf("[矿山战斗] [quickBattle] 结算后未回到战斗页")
	return statemachine.Retry{}
}

func exitTask(sm *statemachine.Machine) statemachine.Result {
	ctx := sm.Ctx.(*battleCtx)
	if ctx.task.page.IsBattlePage() {
		ctx.task.page.TapBackBtn()
		if !ctx.task.home.WaitCurrent(30 * time.Second) {
			logger.Warnf("[矿山战斗] [exit] 返回矿山首页超时")
			return statemachine.Retry{}
		}
	}

	if ctx.task.home.IsCurrent() {
		ctx.task.home.TapBack()
		if !ctx.task.kingdom.WaitHome(30 * time.Second) {
			logger.Warnf("[矿山战斗] [exit] 返回王国首页超时")
			return statemachine.Retry{}
		}
	}

	if ctx.task.kingdom.IsKingdomHome() {
		logger.Infof("[矿山战斗] [exit] 已回到王国首页")
		return statemachine.Done{}
	}

	logger.Warnf("[矿山战斗] [exit] 退出链路失败")
	return statemachine.Retry{}
}
