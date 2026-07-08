package main

import (
	"time"

	"app/internal/config"
	"app/internal/game/arena"
	"app/internal/game/common/kingdom"
	"app/internal/guard"
	"app/internal/hud"
	"app/internal/logger"
	"app/internal/platform/action"
	"app/internal/platform/screen"
	"app/internal/runtime"
	"app/internal/scheduler"
	"app/internal/store"
)

func main() {
	logger.SetLevel(logger.LevelInfo)
	logger.Infof("bot starting")

	cfg, err := config.LoadConfig("config.json")
	if err != nil {
		logger.Errorf("failed to load config: %v", err)
		return
	}

	det := screen.NewDetector(0)
	exec := action.NewExecutor(0)

	s := store.New("data/store.json")
	h := hud.New()
	g := guard.New(det)
	sched := scheduler.New()

	// Register global guard traps here if needed

	// Common kingdom
	kingdomFeature := kingdom.DefaultFeature()
	kingdomPage := kingdom.NewPage(det, exec, kingdomFeature)

	// Arena
	arenaFeature := arena.DefaultFeature()
	arenaSession := arena.NewSession(s)
	arenaTask := arena.NewTask(
		&cfg.Modules.Arena,
		det,
		exec,
		arenaFeature,
		kingdomPage,
		kingdomFeature,
		arenaSession,
		g,
	)

	sched.Build(scheduler.TaskOpts{
		Name:      "王国竞技场",
		ConfigKey: "arena",
		CheckEnabled: func() bool {
			return cfg.Modules.Arena.Enabled
		},
		CheckReady: func() (bool, time.Duration) {
			if arenaSession.IsReachMaxBattles(&cfg.Modules.Arena) {
				return false, 0
			}
			remain := arenaSession.TimeUntilRefresh()
			if remain > 0 {
				return false, remain
			}
			return true, 0
		},
		WaitHUD: func(remain time.Duration) string {
			return "免费刷新等待"
		},
		Action: arenaTask.Run,
	})

	rt := runtime.New(runtime.Options{
		Scheduler: sched,
		Guard:     g,
		HUD:       h,
		RuntimeCfg: runtime.RuntimeConfig{
			GuardInterval: 500 * time.Millisecond,
			IdleDelay:     30 * time.Second,
			StepDelay:     5 * time.Second,
			StopOnError:   false,
		},
	})

	if err := rt.Run(); err != nil {
		logger.Errorf("runtime stopped: %v", err)
	}
}
