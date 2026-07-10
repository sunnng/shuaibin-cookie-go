package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if !cfg.Modules.Arena.Enabled {
		t.Fatal("arena should be enabled by default")
	}
	if cfg.Modules.Arena.AutoBuyCount != 0 {
		t.Fatalf("expected autoBuyCount=0, got %d", cfg.Modules.Arena.AutoBuyCount)
	}
}

func TestLoadConfigMissingFile(t *testing.T) {
	cfg, err := LoadConfig(filepath.Join(t.TempDir(), "missing.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.Modules.Arena.Enabled {
		t.Fatal("expected default arena enabled when file missing")
	}
}

func TestLoadConfigMergesUserValues(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	content := `{
		"modules": {
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
