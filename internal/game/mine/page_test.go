package mine

import (
	"fmt"
	"image"
	"testing"
	"time"

	"app/internal/game/common/kingdom"
	"app/internal/platform/action"
	"app/internal/platform/screen"
)

// ---- mock Detector ----
type mockDetector struct {
	matchByColors map[string]bool // key: colors string
	anchors       []screen.Point
	ocrByKey      map[string]string
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
	return m.anchors
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
	if keyword == "" || m.findOCR == nil {
		return screen.Point{}, false
	}
	return m.findOCR(r, keyword)
}

// ---- mock Executor ----
type mockExecutor struct {
	taps   []action.Point
	swipes [][2]action.Point
	sleeps []int
}

func (e *mockExecutor) Tap(p action.Point)             { e.taps = append(e.taps, p) }
func (e *mockExecutor) LongTap(p action.Point, ms int) {}
func (e *mockExecutor) Swipe(f, t action.Point, ms int) {
	e.swipes = append(e.swipes, [2]action.Point{f, t})
}
func (e *mockExecutor) Back()        {}
func (e *mockExecutor) Home()        {}
func (e *mockExecutor) Sleep(ms int) { e.sleeps = append(e.sleeps, ms) }

func homeColors(f *Feature) string { return f.Home.Identify.Colors }

func TestPageIsCurrent(t *testing.T) {
	f := DefaultFeature()
	d := &mockDetector{matchByColors: map[string]bool{homeColors(f): true}}
	p := NewPage(d, &mockExecutor{}, f)
	if !p.IsCurrent() {
		t.Fatal("IsCurrent should be true when identify matches")
	}
	d.matchByColors[homeColors(f)] = false
	if p.IsCurrent() {
		t.Fatal("IsCurrent should be false when identify misses")
	}
}

func TestPageHasCompletedMiningTask(t *testing.T) {
	f := DefaultFeature()
	d := &mockDetector{matchByColors: map[string]bool{f.Home.CompletedTaskIdentify.Colors: true}}
	p := NewPage(d, &mockExecutor{}, f)
	if !p.HasCompletedMiningTask() {
		t.Fatal("HasCompletedMiningTask should be true when badge matches")
	}
}

func TestPageWaitGone(t *testing.T) {
	f := DefaultFeature()
	d := &mockDetector{matchByColors: map[string]bool{}}
	p := NewPage(d, &mockExecutor{}, f)
	if !p.WaitGone(50 * time.Millisecond) {
		t.Fatal("WaitGone should return immediately when already gone")
	}
	d.matchByColors[homeColors(f)] = true
	if p.WaitGone(50 * time.Millisecond) {
		t.Fatal("WaitGone should time out while page stays current")
	}
}

func TestPageTapActions(t *testing.T) {
	f := DefaultFeature()
	e := &mockExecutor{}
	p := NewPage(&mockDetector{}, e, f)
	p.TapVenture()
	p.TapMining()
	p.TapBattle()
	p.TapJellyEntry()
	p.TapBack()
	p.TapEntryMine()
	if len(e.taps) != 6 {
		t.Fatalf("expected 6 taps, got %d", len(e.taps))
	}
	if len(e.sleeps) != 6 {
		t.Fatalf("every tap action should sleep, got %d sleeps", len(e.sleeps))
	}
}

// ---- route tests ----

type fakeMiningRoutePage struct {
	mining, reward, settlement bool
	backCalls                  int
}

func (f *fakeMiningRoutePage) IsMiningPage() bool      { return f.mining }
func (f *fakeMiningRoutePage) IsRewardPage() bool      { return f.reward }
func (f *fakeMiningRoutePage) IsSettlementRoute() bool { return f.settlement }
func (f *fakeMiningRoutePage) TapBackBtn()             { f.backCalls++ }

func newKingdomPage(d screen.Detector, e action.Executor) *kingdom.Page {
	return kingdom.NewPage(d, e, kingdom.DefaultFeature())
}

func kingdomColors() string { return kingdom.DefaultFeature().Home.Identify.Colors }

func TestRouteKingdomHomeToMineHome(t *testing.T) {
	f := DefaultFeature()
	d := &mockDetector{matchByColors: map[string]bool{
		kingdomColors(): true,
	}}
	home := NewPage(d, &mockExecutor{}, f)
	kp := newKingdomPage(d, &mockExecutor{})
	r := NewRoute(home, kp)

	// 点完入口后矿山首页出现
	d.matchByColors[homeColors(f)] = true
	if !r.KingdomHomeToMineHome() {
		t.Fatal("should reach mine home")
	}

	// 不在王国首页且不在矿山首页 → 失败
	d2 := &mockDetector{matchByColors: map[string]bool{}}
	r2 := NewRoute(NewPage(d2, &mockExecutor{}, f), newKingdomPage(d2, &mockExecutor{}))
	if r2.KingdomHomeToMineHome() {
		t.Fatal("should fail when neither kingdom home nor mine home")
	}
}

func TestRouteMineHomeToKingdom(t *testing.T) {
	f := DefaultFeature()
	d := &mockDetector{matchByColors: map[string]bool{kingdomColors(): true}}
	r := NewRoute(NewPage(d, &mockExecutor{}, f), newKingdomPage(d, &mockExecutor{}))
	if !r.MineHomeToKingdom() {
		t.Fatal("already in kingdom home should succeed")
	}
}

func TestRouteReturnToKingdomFromMiningPage(t *testing.T) {
	f := DefaultFeature()
	d := &mockDetector{matchByColors: map[string]bool{
		kingdomColors(): true, // 开采页返回后直接在王国首页（简化路径）
	}}
	r := NewRoute(NewPage(d, &mockExecutor{}, f), newKingdomPage(d, &mockExecutor{}))
	if !r.ReturnToKingdom(&fakeMiningRoutePage{}) {
		t.Fatal("already in kingdom home should succeed without touching mining page")
	}

	// 在开采页：先 TapBackBtn 回矿山首页，再回王国
	d2 := &mockDetector{matchByColors: map[string]bool{
		homeColors(f):   true, // TapBack 后一直在矿山首页
		kingdomColors(): false,
	}}
	mp := &fakeMiningRoutePage{mining: true}
	r2 := NewRoute(NewPage(d2, &mockExecutor{}, f), newKingdomPage(d2, &mockExecutor{}))
	r2.kingdomWaitTimeout = 20 * time.Millisecond // 失败路径不等满 90s
	// 矿山首页 TapBack 后王国首页仍识别不到 → 失败
	if r2.ReturnToKingdom(mp) {
		t.Fatal("should fail when kingdom home never appears")
	}
	if mp.backCalls != 1 {
		t.Fatalf("expected mining page TapBackBtn once, got %d", mp.backCalls)
	}
}

func TestRouteReturnToKingdomUnknownPage(t *testing.T) {
	f := DefaultFeature()
	d := &mockDetector{matchByColors: map[string]bool{}}
	r := NewRoute(NewPage(d, &mockExecutor{}, f), newKingdomPage(d, &mockExecutor{}))
	if r.ReturnToKingdom(&fakeMiningRoutePage{}) {
		t.Fatal("unknown page should fail")
	}
}
