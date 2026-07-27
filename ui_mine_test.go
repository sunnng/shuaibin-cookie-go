package main

import (
	"testing"

	"app/internal/config"
	"app/ui"
)

func TestMineDescriptorsSeedApplyRoundTrip(t *testing.T) {
	cfg := config.DefaultConfig()
	tasks := mineTaskDescriptors(cfg)
	if len(tasks) != 4 {
		t.Fatalf("len=%d want 4", len(tasks))
	}

	s := ui.NewStore()
	ui.SeedAll(s, tasks)

	// 默认开关：勘查默认开（Lua surveyEnabled=true），其余默认关。
	if s.GetBool("ore_vein_mining_enabled") || s.GetBool("mine_battle_enabled") ||
		s.GetBool("melted_agar_cubes_enabled") {
		t.Fatal("mining/battle/jelly should seed disabled")
	}
	if !s.GetBool("mine_venture_enabled") {
		t.Fatal("mine venture should seed enabled")
	}
	if got := int(s.GetFloat("mine_venture_target_floor")); got != 6 {
		t.Fatalf("seed target_floor=%d", got)
	}
	if got := s.GetString("ore_vein_mining_ore_cards"); got == "" {
		t.Fatal("seed ore cards should not be empty")
	}

	// 面板改动后 Apply 回写 cfg。
	s.SetBool("ore_vein_mining_enabled", true)
	s.SetBool("mine_venture_enabled", false)
	s.SetBool("mine_battle_enabled", true)
	s.SetBool("melted_agar_cubes_enabled", true)
	s.SetFloat("mine_venture_target_floor", 9)
	s.SetFloat("mine_venture_far_gap", 3)
	s.SetFloat("mine_venture_ocr_poll_sec", 30)
	s.SetFloat("mine_venture_far_wait_sec", 300)
	s.SetFloat("ore_vein_mining_interval_sec", 600)
	s.SetString("ore_vein_mining_ore_cards", "sugarOre, flourStone")
	s.SetFloat("mine_battle_interval_sec", 7200)
	s.SetString("mine_battle_soul_stones", "妖精王，莓果")
	s.SetFloat("melted_agar_cubes_interval_sec", 1800)
	ui.ApplyAll(s, tasks)

	if !cfg.Tasks.OreVeinMining.Enabled || cfg.Tasks.MineVenture.Enabled ||
		!cfg.Tasks.MineBattle.Enabled || !cfg.Tasks.MeltedAgarCubes.Enabled {
		t.Fatalf("apply enabled: %+v", cfg.Tasks)
	}
	v := cfg.Tasks.MineVenture
	if v.TargetFloor != 9 || v.FarGap != 3 || v.OCRPollSec != 30 || v.FarWaitSec != 300 {
		t.Fatalf("apply venture: %+v", v)
	}
	m := cfg.Tasks.OreVeinMining
	if m.IntervalSec != 600 || len(m.OreCards) != 2 || m.OreCards[0] != "sugarOre" || m.OreCards[1] != "flourStone" {
		t.Fatalf("apply mining: %+v", m)
	}
	b := cfg.Tasks.MineBattle
	if b.IntervalSec != 7200 || len(b.SoulStones) != 2 || b.SoulStones[0] != "妖精王" || b.SoulStones[1] != "莓果" {
		t.Fatalf("apply battle: %+v", b)
	}
	if cfg.Tasks.MeltedAgarCubes.IntervalSec != 1800 {
		t.Fatalf("apply jelly: %+v", cfg.Tasks.MeltedAgarCubes)
	}
}

func TestMineDescriptorsSummaryAndCategory(t *testing.T) {
	cfg := config.DefaultConfig()
	tasks := mineTaskDescriptors(cfg)
	s := ui.NewStore()
	ui.SeedAll(s, tasks)
	wantIDs := []string{"ore_vein_mining", "mine_venture", "mine_battle", "melted_agar_cubes"}
	wantTitles := []string{"矿山开采", "矿山勘查", "矿山战斗", "解除洋菜冻"}
	for i, task := range tasks {
		if task.ID != wantIDs[i] || task.Title != wantTitles[i] {
			t.Fatalf("[%d] id/title=%s/%s", i, task.ID, task.Title)
		}
		if task.Category != "矿山" {
			t.Fatalf("[%d] category=%q", i, task.Category)
		}
		if got := task.Summary(s); got == "" || got == "未实现" {
			t.Fatalf("[%d] summary=%q", i, got)
		}
	}
	if got := tasks[1].Summary(s); got != "目标 6 层 · 近距阈值 2 · 远距等待 600s" {
		t.Fatalf("venture summary=%q", got)
	}
	if got := tasks[3].Summary(s); got != "冷却 3600s" {
		t.Fatalf("jelly summary=%q", got)
	}
	cats := ui.Categories(append([]ui.Task{arenaTaskDescriptor(cfg)}, tasks...))
	if len(cats) != 2 || cats[0] != "daily" || cats[1] != "矿山" {
		t.Fatalf("categories=%v", cats)
	}
}
