package main

import (
	"app/internal/config"
	"app/ui"
)

// starlightTaskDescriptor 梦幻繁星岛的面板描述符（仅启用开关）。
func starlightTaskDescriptor(cfg *config.Config) ui.Task {
	st := &cfg.Tasks.Starlight
	return ui.Task{
		ID:         "starlight",
		Title:      "梦幻繁星岛",
		Category:   "daily",
		EnabledKey: "starlight_enabled",
		Fields: []ui.Field{
			ui.Bool("starlight_enabled", "启用",
				func() bool { return st.Enabled },
				func(v bool) { st.Enabled = v }),
		},
		Summary: func(s *ui.Store) string {
			if s.GetBool("starlight_enabled") {
				return "已启用"
			}
			return "未启用"
		},
	}
}
