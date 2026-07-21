package ui

import (
	"strconv"
	"strings"
)

// Color RGBA 颜色，分量 0..1。框架自有类型，避免纯逻辑层依赖 imgui；
// android 绘制层负责换算为 imgui.Vec4。
type Color struct {
	R, G, B, A float32
}

// Hex 解析 "#rrggbb" 或 "#rrggbbaa"；非法输入返回不透明黑。
func Hex(s string) Color {
	black := Color{0, 0, 0, 1}
	s = strings.TrimPrefix(s, "#")
	if len(s) != 6 && len(s) != 8 {
		return black
	}
	v, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return black
	}
	if len(s) == 6 {
		return Color{
			R: float32(v>>16&0xff) / 255,
			G: float32(v>>8&0xff) / 255,
			B: float32(v&0xff) / 255,
			A: 1,
		}
	}
	return Color{
		R: float32(v>>24&0xff) / 255,
		G: float32(v>>16&0xff) / 255,
		B: float32(v>>8&0xff) / 255,
		A: float32(v&0xff) / 255,
	}
}

// Theme 框架视觉令牌（CONTEXT.md「主题」）：整套替换或覆盖个别令牌，
// 无运行时切换。全部为可比较字段，零值表示「未指定」。
type Theme struct {
	WindowBg, ChildBg, PopupBg, Border   Color
	FrameBg, FrameHover, FrameActive     Color
	Button, ButtonHover, ButtonActive    Color
	Header, HeaderHover, HeaderActive    Color
	Text, TextDisabled, Accent           Color
	TitleBg, TitleBgActive               Color
	TitleTop, TitleBottom, RailBg, White Color
	Rounding                             float32
}

// DefaultTheme QQ 风浅蓝默认主题（沿用 internal/ui 的 QQ 蓝色值）。
func DefaultTheme() Theme {
	return Theme{
		WindowBg:      Hex("#e9f2fbff"),
		ChildBg:       Hex("#f7fbffff"),
		PopupBg:       Hex("#f2f8feff"),
		Border:        Hex("#9cc3e5ff"),
		FrameBg:       Hex("#ffffffff"),
		FrameHover:    Hex("#e3f0fbff"),
		FrameActive:   Hex("#cde6faff"),
		Button:        Hex("#dcebfaff"),
		ButtonHover:   Hex("#bcdcf7ff"),
		ButtonActive:  Hex("#8fc3efff"),
		Header:        Hex("#dcebfaff"),
		HeaderHover:   Hex("#bcdcf7ff"),
		HeaderActive:  Hex("#8fc3efff"),
		Text:          Hex("#1f3a52ff"),
		TextDisabled:  Hex("#7a8fa3ff"),
		Accent:        Hex("#2f8fd0ff"),
		TitleBg:       Hex("#3d8fd1ff"),
		TitleBgActive: Hex("#3d8fd1ff"),
		TitleTop:      Hex("#5aa9e6ff"),
		TitleBottom:   Hex("#2f7fc4ff"),
		RailBg:        Hex("#cfe4f7ff"),
		White:         Color{1, 1, 1, 1},
		Rounding:      4,
	}
}
