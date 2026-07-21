package ui

import "strings"

// Ctx 组件帧上下文：每帧贯穿组件树的句柄，携带 Store、缩放系数与当前
// 组件嵌套路径。组件状态经 State 按「路径 + 显式键」托管（ADR-0003）。
// Ctx 只在 UI goroutine 使用，非并发安全。
type Ctx struct {
	Store *Store
	Scale float64
	// Theme 本帧主题（RunShell 在启动时注入）；零值时 theme() 回退 DefaultTheme。
	Theme Theme
	// Shell 持有本帧所属的 Shell 实例（RunShell 注入）；面板页面组件经它
	// 访问任务表、路径与控制器。手工构造的 Ctx 可留空。
	Shell *Shell

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
// 同键不同类型重复调用会 panic（视为调用方错误）。
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

// theme 返回生效主题：Ctx.Theme 为零值时回退默认主题。
func (c *Ctx) theme() Theme {
	if c.Theme == (Theme{}) {
		return DefaultTheme()
	}
	return c.Theme
}

// resource 返回 key 对应的缓存资源，仅在缺失时调用 create 创建一次。
// 供 android 绘制层缓存纹理等后端资源；与组件路径无关（全局缓存）。
func (c *Ctx) resource(key string, create func() any) any {
	full := "res\x00" + key
	if v, ok := c.states[full]; ok {
		return v
	}
	v := create()
	c.states[full] = v
	return v
}
