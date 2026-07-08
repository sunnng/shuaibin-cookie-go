package bot

import (
	"encoding/json"
	"os"
)

type ModuleConfig struct {
	CollectResources bool `json:"collectResources"`
	FarmLevels       bool `json:"farmLevels"`
	Arena            bool `json:"arena"`
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
			Arena:            false,
		},
	}
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return DefaultConfig(), nil
	}
	cfg := DefaultConfig()
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
