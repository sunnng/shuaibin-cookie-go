//go:build android && cgo

package ui

import "github.com/Dasongzi1366/AutoGo/imgui"

// 糖果积木（Candy Blocks）共享绘制助手（docs/ui-redesign/design-system.md §3/§7.4）：
// 糖果色板、硬阴影积木块、按压反馈。面板侧颜色尽量走 Theme 令牌；本文件的
// 糖果色用于 Theme 未覆盖的语义色（状态/分类/危险），灵动岛与面板共用。

var (
	candyInk        = imgui.Vec4{X: 0x21 / 255.0, Y: 0x1D / 255.0, Z: 0x19 / 255.0, W: 1} // #211D19
	candyPaper      = imgui.Vec4{X: 0xFF / 255.0, Y: 0xF6 / 255.0, Z: 0xE5 / 255.0, W: 1} // #FFF6E5
	candyRaised     = imgui.Vec4{X: 1, Y: 1, Z: 1, W: 1}                                  // #FFFFFF
	candyYellow     = imgui.Vec4{X: 0xFF / 255.0, Y: 0xC9 / 255.0, Z: 0x3C / 255.0, W: 1} // #FFC93C
	candyYellowSoft = imgui.Vec4{X: 0xFF / 255.0, Y: 0xE9 / 255.0, Z: 0xB8 / 255.0, W: 1} // #FFE9B8
	candyGreen      = imgui.Vec4{X: 0x06 / 255.0, Y: 0xD6 / 255.0, Z: 0xA0 / 255.0, W: 1} // #06D6A0
	candyBlue       = imgui.Vec4{X: 0x4D / 255.0, Y: 0x96 / 255.0, Z: 0xFF / 255.0, W: 1} // #4D96FF
	candyOrange     = imgui.Vec4{X: 0xFF / 255.0, Y: 0x9F / 255.0, Z: 0x1C / 255.0, W: 1} // #FF9F1C
	candyRed        = imgui.Vec4{X: 0xEF / 255.0, Y: 0x47 / 255.0, Z: 0x6F / 255.0, W: 1} // #EF476F
	candyPurple     = imgui.Vec4{X: 0x9B / 255.0, Y: 0x5D / 255.0, Z: 0xE5 / 255.0, W: 1} // #9B5DE5
	candySec        = imgui.Vec4{X: 0x8A / 255.0, Y: 0x80 / 255.0, Z: 0x71 / 255.0, W: 1} // #8A8071
	candyPillOff    = imgui.Vec4{X: 0xE9 / 255.0, Y: 0xE2 / 255.0, Z: 0xD4 / 255.0, W: 1} // #E9E2D4 未启用胶囊
	candyDashLine   = imgui.Vec4{X: 0xE3 / 255.0, Y: 0xD5 / 255.0, Z: 0xB8 / 255.0, W: 1} // #E3D5B8 表单虚线
	candyDangerBg   = imgui.Vec4{X: 0xFD / 255.0, Y: 0xE8 / 255.0, Z: 0xEE / 255.0, W: 1} // #FDE8EE 危险区底
	candyOKBg       = imgui.Vec4{X: 0xD9 / 255.0, Y: 0xF7 / 255.0, Z: 0xEE / 255.0, W: 1} // #D9F7EE 反馈成功底
	candyErrBg      = imgui.Vec4{X: 0xFB / 255.0, Y: 0xD9 / 255.0, Z: 0xE2 / 255.0, W: 1} // #FBD9E2 反馈失败底
	candySelCardBg  = imgui.Vec4{X: 0xFF / 255.0, Y: 0xFD / 255.0, Z: 0xF6 / 255.0, W: 1} // #FFFDF6 选中任务卡

	// 面板内状态文字的深变体（纸面上 ≥4.5:1）。
	candyRunText   = imgui.Vec4{X: 0x0E / 255.0, Y: 0x9F / 255.0, Z: 0x6E / 255.0, W: 1} // #0E9F6E
	candyPauseText = imgui.Vec4{X: 0xC2 / 255.0, Y: 0x41 / 255.0, Z: 0x0C / 255.0, W: 1} // #C2410C
	candyIdleText  = imgui.Vec4{X: 0x25 / 255.0, Y: 0x63 / 255.0, Z: 0xEB / 255.0, W: 1} // #2563EB
)

// candyStateColor 脚本状态对应的糖果色（灵动岛描边/状态点共用）：
// 空闲蓝 / 运行绿 / 暂停橙。
func candyStateColor(state ScriptState) imgui.Vec4 {
	switch state {
	case StateRunning:
		return candyGreen
	case StatePaused:
		return candyOrange
	default:
		return candyBlue
	}
}

// candyCategoryColor 任务分类色条：日常黄 / 活动紫 / 维护蓝 / 未知灰。
func candyCategoryColor(cat string) imgui.Vec4 {
	switch cat {
	case "daily":
		return candyYellow
	case "event":
		return candyPurple
	case "maint":
		return candyBlue
	default:
		return candySec
	}
}

// drawBlock 画一个糖果积木块：先画偏移 (shadowOff,shadowOff) 的墨色实心矩形
// （硬阴影），再在其上画本色矩形 + borderW 描边。shadowOff<=0 时不画阴影。
// fill.A<1 时阴影一并省略（半透明块叠硬阴影会糊）。
func drawBlock(dl *imgui.DrawList, pMin, pMax imgui.Vec2, fill imgui.Vec4, radius, borderW, shadowOff float32) {
	if shadowOff > 0 && fill.W >= 1 {
		dl.AddRectFilledV(
			imgui.Vec2{X: pMin.X + shadowOff, Y: pMin.Y + shadowOff},
			imgui.Vec2{X: pMax.X + shadowOff, Y: pMax.Y + shadowOff},
			imgui.ColorU32Vec4(candyInk), radius, imgui.DrawFlagsRoundCornersAll,
		)
	}
	dl.AddRectFilledV(pMin, pMax, imgui.ColorU32Vec4(fill), radius, imgui.DrawFlagsRoundCornersAll)
	if borderW > 0 {
		dl.AddRectV(pMin, pMax, imgui.ColorU32Vec4(candyInk), radius, imgui.DrawFlagsRoundCornersAll, borderW)
	}
}

// candyButton 糖果积木按钮：3px 墨描边 + 4px 硬阴影，按压时整体 +3,+3 位移且
// 阴影收敛为 1px（"积木被按进桌面"，§3.3）。热区即 size（调用方保证 ≥48）。
// disabled 时 40% 透明、无阴影、不响应。返回本帧是否被点击。
func candyButton(id string, pos imgui.Vec2, size imgui.Vec2, fill imgui.Vec4, radius float32, disabled bool, draw func(dl *imgui.DrawList, pMin, pMax imgui.Vec2, pressed bool)) bool {
	imgui.SetCursorScreenPos(pos)
	if disabled {
		imgui.BeginDisabled()
	}
	clicked := imgui.InvisibleButton(id, size)
	hovered := imgui.IsItemHovered()
	pressed := imgui.IsItemActive()
	if disabled {
		imgui.EndDisabled()
	}

	dl := imgui.WindowDrawList()
	off := float32(0)
	shadow := float32(4)
	if pressed {
		off = 3
		shadow = 1
	}
	body := fill
	if hovered && !pressed {
		body = candyHoverFill(fill)
	}
	if disabled {
		body.W *= 0.4
		shadow = 0
	}
	pMin := imgui.Vec2{X: pos.X + off, Y: pos.Y + off}
	pMax := imgui.Vec2{X: pos.X + size.X + off, Y: pos.Y + size.Y + off}
	drawBlock(dl, pMin, pMax, body, radius, 3, shadow)
	if draw != nil {
		draw(dl, pMin, pMax, pressed)
	}
	return clicked
}

// candyHoverFill hover 微亮：白/纸面块 hover 变浅黄，彩色块保持本色
// （彩色已有足够辨识度，避免 hover 抖动）。
func candyHoverFill(fill imgui.Vec4) imgui.Vec4 {
	if fill == candyRaised || fill == candyPaper {
		return candyYellowSoft
	}
	return fill
}

// candyLerpColor 两色按 t∈[0,1] 线性插值（开关槽底色渐变用）。
func candyLerpColor(a, b imgui.Vec4, t float32) imgui.Vec4 {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	return imgui.Vec4{
		X: a.X + (b.X-a.X)*t,
		Y: a.Y + (b.Y-a.Y)*t,
		Z: a.Z + (b.Z-a.Z)*t,
		W: a.W + (b.W-a.W)*t,
	}
}

// candyLabelInRect 在积木块内居中绘制文字（墨色）。
func candyLabelInRect(dl *imgui.DrawList, pMin, pMax imgui.Vec2, label string, col imgui.Vec4) {
	sz := measureLabelSize(label)
	dl.AddTextVec2V(
		imgui.Vec2{X: pMin.X + (pMax.X-pMin.X-sz.X)/2, Y: pMin.Y + (pMax.Y-pMin.Y-sz.Y)/2},
		imgui.ColorU32Vec4(col),
		label,
	)
}

// drawDashedHLine 表单行分隔虚线（2px 间隔的浅棕虚线，原型 border-bottom dashed）。
func drawDashedHLine(dl *imgui.DrawList, x1, x2, y float32) {
	const seg, gap = float32(6), float32(5)
	for x := x1; x+seg <= x2; x += seg + gap {
		dl.AddLineV(imgui.Vec2{X: x, Y: y}, imgui.Vec2{X: x + seg, Y: y}, imgui.ColorU32Vec4(candyDashLine), 2)
	}
}

// drawFittedText 按 scale 缩放字号绘制文本（design §3.2 字阶：摘要/标签 13px
// ≈ 基准字号 × 0.82）；文本估算宽度超过 maxW 时按字截断并追加省略号，
// 保证不会溢出所在容器。
func drawFittedText(dl *imgui.DrawList, pos imgui.Vec2, col imgui.Vec4, text string, maxW, scale float32) {
	if scale <= 0 {
		scale = 1
	}
	dl.AddTextFontPtr(
		imgui.CurrentFont(), imgui.FontSize()*scale, pos,
		imgui.ColorU32Vec4(col), fitTextToWidth(text, maxW, scale),
	)
}

// fitTextToWidth 文本宽度（按 scale 线性估算）超过 maxW 时按字截断并加省略号。
func fitTextToWidth(text string, maxW, scale float32) string {
	if maxW <= 0 {
		return ""
	}
	if measureLabelSize(text).X*scale <= maxW {
		return text
	}
	r := []rune(text)
	for len(r) > 0 {
		r = r[:len(r)-1]
		if measureLabelSize(string(r)+"…").X*scale <= maxW {
			return string(r) + "…"
		}
	}
	return "…"
}
