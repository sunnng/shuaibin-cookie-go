package main

import (
	"time"

	"app/internal/action"
	"app/internal/bot"
	"app/internal/bot/states"
	"app/internal/config"
	"app/internal/screen"
	"app/internal/utils"
)

func main() {
	utils.Infof("bot starting")

	cfg, err := config.LoadConfig("config.json")
	if err != nil {
		utils.Errorf("failed to load config: %v", err)
		return
	}

	detector := screen.NewDetector(0)
	executor := action.NewExecutor(0)

	unknown := states.NewUnknown()
	battle := states.NewBattle(nil)
	home := states.NewHome(battle)
	battle.SetHome(home)

	reg := bot.NewRegistry()
	reg.Register(home)
	reg.Register(battle)
	reg.Register(unknown)

	ctx := &bot.Context{
		Config:    cfg,
		Detector:  detector,
		Executor:  executor,
		Current:   unknown,
		EnteredAt: time.Now(),
	}

	machine := bot.NewMachine(ctx, reg)
	if err := machine.Run(); err != nil {
		utils.Errorf("machine stopped: %v", err)
	}
}
