package ui

// WidgetKind 字段在面板中使用的控件形态。
type WidgetKind int

const (
	WidgetCheckbox WidgetKind = iota
	WidgetNumberInput
	WidgetTextInput
)

func (k WidgetKind) String() string {
	switch k {
	case WidgetCheckbox:
		return "checkbox"
	case WidgetNumberInput:
		return "number"
	case WidgetTextInput:
		return "text"
	default:
		return "unknown"
	}
}

// NumberConstraints 数字字段的取值约束，供 Form 渲染 NumberInput 使用。
type NumberConstraints struct {
	Min, Max, Step float64
}

// Field 一项配置的唯一声明处（CONTEXT.md「字段」）：种子写入、回写配置、
// 面板渲染都由它推导。实现不可变，构造后安全共享。
type Field interface {
	Key() string
	Label() string
	Widget() WidgetKind
	Constraints() NumberConstraints
	// Seed 在 store 缺少该键时写入应用配置的当前值。
	Seed(*Store)
	// Apply 把 store 值写回应用配置。
	Apply(*Store)
}

type field[T any] struct {
	key    string
	label  string
	widget WidgetKind
	cons   NumberConstraints
	get    func() T
	set    func(T)
	sget   func(*Store, string) T
	sset   func(*Store, string, T)
}

func (f field[T]) Key() string                    { return f.key }
func (f field[T]) Label() string                  { return f.label }
func (f field[T]) Widget() WidgetKind             { return f.widget }
func (f field[T]) Constraints() NumberConstraints { return f.cons }

func (f field[T]) Seed(s *Store) {
	if s == nil || f.get == nil {
		return
	}
	if !s.HasKey(f.key) {
		f.sset(s, f.key, f.get())
	}
}

func (f field[T]) Apply(s *Store) {
	if s == nil || f.set == nil {
		return
	}
	f.set(f.sget(s, f.key))
}

// Bool 声明一个布尔字段（面板渲染为复选框）。
func Bool(key, label string, get func() bool, set func(bool)) Field {
	return field[bool]{
		key: key, label: label, widget: WidgetCheckbox, get: get, set: set,
		sget: func(s *Store, k string) bool { return s.GetBool(k) },
		sset: func(s *Store, k string, v bool) { s.SetBool(k, v) },
	}
}

// Number 声明一个整数字段（面板渲染为步进数字输入框；store 中以 float64 存放，
// 读写时与 int 互转）。min/max/step 仅供渲染层约束，Seed/Apply 不做钳制。
func Number(key, label string, min, max, step float64, get func() int, set func(int)) Field {
	return field[int]{
		key: key, label: label, widget: WidgetNumberInput,
		cons: NumberConstraints{Min: min, Max: max, Step: step}, get: get, set: set,
		sget: func(s *Store, k string) int { return int(s.GetFloat(k)) },
		sset: func(s *Store, k string, v int) { s.SetFloat(k, float64(v)) },
	}
}

// Text 声明一个字符串字段（面板渲染为文本输入框）。
func Text(key, label string, get func() string, set func(string)) Field {
	return field[string]{
		key: key, label: label, widget: WidgetTextInput, get: get, set: set,
		sget: func(s *Store, k string) string { return s.GetString(k) },
		sset: func(s *Store, k string, v string) { s.SetString(k, v) },
	}
}
