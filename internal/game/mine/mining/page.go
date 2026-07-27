package mining

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

// 选卡扫描常量（对应 Lua 开采_页面.lua 顶部常量）。
const (
	dedupRadius        = 80
	selectedNearRadius = 120
	maxSwipes          = 20
	tapYOffset         = -200
	tapSettleMs        = 450
)

// swipeCardListLeft 卡列表向左滑（Lua SWIPE_CARD_LIST）。
var swipeCardListLeft = action.Swipe{
	From:       action.Point{X: 1480, Y: 738},
	To:         action.Point{X: 150, Y: 738},
	DurationMs: 600,
}

// noMineCardHints 矿脉卡清单「没有矿卡」提示词（Lua NO_MINE_CARD_HINTS）。
var noMineCardHints = []string{"没有可选择的矿脉卡", "没有"}

// Page 矿山开采页面封装（对应 Lua 开采_页面.lua）。
type Page struct {
	detector screen.Detector
	executor action.Executor
	feature  *Feature
	home     *mine.Page // IsSettlementRoute 需要判矿山首页

	lastNoMineCard bool
}

func NewPage(det screen.Detector, exec action.Executor, f *Feature, home *mine.Page) *Page {
	if f == nil {
		f = DefaultFeature()
	}
	return &Page{detector: det, executor: exec, feature: f, home: home}
}

func (p *Page) match(f screen.Feature) bool {
	if f.Colors == "" {
		return false
	}
	return p.detector.MatchMultiColor(f.Colors, f.Sim)
}

// findFirst 多点找色取首个命中；Colors 为空视为未配置。
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

func (p *Page) found(cf mine.ColorFind) bool {
	_, ok := p.findFirst(cf)
	return ok
}

func (p *Page) waitMatch(f screen.Feature, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if p.match(f) {
			return true
		}
		p.executor.Sleep(500)
	}
	return false
}

func (p *Page) IsMiningPage() bool { return p.match(p.feature.Page) }

func (p *Page) WaitMiningPage(timeout time.Duration) bool {
	return p.waitMatch(p.feature.Page, timeout)
}

func (p *Page) IsSetup() bool      { return p.match(p.feature.SetupIdentify) }
func (p *Page) IsSetupReady() bool { return p.match(p.feature.SetupReadyIdentify) }

func (p *Page) WaitSetupReady(timeout time.Duration) bool {
	return p.waitMatch(p.feature.SetupReadyIdentify, timeout)
}

// IsRewardPage 获得开采奖励页（Lua：OCR 标题优先，比色兜底；本特征库无比色兜底）。
func (p *Page) IsRewardPage() bool {
	rw := p.feature.RewardPage
	if rw.TitleText == "" {
		return false
	}
	_, ok := p.detector.FindOCRText(rw.TitleOCR, rw.TitleText)
	return ok
}

// IsSettlementRoute 结算链路页：既不在矿山首页也不在开采页（Lua isSettlementRoute）。
func (p *Page) IsSettlementRoute() bool {
	homeCurrent := p.home != nil && p.home.IsCurrent()
	return !homeCurrent && !p.IsMiningPage()
}

// TapUntilMatchMiningPage 反复点奖励确认直到回到开采页（Lua tapUntilMatchMiningPage）。
func (p *Page) TapUntilMatchMiningPage() bool {
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		if p.IsMiningPage() {
			return true
		}
		p.executor.Tap(action.RandomIn(p.feature.RewardPage.ConfirmBtn))
		p.executor.Sleep(500)
	}
	return p.IsMiningPage()
}

func (p *Page) HasCompletedTask() bool { return p.found(p.feature.CompletedTask) }

func (p *Page) TapCompletedSlot() bool {
	pt, ok := p.findFirst(p.feature.CompletedTask)
	if !ok {
		return false
	}
	p.executor.Tap(pt)
	p.executor.Sleep(500)
	return true
}

func (p *Page) HasFreeSlot() bool {
	return p.found(p.feature.FreeLocation) || p.found(p.feature.FreePlus)
}

// findSelectedCardPoints 已选矿卡角标（未取色时为空，Lua selectedMark=nil）。
func (p *Page) findSelectedCardPoints() []screen.Point {
	mark := p.feature.CardSelect.SelectedMark
	if mark.Colors == "" {
		return nil
	}
	return p.detector.FindMultiColorsAll(mark.Region, mark.Colors, mark.Sim, mark.Dir)
}

// HasNoMineCardInList 矿脉卡清单提示没有矿卡（Lua hasNoMineCardInList）。
func (p *Page) HasNoMineCardInList() bool {
	text, err := p.detector.OCRText(p.feature.NoMineCardOCR)
	if err != nil {
		logger.Warnf("[矿山开采.页面] noMineCard OCR failed: %v", err)
		return false
	}
	return hasNoMineCardHint(text)
}

func hasNoMineCardHint(text string) bool {
	if strings.TrimSpace(text) == "" {
		return false
	}
	for _, hint := range noMineCardHints {
		if strings.Contains(text, hint) {
			return true
		}
	}
	return false
}

// EnterMultiSelect 进入「选择多个」模式；清单无矿卡时退回开采页并记 lastNoMineCard。
func (p *Page) EnterMultiSelect() bool {
	p.lastNoMineCard = false
	ocrRegion := p.feature.MultiSelectOCR
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		// 优先检查无矿卡提示（无矿卡时不会出现"选择多个"）
		if p.HasNoMineCardInList() {
			logger.Infof("[矿山开采.页面] 矿脉卡清单提示无矿卡，退出选卡页面")
			p.executor.Tap(action.RandomIn(p.feature.CardSelect.BackBtn))
			p.executor.Sleep(1000)
			if p.WaitMiningPage(30 * time.Second) {
				p.lastNoMineCard = true
			} else {
				logger.Warnf("[矿山开采.页面] 退出选卡页面后未回到矿山开采首页")
			}
			return false
		}
		if _, ok := p.detector.FindOCRText(ocrRegion, "选择多个"); ok {
			p.executor.Tap(action.RandomIn(p.feature.MultiSelectBtn))
			p.executor.Sleep(1000)
			return true
		}
		p.executor.Sleep(500)
	}
	logger.Warnf("[矿山开采.页面] enterMultiSelect 等待选卡页面超时")
	return false
}

func (p *Page) WasNoMineCard() bool { return p.lastNoMineCard }

// TapFreeSlot 点空闲栏位并进入多选模式（Lua tapFreeSlot）。
func (p *Page) TapFreeSlot() bool {
	pt, ok := p.findFirst(p.feature.FreeLocation)
	if !ok {
		pt, ok = p.findFirst(p.feature.FreePlus)
	}
	if !ok {
		return false
	}
	p.executor.Tap(pt)
	p.executor.Sleep(500)
	return p.EnterMultiSelect()
}

// ReadChooseQuota OCR 读取已选/可选数量（Lua Ocr.fraction）。
func (p *Page) ReadChooseQuota() (int, int, bool) {
	text, err := p.detector.OCRText(p.feature.CanChooseNum)
	if err != nil {
		logger.Warnf("[矿山开采.页面] choose quota OCR failed: %v", err)
		return 0, 0, false
	}
	return parseFraction(text)
}

// SelectTargetCards 在卡列表中按方向扫描并点选目标矿卡，直到选满 need 张或列表穷尽。
// 返回 (本次新选数量, 列表是否穷尽)。对应 Lua selectTargetCards。
func (p *Page) SelectTargetCards(target mine.ColorFind, need int, direction string) (int, bool) {
	if need <= 0 {
		return 0, false
	}
	if target.Colors == "" {
		logger.Warnf("[矿山开采.页面] 目标矿卡特征未配置")
		return 0, true
	}

	startCur, startMax, ok := p.ReadChooseQuota()
	if !ok {
		startCur, startMax = 0, need
	}
	targetCur := startCur + need
	if direction == "" {
		direction = "left"
	}

	swipes := 0
	exhausted := false
	for swipes <= maxSwipes {
		cur, max, ok := p.ReadChooseQuota()
		if !ok {
			cur, max = startCur, startMax
		}
		if cur >= max || cur >= targetCur {
			return cur - startCur, false
		}

		selectedMarks := p.findSelectedCardPoints()
		tappedThisPass := []screen.Point{}
		progressed := false
		points := sortByX(p.detector.FindMultiColorsAll(target.Region, target.Colors, target.Sim, target.Dir))
		logger.Infof("[矿山开采.页面] 扫描目标卡 方向:%s 可见:%d 已选:%d 还需:%d 滑动:%d",
			direction, len(points), cur-startCur, targetCur-cur, swipes)

		for _, pt := range points {
			c, _, ok := p.ReadChooseQuota()
			if !ok || c >= targetCur {
				break
			}
			if isNearExisting(pt, tappedThisPass, dedupRadius) {
				continue // 同屏重复命中同一张卡
			}
			if isNearExisting(pt, selectedMarks, selectedNearRadius) {
				continue // 跳过已选标记卡
			}
			tappedThisPass = append(tappedThisPass, pt)
			if p.tapCardIfQuotaIncreases(pt, targetCur) {
				progressed = true
			}
		}

		if c, _, ok := p.ReadChooseQuota(); ok && c >= targetCur {
			return c - startCur, false
		}

		if !progressed {
			if !p.swipeCardList(direction) {
				exhausted = true
				break
			}
			swipes++
		}
	}

	if swipes > maxSwipes {
		exhausted = true
	}
	finalCur, _, ok := p.ReadChooseQuota()
	if !ok {
		finalCur = startCur
	}
	logger.Warnf("[矿山开采.页面] 选卡不足 %d/%d（滑动%d次）", finalCur-startCur, need, swipes)
	return finalCur - startCur, exhausted
}

// tapCardPoint 点矿卡（命中点上移 tapYOffset，Lua tapCardPoint）。
func (p *Page) tapCardPoint(pt screen.Point) {
	p.executor.Tap(action.Point{X: pt.X, Y: pt.Y + tapYOffset})
	p.executor.Sleep(tapSettleMs)
}

// tapCardIfQuotaIncreases 点卡并以配额变化校验是否选中；误触已选卡（配额下降）时点回恢复。
func (p *Page) tapCardIfQuotaIncreases(pt screen.Point, targetCur int) bool {
	before, _, ok := p.ReadChooseQuota()
	if !ok || before >= targetCur {
		return false
	}
	p.tapCardPoint(pt)
	after, _, ok := p.ReadChooseQuota()
	if !ok {
		return false
	}
	if after > before {
		return true
	}
	if after < before {
		logger.Warnf("[矿山开采.页面] 误触已选卡 (%d→%d)，恢复", before, after)
		p.tapCardPoint(pt)
	}
	return false
}

// swipeCardList 滑卡列表并用边缘 OCR 判断是否到尽头（Lua swipeCardList）。
func (p *Page) swipeCardList(direction string) bool {
	s := swipeCardListLeft
	edge := p.feature.CardListEndOCR
	if direction == "right" {
		s = action.Swipe{From: s.To, To: s.From, DurationMs: s.DurationMs}
		edge = p.feature.CardListStartOCR
	}
	p.executor.Swipe(s.From, s.To, s.DurationMs)
	p.executor.Sleep(300)
	if !p.ocrRegionHasText(edge) {
		return false
	}
	p.executor.Sleep(500)
	return true
}

func (p *Page) ocrRegionHasText(r screen.Region) bool {
	text, err := p.detector.OCRText(r)
	if err != nil {
		return false
	}
	return strings.TrimSpace(text) != ""
}

func (p *Page) ConfirmCardSelection() bool {
	p.executor.Tap(action.RandomIn(p.feature.CardSelect.ConfirmBtn))
	p.executor.Sleep(800)
	return true
}

func (p *Page) HasStartableCard() bool { return p.found(p.feature.StartableCard) }

// TapReadySlot 点可开采矿卡（命中点偏移 -100,+100）并等准备开采页（Lua tapReadySlot）。
func (p *Page) TapReadySlot() bool {
	pt, ok := p.findFirst(p.feature.StartableCard)
	if !ok {
		return false
	}
	p.executor.Tap(action.Point{X: pt.X - 100, Y: pt.Y + 100})
	p.executor.Sleep(500)
	return p.waitMatch(p.feature.SetupIdentify, 30*time.Second)
}

// AutoSelectCookieAndStart 自动选饼干并开始开采，处理顺序未知的两个饼干弹窗。
// 对应 Lua autoSelectCookieAndStart + Dialog.resolveUntilIdle。
func (p *Page) AutoSelectCookieAndStart() bool {
	p.executor.Tap(action.RandomIn(p.feature.AutoSelectCookieBtn))
	p.executor.Sleep(500)
	if !p.WaitSetupReady(30 * time.Second) {
		return false
	}
	p.executor.Tap(action.RandomIn(p.feature.ConfirmStartBtn))
	p.executor.Sleep(500)
	return p.resolveCookieDialogs(8 * time.Second)
}

// resolveCookieDialogs 循环处理确认饼干/数量警告两个弹窗：命中即点「今日不再询问」
// 再点确认；连续 800ms 无弹窗视为处理完毕。整体超时时：处理过弹窗才算成功。
func (p *Page) resolveCookieDialogs(timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	lastActivity := time.Now()
	handled := 0
	for time.Now().Before(deadline) {
		d, hit := p.hitCookieDialog()
		if hit {
			p.executor.Tap(action.RandomIn(d.TodayNotAskAgain))
			p.executor.Sleep(300)
			p.executor.Tap(action.RandomIn(d.Confirm))
			p.executor.Sleep(500)
			handled++
			lastActivity = time.Now()
			continue
		}
		if handled > 0 && time.Since(lastActivity) > 800*time.Millisecond {
			return true
		}
		if handled == 0 && time.Since(lastActivity) > 3*time.Second {
			return true // 始终没有弹窗：视为无需处理
		}
		p.executor.Sleep(300)
	}
	if handled == 0 {
		logger.Warnf("[矿山开采.页面] 饼干弹窗处理超时（未出现弹窗）")
	}
	return handled > 0
}

func (p *Page) hitCookieDialog() (DialogDef, bool) {
	d := p.feature.Dialogs.ConfirmCookie
	if p.match(d.Identify) {
		return d, true
	}
	d = p.feature.Dialogs.CookieCountWarning
	if p.match(d.Identify) {
		return d, true
	}
	return DialogDef{}, false
}

// TapBackBtn 开采页/选卡页 → 上一层。
func (p *Page) TapBackBtn() {
	p.executor.Tap(action.RandomIn(p.feature.BackBtn))
	p.executor.Sleep(1000)
}

func sortByX(points []screen.Point) []screen.Point {
	sort.Slice(points, func(i, j int) bool { return points[i].X < points[j].X })
	return points
}

func isNearExisting(pt screen.Point, list []screen.Point, radius int) bool {
	r2 := radius * radius
	for _, p := range list {
		dx := pt.X - p.X
		dy := pt.Y - p.Y
		if dx*dx+dy*dy <= r2 {
			return true
		}
	}
	return false
}

var reFraction = regexp.MustCompile(`(\d+)\s*/\s*(\d+)`)

// parseFraction 解析 "2/3"、"1/12,611" 形式的 已选/可选 数量。
func parseFraction(s string) (int, int, bool) {
	s = strings.ReplaceAll(s, ",", "")
	s = strings.ReplaceAll(s, "，", "")
	m := reFraction.FindStringSubmatch(s)
	if m == nil {
		return 0, 0, false
	}
	cur, err1 := strconv.Atoi(m[1])
	max, err2 := strconv.Atoi(m[2])
	if err1 != nil || err2 != nil {
		return 0, 0, false
	}
	return cur, max, true
}
