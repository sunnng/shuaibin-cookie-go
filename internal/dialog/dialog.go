package dialog

import (
	"time"

	"app/internal/screen"
)

// Def describes a system dialog and its interactive elements.
type Def struct {
	Name       string
	Feature    any
	ConfirmBtn any
	CancelBtn  any
}

// HandleOpts controls how Handle detects and dismisses a dialog.
type HandleOpts struct {
	Mode     string // "ifVisible" or "flow"
	Action   string // "confirm" or "cancel"
	WaitGone time.Duration
	TapDelay time.Duration
	Interval time.Duration
}

// Dialog wraps a dialog definition and provides visibility / handling helpers.
// The current implementation is a placeholder: real button tapping and wait-gone
// logic will be added once the platform action package is wired in.
type Dialog struct {
	def      Def
	tag      string
	detector screen.Detector
}

// New creates a Dialog for the given definition and logging tag.
func New(def Def, tag string) *Dialog {
	return &Dialog{def: def, tag: tag}
}

// IsVisible reports whether the dialog is currently on screen.
// It returns false when no feature is defined or no detector is available.
// A string feature is interpreted as a multi-color match against the detector.
// A func() bool feature is evaluated directly.
func (d *Dialog) IsVisible() bool {
	if d.def.Feature == nil {
		return false
	}
	switch f := d.def.Feature.(type) {
	case func() bool:
		return f()
	case string:
		if d.detector == nil {
			return false
		}
		return d.detector.MatchMultiColor(f, 0.9)
	default:
		return false
	}
}

// Handle attempts to handle the dialog according to opts.
// In "ifVisible" mode it is a no-op when the dialog is not visible.
// The current implementation returns success without actually tapping.
func (d *Dialog) Handle(opts HandleOpts) (bool, string) {
	if opts.Mode == "ifVisible" && !d.IsVisible() {
		return true, ""
	}
	// Placeholder: real implementation taps confirm/cancel and waits gone
	return true, ""
}

// ToGuardHandler adapts Handle to the guard trap handler signature.
func (d *Dialog) ToGuardHandler(opts HandleOpts) func() error {
	return func() error {
		_, _ = d.Handle(opts)
		return nil
	}
}
