package ui

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestSessionControllerStateTransitions(t *testing.T) {
	var started atomic.Int32
	var paused atomic.Int32
	var resumed atomic.Int32
	var stopped atomic.Int32

	c := NewSessionController(SessionHooks{
		OnStart: func() (func() error, func(), func(), func()) {
			started.Add(1)
			stopCh := make(chan struct{})
			var once sync.Once
			run := func() error {
				select {
				case <-stopCh:
					return nil
				case <-time.After(200 * time.Millisecond):
					return nil
				}
			}
			return run,
				func() { paused.Add(1) },
				func() { resumed.Add(1) },
				func() {
					stopped.Add(1)
					once.Do(func() { close(stopCh) })
				}
		},
		OnExit: func() {},
	})

	if c.State() != StateIdle {
		t.Fatalf("want Idle")
	}
	c.Start()
	time.Sleep(20 * time.Millisecond)
	if c.State() != StateRunning {
		t.Fatalf("want Running, got %v", c.State())
	}
	c.Start() // duplicate
	if started.Load() != 1 {
		t.Fatalf("duplicate Start should be ignored, started=%d", started.Load())
	}

	c.Pause()
	if c.State() != StatePaused || paused.Load() != 1 {
		t.Fatalf("want Paused")
	}
	c.Resume()
	if c.State() != StateRunning || resumed.Load() != 1 {
		t.Fatalf("want Running after resume")
	}
	c.Stop()
	deadline := time.Now().Add(500 * time.Millisecond)
	for c.State() != StateIdle && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if c.State() != StateIdle || stopped.Load() != 1 {
		t.Fatalf("want Idle after stop drains run, state=%v stopped=%d", c.State(), stopped.Load())
	}
}

func TestSessionControllerRunEndReturnsIdle(t *testing.T) {
	c := NewSessionController(SessionHooks{
		OnStart: func() (func() error, func(), func(), func()) {
			run := func() error { return nil }
			return run, func() {}, func() {}, func() {}
		},
		OnExit: func() {},
	})
	c.Start()
	time.Sleep(50 * time.Millisecond)
	if c.State() != StateIdle {
		t.Fatalf("want Idle after run ends, got %v", c.State())
	}
}

func TestSessionControllerStopBlocksStartUntilRunEnds(t *testing.T) {
	var runCount atomic.Int32
	var concurrent atomic.Int32
	var maxConcurrent atomic.Int32

	c := NewSessionController(SessionHooks{
		OnStart: func() (func() error, func(), func(), func()) {
			runCount.Add(1)
			stopCh := make(chan struct{})
			var once sync.Once
			run := func() error {
				n := concurrent.Add(1)
				for {
					cur := maxConcurrent.Load()
					if n <= cur || maxConcurrent.CompareAndSwap(cur, n) {
						break
					}
				}
				defer concurrent.Add(-1)
				select {
				case <-stopCh:
					return nil
				case <-time.After(300 * time.Millisecond):
					return nil
				}
			}
			return run, func() {}, func() {}, func() {
				once.Do(func() { close(stopCh) })
			}
		},
		OnExit: func() {},
	})

	c.Start()
	time.Sleep(20 * time.Millisecond)
	c.Stop()
	c.Start() // must be ignored while first run still draining
	if runCount.Load() != 1 {
		t.Fatalf("Start during Stop drain must be ignored, runCount=%d", runCount.Load())
	}
	deadline := time.Now().Add(500 * time.Millisecond)
	for c.State() != StateIdle && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if c.State() != StateIdle {
		t.Fatalf("want Idle after first run ends, got %v", c.State())
	}
	c.Start()
	time.Sleep(20 * time.Millisecond)
	if runCount.Load() != 2 {
		t.Fatalf("want 2 starts after Idle, got %d", runCount.Load())
	}
	c.Stop()
	deadline = time.Now().Add(500 * time.Millisecond)
	for c.State() != StateIdle && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if maxConcurrent.Load() > 1 {
		t.Fatalf("want no overlapping runs, maxConcurrent=%d", maxConcurrent.Load())
	}
}
