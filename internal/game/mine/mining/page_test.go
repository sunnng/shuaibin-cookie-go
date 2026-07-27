package mining

import (
	"fmt"
	"image"
	"testing"

	"app/internal/game/mine"
	"app/internal/platform/action"
	"app/internal/platform/screen"
)

// ---- mock Detector ----
type mockDetector struct {
	matchByColors map[string]bool
	anchors       []screen.Point
	ocrFunc       func(r screen.Region) string
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
	if m.ocrFunc != nil {
		return m.ocrFunc(r), nil
	}
	return "", nil
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
	onTap  func()
}

func (e *mockExecutor) Tap(p action.Point) {
	e.taps = append(e.taps, p)
	if e.onTap != nil {
		e.onTap()
	}
}
func (e *mockExecutor) LongTap(p action.Point, ms int) {}
func (e *mockExecutor) Swipe(f, t action.Point, ms int) {
	e.swipes = append(e.swipes, [2]action.Point{f, t})
}
func (e *mockExecutor) Back()        {}
func (e *mockExecutor) Home()        {}
func (e *mockExecutor) Sleep(ms int) {}

func TestParseFraction(t *testing.T) {
	cases := []struct {
		in       string
		cur, max int
		ok       bool
	}{
		{"2/3", 2, 3, true},
		{"1/12,611", 1, 12611, true},
		{"0 / 5", 0, 5, true},
		{"", 0, 0, false},
		{"abc", 0, 0, false},
		{"3", 0, 0, false},
	}
	for _, c := range cases {
		cur, max, ok := parseFraction(c.in)
		if cur != c.cur || max != c.max || ok != c.ok {
			t.Errorf("parseFraction(%q) = (%d,%d,%v), want (%d,%d,%v)", c.in, cur, max, ok, c.cur, c.max, c.ok)
		}
	}
}

func TestHasNoMineCardHint(t *testing.T) {
	if !hasNoMineCardHint("没有可选择的矿脉卡") {
		t.Error("should match full hint")
	}
	if !hasNoMineCardHint("这里没有什么") {
		t.Error("should match substring 没有")
	}
	if hasNoMineCardHint("") || hasNoMineCardHint("  ") {
		t.Error("empty text should not match")
	}
	if hasNoMineCardHint("可选择 3 张") {
		t.Error("unrelated text should not match")
	}
}

func TestReadChooseQuota(t *testing.T) {
	f := DefaultFeature()
	d := &mockDetector{ocrFunc: func(r screen.Region) string {
		if r == f.CanChooseNum {
			return "2/3"
		}
		return ""
	}}
	p := NewPage(d, &mockExecutor{}, f, nil)
	cur, max, ok := p.ReadChooseQuota()
	if !ok || cur != 2 || max != 3 {
		t.Fatalf("ReadChooseQuota = (%d,%d,%v), want (2,3,true)", cur, max, ok)
	}
}

func TestIsRewardPage(t *testing.T) {
	f := DefaultFeature()
	d := &mockDetector{findOCR: func(r screen.Region, kw string) (screen.Point, bool) {
		if kw == f.RewardPage.TitleText {
			return screen.Point{X: 500, Y: 250}, true
		}
		return screen.Point{}, false
	}}
	p := NewPage(d, &mockExecutor{}, f, nil)
	if !p.IsRewardPage() {
		t.Fatal("IsRewardPage should be true when title OCR hits")
	}
	d.findOCR = nil
	if p.IsRewardPage() {
		t.Fatal("IsRewardPage should be false when title OCR misses")
	}
}

// SelectTargetCards：配额随点选上升，选满 need 张后返回。
func TestSelectTargetCardsFillQuota(t *testing.T) {
	f := DefaultFeature()
	selected := 0
	d := &mockDetector{
		anchors: []screen.Point{{X: 300, Y: 650}},
		ocrFunc: func(r screen.Region) string {
			if r == f.CanChooseNum {
				return fmt.Sprintf("%d/2", selected)
			}
			return "卡列表" // 边缘 OCR 有文字：未到尽头
		},
	}
	e := &mockExecutor{onTap: func() { selected++ }}
	p := NewPage(d, e, f, nil)

	target := f.OreVeinCards[CardButterAmber]
	got, exhausted := p.SelectTargetCards(target, 2, "left")
	if exhausted {
		t.Fatal("should not be exhausted when quota filled")
	}
	if got != 2 {
		t.Fatalf("got = %d, want 2", got)
	}
	if selected != 2 {
		t.Fatalf("selected = %d, want 2", selected)
	}
}

// SelectTargetCards：列表无目标卡且边缘无文字 → 穷尽。
func TestSelectTargetCardsExhausted(t *testing.T) {
	f := DefaultFeature()
	d := &mockDetector{
		anchors: nil,
		ocrFunc: func(r screen.Region) string {
			if r == f.CanChooseNum {
				return "0/1"
			}
			return "" // 边缘 OCR 无文字 → 到尽头
		},
	}
	p := NewPage(d, &mockExecutor{}, f, nil)
	got, exhausted := p.SelectTargetCards(f.OreVeinCards[CardSugarOre], 1, "left")
	if !exhausted {
		t.Fatal("should be exhausted when list end reached")
	}
	if got != 0 {
		t.Fatalf("got = %d, want 0", got)
	}
}

// SelectTargetCards：配额已满直接返回。
func TestSelectTargetCardsQuotaFull(t *testing.T) {
	f := DefaultFeature()
	d := &mockDetector{ocrFunc: func(r screen.Region) string { return "2/2" }}
	p := NewPage(d, &mockExecutor{}, f, nil)
	got, exhausted := p.SelectTargetCards(f.OreVeinCards[CardSugarOre], 1, "left")
	if got != 0 || exhausted {
		t.Fatalf("full quota = (%d,%v), want (0,false)", got, exhausted)
	}
}

func TestResolveCardPriority(t *testing.T) {
	f := DefaultFeature()
	// 配置顺序优先
	cfg := &Config{OreCards: []string{CardSugarOre, CardButterAmber}}
	got := resolveCardPriority(cfg, f.OreVeinCards)
	if len(got) != 2 || got[0] != CardSugarOre || got[1] != CardButterAmber {
		t.Fatalf("priority = %v", got)
	}
	// 未配置键被过滤，空配置回退默认
	cfg2 := &Config{OreCards: []string{"notACard"}}
	got2 := resolveCardPriority(cfg2, f.OreVeinCards)
	if len(got2) != len(DefaultCardPriority) {
		t.Fatalf("fallback priority = %v", got2)
	}
	// 未取色的卡被过滤
	cards := map[string]mine.ColorFind{CardSugarOre: {}}
	got3 := resolveCardPriority(&Config{}, cards)
	if len(got3) != 0 {
		t.Fatalf("unconfigured cards should be filtered, got %v", got3)
	}
}
