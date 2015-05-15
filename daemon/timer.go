package main

import (
	"time"
)

type Timer struct {
	mbu     time.Duration
	time    time.Duration
	ticking bool

	read chan chan time.Duration
	inc  chan chan time.Duration
}

func NewTimer(mbu time.Duration) *Timer {
	t := &Timer{
		mbu:  mbu,
		read: make(chan chan time.Duration),
		inc:  make(chan chan time.Duration),
	}

	//handle read& increments
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

func (t *Timer) Stop() {
	t.ticking = false
}

func (t *Timer) Start() {
	t.ticking = true
	go func() {
		for {
			//previous tick was the last mbu
			//stop ticking
			if !t.ticking {
				return
			}

			//increment with mbu
			i := make(chan time.Duration)
			t.inc <- i
			i <- t.mbu

			//wait for next mbu to arrive
			<-time.After(t.mbu)
		}
	}()
}
