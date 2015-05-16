package main

import (
	"sync"
	"time"
)

type Timer struct {
	mbu     time.Duration
	timeout time.Duration
	time    time.Duration
	ticking bool
	read    chan chan time.Duration
	inc     chan chan time.Duration
	reset   chan struct{}

	Wakeup <-chan interface{}
	*sync.Mutex
}

func NewTimer(mbu time.Duration, to time.Duration) *Timer {
	t := &Timer{
		mbu:     mbu,
		timeout: to,

		read:  make(chan chan time.Duration),
		inc:   make(chan chan time.Duration),
		reset: make(chan struct{}),

		Wakeup: make(chan interface{}),
		Mutex:  &sync.Mutex{},
	}

	//handle read& writes
	go func() {
		for {
			select {
			case r := <-t.read:
				r <- t.time
			case i := <-t.inc:
				t.time += <-i
			case <-t.reset:
				t.time = 0
			}
		}
	}()

	//handle timeout and wakeup
	go func() {
		for {
			select {
			case <-time.After(t.timeout):
				t.Stop()
			case <-t.Wakeup:
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

func (t *Timer) Reset() {
	t.reset <- struct{}{}
}

func (t *Timer) Stop() {
	t.Lock()
	defer t.Unlock()
	t.ticking = false
}

func (t *Timer) Start() {
	t.Lock()
	defer t.Unlock()
	if t.ticking {
		return
	}

	t.ticking = true
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
