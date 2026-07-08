package bot

type State interface {
	Name() string
	Detect(ctx *Context) bool
	Act(ctx *Context) error
	Next(ctx *Context) State
	Recover(ctx *Context) error
}

type Registry struct {
	states []State
}

func NewRegistry() *Registry {
	return &Registry{}
}

func (r *Registry) Register(s State) {
	r.states = append(r.states, s)
}

func (r *Registry) Find(ctx *Context) State {
	for _, s := range r.states {
		if s.Detect(ctx) {
			return s
		}
	}
	return nil
}
