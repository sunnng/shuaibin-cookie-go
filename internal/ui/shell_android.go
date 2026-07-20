//go:build android && cgo

package ui

import (
	"github.com/Dasongzi1366/AutoGo/imgui"
)

// RunShell 启动 imgui 主循环：灵动岛（刘海样式悬浮窗）始终绘制，配置面板按 openPanel 开关。
// 阻塞调用：内部 imgui.Run + select{}。
func RunShell(opts ShellOptions) {
	if opts.Store == nil {
		opts.Store = NewStore()
	}
	if opts.Title == "" {
		opts.Title = "帅宾 Cookie"
	}
	if opts.ConfigPath == "" {
		opts.ConfigPath = DefaultConfigPath
	}
	openPanel := opts.OpenPanelOnStart
	// autoPaused 记录"面板遮挡导致的自动暂停"：只有 UI 自己暂停的才在关面板后自动恢复，
	// 用户手动暂停/停止的不动。
	autoPaused := false

	_ = imgui.Init()
	ApplyQQBlueTheme()
	island := NewFloatingIsland()
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
		// 运行中且任务上报了状态时，灵动岛展示任务状态（如"竞技场 3/10 · 胜率 67%"）。
		taskStatus := ""
		if opts.Status != nil && state == StateRunning {
			taskStatus = opts.Status.Text()
		}

		island.Draw(IslandCallbacks{
			OnSettings: func() {
				openPanel = true
				// 面板大面积遮挡画面，开着面板跑识别必失败：运行中则自动暂停，关面板后恢复。
				if opts.Controller != nil && opts.Controller.State() == StateRunning {
					opts.Controller.Pause()
					autoPaused = true
				}
			},
			OnStartStop: func() {
				if opts.Controller == nil {
					return
				}
				switch opts.Controller.State() {
				case StateIdle:
					_ = opts.Store.SaveConfig(opts.ConfigPath)
					opts.Controller.Start()
					// 面板还开着就启动脚本同样会被遮挡，等同打开面板处理。
					if openPanel && opts.Controller.State() == StateRunning {
						opts.Controller.Pause()
						autoPaused = true
					}
				case StateRunning, StatePaused:
					autoPaused = false
					opts.Controller.Stop()
				}
			},
			OnPauseResume: func() {
				if opts.Controller == nil {
					return
				}
				// 用户手动接管暂停状态，自动恢复失效。
				autoPaused = false
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
		}, state, taskStatus)

		if openPanel {
			stillOpen := drawConfigPanel(opts, &openPanel)
			if !stillOpen {
				openPanel = false
			}
		}
		if !openPanel && autoPaused {
			autoPaused = false
			if opts.Controller != nil && opts.Controller.State() == StatePaused {
				opts.Controller.Resume()
			}
		}
	})
	select {}
}
