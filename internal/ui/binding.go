package ui

import (
	"app/internal/config"
)

const (
	KeyArenaEnabled    = "arena_enabled"
	KeyArenaMaxBattles = "arena_max_battles"
	KeyArenaAutoBuy    = "arena_auto_buy_count"
	KeyArenaTrophyDiff = "arena_trophy_diff"
	KeyCollectEnabled  = "collect_enabled"
	KeyLicense         = "license_key" // 卡密占位，验证逻辑后续接入
)

// SeedFromConfig 用 config.json 的默认值填充 Store（仅填充尚未存在的 key）。
func SeedFromConfig(store *Store, cfg *config.Config) {
	if store == nil || cfg == nil {
		return
	}
	seedBool(store, KeyArenaEnabled, cfg.Modules.Arena.Enabled)
	seedFloat(store, KeyArenaAutoBuy, float64(cfg.Modules.Arena.AutoBuyCount))
	seedFloat(store, KeyArenaTrophyDiff, float64(cfg.Modules.Arena.TrophyDiff))
	if cfg.Modules.Arena.MaxBattles != nil {
		seedFloat(store, KeyArenaMaxBattles, float64(*cfg.Modules.Arena.MaxBattles))
	} else {
		seedFloat(store, KeyArenaMaxBattles, 0)
	}
	seedBool(store, KeyCollectEnabled, cfg.Modules.Collect.Enabled)
	seedString(store, KeyLicense, "")
	EnsureBuiltinModules()
	SeedHubDefaults(store)
}

// ApplyToConfig 将 Store 中的 UI 选项写回 config。
func ApplyToConfig(store *Store, cfg *config.Config) {
	if store == nil || cfg == nil {
		return
	}
	cfg.Modules.Arena.Enabled = store.GetBool(KeyArenaEnabled)
	cfg.Modules.Arena.AutoBuyCount = int(store.GetFloat(KeyArenaAutoBuy))
	cfg.Modules.Arena.TrophyDiff = int(store.GetFloat(KeyArenaTrophyDiff))

	maxBattles := int(store.GetFloat(KeyArenaMaxBattles))
	if maxBattles > 0 {
		cfg.Modules.Arena.MaxBattles = &maxBattles
	} else {
		cfg.Modules.Arena.MaxBattles = nil
	}

	cfg.Modules.Collect.Enabled = store.GetBool(KeyCollectEnabled)
}

func seedBool(store *Store, key string, v bool) {
	if !store.HasKey(key) {
		store.SetBool(key, v)
	}
}

func seedFloat(store *Store, key string, v float64) {
	if !store.HasKey(key) {
		store.SetFloat(key, v)
	}
}

func seedString(store *Store, key string, v string) {
	if !store.HasKey(key) {
		store.SetString(key, v)
	}
}
