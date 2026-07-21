package ui

import "testing"

func TestNumberInputClamp(t *testing.T) {
	p := NumberInputProps{Min: 0, Max: 99, Step: 1}
	if got := p.Clamp(120); got != 99 {
		t.Fatalf("Clamp(120)=%v want 99", got)
	}
	if got := p.Clamp(-5); got != 0 {
		t.Fatalf("Clamp(-5)=%v want 0", got)
	}
	if got := p.Clamp(42); got != 42 {
		t.Fatalf("Clamp(42)=%v want 42", got)
	}

	unbounded := NumberInputProps{Min: 0, Max: 0} // Max <= Min 表示不钳上界
	if got := unbounded.Clamp(12345); got != 12345 {
		t.Fatalf("unbounded Clamp=%v want 12345", got)
	}
	if got := unbounded.Clamp(-3); got != -3 {
		t.Fatalf("unbounded Clamp(-3)=%v want -3", got)
	}
}

func TestFormPropsCarriesFields(t *testing.T) {
	backing := false
	fp := FormProps{
		Store:  NewStore(),
		Fields: []Field{Bool("k", "开关", func() bool { return backing }, func(v bool) { backing = v })},
	}
	if len(fp.Fields) != 1 || fp.Fields[0].Key() != "k" {
		t.Fatalf("FormProps fields: %+v", fp.Fields)
	}
}
