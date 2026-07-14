package screen

import "image"

type Point struct {
	X int
	Y int
}

type Region struct {
	X1, Y1, X2, Y2 int
}

// Rect is an alias for Region (click/bounds rectangles share the same shape).
type Rect = Region

// Feature describes a multi-color screen identity check.
type Feature struct {
	Colors string // 多点比色串
	Sim    float32
}

type Detector interface {
	Capture() *image.NRGBA
	MatchColor(x, y int, color string, sim float32) bool
	FindColor(region Region, color string, sim float32, dir int) (Point, bool)
	FindMultiColorsAll(region Region, colors string, sim float32, dir int) []Point
	MatchMultiColor(colors string, sim float32) bool
	MatchImage(region Region, template []byte, sim float32) (Point, bool)
	OCRText(region Region) string
	// FindOCRText returns the screen-absolute center of the first OCR hit whose
	// label contains keyword. Empty keyword → (0, false).
	FindOCRText(region Region, keyword string) (Point, bool)
}
