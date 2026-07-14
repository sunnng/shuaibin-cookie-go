//go:build android && cgo

package ui

import (
	"unicode"

	"github.com/Dasongzi1366/AutoGo/imgui"
)

// measureLabelSize 测量按钮/标签文字尺寸。
// Android 中文字体字宽常被 CalcTextSize 低估，导致按钮内中文横向裁切；
// 对 CJK 按「国」字宽保底，并保证高度不低于 FontSize。
func measureLabelSize(label string) imgui.Vec2 {
	sz := imgui.CalcTextSize(label)
	fontH := imgui.FontSize()
	if fontH <= 0 {
		fontH = imgui.CalcTextSize("国").Y
	}
	if sz.Y < fontH {
		sz.Y = fontH
	}

	cjk := 0
	other := 0
	for _, r := range label {
		if isCJKRune(r) {
			cjk++
		} else if !unicode.IsSpace(r) {
			other++
		}
	}
	if cjk > 0 {
		ideographW := imgui.CalcTextSize("国").X
		if ideographW < 1 {
			ideographW = fontH
		}
		minW := float32(cjk)*ideographW + float32(other)*(ideographW*0.55)
		if sz.X < minW {
			sz.X = minW
		}
		// 额外安全边，避免抗锯齿/字间距裁切
		sz.X += ideographW * 0.15
	}
	return sz
}

func isCJKRune(r rune) bool {
	switch {
	case r >= 0x4E00 && r <= 0x9FFF: // CJK Unified
		return true
	case r >= 0x3400 && r <= 0x4DBF: // Extension A
		return true
	case r >= 0xF900 && r <= 0xFAFF: // Compatibility
		return true
	case r >= 0x3000 && r <= 0x303F: // CJK punctuation
		return true
	case r >= 0xFF00 && r <= 0xFFEF: // Fullwidth forms
		return true
	default:
		return false
	}
}

// fitButtonSize 按文字自适应按钮宽高（含内边距）。height<=0 时用文字高度+padY。
func fitButtonSize(label string, padX, padY float32) (w, h float32) {
	sz := measureLabelSize(label)
	w = sz.X + padX*2
	h = sz.Y + padY*2
	minH := imgui.FrameHeight()
	if h < minH {
		h = minH
	}
	if w < 48 {
		w = 48
	}
	return w, h
}
