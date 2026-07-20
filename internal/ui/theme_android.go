//go:build android && cgo

package ui

import "github.com/Dasongzi1366/AutoGo/imgui"

// QQBlueTheme 经典 QQ 风浅蓝主题：浅蓝窗体、白色内容区、天蓝主色、小圆角。
// 参考 2009 年代 QQ 客户端配色（sky-blue + white）。
type QQBlueTheme struct {
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

// DefaultQQBlueTheme QQ 风浅蓝默认主题。
func DefaultQQBlueTheme() QQBlueTheme {
	return QQBlueTheme{
		WindowBg:      HexToVec4("#e9f2fbff"),
		ChildBg:       HexToVec4("#f7fbffff"),
		PopupBg:       HexToVec4("#f2f8feff"),
		Border:        HexToVec4("#9cc3e5ff"),
		FrameBg:       HexToVec4("#ffffffff"),
		FrameHover:    HexToVec4("#e3f0fbff"),
		FrameActive:   HexToVec4("#cde6faff"),
		Button:        HexToVec4("#dcebfaff"),
		ButtonHover:   HexToVec4("#bcdcf7ff"),
		ButtonActive:  HexToVec4("#8fc3efff"),
		Header:        HexToVec4("#dcebfaff"),
		HeaderHover:   HexToVec4("#bcdcf7ff"),
		HeaderActive:  HexToVec4("#8fc3efff"),
		Text:          HexToVec4("#1f3a52ff"),
		TextDisabled:  HexToVec4("#7a8fa3ff"),
		Accent:        HexToVec4("#2f8fd0ff"),
		TitleBg:       HexToVec4("#3d8fd1ff"),
		TitleBgActive: HexToVec4("#3d8fd1ff"),
		Rounding:      4,
	}
}

// ApplyQQBlueTheme 将 QQ 蓝主题推入当前 ImGui 样式（在 imgui.Init 之后调用）。
// 颜色/圆角在整个 RunShell 生命周期内保持（进程级 UI）。
func ApplyQQBlueTheme() {
	th := DefaultQQBlueTheme()

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

// QQBlueAccent 主色（天蓝），供选中态、主按钮等使用。
func QQBlueAccent() imgui.Vec4 {
	return HexToVec4("#2f8fd0ff")
}

// 标题栏渐变端色（上浅下深的天蓝，QQ 客户端标题栏风格）。
func QQBlueTitleTop() imgui.Vec4    { return HexToVec4("#5aa9e6ff") }
func QQBlueTitleBottom() imgui.Vec4 { return HexToVec4("#2f7fc4ff") }

// QQBlueRailBg 面板左轨底色（浅蓝）。
func QQBlueRailBg() imgui.Vec4 { return HexToVec4("#cfe4f7ff") }

// QQBlueWhite 标题栏/主按钮上的文字白。
func QQBlueWhite() imgui.Vec4 { return imgui.Vec4{X: 1, Y: 1, Z: 1, W: 1} }
