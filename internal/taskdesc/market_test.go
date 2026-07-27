package taskdesc

import (
	"testing"

	"app/internal/config"
	"app/ui"
)

func TestMarketDescriptorSeedApplyRoundTrip(t *testing.T) {
	cfg := config.DefaultConfig()
	task := Market(cfg)

	s := ui.NewStore()
	ui.SeedAll(s, []ui.Task{task})
	if s.GetBool("market_enabled") {
		t.Fatal("market should seed disabled")
	}
	if got := s.GetString("market_items"); got == "" {
		t.Fatal("seed items should not be empty")
	}
	if int(s.GetFloat("market_restock_buffer_sec")) != 30 {
		t.Fatalf("seed buffer=%v", s.GetFloat("market_restock_buffer_sec"))
	}

	s.SetBool("market_enabled", true)
	s.SetString("market_items", "灿烂的光之碎片, 商品3_罗盘")
	s.SetFloat("market_restock_buffer_sec", 60)
	ui.ApplyAll(s, []ui.Task{task})

	m := cfg.Tasks.SeasideMarket
	if !m.Enabled || m.RestockBufferSec != 60 {
		t.Fatalf("apply: %+v", m)
	}
	if len(m.Items) != 2 || m.Items[0] != "灿烂的光之碎片" || m.Items[1] != "商品3_罗盘" {
		t.Fatalf("apply items: %v", m.Items)
	}
}

func TestMarketDescriptorSummaryAndCategory(t *testing.T) {
	cfg := config.DefaultConfig()
	task := Market(cfg)
	if task.ID != "seaside_market" || task.Title != "海滩交易所" {
		t.Fatalf("id/title=%s/%s", task.ID, task.Title)
	}
	if task.Category != "daily" {
		t.Fatalf("category=%q", task.Category)
	}
	s := ui.NewStore()
	ui.SeedAll(s, []ui.Task{task})
	if got := task.Summary(s); got != "清单 7 项 · 缓冲 30s" {
		t.Fatalf("summary=%q", got)
	}
}
