package main

import (
	"testing"

	"app/internal/config"
	"app/ui"
)

func TestBiscuitDescriptorSeedApplyRoundTrip(t *testing.T) {
	cfg := config.DefaultConfig()
	task := biscuitTaskDescriptor(cfg)

	s := ui.NewStore()
	ui.SeedAll(s, []ui.Task{task})
	if s.GetBool("biscuit_enabled") {
		t.Fatal("biscuit should seed disabled")
	}
	if int(s.GetFloat("biscuit_max_rolls")) != 500 {
		t.Fatalf("seed max_rolls=%v", s.GetFloat("biscuit_max_rolls"))
	}
	if got := s.GetString("biscuit_target1_name"); got != "冷却时间" {
		t.Fatalf("seed target1 name=%q", got)
	}
	if !s.GetBool("biscuit_sum1_enabled") || s.GetString("biscuit_sum1_name") != "攻击力" {
		t.Fatal("seed sum rule mismatch")
	}

	s.SetBool("biscuit_enabled", true)
	s.SetFloat("biscuit_max_rolls", 200)
	s.SetBool("biscuit_target3_enabled", true)
	s.SetString("biscuit_target3_name", "伤害抵抗")
	s.SetFloat("biscuit_target3_min", 4)
	s.SetFloat("biscuit_sum1_count", 3)
	s.SetFloat("biscuit_sum1_min_sum", 15)
	ui.ApplyAll(s, []ui.Task{task})

	b := cfg.Tasks.Biscuit
	if !b.Enabled || b.MaxRolls != 200 {
		t.Fatalf("apply: %+v", b)
	}
	if len(b.Targets) != 4 {
		t.Fatalf("targets len=%d", len(b.Targets))
	}
	if !b.Targets[2].Enabled || b.Targets[2].Name != "伤害抵抗" || b.Targets[2].MinPercent != 4 {
		t.Fatalf("target3: %+v", b.Targets[2])
	}
	if b.SumRules[0].Count != 3 || b.SumRules[0].MinSum != 15 {
		t.Fatalf("sum rule: %+v", b.SumRules[0])
	}
	// 未改动的默认槽位保持不变。
	if !b.Targets[0].Enabled || b.Targets[0].Name != "冷却时间" || b.Targets[0].MinPercent != 5 {
		t.Fatalf("target1: %+v", b.Targets[0])
	}
}

func TestBiscuitDescriptorSummaryAndCategory(t *testing.T) {
	cfg := config.DefaultConfig()
	task := biscuitTaskDescriptor(cfg)
	if task.ID != "biscuit" || task.Title != "洗脆饼词条" {
		t.Fatalf("id/title=%s/%s", task.ID, task.Title)
	}
	if task.Category != "功能" {
		t.Fatalf("category=%q", task.Category)
	}
	s := ui.NewStore()
	ui.SeedAll(s, []ui.Task{task})
	if got := task.Summary(s); got != "上限 500 次 · 目标 冷却时间,会心" {
		t.Fatalf("summary=%q", got)
	}
}
