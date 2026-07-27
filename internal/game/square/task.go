package square

import (
	"time"

	"app/internal/game/common/kingdom"
	"app/internal/guard"
	"app/internal/platform/action"
	"app/internal/platform/screen"
	"app/internal/statemachine"
	"app/internal/status"
)

// Config 布谷鸟广场用户配置（对齐 Lua config.lua USER.square）。
// 后续由编排者统一接线到 internal/config 与 UI 面板。
type Config struct {
	Enabled          bool // 是否启用
	DailyCap         int  // 每日奖励领取上限（默认 240）
	CheckIntervalSec int  // 一轮奖励结算所需的有效停留秒数（默认 60，下限 60）
	ChunkSec         int  // 单次挂机等待的睡眠粒度秒数（默认 10）
}

// DefaultConfig 默认值来自 Lua config.lua USER.square。
func DefaultConfig() *Config {
	return &Config{Enabled: true, DailyCap: 240, CheckIntervalSec: 60, ChunkSec: 10}
}

// dailyCap Lua: cfg().dailyCap or 240。Go 缺失字段零值为 0，故 <=0 按默认。
func (c *Config) dailyCap() int {
	if c == nil || c.DailyCap <= 0 {
		return 240
	}
	return c.DailyCap
}

const minStaySec = 60 // Lua MIN_STAY_SEC

// staySec Lua: max(MIN_STAY_SEC, checkIntervalSec or 60)。
func (c *Config) staySec() time.Duration {
	sec := minStaySec
	if c != nil && c.CheckIntervalSec > sec {
		sec = c.CheckIntervalSec
	}
	return time.Duration(sec) * time.Second
}

// chunkSec Lua: cfg().chunkSec or 10。
func (c *Config) chunkSec() time.Duration {
	sec := 10
	if c != nil && c.ChunkSec > 0 {
		sec = c.ChunkSec
	}
	return time.Duration(sec) * time.Second
}

// page 是 Task 与广场 UI 交互所需的窄接口（学 arena）。
// 公开的 *Page 实现该接口；测试注入 mock。
type page interface {
	IsSquare() bool
	IsLeaveDialog() bool
	WaitLeaveDialog(timeout time.Duration) bool
	TapBack()
	TapCloseDialog()
	TapClaimAll()
	TapUntilDialog() bool
	IsDailyRewardsMaxed() bool
	ReadRewardSum() (pending, total, sum int, ok bool)
	SleepMs(ms int)
}

// route 是 Task 进出广场所需的窄接口；公开的 *Route 实现该接口。
type route interface {
	EnsureSquare() bool
	OpenLeaveDialog() bool
	LeaveToKingdom(timeout time.Duration) bool
	Leave() bool
	IsSquareContext() bool
}

type Task struct {
	cfg        *Config
	page       page
	route      route
	state      *State
	sm         *statemachine.Machine
	guard      *guard.Guard
	shouldStop func() bool
	reporter   *status.Reporter
}

func NewTask(
	cfg *Config,
	det screen.Detector,
	exec action.Executor,
	feature *Feature,
	kingdomPage *kingdom.Page,
	state *State,
	g *guard.Guard,
) *Task {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	p := NewPage(det, exec, feature)
	r := NewRoute(p, kingdomPage)
	return &Task{
		cfg:   cfg,
		page:  p,
		route: r,
		state: state,
		sm:    statemachine.New(),
		guard: g,
	}
}

// SetShouldStop 接入停止探针（通常 runtime.IsStopped），状态机在 tick 间
// 与长睡眠分片中响应停止。
func (t *Task) SetShouldStop(fn func() bool) {
	t.shouldStop = fn
}

// SetStatusReporter 接入任务状态上报（灵动岛一行文案），nil 表示不上报。
func (t *Task) SetStatusReporter(r *status.Reporter) {
	t.reporter = r
}

// pushStatus 上报一行状态；未接入上报时无操作。
func (t *Task) pushStatus(text string) {
	if t.reporter == nil {
		return
	}
	t.reporter.Set("布谷鸟广场 " + text)
}

// Leave 其它任务运行前调用：暂停停留计时并离开广场回王国主城
// （Lua SquareTask.leaveForOtherTask；本轮有效停留进度保留在会话里，
// 对应 Lua 侧的 leaveSquare 语义：卡在广场时先离开再跑别的任务）。
func (t *Task) Leave() bool {
	t.state.PauseStay()
	return t.route.Leave()
}

// InSquareContext 是否正卡在广场页或离开弹窗；供编排者在其它任务运行前先判断，
// 避免对不在广场的情况误调 Leave（Route.Leave 对未知页面会 WaitHome 空等）。
func (t *Task) InSquareContext() bool {
	return t.route.IsSquareContext()
}

// interruptibleSleep 对齐 Lua Guard.sleep / StatusHud.countdownSleep：
// 分片睡眠，每片检查 shouldStop 并跑一遍 Guard（弹窗守卫）。
func (t *Task) interruptibleSleep(d time.Duration) {
	deadline := time.Now().Add(d)
	for {
		remain := time.Until(deadline)
		if remain <= 0 {
			return
		}
		if t.shouldStop != nil && t.shouldStop() {
			return
		}
		if t.guard != nil {
			t.guard.Check()
		}
		step := 500 * time.Millisecond
		if remain < step {
			step = remain
		}
		time.Sleep(step)
	}
}

// Run 执行一轮广场任务推进。对齐 Lua：每次 Run 最多睡一个 chunk 就返回，
// 停留进度持久化在 State，由调度器稍后再唤起，不阻塞其它任务。
func (t *Task) Run() error {
	return t.runWithOptions(statemachine.RunOptions{
		Interval: 500 * time.Millisecond,
		Label:    "布谷鸟广场",
	})
}

// runWithOptions 以指定选项驱动广场状态机；测试用短间隔。
func (t *Task) runWithOptions(opts statemachine.RunOptions) error {
	t.sm.Init("start", statemachine.Options{
		MaxRetry:      3,
		MaxError:      3,
		Timeout:       30 * time.Minute,
		RetryInterval: 1 * time.Second,
	})
	t.sm.Ctx = &squareCtx{task: t, cfg: t.cfg}

	if opts.Label == "" {
		opts.Label = "布谷鸟广场"
	}
	if t.guard != nil {
		opts.Guard = t.guard.Check
	}
	if opts.ShouldStop == nil && t.shouldStop != nil {
		opts.ShouldStop = t.shouldStop
	}
	return t.sm.Run(t.handlers(), opts)
}

// newTask 以注入的 page、route、state 构造 Task，供包内单测使用。
func newTask(cfg *Config, p page, r route, state *State, g *guard.Guard) *Task {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &Task{
		cfg:   cfg,
		page:  p,
		route: r,
		state: state,
		sm:    statemachine.New(),
		guard: g,
	}
}
