//go:build !android || !cgo

package ui

// RunShell 桌面/非 cgo stub：无 UI，Seed 后直接启动脚本，保证 go build/test
// 在任何平台可用。（较旧 internal/ui stub 多了 Seed：填充缺失键，无害。）
func RunShell(opts ShellOptions) {
	shell := NewShell(opts)
	shell.Seed()
	if opts.Controller != nil {
		opts.Controller.Start()
	}
}
