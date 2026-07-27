package mining

import (
	"time"

	"app/internal/game/common/kingdom"
	"app/internal/game/mine"
	"app/internal/guard"
	"app/internal/platform/action"
	"app/internal/platform/screen"
	"app/internal/statemachine"
)

// Config 矿山开采配置（对应 Lua config.lua USER.mine 的开采字段）。
type Config struct {
	Enabled     bool     // miningEnabled，默认 false
	IntervalSec int      // miningIntervalSec，全忙后再次调度间隔秒，默认 1200
	OreCards    []string // miningOreCards，选卡优先级（键名同特征库 oreVeinCards）
}

func DefaultConfig() *Config {
	return &Config{
		Enabled:     false,
		IntervalSec: 1200,
		OreCards:    append([]string(nil), DefaultCardPriority...),
	}
}

// page is the interface required by Task to interact with the mining UI.
// The public *Page type implements this interface; tests inject mocks.
type page interface {
	IsMiningPage() bool
	WaitMiningPage(timeout time.Duration) bool
	IsSetup() bool
	IsSettlementRoute() bool
	TapUntilMatchMiningPage() bool
	HasCompletedTask() bool
	TapCompletedSlot() bool
	HasFreeSlot() bool
	TapFreeSlot() bool
	WasNoMineCard() bool
	HasStartableCard() bool
	TapReadySlot() bool
	ReadChooseQuota() (int, int, bool)
	SelectTargetCards(target mine.ColorFind, need int, direction string) (int, bool)
	ConfirmCardSelection() bool
	AutoSelectCookieAndStart() bool
	TapBackBtn()
}

// homePage 是 Task 需要的矿山首页窄接口（*mine.Page 实现）。
type homePage interface {
	IsCurrent() bool
	WaitGone(timeout time.Duration) bool
	HasCompletedMiningTask() bool
	TapMining()
}

// route 是 Task 需要的共享路由窄接口；ReturnToKingdom 已由适配器绑定本任务 page。
type route interface {
	KingdomHomeToMineHome() bool
	ReturnToKingdom() bool
}

// routeAdapter 把共享 *mine.Route 的 ReturnToKingdom(mp) 绑定到本任务的 *Page。
type routeAdapter struct {
	shared *mine.Route
	page   *Page
}

func (a routeAdapter) KingdomHomeToMineHome() bool { return a.shared.KingdomHomeToMineHome() }
func (a routeAdapter) ReturnToKingdom() bool       { return a.shared.ReturnToKingdom(a.page) }

// kingdomPage 是 Task 需要的王国页窄接口（*kingdom.Page 实现）。
type kingdomPage interface {
	IsKingdomHome() bool
}

type Task struct {
	cfg      *Config
	feature  *Feature
	page     page
	home     homePage
	route    route
	kingdom  kingdomPage
	state    *State
	sm       *statemachine.Machine
	guard    *guard.Guard
	stopFunc func() bool
}

func NewTask(
	cfg *Config,
	det screen.Detector,
	exec action.Executor,
	feature *Feature,
	home *mine.Page,
	sharedRoute *mine.Route,
	kingdomPage *kingdom.Page,
	state *State,
	g *guard.Guard,
) *Task {
	page := NewPage(det, exec, feature, home)
	return &Task{
		cfg:     cfg,
		feature: feature,
		page:    page,
		home:    home,
		route:   routeAdapter{shared: sharedRoute, page: page},
		kingdom: kingdomPage,
		state:   state,
		sm:      statemachine.New(),
		guard:   g,
	}
}

// SetShouldStop wires a stop probe (typically runtime.IsStopped) so the
// task-level state machine can exit between ticks when the script stops.
func (t *Task) SetShouldStop(fn func() bool) {
	t.stopFunc = fn
}

// CanResume 开采页或奖励页上可以直接续跑（Lua 开采 register 的 canResume 语义），
// 供编排者在调度侧决定是否直接拉起任务。
func (t *Task) CanResume() bool {
	mp, ok := t.page.(*Page)
	if !ok {
		return false
	}
	return mp.IsMiningPage() || mp.IsRewardPage()
}

func (t *Task) Run() error {
	return t.runWithOptions(statemachine.RunOptions{
		Interval: 500 * time.Millisecond,
		Label:    "矿山开采",
	})
}

// runWithOptions runs the mining state machine with the supplied run options.
// It is used by tests to drive the task with fast intervals.
func (t *Task) runWithOptions(opts statemachine.RunOptions) error {
	t.sm.Init("detect", statemachine.Options{
		MaxRetry:      3,
		MaxError:      3,
		Timeout:       30 * time.Minute,
		RetryInterval: 1 * time.Second,
	})
	t.sm.Ctx = &miningCtx{task: t, cfg: t.cfg, cardSwipeDirection: "left"}

	if opts.Label == "" {
		opts.Label = "矿山开采"
	}
	if t.guard != nil {
		opts.Guard = t.guard.Check
	}
	if opts.ShouldStop == nil && t.stopFunc != nil {
		opts.ShouldStop = t.stopFunc
	}
	return t.sm.Run(t.handlers(), opts)
}

// newTask constructs a Task with injected page, home, route, kingdom and state.
// It is intended for unit tests in package mining.
func newTask(cfg *Config, f *Feature, p page, h homePage, r route, kp kingdomPage, state *State, g *guard.Guard) *Task {
	return &Task{
		cfg:     cfg,
		feature: f,
		page:    p,
		home:    h,
		route:   r,
		kingdom: kp,
		state:   state,
		sm:      statemachine.New(),
		guard:   g,
	}
}

// resolveCardPriority 过滤出已取色的矿卡；配置为空或全部未取色时回退默认优先级
// （Lua resolveCardPriority）。
func resolveCardPriority(cfg *Config, cards map[string]mine.ColorFind) []string {
	configured := func(key string) bool {
		def, ok := cards[key]
		return ok && def.Colors != ""
	}
	var out []string
	if cfg != nil {
		for _, key := range cfg.OreCards {
			if configured(key) {
				out = append(out, key)
			}
		}
	}
	if len(out) == 0 {
		for _, key := range DefaultCardPriority {
			if configured(key) {
				out = append(out, key)
			}
		}
	}
	return out
}
