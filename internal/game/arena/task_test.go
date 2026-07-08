package arena

import (
	"path/filepath"
	"testing"
	"time"

	"app/internal/config"
	"app/internal/statemachine"
	"app/internal/store"
)

type mockPage struct {
	lobby        bool
	tickets      int
	freeRefresh  bool
	opponent     *OpponentInfo
	battleResult string
	battleOK     bool

	buyTicketCalls      int
	runBattleCalls      int
	swipeCalls          int
	tapFreeRefreshCalls int
}

func (m *mockPage) IsLobby() bool                        { return m.lobby }
func (m *mockPage) WaitLobby(timeout time.Duration) bool { return m.lobby }
func (m *mockPage) ReadMedalAndTicket() (int, int, bool) { return 0, m.tickets, true }
func (m *mockPage) ReadTrophyCount() (int, bool)         { return 1000, true }
func (m *mockPage) FindFirstValidOpponent(cfg *config.Arena, myTrophy int) *OpponentInfo {
	return m.opponent
}
func (m *mockPage) SwipePageLeft()                              { m.swipeCalls++ }
func (m *mockPage) IsFreeRefresh() bool                         { return m.freeRefresh }
func (m *mockPage) TapFreeRefresh()                             { m.tapFreeRefreshCalls++; m.freeRefresh = false }
func (m *mockPage) ReadRefreshCountdown() (time.Duration, bool) { return 0, false }
func (m *mockPage) BuyTicket()                                  { m.buyTicketCalls++; m.tickets++ }
func (m *mockPage) RunBattle() (string, bool) {
	m.runBattleCalls++
	if m.tickets > 0 {
		m.tickets--
	}
	return m.battleResult, m.battleOK
}
func (m *mockPage) TapToLobby() bool { return true }

type mockRoute struct {
	enterCalls int
	leaveCalls int
	enterOK    bool
	leaveOK    bool
}

func (m *mockRoute) Enter() bool { m.enterCalls++; return m.enterOK }
func (m *mockRoute) Leave() bool { m.leaveCalls++; return m.leaveOK }

func newTestTask(t *testing.T, cfg *config.Arena, p page, r route) *Task {
	s := NewSession(store.New(filepath.Join(t.TempDir(), "store.json")))
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

	if task.session.Tickets != 0 {
		t.Errorf("expected tickets to stay 0, got %d", task.session.Tickets)
	}
	if task.session.BuyCount != 0 {
		t.Errorf("expected no ticket purchase, got buyCount=%d", task.session.BuyCount)
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
	if task.session.BuyCount != 1 {
		t.Errorf("expected buyCount=1, got %d", task.session.BuyCount)
	}
	if task.session.Tickets != 1 {
		t.Errorf("expected tickets=1 after purchase, got %d", task.session.Tickets)
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
		freeRefresh:  false,
		opponent:     &OpponentInfo{},
		battleResult: "胜利",
		battleOK:     true,
	}
	r := &mockRoute{leaveOK: true}

	task := newTestTask(t, cfg, p, r)
	if err := task.runWithOptions(fastRunOptions()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if p.runBattleCalls != 1 {
		t.Errorf("expected RunBattle called once, got %d", p.runBattleCalls)
	}
	if task.session.TotalBattles() != 1 {
		t.Errorf("expected total battles=1, got %d", task.session.TotalBattles())
	}
	if task.session.Wins != 1 {
		t.Errorf("expected wins=1, got %d", task.session.Wins)
	}
	if task.session.Tickets != 0 {
		t.Errorf("expected tickets decremented to 0, got %d", task.session.Tickets)
	}
	if r.leaveCalls == 0 {
		t.Error("expected route.Leave to be called")
	}
}

func TestArenaTaskLeavesWhenMaxBattlesReached(t *testing.T) {
	max := 1
	cfg := &config.Arena{Enabled: true, AutoBuyCount: 0, MaxBattles: &max}
	p := &mockPage{lobby: true, tickets: 5, freeRefresh: false, opponent: &OpponentInfo{}}
	r := &mockRoute{leaveOK: true}

	task := newTestTask(t, cfg, p, r)
	task.session.Wins = 1 // already reached max battles

	if err := task.runWithOptions(fastRunOptions()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if p.runBattleCalls != 0 {
		t.Errorf("expected no battle, got %d", p.runBattleCalls)
	}
	if task.session.TotalBattles() != 1 {
		t.Errorf("expected total battles unchanged at 1, got %d", task.session.TotalBattles())
	}
	if r.leaveCalls == 0 {
		t.Error("expected route.Leave to be called")
	}
}
