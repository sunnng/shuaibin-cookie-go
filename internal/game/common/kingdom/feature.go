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
		// Placeholder values; replace with real features from Lua 通用_王国/特征库.lua
		HomeFeature:      screen.Feature{},
		AdventureFeature: screen.Feature{},
		EventBtn:         screen.Rect{},
		AdventureBtn:     screen.Rect{},
		MineBtn:          screen.Rect{},
	}
}
