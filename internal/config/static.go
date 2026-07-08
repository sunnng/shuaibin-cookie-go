package config

type Arena struct {
	Enabled      bool `json:"enabled"`
	MaxBattles   *int `json:"maxBattles"`
	AutoBuyCount int  `json:"autoBuyCount"`
	TrophyDiff   int  `json:"trophyDiff"`
}

type ModuleConfig struct {
	CollectResources bool  `json:"collectResources"`
	FarmLevels       bool  `json:"farmLevels"`
	Arena            Arena `json:"arena"`
}

type Config struct {
	TickIntervalMs      int          `json:"tickIntervalMs"`
	MaxStateDurationSec int          `json:"maxStateDurationSec"`
	MaxUnknownRetries   int          `json:"maxUnknownRetries"`
	MaxRecoveryAttempts int          `json:"maxRecoveryAttempts"`
	LowPowerWaitSec     int          `json:"lowPowerWaitSec"`
	Modules             ModuleConfig `json:"modules"`
}

func DefaultConfig() *Config {
	return &Config{
		TickIntervalMs:      800,
		MaxStateDurationSec: 45,
		MaxUnknownRetries:   5,
		MaxRecoveryAttempts: 3,
		LowPowerWaitSec:     30,
		Modules: ModuleConfig{
			CollectResources: true,
			FarmLevels:       false,
			Arena: Arena{
				Enabled:      true,
				AutoBuyCount: 0,
				TrophyDiff:   0,
			},
		},
	}
}
