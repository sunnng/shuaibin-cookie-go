package guard

import (
	"errors"
	"image"
	"testing"
	"time"

	"app/internal/screen"
)

type mockDetector struct {
	match bool
}

func (m *mockDetector) Capture() *image.NRGBA { return nil }
func (m *mockDetector) MatchColor(x, y int, color string, sim float32) bool {
	return m.match
}
func (m *mockDetector) FindColor(region screen.Region, color string, sim float32, dir int) (screen.Point, bool) {
	return screen.Point{}, false
}
func (m *mockDetector) MatchMultiColor(colors string, sim float32) bool { return m.match }
func (m *mockDetector) MatchImage(region screen.Region, template []byte, sim float32) (screen.Point, bool) {
	return screen.Point{}, false
}
func (m *mockDetector) OCRText(region screen.Region) string { return "" }

func TestGuardCheck(t *testing.T) {
	g := New(&mockDetector{match: true})
	handled := false
	g.Register("popup", "feature", func() error { handled = true; return nil }, 10)
	if !g.Check() {
		t.Fatal("expected guard to handle")
	}
	if !handled {
		t.Fatal("handler not called")
	}
}

func TestGuardCheckNoMatch(t *testing.T) {
	g := New(&mockDetector{match: false})
	handled := false
	g.Register("popup", "feature", func() error { handled = true; return nil }, 10)
	if g.Check() {
		t.Fatal("expected guard not to handle")
	}
	if handled {
		t.Fatal("handler should not be called")
	}
}

func TestGuardPriorityOrder(t *testing.T) {
	g := New(&mockDetector{match: true})
	order := []string{}
	g.Register("low", "low", func() error { order = append(order, "low"); return nil }, 1)
	g.Register("high", "high", func() error { order = append(order, "high"); return nil }, 10)
	g.Register("mid", "mid", func() error { order = append(order, "mid"); return nil }, 5)
	g.Check()
	if len(order) != 1 || order[0] != "high" {
		t.Fatalf("expected high priority handler to run first, got %v", order)
	}
}

func TestGuardHandlerError(t *testing.T) {
	g := New(&mockDetector{match: true})
	g.Register("popup", "feature", func() error { return errors.New("boom") }, 10)
	if g.Check() {
		t.Fatal("expected guard to report not handled when handler errors")
	}
}

func TestGuardSleepChecksPeriodically(t *testing.T) {
	g := New(&mockDetector{match: true})
	handled := false
	g.Register("popup", "feature", func() error { handled = true; return nil }, 10)
	start := time.Now()
	g.Sleep(50 * time.Millisecond)
	if !handled {
		t.Fatal("expected guard to be checked during sleep")
	}
	if time.Since(start) < 50*time.Millisecond {
		t.Fatal("expected sleep to last at least the requested duration")
	}
}
