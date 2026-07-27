package biscuit

import (
	"path/filepath"
	"testing"

	"app/internal/store"
)

func newTestState(t *testing.T) *State {
	t.Helper()
	return NewState(store.New(filepath.Join(t.TempDir(), "store.json")))
}

func TestStateGraduatedPersistence(t *testing.T) {
	dir := t.TempDir()
	s1 := NewState(store.New(filepath.Join(dir, "store.json")))
	if s1.IsGraduated() {
		t.Fatal("fresh state should not be graduated")
	}
	s1.MarkGraduated()

	s2 := NewState(store.New(filepath.Join(dir, "store.json")))
	if !s2.IsGraduated() {
		t.Fatal("graduated flag should persist across State instances")
	}

	s2.ClearGraduated()
	s3 := NewState(store.New(filepath.Join(dir, "store.json")))
	if s3.IsGraduated() {
		t.Fatal("ClearGraduated should clear the persisted flag")
	}
}

func TestStateStatusText(t *testing.T) {
	cfg := DefaultConfig()

	s := newTestState(t)
	if got := s.StatusText(cfg); got != "洗脆饼 0/500" {
		t.Fatalf("fresh = %q", got)
	}

	s.Rolls = 12
	if got := s.StatusText(cfg); got != "洗脆饼 12/500" {
		t.Fatalf("rolling = %q", got)
	}

	s.Rolls = 500
	if got := s.StatusText(cfg); got != "洗脆饼 500/500 · 已达上限" {
		t.Fatalf("max reached = %q", got)
	}

	s2 := newTestState(t)
	s2.Rolls = 37
	s2.MarkGraduated()
	if got := s2.StatusText(cfg); got != "洗脆饼 37/500 · 已毕业" {
		t.Fatalf("graduated = %q", got)
	}

	// nil cfg 不 panic
	if got := s2.StatusText(nil); got != "洗脆饼 37/0 · 已毕业" {
		t.Fatalf("nil cfg = %q", got)
	}
}
