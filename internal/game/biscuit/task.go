package biscuit

import (
	"time"

	"app/internal/platform/action"
	"app/internal/platform/screen"
	"app/internal/statemachine"
	"app/internal/status"
)

// TargetRule 槽位目标规则：需要一条同名且百分比 >= MinPercent 的副词条。
type TargetRule struct {
	Enabled    bool
	Name       string
	MinPercent float64
}

// SumRule 总和规则：同名词条取最高 Count 条，加和 >= MinSum 即毕业。
type SumRule struct {
	Enabled bool
	Name    string
	Count   int
	MinSum  float64
}

// Config 洗脆饼词条任务配置。默认值来自 Lua config.lua USER.biscuit。
type Config struct {
	Enabled  bool
	MaxRolls int          // 洗炼上限次数
	Targets  []TargetRule // 固定 4 条槽位规则
	SumRules []SumRule    // 总和规则（可选）
}

// DefaultConfig 对齐 Lua USER.biscuit 默认值：
// 冷却时间≥5%、会心≥6%，另两条空；总和规则攻击力取最高2条加和≥11。
func DefaultConfig() *Config {
	return &Config{
		Enabled:  false,
		MaxRolls: 500,
		Targets: []TargetRule{
			{Enabled: true, Name: "冷却时间", MinPercent: 5},
			{Enabled: true, Name: "会心", MinPercent: 6},
			{Enabled: false, Name: "", MinPercent: 0},
			{Enabled: false, Name: "", MinPercent: 0},
		},
		SumRules: []SumRule{
			{Enabled: true, Name: "攻击力", Count: 2, MinSum: 11},
		},
	}
}

// page is the interface required by Task to interact with the biscuit reroll UI.
// The public *Page type implements this interface; tests inject mocks.
type page interface {
	ReadEffects() []Effect
	TapReroll()
	ConfirmResetDialog() bool
	ConfirmSameDialog() bool
}

type Task struct {
	cfg        *Config
	page       page
	state      *State
	sm         *statemachine.Machine
	shouldStop func() bool
	reporter   *status.Reporter
}

func NewTask(
	cfg *Config,
	det screen.Detector,
	exec action.Executor,
	feature *Feature,
	state *State,
) *Task {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &Task{
		cfg:   cfg,
		page:  NewPage(det, exec, feature),
		state: state,
		sm:    statemachine.New(),
	}
}

// SetShouldStop wires a stop probe (typically runtime.IsStopped) so the
// task-level state machine can exit between ticks when the script stops.
func (t *Task) SetShouldStop(fn func() bool) {
	t.shouldStop = fn
}

// SetStatusReporter 接入任务状态上报（灵动岛展示洗炼进度），nil 表示不上报。
func (t *Task) SetStatusReporter(r *status.Reporter) {
	t.reporter = r
}

// pushStatus 把当前洗炼进度推给灵动岛；未接入上报时无操作。
func (t *Task) pushStatus() {
	if t.reporter == nil {
		return
	}
	t.reporter.Set(t.state.StatusText(t.cfg))
}

func (t *Task) Run() error {
	return t.runWithOptions(statemachine.RunOptions{
		Interval: 500 * time.Millisecond,
		Label:    "洗脆饼词条",
	})
}

// runWithOptions runs the biscuit state machine with the supplied run options.
// It is used by tests to drive the task with fast intervals.
func (t *Task) runWithOptions(opts statemachine.RunOptions) error {
	t.sm.Init("roll", statemachine.Options{
		MaxRetry:      3,
		MaxError:      3,
		Timeout:       time.Hour, // 500 次 × 每次约 1~3s 的点击/OCR 开销
		RetryInterval: 1 * time.Second,
	})
	t.sm.Ctx = &biscuitCtx{task: t, cfg: t.cfg}

	if opts.Label == "" {
		opts.Label = "洗脆饼词条"
	}
	if opts.ShouldStop == nil && t.shouldStop != nil {
		opts.ShouldStop = t.shouldStop
	}
	return t.sm.Run(t.handlers(), opts)
}

// newTask constructs a Task with injected page and state.
// It is intended for unit tests in package biscuit.
func newTask(cfg *Config, p page, state *State) *Task {
	return &Task{
		cfg:   cfg,
		page:  p,
		state: state,
		sm:    statemachine.New(),
	}
}
