package production

import (
	"time"

	"app/internal/logger"
	"app/internal/platform/action"
	"app/internal/platform/screen"
)

// Page 王国生产页内识别与点击（基于 Feature）。
type Page struct {
	detector screen.Detector
	executor action.Executor
	feature  *Feature
}

// NewPage 构造页面；feature 为 nil 时用 DefaultFeature。
func NewPage(det screen.Detector, exec action.Executor, f *Feature) *Page {
	if f == nil {
		f = DefaultFeature()
	}
	return &Page{detector: det, executor: exec, feature: f}
}

// IsBoard 当前是否在生产总览/产线界面。未取色时恒为 false。
func (p *Page) IsBoard() bool {
	id := p.feature.Board.Identify
	if id.Colors == "" {
		return false
	}
	return p.detector.MatchMultiColor(id.Colors, id.Sim)
}

// WaitBoard 等待进入生产界面。
func (p *Page) WaitBoard(timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if p.IsBoard() {
			return true
		}
		p.executor.Sleep(300)
	}
	return false
}

// TapCollectAll 点击一键收取（或主收取区）。区域未配置时跳过并返回 false。
func (p *Page) TapCollectAll() bool {
	r := p.feature.Board.Actions.CollectAll
	if !action.RegionConfigured(r) {
		logger.Warnf("[Production] CollectAll not configured")
		return false
	}
	p.executor.Tap(action.RandomIn(r))
	p.executor.Sleep(800)
	return true
}
