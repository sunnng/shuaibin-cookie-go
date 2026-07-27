package biscuit

import (
	"path/filepath"
	"testing"
	"time"

	"app/internal/statemachine"
	"app/internal/status"
	"app/internal/store"
)

type mockPage struct {
	effects []Effect

	rerollCalls int
	resetCalls  int
	sameCalls   int
	resetDialog bool
	sameDialog  bool
}

func (m *mockPage) ReadEffects() []Effect { return m.effects }
func (m *mockPage) TapReroll()            { m.rerollCalls++ }
func (m *mockPage) ConfirmResetDialog() bool {
	m.resetCalls++
	return m.resetDialog
}
func (m *mockPage) ConfirmSameDialog() bool {
	m.sameCalls++
	return m.sameDialog
}

func newTestTask(t *testing.T, cfg *Config, p page) *Task {
	t.Helper()
	s := NewState(store.New(filepath.Join(t.TempDir(), "store.json")))
	return newTask(cfg, p, s)
}

func fastRunOptions() statemachine.RunOptions {
	return statemachine.RunOptions{
		Interval: 1 * time.Millisecond,
		Label:    "biscuit-test",
	}
}

func graduatingEffects() []Effect {
	return []Effect{
		{Name: "冷却时间", Value: 5.2},
		{Name: "生命值", Value: 3},
		{Name: "会心", Value: 6.8},
		{Name: "攻击力", Value: 4},
	}
}

func plainEffects() []Effect {
	return []Effect{
		{Name: "生命值", Value: 3},
		{Name: "防御力", Value: 5},
		{Name: "会心", Value: 3.7},
		{Name: "攻击力", Value: 4},
	}
}

// 首次读词条即毕业：不再点洗炼，置毕业标记并关闭 enabled（对齐 Lua 行为）。
func TestTaskGraduatesOnFirstRead(t *testing.T) {
	cfg := DefaultConfig()
	p := &mockPage{effects: graduatingEffects()}
	task := newTestTask(t, cfg, p)

	if err := task.runWithOptions(fastRunOptions()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.state.Rolls != 1 {
		t.Errorf("Rolls = %d, want 1", task.state.Rolls)
	}
	if p.rerollCalls != 0 {
		t.Errorf("rerollCalls = %d, want 0 (毕业前不洗)", p.rerollCalls)
	}
	if !task.state.IsGraduated() {
		t.Error("expected graduated flag persisted")
	}
	if cfg.Enabled {
		t.Error("expected cfg.Enabled=false after graduation (Lua UserConfig.set enabled=false)")
	}
}

// 词条一直不达标：洗到 MaxRolls 上限后正常结束，不置毕业标记。
func TestTaskStopsAtMaxRolls(t *testing.T) {
	cfg := DefaultConfig()
	cfg.MaxRolls = 3
	p := &mockPage{effects: plainEffects()}
	task := newTestTask(t, cfg, p)

	if err := task.runWithOptions(fastRunOptions()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.state.Rolls != 3 {
		t.Errorf("Rolls = %d, want 3", task.state.Rolls)
	}
	if p.rerollCalls != 3 {
		t.Errorf("rerollCalls = %d, want 3", p.rerollCalls)
	}
	if task.state.IsGraduated() {
		t.Error("should not graduate")
	}
	if got := task.state.StatusText(cfg); got != "洗脆饼 3/3 · 已达上限" {
		t.Errorf("status = %q", got)
	}
}

// 已毕业（含历史持久化标记）再次运行：直接结束，不读词条不洗炼。
func TestTaskSkipsWhenAlreadyGraduated(t *testing.T) {
	cfg := DefaultConfig()
	p := &mockPage{effects: plainEffects()}
	task := newTestTask(t, cfg, p)
	task.state.MarkGraduated()

	if err := task.runWithOptions(fastRunOptions()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.state.Rolls != 0 || p.rerollCalls != 0 {
		t.Errorf("Rolls=%d rerolls=%d, want 0/0", task.state.Rolls, p.rerollCalls)
	}
}

// 总和规则毕业：槽位不满足但攻击力 2 条加和达标。
func TestTaskGraduatesViaSumRule(t *testing.T) {
	cfg := DefaultConfig()
	p := &mockPage{effects: []Effect{
		{Name: "攻击力", Value: 5.9},
		{Name: "生命值", Value: 3},
		{Name: "攻击力", Value: 5.2},
		{Name: "会心", Value: 3.7},
	}}
	task := newTestTask(t, cfg, p)

	if err := task.runWithOptions(fastRunOptions()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !task.state.IsGraduated() {
		t.Error("expected graduation via sum rule")
	}
}

// 确认弹窗出现后本轮只处理一次（对齐 Lua isConfirmXxxDialog 标志）。
func TestTaskHandlesDialogsOncePerRun(t *testing.T) {
	cfg := DefaultConfig()
	cfg.MaxRolls = 3
	p := &mockPage{effects: plainEffects(), resetDialog: true, sameDialog: true}
	task := newTestTask(t, cfg, p)

	if err := task.runWithOptions(fastRunOptions()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.resetCalls != 1 || p.sameCalls != 1 {
		t.Errorf("resetCalls=%d sameCalls=%d, want 1/1 (处理后不再检测)", p.resetCalls, p.sameCalls)
	}
}

// shouldStop 在 tick 间生效，任务以 stopped 错误退出。
func TestTaskStopsViaShouldStop(t *testing.T) {
	cfg := DefaultConfig()
	cfg.MaxRolls = 500
	p := &mockPage{effects: plainEffects()}
	task := newTestTask(t, cfg, p)

	calls := 0
	task.SetShouldStop(func() bool {
		calls++
		return calls > 2
	})
	err := task.runWithOptions(fastRunOptions())
	if err == nil {
		t.Fatal("expected stopped error")
	}
	if task.state.IsGraduated() {
		t.Error("stopped run must not graduate")
	}
}

// pushStatus 未接入上报时无操作；接入后写入进度文本。
func TestTaskPushStatus(t *testing.T) {
	s := NewState(store.New(filepath.Join(t.TempDir(), "store.json")))
	s.Rolls = 7
	cfg := DefaultConfig()
	task := newTask(cfg, nil, s)
	task.pushStatus() // reporter 为 nil，不应 panic

	r := status.New()
	task.SetStatusReporter(r)
	task.pushStatus()
	if got := r.Text(); got != "洗脆饼 7/500" {
		t.Fatalf("status = %q", got)
	}
}
