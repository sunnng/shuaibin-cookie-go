//go:build android && cgo

package ui

import (
	"fmt"

	"github.com/Dasongzi1366/AutoGo/device"
	"github.com/Dasongzi1366/AutoGo/imgui"
)

// drawConfigPanel 工业风窗壳：自适应宽高 + 会话条 + 单一主 CTA（无倒计时双入口）。
func drawConfigPanel(opts ShellOptions, open *bool) bool {
	width, height, _, _ := device.GetDisplayInfo(0)
	winW := float32(width) * 0.72
	if winW > 980 {
		winW = 980
	}
	if winW < 640 {
		winW = 640
	}
	winH := float32(height) * 0.78
	if winH > 700 {
		winH = 700
	}
	if winH < 480 {
		winH = 480
	}
	x := (float32(width) - winW) / 2
	y := 12 + (float32(height)-12-winH)/2

	imgui.SetNextWindowSizeV(imgui.Vec2{X: winW, Y: winH}, imgui.CondOnce)
	imgui.SetNextWindowPosV(imgui.Vec2{X: x, Y: y}, imgui.CondOnce, imgui.Vec2{X: 0, Y: 0})

	flags := imgui.WindowFlagsNoMove | imgui.WindowFlagsNoScrollbar | imgui.WindowFlagsNoResize
	if imgui.BeginV(opts.Title, open, flags) {
		renderSessionBar(opts, open)
		imgui.Separator()
		opts.Render(opts.Store)
	}
	imgui.End()

	return *open
}

func renderSessionBar(opts ShellOptions, open *bool) {
	state := StateIdle
	if opts.Controller != nil {
		state = opts.Controller.State()
	}
	stateLabel := "空闲"
	switch state {
	case StateRunning:
		stateLabel = "运行中"
	case StatePaused:
		stateLabel = "已暂停"
	}

	en, total := 0, 0
	if opts.Store != nil {
		EnsureBuiltinModules()
		en, total = CountEnabled(opts.Store)
	}

	imgui.AlignTextToFramePadding()
	imgui.Text(fmt.Sprintf("会话  %s", stateLabel))
	imgui.SameLine()
	imgui.TextDisabled(fmt.Sprintf("启用 %d/%d", en, total))
	imgui.SameLine()

	const padX, padY = float32(14), float32(8)
	pauseLabel := "暂停"
	if state == StatePaused {
		pauseLabel = "继续"
	}
	startLabel := "开始"
	if state == StateRunning || state == StatePaused {
		startLabel = "停止"
	}
	pauseW, pauseH := fitButtonSize(pauseLabel, padX, padY)
	startW, startH := fitButtonSize(startLabel, padX, padY)
	btnH := pauseH
	if startH > btnH {
		btnH = startH
	}
	gap := float32(8)
	avail := imgui.ContentRegionAvail()
	imgui.SetCursorPosX(imgui.CursorPosX() + avail.X - pauseW - startW - gap)

	imgui.PushStyleVarVec2(imgui.StyleVarFramePadding, imgui.Vec2{X: padX, Y: padY})
	defer imgui.PopStyleVar()

	if state == StateRunning {
		if imgui.ButtonV(pauseLabel+"##session", imgui.Vec2{X: pauseW, Y: btnH}) {
			if opts.Controller != nil {
				opts.Controller.Pause()
			}
		}
	} else if state == StatePaused {
		if imgui.ButtonV(pauseLabel+"##session", imgui.Vec2{X: pauseW, Y: btnH}) {
			if opts.Controller != nil {
				opts.Controller.Resume()
			}
		}
	} else {
		imgui.BeginDisabled()
		imgui.ButtonV(pauseLabel+"##session", imgui.Vec2{X: pauseW, Y: btnH})
		imgui.EndDisabled()
	}

	imgui.SameLine()

	imgui.PushStyleColorVec4(imgui.ColButton, IndustrialAccent())
	imgui.PushStyleColorVec4(imgui.ColButtonHovered, IndustrialAccent())
	imgui.PushStyleColorVec4(imgui.ColButtonActive, IndustrialAccent())
	if imgui.ButtonV(startLabel+"##session_start", imgui.Vec2{X: startW, Y: btnH}) {
		if opts.Controller == nil {
			imgui.PopStyleColorV(3)
			return
		}
		switch opts.Controller.State() {
		case StateIdle:
			_ = opts.Store.SaveConfig(opts.ConfigPath)
			opts.Controller.Start()
			*open = false
		case StateRunning, StatePaused:
			opts.Controller.Stop()
		}
	}
	imgui.PopStyleColorV(3)
}
