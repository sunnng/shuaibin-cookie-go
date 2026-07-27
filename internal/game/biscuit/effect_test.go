package biscuit

import (
	"testing"
)

func TestExtractNumber(t *testing.T) {
	cases := []struct {
		in        string
		wantValue float64
		wantName  string
		wantOK    bool
	}{
		{"冷却时间5.2", 5.2, "冷却时间", true},
		{"暗黑属性伤害提升10.8", 10.8, "暗黑属性伤害提升", true},
		{"生命值3", 3, "生命值", true},
		{"会心 3.7", 3.7, "会心", true},
		{"  攻击力 11.5  ", 0, "  攻击力 11.5  ", false}, // 尾部空格 → 无尾部数字（对齐 Lua）
		{"防御力", 0, "防御力", false},
		{"", 0, "", false},
		{"12", 12, "", true}, // 纯数字：value 有效但 name 为空
		{"abc.", 0, "abc", false},
	}
	for _, c := range cases {
		v, name, ok := extractNumber(c.in)
		if ok != c.wantOK || name != c.wantName || (ok && v != c.wantValue) {
			t.Errorf("extractNumber(%q) = (%v,%q,%v), want (%v,%q,%v)",
				c.in, v, name, ok, c.wantValue, c.wantName, c.wantOK)
		}
	}
}

func TestParseRaw(t *testing.T) {
	// Lua 注释里的真实样例
	raw := "暗黑属性伤害提升10.8%生命值3%生命值7.9%会心3.7%"
	effects := parseRaw(raw)
	if len(effects) != 4 {
		t.Fatalf("len = %d, want 4: %+v", len(effects), effects)
	}
	want := []Effect{
		{Name: "暗黑属性伤害提升", Value: 10.8, Raw: "暗黑属性伤害提升10.8%"},
		{Name: "生命值", Value: 3, Raw: "生命值3%"},
		{Name: "生命值", Value: 7.9, Raw: "生命值7.9%"},
		{Name: "会心", Value: 3.7, Raw: "会心3.7%"},
	}
	for i, e := range effects {
		if e != want[i] {
			t.Errorf("effects[%d] = %+v, want %+v", i, e, want[i])
		}
	}
}

func TestParseRawSkipsUnparsable(t *testing.T) {
	cases := []struct {
		in   string
		want int
	}{
		{"", 0},
		{"abc%", 0},         // 无数字
		{"12%", 0},          // 无名称
		{"攻击力11.5%防御力%", 1}, // 第二段无数字被丢弃
		{"%%", 0},
	}
	for _, c := range cases {
		if got := len(parseRaw(c.in)); got != c.want {
			t.Errorf("parseRaw(%q) len = %d, want %d", c.in, got, c.want)
		}
	}
}
