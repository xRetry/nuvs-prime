package main

import (
	"log"
)

type Logger struct {
	isEnabled bool
}

func (l Logger) Printf(msg string, args ...any) {
	if l.isEnabled {
		log.Printf(msg, args...)
	}
}

func (l Logger) Println(args ...any) {
	if l.isEnabled {
		log.Println(args...)
	}
}

func (l Logger) Fatalf(msg string, args ...any) {
	log.Fatalf(msg, args...)
}
