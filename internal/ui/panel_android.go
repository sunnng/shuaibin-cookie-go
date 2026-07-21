//go:build android && cgo

package ui

import (
	"fmt"

	"github.com/Dasongzi1366/AutoGo/device"
	"github.com/Dasongzi1366/AutoGo/imgui"
)

// drawConfigPanel QQ 风窗壳：蓝色渐变标题栏 + 自适应宽高 + 脚本条 + 配置内容。
func drawConfigPanel(opts ShellOptions, open *bool) bool {
	width, height, _, _ := device.GetDisplayInfo(0)
	winW := float32(width) * 0.72
	if winW > 980 {
		winW = 980
	}
	if winW < 640 {
		winW = 640
	}
	if winW > float32(width) { // 屏幕本身不足 640 宽时不得超出屏幕
		winW = float32(width)
	}
	winH := float32(height) * 0.78
	if winH > 700 {
		winH = 700
	}
	if winH < 480 {
		winH = 480
	}
	if winH > float32(height) {
		winH = float32(height)
	}
	x := (float32(width) - winW) / 2
	y := 12 + (float32(height)-12-winH)/2

	imgui.SetNextWindowSizeV(imgui.Vec2{X: winW, Y: winH}, imgui.CondOnce)
	imgui.SetNextWindowPosV(imgui.Vec2{X: x, Y: y}, imgui.CondOnce, imgui.Vec2{X: 0, Y: 0})

	flags := imgui.WindowFlagsNoMove | imgui.WindowFlagsNoScrollbar |
		imgui.WindowFlagsNoResize | imgui.WindowFlagsNoTitleBar
	if imgui.BeginV(opts.Title, open, flags) {
		drawPanelTitleBar(opts.Title, open)
		renderScriptBar(opts, open)
		imgui.Separator()
		renderCookiePanel(opts)
	}
	imgui.End()

	return *open
}

// drawPanelTitleBar QQ 风标题栏：天蓝竖向渐变底 + 左侧白色标题 + 右侧关闭圆钮。
func drawPanelTitleBar(title string, open *bool) {
	const titleH = float32(38)
	// 记下 ImGui 默认内容原点（= border+padding，相对窗口左上角）。
	// SetCursorPos 的坐标是相对窗口左上角的，若末尾把 X 设为 0 会把
	// 后续内容压到窗口绝对边缘（看起来"紧贴边框"）。
	contentX := imgui.CursorPosX()
	winPos := imgui.WindowPos()
	winSize := imgui.WindowSize()
	dl := imgui.WindowDrawList()

	pMax := imgui.Vec2{X: winPos.X + winSize.X, Y: winPos.Y + titleH}
	top := imgui.ColorU32Vec4(QQBlueTitleTop())
	bottom := imgui.ColorU32Vec4(QQBlueTitleBottom())
	dl.AddRectFilledMultiColor(winPos, pMax, top, top, bottom, bottom)

	white := imgui.ColorU32Vec4(QQBlueWhite())
	textSz := measureLabelSize(title)
	dl.AddTextVec2V(
		imgui.Vec2{X: winPos.X + 12, Y: winPos.Y + (titleH-textSz.Y)/2},
		white,
		title,
	)

	// 关闭圆钮（白描边 + X）
	closeR := float32(11)
	cx := pMax.X - titleH/2
	cy := winPos.Y + titleH/2
	dl.AddCircleV(imgui.Vec2{X: cx, Y: cy}, closeR, white, 0, 1.5)
	off := closeR * 0.45
	dl.AddLineV(
		imgui.Vec2{X: cx - off, Y: cy - off},
		imgui.Vec2{X: cx + off, Y: cy + off},
		white, 1.5,
	)
	dl.AddLineV(
		imgui.Vec2{X: cx + off, Y: cy - off},
		imgui.Vec2{X: cx - off, Y: cy + off},
		white, 1.5,
	)

	imgui.SetCursorPos(imgui.Vec2{X: winSize.X - titleH, Y: 0})
	if imgui.InvisibleButton("##panel_close", imgui.Vec2{X: titleH, Y: titleH}) {
		*open = false
	}

	// 内容从标题栏下方开始；X 恢复内容原点，避免贴窗口边缘
	imgui.SetCursorPos(imgui.Vec2{X: contentX, Y: titleH + 4})
}

func renderScriptBar(opts ShellOptions, open *bool) {
	state := StateIdle
	if opts.Controller != nil {
		state = opts.Controller.State()
	}

	en, total := 0, 0
	if opts.Store != nil {
		en, total = CountEnabled(opts.Store, BuiltinTasks())
	}

	// 内容原点（border+padding）即起始位置，与下方左轨/列表自然对齐
	imgui.AlignTextToFramePadding()
	imgui.Text("状态：")
	imgui.SameLine()
	imgui.TextColored(panelStateColor(state), panelStateLabel(state))
	imgui.SameLine()
	imgui.TextDisabled(fmt.Sprintf("　已启用任务 %d/%d", en, total))
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
		if imgui.ButtonV(pauseLabel+"##script", imgui.Vec2{X: pauseW, Y: btnH}) {
			if opts.Controller != nil {
				opts.Controller.Pause()
			}
		}
	} else if state == StatePaused {
		if imgui.ButtonV(pauseLabel+"##script", imgui.Vec2{X: pauseW, Y: btnH}) {
			if opts.Controller != nil {
				opts.Controller.Resume()
			}
		}
	} else {
		imgui.BeginDisabled()
		imgui.ButtonV(pauseLabel+"##script", imgui.Vec2{X: pauseW, Y: btnH})
		imgui.EndDisabled()
	}

	imgui.SameLine()

	// 主按钮：天蓝底 + 白字（QQ 风主操作）
	imgui.PushStyleColorVec4(imgui.ColButton, QQBlueAccent())
	imgui.PushStyleColorVec4(imgui.ColButtonHovered, QQBlueAccent())
	imgui.PushStyleColorVec4(imgui.ColButtonActive, QQBlueAccent())
	imgui.PushStyleColorVec4(imgui.ColText, QQBlueWhite())
	if imgui.ButtonV(startLabel+"##script_start", imgui.Vec2{X: startW, Y: btnH}) {
		if opts.Controller == nil {
			imgui.PopStyleColorV(4)
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
	imgui.PopStyleColorV(4)
}

// panelStateLabel 脚本状态的通俗叫法。
func panelStateLabel(state ScriptState) string {
	switch state {
	case StateRunning:
		return "运行中"
	case StatePaused:
		return "已暂停"
	default:
		return "未运行"
	}
}

// panelStateColor 状态文字颜色。浅底面板不能用灵动岛那套亮色，
// 这里用深色变体保证对比度（绿=运行 / 橙=暂停 / 灰蓝=未运行）。
func panelStateColor(state ScriptState) imgui.Vec4 {
	switch state {
	case StateRunning:
		return imgui.Vec4{X: 0.17, Y: 0.63, Z: 0.17, W: 1}
	case StatePaused:
		return imgui.Vec4{X: 0.85, Y: 0.55, Z: 0.1, W: 1}
	default:
		return imgui.Vec4{X: 0.35, Y: 0.45, Z: 0.6, W: 1}
	}
}
