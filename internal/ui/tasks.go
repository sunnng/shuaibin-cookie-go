package ui

// TaskCategory 任务分类（列表过滤用）。
type TaskCategory string

const (
	CategoryDaily TaskCategory = "daily"
	CategoryEvent TaskCategory = "event"
	CategoryMaint TaskCategory = "maint"
)

// 面板导航 / 列表状态（面板偏好，可随 ui.json 落盘；非业务 config）。
const (
	KeyPanelNav      = "panel_nav"      // tasks | system
	KeyPanelCat      = "panel_cat"      // all | daily | event | maint
	KeyPanelSelected = "panel_selected" // task id

	PanelNavTasks  = "tasks"
	PanelNavSystem = "system"

	PanelCatAll   = "all"
	PanelCatDaily = "daily"
	PanelCatEvent = "event"
	PanelCatMaint = "maint"
)

// Task 面板列表项（编译期静态表，非运行时注册）。
type Task struct {
	ID         string
	Title      string
	Category   TaskCategory
	EnabledKey string
}

// BuiltinTasks 当前已实现任务。新增任务：在此追加一项，并在 Android 面板 switch 里加详情渲染。
func BuiltinTasks() []Task {
	return []Task{
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
		tasks := BuiltinTasks()
		if len(tasks) > 0 {
			store.SetString(KeyPanelSelected, tasks[0].ID)
		}
	}
}

// FilterByCategory cat 为空或 "all" 时返回全部。
func FilterByCategory(tasks []Task, cat string) []Task {
	if cat == "" || cat == PanelCatAll {
		out := make([]Task, len(tasks))
		copy(out, tasks)
		return out
	}
	out := make([]Task, 0, len(tasks))
	for _, task := range tasks {
		if string(task.Category) == cat {
			out = append(out, task)
		}
	}
	return out
}

// FindTask 按 ID 查找。
func FindTask(tasks []Task, id string) (Task, bool) {
	for _, task := range tasks {
		if task.ID == id {
			return task, true
		}
	}
	return Task{}, false
}

// CountEnabled 统计 EnabledKey 为 true 的任务数。
func CountEnabled(store *Store, tasks []Task) (enabled, total int) {
	total = len(tasks)
	if store == nil {
		return 0, total
	}
	for _, task := range tasks {
		if task.EnabledKey != "" && store.GetBool(task.EnabledKey) {
			enabled++
		}
	}
	return enabled, total
}

func categoryLabel(c TaskCategory) string {
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
