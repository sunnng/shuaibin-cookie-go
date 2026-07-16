package config

type Arena struct {
	Enabled      bool `json:"enabled"`
	MaxBattles   *int `json:"maxBattles"`
	AutoBuyCount int  `json:"autoBuyCount"`
	// TrophyDiff：选对手奖杯区间半宽。只打奖杯在 [myTrophy-TrophyDiff, myTrophy+TrophyDiff] 内的对手。
	// 注意：0 表示"严格相等"（只打奖杯完全等于自己的对手），不是"不过滤"。
	TrophyDiff int `json:"trophyDiff"`
}

type ModuleConfig struct {
	Arena Arena `json:"arena"`
}

type Config struct {
	Modules ModuleConfig `json:"modules"`
}

func DefaultConfig() *Config {
	return &Config{
		Modules: ModuleConfig{
			Arena: Arena{
				Enabled:      true,
				AutoBuyCount: 0,
				TrophyDiff:   0,
			},
		},
	}
}
