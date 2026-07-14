package collect

// Session holds collect runtime counters (extend when implementing the feature).
type Session struct{}

func NewSession() *Session {
	return &Session{}
}

func (s *Session) Reset() {}
