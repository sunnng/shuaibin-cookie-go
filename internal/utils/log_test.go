package utils

import "testing"

func TestLoggerExists(t *testing.T) {
    Infof("test %s", "info")
    Errorf("test %s", "error")
}
