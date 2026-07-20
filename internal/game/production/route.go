package production

import (
	"time"

	"app/internal/game/common/kingdom"
	"app/internal/logger"
)

// Route 跨页进出王国生产界面。
type Route struct {
	page        *Page
	kingdomPage *kingdom.Page
}

// NewRoute 构造路线；生产入口通常从王国首页进入（具体点击点待特征填充）。
func NewRoute(page *Page, kingdomPage *kingdom.Page) *Route {
	return &Route{page: page, kingdomPage: kingdomPage}
}

// Enter 尝试进入生产界面。骨架：已在界面则成功；否则仅保证回到王国首页，真正入口待实现。
func (r *Route) Enter() bool {
	if r.page != nil && r.page.IsBoard() {
		logger.Infof("[ProductionRoute] already on board")
		return true
	}
	if r.kingdomPage == nil {
		logger.Warnf("[ProductionRoute] kingdom page nil")
		return false
	}
	if r.kingdomPage.IsKingdomHome() {
		logger.Infof("[ProductionRoute] on kingdom home; entry tap not configured yet")
		return false
	}
	if r.kingdomPage.HasBackHome() {
		r.kingdomPage.TapBackHome()
		if !r.kingdomPage.WaitHome(15 * time.Second) {
			logger.Warnf("[ProductionRoute] wait kingdom home timeout")
			return false
		}
		logger.Infof("[ProductionRoute] back to kingdom home; entry tap not configured yet")
		return false
	}
	logger.Warnf("[ProductionRoute] cannot reach kingdom home")
	return false
}

// Leave 离开生产界面回到王国首页。骨架：已在首页则成功。
func (r *Route) Leave() bool {
	if r.kingdomPage == nil {
		return true
	}
	if r.kingdomPage.IsKingdomHome() {
		return true
	}
	if r.kingdomPage.HasBackHome() {
		r.kingdomPage.TapBackHome()
		return r.kingdomPage.WaitHome(15 * time.Second)
	}
	return false
}
