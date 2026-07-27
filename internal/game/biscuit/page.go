package biscuit

import (
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

// ReadEffects OCR 读取副词条。脆饼固定 4 条：OCR 多识别时截断，不足时
// 补 {Name:"未知"}。OCR 通道故障按空文本处理（对齐 Lua scan 失败 raw=""）。
func (p *Page) ReadEffects() []Effect {
	text, err := p.detector.OCRText(p.feature.Reads.Effects)
	if err != nil {
		logger.Warnf("[Biscuit] effects OCR failed: %v", err)
		text = ""
	}
	effects := parseRaw(text)
	if len(effects) > effectSlotCount {
		effects = effects[:effectSlotCount]
	}
	for len(effects) < effectSlotCount {
		effects = append(effects, Effect{Name: "未知"})
	}
	return effects
}

// TapReroll 点洗炼按钮并等动画。
func (p *Page) TapReroll() {
	p.executor.Tap(action.RandomIn(p.feature.Actions.Reroll))
	p.executor.Sleep(1000)
}

// ConfirmResetDialog 确认重置弹窗出现时点"今日不再显示"再点确认，返回是否处理。
func (p *Page) ConfirmResetDialog() bool {
	return p.confirmDialog(p.feature.Dialogs.ResetConfirm)
}

// ConfirmSameDialog 确认相同脆饼弹窗出现时点"今日不再显示"再点确认，返回是否处理。
func (p *Page) ConfirmSameDialog() bool {
	return p.confirmDialog(p.feature.Dialogs.SameConfirm)
}

func (p *Page) confirmDialog(d DialogDef) bool {
	if d.Identify.Colors == "" || !p.detector.MatchMultiColor(d.Identify.Colors, d.Identify.Sim) {
		return false
	}
	p.executor.Tap(action.RandomIn(d.DontShowAgain))
	p.executor.Sleep(1000)
	p.executor.Tap(action.RandomIn(d.Confirm))
	p.executor.Sleep(1000)
	return true
}
