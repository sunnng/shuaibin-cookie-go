//go:build !android || !cgo

package screen

import "image"

// stubDetector is a non-Android placeholder that never matches anything.
type stubDetector struct{}

func NewDetector(displayId int) Detector { return &stubDetector{} }

func (s *stubDetector) Capture() *image.NRGBA { return nil }
func (s *stubDetector) MatchColor(x, y int, color string, sim float32) bool {
	return false
}
func (s *stubDetector) FindColor(region Region, color string, sim float32, dir int) (Point, bool) {
	return Point{}, false
}
func (s *stubDetector) FindMultiColorsAll(region Region, colors string, sim float32, dir int) []Point {
	return nil
}
func (s *stubDetector) MatchMultiColor(colors string, sim float32) bool { return false }
func (s *stubDetector) MatchImage(region Region, template []byte, sim float32) (Point, bool) {
	return Point{}, false
}
func (s *stubDetector) OCRText(region Region) string { return "" }
