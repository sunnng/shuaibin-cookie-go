package screen

import (
	"strings"

	"github.com/Dasongzi1366/AutoGo/images"
)

func (d *AndroidDetector) OCRText(region Region) string {
	if d.ocr == nil {
		return ""
	}
	img := images.CaptureScreen(region.X1, region.Y1, region.X2, region.Y2, d.displayId)
	if img == nil {
		return ""
	}
	results := d.ocr.OcrFromImage(img, "")
	if len(results) == 0 {
		return ""
	}
	labels := make([]string, 0, len(results))
	for _, r := range results {
		labels = append(labels, r.Label)
	}
	return strings.Join(labels, "\n")
}
