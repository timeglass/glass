package main

import (
	"time"
)

type Timer struct {
	mbu time.Duration
}

func NewTimer(mbu time.Duration) *Timer {
	return &Timer{mbu}
}

//@todo implement
func (t *Timer) Start() error {
	return nil
}
