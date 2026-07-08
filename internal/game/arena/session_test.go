package arena

import (
	"path/filepath"
	"testing"
	"time"

	"app/internal/config"
	"app/internal/store"
)

func TestSessionReachMax(t *testing.T) {
	s := NewSession(store.New(filepath.Join(t.TempDir(), "store.json")))
	max := 3
	cfg := &config.Arena{MaxBattles: &max}
	s.Wins = 2
	if s.IsReachMaxBattles(cfg) {
		t.Fatal("should not reach max")
	}
	s.Wins = 3
	if !s.IsReachMaxBattles(cfg) {
		t.Fatal("should reach max")
	}
}

func TestSessionRefreshPersistence(t *testing.T) {
	dir := t.TempDir()
	s1 := NewSession(store.New(filepath.Join(dir, "store.json")))
	at := time.Now().Add(30 * time.Minute)
	s1.SetNextFreeRefreshAt(at)

	s2 := NewSession(store.New(filepath.Join(dir, "store.json")))
	if s2.TimeUntilRefresh() <= 0 {
		t.Fatal("refresh time should persist")
	}
}
