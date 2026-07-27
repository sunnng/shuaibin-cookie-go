package starlight

import (
	"image"
	"testing"
	"time"

	"app/internal/platform/action"
	"app/internal/platform/screen"
)

// ---- mock Detector ----
type mockDetector struct {
	matchMulti bool
	anchors    []screen.Point
}

func (m *mockDetector) Capture() *image.NRGBA { return nil }
func (m *mockDetector) MatchColor(x, y int, color string, sim float32) bool {
	return false
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
func (m *mockDetector) OCRText(r screen.Region) (string, error) { return "", nil }
func (m *mockDetector) FindOCRText(r screen.Region, keyword string) (screen.Point, bool) {
	return screen.Point{}, false
}

// ---- mock Executor ----
type mockExecutor struct {
	taps   []action.Point
	sleeps []int
}

func (e *mockExecutor) Tap(p action.Point)              { e.taps = append(e.taps, p) }
func (e *mockExecutor) LongTap(p action.Point, ms int)  {}
func (e *mockExecutor) Swipe(f, t action.Point, ms int) {}
func (e *mockExecutor) Back()                           {}
func (e *mockExecutor) Home()                           {}
func (e *mockExecutor) Sleep(ms int)                    { e.sleeps = append(e.sleeps, ms) }

func TestNewPage(t *testing.T) {
	_ = NewPage(nil, nil, nil) // nil feature → DefaultFeature
	_ = NewPage(nil, nil, DefaultFeature())
}

func TestIsHomePage(t *testing.T) {
	d := &mockDetector{matchMulti: true}
	p := NewPage(d, &mockExecutor{}, DefaultFeature())
	if !p.IsHomePage() {
		t.Fatal("IsHomePage should be true when MatchMultiColor returns true")
	}
	d.matchMulti = false
	if p.IsHomePage() {
		t.Fatal("IsHomePage should be false when MatchMultiColor returns false")
	}
}

func TestIsPageUnconfigured(t *testing.T) {
	f := &Feature{} // 全部 Colors 为空
	p := NewPage(&mockDetector{matchMulti: true}, &mockExecutor{}, f)
	if p.IsHomePage() || p.IsManualPage() || p.IsVanillaIslandPage() || p.IsTaskPage() || p.IsEventPage() {
		t.Fatal("empty identify must not match any page")
	}
}

func TestWaitPageTimeout(t *testing.T) {
	d := &mockDetector{matchMulti: false}
	p := NewPage(d, &mockExecutor{}, DefaultFeature())
	if p.WaitHomePage(time.Millisecond) {
		t.Fatal("WaitHomePage should time out when feature never matches")
	}
	if p.WaitManualPage(time.Millisecond) {
		t.Fatal("WaitManualPage should time out when feature never matches")
	}
	if p.WaitVanillaIslandPage(time.Millisecond) {
		t.Fatal("WaitVanillaIslandPage should time out when feature never matches")
	}
	if p.WaitTaskPage(time.Millisecond) {
		t.Fatal("WaitTaskPage should time out when feature never matches")
	}
	if p.WaitEventPage(time.Millisecond) {
		t.Fatal("WaitEventPage should time out when feature never matches")
	}
}

func TestTapSailingManual(t *testing.T) {
	e := &mockExecutor{}
	p := NewPage(&mockDetector{}, e, DefaultFeature())
	if !p.TapSailingManual() {
		t.Fatal("want TapSailingManual true")
	}
	if len(e.taps) != 1 {
		t.Fatalf("want 1 tap, got %v", e.taps)
	}
	pt := e.taps[0]
	r := DefaultFeature().Home.Actions.ManualBtn
	if pt.X < r.X1 || pt.X > r.X2 || pt.Y < r.Y1 || pt.Y > r.Y2 {
		t.Fatalf("tap %+v outside ManualBtn region %+v", pt, r)
	}
	if len(e.sleeps) == 0 || e.sleeps[0] != 1000 {
		t.Fatalf("want 1000ms sleep after tap, got %v", e.sleeps)
	}
}

func TestTapUnconfiguredRegion(t *testing.T) {
	e := &mockExecutor{}
	p := NewPage(&mockDetector{}, e, &Feature{})
	if p.TapSailingManual() || p.TapTaskBtn() || p.TapBackToKingdom() ||
		p.TapLoginIsland() || p.TapBackFromVanilla() || p.TapBackFromTask() || p.TapStarlightEntry() {
		t.Fatal("unconfigured regions must return false")
	}
	if len(e.taps) != 0 {
		t.Fatalf("no taps expected, got %v", e.taps)
	}
}

func TestFindClaimableBtn(t *testing.T) {
	d := &mockDetector{anchors: []screen.Point{{X: 500, Y: 300}, {X: 500, Y: 500}}}
	p := NewPage(d, &mockExecutor{}, DefaultFeature())
	pt, ok := p.FindClaimableBtn()
	if !ok || pt != (screen.Point{X: 500, Y: 300}) {
		t.Fatalf("FindClaimableBtn = (%+v,%v), want first anchor", pt, ok)
	}
	d.anchors = nil
	if _, ok := p.FindClaimableBtn(); ok {
		t.Fatal("no anchors should yield ok=false")
	}
}

func TestFindClaimableBtnUnconfigured(t *testing.T) {
	p := NewPage(&mockDetector{}, &mockExecutor{}, &Feature{})
	if _, ok := p.FindClaimableBtn(); ok {
		t.Fatal("empty claim colors must yield ok=false")
	}
}

func TestTapClaimableBtn(t *testing.T) {
	e := &mockExecutor{}
	p := NewPage(&mockDetector{}, e, DefaultFeature())
	p.TapClaimableBtn(screen.Point{X: 500, Y: 300})
	if len(e.taps) != 1 || e.taps[0] != (action.Point{X: 500, Y: 300}) {
		t.Fatalf("taps = %v, want [(500,300)]", e.taps)
	}
	if len(e.sleeps) == 0 || e.sleeps[0] != 800 {
		t.Fatalf("want 800ms sleep, got %v", e.sleeps)
	}
}

func TestSettleAfterClaim(t *testing.T) {
	e := &mockExecutor{}
	p := NewPage(&mockDetector{}, e, DefaultFeature())
	calls := 0
	p.SettleAfterClaim(func() bool { calls++; return false })
	if calls != 4 {
		t.Fatalf("want guard check 4 times, got %d", calls)
	}
	if len(e.sleeps) != 4 {
		t.Fatalf("want 4 sleeps of 500ms, got %v", e.sleeps)
	}
	for _, s := range e.sleeps {
		if s != 500 {
			t.Fatalf("want 500ms chunks, got %v", e.sleeps)
		}
	}
}

func TestDismissRewardPopupNotNeeded(t *testing.T) {
	d := &mockDetector{matchMulti: true} // 任务页特征仍在 → 无需关闭弹窗
	e := &mockExecutor{}
	p := NewPage(d, e, DefaultFeature())
	p.DismissRewardPopupIfNeeded()
	if len(e.taps) != 0 {
		t.Fatalf("no dismiss tap expected on task page, got %v", e.taps)
	}
}

func TestDismissRewardPopup(t *testing.T) {
	d := &mockDetector{matchMulti: false} // 任务页特征消失 → 点中央关闭
	e := &dismissExec{det: d}             // 点击后任务页特征恢复，避免 5s 等待
	p := NewPage(d, e, DefaultFeature())
	p.DismissRewardPopupIfNeeded()
	if len(e.taps) != 1 || e.taps[0] != (action.Point{X: 800, Y: 450}) {
		t.Fatalf("want dismiss tap at (800,450), got %v", e.taps)
	}
}

// dismissExec 点击后让任务页特征恢复（模拟弹窗被关闭）。
type dismissExec struct {
	mockExecutor
	det *mockDetector
}

func (e *dismissExec) Tap(pt action.Point) {
	e.taps = append(e.taps, pt)
	e.det.matchMulti = true
}
