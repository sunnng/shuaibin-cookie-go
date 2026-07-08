package action

const (
	BaseWidth  = 1600
	BaseHeight = 900
)

type Point struct {
	X int
	Y int
}

func Scale(p Point, actualW, actualH int) Point {
	if actualW <= 0 || actualH <= 0 {
		return Point{0, 0}
	}
	return Point{
		X: p.X * actualW / BaseWidth,
		Y: p.Y * actualH / BaseHeight,
	}
}

func Bound(p Point, maxW, maxH int) Point {
	if p.X < 0 {
		p.X = 0
	}
	if p.Y < 0 {
		p.Y = 0
	}
	if p.X > maxW {
		p.X = maxW
	}
	if p.Y > maxH {
		p.Y = maxH
	}
	return p
}

func SafeTap(p Point, actualW, actualH int) Point {
	return Bound(Scale(p, actualW, actualH), actualW, actualH)
}
