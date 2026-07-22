//go:build android && cgo

package ui

import (
	"strconv"

	"github.com/Dasongzi1366/AutoGo/imgui"
)

// Dropdown 下拉框组件（ADR-0003）。标签在左，触发按钮在右，按钮显示当前选中文本；
// 点击展开弹出层，选择某项后调用 p.OnChange(i) 并收起。不写 Store，完全受控。
//
// 注意：ID 以 Label 区分；同一窗口内多个 Label 相同的 Dropdown 会产生 ID 碰撞，
// 调用方应自行使用 ctx.Push 隔离。本组件使用 State(ctx, "open", false) 托管弹层展开态，
// 同一路径下多个实例会共享 open 缓冲，必须通过 ctx.Push 隔离。
func Dropdown(ctx *Ctx, p DropdownProps) {
	th := ctx.theme()
	if len(p.Options) == 0 {
		return
	}

	const paddingX = float32(8)
	const paddingY = float32(4)
	const maxVisible = 5
	const edgeMargin = float32(6)

	id := p.Label
	if id == "" {
		id = "dropdown"
	}

	selected := p.Selected
	if selected < 0 || selected >= len(p.Options) {
		selected = 0
	}

	if p.Label != "" {
		imgui.AlignTextToFramePadding()
		imgui.PushStyleColorVec4(imgui.ColText, toVec4(th.Text))
		imgui.Text(p.Label)
		imgui.PopStyleColorV(1)
		imgui.SameLine()
	}

	avail := imgui.ContentRegionAvail()
	comboW := avail.X
	if comboW < 60 {
		comboW = 60
	}

	popupID := "DROPDOWN_POPUP_" + id
	buttonID := p.Options[selected] + "   ##DROPDOWN_BUTTON_" + id
	open := State(ctx, "open", false)

	imgui.PushStyleColorVec4(imgui.ColButton, toVec4(th.Button))
	imgui.PushStyleColorVec4(imgui.ColButtonHovered, toVec4(th.ButtonHover))
	imgui.PushStyleColorVec4(imgui.ColButtonActive, toVec4(th.ButtonActive))
	imgui.PushStyleColorVec4(imgui.ColText, toVec4(th.Text))
	imgui.PushStyleColorVec4(imgui.ColBorder, toVec4(th.Border))
	imgui.PushStyleColorVec4(imgui.ColPopupBg, toVec4(th.PopupBg))
	imgui.PushStyleColorVec4(imgui.ColHeader, toVec4(th.Header))
	imgui.PushStyleColorVec4(imgui.ColHeaderHovered, toVec4(th.HeaderHover))
	defer imgui.PopStyleColorV(8)

	imgui.PushStyleVarVec2(imgui.StyleVarFramePadding, imgui.Vec2{X: paddingX, Y: paddingY})
	imgui.PushStyleVarFloat(imgui.StyleVarFrameBorderSize, 1.5)
	imgui.PushStyleVarFloat(imgui.StyleVarFrameRounding, 6)
	imgui.PushStyleVarFloat(imgui.StyleVarPopupRounding, 6)
	imgui.PushStyleVarVec2(imgui.StyleVarButtonTextAlign, imgui.Vec2{X: 0, Y: 0.5})
	defer imgui.PopStyleVarV(5)

	buttonClicked := imgui.ButtonV(buttonID, imgui.Vec2{X: comboW, Y: 0})
	if buttonClicked {
		if imgui.IsPopupOpenStr(popupID) {
			*open = false
		} else {
			*open = true
		}
	}
	if *open && !imgui.IsPopupOpenStr(popupID) && !buttonClicked {
		*open = false
	}
	if *open {
		imgui.OpenPopupStr(popupID)
	}

	itemMin := imgui.ItemRectMin()
	itemMax := imgui.ItemRectMax()
	itemSize := imgui.ItemRectSize()
	itemH := imgui.FrameHeight()
	winPadY := imgui.CurrentStyle().WindowPadding().Y

	visibleCount := len(p.Options)
	if visibleCount > maxVisible {
		visibleCount = maxVisible
	}
	popupH := itemH*float32(visibleCount) + winPadY*2

	viewport := imgui.MainViewport()
	vpPos := viewport.Pos()
	vpSize := viewport.Size()
	screenTop := vpPos.Y + edgeMargin
	screenBottom := vpPos.Y + vpSize.Y - edgeMargin
	spaceAbove := itemMin.Y - screenTop
	spaceBelow := screenBottom - itemMax.Y
	if spaceAbove < 0 {
		spaceAbove = 0
	}
	if spaceBelow < 0 {
		spaceBelow = 0
	}
	openUpward := spaceBelow < popupH && spaceAbove > spaceBelow

	{
		dl := imgui.WindowDrawList()
		fontH := imgui.CalcTextSize("A").Y
		half := fontH * 0.28
		cx := itemMax.X - paddingX - half
		cy := (itemMin.Y + itemMax.Y) * 0.5
		col := imgui.ColorConvertFloat4ToU32(toVec4(th.Text))
		if openUpward {
			dl.AddTriangleFilled(
				imgui.Vec2{X: cx - half, Y: cy + half*0.5},
				imgui.Vec2{X: cx + half, Y: cy + half*0.5},
				imgui.Vec2{X: cx, Y: cy - half*0.5},
				col,
			)
		} else {
			dl.AddTriangleFilled(
				imgui.Vec2{X: cx - half, Y: cy - half*0.5},
				imgui.Vec2{X: cx + half, Y: cy - half*0.5},
				imgui.Vec2{X: cx, Y: cy + half*0.5},
				col,
			)
		}
	}

	finalPopupH := popupH
	if openUpward {
		if finalPopupH > spaceAbove {
			finalPopupH = spaceAbove
		}
	} else {
		if finalPopupH > spaceBelow {
			finalPopupH = spaceBelow
		}
	}
	minPopupH := itemH + winPadY*2
	if finalPopupH < minPopupH {
		finalPopupH = minPopupH
	}

	popupPos := imgui.Vec2{X: itemMin.X, Y: itemMax.Y}
	if openUpward {
		popupPos.Y = itemMin.Y - finalPopupH
	}

	imgui.SetNextWindowPosV(popupPos, imgui.CondAlways, imgui.Vec2{})
	imgui.SetNextWindowSizeV(imgui.Vec2{X: itemSize.X, Y: finalPopupH}, imgui.CondAlways)

	if imgui.BeginPopup(popupID) {
		for i, opt := range p.Options {
			isSelected := i == selected
			if imgui.SelectableBoolV(opt+"##dropdown_opt_"+id+"_"+strconv.Itoa(i), isSelected, 0, imgui.Vec2{X: itemSize.X, Y: 0}) {
				*open = false
				imgui.CloseCurrentPopup()
				if p.OnChange != nil {
					p.OnChange(i)
				}
			}
			if isSelected {
				imgui.SetItemDefaultFocus()
			}
		}
		imgui.EndPopup()
	}
}

// Tabs 顶部标签栏组件（ADR-0003）。渲染为横向按钮行，当前选中项 p.Selected 使用主题
// Accent 底色 + 白字，其余项使用按钮色 + 文字色；点击时调用 p.OnChange(i)。
// 本组件只渲染标签栏，不渲染内容区；调用方按自己的选中变量自行 switch：
//
//	ui.Tabs(ctx, ui.TabsProps{Items: []string{"A", "B"}, Selected: sel, OnChange: func(i int) { sel = i }})
//	switch sel {
//	case 0:
//	    // 渲染 A 页面
//	case 1:
//	    // 渲染 B 页面
//	}
//
// 注意：ID 以首个 Item 区分；同一窗口内多个首个 Item 相同的 Tabs 会产生 ID 碰撞，
// 调用方应自行使用 ctx.Push 隔离（组件状态同理）。
func Tabs(ctx *Ctx, p TabsProps) {
	th := ctx.theme()
	if len(p.Items) == 0 {
		return
	}

	id := p.Items[0]
	if id == "" {
		id = "tabs"
	}

	const tabPadX, tabPadY = float32(14), float32(10)

	imgui.PushStyleVarVec2(imgui.StyleVarFramePadding, imgui.Vec2{X: tabPadX, Y: tabPadY})
	imgui.PushStyleVarVec2(imgui.StyleVarButtonTextAlign, imgui.Vec2{X: 0.5, Y: 0.5})
	imgui.PushStyleColorVec4(imgui.ColChildBg, toVec4(th.ChildBg))
	imgui.PushStyleColorVec4(imgui.ColBorder, toVec4(th.Border))
	imgui.PushStyleVarFloat(imgui.StyleVarChildRounding, 10)
	imgui.PushStyleVarVec2(imgui.StyleVarWindowPadding, imgui.Vec2{X: 8, Y: 10})
	defer imgui.PopStyleVarV(4)
	defer imgui.PopStyleColorV(2)

	tabH := imgui.FrameHeight()
	tabBarHeight := tabH + 20

	if imgui.BeginChildStrV(id+"_tab_bar", imgui.Vec2{X: 0, Y: tabBarHeight}, imgui.ChildFlagsBorders, imgui.WindowFlagsNone) {
		for i, title := range p.Items {
			active := i == p.Selected

			var bg, bgHover, bgActive, text, border imgui.Vec4
			if active {
				bg = toVec4(th.Accent)
				bgHover = toVec4(th.Accent)
				bgActive = toVec4(th.TitleBottom)
				text = toVec4(th.White)
				border = toVec4(th.Border)
			} else {
				bg = toVec4(th.Button)
				bgHover = toVec4(th.ButtonHover)
				bgActive = toVec4(th.ButtonActive)
				text = toVec4(th.Text)
				border = toVec4(th.Border)
			}

			imgui.PushStyleColorVec4(imgui.ColButton, bg)
			imgui.PushStyleColorVec4(imgui.ColButtonHovered, bgHover)
			imgui.PushStyleColorVec4(imgui.ColButtonActive, bgActive)
			imgui.PushStyleColorVec4(imgui.ColText, text)
			imgui.PushStyleColorVec4(imgui.ColBorder, border)
			imgui.PushStyleVarFloat(imgui.StyleVarFrameBorderSize, 1)
			imgui.PushStyleVarFloat(imgui.StyleVarFrameRounding, 10)

			labelSz := measureLabelSize(title)
			tabW := labelSz.X + tabPadX*2
			if tabW < 48 {
				tabW = 48
			}
			thisH := tabH
			if need := labelSz.Y + tabPadY*2; need > thisH {
				thisH = need
			}

			if imgui.ButtonV(title+"##"+id+"_tab_"+strconv.Itoa(i), imgui.Vec2{X: tabW, Y: thisH}) && p.OnChange != nil {
				p.OnChange(i)
			}

			imgui.PopStyleVarV(2)
			imgui.PopStyleColorV(5)

			if i != len(p.Items)-1 {
				imgui.SameLine()
			}
		}
		imgui.EndChild()
	}

	imgui.Spacing()
}

// Collapsible 折叠面板组件（ADR-0003）。标题行是一个全宽按钮，点击切换展开态；
// 展开时调用 content() 渲染内容。展开态使用 State(ctx, "open", p.Open) 托管，
// 切换时调用 p.OnToggle(*open)。
//
// 注意：ID 以 Label 区分；同一窗口内多个 Label 相同的 Collapsible 会产生 ID 碰撞，
// 调用方应自行使用 ctx.Push 隔离。本组件使用 State(ctx, "open", ...) 托管展开态，
// 同一路径下多个实例会共享 open 缓冲，必须通过 ctx.Push 隔离。
func Collapsible(ctx *Ctx, p CollapsibleProps, content func()) {
	th := ctx.theme()

	const paddingX, paddingY = float32(8), float32(6)

	id := p.Label
	if id == "" {
		id = "collapsible"
	}

	open := State(ctx, "open", p.Open)
	avail := imgui.ContentRegionAvail()
	btnW := avail.X

	imgui.PushStyleColorVec4(imgui.ColButton, toVec4(th.Button))
	imgui.PushStyleColorVec4(imgui.ColButtonHovered, toVec4(th.ButtonHover))
	imgui.PushStyleColorVec4(imgui.ColButtonActive, toVec4(th.ButtonActive))
	imgui.PushStyleColorVec4(imgui.ColText, toVec4(th.Text))
	imgui.PushStyleColorVec4(imgui.ColBorder, toVec4(th.Border))
	defer imgui.PopStyleColorV(5)

	imgui.PushStyleVarVec2(imgui.StyleVarFramePadding, imgui.Vec2{X: paddingX, Y: paddingY})
	imgui.PushStyleVarFloat(imgui.StyleVarFrameBorderSize, 1.5)
	imgui.PushStyleVarFloat(imgui.StyleVarFrameRounding, 6)
	imgui.PushStyleVarVec2(imgui.StyleVarButtonTextAlign, imgui.Vec2{X: 0, Y: 0.5})
	defer imgui.PopStyleVarV(4)

	clicked := imgui.ButtonV(p.Label+"##COLLAPSE_"+id, imgui.Vec2{X: btnW, Y: 0})

	{
		dl := imgui.WindowDrawList()
		itemMin := imgui.ItemRectMin()
		itemMax := imgui.ItemRectMax()
		fontH := imgui.CalcTextSize("A").Y
		half := fontH * 0.28
		cx := itemMax.X - paddingX - half
		cy := (itemMin.Y + itemMax.Y) * 0.5
		col := imgui.ColorConvertFloat4ToU32(toVec4(th.Text))
		if *open {
			dl.AddTriangleFilled(
				imgui.Vec2{X: cx - half, Y: cy + half*0.5},
				imgui.Vec2{X: cx + half, Y: cy + half*0.5},
				imgui.Vec2{X: cx, Y: cy - half*0.5},
				col,
			)
		} else {
			dl.AddTriangleFilled(
				imgui.Vec2{X: cx - half, Y: cy - half*0.5},
				imgui.Vec2{X: cx + half, Y: cy - half*0.5},
				imgui.Vec2{X: cx, Y: cy + half*0.5},
				col,
			)
		}
	}

	if clicked {
		*open = !*open
		if p.OnToggle != nil {
			p.OnToggle(*open)
		}
	}

	if *open && content != nil {
		imgui.Spacing()
		content()
		imgui.Spacing()
	}
}
