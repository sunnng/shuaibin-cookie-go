//go:build !android || !cgo

package screen

import "testing"

func TestStubFindMultiColorsAllReturnsNil(t *testing.T) {
	d := NewDetector(0)
	got := d.FindMultiColorsAll(Region{0, 0, 100, 100}, "ffffff", 0.9, 0)
	if got != nil {
		t.Fatalf("stub FindMultiColorsAll should return nil, got %v", got)
	}
}

func TestStubFindOCRTextReturnsFalse(t *testing.T) {
	d := NewDetector(0)
	if _, ok := d.FindOCRText(Region{0, 0, 100, 100}, "王国竞技场"); ok {
		t.Fatal("stub FindOCRText should return false")
	}
	if _, ok := d.FindOCRText(Region{0, 0, 100, 100}, ""); ok {
		t.Fatal("empty keyword should return false")
	}
}
