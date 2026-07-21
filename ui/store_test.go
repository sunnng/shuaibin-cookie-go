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

func TestStoreSaveLoadRoundtrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ui.json")
	s := NewStore()
	s.SetBool("flag", true)
	s.SetFloat("num", 42)
	s.SetString("name", "arena")
	if err := s.SaveConfig(path); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}
	s2 := NewStore()
	if err := s2.LoadConfig(path); err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if !s2.GetBool("flag") || s2.GetFloat("num") != 42 || s2.GetString("name") != "arena" {
		t.Fatalf("roundtrip mismatch: %#v", s2)
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
	s.SetBool("arena_enabled", true)
	reseeded := false
	if err := ClearPanelCache(s, uiPath, kvPath, func(st *Store) {
		reseeded = true
		st.SetBool("arena_enabled", false)
	}); err != nil {
		t.Fatalf("ClearPanelCache: %v", err)
	}
	if !reseeded {
		t.Fatal("expected reseed")
	}
	if s.GetBool("arena_enabled") {
		t.Fatal("reseed should set arena_enabled false")
	}
	if _, err := os.Stat(uiPath); !os.IsNotExist(err) {
		t.Fatalf("ui.json should be removed, err=%v", err)
	}
	if _, err := os.Stat(kvPath); !os.IsNotExist(err) {
		t.Fatalf("store.json should be removed, err=%v", err)
	}
}
