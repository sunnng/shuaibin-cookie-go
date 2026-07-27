package biscuit

import "testing"

func TestEntryNames(t *testing.T) {
	names := EntryNames()
	if len(names) != 15 {
		t.Fatalf("len(EntryNames) = %d, want 15", len(names))
	}
	if names[0] != "攻击力" || names[5] != "冷却时间" {
		t.Fatalf("unexpected order: %v", names[:6])
	}
}

func TestFindEntryAndBounds(t *testing.T) {
	if FindEntry("会心") == nil {
		t.Fatal("会心 should exist")
	}
	if FindEntry("不存在的词条") != nil {
		t.Fatal("unknown name should return nil")
	}
	min, max, ok := ValueBounds("攻击力")
	if !ok || min != 3 || max != 7.5 {
		t.Fatalf("ValueBounds(攻击力) = (%v,%v,%v)", min, max, ok)
	}
	if _, _, ok := ValueBounds("不存在"); ok {
		t.Fatal("unknown name should not be ok")
	}
}

func TestSumBounds(t *testing.T) {
	min, max, ok := SumBounds("攻击力", 2)
	if !ok || min != 6 || max != 15 {
		t.Fatalf("SumBounds(攻击力,2) = (%v,%v,%v)", min, max, ok)
	}
	// count 截断到 4
	_, max, ok = SumBounds("生命值", 99)
	if !ok || max != 60 {
		t.Fatalf("SumBounds(生命值,99) max = %v, want 60", max)
	}
	if _, _, ok := SumBounds("攻击力", 0); ok {
		t.Fatal("count<1 should not be ok")
	}
}

func TestRangeHint(t *testing.T) {
	if got := RangeHint("攻击力"); got != "范围 3%~7.5%" {
		t.Fatalf("RangeHint(攻击力) = %q", got)
	}
	if got := RangeHint("防御力"); got != "范围 5%~7.5%" {
		t.Fatalf("RangeHint(防御力) = %q", got)
	}
	if got := RangeHint(""); got != "" {
		t.Fatalf("RangeHint(\"\") = %q", got)
	}
}
