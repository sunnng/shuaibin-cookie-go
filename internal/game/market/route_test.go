package market

import (
	"testing"

	"app/internal/game/common/kingdom"
	"app/internal/platform/screen"
)

type routeFixture struct {
	det *mockDetector
	exe *mockExecutor
	rt  *Route
}

func newRouteFixture() *routeFixture {
	det := &mockDetector{matchMulti: map[string]bool{}}
	exe := &mockExecutor{}
	mf := DefaultFeature()
	mf.Page.Identify = screen.Feature{Colors: "market", Sim: 0.9}
	mf.Entry.Btn = screen.Region{X1: 200, Y1: 500, X2: 200, Y2: 500}
	mf.Page.CloseBtn = screen.Region{X1: 300, Y1: 20, X2: 300, Y2: 20}
	kf := &kingdom.Feature{
		Home: kingdom.PageSlot{Identify: screen.Feature{Colors: "home", Sim: 0.9}},
		Event: kingdom.OCRPage{
			Region:  screen.Region{X1: 681, Y1: 171, X2: 915, Y2: 242},
			Keyword: "王国活动",
		},
		Actions: kingdom.NavActions{EventBtn: screen.Region{X1: 100, Y1: 800, X2: 100, Y2: 800}},
	}
	mp := NewPage(det, exe, mf)
	kp := kingdom.NewPage(det, exe, kf)
	return &routeFixture{det: det, exe: exe, rt: NewRoute(mp, kp, exe)}
}

func (fx *routeFixture) setEventOCR(hit bool) {
	fx.det.findOCR = func(r screen.Region, keyword string) (screen.Point, bool) {
		if keyword == "王国活动" && hit {
			return screen.Point{X: r.X1 + 1, Y: r.Y1 + 1}, true
		}
		return screen.Point{}, false
	}
}

func TestRouteEnterAlreadyCurrent(t *testing.T) {
	fx := newRouteFixture()
	fx.det.matchMulti["market"] = true
	if !fx.rt.Enter() {
		t.Fatal("want enter true when already in market")
	}
	if len(fx.exe.taps) != 0 {
		t.Fatalf("no taps expected, got %v", fx.exe.taps)
	}
}

func TestRouteEnterFromHome(t *testing.T) {
	fx := newRouteFixture()
	fx.det.matchMulti["home"] = true
	fx.setEventOCR(false)
	fx.exe.onTap = func(n int) {
		switch n {
		case 1: // 首页 → 活动页
			fx.det.matchMulti["home"] = false
			fx.setEventOCR(true)
		case 2: // 活动页 → 交易所
			fx.setEventOCR(false)
			fx.det.matchMulti["market"] = true
		}
	}
	if !fx.rt.Enter() {
		t.Fatal("want enter success from home")
	}
	if len(fx.exe.taps) != 2 {
		t.Fatalf("want event+entry taps, got %v", fx.exe.taps)
	}
}

func TestRouteEnterNotHomeFails(t *testing.T) {
	fx := newRouteFixture()
	fx.setEventOCR(false)
	if fx.rt.Enter() {
		t.Fatal("want enter false when not home/event/market")
	}
	if len(fx.exe.taps) != 0 {
		t.Fatalf("no taps expected, got %v", fx.exe.taps)
	}
}

func TestRouteLeaveViaClose(t *testing.T) {
	fx := newRouteFixture()
	fx.det.matchMulti["market"] = true
	fx.exe.onTap = func(n int) { // 关闭交易所 → 回王国首页
		fx.det.matchMulti["market"] = false
		fx.det.matchMulti["home"] = true
	}
	if !fx.rt.Leave() {
		t.Fatal("want leave success via close button")
	}
	if len(fx.exe.taps) != 1 {
		t.Fatalf("want one close tap, got %v", fx.exe.taps)
	}
}

func TestRouteLeaveAlreadyHome(t *testing.T) {
	fx := newRouteFixture()
	fx.det.matchMulti["home"] = true
	if !fx.rt.Leave() {
		t.Fatal("want leave true when already home")
	}
	if len(fx.exe.taps) != 0 {
		t.Fatalf("no taps expected, got %v", fx.exe.taps)
	}
}
