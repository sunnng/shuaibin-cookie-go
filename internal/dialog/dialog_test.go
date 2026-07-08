package dialog

import (
	"image"
	"testing"

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

func (m *mockDetector) MatchMultiColor(colors string, sim float32) bool {
	return m.match
}

func (m *mockDetector) MatchImage(region screen.Region, template []byte, sim float32) (screen.Point, bool) {
	return screen.Point{}, false
}

func (m *mockDetector) OCRText(region screen.Region) string { return "" }

func TestNew(t *testing.T) {
	d := New(Def{Name: "popup"}, "guard")
	if d == nil {
		t.Fatal("New returned nil")
	}
}

func TestIsVisibleNilFeature(t *testing.T) {
	d := New(Def{Name: "popup"}, "guard")
	if d.IsVisible() {
		t.Fatal("expected invisible when Feature is nil")
	}
}

func TestIsVisibleStringWithoutDetector(t *testing.T) {
	d := New(Def{Name: "popup", Feature: "color"}, "guard")
	if d.IsVisible() {
		t.Fatal("expected invisible without detector")
	}
}

func TestIsVisibleStringWithDetector(t *testing.T) {
	d := New(Def{Name: "popup", Feature: "color"}, "guard")
	d.detector = &mockDetector{match: true}
	if !d.IsVisible() {
		t.Fatal("expected visible when detector matches")
	}
}

func TestIsVisibleFuncFeature(t *testing.T) {
	d := New(Def{Name: "popup", Feature: func() bool { return true }}, "guard")
	if !d.IsVisible() {
		t.Fatal("expected visible when feature func returns true")
	}
}

func TestHandleIfVisibleNotVisible(t *testing.T) {
	d := New(Def{Name: "popup"}, "guard")
	ok, msg := d.Handle(HandleOpts{Mode: "ifVisible"})
	if !ok || msg != "" {
		t.Fatalf("expected (true, \"\"), got (%v, %q)", ok, msg)
	}
}

func TestHandleFlow(t *testing.T) {
	d := New(Def{Name: "popup", Feature: "color"}, "guard")
	ok, msg := d.Handle(HandleOpts{Mode: "flow", Action: "confirm"})
	if !ok || msg != "" {
		t.Fatalf("expected (true, \"\"), got (%v, %q)", ok, msg)
	}
}

func TestToGuardHandler(t *testing.T) {
	d := New(Def{Name: "popup"}, "guard")
	handler := d.ToGuardHandler(HandleOpts{Mode: "ifVisible"})
	if handler == nil {
		t.Fatal("expected non-nil handler")
	}
	if err := handler(); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}
