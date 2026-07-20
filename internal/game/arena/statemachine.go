package arena

import (
	"errors"
	"time"

	"app/internal/config"
	"app/internal/logger"
	"app/internal/statemachine"
)

type arenaCtx struct {
	task             *Task
	cfg              *config.Arena
	selectedOpponent *OpponentInfo
}

// refreshOCRBackoff is used when free-refresh countdown OCR fails, so CheckReady
// does not immediately re-enter the arena.
const refreshOCRBackoff = 60 * time.Second

func (t *Task) handlers() map[string]statemachine.Handler {
	return map[string]statemachine.Handler{
		"detect": func(sm *statemachine.Machine) statemachine.Result {
			ctx := sm.Ctx.(*arenaCtx)
			if ctx.task.page.IsLobby() {
				return statemachine.Next("sync")
			}
			if ctx.task.kingdomPage != nil && (ctx.task.kingdomPage.IsKingdomHome() || ctx.task.kingdomPage.IsAdventurePage()) {
				return statemachine.Next("navigate")
			}
			return statemachine.Fatal{Err: errors.New("无法识别当前页面")}
		},
		"navigate": func(sm *statemachine.Machine) statemachine.Result {
			ctx := sm.Ctx.(*arenaCtx)
			if ctx.task.page.IsLobby() {
				return statemachine.Next("sync")
			}
			if ctx.task.route.Enter() {
				return statemachine.Next("sync")
			}
			return statemachine.Keep{}
		},
		"sync": func(sm *statemachine.Machine) statemachine.Result {
			ctx := sm.Ctx.(*arenaCtx)
			medal, ticket, ok := ctx.task.page.ReadMedalAndTicket()
			if ok {
				ctx.task.session.Medals = medal
				ctx.task.session.Tickets = ticket
			}
			trophies, ok := ctx.task.page.ReadTrophyCount()
			if !ok {
				logger.Warnf("[Arena] trophy OCR failed, retry")
				return statemachine.Retry{}
			}
			ctx.task.session.Trophies = trophies
			logger.Infof("[Arena] sync medals=%d tickets=%d trophies=%d", medal, ticket, trophies)
			ctx.task.pushStatus()
			return statemachine.Next("check")
		},
		"check": func(sm *statemachine.Machine) statemachine.Result {
			ctx := sm.Ctx.(*arenaCtx)
			if ctx.task.session.IsReachMaxBattles(ctx.cfg) {
				logger.Infof("[Arena] max battles reached")
				return statemachine.Next("leave")
			}
			if ctx.task.session.Tickets <= 0 {
				if ctx.cfg.AutoBuyCount <= 0 || ctx.task.session.BuyCount >= ctx.cfg.AutoBuyCount {
					logger.Infof("[Arena] no tickets and cannot buy")
					return statemachine.Next("leave")
				}
				return statemachine.Next("buyTicket")
			}
			return statemachine.Next("selectOpponent")
		},
		"buyTicket": func(sm *statemachine.Machine) statemachine.Result {
			ctx := sm.Ctx.(*arenaCtx)
			ctx.task.page.BuyTicket()
			ctx.task.session.BuyCount++
			return statemachine.Next("sync")
		},
		"selectOpponent": func(sm *statemachine.Machine) statemachine.Result {
			ctx := sm.Ctx.(*arenaCtx)
			info := ctx.task.page.FindFirstValidOpponent(ctx.cfg, ctx.task.session.Trophies)
			if info != nil {
				ctx.selectedOpponent = info
				return statemachine.Next("teamSelect")
			}
			// Try swipe once
			ctx.task.page.SwipePageLeft()
			info = ctx.task.page.FindFirstValidOpponent(ctx.cfg, ctx.task.session.Trophies)
			if info != nil {
				ctx.selectedOpponent = info
				return statemachine.Next("teamSelect")
			}
			if ctx.task.page.IsFreeRefresh() {
				ctx.task.page.TapFreeRefresh()
				ctx.task.session.ClearNextFreeRefresh()
				return statemachine.Next("selectOpponent")
			}
			// Read countdown and persist; OCR miss uses a short backoff to avoid thrash.
			if d, ok := ctx.task.page.ReadRefreshCountdown(); ok {
				ctx.task.session.SetNextFreeRefreshAt(time.Now().Add(d))
			} else {
				logger.Warnf("[Arena] refresh countdown OCR failed, backoff %v", refreshOCRBackoff)
				ctx.task.session.SetNextFreeRefreshAt(time.Now().Add(refreshOCRBackoff))
			}
			return statemachine.Next("leave")
		},
		"teamSelect": func(sm *statemachine.Machine) statemachine.Result {
			ctx := sm.Ctx.(*arenaCtx)
			if ctx.selectedOpponent == nil {
				return statemachine.Fatal{Err: errors.New("未选中对手")}
			}
			ctx.task.page.TapOpponentSite(ctx.selectedOpponent.Site)
			if ctx.task.page.HasTeamSelectPage() {
				if !ctx.task.page.WaitTeamSelect(15 * time.Second) {
					return statemachine.Retry{}
				}
			}
			ctx.task.page.TapStartBattle()
			return statemachine.Next("battle")
		},
		"battle": func(sm *statemachine.Machine) statemachine.Result {
			ctx := sm.Ctx.(*arenaCtx)
			result, ok := ctx.task.page.RunBattle()
			if !ok {
				// 结算未出现/结果 OCR 失败都属于识别类失败，不是战败；
				// 重试（受 MaxRetry 约束），不要当 Fatal 终止整轮。
				logger.Warnf("[Arena] battle result recognition failed, retry")
				return statemachine.Retry{}
			}
			switch result {
			case "胜利":
				ctx.task.session.Wins++
			case "平局":
				ctx.task.session.Draws++
			case "失败":
				ctx.task.session.Losses++
			}
			if ctx.task.session.Tickets > 0 {
				ctx.task.session.Tickets--
			}
			return statemachine.Next("sync")
		},
		"leave": func(sm *statemachine.Machine) statemachine.Result {
			ctx := sm.Ctx.(*arenaCtx)
			if !ctx.task.route.Leave() {
				return statemachine.Fatal{Err: errors.New("离开竞技场失败")}
			}
			return statemachine.Done{}
		},
	}
}
