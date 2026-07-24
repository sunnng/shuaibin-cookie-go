//go:build android && cgo

package ui

import "github.com/Dasongzi1366/AutoGo/imgui"

// TaskListPage 框架可复用的任务列表页（ADR-0002）：分类 chips + 任务卡片列表
// + 详情（RenderDetail 逃生门或 Form 自动渲染）。应用把它作为一个导航
// 条目挂载：ui.NavEntry{ID: "tasks", Title: "任务", Render: ui.TaskListPage()}。
// 视觉为糖果积木语言（design-system.md §4.3）：任务卡片带分类色条与大开关。
func TaskListPage() func(*Ctx) {
	return func(ctx *Ctx) {
		shell := ctx.Shell
		if shell == nil {
			return
		}
		store := shell.Store()
		tasks := shell.Tasks()

		avail := imgui.ContentRegionAvail()
		listW := float32(ctx.S(300))
		if avail.X < float32(ctx.S(600)) {
			listW = avail.X * 0.42
			if listW < float32(ctx.S(180)) {
				listW = float32(ctx.S(180))
			}
		}

		imgui.PushStyleVarFloat(imgui.StyleVarChildBorderSize, 3)
		imgui.BeginChildStrV("panel_list", imgui.Vec2{X: listW, Y: 0}, imgui.ChildFlagsBorders, imgui.WindowFlagsNone)
		renderCatChips(ctx, store, tasks)
		imgui.Dummy(imgui.Vec2{X: 0, Y: float32(ctx.S(10))})
		renderTaskRows(ctx, store, tasks)
		imgui.EndChild()
		imgui.PopStyleVar()

		imgui.SameLine()
		imgui.BeginChildStrV("panel_detail", imgui.Vec2{X: 0, Y: 0}, imgui.ChildFlagsNone, imgui.WindowFlagsNone)
		renderTaskDetail(ctx, store, tasks, ctx.theme())
		imgui.EndChild()
	}
}

// 面板间距令牌：gapL 已由 panel_android.go 定义，避免同包冲突。
const (
	gapXS = float32(4)
	gapS  = float32(8)
)

// renderCatChips 绘制分类筛选 chips（36px 高胶囊）：「全部」+ 任务表动态推导
// 出的分类。选中 = 墨底纸字 + 小硬阴影；未选 = 白底 2px 墨描边。
func renderCatChips(ctx *Ctx, store *Store, tasks []Task) {
	cats := Categories(tasks)
	chips := make([]struct{ id, label string }, 0, len(cats)+1)
	chips = append(chips, struct{ id, label string }{PanelCatAll, "全部"})
	for _, cat := range cats {
		chips = append(chips, struct{ id, label string }{cat, categoryLabel(cat)})
	}

	padX := float32(ctx.S(14))
	chipH := float32(ctx.S(36))
	gap := float32(ctx.S(8))

	lineStart := true
	for _, c := range chips {
		textSz := measureLabelSize(c.label)
		w := textSz.X + padX*2
		if !lineStart {
			remain := imgui.ContentRegionAvail().X
			if remain < w+gap {
				lineStart = true
			} else {
				imgui.SameLineV(0, gap)
			}
		}

		pos := imgui.CursorScreenPos()
		size := imgui.Vec2{X: w, Y: chipH}
		clicked := imgui.InvisibleButton("##cat_"+c.id, size)
		pressed := imgui.IsItemActive()

		active := store.GetString(KeyPanelCat) == c.id
		dl := imgui.WindowDrawList()
		pMin, pMax := pos, imgui.Vec2{X: pos.X + w, Y: pos.Y + chipH}
		if active {
			off := float32(0)
			shadow := float32(3)
			if pressed {
				off = 2
				shadow = 1
			}
			pMin = imgui.Vec2{X: pos.X + off, Y: pos.Y + off}
			pMax = imgui.Vec2{X: pos.X + w + off, Y: pos.Y + chipH + off}
			// 选中：墨底纸字；阴影用半透明墨（小元素避免糊）。
			dl.AddRectFilledV(
				imgui.Vec2{X: pMin.X + shadow, Y: pMin.Y + shadow},
				imgui.Vec2{X: pMax.X + shadow, Y: pMax.Y + shadow},
				imgui.ColorU32Vec4(imgui.Vec4{X: 0.090, Y: 0.075, Z: 0.047, W: 0.3}), chipH/2, imgui.DrawFlagsRoundCornersAll,
			)
			dl.AddRectFilledV(pMin, pMax, imgui.ColorU32Vec4(candyInk), chipH/2, imgui.DrawFlagsRoundCornersAll)
		} else {
			drawBlock(dl, pMin, pMax, candyRaised, chipH/2, 2, 0)
		}
		textCol := candyInk
		if active {
			textCol = candyPaper
		}
		dl.AddTextVec2V(
			imgui.Vec2{X: pMin.X + padX, Y: pMin.Y + (chipH-textSz.Y)/2},
			imgui.ColorU32Vec4(textCol),
			c.label,
		)

		if clicked {
			store.SetString(KeyPanelCat, c.id)
		}
		lineStart = false
	}
}

// renderTaskRows 绘制当前分类下的任务卡片：左侧分类色条 + 标题 + 摘要 +
// 右侧大开关（直接控制启用，不必进入详情）。选中卡 = 硬阴影 + 奶白底 + 黄色条。
func renderTaskRows(ctx *Ctx, store *Store, tasks []Task) {
	cat := store.GetString(KeyPanelCat)
	if cat == "" {
		cat = PanelCatAll
	}
	filtered := FilterByCategory(tasks, cat)
	selected := store.GetString(KeyPanelSelected)

	dl := imgui.WindowDrawList()
	for _, task := range filtered {
		ctx.Push("task:" + task.ID)
		imgui.PushIDStr(task.ID)

		on := task.EnabledKey != "" && store.GetBool(task.EnabledKey)
		selectedHere := selected == task.ID

		summary := ""
		if task.Summary != nil {
			summary = task.Summary(store)
		}

		padV := float32(ctx.S(12))
		titleSz := measureLabelSize(task.Title)
		cardH := padV*2 + titleSz.Y
		var sumSz imgui.Vec2
		if summary != "" {
			sumSz = measureLabelSize(summary)
			// 摘要按 0.82 缩小绘制（见下方 drawFittedText），行高同步收缩。
			cardH += float32(ctx.S(5)) + sumSz.Y*0.82
		}

		avail := imgui.ContentRegionAvail()
		cardW := avail.X
		radius := float32(ctx.S(12))
		rowGap := float32(ctx.S(10))

		// 先用整卡 Dummy 登记行边界（imgui 不允许用 SetCursorPos 扩展父
		// 窗口边界，须先提交 item 撑开），卡内的绝对定位都在已登记范围内。
		pos := imgui.CursorScreenPos()
		imgui.Dummy(imgui.Vec2{X: cardW, Y: cardH + rowGap})

		// 开关热区先提交（后提交的卡片按钮与开关重叠时，先提交者优先命中）；
		// 热区纵向扩到 ≥48（触控目标下限），可视槽保持 56×32 居中。
		var swAnim *float32
		swSize := switchSize(ctx)
		swPos := imgui.Vec2{X: pos.X + cardW - float32(ctx.S(12)) - swSize.X, Y: pos.Y + (cardH-swSize.Y)/2}
		swClicked := false
		if task.EnabledKey != "" {
			swAnim = State(ctx, "swanim", switchAnimTarget(on))
			hitPos, hitSize := switchHitRect(ctx, swPos, swSize)
			imgui.SetCursorScreenPos(hitPos)
			swClicked = imgui.InvisibleButton("##switch", hitSize)
		}

		// 卡片选中热区（整卡）。
		imgui.SetCursorScreenPos(pos)
		cardClicked := imgui.InvisibleButton("##card", imgui.Vec2{X: cardW, Y: cardH})

		// 卡体：白底 3px 墨描边；选中 = 奶白底 + 硬阴影。
		fill := candyRaised
		shadow := float32(0)
		if selectedHere {
			fill = candySelCardBg
			shadow = 4
		}
		drawBlock(dl, pos, imgui.Vec2{X: pos.X + cardW, Y: pos.Y + cardH}, fill, radius, 3, shadow)

		// 左侧分类色条（选中卡变黄），与卡片同圆角贴左缘。
		barCol := candyCategoryColor(task.Category)
		if selectedHere {
			barCol = candyYellow
		}
		barW := float32(ctx.S(6))
		dl.AddRectFilledV(
			imgui.Vec2{X: pos.X, Y: pos.Y},
			imgui.Vec2{X: pos.X + barW, Y: pos.Y + cardH},
			imgui.ColorU32Vec4(barCol), radius, imgui.DrawFlagsRoundCornersLeft,
		)

		textX := pos.X + barW + float32(ctx.S(12))
		// 文本可用宽度：到开关左缘（无开关则到卡片右内边距），超出截断。
		textRight := pos.X + cardW - float32(ctx.S(12))
		if task.EnabledKey != "" {
			textRight = swPos.X - float32(ctx.S(8))
		}
		textMaxW := textRight - textX
		drawFittedText(dl,
			imgui.Vec2{X: textX, Y: pos.Y + padV},
			candyInk, task.Title, textMaxW, 1.0,
		)
		if summary != "" {
			// 摘要按设计字阶缩小到 13px（≈0.82），防止长摘要溢出卡片。
			drawFittedText(dl,
				imgui.Vec2{X: textX, Y: pos.Y + padV + titleSz.Y + float32(ctx.S(5))},
				candySec, summary, textMaxW, 0.82,
			)
		}

		// 大开关（提交于卡片之前，命中优先）。
		if task.EnabledKey != "" {
			stepSwitchAnim(swAnim, on)
			drawSwitchVisual(dl, swPos, swSize, *swAnim)
		}

		switch {
		case swClicked && task.EnabledKey != "":
			store.SetBool(task.EnabledKey, !on)
		case cardClicked:
			store.SetString(KeyPanelSelected, task.ID)
		}

		imgui.PopID()
		ctx.Pop()
	}
	if len(filtered) == 0 {
		imgui.TextDisabled("（该分类暂无任务）")
	}
}

// renderTaskDetail 绘制右侧任务详情：头部（任务名 + 状态胶囊 + 分类胶囊）
// + 自定义 RenderDetail 或自动 Form。
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

	// 头部：任务名 20px + 状态胶囊（已启用=绿 / 未启用=灰）+ 分类胶囊
	// （2px 描边白底）。本行全是文字/手绘元素，不用 AlignTextToFramePadding。
	imgui.Dummy(imgui.Vec2{X: 0, Y: gapXS})
	imgui.Text(task.Title)
	imgui.SameLineV(0, gapS)
	on := task.EnabledKey != "" && store.GetBool(task.EnabledKey)
	drawEnabledPill(ctx, on)
	imgui.SameLineV(0, gapS)
	drawPill(ctx, categoryLabel(task.Category), candyRaised, candyInk, candyInk)
	imgui.Dummy(imgui.Vec2{X: 0, Y: gapS})

	// 头部下方 2px 墨线分隔。
	dl := imgui.WindowDrawList()
	linePos := imgui.CursorScreenPos()
	lineY := linePos.Y - gapS/2
	imgui.Dummy(imgui.Vec2{X: 0, Y: gapS})
	dl.AddLineV(
		imgui.Vec2{X: linePos.X, Y: lineY},
		imgui.Vec2{X: linePos.X + imgui.ContentRegionAvail().X, Y: lineY},
		imgui.ColorU32Vec4(candyInk), 2,
	)

	// 两条详情路径都按任务隔离组件状态（同名字段键的草稿缓冲不互串）。
	ctx.Push("detail:" + task.ID)
	if task.RenderDetail != nil {
		task.RenderDetail(ctx)
	} else {
		Form(ctx, FormProps{Store: store, Fields: task.Fields})
	}
	ctx.Pop()
}

// drawEnabledPill 状态胶囊：已启用=绿底墨字 / 未启用=灰底灰字（均 2px 描边）。
func drawEnabledPill(ctx *Ctx, on bool) {
	if on {
		drawPill(ctx, "已启用", candyGreen, candyInk, candyInk)
	} else {
		drawPill(ctx, "未启用", candyPillOff, candySec, candySec)
	}
}

// drawPill 文字胶囊：圆角底色块 + 2px 描边 + 文字，画完光标随之前移。
func drawPill(ctx *Ctx, text string, bg, border, textCol imgui.Vec4) {
	sz := measureLabelSize(text)
	padX, padY := float32(ctx.S(10)), float32(ctx.S(3))
	w := sz.X + padX*2
	h := sz.Y + padY*2
	pos := imgui.CursorScreenPos()
	dl := imgui.WindowDrawList()
	pMin := pos
	pMax := imgui.Vec2{X: pos.X + w, Y: pos.Y + h}
	dl.AddRectFilledV(pMin, pMax, imgui.ColorU32Vec4(bg), h/2, imgui.DrawFlagsRoundCornersAll)
	dl.AddRectV(pMin, pMax, imgui.ColorU32Vec4(border), h/2, imgui.DrawFlagsRoundCornersAll, 2)
	dl.AddTextVec2V(
		imgui.Vec2{X: pos.X + padX, Y: pos.Y + padY},
		imgui.ColorU32Vec4(textCol),
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
