package action

import (
	"fmt"

	"github.com/Dasongzi1366/AutoGo/device"
	"github.com/Dasongzi1366/AutoGo/motion"
)

type AndroidExecutor struct {
	displayId int
}

func NewAndroidExecutor(displayId int) *AndroidExecutor {
	return &AndroidExecutor{displayId: displayId}
}

func (e *AndroidExecutor) Tap(p Point) error {
	w, h, _, _ := device.GetDisplayInfo(e.displayId)
	if w == 0 || h == 0 {
		return fmt.Errorf("failed to get display info")
	}
	sp := SafeTap(p, w, h)
	motion.Click(sp.X, sp.Y, 0, e.displayId)
	return nil
}

func (e *AndroidExecutor) LongTap(p Point, ms int) error {
	w, h, _, _ := device.GetDisplayInfo(e.displayId)
	if w == 0 || h == 0 {
		return fmt.Errorf("failed to get display info")
	}
	sp := SafeTap(p, w, h)
	motion.LongClick(sp.X, sp.Y, ms, 0, e.displayId)
	return nil
}

func (e *AndroidExecutor) Swipe(from, to Point, ms int) error {
	w, h, _, _ := device.GetDisplayInfo(e.displayId)
	if w == 0 || h == 0 {
		return fmt.Errorf("failed to get display info")
	}
	sf := SafeTap(from, w, h)
	st := SafeTap(to, w, h)
	motion.Swipe(sf.X, sf.Y, st.X, st.Y, ms, 0, e.displayId)
	return nil
}
