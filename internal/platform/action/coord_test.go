package action

import "testing"

func TestScaleSameResolution(t *testing.T) {
	p := Point{X: 800, Y: 450}
	got := Scale(p, 1600, 900)
	if got != p {
		t.Fatalf("expected %v, got %v", p, got)
	}
}

func TestScaleHalfResolution(t *testing.T) {
	p := Point{X: 800, Y: 450}
	got := Scale(p, 800, 450)
	if got != (Point{X: 400, Y: 225}) {
		t.Fatalf("expected {400 225}, got %v", got)
	}
}

func TestBound(t *testing.T) {
	p := Point{X: 2000, Y: -10}
	got := Bound(p, 1600, 900)
	if got != (Point{X: 1600, Y: 0}) {
		t.Fatalf("expected {1600 0}, got %v", got)
	}
}
