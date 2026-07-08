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
