//go:build android && cgo

package ui

import "github.com/Dasongzi1366/AutoGo/imgui"

// Button 按钮组件（ADR-0003）。尺寸语义：Width/Height 为基准分辨率尺寸，
// 经 ctx.S 换算；<=0 表示按文字自适应。Kind 决定配色：
// Primary=主题 Accent 底白字；Secondary=液态玻璃（半透明白）；Danger=系统红底白字。
//
// 注意：Label+"##btn" 以 Label 作为 ID 区分依据；同一窗口内多个 Label 相同的按钮
// 会产生 ID 碰撞，调用方应自行使用 ctx.Push 隔离（组件状态同理）。
func Button(ctx *Ctx, p ButtonProps) {
	th := ctx.theme()
	const padX, padY = float32(16), float32(12)

	w, h := float32(p.Width), float32(p.Height)
	if w > 0 {
		w = ctx.S(p.Width)
	}
	if h > 0 {
		h = ctx.S(p.Height)
	}
	fitW, fitH := fitButtonSize(p.Label, padX, padY)
	if w <= 0 {
		w = fitW
		if w < 64 {
			w = 64
		}
	} else if w < fitW {
		w = fitW
	}
	if h <= 0 {
		h = fitH
	} else if h < fitH {
		h = fitH
	}

	imgui.PushStyleVarVec2(imgui.StyleVarFramePadding, imgui.Vec2{X: padX, Y: padY})
	imgui.PushStyleVarVec2(imgui.StyleVarButtonTextAlign, imgui.Vec2{X: 0.5, Y: 0.5})
	imgui.PushStyleVarFloat(imgui.StyleVarFrameBorderSize, 1)
	imgui.PushStyleVarFloat(imgui.StyleVarFrameRounding, 10)

	var bg, bgHover, bgActive, text imgui.Vec4
	switch p.Kind {
	case ButtonPrimary:
		bg, bgHover, bgActive = toVec4(th.Accent), toVec4(th.Accent), toVec4(th.TitleBottom)
		text = toVec4(th.White)
	case ButtonDanger:
		bg = imgui.Vec4{X: 0.85, Y: 0.16, Z: 0.22, W: 1}
		bgHover = imgui.Vec4{X: 0.95, Y: 0.29, Z: 0.33, W: 1}
		bgActive = imgui.Vec4{X: 0.75, Y: 0.10, Z: 0.16, W: 1}
		text = toVec4(th.White)
	default: // ButtonSecondary：液态玻璃
		bg = imgui.Vec4{X: 1, Y: 1, Z: 1, W: 0.10}
		bgHover = imgui.Vec4{X: 1, Y: 1, Z: 1, W: 0.20}
		bgActive = imgui.Vec4{X: 0.72, Y: 0.86, Z: 1, W: 0.35}
		text = toVec4(th.Text)
	}
	imgui.PushStyleColorVec4(imgui.ColButton, bg)
	imgui.PushStyleColorVec4(imgui.ColButtonHovered, bgHover)
	imgui.PushStyleColorVec4(imgui.ColButtonActive, bgActive)
	imgui.PushStyleColorVec4(imgui.ColBorder, imgui.Vec4{X: 1, Y: 1, Z: 1, W: 0.18})
	imgui.PushStyleColorVec4(imgui.ColText, text)

	if p.Disabled {
		imgui.BeginDisabled()
	}
	if imgui.ButtonV(p.Label+"##btn", imgui.Vec2{X: w, Y: h}) && p.OnClick != nil {
		p.OnClick()
	}
	if p.Disabled {
		imgui.EndDisabled()
	}

	imgui.PopStyleColorV(5)
	imgui.PopStyleVarV(4)
}
