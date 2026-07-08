package screen

import "image"

type Point struct {
	X int
	Y int
}

type Region struct {
	X1, Y1, X2, Y2 int
}

type Detector interface {
	Capture() *image.NRGBA
	MatchColor(x, y int, color string, sim float32) bool
	FindColor(region Region, color string, sim float32, dir int) (Point, bool)
	MatchMultiColor(colors string, sim float32) bool
	MatchImage(region Region, template []byte, sim float32) (Point, bool)
	OCRText(region Region) string
}
