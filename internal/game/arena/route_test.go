package arena

import (
	"image"
	"testing"

	"app/internal/game/common/kingdom"
	"app/internal/platform/action"
	"app/internal/platform/screen"
)

type routeDet struct {
	match map[string]bool
	ocrPt screen.Point
	ocrOK bool
	advOCR bool

	onAdventureTap func()
	onEntryTap     func()
}

func (d *routeDet) Capture() *image.NRGBA { return nil }
func (d *routeDet) MatchColor(x, y int, color string, sim float32) bool {
	return false
}
func (d *routeDet) FindColor(region screen.Region, color string, sim float32, dir int) (screen.Point, bool) {
	return screen.Point{}, false
}
func (d *routeDet) FindMultiColorsAll(region screen.Region, colors string, sim float32, dir int) []screen.Point {
	return nil
}
func (d *routeDet) MatchMultiColor(colors string, sim float32) bool { return d.match[colors] }
func (d *routeDet) MatchImage(region screen.Region, template []byte, sim float32) (screen.Point, bool) {
	return screen.Point{}, false
}
func (d *routeDet) OCRText(region screen.Region) string { return "" }
func (d *routeDet) FindOCRText(region screen.Region, keyword string) (screen.Point, bool) {
	if keyword == "冒险" {
		if !d.advOCR {
			return screen.Point{}, false
		}
		return screen.Point{X: region.X1 + 1, Y: region.Y1 + 1}, true
	}
	if !d.ocrOK {
		return screen.Point{}, false
	}
	return d.ocrPt, true
}

type routeExec struct {
	taps []action.Point
	det  *routeDet
}

func (e *routeExec) Tap(p action.Point) error {
	e.taps = append(e.taps, p)
	if e.det == nil {
		return nil
	}
	if len(e.taps) == 1 && e.det.onAdventureTap != nil {
		e.det.onAdventureTap()
	}
	if len(e.taps) == 2 && e.det.onEntryTap != nil {
		e.det.onEntryTap()
	}
	return nil
}
func (e *routeExec) LongTap(p action.Point, ms int) error              { return nil }
func (e *routeExec) Swipe(from, to action.Point, durationMs int) error { return nil }
func (e *routeExec) Back() error                                       { return nil }
func (e *routeExec) Home() error                                       { return nil }
func (e *routeExec) Sleep(ms int)                                      {}

type leaveExec struct {
	taps []action.Point
	det  *routeDet
}

func (e *leaveExec) Tap(p action.Point) error {
	e.taps = append(e.taps, p)
	e.det.match["lobby"] = false
	return nil
}
func (e *leaveExec) LongTap(p action.Point, ms int) error              { return nil }
func (e *leaveExec) Swipe(from, to action.Point, durationMs int) error { return nil }
func (e *leaveExec) Back() error                                       { return nil }
func (e *leaveExec) Home() error                                       { return nil }
func (e *leaveExec) Sleep(ms int)                                      {}

func TestRouteEnterAlreadyLobby(t *testing.T) {
	det := &routeDet{match: map[string]bool{"lobby": true}}
	exec := &routeExec{}
	af := DefaultFeature()
	af.Lobby.Identify = screen.Feature{Colors: "lobby", Sim: 0.9}
	ap := NewPage(det, exec, af)
	kp := kingdom.NewPage(det, exec, kingdom.DefaultFeature())
	r := NewRoute(ap, kp)
	if !r.Enter() {
		t.Fatal("want enter true when already lobby")
	}
	if len(exec.taps) != 0 {
		t.Fatalf("no taps expected, got %v", exec.taps)
	}
}

func TestRouteEnterFromHome(t *testing.T) {
	det := &routeDet{
		match: map[string]bool{"home": true},
		ocrPt: screen.Point{X: 640, Y: 500},
		ocrOK: true,
	}
	det.onAdventureTap = func() {
		det.match["home"] = false
		det.advOCR = true
	}
	det.onEntryTap = func() {
		det.advOCR = false
		det.match["lobby"] = true
	}
	exec := &routeExec{det: det}
	kf := &kingdom.Feature{
		Home: kingdom.PageSlot{Identify: screen.Feature{Colors: "home", Sim: 0.9}},
		Adventure: kingdom.OCRPage{
			Region:  screen.Region{X1: 89, Y1: 28, X2: 181, Y2: 76},
			Keyword: "冒险",
		},
		Actions: kingdom.NavActions{AdventureBtn: screen.Region{X1: 100, Y1: 800, X2: 100, Y2: 800}},
	}
	af := DefaultFeature()
	af.Lobby.Identify = screen.Feature{Colors: "lobby", Sim: 0.9}
	af.Entry.Region = screen.Region{X1: 0, Y1: 0, X2: 1600, Y2: 900}

	ap := NewPage(det, exec, af)
	kp := kingdom.NewPage(det, exec, kf)
	r := NewRoute(ap, kp)

	if !r.Enter() {
		t.Fatal("want enter success from home")
	}
	if len(exec.taps) < 2 {
		t.Fatalf("want adventure+entry taps, got %v", exec.taps)
	}
}

func TestRouteEnterNotHomeFails(t *testing.T) {
	det := &routeDet{match: map[string]bool{}}
	exec := &routeExec{}
	ap := NewPage(det, exec, DefaultFeature())
	kp := kingdom.NewPage(det, exec, kingdom.DefaultFeature())
	if NewRoute(ap, kp).Enter() {
		t.Fatal("want enter false when not home/adventure/lobby")
	}
}

func TestRouteLeaveNeedsBackHome(t *testing.T) {
	det := &routeDet{match: map[string]bool{"lobby": true}}
	exec := &leaveExec{det: det}
	af := DefaultFeature()
	af.Lobby.Identify = screen.Feature{Colors: "lobby", Sim: 0.9}
	af.Lobby.Actions.Close = screen.Region{X1: 1500, Y1: 20, X2: 1580, Y2: 80}
	ap := NewPage(det, exec, af)
	kp := kingdom.NewPage(det, exec, kingdom.DefaultFeature())
	r := NewRoute(ap, kp)
	if r.Leave() {
		t.Fatal("want leave false without BackHome / home identify")
	}
}

func TestParseBattleResult(t *testing.T) {
	cases := []struct {
		in     string
		want   string
		wantOK bool
	}{
		{"战斗胜利", "胜利", true},
		{"平局结算", "平局", true},
		{"你失败了", "失败", true},
		{"???", "", false},
	}
	for _, c := range cases {
		got, ok := parseBattleResult(c.in)
		if got != c.want || ok != c.wantOK {
			t.Errorf("parseBattleResult(%q)=(%q,%v) want (%q,%v)", c.in, got, ok, c.want, c.wantOK)
		}
	}
}

func TestTapEntry(t *testing.T) {
	det := &routeDet{ocrPt: screen.Point{X: 700, Y: 400}, ocrOK: true}
	exec := &routeExec{}
	f := DefaultFeature()
	f.Entry.Region = screen.Region{X1: 0, Y1: 0, X2: 100, Y2: 100}
	p := NewPage(det, exec, f)
	if !p.TapEntry() {
		t.Fatal("want TapEntry true")
	}
	if len(exec.taps) != 1 || exec.taps[0].X != 700 {
		t.Fatalf("taps=%v", exec.taps)
	}
}

func TestRunBattleRequiresSettlement(t *testing.T) {
	det := &routeDet{match: map[string]bool{}}
	p := NewPage(det, &routeExec{}, DefaultFeature())
	if _, ok := p.RunBattle(); ok {
		t.Fatal("empty settlement identify must fail")
	}
}
