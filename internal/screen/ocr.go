//go:build android

package screen

import (
	"strings"

	"github.com/Dasongzi1366/AutoGo/images"
	"github.com/Dasongzi1366/AutoGo/ppocr"
)

const ocrVersion = "v5"

func (d *AndroidDetector) OCRText(region Region) string {
	img := images.CaptureScreen(region.X1, region.Y1, region.X2, region.Y2, d.displayId)
	if img == nil {
		return ""
	}
	engine := ppocr.New(ocrVersion)
	if engine == nil {
		return ""
	}
	defer engine.Close()
	results := engine.OcrFromImage(img, "")
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
