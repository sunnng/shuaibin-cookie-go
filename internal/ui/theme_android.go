//go:build android && cgo

package ui

import "github.com/Dasongzi1366/AutoGo/imgui"

// IndustrialTheme 工业软件严谨风色板（不透明、小圆角、实线边框）。
type IndustrialTheme struct {
	WindowBg      imgui.Vec4
	ChildBg       imgui.Vec4
	PopupBg       imgui.Vec4
	Border        imgui.Vec4
	FrameBg       imgui.Vec4
	FrameHover    imgui.Vec4
	FrameActive   imgui.Vec4
	Button        imgui.Vec4
	ButtonHover   imgui.Vec4
	ButtonActive  imgui.Vec4
	Header        imgui.Vec4
	HeaderHover   imgui.Vec4
	HeaderActive  imgui.Vec4
	Text          imgui.Vec4
	TextDisabled  imgui.Vec4
	Accent        imgui.Vec4
	TitleBg       imgui.Vec4
	TitleBgActive imgui.Vec4
	Rounding      float32
}

// DefaultIndustrialTheme 深灰工控面板默认主题。
func DefaultIndustrialTheme() IndustrialTheme {
	return IndustrialTheme{
		WindowBg:      HexToVec4("#1b1e22ff"),
		ChildBg:       HexToVec4("#22262bff"),
		PopupBg:       HexToVec4("#1b1e22f2"),
		Border:        HexToVec4("#3d4450ff"),
		FrameBg:       HexToVec4("#2a2f36ff"),
		FrameHover:    HexToVec4("#343b44ff"),
		FrameActive:   HexToVec4("#3a4452ff"),
		Button:        HexToVec4("#2f3640ff"),
		ButtonHover:   HexToVec4("#3a4452ff"),
		ButtonActive:  HexToVec4("#4a90c8ff"),
		Header:        HexToVec4("#2a2f36ff"),
		HeaderHover:   HexToVec4("#343b44ff"),
		HeaderActive:  HexToVec4("#3a4452ff"),
		Text:          HexToVec4("#d6d8dfff"),
		TextDisabled:  HexToVec4("#7a828cff"),
		Accent:        HexToVec4("#4a90c8ff"),
		TitleBg:       HexToVec4("#16191dff"),
		TitleBgActive: HexToVec4("#16191dff"),
		Rounding:      2,
	}
}

// ApplyIndustrialTheme 将工业风推入当前 ImGui 样式（在 imgui.Init 之后调用）。
// 颜色/圆角在整个 RunShell 生命周期内保持（进程级 UI）。
func ApplyIndustrialTheme() {
	th := DefaultIndustrialTheme()

	imgui.PushStyleColorVec4(imgui.ColWindowBg, th.WindowBg)
	imgui.PushStyleColorVec4(imgui.ColChildBg, th.ChildBg)
	imgui.PushStyleColorVec4(imgui.ColPopupBg, th.PopupBg)
	imgui.PushStyleColorVec4(imgui.ColBorder, th.Border)
	imgui.PushStyleColorVec4(imgui.ColFrameBg, th.FrameBg)
	imgui.PushStyleColorVec4(imgui.ColFrameBgHovered, th.FrameHover)
	imgui.PushStyleColorVec4(imgui.ColFrameBgActive, th.FrameActive)
	imgui.PushStyleColorVec4(imgui.ColButton, th.Button)
	imgui.PushStyleColorVec4(imgui.ColButtonHovered, th.ButtonHover)
	imgui.PushStyleColorVec4(imgui.ColButtonActive, th.ButtonActive)
	imgui.PushStyleColorVec4(imgui.ColHeader, th.Header)
	imgui.PushStyleColorVec4(imgui.ColHeaderHovered, th.HeaderHover)
	imgui.PushStyleColorVec4(imgui.ColHeaderActive, th.HeaderActive)
	imgui.PushStyleColorVec4(imgui.ColText, th.Text)
	imgui.PushStyleColorVec4(imgui.ColTextDisabled, th.TextDisabled)
	imgui.PushStyleColorVec4(imgui.ColTitleBg, th.TitleBg)
	imgui.PushStyleColorVec4(imgui.ColTitleBgActive, th.TitleBgActive)
	imgui.PushStyleColorVec4(imgui.ColCheckMark, th.Accent)
	imgui.PushStyleColorVec4(imgui.ColSliderGrab, th.Accent)
	imgui.PushStyleColorVec4(imgui.ColSliderGrabActive, th.Accent)

	imgui.PushStyleVarFloat(imgui.StyleVarWindowRounding, th.Rounding)
	imgui.PushStyleVarFloat(imgui.StyleVarChildRounding, th.Rounding)
	imgui.PushStyleVarFloat(imgui.StyleVarFrameRounding, th.Rounding)
	imgui.PushStyleVarFloat(imgui.StyleVarGrabRounding, th.Rounding)
	imgui.PushStyleVarFloat(imgui.StyleVarPopupRounding, th.Rounding)
	imgui.PushStyleVarFloat(imgui.StyleVarScrollbarRounding, th.Rounding)
	imgui.PushStyleVarFloat(imgui.StyleVarWindowBorderSize, 1)
	imgui.PushStyleVarFloat(imgui.StyleVarChildBorderSize, 1)
	imgui.PushStyleVarFloat(imgui.StyleVarFrameBorderSize, 1)
}

// IndustrialAccent 主色，供顶栏 START 等使用。
func IndustrialAccent() imgui.Vec4 {
	return HexToVec4("#4a90c8ff")
}
