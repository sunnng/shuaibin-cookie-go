package ui

import "strings"

// Ctx 组件帧上下文：每帧贯穿组件树的句柄，携带 Store、缩放系数与当前
// 组件嵌套路径。组件状态经 State 按「路径 + 显式键」托管（ADR-0003）。
// Ctx 只在 UI goroutine 使用，非并发安全。
type Ctx struct {
	Store *Store
	Scale float64

	path   []string
	states map[string]any
}

// NewCtx 创建帧上下文；scale <= 0 归一为 1。
func NewCtx(store *Store, scale float64) *Ctx {
	if scale <= 0 {
		scale = 1
	}
	return &Ctx{Store: store, Scale: scale, states: map[string]any{}}
}

// Push 进入子组件作用域（组件函数惯例：Push(id) + defer Pop()）。
func (c *Ctx) Push(id string) { c.path = append(c.path, id) }

// Pop 离开当前组件作用域；空路径调用安全。
func (c *Ctx) Pop() {
	if n := len(c.path); n > 0 {
		c.path = c.path[:n-1]
	}
}

// S 把基准分辨率（1600×900）下的尺寸换算为设备尺寸。
func (c *Ctx) S(base float64) float64 { return base * c.Scale }

func (c *Ctx) scope() string { return strings.Join(c.path, "/") }

// State 返回当前组件实例内 key 对应的托管状态指针；首次访问写入 initial。
// 同一路径 + 同一键跨帧返回同一指针；不同组件实例（路径不同）各自隔离。
// 规则：同一组件实例内键唯一；条件渲染自由（ADR-0003）。
func State[T any](c *Ctx, key string, initial T) *T {
	full := c.scope() + "\x00" + key
	if v, ok := c.states[full]; ok {
		return v.(*T)
	}
	v := new(T)
	*v = initial
	c.states[full] = v
	return v
}
