package market

import (
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"app/internal/logger"
	"app/internal/platform/action"
	"app/internal/platform/screen"
)

// 购买结果（对齐 Lua tapShelfAndResolve 的 outcome）。
const (
	ResultPurchased = "purchased"
	ResultShortage  = "shortage"
	ResultFailed    = "failed"
)

// PurchaseStats 一轮扫货统计（Lua purchaseWishlist 的 stats）。
type PurchaseStats struct {
	Purchased int
	SoldOut   int
	Shortage  int
	Failed    int
}

// dedupRadius 同屏/同页锚点去重半径（Lua DEDUP_RADIUS）。
const dedupRadius = 80

// 购买弹窗等待时长；包级变量以便测试缩短（生产保持 Lua 的 5000ms）。
var (
	confirmDialogWait  = 5 * time.Second
	purchaseResultWait = 5 * time.Second
)

type stockTarget struct {
	Point screen.Point
	Key   string
}

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

func (p *Page) IsCurrent() bool {
	id := p.feature.Page.Identify
	if id.Colors == "" {
		return false
	}
	return p.detector.MatchMultiColor(id.Colors, id.Sim)
}

func (p *Page) WaitCurrent(timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if p.IsCurrent() {
			return true
		}
		p.executor.Sleep(500)
	}
	return false
}

// TapEntryBtn 点王国活动页上的交易所入口。
func (p *Page) TapEntryBtn() {
	p.executor.Tap(action.RandomIn(p.feature.Entry.Btn))
	p.executor.Sleep(1200)
}

// TapClose 点交易所关闭按钮一次。
func (p *Page) TapClose() {
	p.executor.Tap(action.RandomIn(p.feature.Page.CloseBtn))
	p.executor.Sleep(1200)
}

// EnsureItemTab 确认落在「道具交易所」页：点 Tab 后校验页面身份。
func (p *Page) EnsureItemTab() bool {
	tab := p.feature.Tab.ItemExchange
	if action.RegionConfigured(tab) {
		p.executor.Tap(action.RandomIn(tab))
		p.executor.Sleep(800)
	}
	return p.IsCurrent()
}

func (p *Page) hasNextPage() bool {
	a := p.feature.Page.Arrow
	if a.Colors == "" {
		return false
	}
	_, ok := p.detector.FindColor(a.Region, a.Colors, a.Sim, a.Dir)
	return ok
}

// IsLastPage 右箭头不可见即列表已到右侧尽头。
func (p *Page) IsLastPage() bool {
	return !p.hasNextPage()
}

// SwipeNextPage 右翻一页；已在末页返回 false。
func (p *Page) SwipeNextPage() bool {
	if p.IsLastPage() {
		logger.Infof("[Market] 右箭头不可见，列表已到右侧尽头")
		return false
	}
	s := p.feature.List.Swipe
	p.executor.Swipe(s.From, s.To, s.DurationMs)
	p.executor.Sleep(700)
	return true
}

// slotTapRect 购买按钮点击矩形（锚点 + 固定偏移）。
func (p *Page) slotTapRect(pt screen.Point) screen.Region {
	s := p.feature.Slot
	cx, cy := pt.X, pt.Y+s.BuyBtnOffsetY
	return screen.Region{X1: cx - s.BuyBtnHalfW, Y1: cy - s.BuyBtnHalfH, X2: cx + s.BuyBtnHalfW, Y2: cy + s.BuyBtnHalfH}
}

// slotCrateRect 售罄 OCR 矩形（锚点 + 固定偏移）。
func (p *Page) slotCrateRect(pt screen.Point) screen.Region {
	s := p.feature.Slot
	cy := pt.Y + s.CrateOffsetY
	return screen.Region{X1: pt.X - s.CrateHalfW, Y1: cy - s.CrateHalfH, X2: pt.X + s.CrateHalfW, Y2: cy + s.CrateHalfH}
}

// IsSlotSoldOut 货架格子售罄判定（OCR 命中「售罄」）。
func (p *Page) IsSlotSoldOut(pt screen.Point) bool {
	text, err := p.detector.OCRText(p.slotCrateRect(pt))
	if err != nil {
		logger.Warnf("[Market] sold-out OCR failed: %v", err)
		return false
	}
	return strings.Contains(text, "售罄")
}

func (p *Page) IsConfirmDialog() bool {
	id := p.feature.Dialog.Identify
	if id.Colors == "" {
		return false
	}
	return p.detector.MatchMultiColor(id.Colors, id.Sim)
}

func (p *Page) WaitConfirmDialog(timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if p.IsConfirmDialog() {
			return true
		}
		p.executor.Sleep(300)
	}
	return false
}

func (p *Page) TapDialogConfirm() {
	p.executor.Tap(action.RandomIn(p.feature.Dialog.Confirm))
	p.executor.Sleep(800)
}

func (p *Page) TapDialogClose() {
	p.executor.Tap(action.RandomIn(p.feature.Dialog.Cancel))
	p.executor.Sleep(800)
}

func (p *Page) IsItemShortageDialog() bool {
	id := p.feature.Shortage.Identify
	if id.Colors == "" {
		return false
	}
	return p.detector.MatchMultiColor(id.Colors, id.Sim)
}

func (p *Page) TapItemShortageCancel() {
	p.executor.Tap(action.RandomIn(p.feature.Shortage.Cancel))
	p.executor.Sleep(800)
}

// TapShelfAndResolve 点货架购买按钮并走完确认弹窗流程：
// 等确认弹窗 → 点确认 → 观察结果：道具不足弹窗 → 取消(shortage)；
// 确认弹窗消失 → purchased；确认弹窗未出现/结果超时 → 尝试关闭并 failed。
func (p *Page) TapShelfAndResolve(pt screen.Point) string {
	p.executor.Tap(action.RandomIn(p.slotTapRect(pt)))
	p.executor.Sleep(800)

	if !p.WaitConfirmDialog(confirmDialogWait) {
		logger.Warnf("[Market] 点击货架后确认弹窗未出现")
		return ResultFailed
	}
	p.TapDialogConfirm()

	deadline := time.Now().Add(purchaseResultWait)
	for time.Now().Before(deadline) {
		if p.IsItemShortageDialog() {
			logger.Infof("[Market] 命中道具不足弹窗，取消本次购买")
			p.TapItemShortageCancel()
			if p.IsConfirmDialog() {
				p.TapDialogClose()
			}
			return ResultShortage
		}
		if !p.IsConfirmDialog() {
			return ResultPurchased
		}
		p.executor.Sleep(300)
	}
	logger.Warnf("[Market] 购买确认后结果未知，尝试关闭确认弹窗")
	if p.IsConfirmDialog() {
		p.TapDialogClose()
	}
	return ResultFailed
}

// IsFreeRefresh 刷新按钮状态为「免费刷新」。
func (p *Page) IsFreeRefresh() bool {
	text, err := p.detector.OCRText(p.feature.Page.RefreshOcr)
	if err != nil {
		logger.Warnf("[Market] free-refresh OCR failed: %v", err)
		return false
	}
	return strings.Contains(text, "免费刷新")
}

var reRestock = regexp.MustCompile(`(\d+):(\d+):(\d+)`)

// ReadRestockSeconds 读补货倒计时。返回 (秒, 原始文本, ok)：
// ok=true 且秒=0 → 「免费刷新」（无等待）；ok=false → 读数/解析失败。
func (p *Page) ReadRestockSeconds() (int, string, bool) {
	text, err := p.detector.OCRText(p.feature.Page.RefreshOcr)
	if err != nil {
		logger.Warnf("[Market] restock OCR failed: %v", err)
		return 0, "", false
	}
	raw := strings.TrimSpace(text)
	if raw == "" {
		return 0, raw, false
	}
	if strings.Contains(raw, "免费刷新") {
		return 0, raw, true
	}
	if m := reRestock.FindStringSubmatch(raw); m != nil {
		h, _ := strconv.Atoi(m[1])
		min, _ := strconv.Atoi(m[2])
		sec, _ := strconv.Atoi(m[3])
		return h*3600 + min*60 + sec, raw, true
	}
	logger.Warnf("[Market] 补货倒计时 OCR 解析失败: %q", raw)
	return 0, raw, false
}

// TapRefresh 点免费刷新，并等「免费刷新」字样消失（最长 10s）。
func (p *Page) TapRefresh() {
	p.executor.Tap(action.RandomIn(p.feature.Page.RefreshBtn))
	p.executor.Sleep(1200)
	p.executor.Sleep(1000)
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if !p.IsFreeRefresh() {
			return
		}
		p.executor.Sleep(500)
	}
}

// StockKeys 全部已配置商品键名（排序保证确定性）。
func (p *Page) StockKeys() []string {
	keys := make([]string, 0, len(p.feature.Stock))
	for k := range p.feature.Stock {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func isNearExisting(pt screen.Point, list []screen.Point, radius int) bool {
	r2 := radius * radius
	for _, q := range list {
		dx, dy := pt.X-q.X, pt.Y-q.Y
		if dx*dx+dy*dy <= r2 {
			return true
		}
	}
	return false
}

// collectVisibleTargets 收集当前屏可见的清单商品锚点：跨商品去重（半径 80）后按 X 排序。
func (p *Page) collectVisibleTargets(items []string) []stockTarget {
	var targets []stockTarget
	var seen []screen.Point
	for _, key := range items {
		def, ok := p.feature.Stock[key]
		if !ok || def.Colors == "" {
			continue
		}
		for _, pt := range p.detector.FindMultiColorsAll(def.Region, def.Colors, def.Sim, def.Dir) {
			if isNearExisting(pt, seen, dedupRadius) {
				continue
			}
			seen = append(seen, pt)
			targets = append(targets, stockTarget{Point: pt, Key: key})
		}
	}
	sort.Slice(targets, func(i, j int) bool { return targets[i].Point.X < targets[j].Point.X })
	return targets
}

// PurchaseWishlist 按购买清单扫货：逐页收集可见商品 → 售罄跳过 → 逐个购买，
// 右翻直到末页或达到 MaxSwipes（Lua purchaseWishlist）。
func (p *Page) PurchaseWishlist(items []string) PurchaseStats {
	var stats PurchaseStats
	valid := items[:0]
	for _, key := range items {
		if def, ok := p.feature.Stock[key]; ok && def.Colors != "" {
			valid = append(valid, key)
		} else {
			logger.Warnf("[Market] 未配置 Stock: %s", key)
		}
	}
	if len(valid) == 0 {
		logger.Warnf("[Market] 无可购买道具配置")
		return stats
	}
	maxSwipes := p.feature.List.MaxSwipes
	if maxSwipes <= 0 {
		maxSwipes = 20
	}
	swipes := 0
	for swipes <= maxSwipes {
		var visited []screen.Point
		targets := p.collectVisibleTargets(valid)
		logger.Infof("[Market] 扫描可见商品 目标命中:%d 滑动:%d", len(targets), swipes)
		for _, t := range targets {
			if isNearExisting(t.Point, visited, dedupRadius) {
				continue
			}
			visited = append(visited, t.Point)
			if p.IsSlotSoldOut(t.Point) {
				stats.SoldOut++
				logger.Infof("[Market] %s 已售罄，跳过", t.Key)
				continue
			}
			logger.Infof("[Market] 尝试购买 %s", t.Key)
			switch p.TapShelfAndResolve(t.Point) {
			case ResultPurchased:
				stats.Purchased++
			case ResultShortage:
				stats.Shortage++
			default:
				stats.Failed++
			}
		}
		if p.IsLastPage() {
			logger.Infof("[Market] 已是最后一页，结束扫货")
			break
		}
		if !p.SwipeNextPage() {
			break
		}
		swipes++
	}
	return stats
}
