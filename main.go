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
	"app/internal/status"
	"app/internal/store"
	"app/ui"
)

// 设备上的持久化路径（原 internal/ui/options.go，全项目唯一来源）。
const (
	defaultDataDir    = "/sdcard/shuaibin-cookie"
	defaultConfigPath = defaultDataDir + "/ui.json"
	defaultStorePath  = defaultDataDir + "/store.json"
)

func main() {
	logger.SetLevel(logger.LevelInfo)
	logger.Infof("shuaibin cookie run kingdom start...")

	cfg, err := config.LoadConfig("config.json")
	if err != nil {
		logger.Errorf("failed to load config: %v", err)
		return
	}

	ui.LogErrorf = logger.Errorf
	uiStore := ui.NewStore()
	statusReporter := status.New()
	tasks := []ui.Task{arenaTaskDescriptor(cfg)}

	ctrl := ui.NewScriptController(ui.ScriptHooks{
		OnStart: func() (run func() error, pause, resume, stop func()) {
			ui.ApplyAll(uiStore, tasks) // 面板值回写 cfg 后重建运行时
			rt := buildRuntime(cfg, statusReporter)
			return rt.Run, rt.Pause, rt.Resume, rt.Stop
		},
		OnExit: func() {
			time.Sleep(500 * time.Millisecond)
			os.Exit(0)
		},
	})

	ui.RunShell(ui.ShellOptions{
		Title: "帅宾 Cookie",
		Tasks: tasks,
		Nav: []ui.NavEntry{
			{ID: "tasks", Title: "任务", Render: ui.TaskListPage()},
			{ID: "system", Title: "系统", Render: ui.SystemPage()},
		},
		Store:            uiStore,
		Controller:       ctrl,
		Status:           statusReporter,
		ConfigPath:       defaultConfigPath,
		DataStorePath:    defaultStorePath,
		OpenPanelOnStart: true,
	})
}

func buildRuntime(cfg *config.Config, statusReporter *status.Reporter) *runtime.Runtime {
	det := screen.NewDetector(0)
	exec := action.NewExecutor(0)
	s := store.New(defaultStorePath)
	g := guard.New(det)
	sched := scheduler.New()

	kingdomFeature := kingdom.DefaultFeature()
	kingdomPage := kingdom.NewPage(det, exec, kingdomFeature)
	arenaFeature := arena.DefaultFeature()
	arenaState := arena.NewState(s)
	arenaTask := arena.NewTask(
		&cfg.Tasks.Arena,
		det, exec, arenaFeature, kingdomPage, arenaState, g,
	)

	sched.Build(scheduler.TaskOpts{
		Name: "王国竞技场",
		CheckEnabled: func() bool {
			return cfg.Tasks.Arena.Enabled
		},
		CheckReady: func() (bool, time.Duration) {
			if arenaState.IsReachMaxBattles(&cfg.Tasks.Arena) {
				return false, 0
			}
			remain := arenaState.TimeUntilRefresh()
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
	arenaTask.SetStatusReporter(statusReporter)
	return rt
}
