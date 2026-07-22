//go:build android && cgo

package ui

import (
	"fmt"

	"github.com/Dasongzi1366/AutoGo/imgui"
)

// SystemPage 标准系统页：配置持久化（保存/清缓存）。应用挂载为导航条目。
// 反馈文案经组件状态持有（替代旧包级 settingsStatus）。
func SystemPage() func(*Ctx) {
	return func(ctx *Ctx) {
		shell := ctx.Shell
		if shell == nil {
			return
		}
		store := shell.Store()
		status := State(ctx, "sysStatus", "")

		imgui.Text("系统")
		imgui.Dummy(imgui.Vec2{X: 0, Y: 4})
		imgui.TextDisabled("配置持久化")
		imgui.Separator()
		imgui.Dummy(imgui.Vec2{X: 0, Y: 4})
		imgui.TextDisabled("配置文件  " + shell.ConfigPath())
		imgui.Dummy(imgui.Vec2{X: 0, Y: 8})

		Row(ctx,
			func() {
				Button(ctx, ButtonProps{Label: "保存配置", Kind: ButtonSecondary, OnClick: func() {
					if err := store.SaveConfig(shell.ConfigPath()); err != nil {
						*status = fmt.Sprintf("保存失败：%v", err)
						return
					}
					*status = "配置已保存"
				}})
			},
			func() {
				Button(ctx, ButtonProps{Label: "清除缓存", Kind: ButtonSecondary, OnClick: func() {
					if shell.ScriptState() != StateIdle {
						_ = shell.StartStop() // 停止脚本
					}
					if err := ClearPanelCache(store, shell.ConfigPath(), shell.DataStorePath(), func(*Store) { shell.Seed() }); err != nil {
						*status = fmt.Sprintf("清除失败：%v", err)
						return
					}
					*status = "缓存已清除，默认配置已恢复"
				}})
			},
		)

		if *status != "" {
			imgui.Spacing()
			imgui.TextWrapped(*status)
		}
	}
}
