package arena

import (
	"testing"
	"time"

	"app/internal/platform/screen"
)

func TestNewPage(t *testing.T) {
	_ = NewPage(nil, nil, DefaultFeature())
}

func TestOffsetRegion(t *testing.T) {
	rel := screen.Region{X1: -10, Y1: -20, X2: 10, Y2: 20}
	a := screen.Point{X: 100, Y: 200}
	got := offsetRegion(rel, a)
	want := screen.Region{X1: 90, Y1: 180, X2: 110, Y2: 220}
	if got != want {
		t.Fatalf("offsetRegion = %+v, want %+v", got, want)
	}
}

func TestOffsetPoint(t *testing.T) {
	got := offsetPoint(screen.Point{X: 30, Y: -5}, screen.Point{X: 100, Y: 200})
	want := screen.Point{X: 130, Y: 195}
	if got != want {
		t.Fatalf("offsetPoint = %+v, want %+v", got, want)
	}
}

func TestReadInt(t *testing.T) {
	cases := []struct {
		in     string
		wantN  int
		wantOK bool
	}{
		{"1050", 1050, true},
		{"  99 ", 99, true},
		{"abc", 0, false},
		{"", 0, false},
	}
	for _, c := range cases {
		n, ok := readInt(c.in)
		if n != c.wantN || ok != c.wantOK {
			t.Errorf("readInt(%q) = (%d,%v), want (%d,%v)", c.in, n, ok, c.wantN, c.wantOK)
		}
	}
}

func TestParseCountdown(t *testing.T) {
	cases := []struct {
		in     string
		want   time.Duration
		wantOK bool
	}{
		{"5分30秒", 330 * time.Second, true},
		{"30秒", 30 * time.Second, true},
		{"5分", 300 * time.Second, true},
		{"05:30", 330 * time.Second, true},
		{"", 0, false},
		{"abc", 0, false},
		{"0秒", 0, false},
	}
	for _, c := range cases {
		got, ok := parseCountdown(c.in)
		if got != c.want || ok != c.wantOK {
			t.Errorf("parseCountdown(%q) = (%v,%v), want (%v,%v)", c.in, got, ok, c.want, c.wantOK)
		}
	}
}
