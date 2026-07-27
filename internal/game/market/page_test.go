package market

import (
	"fmt"
	"image"
	"testing"
	"time"

	"app/internal/platform/action"
	"app/internal/platform/screen"
)

// ---- mock Detector ----
type mockDetector struct {
	matchMulti    map[string]bool           // key: colors 串
	ocrByKey      map[string]string         // key: "x1,y1,x2,y2"
	multiByColors map[string][]screen.Point // key: colors 串
	findColorOK   bool
	findColorPt   screen.Point
	findOCR       func(r screen.Region, keyword string) (screen.Point, bool)
}

func (m *mockDetector) Capture() *image.NRGBA { return nil }
func (m *mockDetector) MatchColor(x, y int, color string, sim float32) bool {
	return false
}
func (m *mockDetector) FindColor(r screen.Region, c string, s float32, d int) (screen.Point, bool) {
	return m.findColorPt, m.findColorOK
}
func (m *mockDetector) FindMultiColorsAll(r screen.Region, c string, s float32, d int) []screen.Point {
	return m.multiByColors[c]
}
func (m *mockDetector) MatchMultiColor(colors string, sim float32) bool {
	return m.matchMulti[colors]
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
	backs  int
	onTap  func(n int)
}

func (e *mockExecutor) Tap(p action.Point) {
	e.taps = append(e.taps, p)
	if e.onTap != nil {
		e.onTap(len(e.taps))
	}
}
func (e *mockExecutor) LongTap(p action.Point, ms int) {}
func (e *mockExecutor) Swipe(f, t action.Point, ms int) {
	e.swipes = append(e.swipes, [2]action.Point{f, t})
}
func (e *mockExecutor) Back()        { e.backs++ }
func (e *mockExecutor) Home()        {}
func (e *mockExecutor) Sleep(ms int) { e.sleeps = append(e.sleeps, ms) }

func key(r screen.Region) string { return fmt.Sprintf("%d,%d,%d,%d", r.X1, r.Y1, r.X2, r.Y2) }

func TestNewPage(t *testing.T) {
	p := NewPage(nil, nil, nil)
	if p.feature == nil || len(p.feature.Stock) != 7 {
		t.Fatal("nil feature should fall back to DefaultFeature")
	}
}

func TestSlotRects(t *testing.T) {
	p := NewPage(nil, nil, DefaultFeature())
	pt := screen.Point{X: 500, Y: 600}
	tap := p.slotTapRect(pt)
	wantTap := screen.Region{X1: 395, Y1: 686, X2: 605, Y2: 734} // cy=600+110, halfW=105, halfH=24
	if tap != wantTap {
		t.Errorf("slotTapRect = %+v, want %+v", tap, wantTap)
	}
	crate := p.slotCrateRect(pt)
	wantCrate := screen.Region{X1: 410, Y1: 515, X2: 590, Y2: 645} // cy=600-20, halfW=90, halfH=65
	if crate != wantCrate {
		t.Errorf("slotCrateRect = %+v, want %+v", crate, wantCrate)
	}
}

func TestIsFreeRefresh(t *testing.T) {
	d := &mockDetector{ocrByKey: map[string]string{key(DefaultFeature().Page.RefreshOcr): "免费刷新"}}
	p := NewPage(d, &mockExecutor{}, DefaultFeature())
	if !p.IsFreeRefresh() {
		t.Fatal("IsFreeRefresh should be true for '免费刷新'")
	}
	d.ocrByKey[key(DefaultFeature().Page.RefreshOcr)] = "01:00:00"
	if p.IsFreeRefresh() {
		t.Fatal("IsFreeRefresh should be false for countdown text")
	}
}

func TestReadRestockSeconds(t *testing.T) {
	ocrKey := key(DefaultFeature().Page.RefreshOcr)
	cases := []struct {
		in      string
		wantSec int
		wantOK  bool
	}{
		{"01:23:45", 5025, true},
		{"0:05:30", 330, true},
		{"免费刷新", 0, true},
		{"", 0, false},
		{"abc", 0, false},
	}
	for _, c := range cases {
		d := &mockDetector{ocrByKey: map[string]string{ocrKey: c.in}}
		p := NewPage(d, &mockExecutor{}, DefaultFeature())
		sec, _, ok := p.ReadRestockSeconds()
		if sec != c.wantSec || ok != c.wantOK {
			t.Errorf("ReadRestockSeconds(%q) = (%d,%v), want (%d,%v)", c.in, sec, ok, c.wantSec, c.wantOK)
		}
	}
}

func TestIsSlotSoldOut(t *testing.T) {
	f := DefaultFeature()
	pt := screen.Point{X: 500, Y: 600}
	crateKey := key(screen.Region{X1: 410, Y1: 515, X2: 590, Y2: 645})
	d := &mockDetector{ocrByKey: map[string]string{crateKey: "售罄"}}
	p := NewPage(d, &mockExecutor{}, f)
	if !p.IsSlotSoldOut(pt) {
		t.Fatal("IsSlotSoldOut should be true for '售罄'")
	}
	d.ocrByKey[crateKey] = "250 贝壳"
	if p.IsSlotSoldOut(pt) {
		t.Fatal("IsSlotSoldOut should be false for price text")
	}
}

func TestCollectVisibleTargetsDedupAndSort(t *testing.T) {
	f := DefaultFeature()
	f.Stock = map[string]ColorFind{
		"A": {Region: screen.Region{}, Colors: "cA", Sim: 0.9},
		"B": {Region: screen.Region{}, Colors: "cB", Sim: 0.9},
	}
	d := &mockDetector{multiByColors: map[string][]screen.Point{
		"cA": {{X: 400, Y: 600}, {X: 420, Y: 610}}, // 第二点距第一点 <80 → 去重
		"cB": {{X: 100, Y: 650}},
	}}
	p := NewPage(d, &mockExecutor{}, f)
	targets := p.collectVisibleTargets([]string{"A", "B", "X未知"})
	if len(targets) != 2 {
		t.Fatalf("targets = %d, want 2 (dedup)", len(targets))
	}
	if targets[0].Point.X != 100 || targets[1].Point.X != 400 {
		t.Fatalf("targets not sorted by X: %+v", targets)
	}
	if targets[0].Key != "B" || targets[1].Key != "A" {
		t.Fatalf("target keys = %q,%q", targets[0].Key, targets[1].Key)
	}
}

func shrinkDialogWaits(t *testing.T) {
	t.Helper()
	oldConfirm, oldResult := confirmDialogWait, purchaseResultWait
	confirmDialogWait, purchaseResultWait = 50*time.Millisecond, 50*time.Millisecond
	t.Cleanup(func() { confirmDialogWait, purchaseResultWait = oldConfirm, oldResult })
}

func TestTapShelfAndResolvePurchased(t *testing.T) {
	shrinkDialogWaits(t)
	f := DefaultFeature()
	d := &mockDetector{matchMulti: map[string]bool{f.Dialog.Identify.Colors: true}}
	e := &mockExecutor{}
	e.onTap = func(n int) {
		if n == 2 { // 货架点击后的确认点击 → 弹窗消失
			d.matchMulti[f.Dialog.Identify.Colors] = false
		}
	}
	p := NewPage(d, e, f)
	if got := p.TapShelfAndResolve(screen.Point{X: 500, Y: 600}); got != ResultPurchased {
		t.Fatalf("result = %q, want purchased", got)
	}
	if len(e.taps) != 2 {
		t.Fatalf("taps = %d, want 2 (shelf + confirm)", len(e.taps))
	}
}

func TestTapShelfAndResolveShortage(t *testing.T) {
	shrinkDialogWaits(t)
	f := DefaultFeature()
	d := &mockDetector{matchMulti: map[string]bool{f.Dialog.Identify.Colors: true}}
	e := &mockExecutor{}
	e.onTap = func(n int) {
		switch n {
		case 2: // 确认 → 道具不足弹窗出现
			d.matchMulti[f.Shortage.Identify.Colors] = true
		case 3: // 不足弹窗取消
			d.matchMulti[f.Shortage.Identify.Colors] = false
		case 4: // 确认弹窗关闭
			d.matchMulti[f.Dialog.Identify.Colors] = false
		}
	}
	p := NewPage(d, e, f)
	if got := p.TapShelfAndResolve(screen.Point{X: 500, Y: 600}); got != ResultShortage {
		t.Fatalf("result = %q, want shortage", got)
	}
	if len(e.taps) != 4 {
		t.Fatalf("taps = %d, want 4 (shelf+confirm+shortageCancel+dialogClose)", len(e.taps))
	}
}

func TestTapShelfAndResolveConfirmNotAppear(t *testing.T) {
	shrinkDialogWaits(t)
	f := DefaultFeature()
	d := &mockDetector{matchMulti: map[string]bool{}}
	p := NewPage(d, &mockExecutor{}, f)
	if got := p.TapShelfAndResolve(screen.Point{X: 500, Y: 600}); got != ResultFailed {
		t.Fatalf("result = %q, want failed", got)
	}
}

func TestIsLastPageAndSwipe(t *testing.T) {
	f := DefaultFeature()
	d := &mockDetector{findColorOK: true}
	e := &mockExecutor{}
	p := NewPage(d, e, f)
	if p.IsLastPage() {
		t.Fatal("arrow visible → not last page")
	}
	if !p.SwipeNextPage() {
		t.Fatal("SwipeNextPage should swipe when arrow visible")
	}
	if len(e.swipes) != 1 {
		t.Fatalf("swipes = %d, want 1", len(e.swipes))
	}
	d.findColorOK = false
	if !p.IsLastPage() {
		t.Fatal("arrow invisible → last page")
	}
	if p.SwipeNextPage() {
		t.Fatal("SwipeNextPage should refuse on last page")
	}
}

func TestPurchaseWishlist(t *testing.T) {
	shrinkDialogWaits(t)
	f := DefaultFeature()
	f.Stock = map[string]ColorFind{
		"A": {Region: screen.Region{}, Colors: "cA", Sim: 0.9},
		"B": {Region: screen.Region{}, Colors: "cB", Sim: 0.9},
	}
	// A 锚点 (500,600)：售罄；B 锚点 (800,600)：购买成功
	crateA := key(screen.Region{X1: 410, Y1: 515, X2: 590, Y2: 645})
	d := &mockDetector{
		matchMulti:    map[string]bool{f.Dialog.Identify.Colors: true},
		ocrByKey:      map[string]string{crateA: "售罄"},
		multiByColors: map[string][]screen.Point{"cA": {{X: 500, Y: 600}}, "cB": {{X: 800, Y: 600}}},
		findColorOK:   false, // 无右箭头 → 单页
	}
	e := &mockExecutor{}
	e.onTap = func(n int) {
		if n == 2 { // B 的确认点击 → 弹窗消失
			d.matchMulti[f.Dialog.Identify.Colors] = false
		}
	}
	p := NewPage(d, e, f)
	stats := p.PurchaseWishlist([]string{"A", "B", "X未知"})
	if stats.SoldOut != 1 || stats.Purchased != 1 || stats.Shortage != 0 || stats.Failed != 0 {
		t.Fatalf("stats = %+v, want {Purchased:1 SoldOut:1}", stats)
	}
	if len(e.swipes) != 0 {
		t.Fatalf("last page should not swipe, got %d", len(e.swipes))
	}
}

func TestPurchaseWishlistNoValidItems(t *testing.T) {
	f := DefaultFeature()
	f.Stock = map[string]ColorFind{}
	e := &mockExecutor{}
	p := NewPage(&mockDetector{}, e, f)
	stats := p.PurchaseWishlist([]string{"X未知"})
	if stats != (PurchaseStats{}) {
		t.Fatalf("stats = %+v, want zero", stats)
	}
	if len(e.taps) != 0 || len(e.swipes) != 0 {
		t.Fatal("no valid items → no taps/swipes")
	}
}
