package biscuit

import (
	"fmt"
	"image"
	"testing"

	"app/internal/platform/action"
	"app/internal/platform/screen"
)

// ---- mock Detector ----
type mockDetector struct {
	matchMulti bool
	ocrByKey   map[string]string // key: "x1,y1,x2,y2"
	ocrErr     error
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
func (m *mockDetector) MatchMultiColor(colors string, sim float32) bool { return m.matchMulti }
func (m *mockDetector) MatchImage(r screen.Region, t []byte, s float32) (screen.Point, bool) {
	return screen.Point{}, false
}
func (m *mockDetector) OCRText(r screen.Region) (string, error) {
	return m.ocrByKey[fmt.Sprintf("%d,%d,%d,%d", r.X1, r.Y1, r.X2, r.Y2)], m.ocrErr
}
func (m *mockDetector) FindOCRText(r screen.Region, keyword string) (screen.Point, bool) {
	return screen.Point{}, false
}

// ---- mock Executor ----
type mockExecutor struct {
	taps   []action.Point
	sleeps []int
}

func (m *mockExecutor) Tap(p action.Point)                  { m.taps = append(m.taps, p) }
func (m *mockExecutor) LongTap(p action.Point, ms int)      {}
func (m *mockExecutor) Swipe(from, to action.Point, ms int) {}
func (m *mockExecutor) Back()                               {}
func (m *mockExecutor) Home()                               {}
func (m *mockExecutor) Sleep(ms int)                        { m.sleeps = append(m.sleeps, ms) }

func effectsOCRKey() string {
	r := DefaultFeature().Reads.Effects
	return fmt.Sprintf("%d,%d,%d,%d", r.X1, r.Y1, r.X2, r.Y2)
}

func TestNewPage(t *testing.T) {
	_ = NewPage(nil, nil, nil) // nil feature → DefaultFeature()
}

func TestReadEffects(t *testing.T) {
	det := &mockDetector{ocrByKey: map[string]string{
		effectsOCRKey(): "暗黑属性伤害提升10.8%生命值3%生命值7.9%会心3.7%",
	}}
	p := NewPage(det, &mockExecutor{}, DefaultFeature())

	effects := p.ReadEffects()
	if len(effects) != 4 {
		t.Fatalf("len = %d, want 4", len(effects))
	}
	if effects[0].Name != "暗黑属性伤害提升" || effects[0].Value != 10.8 {
		t.Fatalf("effects[0] = %+v", effects[0])
	}
}

func TestReadEffectsTruncateAndPad(t *testing.T) {
	det := &mockDetector{ocrByKey: map[string]string{
		effectsOCRKey(): "攻击力3%攻击力4%攻击力5%攻击力6%攻击力7%",
	}}
	p := NewPage(det, &mockExecutor{}, DefaultFeature())

	effects := p.ReadEffects()
	if len(effects) != 4 || effects[3].Value != 6 {
		t.Fatalf("truncate to 4 failed: %+v", effects)
	}

	det.ocrByKey[effectsOCRKey()] = "冷却时间5.2%"
	effects = p.ReadEffects()
	if len(effects) != 4 {
		t.Fatalf("pad to 4 failed: %+v", effects)
	}
	for i := 1; i < 4; i++ {
		if effects[i].Name != "未知" || effects[i].Value != 0 || effects[i].Raw != "" {
			t.Fatalf("effects[%d] = %+v, want 未知 placeholder", i, effects[i])
		}
	}
}

// OCR 通道故障按空文本处理（对齐 Lua scan 失败 raw=""），返回 4 条未知。
func TestReadEffectsOCRError(t *testing.T) {
	det := &mockDetector{ocrByKey: map[string]string{}, ocrErr: fmt.Errorf("capture failed")}
	p := NewPage(det, &mockExecutor{}, DefaultFeature())

	effects := p.ReadEffects()
	if len(effects) != 4 || effects[0].Name != "未知" {
		t.Fatalf("effects = %+v, want 4 unknown placeholders", effects)
	}
}

func TestConfirmDialog(t *testing.T) {
	exec := &mockExecutor{}
	p := NewPage(&mockDetector{matchMulti: true}, exec, DefaultFeature())

	if !p.ConfirmResetDialog() {
		t.Fatal("expected reset dialog handled when identify matches")
	}
	if len(exec.taps) != 2 { // 今日不再显示 + 确认
		t.Fatalf("taps = %v, want 2", exec.taps)
	}

	exec2 := &mockExecutor{}
	p2 := NewPage(&mockDetector{matchMulti: false}, exec2, DefaultFeature())
	if p2.ConfirmSameDialog() {
		t.Fatal("dialog not present should return false")
	}
	if len(exec2.taps) != 0 {
		t.Fatalf("taps = %v, want none", exec2.taps)
	}
}

func TestTapReroll(t *testing.T) {
	exec := &mockExecutor{}
	p := NewPage(&mockDetector{}, exec, DefaultFeature())
	p.TapReroll()
	if len(exec.taps) != 1 {
		t.Fatalf("taps = %v, want 1", exec.taps)
	}
	r := DefaultFeature().Actions.Reroll
	pt := exec.taps[0]
	if pt.X < r.X1 || pt.X > r.X2 || pt.Y < r.Y1 || pt.Y > r.Y2 {
		t.Fatalf("tap %+v outside reroll region %+v", pt, r)
	}
}
