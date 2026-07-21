//go:build android && cgo

package ui

import (
	"fmt"

	"github.com/Dasongzi1366/AutoGo/imgui"
)

// settingsStatus 系统区操作反馈（仅面板本次打开期间有效，不落盘）。
var settingsStatus string

// 面板间距令牌：统一行距/列距，替换散落魔数。
const (
	gapXS = float32(4)
	gapS  = float32(8)
	gapM  = float32(12)
	gapL  = float32(16)
)

// renderCookiePanel 左轨（任务/系统）+ 分类列表 + 详情；任务表见 BuiltinTasks。
func renderCookiePanel(opts ShellOptions) {
	store := opts.Store
	if store == nil {
		return
	}
	SeedPanelDefaults(store)

	avail := imgui.ContentRegionAvail()
	railW := float32(120)
	listW := float32(260)
	if avail.X < 520 {
		listW = avail.X * 0.42
		if listW < 160 {
			listW = 160
		}
	}

	imgui.PushStyleColorVec4(imgui.ColChildBg, QQBlueRailBg())
	imgui.BeginChildStrV("panel_rail", imgui.Vec2{X: railW, Y: 0}, imgui.ChildFlagsBorders, imgui.WindowFlagsNone)
	renderPanelRail(store)
	imgui.EndChild()
	imgui.PopStyleColor()

	imgui.SameLine()

	switch store.GetString(KeyPanelNav) {
	case PanelNavSystem:
		imgui.BeginChildStrV("panel_system", imgui.Vec2{X: 0, Y: 0}, imgui.ChildFlagsBorders, imgui.WindowFlagsNone)
		renderSystemPage(opts)
		imgui.EndChild()
	default:
		imgui.BeginChildStrV("panel_list", imgui.Vec2{X: listW, Y: 0}, imgui.ChildFlagsBorders, imgui.WindowFlagsNone)
		renderTaskList(store)
		imgui.EndChild()

		imgui.SameLine()

		imgui.BeginChildStrV("panel_detail", imgui.Vec2{X: 0, Y: 0}, imgui.ChildFlagsBorders, imgui.WindowFlagsNone)
		renderTaskDetail(store)
		imgui.EndChild()
	}
}

func renderPanelRail(store *Store) {
	nav := store.GetString(KeyPanelNav)
	railButton(store, PanelNavTasks, "任务", nav == PanelNavTasks)
	railButton(store, PanelNavSystem, "系统", nav == PanelNavSystem)
}

func railButton(store *Store, id, label string, active bool) {
	const padX, padY = float32(10), float32(10)
	_, btnH := fitButtonSize(label, padX, padY)
	if btnH < 40 {
		btnH = 40
	}
	imgui.PushStyleVarVec2(imgui.StyleVarFramePadding, imgui.Vec2{X: padX, Y: padY})
	if active {
		// 选中态：天蓝底 + 白字（QQ 风选中项）
		imgui.PushStyleColorVec4(imgui.ColButton, QQBlueAccent())
		imgui.PushStyleColorVec4(imgui.ColButtonHovered, QQBlueAccent())
		imgui.PushStyleColorVec4(imgui.ColButtonActive, QQBlueAccent())
		imgui.PushStyleColorVec4(imgui.ColText, QQBlueWhite())
	}
	if imgui.ButtonV(label+"##rail_"+id, imgui.Vec2{X: -1, Y: btnH}) {
		store.SetString(KeyPanelNav, id)
	}
	if active {
		imgui.PopStyleColorV(4)
	}
	imgui.PopStyleVar()
}

func renderTaskList(store *Store) {
	cat := store.GetString(KeyPanelCat)
	if cat == "" {
		cat = PanelCatAll
	}
	renderCatChips(store, cat)
	imgui.Separator()

	tasks := FilterByCategory(BuiltinTasks(), store.GetString(KeyPanelCat))
	selected := store.GetString(KeyPanelSelected)
	for _, task := range tasks {
		on := task.EnabledKey != "" && store.GetBool(task.EnabledKey)
		// 左侧留空给复选框式启用标记（手绘，不依赖字体字形），标题 5 空格缩进。
		label := "     " + task.Title
		selectedHere := selected == task.ID
		if selectedHere {
			imgui.PushStyleColorVec4(imgui.ColHeader, QQBlueAccent())
			imgui.PushStyleColorVec4(imgui.ColHeaderHovered, QQBlueAccent())
			imgui.PushStyleColorVec4(imgui.ColHeaderActive, QQBlueAccent())
			imgui.PushStyleColorVec4(imgui.ColText, QQBlueWhite())
		}
		textH := measureLabelSize(label).Y
		rowH := textH + 12
		if rowH < 32 {
			rowH = 32
		}
		rowPos := imgui.CursorScreenPos()
		if imgui.SelectableBoolV(label+"##task_"+task.ID, selectedHere, imgui.SelectableFlagsNone, imgui.Vec2{X: 0, Y: rowH}) {
			store.SetString(KeyPanelSelected, task.ID)
		}
		drawListCheck(imgui.WindowDrawList(), rowPos, textH, on)
		if selectedHere {
			imgui.PopStyleColorV(4)
		}
		if sum := taskSummary(store, task); sum != "" {
			// 摘要与标题同一缩进（5 空格 ≈ listTitleX）；TextWrapped 折行后
			// 第二行也从该缩进起，不会孤「·」顶行首。
			imgui.PushStyleColorVec4(imgui.ColText, QQBlueTextDisabled())
			imgui.TextWrapped("     " + sum)
			imgui.PopStyleColor()
		}
	}
	if len(tasks) == 0 {
		imgui.TextDisabled("（该分类暂无任务）")
	}
}

// drawListCheck 列表行启用标记（复选框样式）：白底圆角方框 + 描边，启用时
// 框内天蓝对号，未启用空框；蓝底选中行上白框依然醒目。方框按文字高度垂直
// 居中（Selectable 文字顶部有 FramePadding.y≈4 的偏移）。
func drawListCheck(dl *imgui.DrawList, rowPos imgui.Vec2, textH float32, on bool) {
	const box = float32(16)
	const framePadY = float32(4)
	x := rowPos.X + 8
	y := rowPos.Y + framePadY + (textH-box)/2
	pMin := imgui.Vec2{X: x, Y: y}
	pMax := imgui.Vec2{X: x + box, Y: y + box}
	dl.AddRectFilledV(pMin, pMax, imgui.ColorU32Vec4(QQBlueWhite()), 3, imgui.DrawFlagsRoundCornersAll)
	dl.AddRectV(pMin, pMax, imgui.ColorU32Vec4(HexToVec4("#9cc3e5ff")), 3, imgui.DrawFlagsRoundCornersAll, 1)
	if !on {
		return
	}
	col := imgui.ColorU32Vec4(QQBlueAccent())
	dl.AddLineV(
		imgui.Vec2{X: x + box*0.22, Y: y + box*0.55},
		imgui.Vec2{X: x + box*0.45, Y: y + box*0.78},
		col, 2,
	)
	dl.AddLineV(
		imgui.Vec2{X: x + box*0.45, Y: y + box*0.78},
		imgui.Vec2{X: x + box*0.80, Y: y + box*0.28},
		col, 2,
	)
}

func renderCatChips(store *Store, cat string) {
	chips := []struct{ id, label string }{
		{PanelCatAll, "全部"},
		{PanelCatDaily, "日常"},
		{PanelCatEvent, "活动"},
		{PanelCatMaint, "维护"},
	}
	const padX, padY = float32(8), float32(6) // padX 收窄，保证 4 个 chip 在最小列宽内不换行
	const gap = float32(6)
	imgui.PushStyleVarVec2(imgui.StyleVarFramePadding, imgui.Vec2{X: padX, Y: padY})
	imgui.PushStyleVarVec2(imgui.StyleVarItemSpacing, imgui.Vec2{X: gap, Y: gap})
	defer imgui.PopStyleVarV(2)

	lineStart := true
	for _, c := range chips {
		w, h := fitButtonSize(c.label, padX, padY)
		if !lineStart {
			remain := imgui.ContentRegionAvail().X
			if remain < w+gap {
				lineStart = true
			} else {
				imgui.SameLine()
			}
		}
		active := cat == c.id
		if active {
			imgui.PushStyleColorVec4(imgui.ColButton, QQBlueAccent())
			imgui.PushStyleColorVec4(imgui.ColText, QQBlueWhite())
		}
		if imgui.ButtonV(c.label+"##cat_"+c.id, imgui.Vec2{X: w, Y: h}) {
			store.SetString(KeyPanelCat, c.id)
		}
		if active {
			imgui.PopStyleColorV(2)
		}
		lineStart = false
	}
}

func renderTaskDetail(store *Store) {
	tasks := BuiltinTasks()
	id := store.GetString(KeyPanelSelected)
	task, ok := FindTask(tasks, id)
	if !ok {
		if len(tasks) == 0 {
			imgui.TextDisabled("无任务")
			return
		}
		task = tasks[0]
		store.SetString(KeyPanelSelected, task.ID)
	}

	// 头部：任务名 + 状态胶囊 + 分类（弱色）。本行全是文字/手绘元素，
	// 不用 AlignTextToFramePadding（它只下移文字而 SameLine 回到原行起点，
	// 会把胶囊抬高出 FramePadding.y 的偏差）。
	imgui.Text(task.Title)
	imgui.SameLineV(0, gapS)
	on := task.EnabledKey != "" && store.GetBool(task.EnabledKey)
	drawEnabledPill(on)
	imgui.SameLineV(0, gapS)
	drawPill(categoryLabel(task.Category), QQBlueTextDisabled())
	imgui.Dummy(imgui.Vec2{X: 0, Y: gapXS})
	imgui.Separator()
	imgui.Dummy(imgui.Vec2{X: 0, Y: gapXS})

	switch task.ID {
	case "arena":
		renderArenaDetail(store)
	default:
		imgui.TextDisabled("无详情渲染")
	}
}

// drawEnabledPill 状态胶囊：已启用=天蓝底白字 / 未启用=灰蓝底白字。
func drawEnabledPill(on bool) {
	text, bg := "未启用", QQBlueTextDisabled()
	if on {
		text, bg = "已启用", QQBlueAccent()
	}
	drawPill(text, bg)
}

// drawPill 文字胶囊：圆角底色块 + 白字，画完光标随之前移。
func drawPill(text string, bg imgui.Vec4) {
	sz := measureLabelSize(text)
	const padX, padY = float32(8), float32(1)
	w := sz.X + padX*2
	h := sz.Y + padY*2
	pos := imgui.CursorScreenPos()
	dl := imgui.WindowDrawList()
	dl.AddRectFilledV(
		pos,
		imgui.Vec2{X: pos.X + w, Y: pos.Y + h},
		imgui.ColorU32Vec4(bg), h/2, imgui.DrawFlagsRoundCornersAll,
	)
	dl.AddTextVec2V(
		imgui.Vec2{X: pos.X + padX, Y: pos.Y + padY},
		imgui.ColorU32Vec4(QQBlueWhite()),
		text,
	)
	imgui.Dummy(imgui.Vec2{X: w, Y: h})
}

func renderArenaDetail(store *Store) {
	// 两列表单栅格：标签列固定宽（最长标签 + gapM），控件列从 controlX 起排，
	// 调用 widget 传空 showName，控件左边对齐成一条直线。
	labelW := float32(0)
	for _, s := range []string{"启用", "每日战斗上限", "自动购买次数", "奖杯差阈值"} {
		if w := measureLabelSize(s).X; w > labelW {
			labelW = w
		}
	}
	controlX := imgui.CursorPosX() + labelW + gapM

	formRowLabel("启用", controlX)
	formRowCheckbox(store, KeyArenaEnabled)
	formRowGap()
	formRowLabel("每日战斗上限", controlX)
	UI_创建数字输入框(store, KeyArenaMaxBattles, "", "0=不限", float32(-1), float32(0), "#1f3a52", float64(1), float64(0), float64(999))
	formRowGap()
	formRowLabel("自动购买次数", controlX)
	UI_创建数字输入框(store, KeyArenaAutoBuy, "", "0", float32(-1))
	formRowGap()
	formRowLabel("奖杯差阈值", controlX)
	UI_创建数字输入框(store, KeyArenaTrophyDiff, "", "0", float32(-1))
}

// formRowLabel 画表单行标签并把光标移到控件列起点（同行）。
func formRowLabel(label string, controlX float32) {
	imgui.AlignTextToFramePadding()
	imgui.Text(label)
	imgui.SameLineV(0, 0)
	imgui.SetCursorPosX(controlX)
}

// formRowCheckbox 在控件列起点绘制勾选框，样式与 UI_创建复选框 一致
// （白底、圆角 10、天蓝勾）。不走通用复选框：它空文字后接默认 SameLine，
// 会多吃一个 ItemSpacing.x，落在 controlX+8 破坏栅格。
func formRowCheckbox(store *Store, key string) {
	checked := store.GetBool(key)
	imgui.PushStyleVarVec2(imgui.StyleVarFramePadding, imgui.Vec2{X: 5, Y: 4})
	imgui.PushStyleVarFloat(imgui.StyleVarFrameBorderSize, 1.5)
	imgui.PushStyleVarFloat(imgui.StyleVarFrameRounding, 10)
	imgui.PushStyleColorVec4(imgui.ColFrameBg, HexToVec4("#ffffffff"))
	imgui.PushStyleColorVec4(imgui.ColFrameBgHovered, imgui.Vec4{X: 1, Y: 1, Z: 1, W: 0.22})
	imgui.PushStyleColorVec4(imgui.ColFrameBgActive, imgui.Vec4{X: 0.72, Y: 0.86, Z: 1, W: 0.35})
	imgui.PushStyleColorVec4(imgui.ColCheckMark, QQBlueAccent())
	imgui.PushStyleColorVec4(imgui.ColBorder, HexToVec4("#9cc3e5ff"))
	if imgui.Checkbox("##"+key, &checked) {
		store.SetBool(key, checked)
	}
	imgui.PopStyleColorV(5)
	imgui.PopStyleVarV(3)
}

func formRowGap() { imgui.Dummy(imgui.Vec2{X: 0, Y: gapS}) }

func taskSummary(store *Store, task Task) string {
	switch task.ID {
	case "arena":
		max := int(store.GetFloat(KeyArenaMaxBattles))
		buy := int(store.GetFloat(KeyArenaAutoBuy))
		diff := int(store.GetFloat(KeyArenaTrophyDiff))
		maxLabel := "不限"
		if max > 0 {
			maxLabel = fmt.Sprintf("%d", max)
		}
		return fmt.Sprintf("上限 %s · 购买 %d · 奖杯差 %d", maxLabel, buy, diff)
	default:
		return ""
	}
}

func renderSystemPage(opts ShellOptions) {
	store := opts.Store
	imgui.Text("系统")
	imgui.Dummy(imgui.Vec2{X: 0, Y: gapXS})
	imgui.TextDisabled("配置持久化")
	imgui.Separator()
	imgui.Dummy(imgui.Vec2{X: 0, Y: gapXS})

	path := opts.ConfigPath
	if path == "" {
		path = DefaultConfigPath
	}
	imgui.TextDisabled("配置文件  " + path)
	imgui.Dummy(imgui.Vec2{X: 0, Y: gapS})

	UI_创建按钮("save_config", "保存配置", func() {
		if err := store.SaveConfig(path); err != nil {
			settingsStatus = fmt.Sprintf("保存失败：%v", err)
			return
		}
		settingsStatus = "配置已保存"
	}, float32(-2), float32(-2))
	imgui.SameLine()
	UI_创建按钮("clear_cache", "清除缓存", func() {
		if opts.Controller != nil {
			st := opts.Controller.State()
			if st == StateRunning || st == StatePaused {
				opts.Controller.Stop()
			}
		}
		if err := ClearPanelCache(store, path, opts.DataStorePath, opts.Reseed); err != nil {
			settingsStatus = fmt.Sprintf("清除失败：%v", err)
			return
		}
		SeedPanelDefaults(store)
		settingsStatus = "缓存已清除，默认配置已恢复"
	}, float32(-2), float32(-2))

	if settingsStatus != "" {
		imgui.Spacing()
		imgui.TextWrapped(settingsStatus)
	}
}
