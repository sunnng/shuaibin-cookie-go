//go:build android && cgo

package ui

import (
	"github.com/Dasongzi1366/AutoGo/imgui"
)

// RunShell 启动 imgui 主循环：悬浮球始终绘制，配置面板按 openPanel 开关。
// 阻塞调用：内部 imgui.Run + select{}。
// main 应传 OpenPanelOnStart: true（spec 默认打开面板）；零值 false 时仅显示悬浮球。
func RunShell(opts ShellOptions) {
	if opts.Store == nil {
		opts.Store = NewStore()
	}
	if opts.Title == "" {
		opts.Title = "帅宾 Cookie"
	}
	if opts.ConfigPath == "" {
		opts.ConfigPath = "/sdcard/shuaibin-cookie/ui.json"
	}
	if opts.CountdownSec <= 0 {
		opts.CountdownSec = 300
	}
	ConfigureCookiePanel(CookiePanelOptions{
		ConfigPath:    opts.ConfigPath,
		DataStorePath: opts.DataStorePath,
		Controller:    opts.Controller,
		Reseed:        opts.Reseed,
	})
	if opts.Render == nil {
		opts.Render = DefaultCookiePanel
	}
	openPanel := opts.OpenPanelOnStart

	_ = imgui.Init()
	ApplyIndustrialTheme()
	EnsureBuiltinModules()
	ball := NewFloatingBall()
	loaded := false

	imgui.Run(func() {
		if !loaded {
			loaded = true
			_ = opts.Store.LoadConfig(opts.ConfigPath)
			SeedHubDefaults(opts.Store)
		}

		state := StateIdle
		if opts.Controller != nil {
			state = opts.Controller.State()
		}

		ball.Draw(BallCallbacks{
			OnSettings: func() { openPanel = true },
			OnStartStop: func() {
				if opts.Controller == nil {
					return
				}
				switch opts.Controller.State() {
				case StateIdle:
					_ = opts.Store.SaveConfig(opts.ConfigPath)
					opts.Controller.Start()
				case StateRunning, StatePaused:
					opts.Controller.Stop()
				}
			},
			OnPauseResume: func() {
				if opts.Controller == nil {
					return
				}
				switch opts.Controller.State() {
				case StateRunning:
					opts.Controller.Pause()
				case StatePaused:
					opts.Controller.Resume()
				}
			},
			OnClose: func() {
				if opts.Controller != nil {
					opts.Controller.Exit()
				}
			},
		}, state)

		if openPanel {
			stillOpen := drawConfigPanel(opts, &openPanel)
			if !stillOpen {
				openPanel = false
			}
		}
	})
	select {}
}

// RunPanel 兼容旧 PanelOptions：无 Controller 时用 OnRun 包一层 SessionController，
// 然后委托给 RunShell（默认打开面板）。
func RunPanel(opts PanelOptions) {
	var ctrl Controller
	if opts.OnRun != nil {
		ctrl = NewSessionController(SessionHooks{
			OnStart: func() (func() error, func(), func(), func()) {
				run := func() error {
					opts.OnRun(opts.Store)
					return nil
				}
				return run, func() {}, func() {}, func() {}
			},
			OnExit: func() {},
		})
	}
	RunShell(ShellOptions{
		Title:            opts.Title,
		ConfigPath:       opts.ConfigPath,
		CountdownSec:     opts.CountdownSec,
		Store:            opts.Store,
		Render:           opts.Render,
		Controller:       ctrl,
		OpenPanelOnStart: true,
	})
}
