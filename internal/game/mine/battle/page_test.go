package battle

import (
	"image"
	"testing"

	"app/internal/platform/action"
	"app/internal/platform/screen"
)

// ---- mock Detector ----
type mockDetector struct {
	matchByColors map[string]bool
	findAllFunc   func(r screen.Region, colors string) []screen.Point
	ocrByKey      map[string]string
}

func (m *mockDetector) Capture() *image.NRGBA { return nil }
func (m *mockDetector) MatchColor(x, y int, color string, sim float32) bool {
	return false
}
func (m *mockDetector) FindColor(r screen.Region, c string, s float32, d int) (screen.Point, bool) {
	return screen.Point{}, false
}
func (m *mockDetector) FindMultiColorsAll(r screen.Region, c string, s float32, d int) []screen.Point {
	if m.findAllFunc != nil {
		return m.findAllFunc(r, c)
	}
	return nil
}
func (m *mockDetector) MatchMultiColor(colors string, sim float32) bool {
	return m.matchByColors[colors]
}
func (m *mockDetector) MatchImage(r screen.Region, t []byte, s float32) (screen.Point, bool) {
	return screen.Point{}, false
}
func (m *mockDetector) OCRText(r screen.Region) (string, error) {
	return m.ocrByKey[regionKey(r)], nil
}
func (m *mockDetector) FindOCRText(r screen.Region, keyword string) (screen.Point, bool) {
	return screen.Point{}, false
}

func regionKey(r screen.Region) string {
	return string(rune(r.X1)) + string(rune(r.Y1)) + string(rune(r.X2)) + string(rune(r.Y2))
}

// ---- mock Executor ----
type mockExecutor struct {
	taps   []action.Point
	swipes [][2]action.Point
}

func (e *mockExecutor) Tap(p action.Point)             { e.taps = append(e.taps, p) }
func (e *mockExecutor) LongTap(p action.Point, ms int) {}
func (e *mockExecutor) Swipe(f, t action.Point, ms int) {
	e.swipes = append(e.swipes, [2]action.Point{f, t})
}
func (e *mockExecutor) Back()        {}
func (e *mockExecutor) Home()        {}
func (e *mockExecutor) Sleep(ms int) {}

func TestParseClockCount(t *testing.T) {
	cases := []struct {
		in          string
		used, owned int
		ok          bool
	}{
		{"1/12,611", 1, 12611, true},
		{"3/50", 3, 50, true},
		{"0 / 10", 0, 10, true},
		{"", 0, 0, false},
		{"abc", 0, 0, false},
	}
	for _, c := range cases {
		used, owned, ok := parseClockCount(c.in)
		if used != c.used || owned != c.owned || ok != c.ok {
			t.Errorf("parseClockCount(%q) = (%d,%d,%v), want (%d,%d,%v)", c.in, used, owned, ok, c.used, c.owned, c.ok)
		}
	}
}

func TestRecognizeSoulStoneType(t *testing.T) {
	f := DefaultFeature()
	targets := map[string]bool{"妖精王": true, "莓果": true}

	// 命中唯一目标
	d := &mockDetector{findAllFunc: func(r screen.Region, colors string) []screen.Point {
		if colors == f.SoulStones[0].Stones["妖精王"].Colors {
			return []screen.Point{{X: 300, Y: 650}}
		}
		return nil
	}}
	p := NewPage(d, &mockExecutor{}, f)
	if got := p.RecognizeSoulStoneType(targets); got != "妖精王" {
		t.Fatalf("match = %q, want 妖精王", got)
	}

	// 多个目标同时命中 → 无法区分
	d2 := &mockDetector{findAllFunc: func(r screen.Region, colors string) []screen.Point {
		return []screen.Point{{X: 300, Y: 650}}
	}}
	p2 := NewPage(d2, &mockExecutor{}, f)
	if got := p2.RecognizeSoulStoneType(targets); got != "" {
		t.Fatalf("ambiguous match should return empty, got %q", got)
	}

	// 非目标命中不影响
	if got := p2.RecognizeSoulStoneType(map[string]bool{"辣椒素": true}); got != "辣椒素" {
		t.Fatalf("single target match = %q, want 辣椒素", got)
	}

	// 空目标集合
	if got := p.RecognizeSoulStoneType(nil); got != "" {
		t.Fatalf("nil targets should return empty, got %q", got)
	}
}

func TestReadClockCount(t *testing.T) {
	f := DefaultFeature()
	d := &mockDetector{ocrByKey: map[string]string{
		regionKey(f.QuickBattleDialog.ClockCountOCR): "1/12,611",
	}}
	p := NewPage(d, &mockExecutor{}, f)
	used, owned, ok := p.ReadClockCount()
	if !ok || used != 1 || owned != 12611 {
		t.Fatalf("ReadClockCount = (%d,%d,%v), want (1,12611,true)", used, owned, ok)
	}
}

func TestSwipeUpAndCheckLastPage(t *testing.T) {
	f := DefaultFeature()
	lastPage := f.LastPage
	d := &mockDetector{findAllFunc: func(r screen.Region, colors string) []screen.Point {
		if colors == lastPage.Colors {
			return []screen.Point{{X: 600, Y: 550}}
		}
		return nil
	}}
	e := &mockExecutor{}
	p := NewPage(d, e, f)
	if !p.SwipeUpAndCheckLastPage() {
		t.Fatal("should detect last page")
	}
	if len(e.swipes) != 1 {
		t.Fatalf("expected 1 swipe, got %d", len(e.swipes))
	}
}
