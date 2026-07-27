package market

import (
	"path/filepath"
	"testing"
	"time"

	"app/internal/store"
)

func newTestState(t *testing.T) *State {
	t.Helper()
	return NewState(store.New(filepath.Join(t.TempDir(), "store.json")))
}

func TestCheckReadyStartupBypass(t *testing.T) {
	s := newTestState(t)
	ready, remain := s.CheckReady()
	if !ready || remain != 0 {
		t.Fatalf("first CheckReady = (%v,%v), want (true,0)", ready, remain)
	}
	if !s.ConsumeStartupBypass() {
		t.Fatal("bypass should be armed after first CheckReady")
	}
	if s.ConsumeStartupBypass() {
		t.Fatal("bypass should be consumable only once")
	}
	// 无补货记录：第二次起照常就绪
	if ready, _ := s.CheckReady(); !ready {
		t.Fatal("CheckReady should be ready with no schedule")
	}
}

func TestCheckReadyWithoutBypassNotConsumed(t *testing.T) {
	s := newTestState(t)
	if s.ConsumeStartupBypass() {
		t.Fatal("bypass must not be active before first CheckReady")
	}
}

func TestScheduleAndRestoreProgress(t *testing.T) {
	dir := t.TempDir()
	s1 := NewState(store.New(filepath.Join(dir, "store.json")))
	s1.CheckReady() // 消费首轮强制，避免干扰
	s1.ScheduleAfterRestock(time.Hour, 30)

	if remain := s1.TimeUntilRestock(); remain < 59*time.Minute || remain > 61*time.Minute {
		t.Fatalf("TimeUntilRestock = %v, want ~1h30s", remain)
	}
	ready, remain := s1.CheckReady()
	if ready {
		t.Fatal("CheckReady should be false during restock wait")
	}
	if remain < 59*time.Minute {
		t.Fatalf("CheckReady remain = %v, want ~1h", remain)
	}

	// restoreProgress 供 idle provider：跨实例持久化、无副作用
	s2 := NewState(store.New(filepath.Join(dir, "store.json")))
	if got := s2.RestoreProgress(); got < 59*time.Minute {
		t.Fatalf("RestoreProgress = %v, want ~1h", got)
	}
}

func TestScheduleAfterRestockInvalidFallsBack6h(t *testing.T) {
	s := newTestState(t)
	s.ScheduleAfterRestock(-time.Second, 30)
	if remain := s.TimeUntilRestock(); remain < 5*time.Hour {
		t.Fatalf("negative restock should fall back to 6h, remain=%v", remain)
	}
}

func TestStateClear(t *testing.T) {
	s := newTestState(t)
	s.ScheduleAfterRestock(time.Hour, 30)
	s.Clear()
	if got := s.RestoreProgress(); got != 0 {
		t.Fatalf("after clear RestoreProgress = %v, want 0", got)
	}
}

func TestStateStatusText(t *testing.T) {
	s := newTestState(t)
	if got := s.StatusText(); got != "交易所 购0" {
		t.Fatalf("zero stats = %q", got)
	}
	s.Purchased, s.SoldOut, s.Shortage, s.Failed = 3, 1, 1, 0
	if got := s.StatusText(); got != "交易所 购3 · 售罄1 · 不足1" {
		t.Fatalf("stats = %q", got)
	}
	s.Failed = 2
	if got := s.StatusText(); got != "交易所 购3 · 售罄1 · 不足1 · 失败2" {
		t.Fatalf("stats with failed = %q", got)
	}
	s.Reset()
	if got := s.StatusText(); got != "交易所 购0" {
		t.Fatalf("after reset = %q", got)
	}
}
