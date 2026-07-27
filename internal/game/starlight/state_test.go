package starlight

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

func TestStateDoneTodayFlow(t *testing.T) {
	s := newTestState(t)
	if s.IsDoneToday() {
		t.Fatal("fresh state should not be done today")
	}
	if s.Describe() != "今日未完成" {
		t.Fatalf("Describe = %q, want 今日未完成", s.Describe())
	}
	s.MarkDoneToday()
	if !s.IsDoneToday() {
		t.Fatal("IsDoneToday should be true after MarkDoneToday")
	}
	if s.Describe() != "今日已完成" {
		t.Fatalf("Describe = %q, want 今日已完成", s.Describe())
	}
	s.Clear()
	if s.IsDoneToday() {
		t.Fatal("IsDoneToday should be false after Clear")
	}
}

// 完成标记要跨 Store 重载持久化（模拟脚本重启）。
func TestStateDoneTodayPersists(t *testing.T) {
	path := filepath.Join(t.TempDir(), "store.json")
	s1 := NewState(store.New(path))
	s1.MarkDoneToday()
	s2 := NewState(store.New(path))
	if !s2.IsDoneToday() {
		t.Fatal("done mark should survive store reload")
	}
}

// 昨天写入的标记不应算今日完成。
func TestStateDoneYesterdayNotDone(t *testing.T) {
	s := newTestState(t)
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	if err := s.store.Set(doneDateKey, yesterday); err != nil {
		t.Fatal(err)
	}
	if s.IsDoneToday() {
		t.Fatal("yesterday's mark must not count as done today")
	}
}

func TestStateTimeUntilNextDay(t *testing.T) {
	s := newTestState(t)
	remain := s.TimeUntilNextDay()
	if remain <= 0 || remain > 24*time.Hour {
		t.Fatalf("TimeUntilNextDay = %v, want (0, 24h]", remain)
	}
}
