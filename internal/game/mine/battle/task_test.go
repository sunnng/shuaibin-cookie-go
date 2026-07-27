package battle

import (
	"path/filepath"
	"testing"
	"time"

	"app/internal/platform/screen"
	"app/internal/statemachine"
	"app/internal/store"
)

type mockPage struct {
	battlePage   bool
	waitBattleOK bool
	quickBtn     *screen.Point
	dialogOK     bool
	dialogGoneOK bool
	used, owned  int
	clockOK      bool
	settleOK     bool
	cards        []screen.Point
	soulMatch    string
	lastPage     bool

	quickTapCalls  int
	confirmCalls   int
	cancelCalls    int
	cardTapCalls   int
	backCalls      int
	settleCalls    int
	recognizeCalls int

	onSettle func()
}

func (m *mockPage) IsBattlePage() bool { return m.battlePage }
func (m *mockPage) WaitBattlePage(timeout time.Duration) bool {
	return m.waitBattleOK
}
func (m *mockPage) TapBackBtn() { m.backCalls++ }
func (m *mockPage) FindQuickBattleButton() (screen.Point, bool) {
	if m.quickBtn == nil {
		return screen.Point{}, false
	}
	return *m.quickBtn, true
}
func (m *mockPage) TapQuickBattleButton(pt screen.Point) { m.quickTapCalls++ }
func (m *mockPage) WaitQuickBattleDialog(timeout time.Duration) bool {
	return m.dialogOK
}
func (m *mockPage) WaitQuickBattleDialogGone(timeout time.Duration) bool {
	return m.dialogGoneOK
}
func (m *mockPage) ReadClockCount() (int, int, bool) { return m.used, m.owned, m.clockOK }
func (m *mockPage) TapQuickBattleConfirm()           { m.confirmCalls++ }
func (m *mockPage) TapQuickBattleCancel()            { m.cancelCalls++ }
func (m *mockPage) TapSettleUntilBattlePage() bool {
	m.settleCalls++
	if m.onSettle != nil {
		m.onSettle()
	}
	return m.settleOK
}
func (m *mockPage) FindBattleCards() []screen.Point { return m.cards }
func (m *mockPage) TapBattleCard(pt screen.Point)   { m.cardTapCalls++ }
func (m *mockPage) RecognizeSoulStoneType(targets map[string]bool) string {
	m.recognizeCalls++
	return m.soulMatch
}
func (m *mockPage) SwipeUpAndCheckLastPage() bool { return m.lastPage }

type mockHome struct {
	current     bool
	waitOK      bool
	battleCalls int
	backCalls   int
	onTapBattle func()
	onTapBack   func()
}

func (m *mockHome) IsCurrent() bool                        { return m.current }
func (m *mockHome) WaitCurrent(timeout time.Duration) bool { return m.waitOK }
func (m *mockHome) TapBattle() {
	m.battleCalls++
	if m.onTapBattle != nil {
		m.onTapBattle()
	}
}
func (m *mockHome) TapBack() {
	m.backCalls++
	if m.onTapBack != nil {
		m.onTapBack()
	}
}

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
		Label:    "battle-test",
	}
}

func testConfig() *Config {
	return &Config{Enabled: true, IntervalSec: 21600, SoulStones: []string{"妖精王"}}
}

// 主路径：快转按钮 + 灵魂石匹配 + 发条充足 → 快转结算 → 再战无可操作 → 退出回城。
func TestBattleQuickBattleFlow(t *testing.T) {
	pt := screen.Point{X: 580, Y: 770}
	p := &mockPage{
		battlePage:   true,
		quickBtn:     &pt,
		soulMatch:    "妖精王",
		dialogOK:     true,
		used:         1,
		owned:        100,
		clockOK:      true,
		settleOK:     true,
		dialogGoneOK: true,
	}
	// 结算完成后：快转按钮消失、无战斗卡 → exit
	p.onSettle = func() { p.quickBtn = nil; p.cards = nil }
	h := &mockHome{waitOK: true}
	// exit：战斗页 TapBackBtn 后 WaitCurrent 命中；home.IsCurrent=false 跳过 TapBack；
	// kingdom.home=true 直接命中最终分支 → Done。
	kp := &mockKingdom{home: true, waitOK: true}

	task := newTestTask(t, testConfig(), p, h, &mockRoute{ok: true}, kp)
	if err := task.runWithOptions(fastRunOptions()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.quickTapCalls != 1 {
		t.Errorf("expected quick battle tap once, got %d", p.quickTapCalls)
	}
	if p.confirmCalls != 1 {
		t.Errorf("expected confirm once, got %d", p.confirmCalls)
	}
	if p.cancelCalls != 0 {
		t.Errorf("should not cancel, got %d", p.cancelCalls)
	}
}

// 发条不足 → 取消快转 → 退出。
func TestBattleCancelWhenClockInsufficient(t *testing.T) {
	pt := screen.Point{X: 580, Y: 770}
	p := &mockPage{
		battlePage:   true,
		quickBtn:     &pt,
		soulMatch:    "妖精王",
		dialogOK:     true,
		used:         5,
		owned:        2,
		clockOK:      true,
		dialogGoneOK: true,
	}
	h := &mockHome{waitOK: true}
	kp := &mockKingdom{home: true, waitOK: true}

	task := newTestTask(t, testConfig(), p, h, &mockRoute{ok: true}, kp)
	if err := task.runWithOptions(fastRunOptions()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.cancelCalls != 1 {
		t.Errorf("expected cancel once, got %d", p.cancelCalls)
	}
	if p.confirmCalls != 0 {
		t.Errorf("should not confirm, got %d", p.confirmCalls)
	}
}

// 无可快转：仅 1 张战斗卡 → 直接退出。
func TestBattleSingleCardExits(t *testing.T) {
	p := &mockPage{
		battlePage: true,
		cards:      []screen.Point{{X: 200, Y: 300}},
	}
	h := &mockHome{waitOK: true}
	kp := &mockKingdom{home: true, waitOK: true}

	task := newTestTask(t, testConfig(), p, h, &mockRoute{ok: true}, kp)
	if err := task.runWithOptions(fastRunOptions()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.cardTapCalls != 0 {
		t.Errorf("single card should not be tapped, got %d", p.cardTapCalls)
	}
}

// 5 张卡且已到末页 → 退出。
func TestBattleLastPageExits(t *testing.T) {
	cards := []screen.Point{{X: 1, Y: 1}, {X: 2, Y: 2}, {X: 3, Y: 3}, {X: 4, Y: 4}, {X: 5, Y: 5}}
	p := &mockPage{battlePage: true, cards: cards, lastPage: true}
	h := &mockHome{waitOK: true}
	kp := &mockKingdom{home: true, waitOK: true}

	task := newTestTask(t, testConfig(), p, h, &mockRoute{ok: true}, kp)
	if err := task.runWithOptions(fastRunOptions()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// 页面完全未知 → detect Fatal。
func TestBattleDetectFatalOnUnknownPage(t *testing.T) {
	p := &mockPage{}
	task := newTestTask(t, testConfig(), p, &mockHome{}, &mockRoute{}, &mockKingdom{})
	if err := task.runWithOptions(fastRunOptions()); err == nil {
		t.Fatal("expected fatal error on unknown page")
	}
}

// Run 记录战斗时间（即使状态机随后失败/成功）。
func TestBattleRunRecordsBattleTime(t *testing.T) {
	p := &mockPage{
		battlePage: true,
		cards:      []screen.Point{{X: 200, Y: 300}},
	}
	h := &mockHome{waitOK: true}
	kp := &mockKingdom{home: true, waitOK: true}
	task := newTestTask(t, testConfig(), p, h, &mockRoute{ok: true}, kp)
	if err := task.Run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if remain := task.state.GetTimeUntilNext(time.Hour); remain <= 0 {
		t.Error("Run should record battle time")
	}
}
