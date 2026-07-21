package ui

import "testing"

func TestHexParsesRGBA(t *testing.T) {
	c := Hex("#2f8fd0ff")
	want := Color{R: 0x2f / 255.0, G: 0x8f / 255.0, B: 0xd0 / 255.0, A: 1}
	if c != want {
		t.Fatalf("Hex=#%+v want %+v", c, want)
	}
}

func TestHexDefaultsAlphaAndRejectsBadInput(t *testing.T) {
	if c := Hex("#ffffff"); c.A != 1 || c.R != 1 {
		t.Fatalf("6-digit hex should default alpha=1: %+v", c)
	}
	for _, bad := range []string{"", "#12345", "zzzzzz", "#1234567890"} {
		if c := Hex(bad); c != (Color{0, 0, 0, 1}) {
			t.Fatalf("Hex(%q)=%+v want opaque black", bad, c)
		}
	}
}

func TestDefaultThemeIsQQBlue(t *testing.T) {
	th := DefaultTheme()
	if th.Accent != Hex("#2f8fd0ff") {
		t.Fatalf("accent=%+v", th.Accent)
	}
	if th.TitleTop != Hex("#5aa9e6ff") || th.TitleBottom != Hex("#2f7fc4ff") {
		t.Fatal("title gradient colors")
	}
	if th.RailBg != Hex("#cfe4f7ff") {
		t.Fatal("rail bg")
	}
	if th.White != (Color{1, 1, 1, 1}) {
		t.Fatal("white")
	}
	if th.Rounding != 4 {
		t.Fatalf("rounding=%v want 4", th.Rounding)
	}
	// 主题可作零值比较（ShellOptions 缺省判断依赖这一点）
	if th == (Theme{}) {
		t.Fatal("default theme must differ from zero value")
	}
}
