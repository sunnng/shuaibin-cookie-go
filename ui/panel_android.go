//go:build android && cgo

package ui

import (
	"fmt"

	"github.com/Dasongzi1366/AutoGo/device"
	"github.com/Dasongzi1366/AutoGo/imgui"
)

// drawPanel QQ 风窗壳：蓝色渐变标题栏 + 自适应宽高 + 脚本条 + 配置内容。
//
// panelTitleH 标题栏高度；shell.Minimized() 为 true 时面板缩成只剩标题栏一条
// （面板仍算打开，脚本保持自动暂停），再点缩小钮或重开面板恢复。
const panelTitleH = float32(38)

const gapL = float32(16)

func drawPanel(ctx *Ctx, shell *Shell) {
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

	panelH := winH
	if shell.Minimized() {
		panelH = panelTitleH + 8
	}
	imgui.SetNextWindowSizeV(imgui.Vec2{X: winW, Y: panelH}, imgui.CondAlways)
	imgui.SetNextWindowPosV(imgui.Vec2{X: x, Y: y}, imgui.CondOnce, imgui.Vec2{X: 0, Y: 0})

	flags := imgui.WindowFlagsNoMove | imgui.WindowFlagsNoScrollbar |
		imgui.WindowFlagsNoResize | imgui.WindowFlagsNoTitleBar

	open := true
	if imgui.BeginV(shell.Title(), &open, flags) {
		drawPanelTitleBar(ctx, shell)
		if !shell.Minimized() {
			renderScriptBar(ctx, shell)
			imgui.Separator()
			renderPanelContent(ctx, shell)
		}
	}
	imgui.End()

	if !open {
		shell.ClosePanel()
	}
}

// drawPanelTitleBar QQ 风标题栏：天蓝竖向渐变底 + 左侧白色标题 + 右侧窗控按钮。
// 窗控按钮按 QQ2007 风格成对出现：缩小（钢蓝方块 + 白“—”）+ 关闭（红方块 + 白“×”）。
func drawPanelTitleBar(ctx *Ctx, shell *Shell) {
	// 记下 ImGui 默认内容原点（= border+padding，相对窗口左上角）。
	// SetCursorPos 的坐标是相对窗口左上角的，若末尾把 X 设为 0 会把
	// 后续内容压到窗口绝对边缘（看起来"紧贴边框"）。
	contentX := imgui.CursorPosX()
	winPos := imgui.WindowPos()
	winSize := imgui.WindowSize()
	dl := imgui.WindowDrawList()

	pMax := imgui.Vec2{X: winPos.X + winSize.X, Y: winPos.Y + panelTitleH}
	th := ctx.theme()
	top := imgui.ColorU32Vec4(toVec4(th.TitleTop))
	bottom := imgui.ColorU32Vec4(toVec4(th.TitleBottom))
	dl.AddRectFilledMultiColor(winPos, pMax, top, top, bottom, bottom)

	white := imgui.ColorU32Vec4(toVec4(th.White))
	textSz := measureLabelSize(shell.Title())
	dl.AddTextVec2V(
		imgui.Vec2{X: winPos.X + 12, Y: winPos.Y + (panelTitleH-textSz.Y)/2},
		white,
		shell.Title(),
	)

	// 窗控按钮（右对齐成对小方块）：缩小在左，红色关闭在右。
	const btnW, btnH = float32(24), float32(18)
	const gap, margin = float32(4), float32(8)
	btnY := winPos.Y + (panelTitleH-btnH)/2
	closeX := winPos.X + winSize.X - margin - btnW
	minX := closeX - gap - btnW

	// 缩小钮：钢蓝方块 + 白“—”，点击在「整条标题栏 / 完整面板」间切换。
	imgui.SetCursorPos(imgui.Vec2{X: minX - winPos.X, Y: btnY - winPos.Y})
	if imgui.InvisibleButton("##panel_min", imgui.Vec2{X: btnW, Y: btnH}) {
		shell.ToggleMinimized()
	}
	minFill := imgui.Vec4{X: 0.36, Y: 0.52, Z: 0.68, W: 1}
	if imgui.IsItemHovered() {
		minFill = imgui.Vec4{X: 0.47, Y: 0.63, Z: 0.79, W: 1}
	}
	dl.AddRectFilledV(
		imgui.Vec2{X: minX, Y: btnY},
		imgui.Vec2{X: minX + btnW, Y: btnY + btnH},
		imgui.ColorU32Vec4(minFill), 3, imgui.DrawFlagsRoundCornersAll,
	)
	midY := btnY + btnH/2
	dl.AddLineV(imgui.Vec2{X: minX + 6, Y: midY}, imgui.Vec2{X: minX + btnW - 6, Y: midY}, white, 2)

	// 关闭钮：红色方块 + 白“×”，关闭时顺手复位缩小态，下次打开是完整面板。
	imgui.SetCursorPos(imgui.Vec2{X: closeX - winPos.X, Y: btnY - winPos.Y})
	if imgui.InvisibleButton("##panel_close", imgui.Vec2{X: btnW, Y: btnH}) {
		shell.ClosePanel()
		if shell.Minimized() {
			shell.ToggleMinimized()
		}
	}
	closeFill := imgui.Vec4{X: 0.85, Y: 0.16, Z: 0.22, W: 1}
	if imgui.IsItemHovered() {
		closeFill = imgui.Vec4{X: 0.95, Y: 0.29, Z: 0.33, W: 1}
	}
	dl.AddRectFilledV(
		imgui.Vec2{X: closeX, Y: btnY},
		imgui.Vec2{X: closeX + btnW, Y: btnY + btnH},
		imgui.ColorU32Vec4(closeFill), 3, imgui.DrawFlagsRoundCornersAll,
	)
	off := float32(4.5)
	cx, cy := closeX+btnW/2, btnY+btnH/2
	dl.AddLineV(imgui.Vec2{X: cx - off, Y: cy - off}, imgui.Vec2{X: cx + off, Y: cy + off}, white, 1.8)
	dl.AddLineV(imgui.Vec2{X: cx + off, Y: cy - off}, imgui.Vec2{X: cx - off, Y: cy + off}, white, 1.8)

	// 内容从标题栏下方开始；X 恢复内容原点，避免贴窗口边缘。
	// 缩小态窗口只有标题栏高，下移光标会越出 WorkRect 触发 imgui 边界告警，且
	// 缩小态本就没有后续内容，跳过。
	if !shell.Minimized() {
		imgui.SetCursorPos(imgui.Vec2{X: contentX, Y: panelTitleH + 4})
	}
}

func renderScriptBar(ctx *Ctx, shell *Shell) {
	state := shell.ScriptState()

	en, total := CountEnabled(shell.Store(), shell.Tasks())

	// 内容原点（border+padding）即起始位置，与下方左轨/列表自然对齐
	imgui.AlignTextToFramePadding()
	imgui.Text("状态：")
	imgui.SameLine()
	imgui.TextColored(panelStateColor(state), panelStateLabel(state))
	imgui.SameLineV(0, gapL)
	imgui.TextDisabled(fmt.Sprintf("已启用任务 %d/%d", en, total))
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

	if state == StateRunning || state == StatePaused {
		if imgui.ButtonV(pauseLabel+"##script", imgui.Vec2{X: pauseW, Y: btnH}) {
			shell.PauseResume()
		}
	} else {
		imgui.BeginDisabled()
		imgui.ButtonV(pauseLabel+"##script", imgui.Vec2{X: pauseW, Y: btnH})
		imgui.EndDisabled()
	}

	imgui.SameLine()

	// 主按钮：天蓝底 + 白字（QQ 风主操作）
	th := ctx.theme()
	imgui.PushStyleColorVec4(imgui.ColButton, toVec4(th.Accent))
	imgui.PushStyleColorVec4(imgui.ColButtonHovered, toVec4(th.Accent))
	imgui.PushStyleColorVec4(imgui.ColButtonActive, toVec4(th.Accent))
	imgui.PushStyleColorVec4(imgui.ColText, toVec4(th.White))
	if imgui.ButtonV(startLabel+"##script_start", imgui.Vec2{X: startW, Y: btnH}) {
		wasIdle := shell.ScriptState() == StateIdle
		if err := shell.StartStop(); err == nil && wasIdle {
			shell.ClosePanel()
		}
	}
	imgui.PopStyleColorV(4)
}

// renderPanelContent 左轨导航 + 当前条目内容。导航条目来自 shell.Nav()；
// 任务列表页/系统页是应用挂载的普通条目（框架提供 TaskListPage/SystemPage）。
func renderPanelContent(ctx *Ctx, shell *Shell) {
	store := shell.Store()
	nav := shell.Nav()
	if len(nav) == 0 {
		return
	}
	th := ctx.theme()

	avail := imgui.ContentRegionAvail()
	railW := float32(ctx.S(120))

	imgui.PushStyleColorVec4(imgui.ColChildBg, toVec4(th.RailBg))
	imgui.BeginChildStrV("panel_rail", imgui.Vec2{X: railW, Y: 0}, imgui.ChildFlagsBorders, imgui.WindowFlagsNone)
	current := store.GetString(KeyPanelNav)
	for _, entry := range nav {
		railButton(ctx, store, entry.ID, entry.Title, current == entry.ID)
	}
	imgui.EndChild()
	imgui.PopStyleColor()

	imgui.SameLine()
	imgui.BeginChildStrV("panel_body", imgui.Vec2{X: 0, Y: 0}, imgui.ChildFlagsBorders, imgui.WindowFlagsNone)
	for _, entry := range nav {
		if entry.ID == current && entry.Render != nil {
			ctx.Push("nav:" + entry.ID)
			entry.Render(ctx)
			ctx.Pop()
		}
	}
	imgui.EndChild()
}

func railButton(ctx *Ctx, store *Store, id, label string, active bool) {
	const padX, padY = float32(10), float32(10)
	_, btnH := fitButtonSize(label, padX, padY)
	if btnH < 40 {
		btnH = 40
	}
	th := ctx.theme()
	imgui.PushStyleVarVec2(imgui.StyleVarFramePadding, imgui.Vec2{X: padX, Y: padY})
	if active {
		// 选中态：天蓝底 + 白字（QQ 风选中项）
		imgui.PushStyleColorVec4(imgui.ColButton, toVec4(th.Accent))
		imgui.PushStyleColorVec4(imgui.ColButtonHovered, toVec4(th.Accent))
		imgui.PushStyleColorVec4(imgui.ColButtonActive, toVec4(th.Accent))
		imgui.PushStyleColorVec4(imgui.ColText, toVec4(th.White))
	}
	if imgui.ButtonV(label+"##rail_"+id, imgui.Vec2{X: -1, Y: btnH}) {
		store.SetString(KeyPanelNav, id)
	}
	if active {
		imgui.PopStyleColorV(4)
	}
	imgui.PopStyleVar()
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
