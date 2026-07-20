package action

import "app/internal/platform/screen"

const (
	BaseWidth  = 1600
	BaseHeight = 900
)

// Point 是 screen.Point 的别名：识别层输出的坐标可直接传给执行层，
// 无需在 page 层做类型转换。
type Point = screen.Point

func Scale(p Point, actualW, actualH int) Point {
	if actualW <= 0 || actualH <= 0 {
		return Point{X: 0, Y: 0}
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
