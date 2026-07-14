package collect

import (
	"testing"
	"time"

	"app/internal/config"
	"app/internal/statemachine"
)

func TestCollectTaskSkeletonDone(t *testing.T) {
	cfg := &config.Collect{Enabled: true}
	task := NewTask(cfg, nil, nil, DefaultFeature(), NewSession())
	err := task.runWithOptions(statemachine.RunOptions{
		Interval: 1 * time.Millisecond,
		Label:    "collect-test",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
