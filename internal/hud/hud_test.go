package hud

import "testing"

func TestHUD(t *testing.T) {
	h := New()
	h.SetTask("王国竞技场", "执行中")
	h.SetIdle()
	h.SetWait("刷新倒计时")
}
