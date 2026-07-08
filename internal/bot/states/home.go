package states

import (
	"app/internal/action"
	"app/internal/bot"
	"app/internal/utils"
)

// Home represents the game's main / safe screen.
type Home struct {
	battle *Battle
}

// NewHome creates a Home state. Pass a non-nil battle if the bot should
// transition to farming battles from home.
func NewHome(battle *Battle) *Home {
	return &Home{battle: battle}
}

func (h *Home) Name() string { return "home" }

func (h *Home) Detect(ctx *bot.Context) bool {
	// Example: a white-ish home button near the top-left.
	return ctx.Detector.MatchColor(120, 80, "FFFFFF", 0.9)
}

func (h *Home) Act(ctx *bot.Context) error {
	utils.Infof("at home, collecting resources if enabled")
	if ctx.Config.Modules.CollectResources {
		_ = ctx.Executor.Tap(action.Point{X: 200, Y: 300})
		ctx.Executor.Sleep(500)
	}
	return nil
}

func (h *Home) Next(ctx *bot.Context) bot.State {
	if ctx.Config.Modules.FarmLevels && h.battle != nil {
		return h.battle
	}
	return h
}

func (h *Home) Recover(ctx *bot.Context) error {
	// Home is the safe state; nothing to do.
	return nil
}
