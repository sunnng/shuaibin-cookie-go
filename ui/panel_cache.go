package ui

import (
	"os"
)

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
	return nil
}
