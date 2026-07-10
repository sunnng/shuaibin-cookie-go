package screen

import "image"

type Point struct {
	X int
	Y int
}

type Region struct {
	X1, Y1, X2, Y2 int
}

// Feature is a placeholder for a screen feature descriptor.
type Feature struct{
	Colors string // 多点比色串
    Sim    float32
}

// Rect is a placeholder rectangle for UI element bounds.
type Rect struct {
	X1, Y1, X2, Y2 int
}

// FindDef is a placeholder for a color/image find definition.
type FindDef struct{}

// OCRCfg is a placeholder for OCR region and language configuration.
type OCRCfg struct{}

type Detector interface {
	Capture() *image.NRGBA
	MatchColor(x, y int, color string, sim float32) bool
	FindColor(region Region, color string, sim float32, dir int) (Point, bool)
	FindMultiColorsAll(region Region, colors string, sim float32, dir int) []Point
	MatchMultiColor(colors string, sim float32) bool
	MatchImage(region Region, template []byte, sim float32) (Point, bool)
	OCRText(region Region) string
}
