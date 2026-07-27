package mining

import (
	"errors"
	"time"

	"app/internal/game/mine"
	"app/internal/logger"
	"app/internal/statemachine"
)

type miningCtx struct {
	task *Task
	cfg  *Config

	quotaCur           int
	quotaMax           int
	selectedCards      int
	cardSwipeDirection string
	skipSelectOnce     bool
}

func (t *Task) handlers() map[string]statemachine.Handler {
	return map[string]statemachine.Handler{
		"detect": func(sm *statemachine.Machine) statemachine.Result {
			ctx := sm.Ctx.(*miningCtx)
			if ctx.task.page.IsMiningPage() {
				return statemachine.Next("miningPageScan")
			}
			if ctx.task.home.IsCurrent() {
				return statemachine.Next("precheck")
			}
			if ctx.task.kingdom.IsKingdomHome() {
				return statemachine.Next("navigate")
			}
			return statemachine.Fatal{Err: errors.New("矿山开采[detect] 页面识别失败")}
		},
		"navigate": func(sm *statemachine.Machine) statemachine.Result {
			ctx := sm.Ctx.(*miningCtx)
			if ctx.task.route.KingdomHomeToMineHome() {
				return statemachine.Next("precheck")
			}
			return statemachine.Fatal{Err: errors.New("矿山开采[navigate] 王国→矿山首页失败")}
		},
		"precheck": func(sm *statemachine.Machine) statemachine.Result {
			ctx := sm.Ctx.(*miningCtx)
			if ctx.task.home.HasCompletedMiningTask() {
				logger.Infof("[矿山开采] [precheck] 首页发现存在已完成开采任务 → 进入开采页")
			}
			ctx.task.home.TapMining()
			if ctx.task.home.WaitGone(30 * time.Second) {
				if ctx.task.page.IsMiningPage() {
					return statemachine.Next("miningPageScan")
				}
				if ctx.task.page.IsSettlementRoute() {
					return statemachine.Next("confirmRewards")
				}
			}
			logger.Warnf("[矿山开采] [precheck] 未能进入开采页面")
			return statemachine.Retry{}
		},
		"miningPageScan": func(sm *statemachine.Machine) statemachine.Result {
			ctx := sm.Ctx.(*miningCtx)

			// 1. 已完成槽位 → 收奖励
			if ctx.task.page.TapCompletedSlot() {
				return statemachine.Next("confirmRewards")
			}

			// 2. 空闲槽位 → 选矿卡
			if ctx.skipSelectOnce {
				ctx.skipSelectOnce = false
				logger.Infof("[矿山开采] [miningPageScan] 无可用矿卡可填栏位，跳过选卡")
			} else if ctx.task.page.TapFreeSlot() {
				logger.Infof("[矿山开采] [miningPageScan] 有空闲栏位 → selectMineCard")
				return statemachine.Next("selectMineCard")
			} else if ctx.task.page.WasNoMineCard() {
				logger.Infof("[矿山开采] [miningPageScan] 矿脉卡清单无矿卡，准备回城结束")
				return statemachine.Next("noCardReturn")
			}

			// 3. 可开采槽位 → 开始开采
			if ctx.task.page.TapReadySlot() {
				logger.Infof("[矿山开采] [miningPageScan] 有可开始矿卡 → startMining")
				return statemachine.Next("startMining")
			}

			logger.Infof("[矿山开采] [miningPageScan] 无已完成/无空闲/无可开采栏位 → done")
			return statemachine.Next("done")
		},
		"confirmRewards": func(sm *statemachine.Machine) statemachine.Result {
			ctx := sm.Ctx.(*miningCtx)
			if ctx.task.page.TapUntilMatchMiningPage() {
				logger.Infof("[矿山开采] [confirmRewards] 奖励已确认 → miningPageScan")
				return statemachine.Next("miningPageScan")
			}
			logger.Warnf("[矿山开采] [confirmRewards] 确认奖励失败")
			return statemachine.Retry{}
		},
		"startMining": func(sm *statemachine.Machine) statemachine.Result {
			ctx := sm.Ctx.(*miningCtx)
			if ctx.task.page.AutoSelectCookieAndStart() {
				// 弹窗处理后仍在准备页说明开始失败，重试本状态
				if ctx.task.page.IsSetup() {
					return statemachine.Retry{}
				}
				return statemachine.Next("miningPageScan")
			}
			logger.Warnf("[矿山开采] [startMining] 开采矿卡失败")
			return statemachine.Retry{}
		},
		"selectMineCard": selectMineCard,
		"noCardReturn": func(sm *statemachine.Machine) statemachine.Result {
			ctx := sm.Ctx.(*miningCtx)
			logger.Infof("[矿山开采] [noCardReturn] 矿脉卡清单无矿卡，准备回城")
			if !ctx.task.route.ReturnToKingdom() {
				return statemachine.Fatal{Err: errors.New("矿山开采[noCardReturn] 回王国首页失败")}
			}
			ctx.task.enterBusyWait()
			return statemachine.Done{}
		},
		"done": func(sm *statemachine.Machine) statemachine.Result {
			ctx := sm.Ctx.(*miningCtx)
			// 结束前复查一次，避免识别抖动导致误判全忙。
			if ctx.task.page.HasCompletedTask() {
				logger.Infof("[矿山开采] [done] 复查发现已完成任务 → confirmRewards")
				ctx.task.page.TapCompletedSlot()
				return statemachine.Next("confirmRewards")
			}
			if ctx.task.page.HasFreeSlot() {
				logger.Infof("[矿山开采] [done] 复查发现空闲栏位 → selectMineCard")
				ctx.task.page.TapFreeSlot()
				return statemachine.Next("selectMineCard")
			}
			if ctx.task.page.HasStartableCard() {
				logger.Infof("[矿山开采] [done] 复查发现可开采矿卡 → startMining")
				ctx.task.page.TapReadySlot()
				return statemachine.Next("startMining")
			}

			logger.Infof("[矿山开采] [done] 当前页无可操作项，准备回城并记录 busy")
			if !ctx.task.route.ReturnToKingdom() {
				return statemachine.Fatal{Err: errors.New("矿山开采[done] 回王国首页失败")}
			}
			ctx.task.enterBusyWait()
			return statemachine.Done{}
		},
	}
}

// enterBusyWait 记录全忙等待（Lua MiningSession.enterBusyWait 默认读 miningIntervalSec）。
func (t *Task) enterBusyWait() {
	sec := t.cfg.IntervalSec
	if sec <= 0 {
		sec = 6 * 3600
	}
	t.state.EnterBusyWait(time.Duration(sec) * time.Second)
}

// selectMineCard 选矿卡（Lua selectMineCard）：按优先级扫卡列表填满配额后确认。
func selectMineCard(sm *statemachine.Machine) statemachine.Result {
	ctx := sm.Ctx.(*miningCtx)

	initCur, initMax, ok := ctx.task.page.ReadChooseQuota()
	if !ok {
		logger.Warnf("[矿山开采] [selectMineCard] OCR 可选数量失败")
		return statemachine.Retry{}
	}
	ctx.quotaCur = initCur
	ctx.quotaMax = initMax
	ctx.selectedCards = initCur

	if initCur < initMax {
		cardPriority := resolveCardPriority(ctx.cfg, ctx.task.feature.OreVeinCards)
		direction := ctx.cardSwipeDirection
		if direction == "" {
			direction = "left"
		}
		for _, cardKey := range cardPriority {
			cur, max, ok := ctx.task.page.ReadChooseQuota()
			if !ok {
				logger.Warnf("[矿山开采] [selectMineCard] 切换目标前 OCR 失败")
				return statemachine.Retry{}
			}
			if cur >= max {
				break
			}

			need := max - cur
			cardDef, configured := ctx.task.oreCardDef(cardKey)
			if !configured {
				continue
			}
			logger.Infof("[矿山开采] [selectMineCard] 目标矿卡 %s 方向:%s 还需 %d 张", cardKey, direction, need)

			got, exhausted := ctx.task.page.SelectTargetCards(cardDef, need, direction)
			cur, max, ok = ctx.task.page.ReadChooseQuota()
			if !ok {
				logger.Warnf("[矿山开采] [selectMineCard] 选卡后 OCR 失败")
				return statemachine.Retry{}
			}
			ctx.quotaCur = cur
			ctx.quotaMax = max
			ctx.selectedCards = cur

			if cur >= max {
				logger.Infof("[矿山开采] [selectMineCard] 已选满 %d/%d", cur, max)
				break
			}
			if exhausted || got == 0 {
				if exhausted {
					if direction == "left" {
						direction = "right"
					} else {
						direction = "left"
					}
					ctx.cardSwipeDirection = direction
				}
				logger.Infof("[矿山开采] [selectMineCard] %s 已扫完/无新增（+%d），切换下一种，方向:%s", cardKey, got, direction)
				continue
			}
			logger.Warnf("[矿山开采] [selectMineCard] 有新增但未填满，重试当前选卡流程")
			return statemachine.Retry{}
		}
	}

	finalCur, finalMax, ok := ctx.task.page.ReadChooseQuota()
	if !ok {
		logger.Warnf("[矿山开采] [selectMineCard] 最终 OCR 校验失败")
		return statemachine.Retry{}
	}
	ctx.quotaCur = finalCur
	ctx.quotaMax = finalMax
	ctx.selectedCards = finalCur

	if finalCur <= 0 {
		logger.Warnf("[矿山开采] [selectMineCard] 未选择任何矿卡，返回开采页")
		ctx.task.page.TapBackBtn()
		ctx.skipSelectOnce = true
		return statemachine.Next("miningPageScan")
	}

	if !ctx.task.page.ConfirmCardSelection() {
		logger.Warnf("[矿山开采] [selectMineCard] 确认选卡失败")
		if ctx.task.page.IsMiningPage() {
			return statemachine.Next("miningPageScan")
		}
		return statemachine.Retry{}
	}

	if !ctx.task.page.WaitMiningPage(30 * time.Second) {
		logger.Warnf("[矿山开采] [selectMineCard] 等待开采页超时")
		return statemachine.Retry{}
	}

	logger.Infof("[矿山开采] [selectMineCard] 已确认 → startMining")
	return statemachine.Next("startMining")
}

// oreCardDef 取矿卡特征；未配置（未取色）时 configured=false。
func (t *Task) oreCardDef(key string) (mine.ColorFind, bool) {
	if t.feature == nil {
		return mine.ColorFind{}, false
	}
	def, ok := t.feature.OreVeinCards[key]
	if !ok || def.Colors == "" {
		return mine.ColorFind{}, false
	}
	return def, true
}
