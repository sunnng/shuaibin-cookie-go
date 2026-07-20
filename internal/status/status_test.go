package status

import "testing"

func TestReporterDefaultEmpty(t *testing.T) {
	if got := New().Text(); got != "" {
		t.Fatalf("default text = %q, want empty", got)
	}
}

func TestReporterSetAndText(t *testing.T) {
	r := New()
	r.Set("竞技场 3/10 · 胜率 67%")
	if got := r.Text(); got != "竞技场 3/10 · 胜率 67%" {
		t.Fatalf("text = %q", got)
	}
	r.Set("")
	if got := r.Text(); got != "" {
		t.Fatalf("cleared text = %q, want empty", got)
	}
}
