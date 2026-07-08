package bot

import (
	"image"
	"testing"
	"time"

	"app/internal/config"
	"app/internal/platform/action"
	"app/internal/platform/screen"
)

type mockState struct {
	name      string
	detect    bool
	next      State
	actCalled bool
	recCalled bool
}

func (m *mockState) Name() string               { return m.name }
func (m *mockState) Detect(ctx *Context) bool   { return m.detect }
func (m *mockState) Act(ctx *Context) error     { m.actCalled = true; return nil }
func (m *mockState) Next(ctx *Context) State    { return m.next }
func (m *mockState) Recover(ctx *Context) error { m.recCalled = true; return nil }

type mockDetector struct{}

func (m *mockDetector) Capture() *image.NRGBA { return nil }
func (m *mockDetector) MatchColor(x, y int, color string, sim float32) bool {
	return false
}
func (m *mockDetector) FindColor(region screen.Region, color string, sim float32, dir int) (screen.Point, bool) {
	return screen.Point{}, false
}
func (m *mockDetector) MatchMultiColor(colors string, sim float32) bool { return false }
func (m *mockDetector) MatchImage(region screen.Region, template []byte, sim float32) (screen.Point, bool) {
	return screen.Point{}, false
}
func (m *mockDetector) OCRText(region screen.Region) string { return "" }

type mockExecutor struct {
	homeCalled bool
	sleepCalls []int
}

func (m *mockExecutor) Tap(p action.Point) error                  { return nil }
func (m *mockExecutor) LongTap(p action.Point, ms int) error      { return nil }
func (m *mockExecutor) Swipe(from, to action.Point, ms int) error { return nil }
func (m *mockExecutor) Back() error                               { return nil }
func (m *mockExecutor) Home() error                               { m.homeCalled = true; return nil }
func (m *mockExecutor) Sleep(ms int)                              { m.sleepCalls = append(m.sleepCalls, ms) }

func TestMachineTransitions(t *testing.T) {
	home := &mockState{name: "home", detect: true}
	battle := &mockState{name: "battle"}
	home.next = battle

	reg := NewRegistry()
	reg.Register(home)
	reg.Register(battle)

	ctx := &Context{
		Config:    config.DefaultConfig(),
		Detector:  &mockDetector{},
		Executor:  &mockExecutor{},
		Current:   home,
		EnteredAt: time.Now(),
	}

	m := NewMachine(ctx, reg)
	m.tick()

	if !home.actCalled {
		t.Fatal("home.Act should be called")
	}
	if ctx.Current != battle {
		t.Fatalf("expected current state battle, got %v", ctx.Current)
	}
}

func TestMachineStuckRecovery(t *testing.T) {
	state := &mockState{name: "stuck", detect: true}
	reg := NewRegistry()
	reg.Register(state)

	cfg := config.DefaultConfig()
	cfg.MaxStateDurationSec = 1

	ctx := &Context{
		Config:    cfg,
		Detector:  &mockDetector{},
		Executor:  &mockExecutor{},
		Current:   state,
		EnteredAt: time.Now().Add(-10 * time.Second),
	}

	m := NewMachine(ctx, reg)
	m.tick()

	if !state.recCalled {
		t.Fatal("Recover should be called when state exceeds MaxStateDurationSec")
	}
}

func TestMachineUnknownRecovery(t *testing.T) {
	state := &mockState{name: "lost", detect: false}
	reg := NewRegistry()

	cfg := config.DefaultConfig()
	cfg.MaxUnknownRetries = 1

	ctx := &Context{
		Config:    cfg,
		Detector:  &mockDetector{},
		Executor:  &mockExecutor{},
		Current:   state,
		EnteredAt: time.Now(),
	}

	m := NewMachine(ctx, reg)
	m.tick()

	if !state.recCalled {
		t.Fatal("Recover should be called when current state is no longer detected")
	}
}

func TestMachineRegistryFindOnUnknown(t *testing.T) {
	lost := &mockState{name: "lost", detect: false}
	found := &mockState{name: "found", detect: true}
	reg := NewRegistry()
	reg.Register(lost)
	reg.Register(found)

	cfg := config.DefaultConfig()
	cfg.MaxUnknownRetries = 1

	ctx := &Context{
		Config:    cfg,
		Detector:  &mockDetector{},
		Executor:  &mockExecutor{},
		Current:   lost,
		EnteredAt: time.Now(),
	}

	m := NewMachine(ctx, reg)
	m.tick()

	if ctx.Current != found {
		t.Fatalf("expected current state found, got %v", ctx.Current)
	}
}

func TestMachineLowPowerWait(t *testing.T) {
	state := &mockState{name: "lost", detect: false}
	reg := NewRegistry()

	cfg := config.DefaultConfig()
	cfg.MaxUnknownRetries = 1
	cfg.MaxRecoveryAttempts = 1
	cfg.LowPowerWaitSec = 7

	exec := &mockExecutor{}
	ctx := &Context{
		Config:    cfg,
		Detector:  &mockDetector{},
		Executor:  exec,
		Current:   state,
		EnteredAt: time.Now(),
	}

	m := NewMachine(ctx, reg)
	m.tick() // first unknown triggers recover
	m.tick() // second unknown exceeds max recovery attempts

	found := false
	for _, ms := range exec.sleepCalls {
		if ms == cfg.LowPowerWaitSec*1000 {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected low power wait sleep of %d ms, got %v", cfg.LowPowerWaitSec*1000, exec.sleepCalls)
	}
}

func TestMachineUnknownNilCurrent(t *testing.T) {
	reg := NewRegistry()

	cfg := config.DefaultConfig()
	cfg.MaxUnknownRetries = 1

	exec := &mockExecutor{}
	ctx := &Context{
		Config:    cfg,
		Detector:  &mockDetector{},
		Executor:  exec,
		Current:   nil,
		EnteredAt: time.Now(),
	}

	m := NewMachine(ctx, reg)
	m.tick()

	if !exec.homeCalled {
		t.Fatal("Home should be called when no current state and no state detected")
	}
}

func TestMachineFindFromNilCurrent(t *testing.T) {
	found := &mockState{name: "found", detect: true}
	reg := NewRegistry()
	reg.Register(found)

	ctx := &Context{
		Config:    config.DefaultConfig(),
		Detector:  &mockDetector{},
		Executor:  &mockExecutor{},
		Current:   nil,
		EnteredAt: time.Now(),
	}

	m := NewMachine(ctx, reg)
	m.tick()

	if ctx.Current != found {
		t.Fatalf("expected current state found, got %v", ctx.Current)
	}
}
