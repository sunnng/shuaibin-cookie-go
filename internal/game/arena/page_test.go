package arena

import (
	"fmt"
	"image"
	"testing"
	"time"

	"app/internal/config"
	"app/internal/platform/action"
	"app/internal/platform/screen"
)

func TestNewPage(t *testing.T) {
	_ = NewPage(nil, nil, DefaultFeature())
}

func TestOffsetRegion(t *testing.T) {
	rel := screen.Region{X1: -10, Y1: -20, X2: 10, Y2: 20}
	a := screen.Point{X: 100, Y: 200}
	got := offsetRegion(rel, a)
	want := screen.Region{X1: 90, Y1: 180, X2: 110, Y2: 220}
	if got != want {
		t.Fatalf("offsetRegion = %+v, want %+v", got, want)
	}
}

func TestOffsetPoint(t *testing.T) {
	got := offsetPoint(screen.Point{X: 30, Y: -5}, screen.Point{X: 100, Y: 200})
	want := screen.Point{X: 130, Y: 195}
	if got != want {
		t.Fatalf("offsetPoint = %+v, want %+v", got, want)
	}
}

func TestReadInt(t *testing.T) {
	cases := []struct {
		in     string
		wantN  int
		wantOK bool
	}{
		{"1050", 1050, true},
		{"  99 ", 99, true},
		{"abc", 0, false},
		{"", 0, false},
	}
	for _, c := range cases {
		n, ok := readInt(c.in)
		if n != c.wantN || ok != c.wantOK {
			t.Errorf("readInt(%q) = (%d,%v), want (%d,%v)", c.in, n, ok, c.wantN, c.wantOK)
		}
	}
}

func TestParseCountdown(t *testing.T) {
	cases := []struct {
		in     string
		want   time.Duration
		wantOK bool
	}{
		{"5分30秒", 330 * time.Second, true},
		{"30秒", 30 * time.Second, true},
		{"5分", 300 * time.Second, true},
		{"05:30", 330 * time.Second, true},
		{"", 0, false},
		{"abc", 0, false},
		{"0秒", 0, false},
	}
	for _, c := range cases {
		got, ok := parseCountdown(c.in)
		if got != c.want || ok != c.wantOK {
			t.Errorf("parseCountdown(%q) = (%v,%v), want (%v,%v)", c.in, got, ok, c.want, c.wantOK)
		}
	}
}

// ---- mock Detector ----
type mockDetector struct {
	matchMulti bool
	ocrByKey   map[string]string // key: "x1,y1,x2,y2"
	matchByKey map[string]bool   // key: "x,y,color"
	anchors    []screen.Point
}

func (m *mockDetector) Capture() *image.NRGBA { return nil }
func (m *mockDetector) MatchColor(x, y int, color string, sim float32) bool {
	return m.matchByKey[fmt.Sprintf("%d,%d,%s", x, y, color)]
}
func (m *mockDetector) FindColor(r screen.Region, c string, s float32, d int) (screen.Point, bool) {
	return screen.Point{}, false
}
func (m *mockDetector) FindMultiColorsAll(r screen.Region, c string, s float32, d int) []screen.Point {
	return m.anchors
}
func (m *mockDetector) MatchMultiColor(colors string, sim float32) bool { return m.matchMulti }
func (m *mockDetector) MatchImage(r screen.Region, t []byte, s float32) (screen.Point, bool) {
	return screen.Point{}, false
}
func (m *mockDetector) OCRText(r screen.Region) string {
	return m.ocrByKey[fmt.Sprintf("%d,%d,%d,%d", r.X1, r.Y1, r.X2, r.Y2)]
}

// ---- mock Executor ----
type mockExecutor struct {
	taps   []action.Point
	swipes [][2]action.Point
	sleeps []int
}

func (e *mockExecutor) Tap(p action.Point) error { e.taps = append(e.taps, p); return nil }
func (e *mockExecutor) LongTap(p action.Point, ms int) error {
	return nil
}
func (e *mockExecutor) Swipe(f, t action.Point, ms int) error {
	e.swipes = append(e.swipes, [2]action.Point{f, t})
	return nil
}
func (e *mockExecutor) Back() error  { return nil }
func (e *mockExecutor) Home() error  { return nil }
func (e *mockExecutor) Sleep(ms int) { e.sleeps = append(e.sleeps, ms) }

func newLobbyFeature() *Feature {
	f := DefaultFeature()
	f.Lobby.Identify = screen.Feature{Colors: "lobby", Sim: 0.95}
	f.Lobby.Reads.MedalTicket = screen.Region{X1: 1, Y1: 1, X2: 100, Y2: 30}
	f.Lobby.Reads.Trophy = screen.Region{X1: 1, Y1: 40, X2: 100, Y2: 70}
	f.Lobby.Reads.Refresh = screen.Region{X1: 1, Y1: 80, X2: 100, Y2: 110}
	f.Lobby.Reads.FreeRefresh = screen.Region{X1: 1, Y1: 120, X2: 100, Y2: 150}
	f.Lobby.Actions.FreeRefresh = screen.Point{X: 300, Y: 800}
	f.Lobby.Gestures.SwipeLeft = action.Swipe{
		From: action.Point{X: 1400, Y: 450}, To: action.Point{X: 200, Y: 450}, DurationMs: 300}
	return f
}

func TestIsLobby(t *testing.T) {
	d := &mockDetector{matchMulti: true}
	p := NewPage(d, &mockExecutor{}, newLobbyFeature())
	if !p.IsLobby() {
		t.Fatal("IsLobby should be true when MatchMultiColor returns true")
	}
	d.matchMulti = false
	if p.IsLobby() {
		t.Fatal("IsLobby should be false when MatchMultiColor returns false")
	}
}

func TestIsFreeRefresh(t *testing.T) {
	d := &mockDetector{ocrByKey: map[string]string{"1,120,100,150": "免费刷新"}}
	p := NewPage(d, &mockExecutor{}, newLobbyFeature())
	if !p.IsFreeRefresh() {
		t.Fatal("IsFreeRefresh should be true for '免费刷新'")
	}
	d.ocrByKey["1,120,100,150"] = "  刷新  "
	if p.IsFreeRefresh() {
		t.Fatal("IsFreeRefresh should be false for non-exact text")
	}
}

func TestReadMedalAndTicket(t *testing.T) {
	d := &mockDetector{ocrByKey: map[string]string{"1,1,100,30": "123 45"}}
	p := NewPage(d, &mockExecutor{}, newLobbyFeature())
	m, tk, ok := p.ReadMedalAndTicket()
	if !ok || m != 123 || tk != 45 {
		t.Fatalf("ReadMedalAndTicket = (%d,%d,%v), want (123,45,true)", m, tk, ok)
	}
	d.ocrByKey["1,1,100,30"] = "onlyone"
	if _, _, ok := p.ReadMedalAndTicket(); ok {
		t.Fatal("single token should be ok=false")
	}
	d.ocrByKey["1,1,100,30"] = "abc def"
	if _, _, ok := p.ReadMedalAndTicket(); ok {
		t.Fatal("non-numeric should be ok=false")
	}
}

func TestReadTrophyCount(t *testing.T) {
	d := &mockDetector{ocrByKey: map[string]string{"1,40,100,70": "  1024 "}}
	p := NewPage(d, &mockExecutor{}, newLobbyFeature())
	n, ok := p.ReadTrophyCount()
	if !ok || n != 1024 {
		t.Fatalf("ReadTrophyCount = (%d,%v), want (1024,true)", n, ok)
	}
	d.ocrByKey["1,40,100,70"] = "x"
	if _, ok := p.ReadTrophyCount(); ok {
		t.Fatal("non-numeric trophy should be ok=false")
	}
}

func TestReadRefreshCountdown(t *testing.T) {
	d := &mockDetector{ocrByKey: map[string]string{"1,80,100,110": "5分30秒"}}
	p := NewPage(d, &mockExecutor{}, newLobbyFeature())
	got, ok := p.ReadRefreshCountdown()
	if !ok || got != 330*time.Second {
		t.Fatalf("ReadRefreshCountdown = (%v,%v), want (330s,true)", got, ok)
	}
}

func TestTapFreeRefresh(t *testing.T) {
	e := &mockExecutor{}
	p := NewPage(&mockDetector{}, e, newLobbyFeature())
	p.TapFreeRefresh()
	if len(e.taps) != 1 || e.taps[0] != (action.Point{X: 300, Y: 800}) {
		t.Fatalf("TapFreeRefresh taps = %v, want [(300,800)]", e.taps)
	}
	if len(e.sleeps) == 0 {
		t.Fatal("TapFreeRefresh should sleep after tap")
	}
}

func TestSwipePageLeft(t *testing.T) {
	e := &mockExecutor{}
	p := NewPage(&mockDetector{}, e, newLobbyFeature())
	p.SwipePageLeft()
	if len(e.swipes) != 1 {
		t.Fatalf("SwipePageLeft swipes = %v, want 1 swipe", e.swipes)
	}
	if e.swipes[0][0] != (action.Point{X: 1400, Y: 450}) || e.swipes[0][1] != (action.Point{X: 200, Y: 450}) {
		t.Fatalf("SwipePageLeft swipe endpoints = %v", e.swipes[0])
	}
}

func newOpponentFeature() *Feature {
	f := DefaultFeature()
	op := &f.Lobby.Opponent
	op.Anchor = ColorFind{Region: screen.Region{X1: 0, Y1: 0, X2: 1600, Y2: 900}, Colors: "anchor", Sim: 0.95, Dir: 0}
	op.TrophyRect = screen.Region{X1: -10, Y1: -20, X2: 10, Y2: 20} // 相对锚点偏移
	op.ResultOffset = screen.Point{X: 30, Y: 0}
	op.ResultColors = ResultColors{Win: "w", Draw: "d", Lose: "l"}
	op.ResultSim = 0.9
	op.ClickOffset = screen.Point{X: 5, Y: 5}
	return f
}

func key(r screen.Region) string { return fmt.Sprintf("%d,%d,%d,%d", r.X1, r.Y1, r.X2, r.Y2) }

func TestFindFirstValidOpponent(t *testing.T) {
	cfg := &config.Arena{TrophyDiff: 50} // 区间 [950,1050] 当 myTrophy=1000

	t.Run("in-range hit returns first with offset Site", func(t *testing.T) {
		d := &mockDetector{
			anchors: []screen.Point{{X: 100, Y: 200}, {X: 100, Y: 400}},
			ocrByKey: map[string]string{
				key(screen.Region{X1: 90, Y1: 180, X2: 110, Y2: 220}): "1050", // anchor1 奖杯，区间内
			},
			matchByKey: map[string]bool{}, // 三色都不命中 → 未战
		}
		p := NewPage(d, &mockExecutor{}, newOpponentFeature())
		got := p.FindFirstValidOpponent(cfg, 1000)
		if got == nil {
			t.Fatal("want a valid opponent, got nil")
		}
		if got.Site != (action.Point{X: 105, Y: 205}) { // anchor(100,200)+ClickOffset(5,5)
			t.Errorf("Site = %+v, want (105,205)", got.Site)
		}
		if got.Trophies != 1050 || got.IsBattled {
			t.Errorf("Trophies/IsBattled = %d/%v, want 1050/false", got.Trophies, got.IsBattled)
		}
	})

	t.Run("battled card is skipped", func(t *testing.T) {
		d := &mockDetector{
			anchors: []screen.Point{{X: 100, Y: 200}, {X: 100, Y: 400}},
			ocrByKey: map[string]string{
				key(screen.Region{X1: 90, Y1: 180, X2: 110, Y2: 220}): "1000", // anchor1 区间内但已战
				key(screen.Region{X1: 90, Y1: 380, X2: 110, Y2: 420}): "1020", // anchor2 区间内未战
			},
			matchByKey: map[string]bool{
				"130,200,w": true, // anchor1 ResultOffset(100+30,200) 命中 Win → 已战
			},
		}
		p := NewPage(d, &mockExecutor{}, newOpponentFeature())
		got := p.FindFirstValidOpponent(cfg, 1000)
		if got == nil || got.Trophies != 1020 {
			t.Fatalf("want anchor2 (1020), got %+v", got)
		}
	})

	t.Run("out-of-range is skipped", func(t *testing.T) {
		d := &mockDetector{
			anchors:  []screen.Point{{X: 100, Y: 200}},
			ocrByKey: map[string]string{key(screen.Region{X1: 90, Y1: 180, X2: 110, Y2: 220}): "2000"}, // >1050
		}
		p := NewPage(d, &mockExecutor{}, newOpponentFeature())
		if got := p.FindFirstValidOpponent(cfg, 1000); got != nil {
			t.Fatalf("out-of-range should yield nil, got %+v", got)
		}
	})

	t.Run("no anchors returns nil", func(t *testing.T) {
		d := &mockDetector{anchors: nil}
		p := NewPage(d, &mockExecutor{}, newOpponentFeature())
		if got := p.FindFirstValidOpponent(cfg, 1000); got != nil {
			t.Fatalf("no anchors should yield nil, got %+v", got)
		}
	})

	t.Run("OCR failure skips that anchor", func(t *testing.T) {
		d := &mockDetector{
			anchors: []screen.Point{{X: 100, Y: 200}, {X: 100, Y: 400}},
			ocrByKey: map[string]string{
				key(screen.Region{X1: 90, Y1: 180, X2: 110, Y2: 220}): "abc",  // OCR 失败 → 跳过
				key(screen.Region{X1: 90, Y1: 380, X2: 110, Y2: 420}): "1010", // 区间内未战
			},
			matchByKey: map[string]bool{},
		}
		p := NewPage(d, &mockExecutor{}, newOpponentFeature())
		got := p.FindFirstValidOpponent(cfg, 1000)
		if got == nil || got.Trophies != 1010 {
			t.Fatalf("want anchor2 (1010), got %+v", got)
		}
	})

	t.Run("first in-range wins by anchor order", func(t *testing.T) {
		d := &mockDetector{
			anchors: []screen.Point{{X: 100, Y: 200}, {X: 100, Y: 400}}, // dir=0 已按 Y 排
			ocrByKey: map[string]string{
				key(screen.Region{X1: 90, Y1: 180, X2: 110, Y2: 220}): "960",  // 区间内（第一个）
				key(screen.Region{X1: 90, Y1: 380, X2: 110, Y2: 420}): "1040", // 也在区间内但应被前者抢先
			},
			matchByKey: map[string]bool{},
		}
		p := NewPage(d, &mockExecutor{}, newOpponentFeature())
		got := p.FindFirstValidOpponent(cfg, 1000)
		if got == nil || got.Trophies != 960 {
			t.Fatalf("want first anchor (960), got %+v", got)
		}
	})

	t.Run("zero diff means strict equality", func(t *testing.T) {
		eq := &config.Arena{TrophyDiff: 0} // 只接受奖杯==myTrophy
		d := &mockDetector{
			anchors: []screen.Point{{X: 100, Y: 200}, {X: 100, Y: 400}},
			ocrByKey: map[string]string{
				key(screen.Region{X1: 90, Y1: 180, X2: 110, Y2: 220}): "999",  // 不等 → 跳
				key(screen.Region{X1: 90, Y1: 380, X2: 110, Y2: 420}): "1000", // 严格相等 → 中
			},
			matchByKey: map[string]bool{},
		}
		p := NewPage(d, &mockExecutor{}, newOpponentFeature())
		got := p.FindFirstValidOpponent(eq, 1000)
		if got == nil || got.Trophies != 1000 {
			t.Fatalf("diff=0 should only match exact trophy, got %+v", got)
		}
	})
}
