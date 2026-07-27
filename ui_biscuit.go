package main

import (
	"fmt"

	"app/internal/config"
	"app/ui"
)

// biscuitTaskDescriptor 洗脆饼词条的面板描述符：启用 + 洗炼上限 + 4 条槽位
// 目标规则 + 1 条总和规则。minPercent/minSum 配置层是 float64，但 ui.Number
// 只支持 int，面板按整数百分比存取（Lua 默认 5/6/11 皆整数）；小数需求暂未支持。
func biscuitTaskDescriptor(cfg *config.Config) ui.Task {
	b := &cfg.Tasks.Biscuit
	fields := []ui.Field{
		ui.Bool("biscuit_enabled", "启用",
			func() bool { return b.Enabled },
			func(v bool) { b.Enabled = v }),
		ui.Number("biscuit_max_rolls", "洗炼上限次数", 1, 9999, 10,
			func() int { return b.MaxRolls },
			func(v int) { b.MaxRolls = v }),
	}
	for i := 0; i < 4; i++ {
		idx := i
		target := func() *config.BiscuitTarget {
			for len(b.Targets) <= idx {
				b.Targets = append(b.Targets, config.BiscuitTarget{})
			}
			return &b.Targets[idx]
		}
		prefix := fmt.Sprintf("biscuit_target%d_", idx+1)
		label := fmt.Sprintf("目标%d", idx+1)
		fields = append(fields,
			ui.Bool(prefix+"enabled", label+"启用",
				func() bool { return target().Enabled },
				func(v bool) { target().Enabled = v }),
			ui.Text(prefix+"name", label+"词条名",
				func() string { return target().Name },
				func(v string) { target().Name = v }),
			ui.Number(prefix+"min", label+"最低百分比", 0, 100, 1,
				func() int { return int(target().MinPercent) },
				func(v int) { target().MinPercent = float64(v) }),
		)
	}
	sumRule := func() *config.BiscuitSumRule {
		for len(b.SumRules) < 1 {
			b.SumRules = append(b.SumRules, config.BiscuitSumRule{})
		}
		return &b.SumRules[0]
	}
	fields = append(fields,
		ui.Bool("biscuit_sum1_enabled", "总和规则启用",
			func() bool { return sumRule().Enabled },
			func(v bool) { sumRule().Enabled = v }),
		ui.Text("biscuit_sum1_name", "总和词条名",
			func() string { return sumRule().Name },
			func(v string) { sumRule().Name = v }),
		ui.Number("biscuit_sum1_count", "取最高条数", 1, 4, 1,
			func() int { return sumRule().Count },
			func(v int) { sumRule().Count = v }),
		ui.Number("biscuit_sum1_min_sum", "总和阈值", 0, 100, 1,
			func() int { return int(sumRule().MinSum) },
			func(v int) { sumRule().MinSum = float64(v) }),
	)
	return ui.Task{
		ID:         "biscuit",
		Title:      "洗脆饼词条",
		Category:   "功能",
		EnabledKey: "biscuit_enabled",
		Fields:     fields,
		Summary: func(s *ui.Store) string {
			names := []string{}
			for i := 1; i <= 4; i++ {
				prefix := fmt.Sprintf("biscuit_target%d_", i)
				if s.GetBool(prefix + "enabled") {
					if name := s.GetString(prefix + "name"); name != "" {
						names = append(names, name)
					}
				}
			}
			return fmt.Sprintf("上限 %d 次 · 目标 %s",
				int(s.GetFloat("biscuit_max_rolls")), joinList(names))
		},
	}
}
