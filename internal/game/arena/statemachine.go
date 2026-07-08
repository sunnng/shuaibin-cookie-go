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
			trophies, _ := ctx.task.page.ReadTrophyCount()
			ctx.task.session.Trophies = trophies
			logger.Infof("[Arena] sync medals=%d tickets=%d trophies=%d", medal, ticket, trophies)
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
			// Read countdown and persist
			if d, ok := ctx.task.page.ReadRefreshCountdown(); ok {
				ctx.task.session.SetNextFreeRefreshAt(time.Now().Add(d))
			}
			return statemachine.Next("leave")
		},
		"teamSelect": func(sm *statemachine.Machine) statemachine.Result {
			// Placeholder: tap selected opponent and wait team select
			return statemachine.Next("battle")
		},
		"battle": func(sm *statemachine.Machine) statemachine.Result {
			ctx := sm.Ctx.(*arenaCtx)
			result, ok := ctx.task.page.RunBattle()
			if !ok {
				return statemachine.Fatal{Err: errors.New("战斗失败")}
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
