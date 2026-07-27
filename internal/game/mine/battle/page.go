package battle

import (
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"app/internal/game/mine"
	"app/internal/logger"
	"app/internal/platform/action"
	"app/internal/platform/screen"
)

// Page 矿山战斗页面封装（对应 Lua 战斗_页面.lua）。
type Page struct {
	detector screen.Detector
	executor action.Executor
	feature  *Feature
}

func NewPage(det screen.Detector, exec action.Executor, f *Feature) *Page {
	if f == nil {
		f = DefaultFeature()
	}
	return &Page{detector: det, executor: exec, feature: f}
}

func (p *Page) match(f screen.Feature) bool {
	if f.Colors == "" {
		return false
	}
	return p.detector.MatchMultiColor(f.Colors, f.Sim)
}

func (p *Page) findFirst(cf mine.ColorFind) (screen.Point, bool) {
	if cf.Colors == "" {
		return screen.Point{}, false
	}
	pts := p.detector.FindMultiColorsAll(cf.Region, cf.Colors, cf.Sim, cf.Dir)
	if len(pts) == 0 {
		return screen.Point{}, false
	}
	return pts[0], true
}

func (p *Page) IsBattlePage() bool { return p.match(p.feature.Identify) }

func (p *Page) WaitBattlePage(timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if p.IsBattlePage() {
			return true
		}
		p.executor.Sleep(500)
	}
	return false
}

func (p *Page) TapBackBtn() {
	p.executor.Tap(action.RandomIn(p.feature.BackBtn))
	p.executor.Sleep(1000)
}

// FindQuickBattleButton 查找快转按钮。
func (p *Page) FindQuickBattleButton() (screen.Point, bool) {
	return p.findFirst(p.feature.QuickBattleBtn)
}

func (p *Page) TapQuickBattleButton(pt screen.Point) {
	p.executor.Tap(pt)
	p.executor.Sleep(1000)
}

// IsQuickBattleDialog 快转弹窗是否出现。
func (p *Page) IsQuickBattleDialog() bool {
	return p.match(p.feature.QuickBattleDialog.Identify)
}

func (p *Page) WaitQuickBattleDialog(timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if p.IsQuickBattleDialog() {
			return true
		}
		p.executor.Sleep(500)
	}
	return false
}

// WaitQuickBattleDialogGone 等弹窗消失；特征未配置时与 Lua 一致直接返回 true。
func (p *Page) WaitQuickBattleDialogGone(timeout time.Duration) bool {
	if p.feature.QuickBattleDialog.Identify.Colors == "" {
		return true
	}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if !p.IsQuickBattleDialog() {
			return true
		}
		p.executor.Sleep(500)
	}
	return false
}

// ReadClockCount 读取快转发条数量（使用/持有，如 "1/12,611"）。
// 解析失败返回 ok=false（Lua：避免 fraction 兜底误识别，手动解析）。
func (p *Page) ReadClockCount() (int, int, bool) {
	text, err := p.detector.OCRText(p.feature.QuickBattleDialog.ClockCountOCR)
	if err != nil {
		logger.Warnf("[矿山战斗.页面] clock count OCR failed: %v", err)
		return 0, 0, false
	}
	return parseClockCount(text)
}

func (p *Page) TapQuickBattleConfirm() {
	p.executor.Tap(action.RandomIn(p.feature.QuickBattleDialog.ConfirmBtn))
	p.executor.Sleep(1000)
}

func (p *Page) TapQuickBattleCancel() {
	p.executor.Tap(action.RandomIn(p.feature.QuickBattleDialog.CancelBtn))
	p.executor.Sleep(1000)
}

// TapSettleUntilBattlePage 点结算按钮直到战斗页再次出现（Lua tapSettleUntilBattlePage）。
func (p *Page) TapSettleUntilBattlePage() bool {
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		if p.IsBattlePage() {
			return true
		}
		p.executor.Tap(action.RandomIn(p.feature.SettleBtn))
		p.executor.Sleep(800)
	}
	return p.IsBattlePage()
}

// FindBattleCards 查找本页所有战斗卡。
func (p *Page) FindBattleCards() []screen.Point {
	cf := p.feature.BattleCard
	if cf.Colors == "" {
		return nil
	}
	return p.detector.FindMultiColorsAll(cf.Region, cf.Colors, cf.Sim, cf.Dir)
}

func (p *Page) TapBattleCard(pt screen.Point) {
	p.executor.Tap(pt)
	p.executor.Sleep(1000)
}

// RecognizeSoulStoneType 识别灵魂石类型，返回匹配到的目标名称。
// 同一区域命中多个目标灵魂石时无法区分，返回 ""（Lua 同语义返回 nil）。
func (p *Page) RecognizeSoulStoneType(targets map[string]bool) string {
	if len(targets) == 0 {
		return ""
	}
	var matches []string
	for _, cat := range p.feature.SoulStones {
		names := make([]string, 0, len(cat.Stones))
		for name := range cat.Stones {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			def := cat.Stones[name]
			if !targets[name] || def.Colors == "" {
				continue
			}
			if pts := p.detector.FindMultiColorsAll(def.Region, def.Colors, def.Sim, def.Dir); len(pts) > 0 {
				matches = append(matches, cat.Name+"/"+name)
			}
		}
	}
	if len(matches) == 1 {
		name := strings.SplitN(matches[0], "/", 2)[1]
		logger.Debugf("[矿山战斗.页面] 灵魂石匹配 %s", matches[0])
		return name
	}
	if len(matches) > 1 {
		logger.Warnf("[矿山战斗.页面] 灵魂石多个候选命中，无法区分: %s", strings.Join(matches, " , "))
	}
	return ""
}

// SwipeUpAndCheckLastPage 向上翻页并识别是否已到末页。
// 注意：Lua 在滑动按住未松手时识别末页（swipeEx beforeUp）；Executor 接口只有
// 整体 Swipe，这里退化为松手后再识别（末页特征在松手后仍停留，语义等价但时序不同）。
func (p *Page) SwipeUpAndCheckLastPage() bool {
	s := p.feature.PageSwipe
	p.executor.Swipe(s.From, s.To, s.DurationMs)
	p.executor.Sleep(500)
	_, isLast := p.findFirst(p.feature.LastPage)
	logger.Infof("[矿山战斗.页面] 翻页后识别末页=%v", isLast)
	return isLast
}

var reClock = regexp.MustCompile(`(\d+)\s*/\s*(\d+)`)

// parseClockCount 解析 "1/12,611" 形式的 使用/持有 数量。
func parseClockCount(s string) (int, int, bool) {
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, ",", "")
	s = strings.ReplaceAll(s, "，", "")
	m := reClock.FindStringSubmatch(s)
	if m == nil {
		return 0, 0, false
	}
	used, err1 := strconv.Atoi(m[1])
	owned, err2 := strconv.Atoi(m[2])
	if err1 != nil || err2 != nil {
		return 0, 0, false
	}
	return used, owned, true
}
