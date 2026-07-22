//go:build android && cgo

package ui

import (
	"github.com/Dasongzi1366/AutoGo/device"
	"github.com/Dasongzi1366/AutoGo/imgui"
)

// RunShell 启动框架 UI 主循环（阻塞）：灵动岛始终绘制，配置面板按
// Shell 状态开关。首帧加载持久化配置并 Seed。
func RunShell(opts ShellOptions) {
	shell := NewShell(opts)

	_ = imgui.Init()
	ApplyTheme(shell.Theme())

	w, h, _, _ := device.GetDisplayInfo(0)
	bw, bh := shell.BaseSize()
	ctx := NewCtx(shell.Store(), ComputeScale(w, h, bw, bh))
	ctx.Theme = shell.Theme()
	ctx.Shell = shell

	island := newFloatingIsland()
	loaded := false

	imgui.Run(func() {
		if !loaded {
			loaded = true
			if shell.ConfigPath() != "" {
				_ = shell.Store().LoadConfig(shell.ConfigPath())
			}
			shell.Seed()
		}

		island.Draw(ctx, shell)
		if shell.PanelOpen() {
			drawPanel(ctx, shell)
		}
	})
	select {}
}
