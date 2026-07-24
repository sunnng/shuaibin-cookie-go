//go:build android && cgo

package ui

import "github.com/Dasongzi1366/AutoGo/imgui"

// Button 按钮组件（ADR-0003）。尺寸语义：Width/Height 为基准分辨率尺寸，
// 经 ctx.S 换算；<=0 表示按文字自适应。Kind 决定配色（design-system.md §3）：
// Primary=糖果黄底墨字；Secondary=白底墨边墨字；Danger=糖果红底纸字。
// 全部为 3px 墨描边积木块 + 4px 硬阴影，按压 +3,+3 位移且阴影收敛为 1px；
// Disabled 40% 透明、无阴影。
//
// 注意：Label+"##btn" 以 Label 作为 ID 区分依据；同一窗口内多个 Label 相同的按钮
// 会产生 ID 碰撞，调用方应自行使用 ctx.Push 隔离（组件状态同理）。
func Button(ctx *Ctx, p ButtonProps) {
	th := ctx.theme()
	padX, padY := float32(ctx.S(16)), float32(ctx.S(12))

	w, h := float32(p.Width), float32(p.Height)
	if w > 0 {
		w = float32(ctx.S(p.Width))
	}
	if h > 0 {
		h = float32(ctx.S(p.Height))
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

	fill := candyRaised
	text := candyInk
	switch p.Kind {
	case ButtonPrimary:
		fill = toVec4(th.Accent) // 糖果黄（Accent），配墨字
	case ButtonDanger:
		fill = candyRed
		text = candyPaper
	}

	pos := imgui.CursorScreenPos()
	size := imgui.Vec2{X: w, Y: h}
	label := p.Label
	clicked := candyButton(label+"##btn", pos, size, fill, float32(ctx.S(12)), p.Disabled,
		func(dl *imgui.DrawList, pMin, pMax imgui.Vec2, pressed bool) {
			col := text
			if p.Disabled {
				col.W *= 0.4
			}
			candyLabelInRect(dl, pMin, pMax, label, col)
		})
	if clicked && p.OnClick != nil {
		p.OnClick()
	}
}
