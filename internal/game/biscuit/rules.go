package biscuit

import (
	"fmt"
	"sort"
)

// checkSlots 槽位规则：每条启用规则需在 4 条实际词条中独占一条满足
// name 相同且 value >= minPercent 的词条。规则按阈值降序贪心分配
// （最难满足的优先拿词条）。移植自 Lua checkSlots。
func checkSlots(effects []Effect, targets []TargetRule) (bool, string) {
	type rule struct {
		name string
		min  float64
	}
	var active []rule
	for _, r := range targets {
		if r.Enabled && r.Name != "" {
			active = append(active, rule{name: r.Name, min: r.MinPercent})
		}
	}
	if len(active) == 0 {
		return false, "无槽位规则"
	}

	type slot struct {
		name  string
		value float64
		used  bool
	}
	pool := make([]slot, len(effects))
	for i, e := range effects {
		pool[i] = slot{name: e.Name, value: e.Value}
	}

	sort.SliceStable(active, func(i, j int) bool { return active[i].min > active[j].min })

	for _, r := range active {
		found := false
		for i := range pool {
			if !pool[i].used && pool[i].name == r.name && pool[i].value >= r.min {
				pool[i].used = true
				found = true
				break
			}
		}
		if !found {
			return false, fmt.Sprintf("缺[%s>=%s]", r.name, formatNumber(r.min))
		}
	}
	return true, "毕业"
}

// checkSums 总和规则：同名词条数量 >= count 时取最高 count 条求和，
// 加和 >= minSum 即满足。任一启用规则满足即毕业。移植自 Lua checkSums。
func checkSums(effects []Effect, sumRules []SumRule) (bool, string) {
	if len(sumRules) == 0 {
		return false, "未配置总和规则"
	}
	for _, r := range sumRules {
		if !r.Enabled || r.Name == "" || r.Count <= 0 {
			continue
		}
		need := r.Count
		if need > effectSlotCount {
			need = effectSlotCount
		}
		var values []float64
		for _, e := range effects {
			if e.Name == r.Name {
				values = append(values, e.Value)
			}
		}
		if len(values) < need {
			continue
		}
		sort.Sort(sort.Reverse(sort.Float64Slice(values)))
		sum := 0.0
		for i := 0; i < need; i++ {
			sum += values[i]
		}
		if sum >= r.MinSum {
			return true, fmt.Sprintf("[%s]取高%d条 总和%.1f≥%.1f", r.Name, need, sum, r.MinSum)
		}
	}
	return false, "总和规则未满足"
}

// check 整合检查：槽位规则或总和规则满足其一即毕业。移植自 Lua check。
func check(effects []Effect, targets []TargetRule, sumRules []SumRule) (bool, string) {
	ok, msg := checkSlots(effects, targets)
	if ok {
		return true, msg
	}
	if len(sumRules) > 0 {
		if ok2, msg2 := checkSums(effects, sumRules); ok2 {
			return true, msg2
		}
	}
	return false, msg
}
