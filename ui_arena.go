package main

import (
	"fmt"

	"app/internal/config"
	"app/ui"
)

// arenaTaskDescriptor 王国竞技场的面板描述符（ADR-0002）：字段键与旧
// internal/ui/binding.go 完全一致，设备上已有的 ui.json 无缝兼容。
// 详情页由框架 Form 按 Fields 自动渲染（无 RenderDetail 逃生门需求）。
func arenaTaskDescriptor(cfg *config.Config) ui.Task {
	a := &cfg.Tasks.Arena
	return ui.Task{
		ID:         "arena",
		Title:      "王国竞技场",
		Category:   "日常",
		EnabledKey: "arena_enabled",
		Fields: []ui.Field{
			ui.Bool("arena_enabled", "启用",
				func() bool { return a.Enabled },
				func(v bool) { a.Enabled = v }),
			ui.Number("arena_max_battles", "每日战斗上限", 0, 999, 1,
				func() int {
					if a.MaxBattles == nil {
						return 0 // 0 = 不限
					}
					return *a.MaxBattles
				},
				func(v int) {
					if v > 0 {
						a.MaxBattles = &v
					} else {
						a.MaxBattles = nil
					}
				}),
			ui.Number("arena_auto_buy_count", "自动购买次数", 0, 999, 1,
				func() int { return a.AutoBuyCount },
				func(v int) { a.AutoBuyCount = v }),
			ui.Number("arena_trophy_diff", "奖杯差阈值", 0, 999, 1,
				func() int { return a.TrophyDiff },
				func(v int) { a.TrophyDiff = v }),
		},
		Summary: func(s *ui.Store) string {
			max := int(s.GetFloat("arena_max_battles"))
			maxLabel := "不限"
			if max > 0 {
				maxLabel = fmt.Sprintf("%d", max)
			}
			return fmt.Sprintf("上限 %s · 购买 %d · 奖杯差 %d",
				maxLabel, int(s.GetFloat("arena_auto_buy_count")), int(s.GetFloat("arena_trophy_diff")))
		},
	}
}
