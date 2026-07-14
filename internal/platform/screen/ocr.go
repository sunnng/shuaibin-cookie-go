//go:build android && cgo

package screen

import (
	"strings"

	"github.com/Dasongzi1366/AutoGo/images"
	"github.com/Dasongzi1366/AutoGo/ppocr"
)

const ocrVersion = "v5"

func (d *AndroidDetector) OCRText(region Region) string {
	results := d.ocrResults(region)
	if len(results) == 0 {
		return ""
	}
	var sb strings.Builder
	for i, r := range results {
		if i > 0 {
			sb.WriteByte('\n')
		}
		sb.WriteString(r.Label)
	}
	return sb.String()
}

func (d *AndroidDetector) FindOCRText(region Region, keyword string) (Point, bool) {
	if keyword == "" {
		return Point{}, false
	}
	results := d.ocrResults(region)
	for _, r := range results {
		if !strings.Contains(r.Label, keyword) {
			continue
		}
		cx, cy := r.CenterX, r.CenterY
		if cx == 0 && cy == 0 {
			cx = r.X + r.Width/2
			cy = r.Y + r.Height/2
		}
		return Point{X: region.X1 + cx, Y: region.Y1 + cy}, true
	}
	return Point{}, false
}

func (d *AndroidDetector) ocrResults(region Region) []ppocr.Result {
	img := images.CaptureScreen(region.X1, region.Y1, region.X2, region.Y2, d.displayId)
	if img == nil {
		return nil
	}
	engine := ppocr.New(ocrVersion)
	if engine == nil {
		return nil
	}
	defer engine.Close()
	return engine.OcrFromImage(img, "")
}
