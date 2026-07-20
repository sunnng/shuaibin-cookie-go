//go:build android && cgo

package screen

import (
	"errors"
	"strings"
	"sync"

	"github.com/Dasongzi1366/AutoGo/images"
	"github.com/Dasongzi1366/AutoGo/ppocr"
)

const ocrVersion = "v5"

// ocrEngine 缓存 ppocr 引擎：ppocr.New 会加载模型（官方用法是 New 一次、
// 反复复用），每次调用新建+Close 会在每个 OCR 调用上重复付出初始化成本。
// 引擎随 detector 存活整个进程，故意不 Close。初始化失败不缓存，下次调用重试。
var (
	ocrMu     sync.Mutex
	ocrEngine *ppocr.Ppocr
)

func sharedOcrEngine() (*ppocr.Ppocr, error) {
	ocrMu.Lock()
	defer ocrMu.Unlock()
	if ocrEngine != nil {
		return ocrEngine, nil
	}
	engine := ppocr.New(ocrVersion)
	if engine == nil {
		return nil, errors.New("init ocr engine failed")
	}
	ocrEngine = engine
	return ocrEngine, nil
}

func (d *AndroidDetector) OCRText(region Region) (string, error) {
	results, err := d.ocrResults(region)
	if err != nil {
		return "", err
	}
	if len(results) == 0 {
		return "", nil
	}
	var sb strings.Builder
	for i, r := range results {
		if i > 0 {
			sb.WriteByte('\n')
		}
		sb.WriteString(r.Label)
	}
	return sb.String(), nil
}

func (d *AndroidDetector) FindOCRText(region Region, keyword string) (Point, bool) {
	if keyword == "" {
		return Point{}, false
	}
	// 探测语义：OCR 通道故障与"未命中"一样按 false 处理，由上层重试。
	results, err := d.ocrResults(region)
	if err != nil {
		return Point{}, false
	}
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

func (d *AndroidDetector) ocrResults(region Region) ([]ppocr.Result, error) {
	img := images.CaptureScreen(region.X1, region.Y1, region.X2, region.Y2, d.displayId)
	if img == nil {
		return nil, errors.New("capture screen failed")
	}
	engine, err := sharedOcrEngine()
	if err != nil {
		return nil, err
	}
	return engine.OcrFromImage(img, ""), nil
}
