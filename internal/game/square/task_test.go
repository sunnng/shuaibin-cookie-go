package square

import (
	"path/filepath"
	"testing"
	"time"

	"app/internal/statemachine"
	"app/internal/status"
	"app/internal/store"
)

// mockPage 实现 task 的窄 page 接口；SleepMs 无操作以保持测试快速。
type mockPage struct {
	square bool
	dialog bool
	maxed  bool

	pending int
	total   int
	sumOK   bool

	tapBackCalls         int
	tapCloseDialogCalls  int
	tapClaimAllCalls     int
	tapUntilDialogCalls  int
	waitLeaveDialogCalls int
}

func (m *mockPage) IsSquare() bool      { return m.square }
func (m *mockPage) IsLeaveDialog() bool { return m.dialog }
func (m *mockPage) TapBack()            { m.tapBackCalls++ }
func (m *mockPage) TapCloseDialog() {
	m.tapCloseDialogCalls++
	m.dialog = false
	m.square = true
}
func (m *mockPage) TapClaimAll()         { m.tapClaimAllCalls++ }
func (m *mockPage) TapUntilDialog() bool { m.tapUntilDialogCalls++; return true }
func (m *mockPage) IsDailyRewardsMaxed() bool {
	return m.dialog && m.maxed
}
func (m *mockPage) ReadRewardSum() (int, int, int, bool) {
	if !m.sumOK {
		return 0, 0, 0, false
	}
	return m.pending, m.total, m.pending + m.total, true
}
func (m *mockPage) SleepMs(ms int) {}
func (m *mockPage) WaitLeaveDialog(timeout time.Duration) bool {
	m.waitLeaveDialogCalls++
	return m.dialog
}

// mockRoute 实现 task 的窄 route 接口。
type mockRoute struct {
	ensureOK bool
	openOK   bool
	homeOK   bool
	leaveOK  bool
	inCtx    bool
	onOpen   func()

	ensureCalls int
	openCalls   int
	homeCalls   int
	leaveCalls  int
}

func (m *mockRoute) EnsureSquare() bool { m.ensureCalls++; return m.ensureOK }
func (m *mockRoute) OpenLeaveDialog() bool {
	m.openCalls++
	if m.openOK && m.onOpen != nil {
		m.onOpen()
	}
	return m.openOK
}
func (m *mockRoute) LeaveToKingdom(timeout time.Duration) bool {
	m.homeCalls++
	return m.homeOK
}
func (m *mockRoute) Leave() bool { m.leaveCalls++; return m.leaveOK }
func (m *mockRoute) IsSquareContext() bool {
	return m.inCtx
}

func newTestTask(t *testing.T, cfg *Config, p page, r route) *Task {
	t.Helper()
	s := NewState(store.New(filepath.Join(t.TempDir(), "store.json")))
	return newTask(cfg, p, r, s, nil)
}

func fastRunOptions() statemachine.RunOptions {
	return statemachine.RunOptions{
		Interval: time.Millisecond,
		Label:    "square-test",
	}
}

func testCfg() *Config {
	return &Config{Enabled: true, DailyCap: 240, CheckIntervalSec: 60, ChunkSec: 1}
}

func TestRunSkipsWhenDoneToday(t *testing.T) {
	p := &mockPage{}
	r := &mockRoute{ensureOK: true, openOK: true, homeOK: true}
	task := newTestTask(t, testCfg(), p, r)
	task.state.MarkDoneToday()

	if err := task.runWithOptions(fastRunOptions()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.ensureCalls != 0 || r.openCalls != 0 || r.homeCalls != 0 {
		t.Errorf("done today must not navigate: %+v", r)
	}
}

// 未初检 + 未达上限：开弹窗 → 读奖励 → 关弹窗 → 记初检并重置停留计时。
func TestRunFirstCheckBelowCap(t *testing.T) {
	p := &mockPage{square: true, pending: 30, total: 100, sumOK: true}
	r := &mockRoute{ensureOK: true, openOK: true, homeOK: true}
	r.onOpen = func() { p.dialog = true; p.square = false }
	task := newTestTask(t, testCfg(), p, r)

	if err := task.runWithOptions(fastRunOptions()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.openCalls != 1 {
		t.Errorf("expected OpenLeaveDialog once, got %d", r.openCalls)
	}
	if p.tapCloseDialogCalls != 1 {
		t.Errorf("expected TapCloseDialog once, got %d", p.tapCloseDialogCalls)
	}
	if !task.state.HasCheckedToday() {
		t.Error("expected checked-today flag")
	}
	if rem := task.state.StayRemaining(60 * time.Second); rem <= 55*time.Second {
		t.Errorf("stay timer should be reset, remaining=%v", rem)
	}
	if task.state.IsDoneToday() {
		t.Error("below cap must not mark done")
	}
	if p.tapClaimAllCalls != 0 {
		t.Errorf("below cap must not claim, got %d", p.tapClaimAllCalls)
	}
}

// 达到 dailyCap：一次领回 → 回王国主城 → 标记今日完成。
func TestRunClaimsWhenCapReached(t *testing.T) {
	p := &mockPage{square: false, dialog: true, pending: 40, total: 200, sumOK: true}
	r := &mockRoute{ensureOK: true, openOK: true, homeOK: true}
	task := newTestTask(t, testCfg(), p, r)

	if err := task.runWithOptions(fastRunOptions()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.tapClaimAllCalls != 1 {
		t.Errorf("expected TapClaimAll once, got %d", p.tapClaimAllCalls)
	}
	if p.tapUntilDialogCalls != 1 {
		t.Errorf("expected TapUntilDialog once, got %d", p.tapUntilDialogCalls)
	}
	if r.homeCalls != 1 {
		t.Errorf("expected LeaveToKingdom once, got %d", r.homeCalls)
	}
	if !task.state.IsDoneToday() {
		t.Error("cap reached should mark done today")
	}
	if task.state.IsActive() {
		t.Error("done should clear the active session")
	}
}

// 满额标识：不领奖直接结束。
func TestRunFinishWhenMaxed(t *testing.T) {
	p := &mockPage{dialog: true, maxed: true, sumOK: true, pending: 10, total: 10}
	r := &mockRoute{ensureOK: true, openOK: true, homeOK: true}
	task := newTestTask(t, testCfg(), p, r)

	if err := task.runWithOptions(fastRunOptions()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.tapClaimAllCalls != 0 {
		t.Errorf("maxed must not claim, got %d", p.tapClaimAllCalls)
	}
	if !task.state.IsDoneToday() {
		t.Error("maxed should mark done today")
	}
}

// 奖励 OCR 两次都失败：报错，不标记完成。
func TestRunFailsWhenRewardOCRUnreadable(t *testing.T) {
	p := &mockPage{dialog: true, sumOK: false}
	r := &mockRoute{ensureOK: true, openOK: true, homeOK: true}
	task := newTestTask(t, testCfg(), p, r)

	if err := task.runWithOptions(fastRunOptions()); err == nil {
		t.Fatal("expected error when reward OCR keeps failing")
	}
	if task.state.IsDoneToday() {
		t.Error("OCR failure must not mark done")
	}
	if r.homeCalls != 0 {
		t.Errorf("OCR failure must not leave, got %d", r.homeCalls)
	}
}

// 已初检 + 停留未满：睡一个 chunk 后返回，不做任何弹窗动作。
func TestRunAccumulateChunk(t *testing.T) {
	p := &mockPage{square: true, sumOK: true}
	r := &mockRoute{ensureOK: true, openOK: true, homeOK: true}
	task := newTestTask(t, testCfg(), p, r)
	task.state.MarkCheckedToday()
	task.state.ResetStayTimer()

	start := time.Now()
	if err := task.runWithOptions(fastRunOptions()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if elapsed := time.Since(start); elapsed < 900*time.Millisecond {
		t.Errorf("expected ~1s chunk sleep, got %v", elapsed)
	}
	if r.openCalls != 0 {
		t.Errorf("accumulate must not open dialog, got %d", r.openCalls)
	}
	if task.state.IsDoneToday() {
		t.Error("accumulate must not mark done")
	}
	if !task.state.IsActive() {
		t.Error("session should stay active while accumulating")
	}
}

// 已初检 + 停留已满：跳过 chunk 睡眠，直接开弹窗检查。
func TestRunAccumulateElapsedOpensDialog(t *testing.T) {
	p := &mockPage{square: true, pending: 5, total: 5, sumOK: true}
	r := &mockRoute{ensureOK: true, openOK: true, homeOK: true}
	r.onOpen = func() { p.dialog = true; p.square = false }
	task := newTestTask(t, testCfg(), p, r)
	task.state.MarkCheckedToday()
	a := task.state.Ensure()
	a.AccumulatedSec = 120 // 超过 staySec=60
	a.LastEnterAt = 0
	task.state.save(a)

	start := time.Now()
	if err := task.runWithOptions(fastRunOptions()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if elapsed := time.Since(start); elapsed > 500*time.Millisecond {
		t.Errorf("elapsed stay should skip chunk sleep, took %v", elapsed)
	}
	if r.openCalls != 1 {
		t.Errorf("expected OpenLeaveDialog once, got %d", r.openCalls)
	}
}

// 无法进入广场：报错并暂停停留计时。
func TestRunFailsWhenCannotEnter(t *testing.T) {
	p := &mockPage{}
	r := &mockRoute{ensureOK: false}
	task := newTestTask(t, testCfg(), p, r)
	a := task.state.Ensure()
	a.LastEnterAt = time.Now().Unix()
	task.state.save(a)

	if err := task.runWithOptions(fastRunOptions()); err == nil {
		t.Fatal("expected error when square unreachable")
	}
	a, _ = task.state.GetActive()
	if a.LastEnterAt != 0 {
		t.Error("failed enter should pause the stay clock")
	}
}

// 弹窗打不开：报错。
func TestRunFailsWhenDialogNotAppearing(t *testing.T) {
	p := &mockPage{square: true}
	r := &mockRoute{ensureOK: true, openOK: false}
	task := newTestTask(t, testCfg(), p, r)

	if err := task.runWithOptions(fastRunOptions()); err == nil {
		t.Fatal("expected error when leave dialog never appears")
	}
}

// Leave（leaveForOtherTask）：暂停停留计时并委托 route.Leave。
func TestTaskLeave(t *testing.T) {
	p := &mockPage{}
	r := &mockRoute{leaveOK: true}
	task := newTestTask(t, testCfg(), p, r)
	task.state.StartStay()

	if !task.Leave() {
		t.Fatal("Leave should succeed")
	}
	if r.leaveCalls != 1 {
		t.Errorf("expected route.Leave once, got %d", r.leaveCalls)
	}
	a, _ := task.state.GetActive()
	if a.LastEnterAt != 0 {
		t.Error("Leave should pause the stay clock")
	}
}

// interruptibleSleep 响应 shouldStop 提前返回。
func TestInterruptibleSleepStopsEarly(t *testing.T) {
	task := newTestTask(t, testCfg(), &mockPage{}, &mockRoute{})
	stopAt := time.Now().Add(200 * time.Millisecond)
	task.SetShouldStop(func() bool { return time.Now().After(stopAt) })

	start := time.Now()
	task.interruptibleSleep(10 * time.Second)
	if elapsed := time.Since(start); elapsed > 2*time.Second {
		t.Errorf("should stop early, took %v", elapsed)
	}
}

// pushStatus 未接入上报时无操作；接入后带任务名前缀。
func TestTaskPushStatus(t *testing.T) {
	task := newTestTask(t, testCfg(), &mockPage{}, &mockRoute{})
	task.pushStatus("执行中…") // reporter 为 nil，不应 panic

	r := status.New()
	task.SetStatusReporter(r)
	task.pushStatus("执行中…")
	if got := r.Text(); got != "布谷鸟广场 执行中…" {
		t.Fatalf("status = %q", got)
	}
}

func TestStayProgressText(t *testing.T) {
	s := NewState(store.New(filepath.Join(t.TempDir(), "store.json")))
	if got := stayProgressText(s, 60*time.Second); got != "有效停留 0s/60s" {
		t.Fatalf("fresh = %q", got)
	}
	a := s.Ensure()
	a.AccumulatedSec = 45
	s.save(a)
	if got := stayProgressText(s, 60*time.Second); got != "有效停留 45s/60s" {
		t.Fatalf("mid = %q", got)
	}
	a.AccumulatedSec = 60
	s.save(a)
	if got := stayProgressText(s, 60*time.Second); got != "可开弹窗查看奖励" {
		t.Fatalf("full = %q", got)
	}
}
