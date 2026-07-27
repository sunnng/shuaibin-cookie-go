package taskdesc

import (
	"fmt"

	"app/internal/config"
	"app/ui"
)

// Square 布谷鸟广场的面板描述符。
func Square(cfg *config.Config) ui.Task {
	q := &cfg.Tasks.Square
	return ui.Task{
		ID:         "square",
		Title:      "布谷鸟广场",
		Category:   "daily",
		EnabledKey: "square_enabled",
		Fields: []ui.Field{
			ui.Bool("square_enabled", "启用",
				func() bool { return q.Enabled },
				func(v bool) { q.Enabled = v }),
			ui.Number("square_daily_cap", "每日奖励上限", 1, 9999, 1,
				func() int { return q.DailyCap },
				func(v int) { q.DailyCap = v }),
			ui.Number("square_check_interval_sec", "有效停留时长(秒)", 60, 3600, 10,
				func() int { return q.CheckIntervalSec },
				func(v int) { q.CheckIntervalSec = v }),
			ui.Number("square_chunk_sec", "睡眠粒度(秒)", 1, 60, 1,
				func() int { return q.ChunkSec },
				func(v int) { q.ChunkSec = v }),
		},
		Summary: func(s *ui.Store) string {
			return fmt.Sprintf("日上限 %d · 停留 %ds · 粒度 %ds",
				int(s.GetFloat("square_daily_cap")),
				int(s.GetFloat("square_check_interval_sec")),
				int(s.GetFloat("square_chunk_sec")))
		},
	}
}
