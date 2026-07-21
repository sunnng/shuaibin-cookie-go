package ui

// Task 任务描述符（CONTEXT.md「描述符」）：应用启动时显式构造并交给框架，
// 框架据此驱动面板列表、分类 chips、配置绑定与详情页渲染。
// RenderDetail 为 nil 时详情页按 Fields 自动渲染 Form；非 nil 时为自定义
// section 逃生门（ADR-0003），内部可自行组合组件（包括复用 Form）。
type Task struct {
	ID         string
	Title      string
	Category   string // 自由字符串，即 chip 展示文本；空串不参与 chips
	EnabledKey string
	Fields     []Field
	// Summary 列表摘要行（如「已战斗 12 场」），可为 nil。
	Summary func(*Store) string
	// RenderDetail 自定义详情渲染（绘制层，Phase 2 起使用），可为 nil。
	RenderDetail func(*Ctx)
}

// Categories 按首次出现顺序返回去重后的非空分类。
func Categories(tasks []Task) []string {
	seen := map[string]bool{}
	var out []string
	for _, t := range tasks {
		if t.Category == "" || seen[t.Category] {
			continue
		}
		seen[t.Category] = true
		out = append(out, t.Category)
	}
	return out
}

// 面板偏好键（随 ui.json 落盘；非业务配置）。
const (
	KeyPanelNav      = "panel_nav"
	KeyPanelCat      = "panel_cat"
	KeyPanelSelected = "panel_selected"

	PanelCatAll = "all"
)

// FilterByCategory cat 为空或 PanelCatAll 时返回全部。
func FilterByCategory(tasks []Task, cat string) []Task {
	if cat == "" || cat == PanelCatAll {
		return append([]Task(nil), tasks...)
	}
	out := make([]Task, 0, len(tasks))
	for _, t := range tasks {
		if t.Category == cat {
			out = append(out, t)
		}
	}
	return out
}

// FindTask 按 ID 查找。
func FindTask(tasks []Task, id string) (Task, bool) {
	for _, t := range tasks {
		if t.ID == id {
			return t, true
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
	for _, t := range tasks {
		if t.EnabledKey != "" && store.GetBool(t.EnabledKey) {
			enabled++
		}
	}
	return enabled, total
}

// SeedAll 对全部任务的全部字段执行 Seed（仅填缺失键）。
func SeedAll(store *Store, tasks []Task) {
	if store == nil {
		return
	}
	for _, t := range tasks {
		for _, f := range t.Fields {
			f.Seed(store)
		}
	}
}

// ApplyAll 对全部任务的全部字段执行 Apply（写回应用配置）。
func ApplyAll(store *Store, tasks []Task) {
	if store == nil {
		return
	}
	for _, t := range tasks {
		for _, f := range t.Fields {
			f.Apply(store)
		}
	}
}

// SeedPanelDefaults 填充面板偏好默认值（仅缺失键）：导航取 navIDs[0]、
// 分类取全部、选中取首个任务。无任务/无导航则不写对应键。
func SeedPanelDefaults(store *Store, tasks []Task, navIDs []string) {
	if store == nil {
		return
	}
	if !store.HasKey(KeyPanelNav) && len(navIDs) > 0 {
		store.SetString(KeyPanelNav, navIDs[0])
	}
	if !store.HasKey(KeyPanelCat) {
		store.SetString(KeyPanelCat, PanelCatAll)
	}
	if (!store.HasKey(KeyPanelSelected) || store.GetString(KeyPanelSelected) == "") && len(tasks) > 0 {
		store.SetString(KeyPanelSelected, tasks[0].ID)
	}
}
