package jelly

import (
	"path/filepath"
	"testing"
	"time"

	"app/internal/store"
)

func TestStateWaitRoundTrip(t *testing.T) {
	dir := t.TempDir()
	s1 := NewState(store.New(filepath.Join(dir, "store.json")))

	if ready, remain := s1.CheckReady(); !ready || remain != 0 {
		t.Fatalf("empty state should be ready, got ready=%v remain=%v", ready, remain)
	}

	s1.EnterWait(30 * time.Minute)
	if ready, remain := s1.CheckReady(); ready || remain <= 0 {
		t.Fatalf("after EnterWait should not be ready, got ready=%v remain=%v", ready, remain)
	}

	s2 := NewState(store.New(filepath.Join(dir, "store.json")))
	if s2.RestoreProgress() <= 0 {
		t.Fatal("wait should persist across State instances")
	}

	s2.Clear()
	if ready, _ := s2.CheckReady(); !ready {
		t.Fatal("after Clear should be ready")
	}
}
