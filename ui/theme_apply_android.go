//go:build android && cgo

package ui

import "github.com/Dasongzi1366/AutoGo/imgui"

// toVec4 把框架 Color 换算为 imgui.Vec4。
func toVec4(c Color) imgui.Vec4 {
	return imgui.Vec4{X: c.R, Y: c.G, Z: c.B, W: c.A}
}

// ApplyTheme 把主题令牌推入当前 ImGui 样式（在 imgui.Init 之后调用一次；
// 颜色/圆角在整个 RunShell 生命周期内保持）。
func ApplyTheme(t Theme) {
	imgui.PushStyleColorVec4(imgui.ColWindowBg, toVec4(t.WindowBg))
	imgui.PushStyleColorVec4(imgui.ColChildBg, toVec4(t.ChildBg))
	imgui.PushStyleColorVec4(imgui.ColPopupBg, toVec4(t.PopupBg))
	imgui.PushStyleColorVec4(imgui.ColBorder, toVec4(t.Border))
	imgui.PushStyleColorVec4(imgui.ColFrameBg, toVec4(t.FrameBg))
	imgui.PushStyleColorVec4(imgui.ColFrameBgHovered, toVec4(t.FrameHover))
	imgui.PushStyleColorVec4(imgui.ColFrameBgActive, toVec4(t.FrameActive))
	imgui.PushStyleColorVec4(imgui.ColButton, toVec4(t.Button))
	imgui.PushStyleColorVec4(imgui.ColButtonHovered, toVec4(t.ButtonHover))
	imgui.PushStyleColorVec4(imgui.ColButtonActive, toVec4(t.ButtonActive))
	imgui.PushStyleColorVec4(imgui.ColHeader, toVec4(t.Header))
	imgui.PushStyleColorVec4(imgui.ColHeaderHovered, toVec4(t.HeaderHover))
	imgui.PushStyleColorVec4(imgui.ColHeaderActive, toVec4(t.HeaderActive))
	imgui.PushStyleColorVec4(imgui.ColText, toVec4(t.Text))
	imgui.PushStyleColorVec4(imgui.ColTextDisabled, toVec4(t.TextDisabled))
	imgui.PushStyleColorVec4(imgui.ColTitleBg, toVec4(t.TitleBg))
	imgui.PushStyleColorVec4(imgui.ColTitleBgActive, toVec4(t.TitleBgActive))
	imgui.PushStyleColorVec4(imgui.ColCheckMark, toVec4(t.Accent))
	imgui.PushStyleColorVec4(imgui.ColSliderGrab, toVec4(t.Accent))
	imgui.PushStyleColorVec4(imgui.ColSliderGrabActive, toVec4(t.Accent))

	imgui.PushStyleVarFloat(imgui.StyleVarWindowRounding, t.Rounding)
	imgui.PushStyleVarFloat(imgui.StyleVarChildRounding, t.Rounding)
	imgui.PushStyleVarFloat(imgui.StyleVarFrameRounding, t.Rounding)
	imgui.PushStyleVarFloat(imgui.StyleVarGrabRounding, t.Rounding)
	imgui.PushStyleVarFloat(imgui.StyleVarPopupRounding, t.Rounding)
	imgui.PushStyleVarFloat(imgui.StyleVarScrollbarRounding, t.Rounding)
	imgui.PushStyleVarFloat(imgui.StyleVarWindowBorderSize, 1)
	imgui.PushStyleVarFloat(imgui.StyleVarChildBorderSize, 1)
	imgui.PushStyleVarFloat(imgui.StyleVarFrameBorderSize, 1)
}
