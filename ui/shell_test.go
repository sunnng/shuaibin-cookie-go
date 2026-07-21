package ui

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// blockingHooks 返回一对阻塞 run 的 hooks：run 直到 stop 被调用才返回。
func blockingHooks() (ScriptHooks, chan struct{}) {
	stopCh := make(chan struct{})
	hooks := ScriptHooks{
		OnStart: func() (func() error, func(), func(), func()) {
			run := func() error {
				<-stopCh
				return nil
			}
			return run, func() {}, func() {}, func() {}
		},
		OnExit: func() {},
	}
	return hooks, stopCh
}

func waitState(t *testing.T, c *ScriptController, want ScriptState) {
	t.Helper()
	deadline := time.Now().Add(500 * time.Millisecond)
	for c.State() != want && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	if c.State() != want {
		t.Fatalf("state=%v want %v", c.State(), want)
	}
}

func TestShellDefaults(t *testing.T) {
	s := NewShell(ShellOptions{})
	if s.Store() == nil {
		t.Fatal("Store should default to NewStore")
	}
	if s.Theme() != DefaultTheme() {
		t.Fatal("zero Theme should default to DefaultTheme")
	}
	if s.PanelOpen() {
		t.Fatal("panel closed by default")
	}
	if s.Minimized() || s.AutoPaused() {
		t.Fatal("minimized/autoPaused default false")
	}
}

func TestShellOpenPanelAutoPausesAndCloseResumes(t *testing.T) {
	hooks, stopCh := blockingHooks()
	defer close(stopCh)
	ctrl := NewScriptController(hooks)
	s := NewShell(ShellOptions{Controller: ctrl})

	ctrl.Start()
	waitState(t, ctrl, StateRunning)

	s.OpenPanel()
	if !s.PanelOpen() || ctrl.State() != StatePaused || !s.AutoPaused() {
		t.Fatalf("open panel should auto-pause: open=%v state=%v auto=%v",
			s.PanelOpen(), ctrl.State(), s.AutoPaused())
	}

	s.ClosePanel()
	if s.PanelOpen() || ctrl.State() != StateRunning || s.AutoPaused() {
		t.Fatalf("close panel should resume: open=%v state=%v auto=%v",
			s.PanelOpen(), ctrl.State(), s.AutoPaused())
	}
}

func TestShellManualResumeOverridesAutoPause(t *testing.T) {
	hooks, stopCh := blockingHooks()
	defer close(stopCh)
	ctrl := NewScriptController(hooks)
	s := NewShell(ShellOptions{Controller: ctrl})

	ctrl.Start()
	waitState(t, ctrl, StateRunning)
	s.OpenPanel()

	s.PauseResume() // 手动恢复：清除 autoPaused
	if ctrl.State() != StateRunning || s.AutoPaused() {
		t.Fatalf("manual resume: state=%v auto=%v", ctrl.State(), s.AutoPaused())
	}
	s.ClosePanel() // 不得二次动作
	if ctrl.State() != StateRunning {
		t.Fatalf("close after manual resume must not touch controller: %v", ctrl.State())
	}
}

func TestShellStartWhilePanelOpenAutoPausesAndSaves(t *testing.T) {
	hooks, stopCh := blockingHooks()
	defer close(stopCh)
	ctrl := NewScriptController(hooks)
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "ui.json")

	s := NewShell(ShellOptions{Controller: ctrl, ConfigPath: cfgPath, OpenPanelOnStart: true})
	if !s.PanelOpen() {
		t.Fatal("OpenPanelOnStart should open panel")
	}
	s.Store().SetBool("k", true)

	if err := s.StartStop(); err != nil {
		t.Fatalf("StartStop: %v", err)
	}
	waitState(t, ctrl, StatePaused) // 启动后面板仍开 -> 自动暂停
	if !s.AutoPaused() {
		t.Fatal("start while panel open should auto-pause")
	}
	if _, err := os.Stat(cfgPath); err != nil {
		t.Fatalf("start should save config: %v", err)
	}

	if err := s.StartStop(); err != nil { // 停止
		t.Fatalf("StartStop stop: %v", err)
	}
}

func TestShellStatusTextOnlyWhenRunning(t *testing.T) {
	hooks, stopCh := blockingHooks()
	defer close(stopCh)
	ctrl := NewScriptController(hooks)
	s := NewShell(ShellOptions{Controller: ctrl, Status: fakeStatus{"战斗 12 场"}})
	if s.StatusText() != "" {
		t.Fatal("idle -> empty status")
	}
	ctrl.Start()
	waitState(t, ctrl, StateRunning)
	if s.StatusText() != "战斗 12 场" {
		t.Fatalf("status=%q", s.StatusText())
	}
	if NewShell(ShellOptions{Controller: ctrl}).StatusText() != "" {
		t.Fatal("nil StatusSource -> empty")
	}
}

type fakeStatus struct{ text string }

func (f fakeStatus) Text() string { return f.text }

func TestShellSeedAppliesTasksAndPanelDefaults(t *testing.T) {
	enabled := true
	s := NewShell(ShellOptions{
		Tasks: []Task{{
			ID: "arena", Title: "王国竞技场", Category: "日常", EnabledKey: "arena_enabled",
			Fields: []Field{Bool("arena_enabled", "启用", func() bool { return enabled }, func(v bool) { enabled = v })},
		}},
		Nav: []NavEntry{{ID: "tasks", Title: "任务"}, {ID: "system", Title: "系统"}},
	})
	s.Seed()
	if !s.Store().GetBool("arena_enabled") {
		t.Fatal("Seed should seed task fields")
	}
	if s.Store().GetString(KeyPanelNav) != "tasks" || s.Store().GetString(KeyPanelSelected) != "arena" {
		t.Fatal("Seed should seed panel defaults")
	}
	s.Store().SetBool("arena_enabled", false)
	s.Apply()
	if enabled {
		t.Fatal("Apply should write back to app config")
	}
}

func TestShellStopClearsAutoPaused(t *testing.T) {
	hooks, stopCh := blockingHooks()
	defer close(stopCh)
	ctrl := NewScriptController(hooks)
	s := NewShell(ShellOptions{Controller: ctrl, OpenPanelOnStart: true})

	if err := s.StartStop(); err != nil {
		t.Fatalf("StartStop: %v", err)
	}
	waitState(t, ctrl, StatePaused) // 面板开着启动 -> 自动暂停
	if !s.AutoPaused() {
		t.Fatal("expected autoPaused after start with panel open")
	}

	if err := s.StartStop(); err != nil { // 停止
		t.Fatalf("StartStop stop: %v", err)
	}
	if s.AutoPaused() {
		t.Fatal("stop should clear autoPaused")
	}
}
