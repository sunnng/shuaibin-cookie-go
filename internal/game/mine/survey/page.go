package survey

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"app/internal/logger"
	"app/internal/platform/action"
	"app/internal/platform/screen"
)

// Page 勘查域页面封装（对应 Lua 勘查_页面.lua）。
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

func (p *Page) matchAny(fs ...screen.Feature) bool {
	for _, f := range fs {
		if p.match(f) {
			return true
		}
	}
	return false
}

// IsDomain 是否在勘查域（setup/ready/running/settle 任一命中）。
func (p *Page) IsDomain() bool {
	return p.matchAny(
		p.feature.Setup.Identify,
		p.feature.Ready.Identify,
		p.feature.Running.Identify,
		p.feature.Settle.Identify,
	)
}

// WaitDomain 等待进入勘查域（Lua waitMineVentureDomain 默认 60s）。
func (p *Page) WaitDomain(timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if p.IsDomain() {
			return true
		}
		p.executor.Sleep(500)
	}
	return false
}

func (p *Page) IsSetup() bool   { return p.match(p.feature.Setup.Identify) }
func (p *Page) IsReady() bool   { return p.match(p.feature.Ready.Identify) }
func (p *Page) IsRunning() bool { return p.match(p.feature.Running.Identify) }
func (p *Page) IsSettle() bool  { return p.match(p.feature.Settle.Identify) }

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

// tapUntilMatch 反复点 r 直到 f 命中或超时（Lua Color.tapUntilMatch）。
func (p *Page) tapUntilMatch(r screen.Region, f screen.Feature, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if p.match(f) {
			return true
		}
		p.executor.Tap(action.RandomIn(r))
		p.executor.Sleep(800)
	}
	return p.match(f)
}

// Setup 自动选择饼干并开始勘查（对应 Lua mineVenturePage.setup）。
// 步骤：autoSelect → 等 ready → start → 说明弹窗确认 → 确认饼干弹窗确认 → 等 running。
func (p *Page) Setup() bool {
	p.executor.Tap(action.RandomIn(p.feature.Setup.AutoSelectBtn))
	p.executor.Sleep(500)
	if !p.waitMatch(p.feature.Ready.Identify, 30*time.Second) {
		return false
	}

	p.executor.Tap(action.RandomIn(p.feature.Ready.StartBtn))
	p.executor.Sleep(500)
	if !p.waitMatch(p.feature.Dialogs.Info.Identify, 10*time.Second) {
		return false
	}

	p.executor.Tap(action.RandomIn(p.feature.Dialogs.Info.Confirm))
	p.executor.Sleep(500)
	if !p.waitMatch(p.feature.Dialogs.ConfirmCookie.Identify, 10*time.Second) {
		return false
	}

	p.executor.Tap(action.RandomIn(p.feature.Dialogs.ConfirmCookie.Confirm))
	p.executor.Sleep(500)
	if !p.waitMatch(p.feature.Running.Identify, 15*time.Second) {
		return false
	}

	return true
}

// StopVenture 停止勘查并结算（对应 Lua mineVenturePage.stopVenture）。
// 与 Lua 一致：结算页点 finishBtn 到 setup 页重现的等待结果不决定返回值，
// 停止确认弹窗点完即视为成功。
func (p *Page) StopVenture() bool {
	p.executor.Tap(action.RandomIn(p.feature.Running.StopBtn))
	p.executor.Sleep(500)
	if !p.waitMatch(p.feature.Dialogs.Stop.Identify, 10*time.Second) {
		return false
	}

	p.executor.Tap(action.RandomIn(p.feature.Dialogs.Stop.Confirm))
	p.executor.Sleep(500)
	p.tapUntilMatch(p.feature.Settle.FinishBtn, p.feature.Setup.Identify, 20*time.Second)

	return true
}

// GetCurrentFloor OCR 读取当前层数；识别失败返回 (0, false)。
func (p *Page) GetCurrentFloor() (int, bool) {
	text, err := p.detector.OCRText(p.feature.Running.FloorOCR)
	if err != nil {
		logger.Warnf("[矿山勘查.页面] floor OCR failed: %v", err)
		return 0, false
	}
	return parseFirstInt(text)
}

// TapBackBtn 勘查域 → 矿山首页。
func (p *Page) TapBackBtn() {
	p.executor.Tap(action.RandomIn(p.feature.BackBtn))
	p.executor.Sleep(1000)
}

var reFirstInt = regexp.MustCompile(`\d+`)

// parseFirstInt 从 OCR 文本中提取第一个整数（去掉千分位逗号后匹配）。
func parseFirstInt(s string) (int, bool) {
	s = strings.ReplaceAll(s, ",", "")
	s = strings.ReplaceAll(s, "，", "")
	m := reFirstInt.FindString(s)
	if m == "" {
		return 0, false
	}
	n, err := strconv.Atoi(m)
	return n, err == nil
}
