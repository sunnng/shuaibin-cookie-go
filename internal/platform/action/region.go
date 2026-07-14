package action

import "math/rand"

// Region is a tap target box in base coordinates (x1,y1,x2,y2).
// Prefer this over a single Point for buttons so taps can be randomized.
type Region struct {
	X1, Y1, X2, Y2 int
}

// RegionConfigured reports whether r was filled by color-picking.
// All-zero means unset.
func RegionConfigured(r Region) bool {
	return r.X1 != 0 || r.Y1 != 0 || r.X2 != 0 || r.Y2 != 0
}

// RandomIn returns a uniformly random point inside r (inclusive).
// Degenerate boxes (x1==x2 && y1==y2) tap that single point.
func RandomIn(r Region) Point {
	x1, y1, x2, y2 := r.X1, r.Y1, r.X2, r.Y2
	if x2 < x1 {
		x1, x2 = x2, x1
	}
	if y2 < y1 {
		y1, y2 = y2, y1
	}
	if x1 == x2 && y1 == y2 {
		return Point{X: x1, Y: y1}
	}
	return Point{
		X: x1 + rand.Intn(x2-x1+1),
		Y: y1 + rand.Intn(y2-y1+1),
	}
}
