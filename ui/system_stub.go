//go:build !android || !cgo

package ui

// SystemPage 桌面/非 cgo stub：无 UI 可渲染，仅提供符号以满足编译。
func SystemPage() func(*Ctx) {
	return func(*Ctx) {}
}
