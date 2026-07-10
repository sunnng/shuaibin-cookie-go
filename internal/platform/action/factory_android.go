//go:build android && cgo

package action

// NewExecutor returns the Android action executor for the given display.
func NewExecutor(displayId int) Executor {
	return NewAndroidExecutor(displayId)
}
