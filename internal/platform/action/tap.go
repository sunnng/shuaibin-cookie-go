package action

import (
	"app/internal/logger"

	"github.com/Dasongzi1366/AutoGo/device"
	"github.com/Dasongzi1366/AutoGo/motion"
)

type AndroidExecutor struct {
	displayId int
}

func NewAndroidExecutor(displayId int) *AndroidExecutor {
	return &AndroidExecutor{displayId: displayId}
}

// safePoint 把基准坐标缩放并收敛到实际屏幕内；取不到显示信息（通常是设备
// 掉线）时告警并返回 false，调用方跳过本次动作。
func (e *AndroidExecutor) safePoint(p Point) (Point, bool) {
	w, h, _, _ := device.GetDisplayInfo(e.displayId)
	if w == 0 || h == 0 {
		logger.Warnf("[Action] display info unavailable (device offline?), skip input at %+v", p)
		return Point{}, false
	}
	return SafeTap(p, w, h), true
}

func (e *AndroidExecutor) Tap(p Point) {
	sp, ok := e.safePoint(p)
	if !ok {
		return
	}
	motion.Click(sp.X, sp.Y, 0, e.displayId)
}

func (e *AndroidExecutor) LongTap(p Point, ms int) {
	sp, ok := e.safePoint(p)
	if !ok {
		return
	}
	motion.LongClick(sp.X, sp.Y, ms, 0, e.displayId)
}

func (e *AndroidExecutor) Swipe(from, to Point, ms int) {
	sf, ok := e.safePoint(from)
	if !ok {
		return
	}
	st, ok := e.safePoint(to)
	if !ok {
		return
	}
	motion.Swipe(sf.X, sf.Y, st.X, st.Y, ms, 0, e.displayId)
}
