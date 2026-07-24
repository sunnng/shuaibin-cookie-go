//go:build android && cgo

package ui

import "github.com/Dasongzi1366/AutoGo/imgui"

// Form 表单组件（ADR-0003）：按 Fields 自动排版为两列栅格（标签列固定宽 =
// 最长标签 + 12 间距，控件列占剩余宽度），值直连 Store 读写。
// 任务详情页默认渲染器；自定义 section 也可在 RenderDetail 内复用。
//
// 视觉（design-system.md §4.3）：行高 56，行间 2px 浅棕虚线分隔；
// Bool → Switch 开关（仅绘制替换，Props 不变），Number → 三段式步进器，
// Text → 3px 描边积木框。
//
// 每个字段以 ctx.Push(f.Key()) 隔离组件状态、以 imgui.PushIDStr(f.Key())
// 隔离 imgui ID（空 Label 控件的 ID 恒定为 "##xxx_"，无此隔离同类型字段会碰撞）。
// Fields 内 Key 必须唯一（重复键会使状态路径与 imgui ID 双双碰撞）。
func Form(ctx *Ctx, p FormProps) {
	if p.Store == nil {
		return
	}
	const gapM = float32(12)

	rowH := float32(ctx.S(56))

	labelW := float32(0)
	for _, f := range p.Fields {
		if w := measureLabelSize(f.Label()).X; w > labelW {
			labelW = w
		}
	}
	contentX := imgui.CursorScreenPos().X
	controlX := contentX + labelW + gapM
	contentW := imgui.ContentRegionAvail().X
	dl := imgui.WindowDrawList()

	for i, f := range p.Fields {
		rowTop := imgui.CursorScreenPos()
		// 先用整行 Dummy 登记边界（imgui 不允许用 SetCursorPos 扩展父窗口
		// 边界，须先提交 item 撑开），行内的光标移动都在已登记范围内。
		imgui.Dummy(imgui.Vec2{X: contentW, Y: rowH})

		ctx.Push(f.Key())
		imgui.PushIDStr(f.Key())

		// 标签：行内垂直居中。
		labelSz := measureLabelSize(f.Label())
		imgui.SetCursorScreenPos(imgui.Vec2{X: rowTop.X, Y: rowTop.Y + (rowH-labelSz.Y)/2})
		imgui.Text(f.Label())

		// 控件：固定列起点的右列，行内垂直居中。
		ctrlH := formControlHeight(ctx, f)
		imgui.SetCursorScreenPos(imgui.Vec2{X: controlX, Y: rowTop.Y + (rowH-ctrlH)/2})

		switch f.Widget() {
		case WidgetCheckbox:
			Switch(ctx, CheckboxProps{
				Checked: FormFieldValue(p.Store, f).(bool),
				OnChange: func(v bool) {
					FormFieldChanged(p.Store, f, v)
				},
			})
		case WidgetNumberInput:
			c := f.Constraints()
			NumberInput(ctx, NumberInputProps{
				Value: FormFieldValue(p.Store, f).(float64),
				Min:   c.Min, Max: c.Max, Step: c.Step,
				OnChange: func(v float64) {
					FormFieldChanged(p.Store, f, v)
				},
			})
		default:
			TextInput(ctx, InputProps{
				Value: FormFieldValue(p.Store, f).(string),
				OnChange: func(v string) {
					FormFieldChanged(p.Store, f, v)
				},
			})
		}
		imgui.PopID()
		ctx.Pop()

		// 行间画 2px 浅棕虚线分隔；光标回到下一行行首（未越出已登记边界）。
		if i < len(p.Fields)-1 {
			drawDashedHLine(dl, contentX, contentX+contentW, rowTop.Y+rowH)
		}
		imgui.SetCursorScreenPos(imgui.Vec2{X: contentX, Y: rowTop.Y + rowH})
	}
}

// formControlHeight 各类控件的可视高度（用于行内垂直居中）。
func formControlHeight(ctx *Ctx, f Field) float32 {
	switch f.Widget() {
	case WidgetCheckbox:
		_, size := switchVisualRect(ctx)
		return size.Y
	default: // 步进器 / 文本框统一 44
		return float32(ctx.S(44))
	}
}
