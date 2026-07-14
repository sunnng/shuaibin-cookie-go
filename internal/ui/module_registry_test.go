package ui

import "testing"

func TestRegisterModuleAndFilter(t *testing.T) {
	ClearModules()
	t.Cleanup(ClearModules)

	RegisterModule(ModuleDef{
		ID:         "arena",
		Title:      "王国竞技场",
		Category:   CategoryDaily,
		EnabledKey: KeyArenaEnabled,
		Summary:    ArenaSummary,
	})
	RegisterModule(ModuleDef{
		ID:         "collect",
		Title:      "收集",
		Category:   CategoryDaily,
		EnabledKey: KeyCollectEnabled,
		Summary:    CollectSummary,
	})
	RegisterModule(ModuleDef{
		ID:         "pvp",
		Title:      "跨服竞技",
		Category:   CategoryEvent,
		EnabledKey: "pvp_enabled",
	})

	if got := len(Modules()); got != 3 {
		t.Fatalf("Modules() len=%d want 3", got)
	}

	daily := FilterModules(CategoryDaily, "")
	if len(daily) != 2 {
		t.Fatalf("daily filter len=%d want 2", len(daily))
	}

	found := FilterModules("", "arena")
	if len(found) != 1 || found[0].ID != "arena" {
		t.Fatalf("query arena got %#v", found)
	}

	store := NewStore()
	store.SetBool(KeyArenaEnabled, true)
	en, total := CountEnabled(store)
	if en != 1 || total != 3 {
		t.Fatalf("CountEnabled=%d/%d want 1/3", en, total)
	}
}

func TestRegisterModuleOverwrite(t *testing.T) {
	ClearModules()
	t.Cleanup(ClearModules)

	RegisterModule(ModuleDef{ID: "arena", Title: "A", Category: CategoryDaily})
	RegisterModule(ModuleDef{ID: "arena", Title: "B", Category: CategoryEvent})
	mods := Modules()
	if len(mods) != 1 {
		t.Fatalf("len=%d want 1", len(mods))
	}
	if mods[0].Title != "B" || mods[0].Category != CategoryEvent {
		t.Fatalf("overwrite failed: %#v", mods[0])
	}
}

func TestSeedHubDefaults(t *testing.T) {
	ClearModules()
	t.Cleanup(ClearModules)
	RegisterModule(ModuleDef{ID: "arena", Title: "竞技场", Category: CategoryDaily})

	store := NewStore()
	SeedHubDefaults(store)
	if store.GetString(KeyHubNav) != HubNavModules {
		t.Fatalf("nav=%q", store.GetString(KeyHubNav))
	}
	if store.GetString(KeyHubSelected) != "arena" {
		t.Fatalf("selected=%q", store.GetString(KeyHubSelected))
	}
	// 不覆盖已有
	store.SetString(KeyHubNav, HubNavLicense)
	SeedHubDefaults(store)
	if store.GetString(KeyHubNav) != HubNavLicense {
		t.Fatalf("should keep existing nav")
	}
}
