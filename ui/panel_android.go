//go:build android && cgo

package ui

import (
	"fmt"

	"github.com/Dasongzi1366/AutoGo/device"
	"github.com/Dasongzi1366/AutoGo/imgui"
)

// drawPanel 糖果积木窗壳（docs/ui-redesign/design-system.md §4.2）：纸面底 +
// 3px 墨描边 + radius 18 + shadow-lg 硬阴影；黄底积木块标题 + 描边窗控方块 +
// 脚本条（状态摘要 + 主操作）+ 左轨导航与配置内容。
//
// shell.Minimized() 为 true 时面板缩成只剩标题栏一条（面板仍算打开，脚本保持
// 自动暂停），再点缩小钮或重开面板恢复。
func drawPanel(ctx *Ctx, shell *Shell) {
	width, height, _, _ := device.GetDisplayInfo(0)
	winW := float32(width) * 0.72
	if max := float32(ctx.S(1040)); winW > max {
		winW = max
	}
	if min := float32(ctx.S(640)); winW < min {
		winW = min
	}
	if winW > float32(width) { // 屏幕本身不足时不得超出屏幕
		winW = float32(width)
	}
	winH := float32(height) * 0.78
	if max := float32(ctx.S(640)); winH > max {
		winH = max
	}
	if min := float32(ctx.S(480)); winH < min {
		winH = min
	}
	if winH > float32(height) {
		winH = float32(height)
	}
	x := (float32(width) - winW) / 2
	y := 12 + (float32(height)-12-winH)/2

	titleH := float32(ctx.S(64))
	panelH := winH
	if shell.Minimized() {
		panelH = titleH + 8
	}

	// shadow-lg：8px 偏移的半透明墨块（避免在游戏画面上糊成大黑块）。
	radius := float32(ctx.S(18))
	imgui.BackgroundDrawList().AddRectFilledV(
		imgui.Vec2{X: x + 8, Y: y + 8},
		imgui.Vec2{X: x + winW + 8, Y: y + panelH + 8},
		imgui.ColorU32Vec4(imgui.Vec4{X: 0.090, Y: 0.075, Z: 0.047, W: 0.35}),
		radius, imgui.DrawFlagsRoundCornersAll,
	)

	imgui.SetNextWindowSizeV(imgui.Vec2{X: winW, Y: panelH}, imgui.CondAlways)
	imgui.SetNextWindowPosV(imgui.Vec2{X: x, Y: y}, imgui.CondOnce, imgui.Vec2{X: 0, Y: 0})

	imgui.PushStyleColorVec4(imgui.ColBorder, candyInk)
	imgui.PushStyleVarFloat(imgui.StyleVarWindowBorderSize, 3)
	imgui.PushStyleVarFloat(imgui.StyleVarWindowRounding, radius)

	flags := imgui.WindowFlagsNoMove | imgui.WindowFlagsNoScrollbar |
		imgui.WindowFlagsNoResize | imgui.WindowFlagsNoTitleBar

	open := true
	if imgui.BeginV(shell.Title(), &open, flags) {
		contentX := imgui.CursorPosX()
		// 先用整宽 Dummy 登记标题栏+脚本条区域的边界（imgui 不允许用
		// SetCursorPos 扩展父窗口边界，须先提交 item 撑开），栏内的绝对
		// 定位都在已登记范围内。光标从 (0,0) 起登记，覆盖整栏。
		imgui.SetCursorPos(imgui.Vec2{X: 0, Y: 0})
		barH := float32(ctx.S(72))
		reserveH := titleH + 8
		if !shell.Minimized() {
			reserveH = titleH + barH + float32(ctx.S(8))
		}
		imgui.Dummy(imgui.Vec2{X: winW, Y: reserveH})
		drawPanelTitleBar(ctx, shell, titleH)
		if !shell.Minimized() {
			renderScriptBar(ctx, shell, titleH, barH)
			imgui.SetCursorPos(imgui.Vec2{X: contentX, Y: reserveH})
			renderPanelContent(ctx, shell)
		}
	}
	imgui.End()

	imgui.PopStyleVarV(2)
	imgui.PopStyleColor()

	if !open {
		shell.ClosePanel()
	}
}

// drawPanelTitleBar 标题栏（64px）：黄底墨字积木块标题 + 右侧状态胶囊 +
// 两个描边窗控方块（— 缩小 / ✕ 关闭），栏底 3px 墨线分隔。
func drawPanelTitleBar(ctx *Ctx, shell *Shell, titleH float32) {
	winPos := imgui.WindowPos()
	winSize := imgui.WindowSize()
	dl := imgui.WindowDrawList()

	margin := float32(ctx.S(16))
	gap := float32(ctx.S(12))

	// 黄底积木块标题：3px 墨描边 + 硬阴影 + 墨字。
	title := shell.Title()
	textSz := measureLabelSize(title)
	blockPadX, blockPadY := float32(ctx.S(16)), float32(ctx.S(8))
	blockW := textSz.X + blockPadX*2
	blockH := textSz.Y + blockPadY*2
	blockMin := imgui.Vec2{X: winPos.X + margin, Y: winPos.Y + (titleH-blockH)/2}
	drawBlock(dl, blockMin, imgui.Vec2{X: blockMin.X + blockW, Y: blockMin.Y + blockH}, candyYellow, float32(ctx.S(12)), 3, 4)
	dl.AddTextVec2V(
		imgui.Vec2{X: blockMin.X + blockPadX, Y: blockMin.Y + blockPadY},
		imgui.ColorU32Vec4(candyInk),
		title,
	)

	// 窗控方块（右对齐）：缩小在左（白底墨“—”），关闭在右（红底纸“×”）。
	btnW, btnH := float32(ctx.S(48)), float32(ctx.S(42))
	btnY := winPos.Y + (titleH-btnH)/2
	closeX := winPos.X + winSize.X - margin - btnW
	minX := closeX - gap - btnW

	if candyButton("##panel_min", imgui.Vec2{X: minX, Y: btnY}, imgui.Vec2{X: btnW, Y: btnH}, candyRaised, float32(ctx.S(10)), false,
		func(dl *imgui.DrawList, pMin, pMax imgui.Vec2, pressed bool) {
			midY := (pMin.Y + pMax.Y) / 2
			dl.AddLineV(
				imgui.Vec2{X: pMin.X + btnW*0.28, Y: midY},
				imgui.Vec2{X: pMax.X - btnW*0.28, Y: midY},
				imgui.ColorU32Vec4(candyInk), 3,
			)
		}) {
		shell.ToggleMinimized()
	}

	if candyButton("##panel_close", imgui.Vec2{X: closeX, Y: btnY}, imgui.Vec2{X: btnW, Y: btnH}, candyRed, float32(ctx.S(10)), false,
		func(dl *imgui.DrawList, pMin, pMax imgui.Vec2, pressed bool) {
			cx, cy := (pMin.X+pMax.X)/2, (pMin.Y+pMax.Y)/2
			off := float32(ctx.S(6))
			col := imgui.ColorU32Vec4(candyPaper)
			dl.AddLineV(imgui.Vec2{X: cx - off, Y: cy - off}, imgui.Vec2{X: cx + off, Y: cy + off}, col, 3)
			dl.AddLineV(imgui.Vec2{X: cx + off, Y: cy - off}, imgui.Vec2{X: cx - off, Y: cy + off}, col, 3)
		}) {
		shell.ClosePanel()
		if shell.Minimized() {
			shell.ToggleMinimized()
		}
	}

	// 状态胶囊：白底 2px 描边 + 状态点 + 状态文字，位于窗控方块左侧。
	state := shell.ScriptState()
	stateLabel := islandStateLabel(state)
	pillTextSz := measureLabelSize(stateLabel)
	dotD := float32(ctx.S(10))
	pillPadX, pillPadY := float32(ctx.S(14)), float32(ctx.S(6))
	pillGap := float32(ctx.S(8))
	pillW := pillPadX*2 + dotD + pillGap + pillTextSz.X
	pillH := pillTextSz.Y + pillPadY*2
	pillMin := imgui.Vec2{X: minX - gap - pillW, Y: winPos.Y + (titleH-pillH)/2}
	drawBlock(dl, pillMin, imgui.Vec2{X: pillMin.X + pillW, Y: pillMin.Y + pillH}, candyRaised, pillH/2, 2, 0)
	dotC := imgui.ColorU32Vec4(candyStateColor(state))
	dl.AddCircleFilled(
		imgui.Vec2{X: pillMin.X + pillPadX + dotD/2, Y: pillMin.Y + pillH/2},
		dotD/2, dotC,
	)
	dl.AddTextVec2V(
		imgui.Vec2{X: pillMin.X + pillPadX + dotD + pillGap, Y: pillMin.Y + pillPadY},
		imgui.ColorU32Vec4(candyInk),
		stateLabel,
	)

	// 栏底 3px 墨线。
	lineY := winPos.Y + titleH - 1.5
	dl.AddLineV(
		imgui.Vec2{X: winPos.X, Y: lineY},
		imgui.Vec2{X: winPos.X + winSize.X, Y: lineY},
		imgui.ColorU32Vec4(candyInk), 3,
	)
}

// renderScriptBar 脚本条（72px）：左侧状态摘要（状态点 + 状态文字 + 已启用计数，
// 自动暂停时追加橙色「已自动暂停」标记），右侧主操作（暂停/继续 次要白底 +
// 开始绿底/停止红底）。栏底 3px 墨线。
func renderScriptBar(ctx *Ctx, shell *Shell, titleH, barH float32) {
	state := shell.ScriptState()
	winPos := imgui.WindowPos()
	winSize := imgui.WindowSize()
	dl := imgui.WindowDrawList()

	barTop := winPos.Y + titleH
	margin := float32(ctx.S(16))

	// 条底 #FBEFD8 + 栏底墨线。
	dl.AddRectFilledV(
		imgui.Vec2{X: winPos.X, Y: barTop},
		imgui.Vec2{X: winPos.X + winSize.X, Y: barTop + barH},
		imgui.ColorU32Vec4(imgui.Vec4{X: 0xFB / 255.0, Y: 0xEF / 255.0, Z: 0xD8 / 255.0, W: 1}),
		0, imgui.DrawFlagsRoundCornersNone,
	)
	lineY := barTop + barH - 1.5
	dl.AddLineV(
		imgui.Vec2{X: winPos.X, Y: lineY},
		imgui.Vec2{X: winPos.X + winSize.X, Y: lineY},
		imgui.ColorU32Vec4(candyInk), 3,
	)

	// 左侧状态摘要。
	midY := barTop + barH/2
	curX := winPos.X + margin
	dotD := float32(ctx.S(12))
	dl.AddCircleFilled(imgui.Vec2{X: curX + dotD/2, Y: midY}, dotD/2, imgui.ColorU32Vec4(candyStateColor(state)))
	curX += dotD + float32(ctx.S(10))

	stateLabel := panelStateLabel(state)
	stateSz := measureLabelSize(stateLabel)
	dl.AddTextVec2V(imgui.Vec2{X: curX, Y: midY - stateSz.Y/2}, imgui.ColorU32Vec4(panelStateColor(state)), stateLabel)
	curX += stateSz.X + float32(ctx.S(12))

	en, total := CountEnabled(shell.Store(), shell.Tasks())
	enabledText := fmt.Sprintf("已启用任务 %d/%d", en, total)
	enSz := measureLabelSize(enabledText)
	dl.AddTextVec2V(imgui.Vec2{X: curX, Y: midY - enSz.Y/2}, imgui.ColorU32Vec4(candySec), enabledText)
	curX += enSz.X + float32(ctx.S(10))

	// 自动暂停标记：橙色胶囊，让"为什么脚本停了"可见。
	if shell.AutoPaused() {
		badge := "已自动暂停"
		badgeSz := measureLabelSize(badge)
		padX, padY := float32(ctx.S(10)), float32(ctx.S(3))
		bMin := imgui.Vec2{X: curX, Y: midY - (badgeSz.Y+padY*2)/2}
		drawBlock(dl, bMin, imgui.Vec2{X: bMin.X + badgeSz.X + padX*2, Y: bMin.Y + badgeSz.Y + padY*2}, candyOrange, (badgeSz.Y+padY*2)/2, 2, 0)
		dl.AddTextVec2V(
			imgui.Vec2{X: bMin.X + padX, Y: bMin.Y + padY},
			imgui.ColorU32Vec4(candyInk),
			badge,
		)
	}

	// 右侧主操作：暂停/继续（次要 120×48）+ 开始/停止（主操作 140×48）。
	btnH := float32(ctx.S(48))
	pauseW, startW := float32(ctx.S(120)), float32(ctx.S(140))
	gap := float32(ctx.S(12))
	btnY := barTop + (barH-btnH)/2
	startX := winPos.X + winSize.X - margin - startW
	pauseX := startX - gap - pauseW

	pauseLabel := "暂停"
	if state == StatePaused {
		pauseLabel = "继续"
	}
	pauseDisabled := state == StateIdle
	if candyButton(pauseLabel+"##script", imgui.Vec2{X: pauseX, Y: btnY}, imgui.Vec2{X: pauseW, Y: btnH}, candyRaised, float32(ctx.S(12)), pauseDisabled,
		func(dl *imgui.DrawList, pMin, pMax imgui.Vec2, pressed bool) {
			col := candyInk
			if pauseDisabled {
				col.W *= 0.4
			}
			candyLabelInRect(dl, pMin, pMax, pauseLabel, col)
		}) {
		shell.PauseResume()
	}

	startLabel := "开始"
	startFill := candyGreen
	startText := candyInk
	if state == StateRunning || state == StatePaused {
		startLabel = "停止"
		startFill = candyRed
		startText = candyPaper
	}
	if candyButton(startLabel+"##script_start", imgui.Vec2{X: startX, Y: btnY}, imgui.Vec2{X: startW, Y: btnH}, startFill, float32(ctx.S(12)), false,
		func(dl *imgui.DrawList, pMin, pMax imgui.Vec2, pressed bool) {
			candyLabelInRect(dl, pMin, pMax, startLabel, startText)
		}) {
		wasIdle := shell.ScriptState() == StateIdle
		if err := shell.StartStop(); err != nil {
			LogErrorf("panel start/stop: %v", err)
		} else if wasIdle {
			shell.ClosePanel()
		}
	}
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

	railW := float32(ctx.S(136))

	imgui.PushStyleColorVec4(imgui.ColChildBg, toVec4(th.RailBg))
	imgui.PushStyleVarFloat(imgui.StyleVarChildBorderSize, 3)
	imgui.BeginChildStrV("panel_rail", imgui.Vec2{X: railW, Y: 0}, imgui.ChildFlagsBorders, imgui.WindowFlagsNone)
	current := store.GetString(KeyPanelNav)
	for _, entry := range nav {
		railButton(ctx, store, entry.ID, entry.Title, current == entry.ID)
	}
	imgui.EndChild()
	imgui.PopStyleVar()
	imgui.PopStyleColor()

	imgui.SameLine()
	imgui.BeginChildStrV("panel_body", imgui.Vec2{X: 0, Y: 0}, imgui.ChildFlagsNone, imgui.WindowFlagsNone)
	// 持久化的 nav id 可能已失效（应用改版后不再挂载）：回退到首个条目，
	// 与旧 renderCookiePanel 的 default 分支一致；不回写 store，点击左轨自愈。
	active := nav[0]
	for _, entry := range nav {
		if entry.ID == current {
			active = entry
			break
		}
	}
	if active.Render != nil {
		ctx.Push("nav:" + active.ID)
		active.Render(ctx)
		ctx.Pop()
	}
	imgui.EndChild()
}

// railButton 导航轨条目（48px 高积木块）：选中 = 黄底 + 3px 墨描边 + 硬阴影；
// 未选 = 无底色无边框，hover 半透明白。
func railButton(ctx *Ctx, store *Store, id, label string, active bool) {
	btnH := float32(ctx.S(48))
	radius := float32(ctx.S(12))

	avail := imgui.ContentRegionAvail()
	pos := imgui.CursorScreenPos()
	size := imgui.Vec2{X: avail.X, Y: btnH}

	clicked := imgui.InvisibleButton("##rail_"+id, size)
	hovered := imgui.IsItemHovered()
	pressed := imgui.IsItemActive()

	dl := imgui.WindowDrawList()
	off := float32(0)
	shadow := float32(4)
	if pressed {
		off = 3
		shadow = 1
	}
	pMin := imgui.Vec2{X: pos.X + off, Y: pos.Y + off}
	pMax := imgui.Vec2{X: pos.X + size.X + off, Y: pos.Y + size.Y + off}
	switch {
	case active:
		drawBlock(dl, pMin, pMax, candyYellow, radius, 3, shadow)
	case hovered:
		dl.AddRectFilledV(pMin, pMax, imgui.ColorU32Vec4(imgui.Vec4{X: 1, Y: 1, Z: 1, W: 0.5}), radius, imgui.DrawFlagsRoundCornersAll)
	}

	textSz := measureLabelSize(label)
	dl.AddTextVec2V(
		imgui.Vec2{X: pMin.X + float32(ctx.S(14)), Y: pMin.Y + (size.Y-textSz.Y)/2},
		imgui.ColorU32Vec4(candyInk),
		label,
	)

	if clicked {
		store.SetString(KeyPanelNav, id)
	}
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

// panelStateColor 状态文字颜色：纸面上的深变体（≥4.5:1），
// 绿=运行 / 橙=暂停 / 蓝=未运行（design-system.md §3.1）。
func panelStateColor(state ScriptState) imgui.Vec4 {
	switch state {
	case StateRunning:
		return candyRunText
	case StatePaused:
		return candyPauseText
	default:
		return candyIdleText
	}
}
