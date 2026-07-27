package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if !cfg.Tasks.Arena.Enabled {
		t.Fatal("arena should be enabled by default")
	}
	if cfg.Tasks.Arena.AutoBuyCount != 0 {
		t.Fatalf("expected autoBuyCount=0, got %d", cfg.Tasks.Arena.AutoBuyCount)
	}
	if cfg.Tasks.OreVeinMining.Enabled || cfg.Tasks.MineBattle.Enabled ||
		cfg.Tasks.MeltedAgarCubes.Enabled {
		t.Fatal("mining/battle/jelly should be disabled by default")
	}
	if !cfg.Tasks.MineVenture.Enabled {
		t.Fatal("mine venture should be enabled by default (Lua surveyEnabled=true)")
	}
	if cfg.Tasks.MineVenture.TargetFloor != 6 || cfg.Tasks.MineVenture.FarGap != 2 ||
		cfg.Tasks.MineVenture.OCRPollSec != 60 || cfg.Tasks.MineVenture.FarWaitSec != 600 {
		t.Fatalf("mine venture defaults: %+v", cfg.Tasks.MineVenture)
	}
	if cfg.Tasks.OreVeinMining.IntervalSec != 1200 || len(cfg.Tasks.OreVeinMining.OreCards) != 6 {
		t.Fatalf("ore vein mining defaults: %+v", cfg.Tasks.OreVeinMining)
	}
	if cfg.Tasks.MineBattle.IntervalSec != 21600 || len(cfg.Tasks.MineBattle.SoulStones) != 3 {
		t.Fatalf("mine battle defaults: %+v", cfg.Tasks.MineBattle)
	}
	if cfg.Tasks.MeltedAgarCubes.IntervalSec != 3600 {
		t.Fatalf("jelly defaults: %+v", cfg.Tasks.MeltedAgarCubes)
	}
	if !cfg.Tasks.Square.Enabled || cfg.Tasks.Square.DailyCap != 240 ||
		cfg.Tasks.Square.CheckIntervalSec != 60 || cfg.Tasks.Square.ChunkSec != 10 {
		t.Fatalf("square defaults: %+v", cfg.Tasks.Square)
	}
	if cfg.Tasks.Starlight.Enabled {
		t.Fatal("starlight should be disabled by default")
	}
	if cfg.Tasks.SeasideMarket.Enabled || len(cfg.Tasks.SeasideMarket.Items) != 7 ||
		cfg.Tasks.SeasideMarket.RestockBufferSec != 30 {
		t.Fatalf("market defaults: %+v", cfg.Tasks.SeasideMarket)
	}
	b := cfg.Tasks.Biscuit
	if b.Enabled || b.MaxRolls != 500 || len(b.Targets) != 4 || len(b.SumRules) != 1 {
		t.Fatalf("biscuit defaults: %+v", b)
	}
	if !b.Targets[0].Enabled || b.Targets[0].Name != "冷却时间" || b.Targets[0].MinPercent != 5 {
		t.Fatalf("biscuit target1: %+v", b.Targets[0])
	}
	if !b.Targets[1].Enabled || b.Targets[1].Name != "会心" || b.Targets[1].MinPercent != 6 {
		t.Fatalf("biscuit target2: %+v", b.Targets[1])
	}
	if b.Targets[2].Enabled || b.Targets[3].Enabled {
		t.Fatal("biscuit target3/4 should be disabled by default")
	}
	if !b.SumRules[0].Enabled || b.SumRules[0].Name != "攻击力" ||
		b.SumRules[0].Count != 2 || b.SumRules[0].MinSum != 11 {
		t.Fatalf("biscuit sum rule: %+v", b.SumRules[0])
	}
}

func TestLoadConfigMissingFile(t *testing.T) {
	cfg, err := LoadConfig(filepath.Join(t.TempDir(), "missing.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.Tasks.Arena.Enabled {
		t.Fatal("expected default arena enabled when file missing")
	}
}

func TestLoadConfigMergesUserValues(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	content := `{
		"tasks": {
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
	if cfg.Tasks.Arena.Enabled {
		t.Fatal("expected arena enabled=false")
	}
	if cfg.Tasks.Arena.AutoBuyCount != 3 {
		t.Fatalf("expected autoBuyCount=3, got %d", cfg.Tasks.Arena.AutoBuyCount)
	}
	if cfg.Tasks.Arena.MaxBattles != nil {
		t.Fatalf("expected maxBattles nil by default, got %v", *cfg.Tasks.Arena.MaxBattles)
	}
}
