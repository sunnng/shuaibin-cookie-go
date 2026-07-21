package ui

import (
	"app/internal/config"
)

const (
	KeyArenaEnabled    = "arena_enabled"
	KeyArenaMaxBattles = "arena_max_battles"
	KeyArenaAutoBuy    = "arena_auto_buy_count"
	KeyArenaTrophyDiff = "arena_trophy_diff"
)

// SeedFromConfig 用 config.json 的默认值填充 Store（仅填充尚未存在的 key）。
func SeedFromConfig(store *Store, cfg *config.Config) {
	if store == nil || cfg == nil {
		return
	}
	seedBool(store, KeyArenaEnabled, cfg.Tasks.Arena.Enabled)
	seedFloat(store, KeyArenaAutoBuy, float64(cfg.Tasks.Arena.AutoBuyCount))
	seedFloat(store, KeyArenaTrophyDiff, float64(cfg.Tasks.Arena.TrophyDiff))
	if cfg.Tasks.Arena.MaxBattles != nil {
		seedFloat(store, KeyArenaMaxBattles, float64(*cfg.Tasks.Arena.MaxBattles))
	} else {
		seedFloat(store, KeyArenaMaxBattles, 0)
	}
	SeedPanelDefaults(store)
}

// ApplyToConfig 将 Store 中的 UI 选项写回 config。
func ApplyToConfig(store *Store, cfg *config.Config) {
	if store == nil || cfg == nil {
		return
	}
	cfg.Tasks.Arena.Enabled = store.GetBool(KeyArenaEnabled)
	cfg.Tasks.Arena.AutoBuyCount = int(store.GetFloat(KeyArenaAutoBuy))
	cfg.Tasks.Arena.TrophyDiff = int(store.GetFloat(KeyArenaTrophyDiff))

	maxBattles := int(store.GetFloat(KeyArenaMaxBattles))
	if maxBattles > 0 {
		cfg.Tasks.Arena.MaxBattles = &maxBattles
	} else {
		cfg.Tasks.Arena.MaxBattles = nil
	}
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
