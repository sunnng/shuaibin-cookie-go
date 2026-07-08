package bot

type State interface {
	Name() string
	Detect(ctx *Context) bool
	Act(ctx *Context) error
	Next(ctx *Context) State
	Recover(ctx *Context) error
}
