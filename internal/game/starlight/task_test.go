package starlight

import (
	"path/filepath"
	"testing"
	"time"

	"app/internal/platform/screen"
	"app/internal/statemachine"
	"app/internal/status"
	"app/internal/store"
)

// mockPage 按预设页面状态应答；manualWaitFails 控制 WaitManualPage 前 N 次失败。
type mockPage struct {
	home    bool
	manual  bool
	vanilla bool
	task    bool

	claimPt screen.Point
	claimOK bool

	manualWaitFails int

	tapManualCalls   int
	tapIslandCalls   int
	tapBackVanilla   int
	tapTaskBtnCalls  int
	tapClaimCalls    int
	settleCalls      int
	dismissCalls     int
	tapBackTaskCalls int
	tapBackKingdom   int
}

func (m *mockPage) IsHomePage() bool          { return m.home }
func (m *mockPage) IsManualPage() bool        { return m.manual }
func (m *mockPage) IsVanillaIslandPage() bool { return m.vanilla }
func (m *mockPage) IsTaskPage() bool          { return m.task }

func (m *mockPage) WaitHomePage(timeout time.Duration) bool    { return m.home }
func (m *mockPage) WaitVanillaIslandPage(t time.Duration) bool { return m.vanilla }
func (m *mockPage) WaitTaskPage(timeout time.Duration) bool    { return m.task }
func (m *mockPage) WaitManualPage(timeout time.Duration) bool {
	if m.manualWaitFails > 0 {
		m.manualWaitFails--
		return false
	}
	return m.manual
}

func (m *mockPage) TapSailingManual() bool { m.tapManualCalls++; return true }
func (m *mockPage) TapTaskBtn() bool       { m.tapTaskBtnCalls++; return true }
func (m *mockPage) TapBackToKingdom() bool { m.tapBackKingdom++; return true }
func (m *mockPage) TapLoginIsland() bool   { m.tapIslandCalls++; return true }
func (m *mockPage) TapBackFromVanilla() bool {
	m.tapBackVanilla++
	return true
}
func (m *mockPage) TapBackFromTask() bool { m.tapBackTaskCalls++; return true }

func (m *mockPage) FindClaimableBtn() (screen.Point, bool) { return m.claimPt, m.claimOK }
func (m *mockPage) TapClaimableBtn(pt screen.Point)        { m.tapClaimCalls++ }
func (m *mockPage) SettleAfterClaim(check func() bool)     { m.settleCalls++ }
func (m *mockPage) DismissRewardPopupIfNeeded()            { m.dismissCalls++ }

type mockRoute struct {
	ensureCalls int
	ensureOK    bool
	onEnsure    func() // 导航成功后的副作用（如：页面切到繁星岛首页）
}

func (m *mockRoute) IsStarlightHome() bool { return false }
func (m *mockRoute) EnsureHome() bool {
	m.ensureCalls++
	if m.ensureOK && m.onEnsure != nil {
		m.onEnsure()
	}
	return m.ensureOK
}

type mockKingdom struct {
	waitHomeOK bool
}

func (m *mockKingdom) WaitHome(timeout time.Duration) bool { return m.waitHomeOK }

func newTestTask(t *testing.T, p page, r route, k kingdomHome) *Task {
	t.Helper()
	s := NewState(store.New(filepath.Join(t.TempDir(), "store.json")))
	return newTask(nil, p, r, k, s, nil)
}

func fastRunOptions() statemachine.RunOptions {
	return statemachine.RunOptions{
		Interval: 1 * time.Millisecond,
		Label:    "starlight-test",
	}
}

// happyPath 已在繁星岛首页，一路走到领奖并返回王国。
func TestTaskHappyPathFromHome(t *testing.T) {
	p := &mockPage{home: true, manual: true, vanilla: true, task: true, claimPt: screen.Point{X: 500, Y: 300}, claimOK: true}
	r := &mockRoute{ensureOK: true}
	k := &mockKingdom{waitHomeOK: true}

	task := newTestTask(t, p, r, k)
	if err := task.runWithOptions(fastRunOptions()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if p.tapManualCalls != 1 || p.tapIslandCalls != 1 || p.tapBackVanilla != 1 ||
		p.tapTaskBtnCalls != 1 || p.tapBackTaskCalls != 1 || p.tapBackKingdom != 1 {
		t.Fatalf("each step should run once: %+v", p)
	}
	if p.tapClaimCalls != 1 || p.settleCalls != 1 || p.dismissCalls != 1 {
		t.Fatalf("claim flow should run once: claim=%d settle=%d dismiss=%d",
			p.tapClaimCalls, p.settleCalls, p.dismissCalls)
	}
	if !task.state.IsDoneToday() {
		t.Fatal("state should be marked done today")
	}
}

// 无可领奖按钮时也应标记完成并正常收尾（Lua 语义：claim 失败不阻断）。
func TestTaskNoClaimableStillDone(t *testing.T) {
	p := &mockPage{home: true, manual: true, vanilla: true, task: true, claimOK: false}
	r := &mockRoute{ensureOK: true}
	k := &mockKingdom{waitHomeOK: true}

	task := newTestTask(t, p, r, k)
	if err := task.runWithOptions(fastRunOptions()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.tapClaimCalls != 0 || p.settleCalls != 0 || p.dismissCalls != 0 {
		t.Fatalf("no claimable button must skip claim flow: %+v", p)
	}
	if !task.state.IsDoneToday() {
		t.Fatal("state should be marked done today even without claimable button")
	}
}

// 今日已完成：check 直接 Done，不做任何页面操作。
func TestTaskSkipsWhenDoneToday(t *testing.T) {
	p := &mockPage{home: true, manual: true, vanilla: true, task: true}
	r := &mockRoute{ensureOK: true}
	k := &mockKingdom{waitHomeOK: true}

	task := newTestTask(t, p, r, k)
	task.state.MarkDoneToday()

	if err := task.runWithOptions(fastRunOptions()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.tapManualCalls != 0 || r.ensureCalls != 0 {
		t.Fatalf("done-today task must not touch pages/route: taps=%d ensure=%d",
			p.tapManualCalls, r.ensureCalls)
	}
}

// 不在已知页面 → navigate（路由成功后进入 openManual）。
func TestTaskNavigatesFromUnknownPage(t *testing.T) {
	p := &mockPage{home: false, manual: false, vanilla: false, task: false, claimOK: false}
	// 导航成功后视为已按流程走到各页（mock 静态翻页）。
	r := &mockRoute{ensureOK: true, onEnsure: func() { p.home, p.manual, p.vanilla, p.task = true, true, true, true }}
	k := &mockKingdom{waitHomeOK: true}

	task := newTestTask(t, p, r, k)
	if err := task.runWithOptions(fastRunOptions()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.ensureCalls != 1 {
		t.Fatalf("want EnsureHome called once, got %d", r.ensureCalls)
	}
	if !task.state.IsDoneToday() {
		t.Fatal("state should be marked done today")
	}
}

// 导航失败是 Fatal（Lua 返回 false+错误信息），任务报错且不重试导航。
func TestTaskNavigateFailureIsFatal(t *testing.T) {
	p := &mockPage{}
	r := &mockRoute{ensureOK: false}
	k := &mockKingdom{waitHomeOK: true}

	task := newTestTask(t, p, r, k)
	err := task.runWithOptions(fastRunOptions())
	if err == nil {
		t.Fatal("want error when navigation fails")
	}
	if r.ensureCalls != 1 {
		t.Fatalf("fatal navigate must not retry, ensureCalls=%d", r.ensureCalls)
	}
}

// 页面等待超时属于可恢复失败：Retry 后成功。
func TestTaskRetriesAfterWaitTimeout(t *testing.T) {
	p := &mockPage{home: true, manual: true, vanilla: true, task: true, claimOK: false, manualWaitFails: 1}
	r := &mockRoute{ensureOK: true}
	k := &mockKingdom{waitHomeOK: true}

	task := newTestTask(t, p, r, k)
	if err := task.runWithOptions(fastRunOptions()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.tapManualCalls != 2 { // 首次等待超时 + 重试成功
		t.Fatalf("want TapSailingManual twice (1 retry), got %d", p.tapManualCalls)
	}
	if !task.state.IsDoneToday() {
		t.Fatal("state should be marked done today")
	}
}

// 页面等待持续超时：Retry 耗尽后报错，且不得标记完成。
func TestTaskRetryExceeded(t *testing.T) {
	p := &mockPage{home: true, manual: false, manualWaitFails: 10}
	r := &mockRoute{ensureOK: true}
	k := &mockKingdom{waitHomeOK: true}

	task := newTestTask(t, p, r, k)
	task.sm.Init("check", statemachine.Options{MaxRetry: 2, MaxError: 3, Timeout: 5 * time.Second, RetryInterval: time.Millisecond})
	task.sm.Ctx = &starlightCtx{task: task}
	err := task.sm.Run(task.handlers(), fastRunOptions())
	if err == nil {
		t.Fatal("want retry exceeded when manual page never appears")
	}
	if p.tapManualCalls != 3 { // 首次 + MaxRetry=2 次重试
		t.Fatalf("want 3 attempts, got %d", p.tapManualCalls)
	}
	if task.state.IsDoneToday() {
		t.Fatal("failed task must not be marked done")
	}
}

// 回王国首页失败：Retry 耗尽报错。
func TestTaskFinishKingdomWaitFails(t *testing.T) {
	p := &mockPage{home: true, manual: true, vanilla: true, task: true, claimOK: false}
	r := &mockRoute{ensureOK: true}
	k := &mockKingdom{waitHomeOK: false}

	task := newTestTask(t, p, r, k)
	task.sm.Init("check", statemachine.Options{MaxRetry: 1, MaxError: 3, Timeout: 5 * time.Second, RetryInterval: time.Millisecond})
	task.sm.Ctx = &starlightCtx{task: task}
	err := task.sm.Run(task.handlers(), fastRunOptions())
	if err == nil {
		t.Fatal("want error when kingdom home never reached")
	}
	// 注意：finish 之前 claimTask 已标记完成，Lua 语义同样如此（领奖即算完成）。
	if !task.state.IsDoneToday() {
		t.Fatal("claim already happened; done mark expected")
	}
}

// pushStatus 未接入上报时无操作；接入后写入步骤文本。
func TestTaskPushStatus(t *testing.T) {
	task := newTestTask(t, &mockPage{}, &mockRoute{}, &mockKingdom{})
	task.pushStatus("执行中…") // reporter 为 nil，不应 panic

	r := status.New()
	task.SetStatusReporter(r)
	task.pushStatus("打开航海手册…")
	if got := r.Text(); got != "梦幻繁星岛 打开航海手册…" {
		t.Fatalf("status = %q", got)
	}
}

// Run 结束后状态文本：成功 → 今日已完成；失败 → 失败。
func TestTaskStatusAfterRun(t *testing.T) {
	r := status.New()

	okTask := newTestTask(t,
		&mockPage{home: true, manual: true, vanilla: true, task: true},
		&mockRoute{ensureOK: true}, &mockKingdom{waitHomeOK: true})
	okTask.SetStatusReporter(r)
	if err := okTask.runWithOptions(fastRunOptions()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := r.Text(); got != "梦幻繁星岛 今日已完成" {
		t.Fatalf("status = %q, want 今日已完成", got)
	}

	failTask := newTestTask(t, &mockPage{}, &mockRoute{ensureOK: false}, &mockKingdom{})
	failTask.SetStatusReporter(r)
	if err := failTask.runWithOptions(fastRunOptions()); err == nil {
		t.Fatal("want navigation failure")
	}
	if got := r.Text(); got != "梦幻繁星岛 失败" {
		t.Fatalf("status = %q, want 失败", got)
	}
}
