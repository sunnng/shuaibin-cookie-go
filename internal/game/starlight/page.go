package starlight

import (
	"time"

	"app/internal/logger"
	"app/internal/platform/action"
	"app/internal/platform/screen"
)

// Page 梦幻繁星岛的页面识别与交互 API，对应 Lua 繁星岛_页面.lua。
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

func (p *Page) match(id screen.Feature) bool {
	if id.Colors == "" {
		return false
	}
	return p.detector.MatchMultiColor(id.Colors, id.Sim)
}

// waitFor 轮询直到 check 命中或超时，对应 Lua Color.waitMatch。
func (p *Page) waitFor(check func() bool, timeout, interval time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if check() {
			return true
		}
		p.executor.Sleep(int(interval / time.Millisecond))
	}
	return check()
}

// ========== 首页 ==========

func (p *Page) IsHomePage() bool {
	return p.match(p.feature.Home.Identify)
}

func (p *Page) WaitHomePage(timeout time.Duration) bool {
	return p.waitFor(p.IsHomePage, timeout, 500*time.Millisecond)
}

func (p *Page) TapSailingManual() bool {
	return p.tapArea(p.feature.Home.Actions.ManualBtn, "航海手册_按钮", 1000)
}

func (p *Page) TapTaskBtn() bool {
	return p.tapArea(p.feature.Home.Actions.TaskBtn, "taskBtn", 1000)
}

func (p *Page) TapBackToKingdom() bool {
	return p.tapArea(p.feature.Home.Actions.BackBtn, "home.backBtn", 1200)
}

// ========== 航海手册页 ==========

func (p *Page) IsManualPage() bool {
	return p.match(p.feature.Manual.Identify)
}

func (p *Page) WaitManualPage(timeout time.Duration) bool {
	return p.waitFor(p.IsManualPage, timeout, 500*time.Millisecond)
}

func (p *Page) TapLoginIsland() bool {
	return p.tapArea(p.feature.Manual.Actions.LoginIslandBtn, "登陆回忆小岛_按钮", 1000)
}

// ========== 纯香草小岛页 ==========

func (p *Page) IsVanillaIslandPage() bool {
	return p.match(p.feature.Vanilla.Identify)
}

func (p *Page) WaitVanillaIslandPage(timeout time.Duration) bool {
	return p.waitFor(p.IsVanillaIslandPage, timeout, 500*time.Millisecond)
}

func (p *Page) TapBackFromVanilla() bool {
	return p.tapArea(p.feature.Vanilla.Actions.BackBtn, "纯香草小岛.backBtn", 1200)
}

// ========== 任务页 ==========

func (p *Page) IsTaskPage() bool {
	return p.match(p.feature.Task.Identify)
}

func (p *Page) WaitTaskPage(timeout time.Duration) bool {
	return p.waitFor(p.IsTaskPage, timeout, 500*time.Millisecond)
}

func (p *Page) TapBackFromTask() bool {
	return p.tapArea(p.feature.Task.Actions.BackBtn, "任务.backBtn", 1200)
}

// FindClaimableBtn 在任务页查找可领奖按钮，返回首个命中点。
func (p *Page) FindClaimableBtn() (screen.Point, bool) {
	c := p.feature.Task.Claim
	if c.Colors == "" {
		logger.Warnf("[梦幻繁星岛.页面] 可领奖_按钮 未配置")
		return screen.Point{}, false
	}
	pts := p.detector.FindMultiColorsAll(c.Region, c.Colors, c.Sim, c.Dir)
	if len(pts) == 0 {
		return screen.Point{}, false
	}
	return pts[0], true
}

// TapClaimableBtn 点击可领奖按钮（对应 Lua Touch.tapR）。
func (p *Page) TapClaimableBtn(pt screen.Point) {
	p.executor.Tap(action.Point{X: pt.X, Y: pt.Y})
	p.executor.Sleep(800)
}

// SettleAfterClaim 领奖后等待 2 秒；每 500ms 跑一次守卫扫描，对应 Lua
// Guard.sleep(2000, 500)。check 为 nil 时只睡眠。
func (p *Page) SettleAfterClaim(check func() bool) {
	for i := 0; i < 4; i++ {
		if check != nil {
			check() // 守卫处理弹窗后继续剩余等待
		}
		p.executor.Sleep(500)
	}
}

// DismissRewardPopupIfNeeded 领奖后若任务页特征未恢复，点屏幕中央空白处尝试关闭弹窗。
func (p *Page) DismissRewardPopupIfNeeded() {
	if p.IsTaskPage() {
		return
	}
	p.executor.Tap(action.Point{X: p.feature.Task.DismissPoint.X, Y: p.feature.Task.DismissPoint.Y})
	p.executor.Sleep(500)
	p.waitFor(p.IsTaskPage, 5*time.Second, 300*time.Millisecond)
}

// ========== 王国活动页（导航中转） ==========

func (p *Page) IsEventPage() bool {
	return p.match(p.feature.Event.Identify)
}

func (p *Page) WaitEventPage(timeout time.Duration) bool {
	return p.waitFor(p.IsEventPage, timeout, 500*time.Millisecond)
}

func (p *Page) TapStarlightEntry() bool {
	return p.tapArea(p.feature.Event.Actions.StarlightBtn, "梦幻繁星岛_按钮", 1500)
}

// tapArea 在区域内随机点并等待 delayMs；区域未配置时告警并返回 false。
func (p *Page) tapArea(r screen.Region, name string, delayMs int) bool {
	if !action.RegionConfigured(r) {
		logger.Warnf("[梦幻繁星岛.页面] %s 未配置", name)
		return false
	}
	p.executor.Tap(action.RandomIn(r))
	p.executor.Sleep(delayMs)
	return true
}
