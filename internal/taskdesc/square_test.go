package taskdesc

import (
	"testing"

	"app/internal/config"
	"app/ui"
)

func TestSquareDescriptorSeedApplyRoundTrip(t *testing.T) {
	cfg := config.DefaultConfig()
	task := Square(cfg)

	s := ui.NewStore()
	ui.SeedAll(s, []ui.Task{task})
	if !s.GetBool("square_enabled") {
		t.Fatal("square should seed enabled")
	}
	if int(s.GetFloat("square_daily_cap")) != 240 {
		t.Fatalf("seed daily_cap=%v", s.GetFloat("square_daily_cap"))
	}

	s.SetBool("square_enabled", false)
	s.SetFloat("square_daily_cap", 120)
	s.SetFloat("square_check_interval_sec", 90)
	s.SetFloat("square_chunk_sec", 5)
	ui.ApplyAll(s, []ui.Task{task})

	q := cfg.Tasks.Square
	if q.Enabled || q.DailyCap != 120 || q.CheckIntervalSec != 90 || q.ChunkSec != 5 {
		t.Fatalf("apply: %+v", q)
	}
}

func TestSquareDescriptorSummaryAndCategory(t *testing.T) {
	cfg := config.DefaultConfig()
	task := Square(cfg)
	if task.ID != "square" || task.Title != "布谷鸟广场" {
		t.Fatalf("id/title=%s/%s", task.ID, task.Title)
	}
	if task.Category != "daily" {
		t.Fatalf("category=%q", task.Category)
	}
	s := ui.NewStore()
	ui.SeedAll(s, []ui.Task{task})
	if got := task.Summary(s); got != "日上限 240 · 停留 60s · 粒度 10s" {
		t.Fatalf("summary=%q", got)
	}
}
