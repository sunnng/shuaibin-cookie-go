package collect

import (
	"app/internal/platform/action"
	"app/internal/platform/screen"
)

// Page is the collect UI surface. Skeleton: no real recognition yet.
type Page struct {
	detector screen.Detector
	executor action.Executor
	feature  *Feature
}

func NewPage(det screen.Detector, exec action.Executor, f *Feature) *Page {
	if f == nil {
		f = DefaultFeature()
	}
	return &Page{detector: det, executor: exec, feature: f}
}
