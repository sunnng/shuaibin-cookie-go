package market

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"app/internal/statemachine"
	"app/internal/status"
	"app/internal/store"
)

type mockPage struct {
	current     bool
	ensureTabOK bool
	freeRefresh bool
	restockSec  int
	restockRaw  string
	restockOK   bool
	stats       PurchaseStats
	stockKeys   []string

	tapRefreshCalls int
	purchaseCalls   int
	purchaseItems   []string
}

func (m *mockPage) IsCurrent() bool                        { return m.current }
func (m *mockPage) WaitCurrent(timeout time.Duration) bool { return m.current }
func (m *mockPage) EnsureItemTab() bool                    { return m.ensureTabOK }
func (m *mockPage) IsFreeRefresh() bool                    { return m.freeRefresh }
func (m *mockPage) TapRefresh() {
	m.tapRefreshCalls++
	m.freeRefresh = false
}
func (m *mockPage) ReadRestockSeconds() (int, string, bool) {
	return m.restockSec, m.restockRaw, m.restockOK
}
func (m *mockPage) StockKeys() []string { return m.stockKeys }
func (m *mockPage) PurchaseWishlist(items []string) PurchaseStats {
	m.purchaseCalls++
	m.purchaseItems = items
	return m.stats
}

type mockRoute struct {
	enterCalls int
	leaveCalls int
	enterOK    bool
	leaveOK    bool
}

func (m *mockRoute) Enter() bool { m.enterCalls++; return m.enterOK }
func (m *mockRoute) Leave() bool { m.leaveCalls++; return m.leaveOK }

func newTestTask(t *testing.T, cfg *Config, p page, r route) *Task {
	t.Helper()
	s := NewState(store.New(filepath.Join(t.TempDir(), "store.json")))
	return newTask(cfg, p, r, s, nil)
}

func fastRunOptions() statemachine.RunOptions {
	return statemachine.RunOptions{
		Interval: 1 * time.Millisecond,
		Label:    "market-test",
	}
}

func testConfig() *Config {
	return &Config{Enabled: true, Items: []string{"灿烂的光之碎片"}, RestockBufferSec: 30}
}

// 主路径：免费刷新 → 扫货 → 按补货倒计时写调度 → 离场。
func TestMarketTaskHappyPath(t *testing.T) {
	cfg := testConfig()
	p := &mockPage{
		current:     true,
		ensureTabOK: true,
		freeRefresh: true,
		restockSec:  3600,
		restockRaw:  "01:00:00",
		restockOK:   true,
		stats:       PurchaseStats{Purchased: 2, SoldOut: 1},
	}
	r := &mockRoute{leaveOK: true}

	task := newTestTask(t, cfg, p, r)
	reporter := status.New()
	task.SetStatusReporter(reporter)

	if err := task.runWithOptions(fastRunOptions()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.tapRefreshCalls != 1 {
		t.Errorf("TapRefresh = %d, want 1", p.tapRefreshCalls)
	}
	if p.purchaseCalls != 1 {
		t.Errorf("PurchaseWishlist = %d, want 1", p.purchaseCalls)
	}
	if len(p.purchaseItems) != 1 || p.purchaseItems[0] != "灿烂的光之碎片" {
		t.Errorf("purchase items = %v", p.purchaseItems)
	}
	if task.state.Purchased != 2 || task.state.SoldOut != 1 {
		t.Errorf("state stats = %+v", task.state)
	}
	remain := task.state.TimeUntilRestock()
	if remain < 3600*time.Second || remain > 3700*time.Second {
		t.Errorf("TimeUntilRestock = %v, want ~3630s", remain)
	}
	if r.leaveCalls != 1 {
		t.Errorf("Leave = %d, want 1", r.leaveCalls)
	}
	if got := reporter.Text(); got != "交易所 购2 · 售罄1" {
		t.Errorf("status = %q", got)
	}
}

// 冷却中且非首轮：按 OCR 倒计时推迟，不扫货。
func TestMarketTaskDefersWhenCooldown(t *testing.T) {
	cfg := testConfig()
	p := &mockPage{
		current:     true,
		ensureTabOK: true,
		freeRefresh: false,
		restockSec:  1800,
		restockRaw:  "00:30:00",
		restockOK:   true,
	}
	r := &mockRoute{leaveOK: true}

	task := newTestTask(t, cfg, p, r)
	if err := task.runWithOptions(fastRunOptions()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.purchaseCalls != 0 {
		t.Errorf("cooldown should not purchase, got %d", p.purchaseCalls)
	}
	remain := task.state.TimeUntilRestock()
	if remain < 1800*time.Second || remain > 1900*time.Second {
		t.Errorf("TimeUntilRestock = %v, want ~1830s", remain)
	}
	if r.leaveCalls != 1 {
		t.Errorf("Leave = %d, want 1", r.leaveCalls)
	}
}

// 启动首轮强制：即使页面显示补货倒计时也照常扫货。
func TestMarketTaskStartupBypassIgnoresCooldown(t *testing.T) {
	cfg := testConfig()
	p := &mockPage{
		current:     true,
		ensureTabOK: true,
		freeRefresh: false,
		restockSec:  1800,
		restockRaw:  "00:30:00",
		restockOK:   true,
		stats:       PurchaseStats{Purchased: 1},
	}
	r := &mockRoute{leaveOK: true}

	task := newTestTask(t, cfg, p, r)
	task.state.CheckReady() // 武装首轮强制（模拟调度器首次检查）

	if err := task.runWithOptions(fastRunOptions()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.purchaseCalls != 1 {
		t.Errorf("forced first run should purchase, got %d", p.purchaseCalls)
	}
	if task.state.Purchased != 1 {
		t.Errorf("state.Purchased = %d, want 1", task.state.Purchased)
	}
}

// 确认 Tab 持续失败：Retry 耗尽报错，不扫货不离场。
func TestMarketTaskRetriesWhenEnsureTabFails(t *testing.T) {
	cfg := testConfig()
	p := &mockPage{current: true, ensureTabOK: false}
	r := &mockRoute{leaveOK: true}

	task := newTestTask(t, cfg, p, r)
	task.sm.Init("detect", statemachine.Options{MaxRetry: 1, MaxError: 3, Timeout: 2 * time.Second, RetryInterval: time.Millisecond})
	task.sm.Ctx = &marketCtx{task: task, cfg: cfg}
	err := task.sm.Run(task.handlers(), fastRunOptions())
	if err == nil {
		t.Fatal("expected retry exceeded when ensureItemTab keeps failing")
	}
	if p.purchaseCalls != 0 {
		t.Errorf("no purchase expected, got %d", p.purchaseCalls)
	}
	if r.leaveCalls != 0 {
		t.Errorf("no leave expected, got %d", r.leaveCalls)
	}
}

// 无法识别当前页面（不在交易所也不在王国页）：Fatal。
func TestMarketTaskFatalWhenPageUnknown(t *testing.T) {
	cfg := testConfig()
	p := &mockPage{current: false}
	r := &mockRoute{leaveOK: true}

	task := newTestTask(t, cfg, p, r) // kingdomPage 为 nil
	err := task.runWithOptions(fastRunOptions())
	if err == nil || !strings.Contains(err.Error(), "无法识别当前页面") {
		t.Fatalf("want fatal 无法识别当前页面, got %v", err)
	}
}

// 离场失败：Fatal。
func TestMarketTaskFatalWhenLeaveFails(t *testing.T) {
	cfg := testConfig()
	p := &mockPage{current: true, ensureTabOK: true, freeRefresh: true, restockOK: false}
	r := &mockRoute{leaveOK: false}

	task := newTestTask(t, cfg, p, r)
	err := task.runWithOptions(fastRunOptions())
	if err == nil || !strings.Contains(err.Error(), "离开交易所失败") {
		t.Fatalf("want fatal 离开交易所失败, got %v", err)
	}
}

// 补货读数失败：仅告警不写调度，照常离场完成。
func TestMarketTaskSkipsScheduleWhenRestockUnreadable(t *testing.T) {
	cfg := testConfig()
	p := &mockPage{current: true, ensureTabOK: true, freeRefresh: true, restockOK: false}
	r := &mockRoute{leaveOK: true}

	task := newTestTask(t, cfg, p, r)
	if err := task.runWithOptions(fastRunOptions()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := task.state.TimeUntilRestock(); got != 0 {
		t.Errorf("no schedule should be written, remain=%v", got)
	}
}
