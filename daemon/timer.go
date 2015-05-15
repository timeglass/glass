package main

import (
	"sync"
	"time"
)

type Timer struct {
	mbu     time.Duration
	time    time.Duration
	ticking bool
	read    chan chan time.Duration
	inc     chan chan time.Duration

	*sync.Mutex
}

func NewTimer(mbu time.Duration) *Timer {
	t := &Timer{
		mbu:  mbu,
		read: make(chan chan time.Duration),
		inc:  make(chan chan time.Duration),

		Mutex: &sync.Mutex{},
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
	t.Lock()
	defer t.Unlock()
	t.ticking = false
}

func (t *Timer) Start() {
	t.Lock()
	if t.ticking {
		t.Unlock()
		return
	}

	t.ticking = true
	t.Unlock()

	go func() {
		for {

			//wait for next mbu to arrive
			<-time.After(t.mbu)

			//previous tick was the last mbu, don't
			//increment this mbu
			if !t.ticking {
				return
			}

			//increment with mbu
			i := make(chan time.Duration)
			t.inc <- i
			i <- t.mbu

		}
	}()
}
