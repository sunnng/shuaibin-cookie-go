package logger

import "testing"

func TestLoggerInfof(t *testing.T) {
	SetLevel(3)
	Infof("test %s", "info")
}

func TestLoggerLevel(t *testing.T) {
	SetLevel(1)
	Debugf("should not print")
	SetLevel(4)
	Debugf("should print")
}
