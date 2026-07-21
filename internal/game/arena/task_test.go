package arena

import (
	"path/filepath"
	"testing"
	"time"

	"app/internal/config"
	"app/internal/guard"
	"app/internal/platform/action"
	"app/internal/platform/screen"
	"app/internal/statemachine"
	"app/internal/status"
	"app/internal/store"
)

type mockPage struct {
	lobby        bool
	tickets      int
	freeRefresh  bool
	opponent     *OpponentInfo
	battleResult string
	battleOK     bool

	trophyCount       int
	trophyOK          bool
	useTrophyOverride bool

	countdown            time.Duration
	countdownOK          bool
	useCountdownOverride bool

	hasTeamSelect bool
	teamSelectOK  bool

	buyTicketCalls      int
	runBattleCalls      int
	swipeCalls          int
	tapFreeRefreshCalls int
	tapOpponentCalls    int
	tapStartCalls       int
}

func (m *mockPage) IsLobby() bool                        { return m.lobby }
func (m *mockPage) WaitLobby(timeout time.Duration) bool { return m.lobby }
func (m *mockPage) ReadMedalAndTicket() (int, int, bool) { return 0, m.tickets, true }
func (m *mockPage) ReadTrophyCount() (int, bool) {
	if m.useTrophyOverride {
		return m.trophyCount, m.trophyOK
	}
	return 1000, true
}
func (m *mockPage) FindFirstValidOpponent(cfg *config.Arena, myTrophy int) *OpponentInfo {
	return m.opponent
}
func (m *mockPage) SwipePageLeft()      { m.swipeCalls++ }
func (m *mockPage) IsFreeRefresh() bool { return m.freeRefresh }
func (m *mockPage) TapFreeRefresh()     { m.tapFreeRefreshCalls++; m.freeRefresh = false }
func (m *mockPage) ReadRefreshCountdown() (time.Duration, bool) {
	if m.useCountdownOverride {
		return m.countdown, m.countdownOK
	}
	return 0, false
}
func (m *mockPage) BuyTicket() { m.buyTicketCalls++; m.tickets++ }
func (m *mockPage) RunBattle() (string, bool) {
	m.runBattleCalls++
	if m.tickets > 0 {
		m.tickets--
	}
	return m.battleResult, m.battleOK
}
func (m *mockPage) TapToLobby() bool { return true }
func (m *mockPage) TapOpponentSite(site action.Point) {
	m.tapOpponentCalls++
}
func (m *mockPage) HasTeamSelectPage() bool { return m.hasTeamSelect }
func (m *mockPage) WaitTeamSelect(timeout time.Duration) bool {
	return m.teamSelectOK
}
func (m *mockPage) TapStartBattle() { m.tapStartCalls++ }

type mockRoute struct {
	enterCalls int
	leaveCalls int
	enterOK    bool
	leaveOK    bool
}

func (m *mockRoute) Enter() bool { m.enterCalls++; return m.enterOK }
func (m *mockRoute) Leave() bool { m.leaveCalls++; return m.leaveOK }

func newTestTask(t *testing.T, cfg *config.Arena, p page, r route) *Task {
	s := NewState(store.New(filepath.Join(t.TempDir(), "store.json")))
	return newTask(cfg, p, r, s, nil)
}

func fastRunOptions() statemachine.RunOptions {
	return statemachine.RunOptions{
		Interval: 1 * time.Millisecond,
		Label:    "arena-test",
	}
}

func TestArenaTaskLeavesWhenNoTicketsAndNoAutoBuy(t *testing.T) {
	cfg := &config.Arena{Enabled: true, AutoBuyCount: 0}
	p := &mockPage{lobby: true, tickets: 0, freeRefresh: false}
	r := &mockRoute{leaveOK: true}

	task := newTestTask(t, cfg, p, r)
	if err := task.runWithOptions(fastRunOptions()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if task.state.Tickets != 0 {
		t.Errorf("expected tickets to stay 0, got %d", task.state.Tickets)
	}
	if task.state.BuyCount != 0 {
		t.Errorf("expected no ticket purchase, got buyCount=%d", task.state.BuyCount)
	}
	if r.leaveCalls == 0 {
		t.Error("expected route.Leave to be called")
	}
	if p.buyTicketCalls != 0 {
		t.Errorf("expected BuyTicket not called, got %d", p.buyTicketCalls)
	}
}

func TestArenaTaskBuysTicketWhenAllowed(t *testing.T) {
	cfg := &config.Arena{Enabled: true, AutoBuyCount: 1}
	p := &mockPage{lobby: true, tickets: 0, freeRefresh: false, opponent: nil}
	r := &mockRoute{leaveOK: true}

	task := newTestTask(t, cfg, p, r)
	if err := task.runWithOptions(fastRunOptions()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if p.buyTicketCalls != 1 {
		t.Errorf("expected BuyTicket called once, got %d", p.buyTicketCalls)
	}
	if task.state.BuyCount != 1 {
		t.Errorf("expected buyCount=1, got %d", task.state.BuyCount)
	}
	if task.state.Tickets != 1 {
		t.Errorf("expected tickets=1 after purchase, got %d", task.state.Tickets)
	}
	if r.leaveCalls == 0 {
		t.Error("expected route.Leave to be called after failing to find opponent")
	}
}

func TestArenaTaskRunsBattleWhenTicketsAvailable(t *testing.T) {
	cfg := &config.Arena{Enabled: true, AutoBuyCount: 0}
	p := &mockPage{
		lobby:        true,
		tickets:      1,
		opponent:     &OpponentInfo{Site: action.Point{X: 500, Y: 400}},
		battleResult: "胜利",
		battleOK:     true,
	}
	r := &mockRoute{leaveOK: true}

	task := newTestTask(t, cfg, p, r)
	if err := task.runWithOptions(fastRunOptions()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if p.tapOpponentCalls != 1 {
		t.Errorf("expected TapOpponentSite once, got %d", p.tapOpponentCalls)
	}
	if p.tapStartCalls != 1 {
		t.Errorf("expected TapStartBattle once, got %d", p.tapStartCalls)
	}
	if p.runBattleCalls != 1 {
		t.Errorf("expected RunBattle called once, got %d", p.runBattleCalls)
	}
	if task.state.TotalBattles() != 1 {
		t.Errorf("expected total battles=1, got %d", task.state.TotalBattles())
	}
	if task.state.Wins != 1 {
		t.Errorf("expected wins=1, got %d", task.state.Wins)
	}
	if r.leaveCalls == 0 {
		t.Error("expected route.Leave to be called")
	}
}

func TestArenaTaskRetriesWhenTrophyOCRFails(t *testing.T) {
	cfg := &config.Arena{Enabled: true, AutoBuyCount: 0}
	p := &mockPage{lobby: true, tickets: 0, useTrophyOverride: true, trophyOK: false, trophyCount: 0}
	r := &mockRoute{leaveOK: true}

	task := newTestTask(t, cfg, p, r)
	task.sm.Init("detect", statemachine.Options{MaxRetry: 1, MaxError: 3, Timeout: 2 * time.Second})
	task.sm.Ctx = &arenaCtx{task: task, cfg: cfg}
	err := task.sm.Run(task.handlers(), fastRunOptions())
	if err == nil {
		t.Fatal("expected retry exceeded when trophy OCR keeps failing")
	}
	if task.state.Trophies != 0 {
		t.Errorf("failed OCR must not write trophies=0 as success, got %d", task.state.Trophies)
	}
}

func TestArenaTaskBackoffWhenCountdownOCRFails(t *testing.T) {
	cfg := &config.Arena{Enabled: true, AutoBuyCount: 0}
	p := &mockPage{
		lobby:                true,
		tickets:              1,
		opponent:             nil,
		useCountdownOverride: true,
		countdownOK:          false,
	}
	r := &mockRoute{leaveOK: true}

	task := newTestTask(t, cfg, p, r)
	before := time.Now()
	if err := task.runWithOptions(fastRunOptions()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	remain := task.state.TimeUntilRefresh()
	if remain < 50*time.Second {
		t.Fatalf("want refresh backoff ~60s after countdown OCR fail, remain=%v (since %v)", remain, before)
	}
}

func TestArenaTaskLeavesWhenMaxBattlesReached(t *testing.T) {
	max := 1
	cfg := &config.Arena{Enabled: true, AutoBuyCount: 0, MaxBattles: &max}
	p := &mockPage{lobby: true, tickets: 5, freeRefresh: false, opponent: &OpponentInfo{}}
	r := &mockRoute{leaveOK: true}

	task := newTestTask(t, cfg, p, r)
	task.state.Wins = 1 // already reached max battles

	if err := task.runWithOptions(fastRunOptions()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if p.runBattleCalls != 0 {
		t.Errorf("expected no battle, got %d", p.runBattleCalls)
	}
	if task.state.TotalBattles() != 1 {
		t.Errorf("expected total battles unchanged at 1, got %d", task.state.TotalBattles())
	}
	if r.leaveCalls == 0 {
		t.Error("expected route.Leave to be called")
	}
}

// 战斗结果识别失败（结算未出现/OCR 失败）属于可恢复的识别问题：
// 应 Retry 直到 MaxRetry 耗尽报错，而不是当 Fatal 立即终止，也绝不能计入战绩。
func TestArenaTaskRetriesWhenBattleResultUnknown(t *testing.T) {
	cfg := &config.Arena{Enabled: true, AutoBuyCount: 0}
	p := &mockPage{
		lobby:    true,
		tickets:  1,
		opponent: &OpponentInfo{Site: action.Point{X: 500, Y: 400}},
		battleOK: false, // 识别持续失败
	}
	r := &mockRoute{leaveOK: true}

	task := newTestTask(t, cfg, p, r)
	task.sm.Init("detect", statemachine.Options{MaxRetry: 2, MaxError: 3, Timeout: 5 * time.Second, RetryInterval: time.Millisecond})
	task.sm.Ctx = &arenaCtx{task: task, cfg: cfg}
	err := task.sm.Run(task.handlers(), fastRunOptions())
	if err == nil {
		t.Fatal("expected retry exceeded when battle result keeps unrecognizable")
	}
	if p.runBattleCalls != 3 { // 首次 + MaxRetry=2 次重试
		t.Errorf("expected RunBattle called 3 times (1+2 retries), got %d", p.runBattleCalls)
	}
	if task.state.TotalBattles() != 0 {
		t.Errorf("unrecognized battle must not count as battle, got total=%d", task.state.TotalBattles())
	}
}

// 已取色的弹窗注册为 Guard trap 并能点确认；未取色的弹窗被跳过、不注册。
func TestRegisterDialogTraps(t *testing.T) {
	g := guard.New(&mockDetector{matchMulti: true})
	exec := &mockExecutor{}
	dialogs := DialogsFeature{
		MissingTopping: DialogDef{
			Identify: screen.Feature{Colors: "dlg", Sim: 0.9},
			Confirm:  screen.Region{X1: 10, Y1: 10, X2: 20, Y2: 20},
		},
		// DeployMore 未取色：应被 Guard 跳过
	}
	registerDialogTraps(g, exec, dialogs)
	if !g.Check() {
		t.Fatal("expected configured dialog trap to fire")
	}
	if len(exec.taps) != 1 {
		t.Fatalf("expected confirm tap once, got %v", exec.taps)
	}
}

func TestRegisterDialogTrapsNilGuard(t *testing.T) {
	registerDialogTraps(nil, &mockExecutor{}, DialogsFeature{}) // must not panic
}

// pushStatus 未接入上报时无操作；接入后写入任务统计文本。
func TestTaskPushStatus(t *testing.T) {
	s := NewState(store.New(filepath.Join(t.TempDir(), "store.json")))
	s.Wins, s.Losses = 2, 1
	task := newTask(&config.Arena{}, nil, nil, s, nil)
	task.pushStatus() // reporter 为 nil，不应 panic

	r := status.New()
	task.SetStatusReporter(r)
	task.pushStatus()
	if got := r.Text(); got != "竞技场 3 场 · 胜率 66%" {
		t.Fatalf("status = %q", got)
	}
}
