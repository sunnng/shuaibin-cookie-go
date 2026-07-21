package ui

import "testing"

func TestComputeScaleWidthDriven(t *testing.T) {
	cases := []struct {
		dw, dh, bw, bh int
		want           float64
	}{
		{1600, 900, 1600, 900, 1.0},
		{2400, 1080, 1600, 900, 1.5},
		{800, 450, 1600, 900, 0.5},
		{0, 0, 1600, 900, 1.0}, // 无效显示尺寸回退 1
		{1600, 900, 0, 0, 1.0}, // 无效基准回退 1
		{-100, 900, 1600, 900, 1.0},
	}
	for _, c := range cases {
		if got := ComputeScale(c.dw, c.dh, c.bw, c.bh); got != c.want {
			t.Errorf("ComputeScale(%d,%d,%d,%d)=%v want %v", c.dw, c.dh, c.bw, c.bh, got, c.want)
		}
	}
}
