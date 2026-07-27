package jelly

import (
	"regexp"
	"strconv"
	"time"

	"app/internal/logger"
	"app/internal/platform/action"
	"app/internal/platform/screen"
)

// Page 解除洋菜冻页面封装（对应 Lua 解除洋菜冻_页面.lua）。
type Page struct {
	detector screen.Detector
	executor action.Executor
	feature  *Feature
}

func NewPage(det screen.Detector, exec action.Executor, f *Feature) *Page {
	if f == nil {
		f = DefaultFeature()
	}
	return &Page{detector: det, executor: exec, feature: f}
}

func (p *Page) match(f screen.Feature) bool {
	if f.Colors == "" {
		return false
	}
	return p.detector.MatchMultiColor(f.Colors, f.Sim)
}

func (p *Page) waitMatch(f screen.Feature, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if p.match(f) {
			return true
		}
		p.executor.Sleep(500)
	}
	return false
}

func (p *Page) IsJellyPage() bool { return p.match(p.feature.Identify) }

func (p *Page) WaitJellyPage(timeout time.Duration) bool {
	return p.waitMatch(p.feature.Identify, timeout)
}

func (p *Page) CanClaimAll() bool { return p.match(p.feature.ClaimAllIdentify) }

func (p *Page) TapClaimAll() {
	p.executor.Tap(action.RandomIn(p.feature.ClaimAllBtn))
	p.executor.Sleep(800)
}

func (p *Page) TapBack() {
	p.executor.Tap(action.RandomIn(p.feature.BackBtn))
	p.executor.Sleep(1000)
}

// TapSettleUntilJellyPage 领取后反复点结算区直到洋菜冻页恢复（Lua 内联 tapUntilMatch）。
func (p *Page) TapSettleUntilJellyPage() bool {
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		if p.IsJellyPage() {
			return true
		}
		p.executor.Tap(action.RandomIn(p.feature.SettleBtn))
		p.executor.Sleep(800)
	}
	return p.IsJellyPage()
}

// FindConfigBtn OCR 查找「配置」按钮坐标。
func (p *Page) FindConfigBtn() (screen.Point, bool) {
	return p.detector.FindOCRText(p.feature.OCRRegion, "配置")
}

func (p *Page) TapConfigBtn(pt screen.Point) {
	p.executor.Tap(pt)
	p.executor.Sleep(800)
}

func (p *Page) IsConfigPage() bool { return p.match(p.feature.Config.Identify) }

func (p *Page) WaitConfigPage(timeout time.Duration) bool {
	return p.waitMatch(p.feature.Config.Identify, timeout)
}

func (p *Page) CanChoose() bool { return p.match(p.feature.Config.CanChooseIdentify) }

func (p *Page) TapChoose() {
	p.executor.Tap(action.RandomIn(p.feature.Config.ChooseBtn))
	p.executor.Sleep(800)
}

func (p *Page) TapConfigBack() {
	p.executor.Tap(action.RandomIn(p.feature.Config.BackBtn))
	p.executor.Sleep(1000)
}

// ReadRemainTime OCR 识别解除洋菜冻剩余时间。
// 注意：Lua 会逐 OCR item 兜底解析；Go Detector.OCRText 只有合并文本，
// 这里只解析合并文本（item 级兜底无法迁移）。
func (p *Page) ReadRemainTime() (time.Duration, bool) {
	text, err := p.detector.OCRText(p.feature.OCRRegion)
	if err != nil {
		logger.Warnf("[解除洋菜冻.页面] readRemainTime OCR failed: %v", err)
		return 0, false
	}
	return parseRemainTime(text)
}

var (
	reDays    = regexp.MustCompile(`(\d+)\s*天`)
	reHours   = regexp.MustCompile(`(\d+)\s*小时`)
	reMinutes = regexp.MustCompile(`(\d+)\s*分钟`)
	reSeconds = regexp.MustCompile(`(\d+)\s*秒`)
)

// parseRemainTime 解析中文时长文本：X天Y小时Z分钟W秒 的任意组合。
func parseRemainTime(text string) (time.Duration, bool) {
	if text == "" {
		return 0, false
	}
	total := 0
	if m := reDays.FindStringSubmatch(text); m != nil {
		v, _ := strconv.Atoi(m[1])
		total += v * 86400
	}
	if m := reHours.FindStringSubmatch(text); m != nil {
		v, _ := strconv.Atoi(m[1])
		total += v * 3600
	}
	if m := reMinutes.FindStringSubmatch(text); m != nil {
		v, _ := strconv.Atoi(m[1])
		total += v * 60
	}
	if m := reSeconds.FindStringSubmatch(text); m != nil {
		v, _ := strconv.Atoi(m[1])
		total += v
	}
	if total == 0 {
		return 0, false
	}
	return time.Duration(total) * time.Second, true
}
