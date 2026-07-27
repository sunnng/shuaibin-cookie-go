package survey

import (
	"errors"
	"time"

	"app/internal/logger"
	"app/internal/statemachine"
)

type surveyCtx struct {
	task *Task
	cfg  *Config

	nextOcrPollAt time.Time // polling：下次 OCR 轮询点
	lastFloor     int
	lastGap       int
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func (t *Task) handlers() map[string]statemachine.Handler {
	return map[string]statemachine.Handler{
		"detect": func(sm *statemachine.Machine) statemachine.Result {
			ctx := sm.Ctx.(*surveyCtx)
			if ctx.task.page.IsDomain() {
				logger.Infof("[矿山勘查] [detect] 在勘查域 → prepare")
				return statemachine.Next("prepare")
			}
			if ctx.task.home.IsCurrent() || ctx.task.kingdom.IsKingdomHome() {
				logger.Infof("[矿山勘查] [detect] 在王国/矿山首页 → navigate")
				return statemachine.Next("navigate")
			}
			return statemachine.Fatal{Err: errors.New("矿山勘查[detect] 页面识别失败")}
		},
		"navigate": func(sm *statemachine.Machine) statemachine.Result {
			ctx := sm.Ctx.(*surveyCtx)
			switch {
			case ctx.task.kingdom.IsKingdomHome():
				logger.Infof("[矿山勘查] [navigate] 王国首页 → 矿山首页")
				ctx.task.route.KingdomHomeToMineHome()
			case ctx.task.home.IsCurrent():
				logger.Infof("[矿山勘查] [navigate] 矿山首页 → 进入勘查")
				ctx.task.home.TapVenture()
			case ctx.task.page.IsDomain():
				logger.Infof("[矿山勘查] [navigate] 已进入勘查域 → prepare")
				return statemachine.Next("prepare")
			}
			// 本步导航已执行但尚未到达目标页：保持 navigate，不计入重试（与 Lua 一致）
			return statemachine.Keep{}
		},
		"prepare": func(sm *statemachine.Machine) statemachine.Result {
			ctx := sm.Ctx.(*surveyCtx)
			if ctx.task.page.IsRunning() {
				logger.Infof("[矿山勘查] [prepare] 勘查进行中 → running")
				return statemachine.Next("running")
			}
			if ctx.task.page.Setup() {
				waitSec := time.Duration(ctx.cfg.FarWaitSec) * time.Second
				ctx.task.state.EnterFarWait(waitSec)
				logger.Infof("[矿山勘查] [prepare] setup 完成 → farWait %v", waitSec)
				return statemachine.Next("farWait")
			}
			return statemachine.Fatal{Err: errors.New("矿山勘查[prepare] setup 执行失败")}
		},
		"running": func(sm *statemachine.Machine) statemachine.Result {
			ctx := sm.Ctx.(*surveyCtx)
			currentFloor, ok := ctx.task.page.GetCurrentFloor()
			if !ok {
				logger.Warnf("[矿山勘查] [running] OCR 未识别层数，重试")
				return statemachine.Retry{}
			}
			cfg := ctx.cfg
			floorDiff := abs(cfg.TargetFloor - currentFloor)
			ctx.lastFloor = currentFloor
			ctx.lastGap = floorDiff
			logger.Infof("[矿山勘查] [running] 当前层:%d 目标:%d 阈值:%d 轮询:%ds 远距等待:%ds",
				currentFloor, cfg.TargetFloor, cfg.FarGap, cfg.OCRPollSec, cfg.FarWaitSec)

			if currentFloor >= cfg.TargetFloor {
				logger.Infof("[矿山勘查] [running] 已达标 → settle")
				return statemachine.Next("settle")
			}
			if floorDiff > cfg.FarGap {
				ctx.task.state.EnterFarWait(time.Duration(cfg.FarWaitSec) * time.Second)
				logger.Infof("[矿山勘查] [running] 远距(差%d>%d) → farWait，回城等待 %ds",
					floorDiff, cfg.FarGap, cfg.FarWaitSec)
				return statemachine.Next("farWait")
			}
			logger.Infof("[矿山勘查] [running] 近距(差%d<=%d) → polling", floorDiff, cfg.FarGap)
			return statemachine.Next("polling")
		},
		"polling": func(sm *statemachine.Machine) statemachine.Result {
			ctx := sm.Ctx.(*surveyCtx)
			cfg := ctx.cfg
			poll := time.Duration(cfg.OCRPollSec) * time.Second

			if ctx.nextOcrPollAt.IsZero() {
				ctx.nextOcrPollAt = time.Now().Add(poll)
				logger.Debugf("[矿山勘查] [polling] 首次进入，下次 OCR 在 %ds 后", cfg.OCRPollSec)
			}
			if time.Now().Before(ctx.nextOcrPollAt) {
				return statemachine.Keep{}
			}

			currentFloor, ok := ctx.task.page.GetCurrentFloor()
			ctx.nextOcrPollAt = time.Now().Add(poll)
			if !ok {
				logger.Warnf("[矿山勘查] [polling] OCR 未识别层数，重试")
				return statemachine.Retry{}
			}
			ctx.lastFloor = currentFloor
			ctx.lastGap = abs(cfg.TargetFloor - currentFloor)
			if currentFloor >= cfg.TargetFloor {
				logger.Infof("[矿山勘查] [polling] 达标 当前层:%d ≥ 目标:%d → settle",
					currentFloor, cfg.TargetFloor)
				return statemachine.Next("settle")
			}
			logger.Debugf("[矿山勘查] [polling] 当前层:%d 目标:%d，继续等待", currentFloor, cfg.TargetFloor)
			return statemachine.Keep{}
		},
		"settle": func(sm *statemachine.Machine) statemachine.Result {
			ctx := sm.Ctx.(*surveyCtx)
			if ctx.task.page.StopVenture() {
				logger.Infof("[矿山勘查] [settle] 结算完成 → detect（进入下一轮识别）")
				return statemachine.Next("detect")
			}
			logger.Warnf("[矿山勘查] [settle] 停止勘查失败")
			return statemachine.Retry{}
		},
		"farWait": func(sm *statemachine.Machine) statemachine.Result {
			ctx := sm.Ctx.(*surveyCtx)
			logger.Infof("[矿山勘查] [farWait] 导航回王国，本轮结束（等待期满后由调度再次拉起）")
			// 与 Lua 一致：回城失败仅告警，本轮照常 DONE。
			ctx.task.page.TapBackBtn()
			if !ctx.task.home.WaitCurrent(60 * time.Second) {
				logger.Warnf("[矿山勘查] [farWait] 勘查域返回矿山首页超时")
			}
			if !ctx.task.route.MineHomeToKingdom() {
				logger.Warnf("[矿山勘查] [farWait] 矿山首页返回王国超时")
			}
			return statemachine.Done{}
		},
	}
}
