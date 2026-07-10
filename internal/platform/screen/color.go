//go:build android && cgo

package screen

import (
	"image"

	"github.com/Dasongzi1366/AutoGo/images"
)

type AndroidDetector struct {
	displayId int
}

func NewAndroidDetector(displayId int) Detector {
	return &AndroidDetector{
		displayId: displayId,
	}
}

func (d *AndroidDetector) Capture() *image.NRGBA {
	return images.CaptureScreen(0, 0, 0, 0, d.displayId)
}

func (d *AndroidDetector) MatchColor(x, y int, color string, sim float32) bool {
	return images.CmpColor(x, y, color, sim, d.displayId)
}

func (d *AndroidDetector) FindColor(region Region, color string, sim float32, dir int) (Point, bool) {
	x, y := images.FindColor(region.X1, region.Y1, region.X2, region.Y2, color, sim, dir, d.displayId)
	if x < 0 || y < 0 {
		return Point{}, false
	}
	return Point{X: x, Y: y}, true
}

func (d *AndroidDetector) FindMultiColorsAll(region Region, colors string, sim float32, dir int) []Point {
	pts := images.FindMultiColorsAll(region.X1, region.Y1, region.X2, region.Y2, colors, sim, dir, d.displayId)
	out := make([]Point, len(pts))
	for i, p := range pts {
		out[i] = Point{X: p.X, Y: p.Y}
	}
	return out
}

func (d *AndroidDetector) MatchMultiColor(colors string, sim float32) bool {
	return images.DetectsMultiColors(colors, sim, d.displayId)
}
