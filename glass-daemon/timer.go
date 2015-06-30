package main

import (
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/timeglass/glass/_vendor/github.com/hashicorp/errwrap"

	"github.com/timeglass/snow/monitor"
)

type timerData struct {
	paused   bool          `json:"paused"`
	confPath string        `json:"conf_path"`
	latency  time.Duration `json:"latency"`
	timeout  time.Duration `json:"timeout"`
	mbu      time.Duration `json:"mbu"`
	time     time.Duration `json:"time"`
}

type Timer struct {
	timerData *timerData
	monitor   monitor.M
	stop      chan struct{}
}

func NewTimer(confPath string) (*Timer, error) {
	p := &Timer{
		stop: make(chan struct{}),
		timerData: &timerData{
			confPath: confPath,
			latency:  time.Millisecond * 50, //@todo make configurable
			mbu:      time.Minute,           //@todo make configurable
			timeout:  time.Minute * 4,       //@todo make configurable
		},
	}

	return p, nil
}

func (p *Timer) Start() error {
	//lazily initiate monitor
	var err error
	if p.monitor == nil {
		dir := filepath.Dir(p.timerData.confPath)
		p.monitor, err = monitor.New(dir, monitor.Recursive, p.timerData.latency)
		if err != nil {
			return errwrap.Wrapf(fmt.Sprintf("Failed to create monitor for directory '%s': {{err}}", dir), err)
		}
	}

	wakup, err := p.monitor.Start()
	if err != nil {
		return errwrap.Wrapf("Failed to start monitor: {{err}}", err)
	}

	//handle timeouts and wakeups
	log.Printf("Timer for project '%s' was started (and unpaused) explicitely", p.timerData.confPath)
	p.timerData.paused = false
	go func() {
		for {
			select {
			case <-p.stop:
				log.Printf("Timer for project '%s' was stopped (and paused) explicitely", p.timerData.confPath)
				p.timerData.paused = true
				break
			case merr := <-p.monitor.Errors():
				log.Printf("Monitor Error: %s", merr)
			case <-time.After(p.timerData.timeout):
				log.Printf("Timer for project '%s' timed out after %s", p.timerData.confPath, p.timerData.timeout)
				p.timerData.paused = true
			case ev := <-wakup:
				log.Printf("Timer for project '%s' woke up after some activity in '%s'", p.timerData.confPath, ev.Dir())
				p.timerData.paused = false
			}
		}
	}()

	//handle time increments
	go func() {
		for {
			if !p.timerData.paused {
				p.timerData.time += p.timerData.mbu
			}

			select {
			case <-p.stop:
				break
			case <-time.After(p.timerData.mbu):
			}
		}
	}()

	return nil
}

func (p *Timer) IsPaused() bool {
	return p.timerData.paused
}

func (t *Timer) Time() time.Duration {
	return t.timerData.time
}

func (p *Timer) Stop() error {
	p.stop <- struct{}{}

	err := p.monitor.Stop()
	if err != nil {
		return errwrap.Wrapf("Failed to stop monitor: {{err}}", err)
	}

	return nil
}

func (p *Timer) ConfPath() string {
	return p.timerData.confPath
}

func (p *Timer) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &p.timerData)
}

func (p *Timer) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.timerData)
}
