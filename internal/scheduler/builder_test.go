package scheduler

import (
	"testing"
	"time"
)

func TestTaskBuilderConfigKey(t *testing.T) {
	s := New()
	ran := false
	s.Build(TaskOpts{
		Name:      "arena",
		ConfigKey: "arena",
		CheckEnabled: func() bool { return true },
		CheckReady: func() (bool, time.Duration) { return true, 0 },
		Action:    func() error { ran = true; return nil },
	})
	_, err := s.Run(false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ran {
		t.Fatal("action did not run")
	}
}
