//go:build android && cgo

package ui

import (
	"github.com/Dasongzi1366/AutoGo/imgui"
)

// RunShell 启动 imgui 主循环：悬浮球始终绘制，配置面板按 openPanel 开关。
// 阻塞调用：内部 imgui.Run + select{}。
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
	openPanel := opts.OpenPanelOnStart

	_ = imgui.Init()
	ApplyIndustrialTheme()
	ball := NewFloatingBall()
	loaded := false

	imgui.Run(func() {
		if !loaded {
			loaded = true
			_ = opts.Store.LoadConfig(opts.ConfigPath)
			SeedPanelDefaults(opts.Store)
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
