package kingdom

import (
	"image"
	"testing"
	"time"

	"app/internal/platform/action"
	"app/internal/platform/screen"
)

type mockDetector struct {
	matchColors map[string]bool
	findOCR     func(r screen.Region, keyword string) (screen.Point, bool)
}

func (m *mockDetector) Capture() *image.NRGBA { return nil }
func (m *mockDetector) MatchColor(x, y int, color string, sim float32) bool {
	return false
}
func (m *mockDetector) FindColor(region screen.Region, color string, sim float32, dir int) (screen.Point, bool) {
	return screen.Point{}, false
}
func (m *mockDetector) FindMultiColorsAll(region screen.Region, colors string, sim float32, dir int) []screen.Point {
	return nil
}
func (m *mockDetector) MatchMultiColor(colors string, sim float32) bool {
	return m.matchColors[colors]
}
func (m *mockDetector) MatchImage(region screen.Region, template []byte, sim float32) (screen.Point, bool) {
	return screen.Point{}, false
}
func (m *mockDetector) OCRText(region screen.Region) (string, error) { return "", nil }
func (m *mockDetector) FindOCRText(region screen.Region, keyword string) (screen.Point, bool) {
	if m.findOCR != nil {
		return m.findOCR(region, keyword)
	}
	return screen.Point{}, false
}

type mockExecutor struct {
	taps []action.Point
}

func (m *mockExecutor) Tap(p action.Point) {
	m.taps = append(m.taps, p)
}
func (m *mockExecutor) LongTap(p action.Point, ms int)              {}
func (m *mockExecutor) Swipe(from, to action.Point, durationMs int) {}
func (m *mockExecutor) Sleep(ms int)                                {}
func (m *mockExecutor) Back()                                       {}
func (m *mockExecutor) Home()                                       {}

func TestIsKingdomHomeEmptyColorsFalse(t *testing.T) {
	p := NewPage(&mockDetector{matchColors: map[string]bool{"": true}}, &mockExecutor{}, &Feature{})
	if p.IsKingdomHome() {
		t.Fatal("empty Colors must not match as home")
	}
}

func TestIsKingdomHomeUsesFeature(t *testing.T) {
	f := &Feature{Home: PageSlot{Identify: screen.Feature{Colors: "home", Sim: 0.9}}}
	det := &mockDetector{matchColors: map[string]bool{"home": true}}
	p := NewPage(det, &mockExecutor{}, f)
	if !p.IsKingdomHome() {
		t.Fatal("want home true")
	}
	det.matchColors["home"] = false
	if p.IsKingdomHome() {
		t.Fatal("want home false")
	}
}

func TestTapAdventureBtn(t *testing.T) {
	exec := &mockExecutor{}
	f := &Feature{Actions: NavActions{AdventureBtn: screen.Region{X1: 200, Y1: 800, X2: 200, Y2: 800}}}
	p := NewPage(&mockDetector{}, exec, f)
	p.TapAdventureBtn()
	if len(exec.taps) != 1 || exec.taps[0] != (action.Point{X: 200, Y: 800}) {
		t.Fatalf("taps=%v", exec.taps)
	}
}

func TestTapAdventureBtnUnconfiguredNoTap(t *testing.T) {
	exec := &mockExecutor{}
	p := NewPage(&mockDetector{}, exec, &Feature{})
	p.TapAdventureBtn()
	if len(exec.taps) != 0 {
		t.Fatalf("expected no tap, got %v", exec.taps)
	}
}

func TestIsEventPageOCR(t *testing.T) {
	det := &mockDetector{
		findOCR: func(r screen.Region, keyword string) (screen.Point, bool) {
			if keyword == "王国活动" && r.X1 == 681 {
				return screen.Point{X: 700, Y: 200}, true
			}
			return screen.Point{}, false
		},
	}
	f := &Feature{Event: OCRPage{
		Region:  screen.Region{X1: 681, Y1: 171, X2: 915, Y2: 242},
		Keyword: "王国活动",
	}}
	p := NewPage(det, &mockExecutor{}, f)
	if !p.IsEventPage() {
		t.Fatal("want event page when OCR hits")
	}
	f.Event.Keyword = ""
	if p.IsEventPage() {
		t.Fatal("empty keyword must be false")
	}
}

func TestWaitAdventure(t *testing.T) {
	det := &mockDetector{
		findOCR: func(r screen.Region, keyword string) (screen.Point, bool) {
			return screen.Point{}, keyword == "冒险"
		},
	}
	f := &Feature{Adventure: OCRPage{
		Region:  screen.Region{X1: 89, Y1: 28, X2: 181, Y2: 76},
		Keyword: "冒险",
	}}
	p := NewPage(det, &mockExecutor{}, f)
	if !p.WaitAdventure(50 * time.Millisecond) {
		t.Fatal("want adventure")
	}
}
