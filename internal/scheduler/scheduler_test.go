package scheduler

import (
	"errors"
	"testing"
	"time"
)

func TestSchedulerRunsTask(t *testing.T) {
	s := New()
	ran := false
	s.Add("test", func() bool { return true }, func() error { ran = true; return nil })
	hasWork, err := s.Run(false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasWork {
		t.Fatal("expected hasWork=true")
	}
	if !ran {
		t.Fatal("task did not run")
	}
}

func TestSchedulerSkipsTask(t *testing.T) {
	s := New()
	ran := false
	s.Add("test", func() bool { return false }, func() error { ran = true; return nil })
	hasWork, err := s.Run(false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hasWork {
		t.Fatal("expected hasWork=false")
	}
	if ran {
		t.Fatal("task should not run")
	}
}

func TestSchedulerStopOnError(t *testing.T) {
	s := New()
	s.Add("test", func() bool { return true }, func() error { return errors.New("boom") })
	_, err := s.Run(true)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSchedulerIdleProvider(t *testing.T) {
	s := New()
	s.AddIdleProvider("arena", func() (time.Duration, string) { return 30 * time.Second, "arena 30s" })
	wait, label := s.MaxIdleWait()
	if wait != 30*time.Second || label != "arena 30s" {
		t.Fatalf("unexpected wait=%v label=%s", wait, label)
	}
}

func TestSchedulerMaxIdleWaitReturnsMaxProviderLabel(t *testing.T) {
	s := New()
	s.AddIdleProvider("expedition", func() (time.Duration, string) { return 10 * time.Second, "expedition 10s" })
	s.AddIdleProvider("arena", func() (time.Duration, string) { return 30 * time.Second, "arena 30s" })
	wait, label := s.MaxIdleWait()
	if wait != 30*time.Second || label != "arena 30s" {
		t.Fatalf("expected wait=30s label=arena 30s, got wait=%v label=%s", wait, label)
	}
}
