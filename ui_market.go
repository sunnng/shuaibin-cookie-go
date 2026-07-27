package main

import (
	"fmt"

	"app/internal/config"
	"app/ui"
)

// marketTaskDescriptor 海滩交易所的面板描述符。
func marketTaskDescriptor(cfg *config.Config) ui.Task {
	m := &cfg.Tasks.SeasideMarket
	return ui.Task{
		ID:         "seaside_market",
		Title:      "海滩交易所",
		Category:   "daily",
		EnabledKey: "market_enabled",
		Fields: []ui.Field{
			ui.Bool("market_enabled", "启用",
				func() bool { return m.Enabled },
				func(v bool) { m.Enabled = v }),
			ui.Text("market_items", "购买清单(逗号分隔)",
				func() string { return joinList(m.Items) },
				func(v string) { m.Items = splitList(v) }),
			ui.Number("market_restock_buffer_sec", "补货缓冲(秒)", 0, 600, 10,
				func() int { return m.RestockBufferSec },
				func(v int) { m.RestockBufferSec = v }),
		},
		Summary: func(s *ui.Store) string {
			return fmt.Sprintf("清单 %d 项 · 缓冲 %ds",
				len(splitList(s.GetString("market_items"))),
				int(s.GetFloat("market_restock_buffer_sec")))
		},
	}
}
