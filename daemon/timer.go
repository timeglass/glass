package main

import (
	"time"
)

type Timer struct {
	mbu  time.Duration
	time time.Duration

	read chan chan time.Duration
	inc  chan chan time.Duration
}

func NewTimer(mbu time.Duration) *Timer {
	t := &Timer{
		mbu:  mbu,
		read: make(chan chan time.Duration),
		inc:  make(chan chan time.Duration),
	}

	//handle read/writes
	go func() {
		for {
			select {
			case r := <-t.read:
				r <- t.time
			case i := <-t.inc:
				t.time += <-i
			}
		}
	}()

	return t
}

func (t *Timer) Time() time.Duration {
	r := make(chan time.Duration)
	t.read <- r
	return <-r
}

func (t *Timer) Stop() error {

	return nil
}

func (t *Timer) Start() error {
	go func() {
		for {

			//start with increment
			i := make(chan time.Duration)
			t.inc <- i
			i <- t.mbu

			<-time.After(t.mbu)
		}
	}()

	return nil
}
