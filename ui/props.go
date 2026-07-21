package ui

// 组件 Props（ADR-0003）：每组件一个结构体，唯一输入，含数据与回调。
// 本文件仅为纯数据定义；渲染函数在 Phase 2 的 android 组件中。

type ButtonKind int

const (
	ButtonPrimary ButtonKind = iota
	ButtonSecondary
	ButtonDanger
)

type ButtonProps struct {
	Label         string
	Kind          ButtonKind
	Width, Height float64 // 基准分辨率尺寸；0 表示按内容自适应
	Disabled      bool
	OnClick       func()
}

type CheckboxProps struct {
	Label    string
	Checked  bool
	OnChange func(bool)
}

type NumberInputProps struct {
	Label          string
	Value          float64
	Min, Max, Step float64 // Max <= Min 表示不钳上界
	Width          float64
	OnChange       func(float64)
}

// Clamp 把 v 钳入 [Min, Max]（Max > Min 时）；不含步进吸附。
func (p NumberInputProps) Clamp(v float64) float64 {
	if p.Max > p.Min {
		if v < p.Min {
			return p.Min
		}
		if v > p.Max {
			return p.Max
		}
	}
	return v
}

type InputProps struct {
	Label, Hint, Value string
	Width              float64
	OnChange           func(string)
}

type DropdownProps struct {
	Label    string
	Options  []string
	Selected int
	OnChange func(int)
}

type TabsProps struct {
	Items    []string
	Selected int
	OnChange func(int)
}

type CollapsibleProps struct {
	Label    string
	Open     bool
	OnToggle func(bool)
}

// FormProps Form 组件输入：按 Fields 自动排版，值直连 Store 读写。
type FormProps struct {
	Store  *Store
	Fields []Field
}
