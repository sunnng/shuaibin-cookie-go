package utils

import (
	"log"
)

func Infof(format string, args ...any) {
	log.Printf("[INFO] "+format, args...)
}

func Errorf(format string, args ...any) {
	log.Printf("[ERROR] "+format, args...)
}

func PrintStateTransition(from, to string) {
	Infof("state transition: %s -> %s", from, to)
}
