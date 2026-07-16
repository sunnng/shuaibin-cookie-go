package ui

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStoreClear(t *testing.T) {
	s := NewStore()
	s.SetBool("a", true)
	s.SetFloat("b", 1.5)
	s.Clear()
	if s.HasKey("a") || s.HasKey("b") {
		t.Fatalf("Clear should remove all keys")
	}
}

func TestClearPanelCache(t *testing.T) {
	dir := t.TempDir()
	uiPath := filepath.Join(dir, "ui.json")
	kvPath := filepath.Join(dir, "store.json")
	if err := os.WriteFile(uiPath, []byte(`{"arena_enabled":true}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(kvPath, []byte(`{"k":1}`), 0o644); err != nil {
		t.Fatal(err)
	}

	s := NewStore()
	s.SetBool(KeyArenaEnabled, true)
	reseeded := false
	if err := ClearPanelCache(s, uiPath, kvPath, func(st *Store) {
		reseeded = true
		st.SetBool(KeyArenaEnabled, false)
	}); err != nil {
		t.Fatalf("ClearPanelCache: %v", err)
	}
	if !reseeded {
		t.Fatal("expected reseed")
	}
	if s.GetBool(KeyArenaEnabled) {
		t.Fatal("reseed should set arena_enabled false")
	}
	if _, err := os.Stat(uiPath); !os.IsNotExist(err) {
		t.Fatalf("ui.json should be removed, err=%v", err)
	}
	if _, err := os.Stat(kvPath); !os.IsNotExist(err) {
		t.Fatalf("store.json should be removed, err=%v", err)
	}
}
