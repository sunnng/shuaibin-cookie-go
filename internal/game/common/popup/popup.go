// Package popup 通用弹窗守卫（Lua 通用_弹窗/page.lua 迁移）。
package popup

import (
	"app/internal/guard"
	"app/internal/platform/action"
	"app/internal/platform/screen"
)

// networkUnstable「网络联机状态不稳定」弹窗（Lua 特征颜色串，| 已转为 ,）。
var networkUnstable = screen.Feature{
	Colors: "468,224,6a719f-101010,1132,235,363d5f-101010,789,415,505050-101010," +
		"790,462,505050-101010,476,672,dbcfc6-101010,1140,673,aea09b-101010," +
		"855,634,7ace0e-101010,795,631,ffffff-101010,695,613,95d83e-101010",
	Sim: 0.9,
}

// confirmRegion 「确认」按钮区域（1600×900 基准坐标）。
var confirmRegion = screen.Region{X1: 775, Y1: 621, X2: 828, Y2: 647}

const (
	// waitGoneMs 点确认后轮询等弹窗消失的总时长（Lua waitGoneMs=2000）。
	waitGoneMs = 2000
	// pollMs 轮询间隔（Lua 每 200ms 查一次特征）。
	pollMs = 200
)

// Register 把「网络联机状态不稳定」弹窗注册进守卫（priority 10）。
// handler：点确认按钮中心，然后每 200ms 查一次特征，消失或 2s 超时返回。
func Register(g *guard.Guard, exec action.Executor) {
	det := g.Detector()
	g.Register("网络联机状态不稳定", networkUnstable, func() error {
		exec.Tap(screen.Point{
			X: (confirmRegion.X1 + confirmRegion.X2) / 2,
			Y: (confirmRegion.Y1 + confirmRegion.Y2) / 2,
		})
		for elapsed := 0; elapsed < waitGoneMs; elapsed += pollMs {
			exec.Sleep(pollMs)
			if det == nil || !det.MatchMultiColor(networkUnstable.Colors, networkUnstable.Sim) {
				return nil
			}
		}
		return nil
	}, 10)
}
