package ui

// ModuleCategory 任务分类（列表过滤用）。
type ModuleCategory string

const (
	CategoryDaily ModuleCategory = "daily"
	CategoryEvent ModuleCategory = "event"
	CategoryMaint ModuleCategory = "maint"
)

// 面板导航 / 列表状态（UI 会话偏好，可随 ui.json 落盘；非业务 config）。
const (
	KeyPanelNav      = "panel_nav"      // tasks | system
	KeyPanelCat      = "panel_cat"      // all | daily | event | maint
	KeyPanelSelected = "panel_selected" // module id

	PanelNavTasks  = "tasks"
	PanelNavSystem = "system"

	PanelCatAll   = "all"
	PanelCatDaily = "daily"
	PanelCatEvent = "event"
	PanelCatMaint = "maint"
)

// Module 面板列表项（编译期静态表，非运行时注册）。
type Module struct {
	ID         string
	Title      string
	Category   ModuleCategory
	EnabledKey string
}

// BuiltinModules 当前已实现模块。新增模块：在此追加一项，并在 Android 面板 switch 里加详情渲染。
func BuiltinModules() []Module {
	return []Module{
		{
			ID:         "arena",
			Title:      "王国竞技场",
			Category:   CategoryDaily,
			EnabledKey: KeyArenaEnabled,
		},
	}
}

// SeedPanelDefaults 填充列表导航默认值（仅缺失 key）。
func SeedPanelDefaults(store *Store) {
	if store == nil {
		return
	}
	if !store.HasKey(KeyPanelNav) {
		store.SetString(KeyPanelNav, PanelNavTasks)
	}
	if !store.HasKey(KeyPanelCat) {
		store.SetString(KeyPanelCat, PanelCatAll)
	}
	if !store.HasKey(KeyPanelSelected) || store.GetString(KeyPanelSelected) == "" {
		mods := BuiltinModules()
		if len(mods) > 0 {
			store.SetString(KeyPanelSelected, mods[0].ID)
		}
	}
}

// FilterByCategory cat 为空或 "all" 时返回全部。
func FilterByCategory(mods []Module, cat string) []Module {
	if cat == "" || cat == PanelCatAll {
		out := make([]Module, len(mods))
		copy(out, mods)
		return out
	}
	out := make([]Module, 0, len(mods))
	for _, m := range mods {
		if string(m.Category) == cat {
			out = append(out, m)
		}
	}
	return out
}

// FindModule 按 ID 查找。
func FindModule(mods []Module, id string) (Module, bool) {
	for _, m := range mods {
		if m.ID == id {
			return m, true
		}
	}
	return Module{}, false
}

// CountEnabled 统计 EnabledKey 为 true 的模块数。
func CountEnabled(store *Store, mods []Module) (enabled, total int) {
	total = len(mods)
	if store == nil {
		return 0, total
	}
	for _, m := range mods {
		if m.EnabledKey != "" && store.GetBool(m.EnabledKey) {
			enabled++
		}
	}
	return enabled, total
}

func categoryLabel(c ModuleCategory) string {
	switch c {
	case CategoryDaily:
		return "日常"
	case CategoryEvent:
		return "活动"
	case CategoryMaint:
		return "维护"
	default:
		return string(c)
	}
}
