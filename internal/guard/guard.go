package guard

import (
	"sort"

	"app/internal/logger"
	"app/internal/platform/screen"
)

// Trap represents a popup or interrupt that the Guard should detect and handle.
type Trap struct {
	Name     string
	Feature  screen.Feature
	Handler  func() error
	Priority int
}

// Guard scans registered popup traps by priority and handles the first match.
// Runtime 主循环与任务状态机（RunOptions.Guard）都会周期性调用 Check。
type Guard struct {
	detector screen.Detector
	traps    []Trap
}

// New creates a Guard using the supplied screen detector.
func New(detector screen.Detector) *Guard {
	return &Guard{detector: detector}
}

// Register adds a trap. Traps are kept sorted by descending priority.
// A trap whose Feature.Colors is empty (not yet color-picked) is skipped with
// a warning instead of being registered as a silently-dead entry; Sim <= 0
// falls back to 0.9.
func (g *Guard) Register(name string, feature screen.Feature, handler func() error, priority int) {
	if feature.Colors == "" {
		logger.Warnf("[Guard] skip %s: feature colors empty", name)
		return
	}
	if feature.Sim <= 0 {
		feature.Sim = 0.9
	}
	g.traps = append(g.traps, Trap{Name: name, Feature: feature, Handler: handler, Priority: priority})
	sort.SliceStable(g.traps, func(i, j int) bool {
		return g.traps[i].Priority > g.traps[j].Priority
	})
	logger.Infof("[Guard] registered %s priority=%d", name, priority)
}

// Check scans traps in priority order and runs the handler for the first match.
// It returns true if a trap was detected and handled successfully.
func (g *Guard) Check() bool {
	if g.detector == nil {
		return false
	}
	for _, trap := range g.traps {
		if !g.detector.MatchMultiColor(trap.Feature.Colors, trap.Feature.Sim) {
			continue
		}
		logger.Infof("[Guard] hit %s", trap.Name)
		if err := trap.Handler(); err != nil {
			logger.Errorf("[Guard] handle %s failed: %v", trap.Name, err)
			return false
		}
		logger.Infof("[Guard] handled %s", trap.Name)
		return true
	}
	return false
}
