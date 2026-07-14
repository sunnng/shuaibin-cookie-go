package arena

import (
	"time"

	"app/internal/game/common/kingdom"
	"app/internal/logger"
)

type Route struct {
	page        *Page
	kingdomPage *kingdom.Page
}

func NewRoute(page *Page, kingdomPage *kingdom.Page) *Route {
	return &Route{page: page, kingdomPage: kingdomPage}
}

func (r *Route) Enter() bool {
	if r.page.IsLobby() {
		logger.Infof("[ArenaRoute] already in lobby")
		return true
	}
	if !r.kingdomPage.IsAdventurePage() {
		if !r.kingdomPage.IsKingdomHome() {
			logger.Warnf("[ArenaRoute] not in kingdom home, cannot enter")
			return false
		}
		r.kingdomPage.TapAdventureBtn()
		if !r.kingdomPage.WaitAdventure(30 * time.Second) {
			logger.Warnf("[ArenaRoute] wait adventure timeout")
			return false
		}
	}
	for attempt := 0; attempt < 2; attempt++ {
		if !r.page.TapEntry() {
			continue
		}
		if r.page.WaitLobby(30 * time.Second) {
			return true
		}
	}
	logger.Warnf("[ArenaRoute] enter lobby failed")
	return false
}

func (r *Route) Leave() bool {
	if r.kingdomPage.IsKingdomHome() {
		return true
	}
	if r.page.IsLobby() {
		for i := 0; i < 3; i++ {
			if !r.page.IsLobby() {
				break
			}
			r.page.TapLobbyClose()
		}
	}
	if r.kingdomPage.IsKingdomHome() {
		return true
	}
	if !r.kingdomPage.HasBackHome() {
		logger.Warnf("[ArenaRoute] BackHome not configured; cannot return to kingdom home")
		return false
	}
	r.kingdomPage.TapBackHome()
	if r.kingdomPage.WaitHome(15 * time.Second) {
		return true
	}
	logger.Warnf("[ArenaRoute] wait home timeout")
	return false
}
