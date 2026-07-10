package config

type Arena struct {
	Enabled      bool `json:"enabled"`
	MaxBattles   *int `json:"maxBattles"`
	AutoBuyCount int  `json:"autoBuyCount"`
	TrophyDiff   int  `json:"trophyDiff"`
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
