//go:build android && cgo

package ui

import (
	"github.com/Dasongzi1366/AutoGo/device"
	"github.com/Dasongzi1366/AutoGo/imgui"
)

// drawConfigPanel 绘制配置窗。返回 *open：用户关掉窗口（X）或点击「运行脚本」后返回 false。
// 关闭 X：仅 *open=false，不退出进程。
// 点击「运行脚本」：SaveConfig + Controller.Start + *open=false。
func drawConfigPanel(opts ShellOptions, open *bool) bool {
	width, height, _, _ := device.GetDisplayInfo(0)
	winW, winH := float32(700), float32(700)
	x := (float32(width) - winW) / 2
	y := 15 + (float32(height)-15-winH)/2

	imgui.PushStyleVarFloat(imgui.StyleVarWindowRounding, 12)
	defer imgui.PopStyleVar()

	imgui.SetNextWindowSizeV(imgui.Vec2{X: winW, Y: winH}, imgui.CondOnce)
	imgui.SetNextWindowPosV(imgui.Vec2{X: x, Y: y}, imgui.CondOnce, imgui.Vec2{X: 0, Y: 0})

	flags := imgui.WindowFlagsNoMove | imgui.WindowFlagsNoScrollbar | imgui.WindowFlagsNoResize
	if imgui.BeginV(opts.Title, open, flags) {
		UI_创建按钮("overtime", "加时", func() {
			UI_倒计时加时("run_script", 60)
		}, float32(150), float32(-2))
		imgui.SameLine()
		UI_创建倒计时按钮(opts.CountdownSec, "run_script", "运行脚本", func() {
			_ = opts.Store.SaveConfig(opts.ConfigPath)
			if opts.Controller != nil {
				opts.Controller.Start()
			}
			*open = false
		}, float32(-1), float32(-2))

		opts.Render(opts.Store)
	}
	imgui.End()

	return *open
}
