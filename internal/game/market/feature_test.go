package market

import (
	"strings"
	"testing"
)

func TestGoColors(t *testing.T) {
	cases := []struct{ in, want string }{
		{"37|723|f51b67-101010,189|95|14633c-101010", "37,723,f51b67-101010,189,95,14633c-101010"},
		{"-7|-1|fffff3-101010|-34|-1|cfefb9-101010", "-7,-1,fffff3-101010,-34,-1,cfefb9-101010"},
		{"plain", "plain"},
		{"", ""},
	}
	for _, c := range cases {
		if got := goColors(c.in); got != c.want {
			t.Errorf("goColors(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestDefaultFeature(t *testing.T) {
	f := DefaultFeature()

	if len(f.Stock) != 7 {
		t.Fatalf("stock count = %d, want 7", len(f.Stock))
	}
	// 页面身份串应为 Go 多点比色格式（无 "|"）
	if strings.Contains(f.Page.Identify.Colors, "|") {
		t.Errorf("page identify colors not converted: %q", f.Page.Identify.Colors)
	}
	if !strings.HasPrefix(f.Page.Identify.Colors, "37,723,f51b67-101010,") {
		t.Errorf("page identify prefix = %q", f.Page.Identify.Colors)
	}
	// 商品特征：baseColor + 偏移串转换
	got := f.Stock["灿烂的光之碎片"].Colors
	if !strings.HasPrefix(got, "a9e4ff-101010,-7,-1,fffff3-101010,") {
		t.Errorf("stock colors prefix = %q", got)
	}
	if strings.Contains(got, "|") {
		t.Errorf("stock colors not converted: %q", got)
	}
	if r := f.Stock["灿烂的光之碎片"].Region; r.X1 != 3 || r.Y1 != 602 || r.X2 != 1587 || r.Y2 != 707 {
		t.Errorf("stock region = %+v", r)
	}
	// 箭头保留 "|" 备选串（FindColor 原生支持）
	if !strings.Contains(f.Page.Arrow.Colors, "|") {
		t.Errorf("arrow colors should keep '|' alternates: %q", f.Page.Arrow.Colors)
	}
	if f.Entry.Btn.X1 != 574 || f.Entry.Btn.Y1 != 582 || f.Entry.Btn.X2 != 593 || f.Entry.Btn.Y2 != 604 {
		t.Errorf("entry btn = %+v", f.Entry.Btn)
	}
	if f.List.MaxSwipes != 20 {
		t.Errorf("max swipes = %d", f.List.MaxSwipes)
	}
}

func TestDefaultConfig(t *testing.T) {
	c := DefaultConfig()
	if c.Enabled {
		t.Error("default enabled should be false")
	}
	if len(c.Items) != 7 {
		t.Fatalf("default items = %d, want 7", len(c.Items))
	}
	if c.Items[0] != "灿烂的光之碎片" {
		t.Errorf("items[0] = %q", c.Items[0])
	}
	if c.BufferSec() != 30 {
		t.Errorf("buffer = %d, want 30", c.BufferSec())
	}

	var nilCfg *Config
	if nilCfg.BufferSec() != 30 {
		t.Error("nil cfg buffer should fall back to 30")
	}
	bad := &Config{RestockBufferSec: -5}
	if bad.BufferSec() != 30 {
		t.Error("negative buffer should fall back to 30")
	}
}
