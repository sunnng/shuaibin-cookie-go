//go:build android && cgo

package ui

import "github.com/Dasongzi1366/AutoGo/imgui"

// TaskListPage 框架可复用的任务列表页（ADR-0002）：分类 chips + 任务列表
// + 详情（RenderDetail 逃生门或 Form 自动渲染）。应用把它作为一个导航
// 条目挂载：ui.NavEntry{ID: "tasks", Title: "任务", Render: ui.TaskListPage()}。
func TaskListPage() func(*Ctx) {
	return func(ctx *Ctx) {
		shell := ctx.Shell
		if shell == nil {
			return
		}
		store := shell.Store()
		tasks := shell.Tasks()

		avail := imgui.ContentRegionAvail()
		listW := float32(ctx.S(260))
		if avail.X < float32(ctx.S(520)) {
			listW = avail.X * 0.42
			if listW < float32(ctx.S(160)) {
				listW = float32(ctx.S(160))
			}
		}

		imgui.BeginChildStrV("panel_list", imgui.Vec2{X: listW, Y: 0}, imgui.ChildFlagsBorders, imgui.WindowFlagsNone)
		renderCatChips(ctx, store, tasks)
		imgui.Separator()
		renderTaskRows(ctx, store, tasks)
		imgui.EndChild()

		imgui.SameLine()
		imgui.BeginChildStrV("panel_detail", imgui.Vec2{X: 0, Y: 0}, imgui.ChildFlagsBorders, imgui.WindowFlagsNone)
		renderTaskDetail(ctx, store, tasks, ctx.theme())
		imgui.EndChild()
	}
}

// 面板间距令牌：按源 cookie_panel_android.go:15-20 的用法，仅引入本文件需要的
// 部分；gapL 已由 panel_android.go 定义，避免同包冲突。
const (
	gapXS = float32(4)
	gapS  = float32(8)
)

// renderCatChips 绘制分类筛选 chips：「全部」+ 任务表动态推导出的分类。
func renderCatChips(ctx *Ctx, store *Store, tasks []Task) {
	cats := Categories(tasks)
	chips := make([]struct{ id, label string }, 0, len(cats)+1)
	chips = append(chips, struct{ id, label string }{PanelCatAll, "全部"})
	for _, cat := range cats {
		chips = append(chips, struct{ id, label string }{cat, categoryLabel(cat)})
	}

	const padX, padY = float32(8), float32(6) // padX 收窄，保证 4 个 chip 在最小列宽内不换行
	const gap = float32(6)
	th := ctx.theme()
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
		active := store.GetString(KeyPanelCat) == c.id
		if active {
			imgui.PushStyleColorVec4(imgui.ColButton, toVec4(th.Accent))
			imgui.PushStyleColorVec4(imgui.ColText, toVec4(th.White))
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

// renderTaskRows 绘制当前分类下的任务行：勾选标记 + 标题 + 摘要。
func renderTaskRows(ctx *Ctx, store *Store, tasks []Task) {
	cat := store.GetString(KeyPanelCat)
	if cat == "" {
		cat = PanelCatAll
	}
	filtered := FilterByCategory(tasks, cat)
	selected := store.GetString(KeyPanelSelected)
	th := ctx.theme()

	for _, task := range filtered {
		on := task.EnabledKey != "" && store.GetBool(task.EnabledKey)
		// 左侧留空给复选框式启用标记（手绘，不依赖字体字形），标题 5 空格缩进。
		label := "     " + task.Title
		selectedHere := selected == task.ID
		if selectedHere {
			imgui.PushStyleColorVec4(imgui.ColHeader, toVec4(th.Accent))
			imgui.PushStyleColorVec4(imgui.ColHeaderHovered, toVec4(th.Accent))
			imgui.PushStyleColorVec4(imgui.ColHeaderActive, toVec4(th.Accent))
			imgui.PushStyleColorVec4(imgui.ColText, toVec4(th.White))
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
		drawListCheck(ctx, imgui.WindowDrawList(), rowPos, textH, on)
		if selectedHere {
			imgui.PopStyleColorV(4)
		}
		if task.Summary != nil {
			if sum := task.Summary(store); sum != "" {
				// 摘要与标题同一缩进（5 空格 ≈ listTitleX）；TextWrapped 折行后
				// 第二行也从该缩进起，不会孤「·」顶行首。
				imgui.PushStyleColorVec4(imgui.ColText, toVec4(th.TextDisabled))
				imgui.TextWrapped("     " + sum)
				imgui.PopStyleColor()
			}
		}
	}
	if len(filtered) == 0 {
		imgui.TextDisabled("（该分类暂无任务）")
	}
}

// drawListCheck 列表行启用标记（复选框样式）：白底圆角方框 + 描边，启用时
// 框内天蓝对号，未启用空框；蓝底选中行上白框依然醒目。
func drawListCheck(ctx *Ctx, dl *imgui.DrawList, rowPos imgui.Vec2, textH float32, on bool) {
	const box = float32(16)
	const framePadY = float32(4)
	x := rowPos.X + 8
	y := rowPos.Y + framePadY + (textH-box)/2
	pMin := imgui.Vec2{X: x, Y: y}
	pMax := imgui.Vec2{X: x + box, Y: y + box}
	th := ctx.theme()
	dl.AddRectFilledV(pMin, pMax, imgui.ColorU32Vec4(toVec4(th.White)), 3, imgui.DrawFlagsRoundCornersAll)
	dl.AddRectV(pMin, pMax, imgui.ColorU32Vec4(toVec4(th.Border)), 3, imgui.DrawFlagsRoundCornersAll, 1)
	if !on {
		return
	}
	col := imgui.ColorU32Vec4(toVec4(th.Accent))
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

// renderTaskDetail 绘制右侧任务详情：头部 + 自定义 RenderDetail 或自动 Form。
func renderTaskDetail(ctx *Ctx, store *Store, tasks []Task, th Theme) {
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

	// 头部：任务名 + 状态胶囊 + 分类（弱色）。
	imgui.Text(task.Title)
	imgui.SameLineV(0, gapS)
	on := task.EnabledKey != "" && store.GetBool(task.EnabledKey)
	drawEnabledPill(ctx, on)
	imgui.SameLineV(0, gapS)
	drawPill(ctx, categoryLabel(task.Category), th.TextDisabled)
	imgui.Dummy(imgui.Vec2{X: 0, Y: gapXS})
	imgui.Separator()
	imgui.Dummy(imgui.Vec2{X: 0, Y: gapXS})

	if task.RenderDetail != nil {
		ctx.Push("detail:" + task.ID)
		task.RenderDetail(ctx)
		ctx.Pop()
	} else {
		Form(ctx, FormProps{Store: store, Fields: task.Fields})
	}
}

// drawEnabledPill 状态胶囊：已启用=天蓝底白字 / 未启用=灰蓝底白字。
func drawEnabledPill(ctx *Ctx, on bool) {
	text := "未启用"
	bg := ctx.theme().TextDisabled
	if on {
		text = "已启用"
		bg = ctx.theme().Accent
	}
	drawPill(ctx, text, bg)
}

// drawPill 文字胶囊：圆角底色块 + 白字，画完光标随之前移。
func drawPill(ctx *Ctx, text string, bg Color) {
	sz := measureLabelSize(text)
	const padX, padY = float32(8), float32(1)
	w := sz.X + padX*2
	h := sz.Y + padY*2
	pos := imgui.CursorScreenPos()
	dl := imgui.WindowDrawList()
	dl.AddRectFilledV(
		pos,
		imgui.Vec2{X: pos.X + w, Y: pos.Y + h},
		imgui.ColorU32Vec4(toVec4(bg)), h/2, imgui.DrawFlagsRoundCornersAll,
	)
	dl.AddTextVec2V(
		imgui.Vec2{X: pos.X + padX, Y: pos.Y + padY},
		imgui.ColorU32Vec4(toVec4(ctx.theme().White)),
		text,
	)
	imgui.Dummy(imgui.Vec2{X: w, Y: h})
}

// categoryLabel 把任务分类内部标识转换为中文标签；未知或已是中文则原样返回。
func categoryLabel(c string) string {
	switch c {
	case "daily":
		return "日常"
	case "event":
		return "活动"
	case "maint":
		return "维护"
	default:
		return c
	}
}
