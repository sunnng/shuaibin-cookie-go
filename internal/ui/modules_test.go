package ui

import "testing"

func TestFilterByCategory(t *testing.T) {
	mods := []Module{
		{ID: "a", Title: "A", Category: CategoryDaily},
		{ID: "b", Title: "B", Category: CategoryEvent},
		{ID: "c", Title: "C", Category: CategoryDaily},
	}
	daily := FilterByCategory(mods, PanelCatDaily)
	if len(daily) != 2 {
		t.Fatalf("daily len=%d want 2", len(daily))
	}
	all := FilterByCategory(mods, PanelCatAll)
	if len(all) != 3 {
		t.Fatalf("all len=%d want 3", len(all))
	}
	maint := FilterByCategory(mods, PanelCatMaint)
	if len(maint) != 0 {
		t.Fatalf("maint len=%d want 0", len(maint))
	}
}

func TestFindModuleAndCountEnabled(t *testing.T) {
	mods := BuiltinModules()
	m, ok := FindModule(mods, "arena")
	if !ok || m.Title != "王国竞技场" {
		t.Fatalf("FindModule arena: %#v ok=%v", m, ok)
	}
	if _, ok := FindModule(mods, "nope"); ok {
		t.Fatal("expected missing")
	}

	store := NewStore()
	store.SetBool(KeyArenaEnabled, true)
	en, total := CountEnabled(store, mods)
	if en != 1 || total != 1 {
		t.Fatalf("CountEnabled=%d/%d want 1/1", en, total)
	}
}

func TestSeedPanelDefaults(t *testing.T) {
	store := NewStore()
	SeedPanelDefaults(store)
	if store.GetString(KeyPanelNav) != PanelNavTasks {
		t.Fatalf("nav=%q", store.GetString(KeyPanelNav))
	}
	if store.GetString(KeyPanelSelected) != "arena" {
		t.Fatalf("selected=%q", store.GetString(KeyPanelSelected))
	}
	store.SetString(KeyPanelNav, PanelNavSystem)
	SeedPanelDefaults(store)
	if store.GetString(KeyPanelNav) != PanelNavSystem {
		t.Fatal("should keep existing nav")
	}
}
