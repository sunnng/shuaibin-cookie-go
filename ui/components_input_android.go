//go:build android && cgo

package ui

import (
	"strconv"

	"github.com/Dasongzi1366/AutoGo/imgui"
)

// Switch 开关组件（design-system.md §4.3）：Bool 字段的积木化渲染，替换复选框。
// 56×32 全圆角槽（3px 墨描边），on=绿底滑块右移 / off=白底滑块居左，150ms 滑动。
// 勾选状态读取 p.Checked，切换时调用 p.OnChange(!checked)。热区纵向扩到 ≥48。
//
// 注意：ID 与 Button 同约（"##switch_"+Label）；同一窗口内多个 Label 相同（或同为空）
// 的开关会产生 ID 碰撞，调用方应自行使用 ctx.Push 隔离（组件状态同理）。
func Switch(ctx *Ctx, p CheckboxProps) {
	th := ctx.theme()

	if p.Label != "" {
		imgui.AlignTextToFramePadding()
		imgui.PushStyleColorVec4(imgui.ColText, toVec4(th.Text))
		imgui.Text(p.Label)
		imgui.PopStyleColorV(1)
		imgui.SameLine()
	}

	pos, size := switchVisualRect(ctx)
	anim := State(ctx, "swanim", switchAnimTarget(p.Checked))

	hitPos, hitSize := switchHitRect(ctx, pos, size)
	imgui.SetCursorScreenPos(hitPos)
	clicked := imgui.InvisibleButton("##switch_"+p.Label, hitSize)

	stepSwitchAnim(anim, p.Checked)
	drawSwitchVisual(imgui.WindowDrawList(), pos, size, *anim)

	if clicked && p.OnChange != nil {
		p.OnChange(!p.Checked)
	}
}

// switchSize 开关可视尺寸（基准 56×32，按 ctx.S 缩放）。
func switchSize(ctx *Ctx) imgui.Vec2 {
	return imgui.Vec2{X: float32(ctx.S(56)), Y: float32(ctx.S(32))}
}

// switchVisualRect 当前光标处的开关可视矩形；光标前进一个开关位（含热区高度）。
func switchVisualRect(ctx *Ctx) (pos, size imgui.Vec2) {
	size = switchSize(ctx)
	cursor := imgui.CursorScreenPos()
	// 热区 ≥48 高时可视槽垂直居中于热区。
	_, hitSize := switchHitRect(ctx, cursor, size)
	pos = imgui.Vec2{X: cursor.X, Y: cursor.Y + (hitSize.Y-size.Y)/2}
	return pos, size
}

// switchHitRect 开关热区：横向同可视宽，纵向不足 48 时扩到 48（触控目标下限）。
func switchHitRect(ctx *Ctx, pos, size imgui.Vec2) (hitPos, hitSize imgui.Vec2) {
	hitSize = size
	if min := float32(ctx.S(48)); hitSize.Y < min {
		hitSize.Y = min
	}
	hitPos = imgui.Vec2{X: pos.X, Y: pos.Y + (size.Y-hitSize.Y)/2}
	return hitPos, hitSize
}

// switchAnimTarget 开关动画目标位：on=1（滑块居右）/ off=0（居左）。
func switchAnimTarget(on bool) float32 {
	if on {
		return 1
	}
	return 0
}

// stepSwitchAnim 每帧把动画值向目标推进一步（≈150ms 完成，60fps）。
func stepSwitchAnim(anim *float32, on bool) {
	target := switchAnimTarget(on)
	const step = float32(0.12)
	if *anim < target {
		*anim += step
		if *anim > target {
			*anim = target
		}
	} else if *anim > target {
		*anim -= step
		if *anim < target {
			*anim = target
		}
	}
}

// drawSwitchVisual 绘制开关：白→绿渐变槽 + 3px 墨描边 + 墨色滑块（t=0 居左，1 居右）。
func drawSwitchVisual(dl *imgui.DrawList, pos, size imgui.Vec2, t float32) {
	radius := size.Y / 2
	pMax := imgui.Vec2{X: pos.X + size.X, Y: pos.Y + size.Y}
	dl.AddRectFilledV(pos, pMax, imgui.ColorU32Vec4(candyLerpColor(candyRaised, candyGreen, t)), radius, imgui.DrawFlagsRoundCornersAll)
	dl.AddRectV(pos, pMax, imgui.ColorU32Vec4(candyInk), radius, imgui.DrawFlagsRoundCornersAll, 3)

	knobD := size.Y * 22 / 32
	inset := (size.Y - knobD) / 2
	knobX := pos.X + inset + (size.X-knobD-inset*2)*t
	dl.AddCircleFilled(
		imgui.Vec2{X: knobX + knobD/2, Y: pos.Y + size.Y/2},
		knobD/2, imgui.ColorU32Vec4(candyInk),
	)
}

// Checkbox 复选框组件（ADR-0003）。标签在左，勾选框紧跟右侧；勾选状态读取
// p.Checked，切换时调用 p.OnChange(checked)。若 p.Label 为空则只显示勾选框。
//
// 注意：ID 与 Button 同约（"##chk_"+Label）；同一窗口内多个 Label 相同（或同为空）
// 的复选框会产生 ID 碰撞，调用方应自行使用 ctx.Push 隔离（组件状态同理）。
//
// 表单（Form）与任务卡片已改用 Switch；本组件保留给需要复选框语义的调用方。
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
	imgui.PushStyleVarFloat(imgui.StyleVarFrameBorderSize, 3)
	imgui.PushStyleVarFloat(imgui.StyleVarFrameRounding, 10)
	imgui.PushStyleColorVec4(imgui.ColFrameBg, toVec4(th.FrameBg))
	imgui.PushStyleColorVec4(imgui.ColFrameBgHovered, toVec4(th.FrameHover))
	imgui.PushStyleColorVec4(imgui.ColFrameBgActive, toVec4(th.FrameActive))
	imgui.PushStyleColorVec4(imgui.ColCheckMark, candyInk)
	imgui.PushStyleColorVec4(imgui.ColBorder, toVec4(th.Border))

	checked := p.Checked
	if imgui.Checkbox("##chk_"+p.Label, &checked) && p.OnChange != nil {
		p.OnChange(checked)
	}

	imgui.PopStyleColorV(5)
	imgui.PopStyleVarV(3)
}

// NumberInput 步进数字输入组件（ADR-0003）。左侧显示 p.Label，右侧为
// [-] [输入] [+] 三段式步进控件（按钮 44×44 积木块，数值框 72×44 白底 3px
// 墨描边）。数值读取 p.Value，p.Step<=0 时视为 1；按钮或编辑提交后先经
// p.Clamp 钳制再调用 p.OnChange。
//
// 输入框使用组件状态缓存编辑字符串（State(ctx, "numbuf", "")），初始时从
// p.Value 格式化。v1 不回写外部 p.Value 的变更；且缓冲为空即重填，用户无法
// 清空草稿重新输入（全选删除会在下一帧被还原，继承移植源行为）。
//
// 注意：缓冲键只按 Ctx 路径寻址；同一路径下渲染多个 NumberInput 会共享同一
// 缓冲（互相改写），调用方必须 ctx.Push 为每个实例隔离路径。ID 与 Button 同约
// （"##num_/##sub_/##add_"+Label），同 Label 实例同理会碰撞。
func NumberInput(ctx *Ctx, p NumberInputProps) {
	th := ctx.theme()

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

	btnS := float32(ctx.S(44))
	gap := float32(ctx.S(8))
	inputW := float32(ctx.S(72))
	if p.Width > 0 {
		if w := float32(ctx.S(p.Width)) - btnS*2 - gap*2; w > inputW {
			inputW = w
		}
	}

	buf := State(ctx, "numbuf", "")
	if *buf == "" {
		*buf = strconv.FormatFloat(p.Value, 'f', -1, 64)
	}

	rowPos := imgui.CursorScreenPos()

	// 先用 Dummy 登记整组控件边界（imgui 不允许用 SetCursorPos 扩展父窗口
	// 边界，须先提交 item 撑开），组内的绝对定位都在已登记范围内；
	// Dummy 同时把光标推进到下一行。
	imgui.Dummy(imgui.Vec2{X: btnS*2 + gap*2 + inputW, Y: btnS})

	// − 按钮（积木块：3px 描边 + 硬阴影 + 按压位移）。
	if candyButton("-##sub_"+p.Label, rowPos, imgui.Vec2{X: btnS, Y: btnS}, candyRaised, float32(ctx.S(10)), false,
		func(dl *imgui.DrawList, pMin, pMax imgui.Vec2, pressed bool) {
			candyLabelInRect(dl, pMin, pMax, "-", candyInk)
		}) && p.OnChange != nil {
		v := p.Clamp(p.Value - step)
		*buf = strconv.FormatFloat(v, 'f', -1, 64)
		p.OnChange(v)
	}

	// 数值框：白底 3px 墨描边，高度与按钮齐平。
	fontH := imgui.CalcTextSize("A").Y
	padY := (btnS - fontH) / 2
	if padY < 0 {
		padY = 0
	}
	imgui.PushStyleVarVec2(imgui.StyleVarFramePadding, imgui.Vec2{X: float32(ctx.S(8)), Y: padY})
	imgui.PushStyleVarFloat(imgui.StyleVarFrameBorderSize, 3)
	imgui.PushStyleVarFloat(imgui.StyleVarFrameRounding, float32(ctx.S(10)))
	imgui.PushStyleColorVec4(imgui.ColFrameBg, toVec4(th.FrameBg))
	imgui.PushStyleColorVec4(imgui.ColBorder, toVec4(th.Border))
	imgui.PushStyleColorVec4(imgui.ColText, toVec4(th.Text))

	inputPos := imgui.Vec2{X: rowPos.X + btnS + gap, Y: rowPos.Y}
	imgui.SetCursorScreenPos(inputPos)
	imgui.SetNextItemWidth(inputW)
	if imgui.InputTextWithHint("##num_"+p.Label, p.Hint, buf, imgui.InputTextFlagsCharsDecimal, nil) && p.OnChange != nil {
		if n, err := strconv.ParseFloat(*buf, 64); err == nil {
			v := p.Clamp(n)
			*buf = strconv.FormatFloat(v, 'f', -1, 64)
			p.OnChange(v)
		}
	}

	imgui.PopStyleColorV(3)
	imgui.PopStyleVarV(3)

	// + 按钮。
	plusPos := imgui.Vec2{X: inputPos.X + inputW + gap, Y: rowPos.Y}
	if candyButton("+##add_"+p.Label, plusPos, imgui.Vec2{X: btnS, Y: btnS}, candyRaised, float32(ctx.S(10)), false,
		func(dl *imgui.DrawList, pMin, pMax imgui.Vec2, pressed bool) {
			candyLabelInRect(dl, pMin, pMax, "+", candyInk)
		}) && p.OnChange != nil {
		v := p.Clamp(p.Value + step)
		*buf = strconv.FormatFloat(v, 'f', -1, 64)
		p.OnChange(v)
	}
}

// TextInput 文本输入组件（ADR-0003）。p.Multiline 为 true 时渲染多行输入框，
// 否则单行。p.Label 非空时显示为标签（单行在左侧，多行在上方）。p.Hint 为
// 占位提示。宽度 p.Width>0 时经 ctx.S 缩放，否则占满剩余宽度。
// 视觉（design-system.md §4.3）：白底 3px 墨描边积木框（高 44），聚焦时
// 4px 糖果黄外发光。
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

	buf := State(ctx, "inbuf", p.Value)

	fontH := imgui.CalcTextSize("A").Y
	frameH := float32(ctx.S(44))
	padY := (frameH - fontH) / 2
	if padY < 0 {
		padY = 0
	}

	imgui.PushStyleColorVec4(imgui.ColFrameBg, toVec4(th.FrameBg))
	imgui.PushStyleColorVec4(imgui.ColText, toVec4(th.Text))
	imgui.PushStyleColorVec4(imgui.ColTextDisabled, toVec4(th.TextDisabled))
	imgui.PushStyleColorVec4(imgui.ColBorder, toVec4(th.Border))
	imgui.PushStyleColorVec4(imgui.ColFrameBgHovered, toVec4(th.FrameBg))
	imgui.PushStyleColorVec4(imgui.ColFrameBgActive, toVec4(th.FrameBg))
	defer imgui.PopStyleColorV(6)

	imgui.PushStyleVarVec2(imgui.StyleVarFramePadding, imgui.Vec2{X: float32(ctx.S(12)), Y: padY})
	imgui.PushStyleVarFloat(imgui.StyleVarFrameBorderSize, 3)
	imgui.PushStyleVarFloat(imgui.StyleVarFrameRounding, float32(ctx.S(10)))
	defer imgui.PopStyleVarV(3)

	if p.Multiline {
		if p.Label != "" {
			imgui.PushStyleColorVec4(imgui.ColText, toVec4(th.Text))
			imgui.Text(p.Label)
			imgui.PopStyleColorV(1)
		}

		inputW := textInputWidth(ctx, p.Width)
		const multilineH = float32(80)
		if imgui.InputTextMultiline("##multi_"+p.Label, buf, imgui.Vec2{X: inputW, Y: multilineH}, imgui.InputTextFlags(0), nil) && p.OnChange != nil {
			p.OnChange(*buf)
		}
		drawFocusGlow(ctx)
		return
	}

	if p.Label != "" {
		imgui.AlignTextToFramePadding()
		imgui.Text(p.Label)
		imgui.SameLine()
	}

	inputW := textInputWidth(ctx, p.Width)
	imgui.SetNextItemWidth(inputW)
	if imgui.InputTextWithHint("##input_"+p.Label, p.Hint, buf, imgui.InputTextFlags(0), nil) && p.OnChange != nil {
		p.OnChange(*buf)
	}
	drawFocusGlow(ctx)
}

// textInputWidth 输入框宽度：p.Width>0 经 ctx.S 缩放，否则占满剩余宽度（下限 40）。
func textInputWidth(ctx *Ctx, width float64) float32 {
	var inputW float32
	if width > 0 {
		inputW = float32(ctx.S(width))
	} else {
		inputW = imgui.ContentRegionAvail().X
	}
	if inputW < 40 {
		inputW = 40
	}
	return inputW
}

// drawFocusGlow 聚焦光环：输入框激活时沿外缘描 4px 糖果黄（§4.3 聚焦外发光）。
func drawFocusGlow(ctx *Ctx) {
	if !imgui.IsItemActive() {
		return
	}
	dl := imgui.WindowDrawList()
	pMin := imgui.ItemRectMin()
	pMax := imgui.ItemRectMax()
	const expand = float32(2)
	dl.AddRectV(
		imgui.Vec2{X: pMin.X - expand, Y: pMin.Y - expand},
		imgui.Vec2{X: pMax.X + expand, Y: pMax.Y + expand},
		imgui.ColorU32Vec4(candyYellow), float32(ctx.S(12)), imgui.DrawFlagsRoundCornersAll, 4,
	)
}
