package kingdom

import (
	"time"

	"app/internal/logger"
	"app/internal/platform/action"
	"app/internal/platform/screen"
)

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

func (p *Page) IsKingdomHome() bool {
	id := p.feature.Home.Identify
	if id.Colors == "" {
		return false
	}
	return p.detector.MatchMultiColor(id.Colors, id.Sim)
}

func (p *Page) IsAdventurePage() bool {
	adv := p.feature.Adventure
	if adv.Keyword == "" {
		return false
	}
	_, ok := p.detector.FindOCRText(adv.Region, adv.Keyword)
	return ok
}

func (p *Page) IsEventPage() bool {
	ev := p.feature.Event
	if ev.Keyword == "" {
		return false
	}
	_, ok := p.detector.FindOCRText(ev.Region, ev.Keyword)
	return ok
}

func (p *Page) TapAdventureBtn() {
	r := p.feature.Actions.AdventureBtn
	if !RegionConfigured(r) {
		logger.Warnf("[Kingdom] AdventureBtn not configured")
		return
	}
	_ = p.executor.Tap(tapRegion(r))
	p.executor.Sleep(1200)
}

func (p *Page) TapEventBtn() {
	r := p.feature.Actions.EventBtn
	if !RegionConfigured(r) {
		logger.Warnf("[Kingdom] EventBtn not configured")
		return
	}
	_ = p.executor.Tap(tapRegion(r))
	p.executor.Sleep(1200)
}

func (p *Page) TapBackHome() {
	r := p.feature.Actions.BackHome
	if !RegionConfigured(r) {
		logger.Warnf("[Kingdom] BackHome not configured")
		return
	}
	_ = p.executor.Tap(tapRegion(r))
	p.executor.Sleep(800)
}

func (p *Page) HasBackHome() bool {
	return RegionConfigured(p.feature.Actions.BackHome)
}

func (p *Page) WaitAdventure(timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if p.IsAdventurePage() {
			return true
		}
		p.executor.Sleep(500)
	}
	return false
}

func (p *Page) WaitHome(timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if p.IsKingdomHome() {
			return true
		}
		p.executor.Sleep(500)
	}
	return false
}
