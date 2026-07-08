package screen

import (
	"image"

	"github.com/Dasongzi1366/AutoGo/images"
	"github.com/Dasongzi1366/AutoGo/ppocr"
)

type AndroidDetector struct {
	displayId int
	ocr       *ppocr.Ppocr
}

func NewAndroidDetector(displayId int) Detector {
	return &AndroidDetector{
		displayId: displayId,
		ocr:       ppocr.New("v5"),
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

func (d *AndroidDetector) MatchMultiColor(colors string, sim float32) bool {
	return images.DetectsMultiColors(colors, sim, d.displayId)
}
