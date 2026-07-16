package ui

type ShellOptions struct {
	Title            string
	ConfigPath       string
	DataStorePath    string // 业务 KV，如 /sdcard/shuaibin-cookie/store.json；清除缓存时删除
	Store            *Store
	Controller       *SessionController
	OpenPanelOnStart bool
	// Reseed 在清除缓存后回填默认控件值（通常为 SeedFromConfig）。
	Reseed func(*Store)
}
