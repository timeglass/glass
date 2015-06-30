package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/timeglass/glass/_vendor/github.com/hashicorp/errwrap"

	"github.com/timeglass/snow/monitor"
)

type timerData struct {
	Paused  bool          `json:"paused"`
	Dir     string        `json:"conf_path"`
	Latency time.Duration `json:"latency"`
	Timeout time.Duration `json:"timeout"`
	MBU     time.Duration `json:"mbu"`
	Time    time.Duration `json:"time"`
}

type Timer struct {
	timerData *timerData
	monitor   monitor.M
	stop      chan struct{}
}

func NewTimer(dir string) (*Timer, error) {
	t := &Timer{
		stop: make(chan struct{}),
		timerData: &timerData{
			Dir:     dir,
			Latency: time.Millisecond * 50, //@todo make configurable
			MBU:     time.Minute,           //@todo make configurable
			Timeout: time.Minute * 4,       //@todo make configurable
		},
	}

	return t, nil
}

func (t *Timer) Start() error {
	//lazily initiate monitor
	var err error
	if t.monitor == nil {
		t.monitor, err = monitor.New(t.Dir(), monitor.Recursive, t.timerData.Latency)
		if err != nil {
			return errwrap.Wrapf(fmt.Sprintf("Failed to create monitor for directory '%s': {{err}}", t.Dir()), err)
		}
	}

	wakup, err := t.monitor.Start()
	if err != nil {
		return errwrap.Wrapf("Failed to start monitor: {{err}}", err)
	}

	//handle timeouts and wakeups
	log.Printf("Timer for project '%s' was started (and unpaused) explicitely", t.Dir())
	t.timerData.Paused = false
	go func() {
		for {
			select {
			case <-t.stop:
				log.Printf("Timer for project '%s' was stopped (and paused) explicitely", t.Dir())
				t.timerData.Paused = true
				break
			case merr := <-t.monitor.Errors():
				log.Printf("Monitor Error: %s", merr)
			case <-time.After(t.timerData.Timeout):
				log.Printf("Timer for project '%s' timed out after %s", t.Dir(), t.timerData.Timeout)
				t.timerData.Paused = true
			case ev := <-wakup:
				log.Printf("Timer for project '%s' woke up after some activity in '%s'", t.Dir(), ev.Dir())
				t.timerData.Paused = false
			}
		}
	}()

	//handle time increments
	go func() {
		for {
			if !t.timerData.Paused {
				t.timerData.Time += t.timerData.MBU
			}

			select {
			case <-t.stop:
				break
			case <-time.After(t.timerData.MBU):
			}
		}
	}()

	return nil
}

func (t *Timer) Stop() error {
	t.stop <- struct{}{}
	err := t.monitor.Stop()
	if err != nil {
		return errwrap.Wrapf("Failed to stop monitor: {{err}}", err)
	}

	return nil
}

func (t *Timer) IsPaused() bool {
	return t.timerData.Paused
}

func (t *Timer) Time() time.Duration {
	return t.timerData.Time
}

func (t *Timer) Dir() string {
	return t.timerData.Dir
}

func (t *Timer) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &t.timerData)
}

func (t *Timer) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.timerData)
}
