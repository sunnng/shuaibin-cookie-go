package jelly

import (
	"path/filepath"
	"testing"
	"time"

	"app/internal/platform/screen"
	"app/internal/statemachine"
	"app/internal/store"
)

type mockPage struct {
	jellyPage    bool
	waitJellyOK  bool
	claimAll     bool
	settleOK     bool
	configPt     *screen.Point
	waitConfigOK bool
	canChoose    bool
	remain       time.Duration
	remainOK     bool

	claimCalls      int
	configTapCalls  int
	chooseCalls     int
	backCalls       int
	configBackCalls int

	onClaim  func()
	onChoose func()
}

func (m *mockPage) IsJellyPage() bool { return m.jellyPage }
func (m *mockPage) WaitJellyPage(timeout time.Duration) bool {
	return m.waitJellyOK
}
func (m *mockPage) CanClaimAll() bool { return m.claimAll }
func (m *mockPage) TapClaimAll() {
	m.claimCalls++
	if m.onClaim != nil {
		m.onClaim()
	}
}
func (m *mockPage) TapSettleUntilJellyPage() bool { return m.settleOK }
func (m *mockPage) TapBack()                      { m.backCalls++ }
func (m *mockPage) FindConfigBtn() (screen.Point, bool) {
	if m.configPt == nil {
		return screen.Point{}, false
	}
	return *m.configPt, true
}
func (m *mockPage) TapConfigBtn(pt screen.Point) { m.configTapCalls++ }
func (m *mockPage) WaitConfigPage(timeout time.Duration) bool {
	return m.waitConfigOK
}
func (m *mockPage) CanChoose() bool { return m.canChoose }
func (m *mockPage) TapChoose() {
	m.chooseCalls++
	if m.onChoose != nil {
		m.onChoose()
	}
}
func (m *mockPage) TapConfigBack() { m.configBackCalls++ }
func (m *mockPage) ReadRemainTime() (time.Duration, bool) {
	return m.remain, m.remainOK
}

type mockHome struct {
	current    bool
	waitOK     bool
	jellyCalls int
	backCalls  int
}

func (m *mockHome) IsCurrent() bool                        { return m.current }
func (m *mockHome) WaitCurrent(timeout time.Duration) bool { return m.waitOK }
func (m *mockHome) TapJellyEntry()                         { m.jellyCalls++ }
func (m *mockHome) TapBack()                               { m.backCalls++ }

type mockRoute struct {
	ok          bool
	toMineCalls int
}

func (m *mockRoute) KingdomHomeToMineHome() bool { m.toMineCalls++; return m.ok }

type mockKingdom struct {
	home   bool
	waitOK bool
}

func (m *mockKingdom) IsKingdomHome() bool                 { return m.home }
func (m *mockKingdom) WaitHome(timeout time.Duration) bool { return m.waitOK }

func newTestTask(t *testing.T, cfg *Config, p page, h homePage, r route, kp kingdomPage) *Task {
	s := NewState(store.New(filepath.Join(t.TempDir(), "store.json")))
	return newTask(cfg, p, h, r, kp, s, nil)
}

func fastRunOptions() statemachine.RunOptions {
	return statemachine.RunOptions{
		Interval: 1 * time.Millisecond,
		Label:    "jelly-test",
	}
}

func testConfig() *Config {
	return &Config{Enabled: true, IntervalSec: 3600}
}

// 主路径：全部领取 → 配置洋菜冻 → 再处理读剩余时间 → 回城按剩余时间冷却。
func TestJellyFullFlow(t *testing.T) {
	pt := screen.Point{X: 800, Y: 620}
	p := &mockPage{
		jellyPage:    true,
		waitJellyOK:  true,
		claimAll:     true,
		settleOK:     true,
		configPt:     &pt,
		waitConfigOK: true,
		canChoose:    true,
		remain:       1800 * time.Second,
		remainOK:     true,
	}
	p.onClaim = func() { p.claimAll = false }
	p.onChoose = func() { p.configPt = nil; p.canChoose = false }
	h := &mockHome{current: true, waitOK: true}
	kp := &mockKingdom{home: true, waitOK: true}

	task := newTestTask(t, testConfig(), p, h, &mockRoute{ok: true}, kp)
	if err := task.runWithOptions(fastRunOptions()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.claimCalls != 1 {
		t.Errorf("expected claim-all once, got %d", p.claimCalls)
	}
	if p.chooseCalls != 1 {
		t.Errorf("expected choose once, got %d", p.chooseCalls)
	}
	ready, remain := task.state.CheckReady()
	if ready || remain < 1700*time.Second || remain > 1800*time.Second {
		t.Errorf("cooldown should follow remain time ~1800s, got ready=%v remain=%v", ready, remain)
	}
}

// 无配置按钮且剩余时间识别失败 → 按冷却间隔等待。
func TestJellyCooldownFallbackToInterval(t *testing.T) {
	p := &mockPage{
		jellyPage: true,
		remainOK:  false,
	}
	h := &mockHome{current: true, waitOK: true}
	kp := &mockKingdom{home: true, waitOK: true}

	task := newTestTask(t, testConfig(), p, h, &mockRoute{ok: true}, kp)
	if err := task.runWithOptions(fastRunOptions()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ready, remain := task.state.CheckReady()
	if ready || remain < 3500*time.Second || remain > 3600*time.Second {
		t.Errorf("cooldown should fall back to interval ~3600s, got ready=%v remain=%v", ready, remain)
	}
}

// 配置按钮点了但进不了配置页（无可选择洋菜冻）→ 直接回城。
func TestJellyConfigPageNotEntered(t *testing.T) {
	pt := screen.Point{X: 800, Y: 620}
	p := &mockPage{
		jellyPage:    true,
		configPt:     &pt,
		waitConfigOK: false,
		remain:       900 * time.Second,
		remainOK:     true,
	}
	h := &mockHome{current: true, waitOK: true}
	kp := &mockKingdom{home: true, waitOK: true}

	task := newTestTask(t, testConfig(), p, h, &mockRoute{ok: true}, kp)
	if err := task.runWithOptions(fastRunOptions()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.chooseCalls != 0 {
		t.Errorf("should not choose, got %d", p.chooseCalls)
	}
	// 未进配置页 → remain=0 → 按冷却间隔等待
	ready, remain := task.state.CheckReady()
	if ready || remain < 3500*time.Second {
		t.Errorf("cooldown should be interval, got ready=%v remain=%v", ready, remain)
	}
}

// 配置页不可选择 → 返回洋菜冻页 → 回城。
func TestJellyConfigNotChoosable(t *testing.T) {
	pt := screen.Point{X: 800, Y: 620}
	p := &mockPage{
		jellyPage:    true,
		waitJellyOK:  true,
		configPt:     &pt,
		waitConfigOK: true,
		canChoose:    false,
		remain:       600 * time.Second,
		remainOK:     true,
	}
	h := &mockHome{current: true, waitOK: true}
	kp := &mockKingdom{home: true, waitOK: true}

	task := newTestTask(t, testConfig(), p, h, &mockRoute{ok: true}, kp)
	if err := task.runWithOptions(fastRunOptions()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.configBackCalls != 1 {
		t.Errorf("expected config back once, got %d", p.configBackCalls)
	}
}

// 从王国首页导航进洋菜冻页。
func TestJellyNavigateFromKingdom(t *testing.T) {
	p := &mockPage{jellyPage: false, waitJellyOK: true, remainOK: false}
	h := &mockHome{current: false, waitOK: true}
	r := &mockRoute{ok: true}
	kp := &mockKingdom{home: true, waitOK: true}
	// navigate → enterJelly：点入口后 jellyPage 出现
	// detect: jellyPage false → home false → kingdom true → navigate →
	// KingdomHomeToMineHome ok → enterJelly → TapJellyEntry → WaitJellyPage ok → processPage
	// processPage: claimAll false, configPt nil → remain fail → returnHome →
	// IsJellyPage false → home.current? false → kingdom.home true → Done
	task := newTestTask(t, testConfig(), p, h, r, kp)
	if err := task.runWithOptions(fastRunOptions()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.toMineCalls != 1 {
		t.Errorf("expected KingdomHomeToMineHome once, got %d", r.toMineCalls)
	}
	if h.jellyCalls != 1 {
		t.Errorf("expected TapJellyEntry once, got %d", h.jellyCalls)
	}
}

// 页面完全未知 → detect Fatal。
func TestJellyDetectFatalOnUnknownPage(t *testing.T) {
	p := &mockPage{}
	task := newTestTask(t, testConfig(), p, &mockHome{}, &mockRoute{}, &mockKingdom{})
	if err := task.runWithOptions(fastRunOptions()); err == nil {
		t.Fatal("expected fatal error on unknown page")
	}
}
