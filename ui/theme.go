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

// DefaultTheme 糖果积木（Candy Blocks）默认主题：纸面底 + 墨色描边 + 糖果黄
// 主色（docs/ui-redesign/design-system.md §3/§7.1）。Accent 为糖果黄，压在
// 其上的文字必须用墨色 Text，不能用白字（黄底白字对比度不达标）。
func DefaultTheme() Theme {
	return Theme{
		WindowBg:      Hex("#FFF6E5ff"), // paper
		ChildBg:       Hex("#FFFBF2ff"),
		PopupBg:       Hex("#FFF3D6ff"),
		Border:        Hex("#211D19ff"), // ink
		FrameBg:       Hex("#ffffffff"),
		FrameHover:    Hex("#FFE9B8ff"), // 浅黄 hover
		FrameActive:   Hex("#FFC93Cff"), // candy-yellow
		Button:        Hex("#ffffffff"),
		ButtonHover:   Hex("#FFE9B8ff"),
		ButtonActive:  Hex("#FFC93Cff"),
		Header:        Hex("#ffffffff"),
		HeaderHover:   Hex("#FFE9B8ff"),
		HeaderActive:  Hex("#FFC93Cff"),
		Text:          Hex("#211D19ff"), // ink
		TextDisabled:  Hex("#8A8071ff"),
		Accent:        Hex("#FFC93Cff"), // candy-yellow（配墨色文字）
		TitleBg:       Hex("#FFC93Cff"), // 平涂，取消渐变
		TitleBgActive: Hex("#FFC93Cff"),
		TitleTop:      Hex("#FFC93Cff"),
		TitleBottom:   Hex("#FFC93Cff"),
		RailBg:        Hex("#F3E7CEff"),
		White:         Color{1, 1, 1, 1},
		Rounding:      10,
	}
}
