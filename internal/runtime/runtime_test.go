package runtime

import (
	"image"
	"testing"
	"time"

	"app/internal/guard"
	"app/internal/hud"
	"app/internal/screen"
	"app/internal/scheduler"
)

type mockDetector struct{}

func (m *mockDetector) Capture() *image.NRGBA { return nil }
func (m *mockDetector) MatchColor(x, y int, color string, sim float32) bool { return false }
func (m *mockDetector) FindColor(region screen.Region, color string, sim float32, dir int) (screen.Point, bool) { return screen.Point{}, false }
func (m *mockDetector) MatchMultiColor(colors string, sim float32) bool { return false }
func (m *mockDetector) MatchImage(region screen.Region, template []byte, sim float32) (screen.Point, bool) { return screen.Point{}, false }
func (m *mockDetector) OCRText(region screen.Region) string { return "" }

func TestRuntimeRegistersAndRunsOnce(t *testing.T) {
	s := scheduler.New()
	g := guard.New(&mockDetector{})
	h := hud.New()
	rt := New(Options{Scheduler: s, Guard: g, HUD: h, RuntimeCfg: RuntimeConfig{
		GuardInterval: 50 * time.Millisecond,
		IdleDelay:     50 * time.Millisecond,
		StepDelay:     50 * time.Millisecond,
		StopOnError:   false,
	}})

	registered := false
	rt.Register(func() { registered = true })

	// Run in background and stop after first tick
	done := make(chan struct{})
	go func() {
		_ = rt.Run()
		close(done)
	}()
	time.Sleep(100 * time.Millisecond)
	rt.Stop()
	<-done

	if !registered {
		t.Fatal("register not called")
	}
}
