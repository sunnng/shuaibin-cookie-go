package kingdom

import (
	"time"

	"app/internal/platform/action"
	"app/internal/platform/screen"
)

type Page struct {
	detector screen.Detector
	executor action.Executor
	feature  *Feature
}

func NewPage(det screen.Detector, exec action.Executor, f *Feature) *Page {
	return &Page{detector: det, executor: exec, feature: f}
}

func (p *Page) IsKingdomHome() bool {
	return p.detector.MatchMultiColor("", 0.9) // placeholder
}

func (p *Page) TapAdventureBtn() {
	_ = p.executor.Tap(action.Point{X: 100, Y: 100}) // placeholder
	p.executor.Sleep(1200)
}

func (p *Page) IsAdventurePage() bool {
	return p.detector.MatchMultiColor("", 0.9) // placeholder
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
