package ui

import (
	"fmt"
	"os"
)

// CookiePanelOptions 配置 DefaultCookiePanel（由 RunShell 从 ShellOptions 注入）。
type CookiePanelOptions struct {
	ConfigPath    string
	DataStorePath string
	Controller    Controller
	Reseed        func(*Store)
}

var cookiePanelOpts CookiePanelOptions

// ConfigureCookiePanel 设置默认面板的路径与回调。
func ConfigureCookiePanel(opts CookiePanelOptions) {
	cookiePanelOpts = opts
}

// ArenaSummary 生成竞技场模块卡片摘要。
func ArenaSummary(store *Store) string {
	if store == nil {
		return ""
	}
	max := int(store.GetFloat(KeyArenaMaxBattles))
	buy := int(store.GetFloat(KeyArenaAutoBuy))
	diff := int(store.GetFloat(KeyArenaTrophyDiff))
	maxLabel := "不限"
	if max > 0 {
		maxLabel = fmt.Sprintf("%d", max)
	}
	return fmt.Sprintf("上限 %s · 购买 %d · 奖杯差 %d", maxLabel, buy, diff)
}

// CollectSummary 生成收集模块卡片摘要。
func CollectSummary(store *Store) string {
	if store == nil {
		return ""
	}
	if store.GetBool(KeyCollectEnabled) {
		return "已启用 · 骨架占位"
	}
	return "骨架占位"
}

// ClearPanelCache 清空 UI Store、ui.json 与业务 store.json，再按需回填默认值。
func ClearPanelCache(store *Store, configPath, dataStorePath string, reseed func(*Store)) error {
	if store != nil {
		store.Clear()
	}
	if configPath != "" {
		if err := os.Remove(configPath); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	if dataStorePath != "" {
		if err := os.Remove(dataStorePath); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	if reseed != nil && store != nil {
		reseed(store)
	}
	if store != nil {
		EnsureBuiltinModules()
		store.SetString(KeyHubNav, HubNavModules)
		store.SetString(KeyHubCat, HubCatAll)
		store.SetString(KeyHubFilter, "")
		store.SetString(KeyHubSelected, "")
		store.SetString(KeyLicenseStatus, "unverified")
		SeedHubDefaults(store)
	}
	return nil
}
