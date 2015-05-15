package main

import (
	"sync"
	"time"
)

type Timer struct {
	mbu  time.Duration
	time time.Duration

	*sync.Mutex
}

func NewTimer(mbu time.Duration) *Timer {
	return &Timer{
		mbu:   mbu,
		Mutex: &sync.Mutex{},
	}
}

func (t *Timer) Time() time.Duration {
	t.Lock()
	defer t.Unlock()
	return t.time
}

func (t *Timer) Start() error {
	go func() {
		for {

			//reserve at least one mbu
			//upon starting
			t.Lock()
			t.time += t.mbu
			t.Unlock()

			<-time.After(t.mbu)
		}
	}()

	return nil
}
