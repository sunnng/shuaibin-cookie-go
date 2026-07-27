package starlight

import (
	"testing"

	"app/internal/game/common/kingdom"
	"app/internal/platform/action"
	"app/internal/platform/screen"
)

// routeDet 按颜色串区分页面：kingdom home / event / starlight home。
type routeDet struct {
	mockDetector
	match map[string]bool
}

func (d *routeDet) MatchMultiColor(colors string, sim float32) bool { return d.match[colors] }

// routeExec 记录点击；第一次点击（事件按钮）后进入活动页，
// 第二次点击（繁星岛入口）后进入繁星岛首页。
type routeExec struct {
	mockExecutor
	det *routeDet
}

func (e *routeExec) Tap(p action.Point) {
	e.taps = append(e.taps, p)
	if e.det == nil {
		return
	}
	switch len(e.taps) {
	case 1:
		e.det.match["kingdom"] = false
		e.det.match["event"] = true
	case 2:
		e.det.match["event"] = false
		e.det.match["home"] = true
	}
}

func newRoutePages(det *routeDet, exec action.Executor) (*Page, *kingdom.Page) {
	f := DefaultFeature()
	f.Home.Identify = screen.Feature{Colors: "home", Sim: 0.9}
	f.Event.Identify = screen.Feature{Colors: "event", Sim: 0.9}
	p := NewPage(det, exec, f)
	kf := &kingdom.Feature{
		Home:    kingdom.PageSlot{Identify: screen.Feature{Colors: "kingdom", Sim: 0.9}},
		Actions: kingdom.NavActions{EventBtn: screen.Region{X1: 246, Y1: 802, X2: 271, Y2: 832}},
	}
	kp := kingdom.NewPage(det, exec, kf)
	return p, kp
}

func TestRouteEnsureHomeAlreadyHome(t *testing.T) {
	det := &routeDet{match: map[string]bool{"home": true}}
	exec := &mockExecutor{}
	p, kp := newRoutePages(det, exec)
	r := NewRoute(p, kp)
	if !r.EnsureHome() {
		t.Fatal("want EnsureHome true when already on starlight home")
	}
	if len(exec.taps) != 0 {
		t.Fatalf("no taps expected, got %v", exec.taps)
	}
}

func TestRouteKingdomToHome(t *testing.T) {
	det := &routeDet{match: map[string]bool{"kingdom": true}}
	exec := &routeExec{det: det}
	p, kp := newRoutePages(det, exec)
	r := NewRoute(p, kp)
	if !r.KingdomToHome() {
		t.Fatal("want KingdomToHome success from kingdom home")
	}
	if len(exec.taps) != 2 {
		t.Fatalf("want event btn + starlight entry taps, got %v", exec.taps)
	}
}

func TestRouteNotKingdomHomeFails(t *testing.T) {
	det := &routeDet{match: map[string]bool{}}
	exec := &routeExec{det: det}
	p, kp := newRoutePages(det, exec)
	if NewRoute(p, kp).EnsureHome() {
		t.Fatal("want EnsureHome false when not on kingdom home")
	}
	if len(exec.taps) != 0 {
		t.Fatalf("no taps expected, got %v", exec.taps)
	}
}

func TestRouteEntryUnconfiguredFails(t *testing.T) {
	det := &routeDet{match: map[string]bool{"kingdom": true}}
	exec := &routeExec{det: det}
	p, kp := newRoutePages(det, exec)
	p.feature.Event.Actions.StarlightBtn = screen.Region{} // 入口未配置
	if NewRoute(p, kp).KingdomToHome() {
		t.Fatal("want KingdomToHome false when starlight entry not configured")
	}
}
