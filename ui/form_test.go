package ui

import "testing"

func TestFormFieldValueAndChanged(t *testing.T) {
	bf := Bool("b", "开关", func() bool { return false }, func(bool) {})
	nf := Number("n", "数量", 0, 99, 1, func() int { return 0 }, func(int) {})
	tf := Text("t", "文本", func() string { return "" }, func(string) {})

	s := NewStore()
	s.SetBool("b", true)
	s.SetFloat("n", 7)
	s.SetString("t", "hello")

	if v := FormFieldValue(s, bf); v != true {
		t.Fatalf("bool value=%v (%T)", v, v)
	}
	if v := FormFieldValue(s, nf); v != float64(7) {
		t.Fatalf("number value=%v (%T)", v, v)
	}
	if v := FormFieldValue(s, tf); v != "hello" {
		t.Fatalf("text value=%v (%T)", v, v)
	}

	FormFieldChanged(s, bf, false)
	FormFieldChanged(s, nf, 8.6)
	FormFieldChanged(s, tf, "world")
	if s.GetBool("b") || s.GetFloat("n") != 8.6 || s.GetString("t") != "world" {
		t.Fatalf("changed: b=%v n=%v t=%v", s.GetBool("b"), s.GetFloat("n"), s.GetString("t"))
	}

	// 类型不符的 v 安全忽略，nil store 不 panic
	FormFieldChanged(s, bf, "oops")
	if s.GetBool("b") != false {
		t.Fatal("mismatched type should be ignored")
	}
	FormFieldChanged(nil, bf, true)
	if v := FormFieldValue(nil, bf); v != false {
		t.Fatalf("nil store bool zero value, got %v", v)
	}
}
