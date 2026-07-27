package survey

import (
	"path/filepath"
	"testing"
	"time"

	"app/internal/store"
)

func TestStateFarWaitRoundTrip(t *testing.T) {
	dir := t.TempDir()
	s1 := NewState(store.New(filepath.Join(dir, "store.json")))

	ready, remain := s1.CheckFarWait()
	if !ready || remain != 0 {
		t.Fatalf("empty state should be ready, got ready=%v remain=%v", ready, remain)
	}

	s1.EnterFarWait(10 * time.Minute)
	ready, remain = s1.CheckFarWait()
	if ready || remain <= 0 {
		t.Fatalf("after EnterFarWait should not be ready, got ready=%v remain=%v", ready, remain)
	}

	// 持久化：新 State 实例读同一 store 仍在等待
	s2 := NewState(store.New(filepath.Join(dir, "store.json")))
	if ready, remain := s2.CheckFarWait(); ready || remain <= 0 {
		t.Fatal("far wait should persist across State instances")
	}
	if s2.RestoreProgress() <= 0 {
		t.Fatal("RestoreProgress should return remaining wait")
	}

	s2.Clear()
	if ready, _ := s2.CheckFarWait(); !ready {
		t.Fatal("after Clear should be ready")
	}
}

func TestStateFarWaitExpired(t *testing.T) {
	s := NewState(store.New(filepath.Join(t.TempDir(), "store.json")))
	s.EnterFarWait(-time.Second) // 已到期
	if ready, _ := s.CheckFarWait(); !ready {
		t.Fatal("expired far wait should be ready")
	}
}
