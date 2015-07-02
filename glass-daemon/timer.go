package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/timeglass/glass/_vendor/github.com/hashicorp/errwrap"

	"github.com/timeglass/glass/config"
	"github.com/timeglass/snow/monitor"
)

type timerData struct {
	Failed  string        `json:"failed"`
	Paused  bool          `json:"paused"`
	Dir     string        `json:"conf_path"`
	Latency time.Duration `json:"latency"`
	Timeout time.Duration `json:"timeout"`
	MBU     time.Duration `json:"mbu"`
	Time    time.Duration `json:"time"`
}

type Timer struct {
	running   bool
	timerData *timerData
	monitor   monitor.M
	save      chan struct{}
	stopto    chan struct{}
	stoptick  chan struct{}
	reset     chan struct{}
}

func NewTimer(dir string) (*Timer, error) {
	t := &Timer{
		timerData: &timerData{
			Dir:     dir,
			MBU:     time.Minute,
			Latency: time.Millisecond * 50, //@todo make configurable
			Timeout: time.Minute * 4,       //@todo make configurable
		},
	}

	return t, nil
}

// Start get called in a multitude of different situations:
//  - when the service starts after a reboot and loads timer state from the ledger
//  - when a new timer is added (for a new project)
func (t *Timer) Start() {
	var err error

	//already running and not failed? no-op
	if t.running && t.HasFailed() == "" {
		return
	}

	t.timerData.Failed = ""

	//load project specific configuration
	conf, err := config.ReadConfig(t.Dir())
	if err != nil {
		err = errwrap.Wrapf(fmt.Sprintf("Failed to read configuration for '%s': {{err}}, using default", t.Dir()), err)
		t.timerData.Failed = err.Error()
		conf = config.DefaultConfig
	}

	t.timerData.MBU = time.Duration(conf.MBU)
	t.timerData.Timeout = 4 * t.timerData.MBU

	//lazily initiate control members
	t.stopto = make(chan struct{})
	t.stoptick = make(chan struct{})
	t.reset = make(chan struct{})

	//setup monitor, if not done yet
	wakeup := make(chan monitor.DirEvent)
	merrs := make(chan error)
	if t.monitor == nil {
		t.monitor, err = monitor.New(t.Dir(), monitor.Recursive, t.timerData.Latency)
		if err != nil {
			err = errwrap.Wrapf(fmt.Sprintf("Failed to create monitor for directory '%s': {{err}}", t.Dir()), err)
			t.timerData.Failed = err.Error()
			log.Print(err)
		} else {
			wakeup, err = t.monitor.Start()
			if err != nil {
				err = errwrap.Wrapf("Failed to start monitor: {{err}}", err)
				t.timerData.Failed = err.Error()
				log.Print(err)
			}

			merrs = t.monitor.Errors()
		}
	} else {
		wakeup = t.monitor.Events()
		merrs = t.monitor.Errors()
	}

	//handle stops, pauses, timeouts and wakeups
	log.Printf("Timer for project '%s' was started (and unpaused) explicitely", t.Dir())
	t.timerData.Paused = false
	t.running = true
	go func() {
		for {

			t.EmitSave()
			select {
			case <-t.stopto:
				log.Printf("Timer for project '%s' was stopped (and paused) explicitely", t.Dir())
				return
			case merr := <-merrs:
				log.Printf("Monitor Error: %s", merr)
				t.timerData.Failed = merr.Error()
			case <-time.After(t.timerData.Timeout):
				if !t.IsPaused() {
					log.Printf("Timer for project '%s' timed out after %s", t.Dir(), t.timerData.Timeout)
				}
				t.Pause()
			case ev := <-wakeup:
				if t.IsPaused() {
					log.Printf("Timer for project '%s' woke up after some activity in '%s'", t.Dir(), ev.Dir())
					t.Unpause()
				} else {
					log.Printf("Timer saw activity for project '%s' in '%s' but is already unpaused", t.Dir(), ev.Dir())
				}
			}
		}
	}()

	//handle time modifications here
	go func() {
		for {
			if !t.timerData.Paused {
				t.timerData.Time += t.timerData.MBU
			}

			t.EmitSave()
			select {
			case <-t.stoptick:
				return
			case <-t.reset:
				t.timerData.Time = 0
				log.Printf("Timer for project '%s' was reset", t.Dir())
			case <-time.After(t.timerData.MBU):
			}
		}
	}()
}

func (t *Timer) Pause() {
	if !t.running || t.IsPaused() {
		return
	}

	t.timerData.Paused = true
	log.Printf("Timer for project '%s' was paused", t.Dir())
}

func (t *Timer) Unpause() {
	if !t.running || !t.IsPaused() {
		return
	}

	t.timerData.Paused = false
	log.Printf("Timer for project '%s' was unpaused", t.Dir())
}

func (t *Timer) Reset() {
	if !t.running {
		//if running state is not correct
		//this migth cause race conditions
		t.timerData.Time = 0
		return
	}

	t.reset <- struct{}{}
}

func (t *Timer) Stop() {
	if !t.running {
		return
	}

	if t.monitor != nil {

		//@todo Remove this at some point, it normalizes
		//time after rapid stop start usage on OSX
		//for unkown reasons, long term solution should probably
		//involve some mechanism that prevents the darwin monitor
		//form stopping to quickly after being started
		<-time.After(time.Millisecond)

		err := t.monitor.Stop()
		if err != nil {
			log.Print(errwrap.Wrapf("Failed to stop monitor: {{err}}", err))
		}

		t.monitor = nil
	}

	t.stopto <- struct{}{}
	t.stoptick <- struct{}{}

	t.timerData.Paused = true
	t.running = false
}

func (t *Timer) EmitSave() {
	if t.save != nil {
		t.save <- struct{}{}
	}
}

func (t *Timer) SetSave(ch chan struct{}) {
	t.save = ch
}

func (t *Timer) HasFailed() string {
	return t.timerData.Failed
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
