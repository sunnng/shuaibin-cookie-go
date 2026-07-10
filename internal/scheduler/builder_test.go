package scheduler

import (
	"testing"
	"time"
)

func TestTaskBuilderRunsWhenReady(t *testing.T) {
	s := New()
	ran := false
	s.Build(TaskOpts{
		Name:         "arena",
		CheckEnabled: func() bool { return true },
		CheckReady:   func() (bool, time.Duration) { return true, 0 },
		Action:       func() error { ran = true; return nil },
	})
	_, err := s.Run(false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ran {
		t.Fatal("action did not run")
	}
}

func TestTaskBuilderRegistersIdleProvider(t *testing.T) {
	s := New()
	s.Build(TaskOpts{
		Name:         "arena",
		CheckEnabled: func() bool { return true },
		CheckReady:   func() (bool, time.Duration) { return false, 30 * time.Second },
		WaitHUD:      func(remain time.Duration) string { return "免费刷新等待" },
		Action:       func() error { return nil },
	})

	wait, label := s.MaxIdleWait()
	if wait != 30*time.Second {
		t.Fatalf("expected idle wait 30s, got %v", wait)
	}
	if label != "免费刷新等待" {
		t.Fatalf("expected wait label, got %q", label)
	}

	hasWork, err := s.Run(false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hasWork {
		t.Fatal("task should not run while not ready")
	}
}

func TestTaskBuilderIdleProviderRespectsDisabled(t *testing.T) {
	s := New()
	s.Build(TaskOpts{
		Name:         "arena",
		CheckEnabled: func() bool { return false },
		CheckReady:   func() (bool, time.Duration) { return false, 30 * time.Second },
		WaitHUD:      func(remain time.Duration) string { return "免费刷新等待" },
		Action:       func() error { return nil },
	})

	wait, _ := s.MaxIdleWait()
	if wait != 0 {
		t.Fatalf("expected no idle wait when disabled, got %v", wait)
	}
}
