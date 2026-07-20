package arena

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"app/internal/config"
	"app/internal/logger"
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
	id := p.feature.Lobby.Identify
	if id.Colors == "" {
		return false
	}
	return p.detector.MatchMultiColor(id.Colors, id.Sim)
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
	text, err := p.detector.OCRText(p.feature.Lobby.Reads.MedalTicket)
	if err != nil {
		logger.Warnf("[Arena] medal/ticket OCR failed: %v", err)
		return 0, 0, false
	}
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
	text, err := p.detector.OCRText(p.feature.Lobby.Reads.Trophy)
	if err != nil {
		logger.Warnf("[Arena] trophy OCR failed: %v", err)
		return 0, false
	}
	n, err := strconv.Atoi(strings.TrimSpace(text))
	return n, err == nil
}

func (p *Page) FindFirstValidOpponent(cfg *config.Arena, myTrophy int) *OpponentInfo {
	op := p.feature.Lobby.Opponent
	anchors := p.detector.FindMultiColorsAll(op.Anchor.Region, op.Anchor.Colors, op.Anchor.Sim, op.Anchor.Dir)
	lo, hi := myTrophy-cfg.TrophyDiff, myTrophy+cfg.TrophyDiff

	for _, a := range anchors {
		text, err := p.detector.OCRText(offsetRegion(op.TrophyRect, a))
		if err != nil {
			logger.Warnf("[Arena] opponent trophy OCR failed at %+v: %v", a, err)
			continue
		}
		trophy, ok := readInt(text)
		if !ok {
			continue // OCR 失败：跳过该锚点，不漏后面的卡
		}
		if battledAt(p.detector, offsetPoint(op.ResultOffset, a), op) {
			continue // 已战：跳过
		}
		if trophy < lo || trophy > hi {
			continue // 奖杯不在区间：跳过
		}
		return &OpponentInfo{
			Site:      action.Point{X: a.X + op.ClickOffset.X, Y: a.Y + op.ClickOffset.Y},
			Trophies:  trophy,
			IsBattled: false,
		}
	}
	return nil
}

func (p *Page) SwipePageLeft() {
	s := p.feature.Lobby.Gestures.SwipeLeft
	p.executor.Swipe(s.From, s.To, s.DurationMs)
	p.executor.Sleep(1000)
}

func (p *Page) IsFreeRefresh() bool {
	text, err := p.detector.OCRText(p.feature.Lobby.Reads.FreeRefresh)
	if err != nil {
		logger.Warnf("[Arena] free-refresh OCR failed: %v", err)
		return false
	}
	return strings.TrimSpace(text) == "免费刷新"
}

func (p *Page) TapFreeRefresh() {
	p.executor.Tap(action.RandomIn(p.feature.Lobby.Actions.FreeRefresh))
	p.executor.Sleep(1000)
}

func (p *Page) ReadRefreshCountdown() (time.Duration, bool) {
	text, err := p.detector.OCRText(p.feature.Lobby.Reads.Refresh)
	if err != nil {
		logger.Warnf("[Arena] refresh countdown OCR failed: %v", err)
		return 0, false
	}
	return parseCountdown(strings.TrimSpace(text))
}

func (p *Page) BuyTicket() {
	p.executor.Tap(action.RandomIn(p.feature.Lobby.Actions.BuyTicket))
	p.executor.Sleep(1500)
	s := p.feature.Lobby.Actions.BuyTicketSlider
	p.executor.Swipe(s.From, s.To, s.DurationMs)
	p.executor.Sleep(1000)
	p.executor.Tap(action.RandomIn(p.feature.Lobby.Actions.BuyTicketConfirm))
}

// TapEntry OCR-taps the arena entrance on the adventure page.
func (p *Page) TapEntry() bool {
	kw := strings.TrimSpace(p.feature.Entry.Keyword)
	if kw == "" {
		kw = "王国竞技场"
	}
	pt, ok := p.detector.FindOCRText(p.feature.Entry.Region, kw)
	if !ok {
		logger.Warnf("[Arena] entry OCR miss keyword=%q region=%+v", kw, p.feature.Entry.Region)
		return false
	}
	p.executor.Tap(action.Point{X: pt.X, Y: pt.Y})
	p.executor.Sleep(1500)
	return true
}

// TapLobbyClose taps the lobby close control once (does not require ending in lobby).
func (p *Page) TapLobbyClose() {
	p.executor.Tap(action.RandomIn(p.feature.Lobby.Actions.Close))
	p.executor.Sleep(800)
}

// TapToLobby taps lobby Close until IsLobby or retries/timeout exhausted.
func (p *Page) TapToLobby() bool {
	if p.IsLobby() {
		return true
	}
	deadline := time.Now().Add(10 * time.Second)
	for i := 0; i < 3 && time.Now().Before(deadline); i++ {
		p.TapLobbyClose()
		if p.IsLobby() {
			return true
		}
	}
	return p.IsLobby()
}

func (p *Page) TapOpponentSite(site action.Point) {
	p.executor.Tap(site)
	p.executor.Sleep(1000)
}

func (p *Page) IsTeamSelect() bool {
	id := p.feature.TeamSelect.Identify
	if id.Colors == "" {
		return false
	}
	return p.detector.MatchMultiColor(id.Colors, id.Sim)
}

// HasTeamSelectPage reports whether TeamSelect.Identify is configured.
func (p *Page) HasTeamSelectPage() bool {
	return p.feature.TeamSelect.Identify.Colors != ""
}

func (p *Page) WaitTeamSelect(timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if p.IsTeamSelect() {
			return true
		}
		p.executor.Sleep(500)
	}
	return false
}

func (p *Page) TapStartBattle() {
	p.executor.Tap(action.RandomIn(p.feature.TeamSelect.Actions.StartBattle))
	p.executor.Sleep(1000)
}

const (
	battleTimeout       = 3 * time.Minute
	leaveToLobbyTimeout = 30 * time.Second
)

func (p *Page) RunBattle() (string, bool) {
	id := p.feature.Settlement.Identify
	if id.Colors == "" {
		logger.Warnf("[Arena] Settlement.Identify not configured")
		return "", false
	}
	deadline := time.Now().Add(battleTimeout)
	for time.Now().Before(deadline) {
		if p.detector.MatchMultiColor(id.Colors, id.Sim) {
			break
		}
		p.executor.Sleep(500)
	}
	if !p.detector.MatchMultiColor(id.Colors, id.Sim) {
		logger.Warnf("[Arena] settlement not reached before timeout")
		return "", false
	}
	text, err := p.detector.OCRText(p.feature.Settlement.Reads.Result)
	if err != nil {
		logger.Warnf("[Arena] battle result OCR failed: %v", err)
		return "", false
	}
	result, ok := parseBattleResult(text)
	if !ok {
		logger.Warnf("[Arena] battle result OCR unparsable text=%q", text)
		return "", false
	}
	if !p.leaveSettlementToLobby() {
		return "", false
	}
	return result, true
}

func (p *Page) leaveSettlementToLobby() bool {
	leaveID := p.feature.Settlement.Actions.LeaveIdentify
	leaveR := p.feature.Settlement.Actions.Leave
	deadline := time.Now().Add(leaveToLobbyTimeout)
	for time.Now().Before(deadline) {
		if p.IsLobby() {
			return true
		}
		if leaveID.Colors != "" && !p.detector.MatchMultiColor(leaveID.Colors, leaveID.Sim) {
			p.executor.Sleep(400)
			continue
		}
		p.executor.Tap(action.RandomIn(leaveR))
		p.executor.Sleep(800)
		if p.IsLobby() {
			return true
		}
		if p.TapToLobby() {
			return true
		}
	}
	return p.IsLobby()
}

func parseBattleResult(s string) (string, bool) {
	s = strings.TrimSpace(s)
	switch {
	case strings.Contains(s, "胜利"):
		return "胜利", true
	case strings.Contains(s, "平局"):
		return "平局", true
	case strings.Contains(s, "失败"):
		return "失败", true
	default:
		return "", false
	}
}

func offsetRegion(rel screen.Region, a screen.Point) screen.Region {
	return screen.Region{X1: a.X + rel.X1, Y1: a.Y + rel.Y1, X2: a.X + rel.X2, Y2: a.Y + rel.Y2}
}

func offsetPoint(rel screen.Point, a screen.Point) screen.Point {
	return screen.Point{X: a.X + rel.X, Y: a.Y + rel.Y}
}

func readInt(s string) (int, bool) {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	return n, err == nil
}

var (
	reColon  = regexp.MustCompile(`^\s*(\d{1,3}):(\d{1,2})\s*$`)
	reDigits = regexp.MustCompile(`^\s*(\d+)\s*$`)
	reMin    = regexp.MustCompile(`(\d+)\s*分`)
	reSec    = regexp.MustCompile(`(\d+)\s*秒`)
)

// parseCountdown 解析刷新倒计时。支持 "5分30秒"/"30秒"/"5分"/"05:30"/"330"。
// 抓不到数字或合计为 0 → (0, false)。
func parseCountdown(s string) (time.Duration, bool) {
	if m := reColon.FindStringSubmatch(s); m != nil {
		min, _ := strconv.Atoi(m[1])
		sec, _ := strconv.Atoi(m[2])
		d := time.Duration(min)*time.Minute + time.Duration(sec)*time.Second
		if d == 0 {
			return 0, false
		}
		return d, true
	}
	if m := reDigits.FindStringSubmatch(s); m != nil {
		sec, _ := strconv.Atoi(m[1])
		if sec == 0 {
			return 0, false
		}
		return time.Duration(sec) * time.Second, true
	}
	total := 0
	if m := reMin.FindStringSubmatch(s); m != nil {
		v, _ := strconv.Atoi(m[1])
		total += v * 60
	}
	if m := reSec.FindStringSubmatch(s); m != nil {
		v, _ := strconv.Atoi(m[1])
		total += v
	}
	if total == 0 {
		return 0, false
	}
	return time.Duration(total) * time.Second, true
}

// battledAt 判断结果点 pt 是否显示已战标记色。
// 命中 Win/Draw/Lose 任一 → 已战(true)。三色都不命中 → 未战(false)。
// 注：AutoGo SDK 的比色接口只返 bool、无 error 通道，无法区分"未战中性态"与
// "比色异常"，因此"异常当已战"的保守分支无法实现；若 SDK 未来提供 error 通道再补。
func battledAt(det screen.Detector, pt screen.Point, op OpponentFeature) bool {
	if det.MatchColor(pt.X, pt.Y, op.ResultColors.Win, op.ResultSim) {
		return true
	}
	if det.MatchColor(pt.X, pt.Y, op.ResultColors.Draw, op.ResultSim) {
		return true
	}
	if det.MatchColor(pt.X, pt.Y, op.ResultColors.Lose, op.ResultSim) {
		return true
	}
	return false
}
