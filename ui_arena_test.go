package main

import (
	"testing"

	"app/internal/config"
	"app/ui"
)

func TestArenaDescriptorSeedApplyRoundTrip(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Tasks.Arena.Enabled = true
	cfg.Tasks.Arena.AutoBuyCount = 2
	cfg.Tasks.Arena.TrophyDiff = 50
	maxB := 8
	cfg.Tasks.Arena.MaxBattles = &maxB

	task := arenaTaskDescriptor(cfg)
	tasks := []ui.Task{task}

	s := ui.NewStore()
	ui.SeedAll(s, tasks)
	if !s.GetBool("arena_enabled") || s.GetFloat("arena_max_battles") != 8 ||
		s.GetFloat("arena_auto_buy_count") != 2 || s.GetFloat("arena_trophy_diff") != 50 {
		t.Fatalf("seed: %v %v %v %v", s.GetBool("arena_enabled"),
			s.GetFloat("arena_max_battles"), s.GetFloat("arena_auto_buy_count"),
			s.GetFloat("arena_trophy_diff"))
	}

	// 面板改值 → Apply 回写 cfg；上限 0 → MaxBattles 归 nil（不限）
	s.SetBool("arena_enabled", false)
	s.SetFloat("arena_max_battles", 0)
	s.SetFloat("arena_auto_buy_count", 3)
	ui.ApplyAll(s, tasks)
	a := cfg.Tasks.Arena
	if a.Enabled || a.MaxBattles != nil || a.AutoBuyCount != 3 || a.TrophyDiff != 50 {
		t.Fatalf("apply: %+v", a)
	}
}

func TestArenaDescriptorSummary(t *testing.T) {
	cfg := config.DefaultConfig()
	task := arenaTaskDescriptor(cfg)
	s := ui.NewStore()
	s.SetFloat("arena_max_battles", 10)
	s.SetFloat("arena_auto_buy_count", 1)
	s.SetFloat("arena_trophy_diff", 30)
	if got := task.Summary(s); got != "上限 10 · 购买 1 · 奖杯差 30" {
		t.Fatalf("summary=%q", got)
	}
	s.SetFloat("arena_max_battles", 0)
	if got := task.Summary(s); got != "上限 不限 · 购买 1 · 奖杯差 30" {
		t.Fatalf("summary unlimited=%q", got)
	}
}
