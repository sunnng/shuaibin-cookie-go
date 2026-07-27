package main

import (
	"os"
	"time"

	"app/internal/config"
	"app/internal/game/arena"
	"app/internal/game/biscuit"
	"app/internal/game/common/kingdom"
	"app/internal/game/common/popup"
	"app/internal/game/market"
	"app/internal/game/mine"
	"app/internal/game/mine/battle"
	"app/internal/game/mine/jelly"
	"app/internal/game/mine/mining"
	"app/internal/game/mine/survey"
	"app/internal/game/square"
	"app/internal/game/starlight"
	"app/internal/guard"
	"app/internal/logger"
	"app/internal/platform/action"
	"app/internal/platform/screen"
	"app/internal/runtime"
	"app/internal/scheduler"
	"app/internal/status"
	"app/internal/store"
	"app/internal/taskdesc"
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
	tasks := taskdesc.Tasks(cfg)

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
		Title: "帅斌饼干助手",
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

	popup.Register(g, exec) // 「网络联机状态不稳定」弹窗守卫

	kingdomPage := kingdom.NewPage(det, exec, kingdom.DefaultFeature())
	mineHome := mine.NewPage(det, exec, mine.DefaultFeature())
	mineRoute := mine.NewRoute(mineHome, kingdomPage)

	// ---- 任务与状态构造（全部在 sched.Build 之前完成）----

	surveyCfg := &survey.Config{
		Enabled:     cfg.Tasks.MineVenture.Enabled,
		TargetFloor: cfg.Tasks.MineVenture.TargetFloor,
		FarGap:      cfg.Tasks.MineVenture.FarGap,
		OCRPollSec:  cfg.Tasks.MineVenture.OCRPollSec,
		FarWaitSec:  cfg.Tasks.MineVenture.FarWaitSec,
	}
	surveyState := survey.NewState(s)
	surveyTask := survey.NewTask(surveyCfg, det, exec, survey.DefaultFeature(),
		mineHome, mineRoute, kingdomPage, surveyState, g)

	oreCards := cfg.Tasks.OreVeinMining.OreCards
	if len(oreCards) == 0 { // 空配置回退包内默认优先级（Lua resolveCardPriority）
		oreCards = append([]string(nil), mining.DefaultCardPriority...)
	}
	miningCfg := &mining.Config{
		Enabled:     cfg.Tasks.OreVeinMining.Enabled,
		IntervalSec: cfg.Tasks.OreVeinMining.IntervalSec,
		OreCards:    oreCards,
	}
	miningState := mining.NewState(s)
	miningTask := mining.NewTask(miningCfg, det, exec, mining.DefaultFeature(),
		mineHome, mineRoute, kingdomPage, miningState, g)

	battleCfg := &battle.Config{
		Enabled:     cfg.Tasks.MineBattle.Enabled,
		IntervalSec: cfg.Tasks.MineBattle.IntervalSec,
		SoulStones:  cfg.Tasks.MineBattle.SoulStones,
	}
	battleState := battle.NewState(s)
	battleTask := battle.NewTask(battleCfg, det, exec, battle.DefaultFeature(),
		mineHome, mineRoute, kingdomPage, battleState, g)

	jellyCfg := &jelly.Config{
		Enabled:     cfg.Tasks.MeltedAgarCubes.Enabled,
		IntervalSec: cfg.Tasks.MeltedAgarCubes.IntervalSec,
	}
	jellyState := jelly.NewState(s)
	jellyTask := jelly.NewTask(jellyCfg, det, exec, jelly.DefaultFeature(),
		mineHome, mineRoute, kingdomPage, jellyState, g)

	marketCfg := &market.Config{
		Enabled:          cfg.Tasks.SeasideMarket.Enabled,
		Items:            cfg.Tasks.SeasideMarket.Items,
		RestockBufferSec: cfg.Tasks.SeasideMarket.RestockBufferSec,
	}
	marketState := market.NewState(s)
	marketTask := market.NewTask(marketCfg, det, exec, market.DefaultFeature(),
		kingdomPage, marketState, g)

	arenaFeature := arena.DefaultFeature()
	arenaState := arena.NewState(s)
	arenaTask := arena.NewTask(
		&cfg.Tasks.Arena,
		det, exec, arenaFeature, kingdomPage, arenaState, g,
	)

	starlightCfg := &starlight.Config{Enabled: cfg.Tasks.Starlight.Enabled}
	starlightState := starlight.NewState(s)
	starlightTask := starlight.NewTask(starlightCfg, det, exec, starlight.DefaultFeature(),
		kingdomPage, starlightState, g)

	squareCfg := &square.Config{
		Enabled:          cfg.Tasks.Square.Enabled,
		DailyCap:         cfg.Tasks.Square.DailyCap,
		CheckIntervalSec: cfg.Tasks.Square.CheckIntervalSec,
		ChunkSec:         cfg.Tasks.Square.ChunkSec,
	}
	squareState := square.NewState(s)
	squareTask := square.NewTask(squareCfg, det, exec, square.DefaultFeature(),
		kingdomPage, squareState, g)

	biscuitCfg := &biscuit.Config{
		Enabled:  cfg.Tasks.Biscuit.Enabled,
		MaxRolls: cfg.Tasks.Biscuit.MaxRolls,
	}
	for _, t := range cfg.Tasks.Biscuit.Targets {
		biscuitCfg.Targets = append(biscuitCfg.Targets, biscuit.TargetRule{
			Enabled: t.Enabled, Name: t.Name, MinPercent: t.MinPercent,
		})
	}
	for _, r := range cfg.Tasks.Biscuit.SumRules {
		biscuitCfg.SumRules = append(biscuitCfg.SumRules, biscuit.SumRule{
			Enabled: r.Enabled, Name: r.Name, Count: r.Count, MinSum: r.MinSum,
		})
	}
	biscuitState := biscuit.NewState(s)
	biscuitTask := biscuit.NewTask(biscuitCfg, det, exec, biscuit.DefaultFeature(), biscuitState)

	// ---- 调度辅助闭包 ----

	// mineIdle 对应 Lua isMineSchedulerIdle：只查勘查/开采/洋菜冻（不含战斗），
	// 任一矿山调度就绪时其它非矿山任务让路。
	mineIdle := func() bool {
		if cfg.Tasks.MineVenture.Enabled {
			if ready, _ := surveyState.CheckFarWait(); ready {
				return false
			}
		}
		if cfg.Tasks.OreVeinMining.Enabled {
			if ready, _ := miningState.CheckReady(); ready {
				return false
			}
		}
		if cfg.Tasks.MeltedAgarCubes.Enabled {
			if ready, _ := jellyState.CheckReady(); ready {
				return false
			}
		}
		return true
	}

	// leaveSquare 对应 Lua leaveSquare：除广场外的任务运行前若卡在广场先离开
	// （停留进度由 square.Task.Leave 保留）。先判断避免对未知页面空等。
	leaveSquare := func(run func() error) func() error {
		return func() error {
			if squareTask.InSquareContext() {
				squareTask.Leave()
			}
			return run()
		}
	}

	// ---- 注册（顺序严格按 Lua register.lua）----

	sched.Build(scheduler.TaskOpts{
		Name:         "矿山勘查",
		CheckEnabled: func() bool { return cfg.Tasks.MineVenture.Enabled },
		CheckReady:   surveyState.CheckFarWait,
		WaitHUD:      func(remain time.Duration) string { return "远距等待" },
		Action:       leaveSquare(surveyTask.Run),
	})

	sched.Build(scheduler.TaskOpts{
		Name:         "矿山开采",
		CheckEnabled: func() bool { return cfg.Tasks.OreVeinMining.Enabled },
		CheckReady: func() (bool, time.Duration) {
			if miningTask.CanResume() {
				return true, 0
			}
			return miningState.CheckReady()
		},
		WaitHUD: func(remain time.Duration) string { return "busy 等待" },
		Action:  leaveSquare(miningTask.Run),
	})

	battleInterval := func() time.Duration {
		return time.Duration(cfg.Tasks.MineBattle.IntervalSec) * time.Second
	}
	sched.Build(scheduler.TaskOpts{
		Name:         "矿山战斗",
		CheckEnabled: func() bool { return cfg.Tasks.MineBattle.Enabled },
		CheckReady: func() (bool, time.Duration) {
			return battleState.CheckReady(battleInterval())
		},
		WaitHUD: func(remain time.Duration) string { return "冷却等待" },
		Action:  leaveSquare(battleTask.Run),
	})

	sched.Build(scheduler.TaskOpts{
		Name:         "解除洋菜冻",
		CheckEnabled: func() bool { return cfg.Tasks.MeltedAgarCubes.Enabled },
		CheckReady:   jellyState.CheckReady,
		WaitHUD:      func(remain time.Duration) string { return "冷却等待" },
		Action:       leaveSquare(jellyTask.Run),
	})

	sched.Build(scheduler.TaskOpts{
		Name: "海滩交易所",
		// precondition：矿山调度空闲时才跑（Lua isMineSchedulerIdle）。
		CheckEnabled: func() bool { return cfg.Tasks.SeasideMarket.Enabled && mineIdle() },
		// marketState.CheckReady 有首轮强制副作用（启动后首次调用直接就绪），如实保留。
		CheckReady: marketState.CheckReady,
		WaitHUD:    func(remain time.Duration) string { return "补货等待" },
		Action:     leaveSquare(marketTask.Run),
	})

	sched.Build(scheduler.TaskOpts{
		Name: "王国竞技场",
		CheckEnabled: func() bool {
			return cfg.Tasks.Arena.Enabled && mineIdle() // precondition
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
		Action: leaveSquare(arenaTask.Run),
	})

	sched.Build(scheduler.TaskOpts{
		Name:         "梦幻繁星岛",
		CheckEnabled: func() bool { return cfg.Tasks.Starlight.Enabled && mineIdle() }, // precondition
		CheckReady: func() (bool, time.Duration) {
			if starlightState.IsDoneToday() {
				return false, starlightState.TimeUntilNextDay()
			}
			return true, 0
		},
		Action: leaveSquare(starlightTask.Run),
	})

	sched.Build(scheduler.TaskOpts{
		Name:         "布谷鸟广场",
		CheckEnabled: func() bool { return cfg.Tasks.Square.Enabled && mineIdle() }, // precondition
		CheckReady:   squareState.CheckReady,
		Action:       squareTask.Run, // 广场自身不 leaveSquare
	})

	sched.Build(scheduler.TaskOpts{
		Name: "洗脆饼词条",
		CheckEnabled: func() bool {
			return cfg.Tasks.Biscuit.Enabled && !biscuitState.IsGraduated()
		},
		Action: leaveSquare(biscuitTask.Run),
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
	surveyTask.SetShouldStop(rt.IsStopped)
	miningTask.SetShouldStop(rt.IsStopped)
	battleTask.SetShouldStop(rt.IsStopped)
	jellyTask.SetShouldStop(rt.IsStopped)
	marketTask.SetShouldStop(rt.IsStopped)
	arenaTask.SetShouldStop(rt.IsStopped)
	starlightTask.SetShouldStop(rt.IsStopped)
	squareTask.SetShouldStop(rt.IsStopped)
	biscuitTask.SetShouldStop(rt.IsStopped)

	// 灵动岛状态上报（mine 系四个任务无 SetStatusReporter，不接）。
	marketTask.SetStatusReporter(statusReporter)
	arenaTask.SetStatusReporter(statusReporter)
	starlightTask.SetStatusReporter(statusReporter)
	squareTask.SetStatusReporter(statusReporter)
	biscuitTask.SetStatusReporter(statusReporter)
	return rt
}
