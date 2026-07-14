package store

import (
	"path/filepath"
	"testing"
)

func TestStoreSetGet(t *testing.T) {
	dir := t.TempDir()
	s := New(filepath.Join(dir, "store.json"))
	if err := s.Set("arena_refresh", int64(1234567890)); err != nil {
		t.Fatalf("set failed: %v", err)
	}
	got, ok := s.GetInt64("arena_refresh")
	if !ok || got != 1234567890 {
		t.Fatalf("expected 1234567890, got %d ok=%v", got, ok)
	}
}

func TestStoreClearAll(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "store.json")
	s := New(path)
	if err := s.Set("arena_refresh", int64(1)); err != nil {
		t.Fatal(err)
	}
	if err := s.ClearAll(); err != nil {
		t.Fatal(err)
	}
	if _, ok := s.GetInt64("arena_refresh"); ok {
		t.Fatal("ClearAll should remove keys")
	}
	s2 := New(path)
	if _, ok := s2.GetInt64("arena_refresh"); ok {
		t.Fatal("ClearAll should persist empty store")
	}
}
