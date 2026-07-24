//go:build android && cgo

package ui

import (
	"fmt"

	"github.com/Dasongzi1366/AutoGo/imgui"
)

// SystemPage 标准系统页（design-system.md §4.4）：两张卡片纵向排列，按风险分级——
// 「配置持久化」白卡（黄底主按钮 + 行内反馈条）与「危险区」红描边浅红底卡
// （清除缓存需二次确认：点击后按钮变为「确认清除？」，3 秒内再点生效）。
// 应用挂载为导航条目。反馈文案经组件状态持有。
func SystemPage() func(*Ctx) {
	return func(ctx *Ctx) {
		shell := ctx.Shell
		if shell == nil {
			return
		}
		store := shell.Store()
		status := State(ctx, "sysStatus", "")
		statusOK := State(ctx, "sysStatusOK", true)
		armed := State(ctx, "clearArmed", false)
		armedAt := State(ctx, "clearArmedAt", float64(0))

		// 二次确认 3 秒超时自动复位。
		if *armed && imgui.Time()-*armedAt > 3 {
			*armed = false
		}

		imgui.Dummy(imgui.Vec2{X: 0, Y: float32(ctx.S(8))})

		// 配置持久化卡。
		sysCard(ctx, "save", false, func() {
			imgui.Text("配置持久化")
			imgui.Dummy(imgui.Vec2{X: 0, Y: float32(ctx.S(8))})
			imgui.TextDisabled("配置文件  " + shell.ConfigPath())
			imgui.Dummy(imgui.Vec2{X: 0, Y: float32(ctx.S(14))})

			Button(ctx, ButtonProps{Label: "保存配置", Kind: ButtonPrimary, Width: 150, Height: 48, OnClick: func() {
				if err := store.SaveConfig(shell.ConfigPath()); err != nil {
					*status = fmt.Sprintf("保存失败：%v", err)
					*statusOK = false
					return
				}
				*status = "配置已保存"
				*statusOK = true
			}})

			if *status != "" {
				imgui.Dummy(imgui.Vec2{X: 0, Y: float32(ctx.S(12))})
				drawFeedbackBar(ctx, *status, *statusOK)
			}
		})

		// 危险区卡：红描边 + 浅红底，清除缓存二次确认。
		sysCard(ctx, "danger", true, func() {
			imgui.Text("危险区")
			imgui.Dummy(imgui.Vec2{X: 0, Y: float32(ctx.S(8))})
			imgui.TextDisabled("清除缓存将停止脚本、删除设备上的配置与业务数据，并恢复为默认配置。此操作不可撤销。")
			imgui.Dummy(imgui.Vec2{X: 0, Y: float32(ctx.S(14))})

			label := "清除缓存"
			if *armed {
				label = "确认清除？"
			}
			Button(ctx, ButtonProps{Label: label, Kind: ButtonDanger, Width: 170, Height: 48, OnClick: func() {
				if !*armed {
					*armed = true
					*armedAt = imgui.Time()
					return
				}
				*armed = false
				if shell.ScriptState() != StateIdle {
					_ = shell.StartStop() // 停止脚本
				}
				if err := ClearPanelCache(store, shell.ConfigPath(), shell.DataStorePath(), func(*Store) { shell.Seed() }); err != nil {
					*status = fmt.Sprintf("清除失败：%v", err)
					*statusOK = false
					return
				}
				*status = "缓存已清除，默认配置已恢复"
				*statusOK = true
			}})
		})
	}
}

// sysCard 系统页卡片：白底（危险区浅红底）3px 描边积木卡 + 4px 硬阴影，
// 高度随内容自适应（ChildFlagsAutoResizeY）；硬阴影用上一帧高度先画（一帧收敛）。
func sysCard(ctx *Ctx, id string, danger bool, content func()) {
	pad := float32(ctx.S(18))
	radius := float32(ctx.S(14))

	cardH := State(ctx, "cardH_"+id, float32(ctx.S(140)))
	pos := imgui.CursorScreenPos()
	avail := imgui.ContentRegionAvail().X

	// 硬阴影（画在卡片之前 = 垫在底下）。
	dl := imgui.WindowDrawList()
	dl.AddRectFilledV(
		imgui.Vec2{X: pos.X + 4, Y: pos.Y + 4},
		imgui.Vec2{X: pos.X + avail + 4, Y: pos.Y + *cardH + 4},
		imgui.ColorU32Vec4(candyInk), radius, imgui.DrawFlagsRoundCornersAll,
	)

	bg := candyRaised
	border := candyInk
	if danger {
		bg = candyDangerBg
		border = candyRed
	}
	imgui.PushStyleColorVec4(imgui.ColChildBg, bg)
	imgui.PushStyleColorVec4(imgui.ColBorder, border)
	imgui.PushStyleVarFloat(imgui.StyleVarChildBorderSize, 3)
	imgui.PushStyleVarFloat(imgui.StyleVarChildRounding, radius)
	imgui.PushStyleVarVec2(imgui.StyleVarWindowPadding, imgui.Vec2{X: pad, Y: pad})

	imgui.BeginChildStrV("syscard_"+id, imgui.Vec2{X: avail, Y: 0}, imgui.ChildFlagsBorders|imgui.ChildFlagsAutoResizeY, imgui.WindowFlagsNone)
	content()
	imgui.EndChild()

	imgui.PopStyleVarV(3)
	imgui.PopStyleColorV(2)

	*cardH = imgui.ItemRectSize().Y
	imgui.Dummy(imgui.Vec2{X: 0, Y: float32(ctx.S(18))})
}

// drawFeedbackBar 行内反馈条：成功浅绿底 / 失败浅红底，2px 墨描边 + 墨字。
func drawFeedbackBar(ctx *Ctx, text string, ok bool) {
	bg := candyOKBg
	if !ok {
		bg = candyErrBg
	}
	sz := measureLabelSize(text)
	padX, padY := float32(ctx.S(12)), float32(ctx.S(8))
	avail := imgui.ContentRegionAvail().X
	w := sz.X + padX*2
	if avail > 0 && w > avail {
		w = avail
	}
	h := sz.Y + padY*2
	pos := imgui.CursorScreenPos()
	drawBlock(imgui.WindowDrawList(), pos, imgui.Vec2{X: pos.X + w, Y: pos.Y + h}, bg, float32(ctx.S(10)), 2, 0)
	imgui.WindowDrawList().AddTextVec2V(
		imgui.Vec2{X: pos.X + padX, Y: pos.Y + padY},
		imgui.ColorU32Vec4(candyInk),
		text,
	)
	imgui.Dummy(imgui.Vec2{X: w, Y: h})
}
