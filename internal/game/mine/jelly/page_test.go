package jelly

import (
	"image"
	"testing"
	"time"

	"app/internal/platform/action"
	"app/internal/platform/screen"
)

// ---- mock Detector ----
type mockDetector struct {
	matchByColors map[string]bool
	ocrText       string
	findOCR       func(r screen.Region, keyword string) (screen.Point, bool)
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
func (m *mockDetector) OCRText(r screen.Region) (string, error) { return m.ocrText, nil }
func (m *mockDetector) FindOCRText(r screen.Region, keyword string) (screen.Point, bool) {
	if keyword == "" || m.findOCR == nil {
		return screen.Point{}, false
	}
	return m.findOCR(r, keyword)
}

// ---- mock Executor ----
type mockExecutor struct{ taps []action.Point }

func (e *mockExecutor) Tap(p action.Point)              { e.taps = append(e.taps, p) }
func (e *mockExecutor) LongTap(p action.Point, ms int)  {}
func (e *mockExecutor) Swipe(f, t action.Point, ms int) {}
func (e *mockExecutor) Back()                           {}
func (e *mockExecutor) Home()                           {}
func (e *mockExecutor) Sleep(ms int)                    {}

func TestParseRemainTime(t *testing.T) {
	cases := []struct {
		in   string
		want time.Duration
		ok   bool
	}{
		{"1天2小时3分钟4秒", 86400 + 2*3600 + 3*60 + 4, true},
		{"2小时30分钟", 2*3600 + 30*60, true},
		{"45分钟", 45 * 60, true},
		{"剩余 10秒", 10, true},
		{"", 0, false},
		{"暂无", 0, false},
		{"0秒", 0, false},
	}
	for _, c := range cases {
		got, ok := parseRemainTime(c.in)
		want := time.Duration(c.want) * time.Second
		if got != want || ok != c.ok {
			t.Errorf("parseRemainTime(%q) = (%v,%v), want (%v,%v)", c.in, got, ok, want, c.ok)
		}
	}
}

func TestReadRemainTime(t *testing.T) {
	p := NewPage(&mockDetector{ocrText: "1小时23分钟"}, &mockExecutor{}, DefaultFeature())
	got, ok := p.ReadRemainTime()
	if !ok || got != (3600+23*60)*time.Second {
		t.Fatalf("ReadRemainTime = (%v,%v)", got, ok)
	}
	p2 := NewPage(&mockDetector{ocrText: "???"}, &mockExecutor{}, DefaultFeature())
	if _, ok := p2.ReadRemainTime(); ok {
		t.Fatal("unparsable text should be ok=false")
	}
}

func TestFindConfigBtn(t *testing.T) {
	d := &mockDetector{findOCR: func(r screen.Region, kw string) (screen.Point, bool) {
		if kw == "配置" {
			return screen.Point{X: 800, Y: 620}, true
		}
		return screen.Point{}, false
	}}
	p := NewPage(d, &mockExecutor{}, DefaultFeature())
	pt, ok := p.FindConfigBtn()
	if !ok || pt.X != 800 {
		t.Fatalf("FindConfigBtn = (%+v,%v)", pt, ok)
	}
	d.findOCR = nil
	if _, ok := p.FindConfigBtn(); ok {
		t.Fatal("should miss when OCR finds nothing")
	}
}

func TestCanClaimAllAndChoose(t *testing.T) {
	f := DefaultFeature()
	d := &mockDetector{matchByColors: map[string]bool{
		f.ClaimAllIdentify.Colors:         true,
		f.Config.CanChooseIdentify.Colors: true,
	}}
	p := NewPage(d, &mockExecutor{}, f)
	if !p.CanClaimAll() {
		t.Fatal("CanClaimAll should be true when feature matches")
	}
	if !p.CanChoose() {
		t.Fatal("CanChoose should be true when feature matches")
	}
}
