package mining

import (
	"path/filepath"
	"testing"
	"time"

	"app/internal/game/mine"
	"app/internal/statemachine"
	"app/internal/store"
)

type mockPage struct {
	miningPage   bool
	settlement   bool
	setup        bool
	waitMiningOK bool
	untilMatchOK bool

	completed  bool
	freeSlot   bool
	noMineCard bool
	startable  bool

	quotaCur, quotaMax int
	quotaOK            bool
	selectGot          int
	selectExhausted    bool
	confirmOK          bool
	autoStartOK        bool

	tapCompletedCalls int
	tapFreeCalls      int
	tapReadyCalls     int
	tapBackCalls      int
	autoStartCalls    int
	selectCalls       int

	onTapCompleted func()
	onSelect       func()
	onTapFree      func()
	onTapReady     func()
}

func (m *mockPage) IsMiningPage() bool                        { return m.miningPage }
func (m *mockPage) WaitMiningPage(timeout time.Duration) bool { return m.waitMiningOK }
func (m *mockPage) IsSetup() bool                             { return m.setup }
func (m *mockPage) IsSettlementRoute() bool                   { return m.settlement }
func (m *mockPage) TapUntilMatchMiningPage() bool             { return m.untilMatchOK }
func (m *mockPage) HasCompletedTask() bool                    { return m.completed }
func (m *mockPage) TapCompletedSlot() bool {
	if !m.completed {
		return false
	}
	m.tapCompletedCalls++
	if m.onTapCompleted != nil {
		m.onTapCompleted()
	}
	return true
}
func (m *mockPage) HasFreeSlot() bool { return m.freeSlot }
func (m *mockPage) TapFreeSlot() bool {
	if !m.freeSlot {
		return false
	}
	m.tapFreeCalls++
	if m.onTapFree != nil {
		m.onTapFree()
	}
	return true
}
func (m *mockPage) WasNoMineCard() bool    { return m.noMineCard }
func (m *mockPage) HasStartableCard() bool { return m.startable }
func (m *mockPage) TapReadySlot() bool {
	if !m.startable {
		return false
	}
	m.tapReadyCalls++
	if m.onTapReady != nil {
		m.onTapReady()
	}
	return true
}
func (m *mockPage) ReadChooseQuota() (int, int, bool) { return m.quotaCur, m.quotaMax, m.quotaOK }
func (m *mockPage) SelectTargetCards(target mine.ColorFind, need int, direction string) (int, bool) {
	m.selectCalls++
	if m.onSelect != nil {
		m.onSelect()
	}
	return m.selectGot, m.selectExhausted
}
func (m *mockPage) ConfirmCardSelection() bool { return m.confirmOK }
func (m *mockPage) AutoSelectCookieAndStart() bool {
	m.autoStartCalls++
	return m.autoStartOK
}
func (m *mockPage) TapBackBtn() { m.tapBackCalls++ }

type mockHome struct {
	current     bool
	goneOK      bool
	completed   bool
	miningCalls int
	onTapMining func()
}

func (m *mockHome) IsCurrent() bool                     { return m.current }
func (m *mockHome) WaitGone(timeout time.Duration) bool { return m.goneOK }
func (m *mockHome) HasCompletedMiningTask() bool        { return m.completed }
func (m *mockHome) TapMining() {
	m.miningCalls++
	if m.onTapMining != nil {
		m.onTapMining()
	}
}

type mockRoute struct {
	mineHomeOK bool
	returnOK   bool

	toMineCalls int
	returnCalls int
}

func (m *mockRoute) KingdomHomeToMineHome() bool { m.toMineCalls++; return m.mineHomeOK }
func (m *mockRoute) ReturnToKingdom() bool       { m.returnCalls++; return m.returnOK }

type mockKingdom struct{ home bool }

func (m *mockKingdom) IsKingdomHome() bool { return m.home }

func newTestTask(t *testing.T, cfg *Config, p page, h homePage, r route, kp kingdomPage) *Task {
	s := NewState(store.New(filepath.Join(t.TempDir(), "store.json")))
	return newTask(cfg, DefaultFeature(), p, h, r, kp, s, nil)
}

func fastRunOptions() statemachine.RunOptions {
	return statemachine.RunOptions{
		Interval: 1 * time.Millisecond,
		Label:    "mining-test",
	}
}

func testConfig() *Config {
	return &Config{Enabled: true, IntervalSec: 100, OreCards: []string{CardButterAmber}}
}

// 主路径：开采页无可操作项 → 复查仍无 → 回城记录 busy → Done。
func TestMiningDoneWhenNothingToDo(t *testing.T) {
	p := &mockPage{miningPage: true, quotaOK: true}
	r := &mockRoute{returnOK: true}
	task := newTestTask(t, testConfig(), p, &mockHome{}, r, &mockKingdom{})

	if err := task.runWithOptions(fastRunOptions()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.returnCalls != 1 {
		t.Errorf("expected ReturnToKingdom once, got %d", r.returnCalls)
	}
	if ready, remain := task.state.CheckReady(); ready || remain <= 0 {
		t.Errorf("busy wait should be recorded, got ready=%v remain=%v", ready, remain)
	}
}

// 已完成槽位 → 收奖励 → 再扫描无事 → done。
func TestMiningConfirmsRewards(t *testing.T) {
	p := &mockPage{miningPage: true, completed: true, untilMatchOK: true, quotaOK: true}
	p.onTapCompleted = func() { p.completed = false }
	r := &mockRoute{returnOK: true}
	task := newTestTask(t, testConfig(), p, &mockHome{}, r, &mockKingdom{})

	if err := task.runWithOptions(fastRunOptions()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.tapCompletedCalls != 1 {
		t.Errorf("expected TapCompletedSlot once, got %d", p.tapCompletedCalls)
	}
}

// 空闲栏位 → 选卡 → 确认 → 开始开采 → 再扫描无事 → done。
func TestMiningSelectCardAndStart(t *testing.T) {
	p := &mockPage{
		miningPage:   true,
		freeSlot:     true,
		quotaCur:     0,
		quotaMax:     1,
		quotaOK:      true,
		selectGot:    1,
		confirmOK:    true,
		waitMiningOK: true,
		autoStartOK:  true,
	}
	p.onSelect = func() { p.freeSlot = false; p.quotaCur = 1 }
	r := &mockRoute{returnOK: true}
	task := newTestTask(t, testConfig(), p, &mockHome{}, r, &mockKingdom{})

	if err := task.runWithOptions(fastRunOptions()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.tapFreeCalls != 1 {
		t.Errorf("expected TapFreeSlot once, got %d", p.tapFreeCalls)
	}
	if p.selectCalls != 1 {
		t.Errorf("expected SelectTargetCards once, got %d", p.selectCalls)
	}
	if p.autoStartCalls != 1 {
		t.Errorf("expected AutoSelectCookieAndStart once, got %d", p.autoStartCalls)
	}
}

// 可开采槽位 → 直接开始开采。
func TestMiningStartableSlot(t *testing.T) {
	p := &mockPage{miningPage: true, startable: true, autoStartOK: true, quotaOK: true}
	p.onTapReady = func() { p.startable = false } // 启动后可开采卡消失
	r := &mockRoute{returnOK: true}
	task := newTestTask(t, testConfig(), p, &mockHome{}, r, &mockKingdom{})

	if err := task.runWithOptions(fastRunOptions()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.tapReadyCalls != 1 {
		t.Errorf("expected TapReadySlot once, got %d", p.tapReadyCalls)
	}
	if p.autoStartCalls != 1 {
		t.Errorf("expected AutoSelectCookieAndStart once, got %d", p.autoStartCalls)
	}
}

// 选卡页一张都选不到 → 返回开采页并跳过一次选卡 → 下轮扫描不再进选卡 → done。
func TestMiningNoCardSelectedSkipsOnce(t *testing.T) {
	p := &mockPage{
		miningPage: true,
		freeSlot:   true,
		quotaCur:   0,
		quotaMax:   1,
		quotaOK:    true,
		selectGot:  0,
	}
	p.onSelect = func() {}                      // 配额不变
	p.onTapFree = func() { p.freeSlot = false } // 点过一次后空闲栏位不再出现，避免 done 复查死循环
	r := &mockRoute{returnOK: true}
	task := newTestTask(t, testConfig(), p, &mockHome{}, r, &mockKingdom{})

	if err := task.runWithOptions(fastRunOptions()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.tapBackCalls != 1 {
		t.Errorf("expected TapBackBtn once after empty selection, got %d", p.tapBackCalls)
	}
	if p.tapFreeCalls != 1 {
		t.Errorf("skipSelectOnce should prevent second TapFreeSlot, got %d", p.tapFreeCalls)
	}
}

// 空闲栏位点进去发现清单无矿卡 → noCardReturn 回城。
func TestMiningNoMineCardReturns(t *testing.T) {
	p := &mockPage{miningPage: true, freeSlot: false, noMineCard: true, quotaOK: true}
	// freeSlot=false 时 TapFreeSlot 返回 false，随后 WasNoMineCard=true → noCardReturn
	r := &mockRoute{returnOK: true}
	task := newTestTask(t, testConfig(), p, &mockHome{}, r, &mockKingdom{})

	if err := task.runWithOptions(fastRunOptions()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.returnCalls != 1 {
		t.Errorf("expected ReturnToKingdom once, got %d", r.returnCalls)
	}
}

// 从王国首页导航进矿山的完整链路。
func TestMiningNavigateFromKingdom(t *testing.T) {
	p := &mockPage{quotaOK: true}
	h := &mockHome{goneOK: true}
	r := &mockRoute{mineHomeOK: true, returnOK: true}
	kp := &mockKingdom{home: true}
	h.onTapMining = func() { p.miningPage = true } // precheck 点开采后进入开采页
	task := newTestTask(t, testConfig(), p, h, r, kp)

	if err := task.runWithOptions(fastRunOptions()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.toMineCalls != 1 {
		t.Errorf("expected KingdomHomeToMineHome once, got %d", r.toMineCalls)
	}
	if h.miningCalls != 1 {
		t.Errorf("expected TapMining once in precheck, got %d", h.miningCalls)
	}
}

// 页面完全未知 → detect Fatal。
func TestMiningDetectFatalOnUnknownPage(t *testing.T) {
	p := &mockPage{}
	task := newTestTask(t, testConfig(), p, &mockHome{}, &mockRoute{}, &mockKingdom{})
	if err := task.runWithOptions(fastRunOptions()); err == nil {
		t.Fatal("expected fatal error on unknown page")
	}
}

// 回城失败 → Fatal。
func TestMiningFatalWhenReturnFails(t *testing.T) {
	p := &mockPage{miningPage: true, quotaOK: true}
	r := &mockRoute{returnOK: false}
	task := newTestTask(t, testConfig(), p, &mockHome{}, r, &mockKingdom{})
	if err := task.runWithOptions(fastRunOptions()); err == nil {
		t.Fatal("expected fatal error when ReturnToKingdom fails")
	}
}
