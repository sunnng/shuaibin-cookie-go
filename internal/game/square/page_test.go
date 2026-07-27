package square

import (
	"image"
	"testing"
	"time"

	"app/internal/platform/action"
	"app/internal/platform/screen"
)

// pageDet 按区域文本表应答 OCR，按颜色串表应答多点比色。
type pageDet struct {
	match map[string]bool
	ocr   map[screen.Region]string
}

func (d *pageDet) Capture() *image.NRGBA { return nil }
func (d *pageDet) MatchColor(x, y int, color string, sim float32) bool {
	return false
}
func (d *pageDet) FindColor(r screen.Region, color string, sim float32, dir int) (screen.Point, bool) {
	return screen.Point{}, false
}
func (d *pageDet) FindMultiColorsAll(r screen.Region, colors string, sim float32, dir int) []screen.Point {
	return nil
}
func (d *pageDet) MatchMultiColor(colors string, sim float32) bool { return d.match[colors] }
func (d *pageDet) MatchImage(r screen.Region, tpl []byte, sim float32) (screen.Point, bool) {
	return screen.Point{}, false
}
func (d *pageDet) OCRText(r screen.Region) (string, error) { return d.ocr[r], nil }
func (d *pageDet) FindOCRText(r screen.Region, keyword string) (screen.Point, bool) {
	return screen.Point{}, false
}

type pageExec struct {
	taps []action.Point
}

func (e *pageExec) Tap(p action.Point)                  { e.taps = append(e.taps, p) }
func (e *pageExec) LongTap(p action.Point, ms int)      {}
func (e *pageExec) Swipe(from, to action.Point, ms int) {}
func (e *pageExec) Back()                               {}
func (e *pageExec) Home()                               {}
func (e *pageExec) Sleep(ms int)                        {}

func dialogPage(det *pageDet) (*Page, *Feature) {
	f := DefaultFeature()
	det.match[f.Dialog.Identify.Colors] = true
	return NewPage(det, &pageExec{}, f), f
}

func TestStripNonDigits(t *testing.T) {
	cases := map[string]string{
		"240":     "240",
		" 30 个 ":  "30",
		"+12/240": "12240",
		"abc":     "",
		"":        "",
	}
	for in, want := range cases {
		if got := stripNonDigits(in); got != want {
			t.Errorf("stripNonDigits(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestTextIndicatesMaxed(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"已达到当日最大", true},
		{"最大", true},
		{"已领取全部奖励", true},
		{"已领取", false},
		{"奖励", false},
		{"12/240", false},
		{"", false},
	}
	for _, c := range cases {
		if got := textIndicatesMaxed(c.in); got != c.want {
			t.Errorf("textIndicatesMaxed(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestPageReadCount(t *testing.T) {
	det := &pageDet{match: map[string]bool{}, ocr: map[screen.Region]string{}}
	p, f := dialogPage(det)
	r := f.Dialog.Reads.RewardNow

	det.ocr[r] = " 240个 "
	n, ok := p.RewardNow()
	if !ok || n != 240 {
		t.Fatalf("RewardNow = (%d, %v), want (240, true)", n, ok)
	}

	det.ocr[r] = "abc"
	if _, ok := p.RewardNow(); ok {
		t.Fatal("non-numeric OCR should fail")
	}

	det.ocr[r] = ""
	if _, ok := p.RewardNow(); ok {
		t.Fatal("empty OCR should fail")
	}
}

func TestPageIsDailyRewardsMaxed(t *testing.T) {
	det := &pageDet{match: map[string]bool{}, ocr: map[screen.Region]string{}}
	p, f := dialogPage(det)

	det.ocr[f.Dialog.Reads.IsFinish] = "已达到当日最大"
	if !p.IsDailyRewardsMaxed() {
		t.Fatal("满额 OCR 应判定已满额")
	}

	det.ocr[f.Dialog.Reads.IsFinish] = "12/240"
	if p.IsDailyRewardsMaxed() {
		t.Fatal("普通数字不应判定满额")
	}

	// 不在离开弹窗 → 一律 false。
	det.match[f.Dialog.Identify.Colors] = false
	det.ocr[f.Dialog.Reads.IsFinish] = "最大"
	if p.IsDailyRewardsMaxed() {
		t.Fatal("不在离开弹窗时不应判定满额")
	}
}

func TestPageReadRewardSum(t *testing.T) {
	det := &pageDet{match: map[string]bool{}, ocr: map[screen.Region]string{}}
	p, f := dialogPage(det)

	det.ocr[f.Dialog.Reads.RewardNow] = "30"
	det.ocr[f.Dialog.Reads.RewardTotal] = "210"
	pending, total, sum, ok := p.ReadRewardSum()
	if !ok || pending != 30 || total != 210 || sum != 240 {
		t.Fatalf("ReadRewardSum = (%d, %d, %d, %v)", pending, total, sum, ok)
	}

	// 累计 OCR 失败 → 整体失败。
	det.ocr[f.Dialog.Reads.RewardTotal] = "--"
	if _, _, _, ok := p.ReadRewardSum(); ok {
		t.Fatal("partial OCR failure should fail the sum")
	}

	// 不在弹窗 → 失败。
	det.match[f.Dialog.Identify.Colors] = false
	if _, _, _, ok := p.ReadRewardSum(); ok {
		t.Fatal("ReadRewardSum outside dialog should fail")
	}
}

func TestPageTapUntilDialog(t *testing.T) {
	orig := tapUntilDialogTimeout
	tapUntilDialogTimeout = 50 * time.Millisecond
	t.Cleanup(func() { tapUntilDialogTimeout = orig })

	det := &pageDet{match: map[string]bool{}, ocr: map[screen.Region]string{}}
	exec := &pageExec{}
	p := NewPage(det, exec, DefaultFeature())

	// 弹窗始终不出现：超时后 false，且确实点过 TapUntilRect 区域。
	if p.TapUntilDialog() {
		t.Fatal("dialog never appears → false")
	}
	if len(exec.taps) == 0 {
		t.Fatal("expected taps on TapUntilRect")
	}

	// 弹窗已在：立即 true，不再点。
	det.match[DefaultFeature().Dialog.Identify.Colors] = true
	exec.taps = nil
	if !p.TapUntilDialog() {
		t.Fatal("dialog present → true")
	}
	if len(exec.taps) != 0 {
		t.Fatalf("dialog already present, want 0 taps, got %d", len(exec.taps))
	}
}

func TestPageTapsUseConfiguredRegions(t *testing.T) {
	det := &pageDet{match: map[string]bool{}, ocr: map[screen.Region]string{}}
	exec := &pageExec{}
	f := DefaultFeature()
	p := NewPage(det, exec, f)

	inRegion := func(pt action.Point, r screen.Region) bool {
		return pt.X >= r.X1 && pt.X <= r.X2 && pt.Y >= r.Y1 && pt.Y <= r.Y2
	}

	p.TapBack()
	p.TapCloseDialog()
	p.TapReturnKingdom()
	p.TapClaimAll()
	p.TapEntryBtn()
	if len(exec.taps) != 5 {
		t.Fatalf("want 5 taps, got %d", len(exec.taps))
	}
	regions := []screen.Region{
		f.Home.Actions.BackBtn,
		f.Dialog.Actions.CancelBtn,
		f.Dialog.Actions.LeaveBtn,
		f.Dialog.Actions.ConfirmRewardBtn,
		f.EntryBtn,
	}
	for i, r := range regions {
		if !inRegion(exec.taps[i], r) {
			t.Errorf("tap %d = %+v, not in region %+v", i, exec.taps[i], r)
		}
	}
}
