package starlight

import (
	"time"

	"app/internal/game/common/kingdom"
	"app/internal/logger"
)

// Route 从王国首页导航到梦幻繁星岛首页，对应 Lua 繁星岛_路由.lua。
type Route struct {
	page        *Page
	kingdomPage *kingdom.Page
}

func NewRoute(page *Page, kingdomPage *kingdom.Page) *Route {
	return &Route{page: page, kingdomPage: kingdomPage}
}

func (r *Route) IsStarlightHome() bool {
	return r.page.IsHomePage()
}

// KingdomToHome 从王国首页导航到梦幻繁星岛首页：
// 点事件按钮 → 等活动页 → 点梦幻繁星岛入口 → 等繁星岛首页。
func (r *Route) KingdomToHome() bool {
	if r.page.IsHomePage() {
		return true
	}
	if !r.kingdomPage.IsKingdomHome() {
		logger.Warnf("[梦幻繁星岛.路由] 当前不在王国首页，无法导航")
		return false
	}
	r.kingdomPage.TapEventBtn()
	if !r.page.WaitEventPage(10 * time.Second) {
		logger.Warnf("[梦幻繁星岛.路由] 等待事件页超时")
		return false
	}
	if !r.page.TapStarlightEntry() {
		return false
	}
	return r.page.WaitHomePage(30 * time.Second)
}

// EnsureHome 确保位于繁星岛首页。
func (r *Route) EnsureHome() bool {
	if r.page.IsHomePage() {
		return true
	}
	return r.KingdomToHome()
}
