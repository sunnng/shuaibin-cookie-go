package guard

import (
	"errors"
	"image"
	"testing"

	"app/internal/platform/screen"
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
func (m *mockDetector) FindMultiColorsAll(region screen.Region, colors string, sim float32, dir int) []screen.Point {
	return nil
}
func (m *mockDetector) MatchImage(region screen.Region, template []byte, sim float32) (screen.Point, bool) {
	return screen.Point{}, false
}
func (m *mockDetector) OCRText(region screen.Region) (string, error) { return "", nil }
func (m *mockDetector) FindOCRText(region screen.Region, keyword string) (screen.Point, bool) {
	return screen.Point{}, false
}

func feature(colors string) screen.Feature {
	return screen.Feature{Colors: colors, Sim: 0.9}
}

func TestGuardCheck(t *testing.T) {
	g := New(&mockDetector{match: true})
	handled := false
	g.Register("popup", feature("feature"), func() error { handled = true; return nil }, 10)
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
	g.Register("popup", feature("feature"), func() error { handled = true; return nil }, 10)
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
	g.Register("low", feature("low"), func() error { order = append(order, "low"); return nil }, 1)
	g.Register("high", feature("high"), func() error { order = append(order, "high"); return nil }, 10)
	g.Register("mid", feature("mid"), func() error { order = append(order, "mid"); return nil }, 5)
	g.Check()
	if len(order) != 1 || order[0] != "high" {
		t.Fatalf("expected high priority handler to run first, got %v", order)
	}
}

func TestGuardHandlerError(t *testing.T) {
	g := New(&mockDetector{match: true})
	g.Register("popup", feature("feature"), func() error { return errors.New("boom") }, 10)
	if g.Check() {
		t.Fatal("expected guard to report not handled when handler errors")
	}
}

func TestGuardSkipsEmptyFeature(t *testing.T) {
	g := New(&mockDetector{match: true})
	handled := false
	g.Register("unconfigured", screen.Feature{}, func() error { handled = true; return nil }, 10)
	if g.Check() {
		t.Fatal("empty-colors trap must never fire")
	}
	if handled {
		t.Fatal("handler of skipped trap should not run")
	}
}
