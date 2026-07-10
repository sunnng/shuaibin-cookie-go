//go:build !android || !cgo

package ui

func RunPanel(opts PanelOptions) {
	if opts.Store == nil {
		opts.Store = NewStore()
	}
	if opts.OnRun != nil {
		opts.OnRun(opts.Store)
	}
}

func RunShell(opts ShellOptions) {
	if opts.Store == nil {
		opts.Store = NewStore()
	}
	if opts.Controller != nil {
		opts.Controller.Start()
	}
}

func DefaultCookiePanel(store *Store) {}
