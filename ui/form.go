package ui

// FormFieldValue 按字段控件类型从 store 读当前值：
// checkbox→bool、number→float64、text→string。供 Form 渲染与测试使用。
func FormFieldValue(s *Store, f Field) any {
	if s == nil {
		switch f.Widget() {
		case WidgetCheckbox:
			return false
		case WidgetNumberInput:
			return float64(0)
		default:
			return ""
		}
	}
	switch f.Widget() {
	case WidgetCheckbox:
		return s.GetBool(f.Key())
	case WidgetNumberInput:
		return s.GetFloat(f.Key())
	default:
		return s.GetString(f.Key())
	}
}

// FormFieldChanged 把控件新值写回 store：checkbox 收 bool、number 收 float64、
// text 收 string；类型不符安全忽略，nil store 不 panic。
func FormFieldChanged(s *Store, f Field, v any) {
	if s == nil {
		return
	}
	switch f.Widget() {
	case WidgetCheckbox:
		if b, ok := v.(bool); ok {
			s.SetBool(f.Key(), b)
		}
	case WidgetNumberInput:
		if n, ok := v.(float64); ok {
			s.SetFloat(f.Key(), n)
		}
	default:
		if str, ok := v.(string); ok {
			s.SetString(f.Key(), str)
		}
	}
}
