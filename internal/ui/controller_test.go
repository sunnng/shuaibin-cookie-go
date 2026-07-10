package ui

import (
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
			run := func() error {
				time.Sleep(200 * time.Millisecond)
				return nil
			}
			return run, func() { paused.Add(1) }, func() { resumed.Add(1) }, func() { stopped.Add(1) }
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
	time.Sleep(20 * time.Millisecond)
	if c.State() != StateIdle || stopped.Load() != 1 {
		t.Fatalf("want Idle after stop")
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

func TestSessionControllerStopStartGen(t *testing.T) {
	var runCount atomic.Int32

	c := NewSessionController(SessionHooks{
		OnStart: func() (func() error, func(), func(), func()) {
			runCount.Add(1)
			run := func() error {
				time.Sleep(500 * time.Millisecond)
				return nil
			}
			return run, func() {}, func() {}, func() {}
		},
		OnExit: func() {},
	})

	c.Start()
	time.Sleep(20 * time.Millisecond)
	c.Stop()
	c.Start()
	time.Sleep(20 * time.Millisecond)
	if c.State() != StateRunning {
		t.Fatalf("want Running after Stop+Start, got %v", c.State())
	}
	time.Sleep(600 * time.Millisecond)
	if c.State() != StateIdle {
		t.Fatalf("want Idle after second run ends, got %v", c.State())
	}
	if runCount.Load() != 2 {
		t.Fatalf("want 2 starts, got %d", runCount.Load())
	}
}
