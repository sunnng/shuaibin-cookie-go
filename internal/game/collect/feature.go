package collect

import "app/internal/platform/screen"

// Feature holds UI constants for the collect module (fill via color-picking later).
type Feature struct {
	// Identify placeholder for a future page identity check.
	Identify screen.Feature
}

func DefaultFeature() *Feature {
	return &Feature{}
}
