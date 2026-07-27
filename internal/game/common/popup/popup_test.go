package popup

import (
	"image"
	"testing"

	"app/internal/guard"
	"app/internal/platform/screen"
)

type mockDetector struct {
	matches int // MatchMultiColor 剩余返回 true 的次数，之后返回 false
	calls   int
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
	m.calls++
	if m.matches > 0 {
		m.matches--
		return true
	}
	return false
}
func (m *mockDetector) MatchImage(region screen.Region, template []byte, sim float32) (screen.Point, bool) {
	return screen.Point{}, false
}
func (m *mockDetector) OCRText(region screen.Region) (string, error) { return "", nil }
func (m *mockDetector) FindOCRText(region screen.Region, keyword string) (screen.Point, bool) {
	return screen.Point{}, false
}

type mockExecutor struct {
	taps   []screen.Point
	sleeps int
}

func (m *mockExecutor) Tap(p screen.Point)                  { m.taps = append(m.taps, p) }
func (m *mockExecutor) LongTap(p screen.Point, ms int)      {}
func (m *mockExecutor) Swipe(from, to screen.Point, ms int) {}
func (m *mockExecutor) Back()                               {}
func (m *mockExecutor) Home()                               {}
func (m *mockExecutor) Sleep(ms int)                        { m.sleeps++ }

// 命中弹窗：点确认按钮中心，特征消失后返回。
func TestRegisterHandlesPopup(t *testing.T) {
	det := &mockDetector{matches: 2} // guard.Check 命中 1 次 + handler 轮询 1 次仍可见，第 2 次消失
	exec := &mockExecutor{}
	g := guard.New(det)
	Register(g, exec)

	if !g.Check() {
		t.Fatal("guard should hit the popup trap")
	}
	if len(exec.taps) != 1 {
		t.Fatalf("taps=%v", exec.taps)
	}
	want := screen.Point{X: (775 + 828) / 2, Y: (621 + 647) / 2}
	if exec.taps[0] != want {
		t.Fatalf("tap=%v want %v", exec.taps[0], want)
	}
	if exec.sleeps == 0 {
		t.Fatal("handler should poll with sleeps")
	}
}

// 弹窗一直不消失时 handler 在 waitGoneMs 后超时返回（不卡死）。
func TestHandlerTimesOut(t *testing.T) {
	det := &mockDetector{matches: 1 << 30} // 永远可见
	exec := &mockExecutor{}
	g := guard.New(det)
	Register(g, exec)

	if !g.Check() {
		t.Fatal("guard should hit the popup trap")
	}
	// 轮询次数 = waitGoneMs/pollMs（Check 命中消耗 1 次 matches 不影响）。
	if exec.sleeps != waitGoneMs/pollMs {
		t.Fatalf("sleeps=%d want %d", exec.sleeps, waitGoneMs/pollMs)
	}
}

// 未命中特征时 guard 不触发 handler。
func TestNoMatchNoHandle(t *testing.T) {
	det := &mockDetector{matches: 0}
	exec := &mockExecutor{}
	g := guard.New(det)
	Register(g, exec)

	if g.Check() {
		t.Fatal("guard should not hit without matching feature")
	}
	if len(exec.taps) != 0 {
		t.Fatalf("taps=%v", exec.taps)
	}
}
