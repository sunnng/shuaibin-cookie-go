//go:build !android || !cgo

package action

import "time"

// stubExecutor is a non-Android placeholder that performs no real input.
type stubExecutor struct{}

func NewExecutor(displayId int) Executor { return &stubExecutor{} }

func (s *stubExecutor) Tap(p Point)                  {}
func (s *stubExecutor) LongTap(p Point, ms int)      {}
func (s *stubExecutor) Swipe(from, to Point, ms int) {}
func (s *stubExecutor) Back()                        {}
func (s *stubExecutor) Home()                        {}
func (s *stubExecutor) Sleep(ms int) {
	time.Sleep(time.Duration(ms) * time.Millisecond)
}
