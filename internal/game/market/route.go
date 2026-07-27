package market

import (
	"time"

	"app/internal/game/common/kingdom"
	"app/internal/logger"
	"app/internal/platform/action"
)

// Route 进出海滩交易所：王国首页 → 活动页 → 交易所；离开反向回王国首页。
type Route struct {
	page        *Page
	kingdomPage *kingdom.Page
	exec        action.Executor
}

func NewRoute(page *Page, kingdomPage *kingdom.Page, exec action.Executor) *Route {
	return &Route{page: page, kingdomPage: kingdomPage, exec: exec}
}

func (r *Route) waitEvent(timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if r.kingdomPage.IsEventPage() {
			return true
		}
		r.exec.Sleep(500)
	}
	return false
}

func (r *Route) Enter() bool {
	if r.page.IsCurrent() {
		return true
	}
	if !r.kingdomPage.IsEventPage() {
		if !r.kingdomPage.IsKingdomHome() {
			logger.Warnf("[MarketRoute] 不在王国主城/活动页，无法进入")
			return false
		}
		r.kingdomPage.TapEventBtn()
		if !r.waitEvent(30 * time.Second) {
			logger.Warnf("[MarketRoute] 等待王国活动页超时")
			return false
		}
	}
	r.page.TapEntryBtn()
	if r.page.WaitCurrent(30 * time.Second) {
		logger.Infof("[MarketRoute] 已进入海滩交易所")
		return true
	}
	logger.Warnf("[MarketRoute] 进入海滩交易所超时")
	return false
}

func (r *Route) Leave() bool {
	if r.kingdomPage.IsKingdomHome() {
		return true
	}
	if r.page.IsCurrent() {
		r.page.TapClose()
	}
	if r.kingdomPage.WaitHome(15 * time.Second) {
		logger.Infof("[MarketRoute] 已回王国主城")
		return true
	}
	if r.kingdomPage.IsEventPage() {
		r.exec.Back()
		r.exec.Sleep(1200)
		if r.kingdomPage.WaitHome(15 * time.Second) {
			logger.Infof("[MarketRoute] 已从活动页回王国主城")
			return true
		}
	}
	logger.Warnf("[MarketRoute] 回王国主城失败")
	return false
}
