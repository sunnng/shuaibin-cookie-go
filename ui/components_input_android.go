//go:build android && cgo

package ui

import (
	"strconv"

	"github.com/Dasongzi1366/AutoGo/imgui"
)

// Checkbox 复选框组件（ADR-0003）。标签在左，勾选框紧跟右侧；勾选状态读取
// p.Checked，切换时调用 p.OnChange(checked)。若 p.Label 为空则只显示勾选框。
//
// 注意：ID 与 Button 同约（"##chk_"+Label）；同一窗口内多个 Label 相同（或同为空）
// 的复选框会产生 ID 碰撞，调用方应自行使用 ctx.Push 隔离（组件状态同理）。
func Checkbox(ctx *Ctx, p CheckboxProps) {
	th := ctx.theme()
	const paddingX, paddingY = float32(5), float32(4)

	if p.Label != "" {
		imgui.AlignTextToFramePadding()
		imgui.PushStyleColorVec4(imgui.ColText, toVec4(th.Text))
		imgui.Text(p.Label)
		imgui.PopStyleColorV(1)
		imgui.SameLine()
	}

	imgui.PushStyleVarVec2(imgui.StyleVarFramePadding, imgui.Vec2{X: paddingX, Y: paddingY})
	imgui.PushStyleVarFloat(imgui.StyleVarFrameBorderSize, 1.5)
	imgui.PushStyleVarFloat(imgui.StyleVarFrameRounding, 10)
	imgui.PushStyleColorVec4(imgui.ColFrameBg, toVec4(th.FrameBg))
	imgui.PushStyleColorVec4(imgui.ColFrameBgHovered, toVec4(th.FrameHover))
	imgui.PushStyleColorVec4(imgui.ColFrameBgActive, toVec4(th.FrameActive))
	imgui.PushStyleColorVec4(imgui.ColCheckMark, toVec4(th.Accent))
	imgui.PushStyleColorVec4(imgui.ColBorder, toVec4(th.Border))

	checked := p.Checked
	if imgui.Checkbox("##chk_"+p.Label, &checked) && p.OnChange != nil {
		p.OnChange(checked)
	}

	imgui.PopStyleColorV(5)
	imgui.PopStyleVarV(3)
}

// NumberInput 步进数字输入组件（ADR-0003）。左侧显示 p.Label，右侧为
// [-] [输入] [+] 步进控件。数值读取 p.Value，p.Step<=0 时视为 1；按钮或
// 编辑提交后先经 p.Clamp 钳制再调用 p.OnChange。
//
// 输入框使用组件状态缓存编辑字符串（State(ctx, "numbuf", "")），初始时从
// p.Value 格式化。v1 不回写外部 p.Value 的变更。
//
// 注意：缓冲键只按 Ctx 路径寻址；同一路径下渲染多个 NumberInput 会共享同一
// 缓冲（互相改写），调用方必须 ctx.Push 为每个实例隔离路径。ID 与 Button 同约
// （"##num_/##sub_/##add_"+Label），同 Label 实例同理会碰撞。
func NumberInput(ctx *Ctx, p NumberInputProps) {
	th := ctx.theme()
	const paddingX, paddingY = float32(8), float32(5)

	step := p.Step
	if step <= 0 {
		step = 1
	}

	if p.Label != "" {
		imgui.AlignTextToFramePadding()
		imgui.PushStyleColorVec4(imgui.ColText, toVec4(th.Text))
		imgui.Text(p.Label)
		imgui.PopStyleColorV(1)
		imgui.SameLine()
	}

	avail := imgui.ContentRegionAvail()
	var totalW float32
	if p.Width > 0 {
		totalW = float32(ctx.S(p.Width))
	} else {
		totalW = avail.X
	}
	if totalW < 80 {
		totalW = 80
	}

	fontH := imgui.CalcTextSize("A").Y
	btnW := fontH + paddingY*2
	inputW := totalW - btnW*2 - 32
	if inputW < 40 {
		inputW = 40
	}

	buf := State(ctx, "numbuf", "")
	if *buf == "" {
		*buf = strconv.FormatFloat(p.Value, 'f', -1, 64)
	}

	imgui.PushStyleVarVec2(imgui.StyleVarFramePadding, imgui.Vec2{X: paddingX, Y: paddingY})
	imgui.PushStyleVarFloat(imgui.StyleVarFrameBorderSize, 1.5)
	imgui.PushStyleVarFloat(imgui.StyleVarFrameRounding, 6)
	imgui.PushStyleColorVec4(imgui.ColFrameBg, toVec4(th.FrameBg))
	imgui.PushStyleColorVec4(imgui.ColBorder, toVec4(th.Border))
	imgui.PushStyleColorVec4(imgui.ColText, toVec4(th.Text))
	imgui.PushStyleColorVec4(imgui.ColButton, toVec4(th.Button))
	imgui.PushStyleColorVec4(imgui.ColButtonHovered, toVec4(th.ButtonHover))
	imgui.PushStyleColorVec4(imgui.ColButtonActive, toVec4(th.ButtonActive))
	defer imgui.PopStyleColorV(6)
	defer imgui.PopStyleVarV(3)

	if imgui.ButtonV("-##sub_"+p.Label, imgui.Vec2{X: btnW, Y: btnW}) && p.OnChange != nil {
		v := p.Clamp(p.Value - step)
		*buf = strconv.FormatFloat(v, 'f', -1, 64)
		p.OnChange(v)
	}
	imgui.SameLine()

	imgui.SetNextItemWidth(inputW)
	if imgui.InputTextWithHint("##num_"+p.Label, p.Hint, buf, imgui.InputTextFlagsCharsDecimal, nil) && p.OnChange != nil {
		if n, err := strconv.ParseFloat(*buf, 64); err == nil {
			v := p.Clamp(n)
			*buf = strconv.FormatFloat(v, 'f', -1, 64)
			p.OnChange(v)
		}
	}
	imgui.SameLine()

	if imgui.ButtonV("+##add_"+p.Label, imgui.Vec2{X: btnW, Y: btnW}) && p.OnChange != nil {
		v := p.Clamp(p.Value + step)
		*buf = strconv.FormatFloat(v, 'f', -1, 64)
		p.OnChange(v)
	}
}

// TextInput 文本输入组件（ADR-0003）。p.Multiline 为 true 时渲染多行输入框，
// 否则单行。p.Label 非空时显示为标签（单行在左侧，多行在上方）。p.Hint 为
// 占位提示。宽度 p.Width>0 时经 ctx.S 缩放，否则占满剩余宽度。
//
// 受控缓冲使用 State(ctx, "inbuf", p.Value) 托管；imgui 返回编辑时调用
// p.OnChange(*buf)。v1 不回写外部 p.Value 的变更。
//
// 注意：缓冲键只按 Ctx 路径寻址；同一路径下渲染多个 TextInput 会共享同一
// 缓冲（互相改写），调用方必须 ctx.Push 为每个实例隔离路径。ID 与 Button 同约
// （"##input_/##multi_"+Label），同 Label 实例同理会碰撞。多行框高度暂为
// 常量 80，未随 ctx.S 缩放（沿用移植源行为）。
func TextInput(ctx *Ctx, p InputProps) {
	th := ctx.theme()
	const paddingX, paddingY = float32(8), float32(5)

	buf := State(ctx, "inbuf", p.Value)

	if p.Multiline {
		if p.Label != "" {
			imgui.PushStyleColorVec4(imgui.ColText, toVec4(th.Text))
			imgui.Text(p.Label)
			imgui.PopStyleColorV(1)
		}

		avail := imgui.ContentRegionAvail()
		var inputW float32
		if p.Width > 0 {
			inputW = float32(ctx.S(p.Width))
		} else {
			inputW = avail.X
		}
		if inputW < 40 {
			inputW = 40
		}

		imgui.PushStyleColorVec4(imgui.ColFrameBg, toVec4(th.FrameBg))
		imgui.PushStyleColorVec4(imgui.ColText, toVec4(th.Text))
		imgui.PushStyleColorVec4(imgui.ColTextDisabled, toVec4(th.TextDisabled))
		imgui.PushStyleColorVec4(imgui.ColBorder, toVec4(th.Border))
		imgui.PushStyleColorVec4(imgui.ColFrameBgHovered, toVec4(th.FrameHover))
		imgui.PushStyleColorVec4(imgui.ColFrameBgActive, toVec4(th.FrameActive))
		defer imgui.PopStyleColorV(6)

		imgui.PushStyleVarVec2(imgui.StyleVarFramePadding, imgui.Vec2{X: paddingX, Y: paddingY})
		imgui.PushStyleVarFloat(imgui.StyleVarFrameBorderSize, 1.5)
		imgui.PushStyleVarFloat(imgui.StyleVarFrameRounding, 6)
		defer imgui.PopStyleVarV(3)

		const multilineH = float32(80)
		if imgui.InputTextMultiline("##multi_"+p.Label, buf, imgui.Vec2{X: inputW, Y: multilineH}, imgui.InputTextFlags(0), nil) && p.OnChange != nil {
			p.OnChange(*buf)
		}
		return
	}

	imgui.PushStyleColorVec4(imgui.ColFrameBg, toVec4(th.FrameBg))
	imgui.PushStyleColorVec4(imgui.ColText, toVec4(th.Text))
	imgui.PushStyleColorVec4(imgui.ColTextDisabled, toVec4(th.TextDisabled))
	imgui.PushStyleColorVec4(imgui.ColBorder, toVec4(th.Border))
	imgui.PushStyleColorVec4(imgui.ColFrameBgHovered, toVec4(th.FrameHover))
	imgui.PushStyleColorVec4(imgui.ColFrameBgActive, toVec4(th.FrameActive))
	defer imgui.PopStyleColorV(6)

	imgui.PushStyleVarVec2(imgui.StyleVarFramePadding, imgui.Vec2{X: paddingX, Y: paddingY})
	imgui.PushStyleVarFloat(imgui.StyleVarFrameBorderSize, 1.5)
	imgui.PushStyleVarFloat(imgui.StyleVarFrameRounding, 6)
	defer imgui.PopStyleVarV(3)

	if p.Label != "" {
		imgui.AlignTextToFramePadding()
		imgui.Text(p.Label)
		imgui.SameLine()
	}

	avail := imgui.ContentRegionAvail()
	var inputW float32
	if p.Width > 0 {
		inputW = float32(ctx.S(p.Width))
	} else {
		inputW = avail.X
	}
	if inputW < 40 {
		inputW = 40
	}

	imgui.SetNextItemWidth(inputW)
	if imgui.InputTextWithHint("##input_"+p.Label, p.Hint, buf, imgui.InputTextFlags(0), nil) && p.OnChange != nil {
		p.OnChange(*buf)
	}
}
