package mining

import (
	"path/filepath"
	"testing"
	"time"

	"app/internal/store"
)

func TestStateBusyWaitRoundTrip(t *testing.T) {
	dir := t.TempDir()
	s1 := NewState(store.New(filepath.Join(dir, "store.json")))

	if ready, remain := s1.CheckReady(); !ready || remain != 0 {
		t.Fatalf("empty state should be ready, got ready=%v remain=%v", ready, remain)
	}

	s1.EnterBusyWait(20 * time.Minute)
	if ready, remain := s1.CheckReady(); ready || remain <= 0 {
		t.Fatalf("after EnterBusyWait should not be ready, got ready=%v remain=%v", ready, remain)
	}

	s2 := NewState(store.New(filepath.Join(dir, "store.json")))
	if ready, _ := s2.CheckReady(); ready {
		t.Fatal("busy wait should persist across State instances")
	}
	if s2.RestoreProgress() <= 0 {
		t.Fatal("RestoreProgress should return remaining wait")
	}

	s2.Clear()
	if ready, _ := s2.CheckReady(); !ready {
		t.Fatal("after Clear should be ready")
	}
}

func TestStateBusyWaitExpired(t *testing.T) {
	s := NewState(store.New(filepath.Join(t.TempDir(), "store.json")))
	s.EnterBusyWait(-time.Second)
	if ready, _ := s.CheckReady(); !ready {
		t.Fatal("expired busy wait should be ready")
	}
}
