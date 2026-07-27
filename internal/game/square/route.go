package square

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

// Enter 王国主城 → 布谷鸟广场（Lua SquareRoute.kingdomToSquare）。
func (r *Route) Enter() bool {
	if r.page.IsSquare() {
		return true
	}
	if !r.kingdomPage.IsKingdomHome() {
		logger.Warnf("[SquareRoute] 不在王国主城，无法进广场")
		return false
	}
	r.page.TapEntryBtn()
	if r.page.WaitSquare(30 * time.Second) {
		logger.Infof("[SquareRoute] 已进入广场")
		return true
	}
	logger.Warnf("[SquareRoute] 进入广场超时")
	return false
}

// EnsureSquare 确保处于广场上下文（广场页或离开弹窗）；身处王国主城时尝试进入
// （Lua 广场_任务.lua 的 ensureSquarePage）。
func (r *Route) EnsureSquare() bool {
	if r.page.IsSquare() || r.page.IsLeaveDialog() {
		return true
	}
	if r.kingdomPage.IsKingdomHome() {
		return r.Enter()
	}
	logger.Warnf("[SquareRoute] 当前界面未知，无法进入广场")
	return false
}

// IsSquareContext 是否在广场页或离开弹窗（Lua SquareRoute.isSquareContext）。
func (r *Route) IsSquareContext() bool {
	return r.page.IsSquare() || r.page.IsLeaveDialog()
}

// OpenLeaveDialog 打开「离开广场」弹窗（Lua SquareRoute.openLeaveDialog）。
func (r *Route) OpenLeaveDialog() bool {
	if r.page.IsLeaveDialog() {
		return true
	}
	if !r.page.IsSquare() {
		if !r.Enter() {
			return false
		}
	}
	r.page.TapBack()
	if r.page.WaitLeaveDialog(15 * time.Second) {
		logger.Infof("[SquareRoute] 已打开离开广场弹窗")
		return true
	}
	logger.Warnf("[SquareRoute] 离开广场弹窗未出现")
	return false
}

// LeaveToKingdom 经弹窗或广场返回王国主城（Lua SquareRoute.leaveDialogToKingdom）。
func (r *Route) LeaveToKingdom(timeout time.Duration) bool {
	if r.kingdomPage.IsKingdomHome() {
		return true
	}
	if r.page.IsLeaveDialog() {
		r.page.TapReturnKingdom()
	} else if r.page.IsSquare() {
		r.page.TapBack()
		if r.page.WaitLeaveDialog(8 * time.Second) {
			r.page.TapReturnKingdom()
		}
	}
	if r.kingdomPage.WaitHome(timeout) {
		logger.Infof("[SquareRoute] 已回王国主城")
		return true
	}
	logger.Warnf("[SquareRoute] 回王国主城超时")
	return false
}

// Leave 供其它任务在运行前调用：若卡在广场（页或弹窗）则离开回王国主城
// （Lua SquareTask.leaveForOtherTask 的页面流转部分；停留计时的暂停由
// Task.Leave 负责，Route 不持有会话）。
func (r *Route) Leave() bool {
	if r.kingdomPage.IsKingdomHome() {
		return true
	}
	if r.page.IsLeaveDialog() {
		return r.LeaveToKingdom(30 * time.Second)
	}
	if r.page.IsSquare() {
		r.page.TapBack()
		if r.page.WaitLeaveDialog(8 * time.Second) {
			return r.LeaveToKingdom(30 * time.Second)
		}
	}
	return r.LeaveToKingdom(30 * time.Second)
}
