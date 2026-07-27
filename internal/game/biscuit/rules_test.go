package biscuit

import (
	"strings"
	"testing"
)

func TestCheckSlotsGraduate(t *testing.T) {
	cfg := DefaultConfig()
	effects := []Effect{
		{Name: "冷却时间", Value: 5.2},
		{Name: "生命值", Value: 3},
		{Name: "会心", Value: 6.8},
		{Name: "攻击力", Value: 4},
	}
	ok, msg := checkSlots(effects, cfg.Targets)
	if !ok || msg != "毕业" {
		t.Fatalf("checkSlots = (%v,%q), want (true,毕业)", ok, msg)
	}
}

func TestCheckSlotsMissing(t *testing.T) {
	cfg := DefaultConfig()
	effects := []Effect{
		{Name: "冷却时间", Value: 5.2},
		{Name: "生命值", Value: 3},
		{Name: "会心", Value: 3.7}, // 会心 < 6 不达标
		{Name: "攻击力", Value: 4},
	}
	ok, msg := checkSlots(effects, cfg.Targets)
	if ok || !strings.HasPrefix(msg, "缺[会心>=") {
		t.Fatalf("checkSlots = (%v,%q), want 缺会心", ok, msg)
	}
}

// 同名两条词条、两条同名规则：规则按阈值降序分配才能都满足（高阈值先拿高值）。
func TestCheckSlotsGreedyByMinDesc(t *testing.T) {
	targets := []TargetRule{
		{Enabled: true, Name: "生命值", MinPercent: 3},
		{Enabled: true, Name: "生命值", MinPercent: 7},
	}
	effects := []Effect{
		{Name: "生命值", Value: 3},
		{Name: "生命值", Value: 7.9},
	}
	ok, msg := checkSlots(effects, targets)
	if !ok {
		t.Fatalf("checkSlots = (%v,%q), want graduate", ok, msg)
	}

	// 只有一条达标时高阈值规则必须失败
	effects[1].Value = 6
	ok, msg = checkSlots(effects, targets)
	if ok || !strings.HasPrefix(msg, "缺[生命值>=7]") {
		t.Fatalf("checkSlots = (%v,%q), want 缺生命值>=7", ok, msg)
	}
}

func TestCheckSlotsNoActiveRule(t *testing.T) {
	ok, msg := checkSlots([]Effect{{Name: "攻击力", Value: 9}}, []TargetRule{{Enabled: false, Name: "攻击力", MinPercent: 1}})
	if ok || msg != "无槽位规则" {
		t.Fatalf("checkSlots = (%v,%q), want 无槽位规则", ok, msg)
	}
}

func TestCheckSums(t *testing.T) {
	rules := []SumRule{{Enabled: true, Name: "攻击力", Count: 2, MinSum: 11}}
	effects := []Effect{
		{Name: "攻击力", Value: 5.9},
		{Name: "生命值", Value: 3},
		{Name: "攻击力", Value: 5.2},
		{Name: "会心", Value: 3.7},
	}
	ok, msg := checkSums(effects, rules)
	if !ok || !strings.Contains(msg, "总和11.1≥11.0") {
		t.Fatalf("checkSums = (%v,%q), want 总和11.1≥11.0", ok, msg)
	}

	// 加和不够
	rules[0].MinSum = 12
	if ok, _ = checkSums(effects, rules); ok {
		t.Fatal("checkSums should fail when sum < minSum")
	}

	// 同名词条数量不足 count
	rules[0].MinSum = 1
	rules[0].Count = 3
	if ok, _ = checkSums(effects, rules); ok {
		t.Fatal("checkSums should fail when not enough same-name effects")
	}

	// 规则未启用 / 无规则
	if ok, msg := checkSums(effects, nil); ok || msg != "未配置总和规则" {
		t.Fatalf("checkSums(nil) = (%v,%q)", ok, msg)
	}
	if ok, _ = checkSums(effects, []SumRule{{Enabled: false, Name: "攻击力", Count: 2, MinSum: 1}}); ok {
		t.Fatal("disabled sum rule must not graduate")
	}
}

// check：槽位不满足但总和规则满足也算毕业；都不满足时返回槽位的失败消息。
func TestCheckCombined(t *testing.T) {
	cfg := DefaultConfig()
	effects := []Effect{
		{Name: "攻击力", Value: 5.9},
		{Name: "生命值", Value: 3},
		{Name: "攻击力", Value: 5.2},
		{Name: "会心", Value: 3.7},
	}
	ok, msg := check(effects, cfg.Targets, cfg.SumRules)
	if !ok || !strings.Contains(msg, "攻击力") {
		t.Fatalf("check = (%v,%q), want sum-rule graduate", ok, msg)
	}

	ok, msg = check(effects, cfg.Targets, nil)
	if ok || !strings.HasPrefix(msg, "缺[") {
		t.Fatalf("check without sumRules = (%v,%q), want slot failure msg", ok, msg)
	}
}
