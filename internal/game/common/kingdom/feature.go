package kingdom

import "app/internal/platform/screen"

type Feature struct {
	HomeFeature      screen.Feature
	AdventureFeature screen.Feature
	EventBtn         screen.Rect
	AdventureBtn     screen.Rect
	MineBtn          screen.Rect
}

func DefaultFeature() *Feature {
	return &Feature{
		HomeFeature:      screen.Feature{},
		AdventureFeature: screen.Feature{},
		EventBtn:         screen.Rect{},
		AdventureBtn:     screen.Rect{},
		MineBtn:          screen.Rect{},
	}
}
