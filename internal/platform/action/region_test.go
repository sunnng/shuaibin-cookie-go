package action

import "testing"

func TestRegionConfigured(t *testing.T) {
	if RegionConfigured(Region{}) {
		t.Fatal("zero region should be unset")
	}
	if !RegionConfigured(Region{X1: 10, Y1: 20, X2: 30, Y2: 40}) {
		t.Fatal("want configured")
	}
}

func TestRandomInSinglePoint(t *testing.T) {
	p := RandomIn(Region{X1: 100, Y1: 200, X2: 100, Y2: 200})
	if p.X != 100 || p.Y != 200 {
		t.Fatalf("got %+v", p)
	}
}

func TestRandomInStaysInside(t *testing.T) {
	r := Region{X1: 10, Y1: 20, X2: 15, Y2: 25}
	for i := 0; i < 50; i++ {
		p := RandomIn(r)
		if p.X < 10 || p.X > 15 || p.Y < 20 || p.Y > 25 {
			t.Fatalf("out of range: %+v", p)
		}
	}
}
