package mine

import (
	"time"

	"app/internal/logger"
	"app/internal/platform/action"
	"app/internal/platform/screen"
)

// Page 是矿山首页：四个子任务的入口页 + 回王国首页的中转页。
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

func (p *Page) IsCurrent() bool {
	id := p.feature.Home.Identify
	if id.Colors == "" {
		return false
	}
	return p.detector.MatchMultiColor(id.Colors, id.Sim)
}

// HasCompletedMiningTask 首页角标：存在已完成的开采任务（开采子任务预检用）。
func (p *Page) HasCompletedMiningTask() bool {
	id := p.feature.Home.CompletedTaskIdentify
	if id.Colors == "" {
		return false
	}
	return p.detector.MatchMultiColor(id.Colors, id.Sim)
}

// WaitCurrent 等待矿山首页出现（Lua MineHomePage.wait 默认 60s）。
func (p *Page) WaitCurrent(timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if p.IsCurrent() {
			return true
		}
		p.executor.Sleep(500)
	}
	return false
}

// WaitGone 等待矿山首页消失（Lua MineHomePage.waitGone 默认 30s）。
func (p *Page) WaitGone(timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if !p.IsCurrent() {
			return true
		}
		p.executor.Sleep(500)
	}
	return false
}

func (p *Page) TapVenture() {
	p.executor.Tap(action.RandomIn(p.feature.Home.Actions.VentureBtn))
	p.executor.Sleep(1000)
}

func (p *Page) TapMining() {
	p.executor.Tap(action.RandomIn(p.feature.Home.Actions.MiningBtn))
	p.executor.Sleep(1000)
}

func (p *Page) TapBattle() {
	p.executor.Tap(action.RandomIn(p.feature.Home.Actions.BattleBtn))
	p.executor.Sleep(1000)
}

// TapJellyEntry 点「解除洋菜冻」入口（Lua JellyPage.tapEnterBtn，
// 坐标定义在 Lua mineVenture 特征库，本包归入首页动作）。
func (p *Page) TapJellyEntry() {
	p.executor.Tap(action.RandomIn(p.feature.Home.Actions.JellyBtn))
	p.executor.Sleep(1000)
}

// TapBack 矿山首页 → 王国首页。
func (p *Page) TapBack() {
	p.executor.Tap(action.RandomIn(p.feature.Home.Actions.BackBtn))
	p.executor.Sleep(1000)
}

// TapEntryMine 王国活动页 → 点矿山入口（Lua KingdomPage.tapMineBtn）。
func (p *Page) TapEntryMine() {
	r := p.feature.Entry.MineBtn
	if !action.RegionConfigured(r) {
		logger.Warnf("[Mine] Entry.MineBtn not configured")
		return
	}
	p.executor.Tap(action.RandomIn(r))
	p.executor.Sleep(1200)
}
