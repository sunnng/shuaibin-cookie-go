package production

import (
	"testing"
	"time"

	"app/internal/statemachine"
)

type stubPage struct {
	board bool
}

func (s *stubPage) IsBoard() bool                      { return s.board }
func (s *stubPage) WaitBoard(timeout time.Duration) bool { return s.board }
func (s *stubPage) TapCollectAll() bool { return true }

type stubRoute struct {
	entered bool
	left    bool
}

func (s *stubRoute) Enter() bool {
	s.entered = true
	return true
}
func (s *stubRoute) Leave() bool {
	s.left = true
	return true
}

func TestTaskRunSkeletonDone(t *testing.T) {
	state := NewState(nil)
	task := newTask(&stubPage{}, &stubRoute{}, state, nil)
	err := task.runWithOptions(statemachine.RunOptions{
		Interval: time.Millisecond,
		Label:    "王国生产",
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
}

func TestStateStatusText(t *testing.T) {
	s := NewState(nil)
	if got := s.StatusText(); got != "生产 0" {
		t.Fatalf("StatusText=%q", got)
	}
	s.Collected = 3
	if got := s.StatusText(); got != "生产 3" {
		t.Fatalf("StatusText=%q", got)
	}
}

func TestDefaultFeature(t *testing.T) {
	if DefaultFeature() == nil {
		t.Fatal("DefaultFeature nil")
	}
}
