package statemachine

import (
	"errors"
	"testing"
	"time"
)

func TestMachineNextTransition(t *testing.T) {
	m := New()
	m.Init("a", Options{Timeout: 5 * time.Second})
	handlers := map[string]Handler{
		"a": func(sm *Machine) Result { return Next("b") },
		"b": func(sm *Machine) Result { return Done{} },
	}
	if err := m.Run(handlers, RunOptions{Interval: 10 * time.Millisecond}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Current != "b" {
		t.Fatalf("expected end state b, got %s", m.Current)
	}
}

func TestMachineRetryLimit(t *testing.T) {
	m := New()
	m.Init("a", Options{MaxRetry: 1, Timeout: 5 * time.Second})
	count := 0
	handlers := map[string]Handler{
		"a": func(sm *Machine) Result {
			count++
			if count < 3 {
				return Retry{}
			}
			return Done{}
		},
	}
	err := m.Run(handlers, RunOptions{Interval: 10 * time.Millisecond})
	if err == nil {
		t.Fatal("expected error from retry limit")
	}
}

func TestMachineGuardSkipTick(t *testing.T) {
	m := New()
	m.Init("a", Options{Timeout: 5 * time.Second})
	handlerCalls := 0
	guardCalls := 0
	handlers := map[string]Handler{
		"a": func(sm *Machine) Result {
			handlerCalls++
			return Done{}
		},
	}
	guard := func() bool {
		guardCalls++
		return guardCalls < 3
	}
	err := m.Run(handlers, RunOptions{Interval: 10 * time.Millisecond, Guard: guard})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if handlerCalls != 1 {
		t.Fatalf("expected handler called once, got %d", handlerCalls)
	}
	if guardCalls < 3 {
		t.Fatalf("expected guard called at least 3 times, got %d", guardCalls)
	}
}

func TestMachineZeroTimeout(t *testing.T) {
	m := New()
	m.Init("a", Options{Timeout: 0})
	handlers := map[string]Handler{
		"a": func(sm *Machine) Result { return Done{} },
	}
	if err := m.Run(handlers, RunOptions{Interval: 10 * time.Millisecond}); err != nil {
		t.Fatalf("expected zero timeout to mean no timeout, got error: %v", err)
	}
}

func TestMachineRunWithoutInit(t *testing.T) {
	m := New()
	handlers := map[string]Handler{
		"a": func(sm *Machine) Result { return Done{} },
	}
	err := m.Run(handlers, RunOptions{Interval: 10 * time.Millisecond})
	if err == nil {
		t.Fatal("expected error when Run is called without Init")
	}
}

func TestMachineFatal(t *testing.T) {
	m := New()
	m.Init("a", Options{Timeout: 5 * time.Second})
	handlers := map[string]Handler{
		"a": func(sm *Machine) Result { return Fatal{Err: errors.New("boom")} },
	}
	err := m.Run(handlers, RunOptions{Interval: 10 * time.Millisecond})
	if err == nil {
		t.Fatal("expected fatal error")
	}
}
