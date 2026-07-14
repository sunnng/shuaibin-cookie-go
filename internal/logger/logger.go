package logger

import (
	"log"
	"sync"
)

const (
	LevelError = 1
	LevelWarn  = 2
	LevelInfo  = 3
	LevelDebug = 4
)

var (
	level int = LevelInfo
	mu    sync.RWMutex
)

func SetLevel(l int) {
	mu.Lock()
	defer mu.Unlock()
	level = l
}

func getLevel() int {
	mu.RLock()
	defer mu.RUnlock()
	return level
}

func logf(minLevel int, tag string, format string, args ...any) {
	if getLevel() < minLevel {
		return
	}
	log.Printf("["+tag+"] "+format, args...)
}

func Errorf(format string, args ...any) { logf(LevelError, "ERROR", format, args...) }
func Warnf(format string, args ...any)  { logf(LevelWarn, "WARN", format, args...) }
func Infof(format string, args ...any)  { logf(LevelInfo, "INFO", format, args...) }
func Debugf(format string, args ...any) { logf(LevelDebug, "DEBUG", format, args...) }
