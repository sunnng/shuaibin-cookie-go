package ui

import "testing"

func TestBoolFieldSeedAndApply(t *testing.T) {
	backing := true
	f := Bool("arena_enabled", "启用", func() bool { return backing }, func(v bool) { backing = v })

	s := NewStore()
	f.Seed(s)
	if !s.GetBool("arena_enabled") {
		t.Fatal("seed should write default when key missing")
	}

	s.SetBool("arena_enabled", false)
	f.Seed(s)
	if s.GetBool("arena_enabled") {
		t.Fatal("seed must not overwrite existing key")
	}

	f.Apply(s)
	if backing != false {
		t.Fatalf("apply should write store value back, backing=%v", backing)
	}
	if f.Widget() != WidgetCheckbox {
		t.Fatalf("widget=%v want checkbox", f.Widget())
	}
}

func TestNumberFieldStoresFloatAndAppliesInt(t *testing.T) {
	backing := 3
	f := Number("arena_max_battles", "战斗上限", 0, 99, 1,
		func() int { return backing }, func(v int) { backing = v })

	s := NewStore()
	f.Seed(s)
	if got := s.GetFloat("arena_max_battles"); got != 3 {
		t.Fatalf("seed stored %v want 3", got)
	}

	s.SetFloat("arena_max_battles", 7.9)
	f.Apply(s)
	if backing != 7 {
		t.Fatalf("apply truncated float to int, backing=%d want 7", backing)
	}

	c := f.Constraints()
	if c.Min != 0 || c.Max != 99 || c.Step != 1 {
		t.Fatalf("constraints=%+v", c)
	}
	if f.Widget() != WidgetNumberInput {
		t.Fatalf("widget=%v want number", f.Widget())
	}
}

func TestTextFieldAndNilStoreSafety(t *testing.T) {
	backing := "x"
	f := Text("note", "备注", func() string { return backing }, func(v string) { backing = v })
	s := NewStore()
	f.Seed(s)
	if s.GetString("note") != "x" {
		t.Fatal("seed string")
	}
	s.SetString("note", "y")
	f.Apply(s)
	if backing != "y" {
		t.Fatalf("apply string, backing=%q", backing)
	}
	if f.Widget() != WidgetTextInput {
		t.Fatalf("widget=%v want text", f.Widget())
	}

	// nil Store 不得 panic
	f.Seed(nil)
	f.Apply(nil)
}

func TestWidgetKindString(t *testing.T) {
	if WidgetCheckbox.String() != "checkbox" || WidgetNumberInput.String() != "number" || WidgetTextInput.String() != "text" {
		t.Fatal("widget kind strings")
	}
}
