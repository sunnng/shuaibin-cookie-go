//go:build android && cgo

package ui

import "github.com/Dasongzi1366/AutoGo/imgui"

// Form 表单组件（ADR-0003）：按 Fields 自动排版为两列栅格（标签列固定宽 =
// 最长标签 + 12 间距，控件列占剩余宽度），值直连 Store 读写。
// 任务详情页默认渲染器；自定义 section 也可在 RenderDetail 内复用。
//
// 每个字段以 ctx.Push(f.Key()) 隔离组件状态、以 imgui.PushIDStr(f.Key())
// 隔离 imgui ID（空 Label 控件的 ID 恒定为 "##xxx_"，无此隔离同类型字段会碰撞）。
// Fields 内 Key 必须唯一（重复键会使状态路径与 imgui ID 双双碰撞）。
func Form(ctx *Ctx, p FormProps) {
	if p.Store == nil {
		return
	}
	const gapM = float32(12)
	// 行间距用包级 gapS（taskpage_android.go），不在此重复定义。

	labelW := float32(0)
	for _, f := range p.Fields {
		if w := measureLabelSize(f.Label()).X; w > labelW {
			labelW = w
		}
	}
	controlX := imgui.CursorPosX() + labelW + gapM

	for i, f := range p.Fields {
		ctx.Push(f.Key())
		imgui.PushIDStr(f.Key())
		imgui.AlignTextToFramePadding()
		imgui.Text(f.Label())
		imgui.SameLineV(0, 0)
		imgui.SetCursorPosX(controlX)

		switch f.Widget() {
		case WidgetCheckbox:
			Checkbox(ctx, CheckboxProps{
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
		if i < len(p.Fields)-1 {
			imgui.Dummy(imgui.Vec2{X: 0, Y: gapS})
		}
	}
}
