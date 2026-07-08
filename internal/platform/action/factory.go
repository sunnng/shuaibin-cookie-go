//go:build !android

package action

import "time"

// stubExecutor is a non-Android placeholder that performs no real input.
type stubExecutor struct{}

func NewExecutor(displayId int) Executor { return &stubExecutor{} }

func (s *stubExecutor) Tap(p Point) error                  { return nil }
func (s *stubExecutor) LongTap(p Point, ms int) error      { return nil }
func (s *stubExecutor) Swipe(from, to Point, ms int) error { return nil }
func (s *stubExecutor) Back() error                        { return nil }
func (s *stubExecutor) Home() error                        { return nil }
func (s *stubExecutor) Sleep(ms int)                       { time.Sleep(time.Duration(ms) * time.Millisecond) }
