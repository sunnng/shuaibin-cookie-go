// Package taskdesc 应用侧的任务面板描述符（ADR-0002）：每个游戏任务一个
// ui.Task 描述符，字段键即设备 ui.json 的持久化键。框架包 ui 只负责渲染
// 与配置流，任务/字段的声明全部集中在该应用侧包，main.go 只调 Tasks 聚合。
package taskdesc

import (
	"app/internal/config"
	"app/ui"
)

// Tasks 返回全部任务的面板描述符（顺序即面板分组内的展示顺序）。
func Tasks(cfg *config.Config) []ui.Task {
	tasks := append([]ui.Task{Arena(cfg)}, Mine(cfg)...)
	return append(tasks,
		Square(cfg),
		Starlight(cfg),
		Market(cfg),
		Biscuit(cfg),
	)
}
