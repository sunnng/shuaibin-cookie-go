package survey

import (
	"path/filepath"
	"testing"
	"time"

	"app/internal/statemachine"
	"app/internal/store"
)

type mockPage struct {
	domain  bool
	running bool
	floor   int
	floorOK bool
	setupOK bool
	stopOK  bool

	setupCalls int
	stopCalls  int
	backCalls  int
	onStop     func()
	onRead     func()
}

func (m *mockPage) IsDomain() bool  { return m.domain }
func (m *mockPage) IsRunning() bool { return m.running }
func (m *mockPage) Setup() bool     { m.setupCalls++; return m.setupOK }
func (m *mockPage) StopVenture() bool {
	m.stopCalls++
	if m.onStop != nil {
		m.onStop()
	}
	return m.stopOK
}
func (m *mockPage) GetCurrentFloor() (int, bool) {
	if m.onRead != nil {
		m.onRead()
	}
	return m.floor, m.floorOK
}
func (m *mockPage) TapBackBtn() { m.backCalls++ }

type mockHome struct {
	current      bool
	waitOK       bool
	ventureCalls int
	onVenture    func()
}

func (m *mockHome) IsCurrent() bool { return m.current }
func (m *mockHome) WaitCurrent(timeout time.Duration) bool {
	return m.waitOK
}
func (m *mockHome) TapVenture() {
	m.ventureCalls++
	if m.onVenture != nil {
		m.onVenture()
	}
}

type mockRoute struct {
	mineHomeOK     bool
	kingdomOK      bool
	toMineCalls    int
	toKingdomCalls int
}

func (m *mockRoute) KingdomHomeToMineHome() bool { m.toMineCalls++; return m.mineHomeOK }
func (m *mockRoute) MineHomeToKingdom() bool     { m.toKingdomCalls++; return m.kingdomOK }

type mockKingdom struct{ home bool }

func (m *mockKingdom) IsKingdomHome() bool { return m.home }

func newTestTask(t *testing.T, cfg *Config, p page, h homePage, r route, kp kingdomPage) *Task {
	s := NewState(store.New(filepath.Join(t.TempDir(), "store.json")))
	return newTask(cfg, p, h, r, kp, s, nil)
}

func fastRunOptions() statemachine.RunOptions {
	return statemachine.RunOptions{
		Interval: 1 * time.Millisecond,
		Label:    "survey-test",
	}
}

func testConfig() *Config {
	return &Config{Enabled: true, TargetFloor: 6, FarGap: 2, OCRPollSec: 0, FarWaitSec: 100}
}

// 主路径：已在勘查域且勘查中 → 读层达标 → 停止结算 → 重新 setup → farWait 回城 Done。
func TestSurveySettleThenReSetup(t *testing.T) {
	p := &mockPage{domain: true, running: true, floor: 6, floorOK: true, setupOK: true, stopOK: true}
	p.onStop = func() { p.running = false; p.floor = 0 } // 结算后回到未启动态
	h := &mockHome{waitOK: true}
	r := &mockRoute{kingdomOK: true}
	kp := &mockKingdom{}

	task := newTestTask(t, testConfig(), p, h, r, kp)
	if err := task.runWithOptions(fastRunOptions()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.stopCalls != 1 {
		t.Errorf("expected StopVenture once, got %d", p.stopCalls)
	}
	if p.setupCalls != 1 {
		t.Errorf("expected Setup once after settle, got %d", p.setupCalls)
	}
	if r.toKingdomCalls == 0 {
		t.Error("farWait should navigate back to kingdom")
	}
	if ready, remain := task.state.CheckFarWait(); ready || remain <= 0 {
		t.Errorf("far wait should be recorded, got ready=%v remain=%v", ready, remain)
	}
}

// 远距：当前层离目标超过 farGap → 直接 EnterFarWait 回城 Done，不停止勘查。
func TestSurveyFarDistance(t *testing.T) {
	p := &mockPage{domain: true, running: true, floor: 1, floorOK: true, setupOK: true, stopOK: true}
	h := &mockHome{waitOK: true}
	r := &mockRoute{kingdomOK: true}
	kp := &mockKingdom{}

	task := newTestTask(t, testConfig(), p, h, r, kp)
	if err := task.runWithOptions(fastRunOptions()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.stopCalls != 0 {
		t.Errorf("far distance should not stop venture, got %d stops", p.stopCalls)
	}
	if ready, remain := task.state.CheckFarWait(); ready || remain <= 0 {
		t.Errorf("far wait should be recorded, got ready=%v remain=%v", ready, remain)
	}
}

// 近距轮询：差 1 层 ≤ farGap → polling；轮询读层达标 → settle → 重新 setup → Done。
func TestSurveyPollingReachesTarget(t *testing.T) {
	p := &mockPage{domain: true, running: true, floor: 5, floorOK: true, stopOK: true, setupOK: true}
	reads := 0
	p.onRead = func() {
		reads++
		if reads >= 2 { // running 读第 1 次仍是 5，polling 第 1 次轮询即达标
			p.floor = 6
		}
	}
	p.onStop = func() { p.running = false; p.floor = 0 }
	h := &mockHome{waitOK: true}
	r := &mockRoute{kingdomOK: true}

	task := newTestTask(t, testConfig(), p, h, r, &mockKingdom{})
	if err := task.runWithOptions(fastRunOptions()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.stopCalls != 1 {
		t.Errorf("polling reach target should settle once, got %d stops", p.stopCalls)
	}
	if ready, remain := task.state.CheckFarWait(); ready || remain <= 0 {
		t.Errorf("after re-setup far wait should be recorded, got ready=%v remain=%v", ready, remain)
	}
}

// OCR 持续失败 → Retry 耗尽报错。
func TestSurveyRetriesWhenFloorOCRFails(t *testing.T) {
	p := &mockPage{domain: true, running: true, floorOK: false}
	task := newTestTask(t, testConfig(), p, &mockHome{}, &mockRoute{}, &mockKingdom{})
	task.sm.Init("running", statemachine.Options{MaxRetry: 1, MaxError: 3, Timeout: 2 * time.Second, RetryInterval: time.Millisecond})
	task.sm.Ctx = &surveyCtx{task: task, cfg: task.cfg}
	if err := task.sm.Run(task.handlers(), fastRunOptions()); err == nil {
		t.Fatal("expected retry exceeded when floor OCR keeps failing")
	}
}

// 页面完全未知 → detect Fatal。
func TestSurveyDetectFatalOnUnknownPage(t *testing.T) {
	p := &mockPage{}
	task := newTestTask(t, testConfig(), p, &mockHome{}, &mockRoute{}, &mockKingdom{})
	if err := task.runWithOptions(fastRunOptions()); err == nil {
		t.Fatal("expected fatal error on unknown page")
	}
}

// 导航：在矿山首页 → 点勘查入口 → 进入勘查域 → setup → farWait Done。
func TestSurveyNavigateFromMineHome(t *testing.T) {
	p := &mockPage{setupOK: true}
	h := &mockHome{current: true, waitOK: true}
	h.onVenture = func() { h.current = false; p.domain = true }
	r := &mockRoute{kingdomOK: true}

	task := newTestTask(t, testConfig(), p, h, r, &mockKingdom{})
	if err := task.runWithOptions(fastRunOptions()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if h.ventureCalls == 0 {
		t.Error("expected TapVenture on mine home")
	}
	if p.setupCalls != 1 {
		t.Errorf("expected Setup once, got %d", p.setupCalls)
	}
}

// Run 双保险：远距等待未到期时本轮直接跳过，不跑状态机。
func TestSurveyRunSkipsDuringFarWait(t *testing.T) {
	p := &mockPage{domain: true, running: true, floor: 6, floorOK: true}
	task := newTestTask(t, testConfig(), p, &mockHome{}, &mockRoute{}, &mockKingdom{})
	task.state.EnterFarWait(10 * time.Minute)
	if err := task.Run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.stopCalls != 0 || p.setupCalls != 0 {
		t.Error("state machine should not run during far wait")
	}
}
