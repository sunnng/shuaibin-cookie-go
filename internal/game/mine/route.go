package mine

import (
	"time"

	"app/internal/game/common/kingdom"
	"app/internal/logger"
)

// Route 矿山共享路由：王国首页 ⇄ 矿山首页，以及「矿山相关任意页面 → 王国首页」。
type Route struct {
	home        *Page
	kingdomPage *kingdom.Page

	// 等待超时可由测试缩短；NewRoute 填入与 Lua 一致的默认值。
	homeWaitTimeout    time.Duration // 等矿山首页，默认 60s
	kingdomWaitTimeout time.Duration // 等王国首页，默认 90s
}

func NewRoute(home *Page, kingdomPage *kingdom.Page) *Route {
	return &Route{
		home:               home,
		kingdomPage:        kingdomPage,
		homeWaitTimeout:    60 * time.Second,
		kingdomWaitTimeout: 90 * time.Second,
	}
}

// MiningRoutePage 是 ReturnToKingdom 需要的开采页窄接口（由 mining.Page 实现），
// 避免共享包反向依赖子任务包。
type MiningRoutePage interface {
	IsMiningPage() bool
	IsRewardPage() bool
	IsSettlementRoute() bool
	TapBackBtn()
}

// KingdomHomeToMineHome 王国首页 →（活动页）→ 矿山首页。
// 对应 Lua Route.kingdomHomeToMineHome：tapEventBtn → tapMineBtn → MineHomePage.wait。
func (r *Route) KingdomHomeToMineHome() bool {
	if r.home.IsCurrent() {
		return true
	}
	if !r.kingdomPage.IsKingdomHome() {
		logger.Warnf("[MineRoute] not in kingdom home, cannot enter mine")
		return false
	}
	r.kingdomPage.TapEventBtn()
	r.home.TapEntryMine()
	if r.home.WaitCurrent(r.homeWaitTimeout) {
		return true
	}
	logger.Warnf("[MineRoute] wait mine home timeout")
	return false
}

// MineHomeToKingdom 矿山首页 → 王国首页（Lua Route.mineHomeToKingdom，wait 默认 90s）。
func (r *Route) MineHomeToKingdom() bool {
	if r.kingdomPage.IsKingdomHome() {
		return true
	}
	r.home.TapBack()
	if r.kingdomPage.WaitHome(r.kingdomWaitTimeout) {
		return true
	}
	logger.Warnf("[MineRoute] wait kingdom home timeout")
	return false
}

// ReturnToKingdom 矿山相关任意页面 → 王国首页。
// 对应 Lua Route.returnToKingdom：开采页/奖励页先退回矿山首页，再回王国。
// 开采页返回矿山首页超时不视为失败（与 Lua 一致：仅告警后继续判断）。
func (r *Route) ReturnToKingdom(mp MiningRoutePage) bool {
	if r.kingdomPage.IsKingdomHome() {
		return true
	}
	if mp != nil && (mp.IsMiningPage() || mp.IsRewardPage() || mp.IsSettlementRoute()) {
		mp.TapBackBtn()
		if !r.home.WaitCurrent(r.homeWaitTimeout) {
			logger.Warnf("[MineRoute] 开采页返回矿山首页超时")
		}
	}
	if r.home.IsCurrent() {
		if r.MineHomeToKingdom() {
			logger.Infof("[MineRoute] 已回王国首页")
			return true
		}
		logger.Warnf("[MineRoute] 矿山首页返回王国超时")
		return false
	}
	if r.kingdomPage.IsKingdomHome() {
		return true
	}
	logger.Warnf("[MineRoute] 回王国首页失败，当前页面未知")
	return false
}
