package states

import "app/internal/bot"

// Unknown is a placeholder state used before the machine has identified
// the current screen. It never detects itself and has no actions.
type Unknown struct{}

func NewUnknown() *Unknown { return &Unknown{} }

func (u *Unknown) Name() string                    { return "unknown" }
func (u *Unknown) Detect(ctx *bot.Context) bool    { return false }
func (u *Unknown) Act(ctx *bot.Context) error      { return nil }
func (u *Unknown) Next(ctx *bot.Context) bot.State { return nil }
func (u *Unknown) Recover(ctx *bot.Context) error  { return nil }
