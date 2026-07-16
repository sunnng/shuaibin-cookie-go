package main

import (
	"os"
	"time"

	"app/internal/config"
	"app/internal/game/arena"
	"app/internal/game/common/kingdom"
	"app/internal/guard"
	"app/internal/logger"
	"app/internal/platform/action"
	"app/internal/platform/screen"
	"app/internal/runtime"
	"app/internal/scheduler"
	"app/internal/store"
	"app/internal/ui"
)

func main() {
	logger.SetLevel(logger.LevelInfo)
	logger.Infof("superbin cookie run kingdom start...")

	cfg, err := config.LoadConfig("config.json")
	if err != nil {
		logger.Errorf("failed to load config: %v", err)
		return
	}

	uiStore := ui.NewStore()
	ui.SeedFromConfig(uiStore, cfg)

	ctrl := ui.NewSessionController(ui.SessionHooks{
		OnStart: func() (run func() error, pause, resume, stop func()) {
			ui.ApplyToConfig(uiStore, cfg)
			rt := buildRuntime(cfg)
			return rt.Run, rt.Pause, rt.Resume, rt.Stop
		},
		OnExit: func() {
			time.Sleep(500 * time.Millisecond)
			os.Exit(0)
		},
	})

	ui.RunShell(ui.ShellOptions{
		Title:            "Superbin Cookie",
		ConfigPath:       "/sdcard/shuaibin-cookie/ui.json",
		DataStorePath:    "/sdcard/shuaibin-cookie/store.json",
		Store:            uiStore,
		Controller:       ctrl,
		OpenPanelOnStart: true,
		Reseed: func(s *ui.Store) {
			ui.SeedFromConfig(s, cfg)
		},
	})
}

func buildRuntime(cfg *config.Config) *runtime.Runtime {
	det := screen.NewDetector(0)
	exec := action.NewExecutor(0)
	s := store.New("/sdcard/shuaibin-cookie/store.json")
	g := guard.New(det)
	sched := scheduler.New()

	kingdomFeature := kingdom.DefaultFeature()
	kingdomPage := kingdom.NewPage(det, exec, kingdomFeature)
	arenaFeature := arena.DefaultFeature()
	arenaSession := arena.NewSession(s)
	arenaTask := arena.NewTask(
		&cfg.Modules.Arena,
		det, exec, arenaFeature, kingdomPage, arenaSession, g,
	)

	sched.Build(scheduler.TaskOpts{
		Name: "王国竞技场",
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
		RuntimeCfg: runtime.RuntimeConfig{
			GuardInterval: 500 * time.Millisecond,
			IdleDelay:     30 * time.Second,
			StepDelay:     5 * time.Second,
			StopOnError:   false,
		},
	})
	arenaTask.SetShouldStop(rt.IsStopped)
	return rt
}
