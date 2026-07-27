package main

import (
	"testing"

	"app/internal/config"
	"app/ui"
)

func TestStarlightDescriptorSeedApplyRoundTrip(t *testing.T) {
	cfg := config.DefaultConfig()
	task := starlightTaskDescriptor(cfg)

	s := ui.NewStore()
	ui.SeedAll(s, []ui.Task{task})
	if s.GetBool("starlight_enabled") {
		t.Fatal("starlight should seed disabled")
	}

	s.SetBool("starlight_enabled", true)
	ui.ApplyAll(s, []ui.Task{task})
	if !cfg.Tasks.Starlight.Enabled {
		t.Fatal("apply: starlight should be enabled")
	}
}

func TestStarlightDescriptorSummaryAndCategory(t *testing.T) {
	cfg := config.DefaultConfig()
	task := starlightTaskDescriptor(cfg)
	if task.ID != "starlight" || task.Title != "梦幻繁星岛" {
		t.Fatalf("id/title=%s/%s", task.ID, task.Title)
	}
	if task.Category != "daily" {
		t.Fatalf("category=%q", task.Category)
	}
	s := ui.NewStore()
	ui.SeedAll(s, []ui.Task{task})
	if got := task.Summary(s); got != "未启用" {
		t.Fatalf("summary=%q", got)
	}
	s.SetBool("starlight_enabled", true)
	if got := task.Summary(s); got != "已启用" {
		t.Fatalf("summary=%q", got)
	}
}
