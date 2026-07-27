package square

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

func TestStateDoneToday(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "store.json")

	s := NewState(store.New(path))
	if s.IsDoneToday() {
		t.Fatal("new state should not be done")
	}
	s.Ensure()
	s.MarkDoneToday()
	if !s.IsDoneToday() {
		t.Fatal("should be done after MarkDoneToday")
	}
	if s.IsActive() {
		t.Fatal("MarkDoneToday should clear the active session")
	}

	// 重新加载（模拟重启）完成标记仍在。
	s2 := NewState(store.New(path))
	if !s2.IsDoneToday() {
		t.Fatal("done flag should persist across reload")
	}
}

func TestStateCheckedToday(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "store.json")

	s := NewState(store.New(path))
	if s.HasCheckedToday() {
		t.Fatal("no session → not checked")
	}
	s.MarkCheckedToday()
	if !s.HasCheckedToday() {
		t.Fatal("should be checked after MarkCheckedToday")
	}

	s2 := NewState(store.New(path))
	if !s2.HasCheckedToday() {
		t.Fatal("checked flag should persist across reload")
	}
}

func TestStateStayAccumulate(t *testing.T) {
	s := newTestState(t)
	if got := s.StayElapsed(); got != 0 {
		t.Fatalf("no session → elapsed 0, got %v", got)
	}

	// 构造一段"5 秒前开始计时"的会话。
	a := s.Ensure()
	a.LastEnterAt = time.Now().Add(-5 * time.Second).Unix()
	s.save(a)

	if got := s.StayElapsed(); got < 5*time.Second {
		t.Fatalf("elapsed should be >= 5s, got %v", got)
	}
	if rem := s.StayRemaining(60 * time.Second); rem <= 0 || rem > 55*time.Second {
		t.Fatalf("remaining should be in (0, 55s], got %v", rem)
	}

	s.PauseStay()
	a, ok := s.GetActive()
	if !ok {
		t.Fatal("session should exist")
	}
	if a.LastEnterAt != 0 {
		t.Fatalf("PauseStay should stop the clock, lastEnterAt=%d", a.LastEnterAt)
	}
	if a.AccumulatedSec < 5 {
		t.Fatalf("PauseStay should settle >= 5s, got %d", a.AccumulatedSec)
	}
	// 暂停后停留不再增长。
	if got := s.StayElapsed(); got < 5*time.Second || got > 6*time.Second {
		t.Fatalf("paused elapsed should stay ~5s, got %v", got)
	}
}

func TestStateResetStayTimer(t *testing.T) {
	s := newTestState(t)
	a := s.Ensure()
	a.AccumulatedSec = 120
	a.LastEnterAt = 0
	s.save(a)

	s.ResetStayTimer()
	a, _ = s.GetActive()
	if a.AccumulatedSec != 0 {
		t.Fatalf("reset should zero accumulated, got %d", a.AccumulatedSec)
	}
	if a.LastEnterAt == 0 {
		t.Fatal("reset should restart the clock")
	}
	if rem := s.StayRemaining(60 * time.Second); rem <= 55*time.Second {
		t.Fatalf("after reset remaining should be ~60s, got %v", rem)
	}
}

func TestStateStartStayIdempotent(t *testing.T) {
	s := newTestState(t)
	s.StartStay()
	a, _ := s.GetActive()
	first := a.LastEnterAt
	if first == 0 {
		t.Fatal("StartStay should start the clock")
	}
	time.Sleep(1100 * time.Millisecond)
	s.StartStay()
	a, _ = s.GetActive()
	if a.LastEnterAt != first {
		t.Fatal("StartStay while running must not restart the clock")
	}
}

func TestStateCheckReady(t *testing.T) {
	s := newTestState(t)
	ready, remain := s.CheckReady()
	if !ready || remain != 0 {
		t.Fatalf("not done → (true, 0), got (%v, %v)", ready, remain)
	}
	s.MarkDoneToday()
	ready, remain = s.CheckReady()
	if ready || remain != 0 {
		t.Fatalf("done → (false, 0), got (%v, %v)", ready, remain)
	}
}

func TestStateDescribe(t *testing.T) {
	s := newTestState(t)
	if got := s.Describe(); got != "今日未完成，无挂机会话" {
		t.Fatalf("no session = %q", got)
	}
	s.Ensure()
	a, _ := s.GetActive()
	a.AccumulatedSec = 30
	s.save(a)
	if got := s.Describe(); got != "今日未完成，未初检，有效停留 30s" {
		t.Fatalf("active = %q", got)
	}
	s.MarkCheckedToday()
	if got := s.Describe(); got != "今日未完成，已初检，有效停留 30s" {
		t.Fatalf("checked = %q", got)
	}
	s.MarkDoneToday()
	if got := s.Describe(); got != "今日已完成" {
		t.Fatalf("done = %q", got)
	}
}
