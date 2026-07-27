package main

import (
	"fmt"

	"app/internal/config"
	"app/ui"
)

// mineTaskDescriptors 矿山分类下四个任务的面板描述符。任务 ID、Title、
// Category 与 EnabledKey 与占位版一致，设备上已有的 ui.json 无缝兼容。
func mineTaskDescriptors(cfg *config.Config) []ui.Task {
	return []ui.Task{
		oreVeinMiningTaskDescriptor(cfg),
		mineVentureTaskDescriptor(cfg),
		mineBattleTaskDescriptor(cfg),
		meltedAgarCubesTaskDescriptor(cfg),
	}
}

// oreVeinMiningTaskDescriptor 矿山开采：启用 + 调度间隔 + 矿卡优先级（逗号分隔）。
func oreVeinMiningTaskDescriptor(cfg *config.Config) ui.Task {
	m := &cfg.Tasks.OreVeinMining
	return ui.Task{
		ID:         "ore_vein_mining",
		Title:      "矿山开采",
		Category:   "矿山",
		EnabledKey: "ore_vein_mining_enabled",
		Fields: []ui.Field{
			ui.Bool("ore_vein_mining_enabled", "启用",
				func() bool { return m.Enabled },
				func(v bool) { m.Enabled = v }),
			ui.Number("ore_vein_mining_interval_sec", "全忙后再调度间隔(秒)", 60, 86400, 60,
				func() int { return m.IntervalSec },
				func(v int) { m.IntervalSec = v }),
			ui.Text("ore_vein_mining_ore_cards", "矿卡优先级(逗号分隔)",
				func() string { return joinList(m.OreCards) },
				func(v string) { m.OreCards = splitList(v) }),
		},
		Summary: func(s *ui.Store) string {
			return fmt.Sprintf("间隔 %ds · 矿卡 %d 张",
				int(s.GetFloat("ore_vein_mining_interval_sec")),
				len(splitList(s.GetString("ore_vein_mining_ore_cards"))))
		},
	}
}

// mineVentureTaskDescriptor 矿山勘查：启用 + 目标层数/近距阈值/轮询与远距等待。
func mineVentureTaskDescriptor(cfg *config.Config) ui.Task {
	m := &cfg.Tasks.MineVenture
	return ui.Task{
		ID:         "mine_venture",
		Title:      "矿山勘查",
		Category:   "矿山",
		EnabledKey: "mine_venture_enabled",
		Fields: []ui.Field{
			ui.Bool("mine_venture_enabled", "启用",
				func() bool { return m.Enabled },
				func(v bool) { m.Enabled = v }),
			ui.Number("mine_venture_target_floor", "目标层数", 1, 100, 1,
				func() int { return m.TargetFloor },
				func(v int) { m.TargetFloor = v }),
			ui.Number("mine_venture_far_gap", "近距阈值(层)", 0, 20, 1,
				func() int { return m.FarGap },
				func(v int) { m.FarGap = v }),
			ui.Number("mine_venture_ocr_poll_sec", "近距轮询间隔(秒)", 10, 600, 10,
				func() int { return m.OCRPollSec },
				func(v int) { m.OCRPollSec = v }),
			ui.Number("mine_venture_far_wait_sec", "远距回城等待(秒)", 60, 7200, 60,
				func() int { return m.FarWaitSec },
				func(v int) { m.FarWaitSec = v }),
		},
		Summary: func(s *ui.Store) string {
			return fmt.Sprintf("目标 %d 层 · 近距阈值 %d · 远距等待 %ds",
				int(s.GetFloat("mine_venture_target_floor")),
				int(s.GetFloat("mine_venture_far_gap")),
				int(s.GetFloat("mine_venture_far_wait_sec")))
		},
	}
}

// mineBattleTaskDescriptor 矿山战斗：启用 + 检测间隔 + 目标灵魂石（逗号分隔）。
func mineBattleTaskDescriptor(cfg *config.Config) ui.Task {
	m := &cfg.Tasks.MineBattle
	return ui.Task{
		ID:         "mine_battle",
		Title:      "矿山战斗",
		Category:   "矿山",
		EnabledKey: "mine_battle_enabled",
		Fields: []ui.Field{
			ui.Bool("mine_battle_enabled", "启用",
				func() bool { return m.Enabled },
				func(v bool) { m.Enabled = v }),
			ui.Number("mine_battle_interval_sec", "战斗检测间隔(秒)", 60, 86400, 60,
				func() int { return m.IntervalSec },
				func(v int) { m.IntervalSec = v }),
			ui.Text("mine_battle_soul_stones", "目标灵魂石(逗号分隔)",
				func() string { return joinList(m.SoulStones) },
				func(v string) { m.SoulStones = splitList(v) }),
		},
		Summary: func(s *ui.Store) string {
			return fmt.Sprintf("间隔 %ds · 灵魂石 %d 个",
				int(s.GetFloat("mine_battle_interval_sec")),
				len(splitList(s.GetString("mine_battle_soul_stones"))))
		},
	}
}

// meltedAgarCubesTaskDescriptor 解除洋菜冻：启用 + 冷却间隔。
func meltedAgarCubesTaskDescriptor(cfg *config.Config) ui.Task {
	m := &cfg.Tasks.MeltedAgarCubes
	return ui.Task{
		ID:         "melted_agar_cubes",
		Title:      "解除洋菜冻",
		Category:   "矿山",
		EnabledKey: "melted_agar_cubes_enabled",
		Fields: []ui.Field{
			ui.Bool("melted_agar_cubes_enabled", "启用",
				func() bool { return m.Enabled },
				func(v bool) { m.Enabled = v }),
			ui.Number("melted_agar_cubes_interval_sec", "冷却间隔(秒)", 60, 86400, 60,
				func() int { return m.IntervalSec },
				func(v int) { m.IntervalSec = v }),
		},
		Summary: func(s *ui.Store) string {
			return fmt.Sprintf("冷却 %ds", int(s.GetFloat("melted_agar_cubes_interval_sec")))
		},
	}
}
