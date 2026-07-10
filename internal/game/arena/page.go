package arena

import (
	"strconv"
	"strings"
	"time"

	"app/internal/config"
	"app/internal/platform/action"
	"app/internal/platform/screen"
)

type OpponentInfo struct {
	Site         action.Point
	Trophies     int
	IsBattled    bool
	BattleResult string
}

type Page struct {
	detector screen.Detector
	executor action.Executor
	feature  *Feature
}

func NewPage(det screen.Detector, exec action.Executor, f *Feature) *Page {
	return &Page{detector: det, executor: exec, feature: f}
}

func (p *Page) IsLobby() bool {
	return p.detector.MatchMultiColor("", 0.95) // use feature
}

func (p *Page) WaitLobby(timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if p.IsLobby() {
			return true
		}
		p.executor.Sleep(500)
	}
	return false
}

func (p *Page) ReadMedalAndTicket() (int, int, bool) {
	text := p.detector.OCRText(p.feature.Lobby.Reads.MedalTicket)
	parts := strings.Fields(text)
	if len(parts) < 2 {
		return 0, 0, false
	}
	medal, err1 := strconv.Atoi(parts[0])
	ticket, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil {
		return 0, 0, false
	}
	return medal, ticket, true
}

func (p *Page) ReadTrophyCount() (int, bool) {
	text := p.detector.OCRText(p.feature.Lobby.Reads.Trophy)
	n, err := strconv.Atoi(strings.TrimSpace(text))
	return n, err == nil
}

func (p *Page) FindFirstValidOpponent(cfg *config.Arena, myTrophy int) *OpponentInfo {
	// Placeholder: implement opponent scanning using feature.Lobby.Opponent
	return nil
}

func (p *Page) SwipePageLeft() {
	s := p.feature.Lobby.Gestures.SwipeLeft
	_ = p.executor.Swipe(s.From, s.To, s.DurationMs)
	p.executor.Sleep(1000)
}

func (p *Page) IsFreeRefresh() bool {
	text := p.detector.OCRText(p.feature.Lobby.Reads.FreeRefresh)
	return strings.TrimSpace(text) == "免费刷新"
}

func (p *Page) TapFreeRefresh() {
	pt := p.feature.Lobby.Actions.FreeRefresh
	_ = p.executor.Tap(action.Point{X: pt.X, Y: pt.Y})
	p.executor.Sleep(1000)
}

func (p *Page) ReadRefreshCountdown() (time.Duration, bool) {
	text := p.detector.OCRText(p.feature.Lobby.Reads.Refresh)
	// Parse text like "5分30秒" or "30秒"
	// Placeholder
	_ = text
	return 0, false
}

func (p *Page) BuyTicket() {
	btn := p.feature.Lobby.Actions.BuyTicket
	_ = p.executor.Tap(action.Point{X: btn.X, Y: btn.Y})
	p.executor.Sleep(1500)
	s := p.feature.Lobby.Actions.BuyTicketSlider
	_ = p.executor.Swipe(s.From, s.To, s.DurationMs)
	p.executor.Sleep(1000)
	confirm := p.feature.Lobby.Actions.BuyTicketConfirm
	_ = p.executor.Tap(action.Point{X: confirm.X, Y: confirm.Y})
}

func (p *Page) RunBattle() (string, bool) {
	// Placeholder: wait team select, start battle, handle dialogs, wait settlement, read result, leave
	return "胜利", true
}

func (p *Page) TapToLobby() bool {
	// Placeholder: tap leave button until lobby
	return true
}
