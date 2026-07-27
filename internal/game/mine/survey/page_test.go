package survey

import (
	"fmt"
	"image"
	"testing"

	"app/internal/platform/action"
	"app/internal/platform/screen"
)

// ---- mock Detector ----
type mockDetector struct {
	matchByColors map[string]bool
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
	return nil
}
func (m *mockDetector) MatchMultiColor(colors string, sim float32) bool {
	return m.matchByColors[colors]
}
func (m *mockDetector) MatchImage(r screen.Region, t []byte, s float32) (screen.Point, bool) {
	return screen.Point{}, false
}
func (m *mockDetector) OCRText(r screen.Region) (string, error) {
	return m.ocrByKey[fmt.Sprintf("%d,%d,%d,%d", r.X1, r.Y1, r.X2, r.Y2)], nil
}
func (m *mockDetector) FindOCRText(r screen.Region, keyword string) (screen.Point, bool) {
	return screen.Point{}, false
}

// ---- mock Executor ----
type mockExecutor struct {
	taps []action.Point
}

func (e *mockExecutor) Tap(p action.Point)              { e.taps = append(e.taps, p) }
func (e *mockExecutor) LongTap(p action.Point, ms int)  {}
func (e *mockExecutor) Swipe(f, t action.Point, ms int) {}
func (e *mockExecutor) Back()                           {}
func (e *mockExecutor) Home()                           {}
func (e *mockExecutor) Sleep(ms int)                    {}

func TestParseFirstInt(t *testing.T) {
	cases := []struct {
		in     string
		wantN  int
		wantOK bool
	}{
		{"12", 12, true},
		{"当前 6 层", 6, true},
		{"1,234", 1234, true},
		{"", 0, false},
		{"层数未知", 0, false},
	}
	for _, c := range cases {
		n, ok := parseFirstInt(c.in)
		if n != c.wantN || ok != c.wantOK {
			t.Errorf("parseFirstInt(%q) = (%d,%v), want (%d,%v)", c.in, n, ok, c.wantN, c.wantOK)
		}
	}
}

func floorKey(f *Feature) string {
	r := f.Running.FloorOCR
	return fmt.Sprintf("%d,%d,%d,%d", r.X1, r.Y1, r.X2, r.Y2)
}

func TestGetCurrentFloor(t *testing.T) {
	f := DefaultFeature()
	d := &mockDetector{ocrByKey: map[string]string{floorKey(f): "当前 7 层"}}
	p := NewPage(d, &mockExecutor{}, f)
	n, ok := p.GetCurrentFloor()
	if !ok || n != 7 {
		t.Fatalf("GetCurrentFloor = (%d,%v), want (7,true)", n, ok)
	}
	d.ocrByKey[floorKey(f)] = "??"
	if _, ok := p.GetCurrentFloor(); ok {
		t.Fatal("unparsable floor should be ok=false")
	}
}

func TestIsDomain(t *testing.T) {
	f := DefaultFeature()
	d := &mockDetector{matchByColors: map[string]bool{}}
	p := NewPage(d, &mockExecutor{}, f)
	if p.IsDomain() {
		t.Fatal("no feature matched → not in domain")
	}
	d.matchByColors[f.Running.Identify.Colors] = true
	if !p.IsDomain() || !p.IsRunning() {
		t.Fatal("running feature matched → in domain and running")
	}
}

// Setup 全流程特征依次命中（mock 对全部特征返回 true）→ 成功。
func TestSetupSuccess(t *testing.T) {
	f := DefaultFeature()
	d := &mockDetector{matchByColors: map[string]bool{
		f.Ready.Identify.Colors:                 true,
		f.Dialogs.Info.Identify.Colors:          true,
		f.Dialogs.ConfirmCookie.Identify.Colors: true,
		f.Running.Identify.Colors:               true,
	}}
	e := &mockExecutor{}
	p := NewPage(d, e, f)
	if !p.Setup() {
		t.Fatal("Setup should succeed when every stage matches")
	}
	if len(e.taps) != 4 {
		t.Fatalf("Setup should tap 4 times (autoSelect/start/info/cookie), got %d", len(e.taps))
	}
}

// StopVenture：停止确认弹窗命中 → 成功；结算等待结果不影响返回值（与 Lua 一致）。
func TestStopVenture(t *testing.T) {
	f := DefaultFeature()
	d := &mockDetector{matchByColors: map[string]bool{
		f.Dialogs.Stop.Identify.Colors: true,
		f.Setup.Identify.Colors:        true, // 结算后立即回到 setup 页，tapUntilMatch 不空等
	}}
	e := &mockExecutor{}
	p := NewPage(d, e, f)
	if !p.StopVenture() {
		t.Fatal("StopVenture should succeed when stop dialog matches")
	}
	if len(e.taps) < 2 {
		t.Fatalf("StopVenture should tap stop + confirm, got %d taps", len(e.taps))
	}
}
