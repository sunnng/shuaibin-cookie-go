//go:build !android || !cgo

package ui

func RunShell(opts ShellOptions) {
	if opts.Store == nil {
		opts.Store = NewStore()
	}
	if opts.Controller != nil {
		opts.Controller.Start()
	}
}
