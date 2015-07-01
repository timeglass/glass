package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/timeglass/glass/_vendor/github.com/hashicorp/errwrap"

	"github.com/timeglass/glass/model"
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
	timerData *timerData
	monitor   monitor.M
	stopto    chan struct{}
	stoptick  chan struct{}
	reset     chan struct{}
	pause     chan struct{}
	unpause   chan struct{}
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
func (t *Timer) Start() error {
	var err error

	//load project specific configuration
	conf := &model.Config{}
	confp := filepath.Join(t.Dir(), "timeglass.json")
	confdata, err := ioutil.ReadFile(confp)
	if err != nil && !os.IsNotExist(err) {
		log.Printf("Failed to read project configuration even though it exist: %s", err)
	} else if os.IsNotExist(err) {
		conf = model.DefaultConfig
	} else {
		err := json.Unmarshal(confdata, &conf)
		if err != nil {
			log.Printf("Failed to parse configuration JSON: %s, using default config", err)
			conf = model.DefaultConfig
		}
	}

	t.timerData.MBU = time.Duration(conf.MBU)
	t.timerData.Timeout = 4 * t.timerData.MBU

	//lazily initiate control members
	t.timerData.Failed = ""
	t.stopto = make(chan struct{})
	t.stoptick = make(chan struct{})
	t.reset = make(chan struct{})
	t.pause = make(chan struct{})
	t.unpause = make(chan struct{})
	if t.monitor == nil {
		t.monitor, err = monitor.New(t.Dir(), monitor.Recursive, t.timerData.Latency)
		if err != nil {
			err = errwrap.Wrapf(fmt.Sprintf("Failed to create monitor for directory '%s': {{err}}", t.Dir()), err)
			t.timerData.Failed = err.Error()
			return err
		}
	}

	wakup, err := t.monitor.Start()
	if err != nil {
		err = errwrap.Wrapf("Failed to start monitor: {{err}}", err)
		t.timerData.Failed = err.Error()
		return err
	}

	//handle stops, pauses, timeouts and wakeups
	log.Printf("Timer for project '%s' was started (and unpaused) explicitely", t.Dir())
	t.timerData.Paused = false
	go func() {
		defer close(t.stopto)
		defer close(wakup)
		defer close(t.pause)
		defer close(t.unpause)

		for {
			select {
			case <-t.pause:
				log.Printf("Timer for project '%s' was paused explicitely", t.Dir())
				t.timerData.Paused = true
			case <-t.unpause:
				log.Printf("Timer for project '%s' was unpaused explicitely", t.Dir())
				t.timerData.Paused = false
			case <-t.stopto:
				log.Printf("Timer for project '%s' was stopped (and paused) explicitely", t.Dir())
				t.timerData.Paused = true
				break
			case merr := <-t.monitor.Errors():
				log.Printf("Monitor Error: %s", merr)
			case <-time.After(t.timerData.Timeout):
				log.Printf("Timer for project '%s' timed out after %s", t.Dir(), t.timerData.Timeout)
				t.timerData.Paused = true
			case ev := <-wakup:
				if t.IsPaused() {
					log.Printf("Timer for project '%s' woke up after some activity in '%s'", t.Dir(), ev.Dir())
					t.timerData.Paused = false
				} else {
					log.Printf("Timer saw activity for project '%s' in '%s' but is already unpaused", t.Dir(), ev.Dir())
				}
			}
		}
	}()

	//handle time modifications here
	go func() {
		defer close(t.reset)
		defer close(t.stoptick)

		for {
			if !t.timerData.Paused {
				t.timerData.Time += t.timerData.MBU
			}

			select {
			case <-t.stoptick:
				break
			case <-t.reset:
				t.timerData.Time = 0
				log.Printf("Timer for project '%s' was reset explicitely", t.Dir())
			case <-time.After(t.timerData.MBU):
			}
		}
	}()

	return nil
}

func (t *Timer) Pause() {
	t.pause <- struct{}{}
}

func (t *Timer) Unpause() {
	t.unpause <- struct{}{}
}

func (t *Timer) Reset() {
	t.reset <- struct{}{}
}

func (t *Timer) Stop() error {
	t.stopto <- struct{}{}
	t.stoptick <- struct{}{}
	err := t.monitor.Stop()
	if err != nil {
		return errwrap.Wrapf("Failed to stop monitor: {{err}}", err)
	}

	return nil
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
