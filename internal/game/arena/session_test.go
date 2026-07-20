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

func TestSessionStatusText(t *testing.T) {
	max := 10
	cfg := &config.Arena{MaxBattles: &max}
	newS := func() *Session { return NewSession(store.New(filepath.Join(t.TempDir(), "store.json"))) }

	s := newS()
	if got := s.StatusText(cfg); got != "竞技场 0/10" {
		t.Fatalf("zero battles = %q", got)
	}

	s.Wins, s.Draws, s.Losses = 2, 1, 1
	if got := s.StatusText(cfg); got != "竞技场 4/10 · 胜率 50%" {
		t.Fatalf("with max = %q", got)
	}

	s.Wins = 10 // 11/10 超上限也照实显示
	if got := s.StatusText(cfg); got != "竞技场 12/10 · 胜率 83%" {
		t.Fatalf("over max = %q", got)
	}

	s2 := newS()
	s2.Wins = 3
	if got := s2.StatusText(&config.Arena{}); got != "竞技场 3 场 · 胜率 100%" {
		t.Fatalf("without max = %q", got)
	}

	if got := s2.StatusText(nil); got != "竞技场 3 场 · 胜率 100%" {
		t.Fatalf("nil cfg = %q", got)
	}
}
