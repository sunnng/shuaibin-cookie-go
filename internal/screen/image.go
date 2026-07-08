package screen

import "github.com/Dasongzi1366/AutoGo/opencv"

func (d *AndroidDetector) MatchImage(region Region, template []byte, sim float32) (Point, bool) {
	x, y := opencv.FindImage(region.X1, region.Y1, region.X2, region.Y2, &template, false, false, sim, d.displayId)
	if x < 0 || y < 0 {
		return Point{}, false
	}
	return Point{X: x, Y: y}, true
}
