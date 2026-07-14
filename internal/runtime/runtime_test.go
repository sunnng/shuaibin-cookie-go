package runtime

import (
	"image"
	"sync"
	"testing"
	"time"

	"app/internal/guard"
	"app/internal/platform/screen"
	"app/internal/scheduler"
)

type mockDetector struct{}

func (m *mockDetector) Capture() *image.NRGBA                               { return nil }
func (m *mockDetector) MatchColor(x, y int, color string, sim float32) bool { return false }
func (m *mockDetector) FindColor(region screen.Region, color string, sim float32, dir int) (screen.Point, bool) {
	return screen.Point{}, false
}
func (m *mockDetector) MatchMultiColor(colors string, sim float32) bool { return false }
func (m *mockDetector) FindMultiColorsAll(region screen.Region, colors string, sim float32, dir int) []screen.Point {
	return nil
}
func (m *mockDetector) MatchImage(region screen.Region, template []byte, sim float32) (screen.Point, bool) {
	return screen.Point{}, false
}
func (m *mockDetector) OCRText(region screen.Region) string { return "" }
func (m *mockDetector) FindOCRText(region screen.Region, keyword string) (screen.Point, bool) {
	return screen.Point{}, false
}

func TestRuntimeRunsAndStops(t *testing.T) {
	s := scheduler.New()
	g := guard.New(&mockDetector{})
	rt := New(Options{Scheduler: s, Guard: g, RuntimeCfg: RuntimeConfig{
		GuardInterval: 50 * time.Millisecond,
		IdleDelay:     50 * time.Millisecond,
		StepDelay:     50 * time.Millisecond,
		StopOnError:   false,
	}})

	done := make(chan struct{})
	go func() {
		_ = rt.Run()
		close(done)
	}()

	time.Sleep(80 * time.Millisecond)
	rt.Stop()
	<-done
}

func TestRuntimeIdleUsesMaxIdleWait(t *testing.T) {
	s := scheduler.New()
	s.AddIdleProvider("arena", func() (time.Duration, string) {
		return 200 * time.Millisecond, "免费刷新等待"
	})
	rt := New(Options{Scheduler: s, RuntimeCfg: RuntimeConfig{
		GuardInterval: 20 * time.Millisecond,
		IdleDelay:     5 * time.Second,
		StepDelay:     50 * time.Millisecond,
	}})

	start := time.Now()
	done := make(chan struct{})
	go func() {
		_ = rt.Run()
		close(done)
	}()

	time.Sleep(120 * time.Millisecond)
	rt.Stop()
	<-done

	elapsed := time.Since(start)
	if elapsed >= 500*time.Millisecond {
		t.Fatalf("expected idle wait capped by provider remain, elapsed=%v", elapsed)
	}
}

func TestRuntimePauseBlocksScheduling(t *testing.T) {
	var mu sync.Mutex
	calls := 0

	s := scheduler.New()
	s.Add("tick", func() bool { return true }, func() error {
		mu.Lock()
		calls++
		mu.Unlock()
		return nil
	})

	rt := New(Options{Scheduler: s, RuntimeCfg: RuntimeConfig{
		GuardInterval: 10 * time.Millisecond,
		IdleDelay:     10 * time.Millisecond,
		StepDelay:     20 * time.Millisecond,
	}})

	done := make(chan struct{})
	go func() {
		_ = rt.Run()
		close(done)
	}()

	time.Sleep(60 * time.Millisecond)
	rt.Pause()
	time.Sleep(40 * time.Millisecond)

	mu.Lock()
	pausedCalls := calls
	mu.Unlock()

	time.Sleep(80 * time.Millisecond)

	mu.Lock()
	after := calls
	mu.Unlock()
	if after != pausedCalls {
		t.Fatalf("expected no new calls while paused, before=%d after=%d", pausedCalls, after)
	}

	rt.Resume()
	time.Sleep(80 * time.Millisecond)

	mu.Lock()
	resumed := calls
	mu.Unlock()
	if resumed <= pausedCalls {
		t.Fatalf("expected calls to resume, paused=%d resumed=%d", pausedCalls, resumed)
	}

	rt.Stop()
	<-done
}

func TestRuntimeStopInterruptsSleep(t *testing.T) {
	s := scheduler.New()
	s.Add("idle", func() bool { return false }, func() error { return nil })
	rt := New(Options{Scheduler: s, RuntimeCfg: RuntimeConfig{
		GuardInterval: 50 * time.Millisecond,
		IdleDelay:     2 * time.Second,
		StepDelay:     50 * time.Millisecond,
	}})

	done := make(chan struct{})
	go func() {
		_ = rt.Run()
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)
	stopAt := time.Now()
	rt.Stop()

	select {
	case <-done:
		if elapsed := time.Since(stopAt); elapsed > 300*time.Millisecond {
			t.Fatalf("Stop did not interrupt idle sleep promptly, elapsed=%v", elapsed)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Run did not return after Stop during idle sleep")
	}
}

func TestRuntimeStopUnblocksPause(t *testing.T) {
	s := scheduler.New()
	rt := New(Options{Scheduler: s, RuntimeCfg: RuntimeConfig{
		GuardInterval: 10 * time.Millisecond,
		IdleDelay:     50 * time.Millisecond,
		StepDelay:     20 * time.Millisecond,
	}})

	done := make(chan struct{})
	go func() {
		_ = rt.Run()
		close(done)
	}()

	time.Sleep(30 * time.Millisecond)
	rt.Pause()

	select {
	case <-done:
		t.Fatal("Run returned while paused without Stop")
	case <-time.After(50 * time.Millisecond):
	}

	rt.Stop()
	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Stop did not unblock paused Run")
	}
}
