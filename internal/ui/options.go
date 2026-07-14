package ui

type PanelOptions struct {
	Title        string
	ConfigPath   string
	CountdownSec float64
	Store        *Store
	Render       func(store *Store)
	OnRun        func(store *Store)
	OnClose      func(store *Store)
}

type ShellOptions struct {
	Title            string
	ConfigPath       string
	DataStorePath    string // 业务 KV，如 /sdcard/shuaibin-cookie/store.json；清除缓存时删除
	CountdownSec     float64
	Store            *Store
	Render           func(store *Store)
	Controller       Controller
	OpenPanelOnStart bool
	// Reseed 在清除缓存后回填默认控件值（通常为 SeedFromConfig）。
	Reseed func(*Store)
}
