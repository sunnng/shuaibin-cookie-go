package biscuit

import "strconv"

// Entry 脆饼副词条参考表条目（名称及典型数值区间），
// 原样搬自 Lua game/功能_洗脆饼/词条库.lua。
type Entry struct {
	Name     string
	MinValue float64
	MaxValue float64
}

// entries 游戏内可出现的副词条名称及典型数值区间。
var entries = []Entry{
	{Name: "攻击力", MinValue: 3, MaxValue: 7.5},
	{Name: "防御力", MinValue: 5, MaxValue: 7.5},
	{Name: "生命值", MinValue: 3, MaxValue: 15},
	{Name: "攻击速度", MinValue: 3, MaxValue: 10},
	{Name: "会心", MinValue: 3, MaxValue: 7},
	{Name: "冷却时间", MinValue: 2, MaxValue: 6},
	{Name: "伤害减免", MinValue: 5, MaxValue: 10},
	{Name: "会心伤害减免", MinValue: 4, MaxValue: 10},
	{Name: "增益效果增强", MinValue: 2, MaxValue: 5},
	{Name: "减益效果减免", MinValue: 2, MaxValue: 5},
	{Name: "无视伤害减免", MinValue: 5, MaxValue: 15},
	{Name: "电属性伤害提升", MinValue: 8, MaxValue: 15},
	{Name: "火属性伤害提升", MinValue: 8, MaxValue: 15},
	{Name: "暗属性伤害提升", MinValue: 8, MaxValue: 15},
	{Name: "毒属性伤害提升", MinValue: 8, MaxValue: 15},
}

// EntryNames 返回全部副词条名称（顺序与词条库一致）。
func EntryNames() []string {
	list := make([]string, 0, len(entries))
	for _, e := range entries {
		list = append(list, e.Name)
	}
	return list
}

// FindEntry 按名称查词条，未命中返回 nil。
func FindEntry(name string) *Entry {
	for i := range entries {
		if entries[i].Name == name {
			return &entries[i]
		}
	}
	return nil
}

// ValueBounds 返回词条的单条数值区间；未知名称 ok=false。
func ValueBounds(name string) (min, max float64, ok bool) {
	e := FindEntry(name)
	if e == nil {
		return 0, 0, false
	}
	return e.MinValue, e.MaxValue, true
}

// SumBounds 返回同名词条取最高 count 条时的加和区间；count 截断到 [1,4]。
// 未知名称或 count<1 时 ok=false。
func SumBounds(name string, count int) (minSum, maxSum float64, ok bool) {
	min, max, ok := ValueBounds(name)
	if !ok || count < 1 {
		return 0, 0, false
	}
	if count > effectSlotCount {
		count = effectSlotCount
	}
	return min * float64(count), max * float64(count), true
}

// RangeHint 返回词条数值区间提示，如 "范围 3%~7.5%"；未知名称返回空串。
func RangeHint(name string) string {
	e := FindEntry(name)
	if e == nil {
		return ""
	}
	return "范围 " + formatNumber(e.MinValue) + "%~" + formatNumber(e.MaxValue) + "%"
}

// formatNumber 对齐 Lua %g 的显示风格（整数不带小数点）。
func formatNumber(v float64) string {
	return strconv.FormatFloat(v, 'g', -1, 64)
}
