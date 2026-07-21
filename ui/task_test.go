package ui

import "testing"

func testTasks() []Task {
	return []Task{
		{ID: "arena", Title: "王国竞技场", Category: "日常", EnabledKey: "arena_enabled"},
		{ID: "tower", Title: "混沌塔", Category: "日常", EnabledKey: "tower_enabled"},
		{ID: "raid", Title: "讨伐", Category: "活动", EnabledKey: "raid_enabled"},
		{ID: "nocat", Title: "未分类", Category: "", EnabledKey: "nocat_enabled"},
	}
}

func TestCategoriesDedupKeepsOrder(t *testing.T) {
	got := Categories(testTasks())
	if len(got) != 2 || got[0] != "日常" || got[1] != "活动" {
		t.Fatalf("Categories=%v", got)
	}
	if got := Categories(nil); len(got) != 0 {
		t.Fatalf("Categories(nil)=%v", got)
	}
}

func TestFilterByCategory(t *testing.T) {
	tasks := testTasks()
	if got := FilterByCategory(tasks, PanelCatAll); len(got) != 4 {
		t.Fatalf("all len=%d want 4", len(got))
	}
	if got := FilterByCategory(tasks, ""); len(got) != 4 {
		t.Fatalf("empty cat len=%d want 4", len(got))
	}
	if got := FilterByCategory(tasks, "日常"); len(got) != 2 {
		t.Fatalf("日常 len=%d want 2", len(got))
	}
	if got := FilterByCategory(tasks, "维护"); len(got) != 0 {
		t.Fatalf("维护 len=%d want 0", len(got))
	}
}

func TestFindTaskAndCountEnabled(t *testing.T) {
	tasks := testTasks()
	m, ok := FindTask(tasks, "arena")
	if !ok || m.Title != "王国竞技场" {
		t.Fatalf("FindTask arena: %#v ok=%v", m, ok)
	}
	if _, ok := FindTask(tasks, "nope"); ok {
		t.Fatal("expected missing")
	}

	store := NewStore()
	store.SetBool("arena_enabled", true)
	store.SetBool("raid_enabled", true)
	en, total := CountEnabled(store, tasks)
	if en != 2 || total != 4 {
		t.Fatalf("CountEnabled=%d/%d want 2/4", en, total)
	}
	if en, total := CountEnabled(nil, tasks); en != 0 || total != 4 {
		t.Fatalf("CountEnabled nil store=%d/%d", en, total)
	}
}

func TestSeedAllAndApplyAll(t *testing.T) {
	enabled := false
	maxBattles := 10
	tasks := []Task{
		{ID: "arena", Fields: []Field{
			Bool("arena_enabled", "启用", func() bool { return enabled }, func(v bool) { enabled = v }),
			Number("arena_max_battles", "上限", 0, 99, 1, func() int { return maxBattles }, func(v int) { maxBattles = v }),
		}},
	}
	s := NewStore()
	SeedAll(s, tasks)
	if s.GetFloat("arena_max_battles") != 10 {
		t.Fatal("SeedAll should seed number field")
	}
	s.SetBool("arena_enabled", true)
	s.SetFloat("arena_max_battles", 20)
	ApplyAll(s, tasks)
	if !enabled || maxBattles != 20 {
		t.Fatalf("ApplyAll: enabled=%v max=%d", enabled, maxBattles)
	}
}

func TestSeedPanelDefaults(t *testing.T) {
	store := NewStore()
	SeedPanelDefaults(store, testTasks(), []string{"tasks", "system"})
	if store.GetString(KeyPanelNav) != "tasks" {
		t.Fatalf("nav=%q", store.GetString(KeyPanelNav))
	}
	if store.GetString(KeyPanelCat) != PanelCatAll {
		t.Fatalf("cat=%q", store.GetString(KeyPanelCat))
	}
	if store.GetString(KeyPanelSelected) != "arena" {
		t.Fatalf("selected=%q", store.GetString(KeyPanelSelected))
	}

	store.SetString(KeyPanelNav, "system")
	SeedPanelDefaults(store, testTasks(), []string{"tasks", "system"})
	if store.GetString(KeyPanelNav) != "system" {
		t.Fatal("should keep existing nav")
	}

	empty := NewStore()
	SeedPanelDefaults(empty, nil, nil)
	if empty.HasKey(KeyPanelNav) || empty.HasKey(KeyPanelSelected) {
		t.Fatal("no tasks/nav -> no defaults")
	}
	SeedPanelDefaults(nil, testTasks(), nil) // nil store 不得 panic
}
