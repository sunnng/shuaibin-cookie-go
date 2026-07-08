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
			return false
		}
	}
	// TODO: OCR tap "王国竞技场"
	return r.page.TapToLobby()
}

func (r *Route) Leave() bool {
	if r.kingdomPage.IsKingdomHome() {
		return true
	}
	if r.page.IsLobby() {
		// tap close
	}
	// navigate back to kingdom home
	return true
}
