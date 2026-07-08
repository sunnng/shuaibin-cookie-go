package guard

import (
	"sort"
	"time"

	"app/internal/logger"
	"app/internal/platform/screen"
)

// Trap represents a popup or interrupt that the Guard should detect and handle.
type Trap struct {
	Name     string
	Feature  any
	Handler  func() error
	Priority int
}

// Guard scans registered popup traps by priority and handles the first match.
// It also provides a segmented sleep that checks guards periodically.
type Guard struct {
	detector screen.Detector
	traps    []Trap
}

// New creates a Guard using the supplied screen detector.
func New(detector screen.Detector) *Guard {
	return &Guard{detector: detector}
}

// Register adds a trap. Traps are kept sorted by descending priority.
func (g *Guard) Register(name string, feature any, handler func() error, priority int) {
	g.traps = append(g.traps, Trap{Name: name, Feature: feature, Handler: handler, Priority: priority})
	sort.SliceStable(g.traps, func(i, j int) bool {
		return g.traps[i].Priority > g.traps[j].Priority
	})
	logger.Infof("[Guard] registered %s priority=%d", name, priority)
}

// Check scans traps in priority order and runs the handler for the first match.
// It returns true if a trap was detected and handled successfully.
func (g *Guard) Check() bool {
	for _, trap := range g.traps {
		if g.match(trap.Feature) {
			logger.Infof("[Guard] hit %s", trap.Name)
			if err := trap.Handler(); err != nil {
				logger.Errorf("[Guard] handle %s failed: %v", trap.Name, err)
				return false
			}
			logger.Infof("[Guard] handled %s", trap.Name)
			return true
		}
	}
	return false
}

// Sleep pauses for the requested duration, calling Check every 500ms so that
// popup traps can be handled during the wait.
func (g *Guard) Sleep(d time.Duration) {
	g.SleepWithInterval(d, 500*time.Millisecond)
}

// SleepWithInterval pauses for the requested duration, calling Check before each
// step chunk so that popup traps can be handled during the wait.
func (g *Guard) SleepWithInterval(d, step time.Duration) {
	if step <= 0 {
		step = 500 * time.Millisecond
	}
	left := d
	for left > 0 {
		g.Check()
		chunk := step
		if left < chunk {
			chunk = left
		}
		time.Sleep(chunk)
		left -= chunk
	}
}

// match evaluates a feature against the current screen.
// Supported features:
//   - string: interpreted as a multi-color string and passed to MatchMultiColor.
//   - func() bool: a custom predicate; true means the trap matched.
//   - other: currently returns false (structured screen.Feature support pending).
func (g *Guard) match(feature any) bool {
	switch f := feature.(type) {
	case string:
		return g.detector.MatchMultiColor(f, 0.9)
	case func() bool:
		return f()
	default:
		return false
	}
}
