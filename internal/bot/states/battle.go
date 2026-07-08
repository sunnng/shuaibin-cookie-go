package states

import (
	"app/internal/platform/action"
	"app/internal/bot"
	"app/internal/utils"
)

// Battle represents an in-battle / pre-battle screen.
type Battle struct {
	home *Home
}

// NewBattle creates a Battle state. The home reference can be set later via
// SetHome to break the circular dependency between Home and Battle.
func NewBattle(home *Home) *Battle {
	return &Battle{home: home}
}

// SetHome sets the home state after both states have been constructed.
func (b *Battle) SetHome(home *Home) {
	b.home = home
}

func (b *Battle) Name() string { return "battle" }

func (b *Battle) Detect(ctx *bot.Context) bool {
	// Example: detect a battle-start button via multi-color.
	return ctx.Detector.MatchMultiColor("800,500,FFAA00-101010,820,500,FFAA00-101010", 0.9)
}

func (b *Battle) Act(ctx *bot.Context) error {
	utils.Infof("starting battle")
	_ = ctx.Executor.Tap(action.Point{X: 800, Y: 500})
	ctx.Executor.Sleep(2000)
	if ctx.Detector.MatchColor(800, 200, "00FF00", 0.8) {
		_ = ctx.Executor.Tap(action.Point{X: 800, Y: 200})
	}
	return nil
}

func (b *Battle) Next(ctx *bot.Context) bot.State {
	if b.home != nil {
		return b.home
	}
	return b
}

func (b *Battle) Recover(ctx *bot.Context) error {
	// Exit battle: back -> confirm.
	_ = ctx.Executor.Back()
	ctx.Executor.Sleep(800)
	_ = ctx.Executor.Tap(action.Point{X: 800, Y: 600})
	ctx.Executor.Sleep(500)
	return nil
}
