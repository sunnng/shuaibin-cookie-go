package ui

import "app/internal/status"

// 设备上的持久化路径，全项目唯一来源（main、shell、panel 共用）。
const (
	DefaultDataDir    = "/sdcard/shuaibin-cookie"
	DefaultConfigPath = DefaultDataDir + "/ui.json"
	DefaultStorePath  = DefaultDataDir + "/store.json"
)

type ShellOptions struct {
	Title            string
	ConfigPath       string
	DataStorePath    string // 业务 KV，如 DefaultStorePath；清除缓存时删除
	Store            *Store
	Controller       *ScriptController
	OpenPanelOnStart bool
	// Status 任务状态上报通道；非 nil 且脚本运行中时，灵动岛展示任务状态文本。
	Status *status.Reporter
	// Reseed 在清除缓存后回填默认控件值（通常为 SeedFromConfig）。
	Reseed func(*Store)
}
