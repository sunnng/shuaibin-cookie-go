package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.TickIntervalMs != 800 {
		t.Fatalf("unexpected tick interval")
	}
	if !cfg.Modules.Arena.Enabled {
		t.Fatal("arena should be enabled by default for test")
	}
}

func TestLoadConfigMissingFile(t *testing.T) {
	cfg, err := LoadConfig(filepath.Join(t.TempDir(), "missing.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.TickIntervalMs != 800 {
		t.Fatalf("expected default config when file missing, got tickIntervalMs=%d", cfg.TickIntervalMs)
	}
}

func TestLoadConfigMergesUserValues(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	content := `{
		"tickIntervalMs": 1200,
		"modules": {
			"collectResources": false,
			"arena": {
				"enabled": false,
				"autoBuyCount": 3
			}
		}
	}`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.TickIntervalMs != 1200 {
		t.Fatalf("expected tickIntervalMs=1200, got %d", cfg.TickIntervalMs)
	}
	if cfg.MaxStateDurationSec != 45 {
		t.Fatalf("expected default maxStateDurationSec unchanged, got %d", cfg.MaxStateDurationSec)
	}
	if cfg.Modules.CollectResources {
		t.Fatal("expected collectResources=false")
	}
	if cfg.Modules.Arena.Enabled {
		t.Fatal("expected arena enabled=false")
	}
	if cfg.Modules.Arena.AutoBuyCount != 3 {
		t.Fatalf("expected autoBuyCount=3, got %d", cfg.Modules.Arena.AutoBuyCount)
	}
	if cfg.Modules.Arena.MaxBattles != nil {
		t.Fatalf("expected maxBattles nil by default, got %v", *cfg.Modules.Arena.MaxBattles)
	}
}
