package ui

import "testing"

func TestClampIslandPos(t *testing.T) {
	cases := []struct {
		name               string
		x, y, w, h, sw, sh float32
		wantX, wantY       float32
	}{
		{"inside", 100, 50, 240, 64, 1600, 900, 100, 50},
		{"left overflow", -30, 50, 240, 64, 1600, 900, 0, 50},
		{"top overflow", 100, -10, 240, 64, 1600, 900, 100, 0},
		{"right overflow", 1500, 50, 240, 64, 1600, 900, 1360, 50},
		{"bottom overflow", 100, 880, 240, 64, 1600, 900, 100, 836},
		{"corner overflow", 1500, 880, 240, 64, 1600, 900, 1360, 836},
		{"wider than screen", -5, 10, 2000, 64, 1600, 900, 0, 10},
		{"taller than screen", 10, -5, 240, 1000, 1600, 900, 10, 0},
	}
	for _, c := range cases {
		gotX, gotY := clampIslandPos(c.x, c.y, c.w, c.h, c.sw, c.sh)
		if gotX != c.wantX || gotY != c.wantY {
			t.Errorf("%s: clamp(%v,%v) = (%v,%v), want (%v,%v)",
				c.name, c.x, c.y, gotX, gotY, c.wantX, c.wantY)
		}
	}
}
