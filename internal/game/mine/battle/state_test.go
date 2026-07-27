package battle

import (
	"path/filepath"
	"testing"
	"time"

	"app/internal/store"
)

func TestStateRecordAndCooldown(t *testing.T) {
	dir := t.TempDir()
	s := NewState(store.New(filepath.Join(dir, "store.json")))

	if remain := s.GetTimeUntilNext(time.Hour); remain != 0 {
		t.Fatalf("no record should be ready, remain=%v", remain)
	}

	s.RecordBattle()
	ready, remain := s.CheckReady(time.Hour)
	if ready || remain <= 0 {
		t.Fatalf("just recorded should be cooling down, ready=%v remain=%v", ready, remain)
	}

	// 持久化
	s2 := NewState(store.New(filepath.Join(dir, "store.json")))
	if remain := s2.GetTimeUntilNext(time.Hour); remain <= 0 {
		t.Fatal("battle time should persist across State instances")
	}
	// 间隔已过则就绪
	if ready, _ := s2.CheckReady(time.Nanosecond); !ready {
		t.Fatal("past interval should be ready")
	}

	s2.Clear()
	if remain := s2.GetTimeUntilNext(time.Hour); remain != 0 {
		t.Fatal("after Clear should be ready")
	}
}
