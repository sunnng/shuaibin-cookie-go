package square

import (
	"strconv"
	"strings"
	"time"
	"unicode"

	"app/internal/logger"
	"app/internal/platform/action"
	"app/internal/platform/screen"
)

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

// IsSquare 当前是否在布谷鸟广场页（Lua SquarePage.isCurrent）。
func (p *Page) IsSquare() bool {
	id := p.feature.Home.Identify
	if id.Colors == "" {
		return false
	}
	return p.detector.MatchMultiColor(id.Colors, id.Sim)
}

// WaitSquare 轮询等待进入广场页（Lua SquarePage.waitHome）。
func (p *Page) WaitSquare(timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if p.IsSquare() {
			return true
		}
		p.executor.Sleep(500)
	}
	return false
}

// IsLeaveDialog 当前是否在「离开广场」弹窗（Lua SquarePage.isLeaveDialog）。
func (p *Page) IsLeaveDialog() bool {
	id := p.feature.Dialog.Identify
	if id.Colors == "" {
		return false
	}
	return p.detector.MatchMultiColor(id.Colors, id.Sim)
}

// WaitLeaveDialog 轮询等待「离开广场」弹窗出现（Lua SquarePage.waitLeaveDialog）。
func (p *Page) WaitLeaveDialog(timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if p.IsLeaveDialog() {
			return true
		}
		p.executor.Sleep(500)
	}
	return false
}

// TapEntryBtn 王国首页点广场入口（Lua Touch.tapArea(home.squareBtn)）。
func (p *Page) TapEntryBtn() {
	p.executor.Tap(action.RandomIn(p.feature.EntryBtn))
	p.executor.Sleep(1200)
}

// TapBack 广场页点返回键，弹出「离开广场」弹窗（Lua SquarePage.tapBack）。
func (p *Page) TapBack() {
	p.executor.Tap(action.RandomIn(p.feature.Home.Actions.BackBtn))
	p.executor.Sleep(1000)
}

// TapCloseDialog 点弹窗右上 X 关闭，留在广场（Lua SquarePage.tapCloseDialog）。
func (p *Page) TapCloseDialog() {
	p.executor.Tap(action.RandomIn(p.feature.Dialog.Actions.CancelBtn))
	p.executor.Sleep(1200)
}

// TapReturnKingdom 点弹窗「离开」按钮回王国主城（Lua SquarePage.tapReturnKingdom）。
func (p *Page) TapReturnKingdom() {
	p.executor.Tap(action.RandomIn(p.feature.Dialog.Actions.LeaveBtn))
	p.executor.Sleep(1200)
}

// TapClaimAll 点「一次领回」领取奖励（Lua SquarePage.tapClaimAll）。
func (p *Page) TapClaimAll() {
	p.executor.Tap(action.RandomIn(p.feature.Dialog.Actions.ConfirmRewardBtn))
	p.executor.Sleep(1000)
}

// TapUntilDialog 领奖后反复点 TapUntilRect，直到离开弹窗重新出现
// tapUntilDialogTimeout 是 TapUntilDialog 的上限（Lua tapUntilMatch 默认 15s）；
// 变量而非常量，测试可缩短。
var tapUntilDialogTimeout = 15 * time.Second

// （Lua SquarePage.tapUtilDialog → Color.tapUntilMatch，默认 15s 超时）。
func (p *Page) TapUntilDialog() bool {
	deadline := time.Now().Add(tapUntilDialogTimeout)
	for time.Now().Before(deadline) {
		if p.IsLeaveDialog() {
			return true
		}
		p.executor.Tap(action.RandomIn(p.feature.Dialog.Actions.TapUntilRect))
		p.executor.Sleep(800)
		if p.IsLeaveDialog() {
			return true
		}
		p.executor.Sleep(500)
	}
	return p.IsLeaveDialog()
}

// RewardNow OCR 读取「目前可获得奖励」数量。
func (p *Page) RewardNow() (int, bool) {
	return p.readCount(p.feature.Dialog.Reads.RewardNow, "目前可获得奖励")
}

// RewardTotal OCR 读取「累计获得奖励」数量。
func (p *Page) RewardTotal() (int, bool) {
	return p.readCount(p.feature.Dialog.Reads.RewardTotal, "累计获得奖励")
}

// readCount 对齐 Lua readCount：OCR 文本剔除全部非数字字符后解析；
// 有字无数时告警。Go 的 Detector 没有独立数字模式，退化为单次 OCRText。
func (p *Page) readCount(r screen.Region, label string) (int, bool) {
	if r == (screen.Region{}) {
		logger.Warnf("[Square] OCR 区域未配置: %s", label)
		return 0, false
	}
	text, err := p.detector.OCRText(r)
	if err != nil {
		logger.Warnf("[Square] %s OCR failed: %v", label, err)
		return 0, false
	}
	digits := stripNonDigits(text)
	if digits == "" {
		if strings.TrimSpace(text) != "" {
			logger.Warnf("[Square] %s OCR 有字无数: %s", label, strings.TrimSpace(text))
		}
		return 0, false
	}
	n, err := strconv.Atoi(digits)
	if err != nil {
		return 0, false
	}
	return n, true
}

// IsDailyRewardsMaxed 判定每日奖励是否已满额
// （Lua SquarePage.isDailyRewardsMaxed / isFinishOcr）。
func (p *Page) IsDailyRewardsMaxed() bool {
	if !p.IsLeaveDialog() {
		return false
	}
	r := p.feature.Dialog.Reads.IsFinish
	if r == (screen.Region{}) {
		r = p.feature.Dialog.Reads.DailyMax
	}
	text, err := p.detector.OCRText(r)
	if err != nil {
		logger.Warnf("[Square] 满额标识 OCR failed: %v", err)
		return false
	}
	if textIndicatesMaxed(text) {
		logger.Infof("[Square] 满额标识 OCR: %s", strings.TrimSpace(text))
		return true
	}
	return false
}

// ReadRewardSum 读取 目前+累计 奖励并求和（Lua SquarePage.readRewardSum）。
// 任一 OCR 失败则 ok=false；pending/total 返回各自已读到的值（未读到为 0）。
func (p *Page) ReadRewardSum() (pending, total, sum int, ok bool) {
	if !p.IsLeaveDialog() {
		logger.Warnf("[Square] 不在离开广场弹窗，无法 OCR")
		return 0, 0, 0, false
	}
	pending, okP := p.RewardNow()
	total, okT := p.RewardTotal()
	if !okP || !okT {
		logger.Warnf("[Square] 奖励 OCR 失败 目前=%v 累计=%v", okP, okT)
		return pending, total, 0, false
	}
	sum = pending + total
	logger.Infof("[Square] 奖励 可获得=%d 累计=%d 总计=%d", pending, total, sum)
	return pending, total, sum, true
}

// SleepMs 供任务层做 Lua Guard.sleep 形态的短等待（页面间缓冲）。
func (p *Page) SleepMs(ms int) {
	p.executor.Sleep(ms)
}

// textIndicatesMaxed 对齐 Lua textIndicatesMaxed：
// 含「最大」，或同时含「已领取」与「奖励」。
func textIndicatesMaxed(text string) bool {
	if strings.Contains(text, "最大") {
		return true
	}
	return strings.Contains(text, "已领取") && strings.Contains(text, "奖励")
}

func stripNonDigits(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsDigit(r) {
			return r
		}
		return -1
	}, s)
}
