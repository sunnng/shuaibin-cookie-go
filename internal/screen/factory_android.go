//go:build android

package screen

// NewDetector returns the Android screen detector for the given display.
func NewDetector(displayId int) Detector {
	return NewAndroidDetector(displayId)
}
