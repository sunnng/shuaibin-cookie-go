//go:build !android || !cgo

package ui

// EnsureBuiltinModules 桌面 stub：仅注册元数据（无 RenderDetail），供单测与类型检查。
func EnsureBuiltinModules() {
	if len(Modules()) > 0 {
		return
	}
	RegisterModule(ModuleDef{
		ID:         "arena",
		Title:      "王国竞技场",
		Category:   CategoryDaily,
		EnabledKey: KeyArenaEnabled,
		Summary:    ArenaSummary,
	})
	RegisterModule(ModuleDef{
		ID:         "collect",
		Title:      "收集",
		Category:   CategoryDaily,
		EnabledKey: KeyCollectEnabled,
		Summary:    CollectSummary,
	})
}
