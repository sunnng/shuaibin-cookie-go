//go:build android && cgo

package ui

import (
	"sync"

	"github.com/Dasongzi1366/AutoGo/imgui"
)

var builtinModulesOnce sync.Once

// EnsureBuiltinModules 注册内置任务模块（幂等）。Android 面板首次绘制前调用。
func EnsureBuiltinModules() {
	builtinModulesOnce.Do(func() {
		RegisterModule(ModuleDef{
			ID:         "arena",
			Title:      "王国竞技场",
			Category:   CategoryDaily,
			EnabledKey: KeyArenaEnabled,
			Summary:    ArenaSummary,
			RenderDetail: func(store *Store) {
				UI_创建复选框(store, KeyArenaEnabled, "启用")
				UI_创建数字输入框(store, KeyArenaMaxBattles, "每日战斗上限", "0=不限", float32(-1), float32(0), "#d6d8df", float64(1), float64(0), float64(999))
				UI_创建数字输入框(store, KeyArenaAutoBuy, "自动购买次数", "0", float32(-1))
				UI_创建数字输入框(store, KeyArenaTrophyDiff, "奖杯差阈值", "0", float32(-1))
			},
		})
		RegisterModule(ModuleDef{
			ID:         "collect",
			Title:      "收集",
			Category:   CategoryDaily,
			EnabledKey: KeyCollectEnabled,
			Summary:    CollectSummary,
			RenderDetail: func(store *Store) {
				UI_创建复选框(store, KeyCollectEnabled, "启用")
				imgui.TextWrapped("骨架占位：启用后调度器跑空状态机（detect→Done）。")
			},
		})
	})
}
