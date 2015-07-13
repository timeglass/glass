package timer

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/timeglass/glass/_vendor/github.com/hashicorp/errwrap"

	"github.com/timeglass/glass/config"
	"github.com/timeglass/snow/index"
	"github.com/timeglass/snow/monitor"
)

// a timer without explicit start
// and stop methods, let garbage collector
// remove any routines, channels etc when
// a refence is removed
type Timer struct {
	err     error
	running bool
	read    chan chan time.Duration
	marshal chan chan []byte
	unpause chan struct{}
	save    chan struct{}
	pause   chan struct{}
	reset   chan struct{}
	stop    chan chan struct{}
	monitor monitor.M
	index   index.I

	timerData *timerData
}

type timerData struct {
	Dir         string        `json:"conf_path"`
	Time        time.Duration `json:"total"`
	MBU         time.Duration `json:"mbu"`
	Timeout     time.Duration `json:"timeout"`
	Paused      bool          `json:"paused"`
	Latency     time.Duration `json:"latency"`
	Distributor *Distributor  `json:"distributor"`
}

func NewTimer(dir string) (*Timer, error) {
	t := &Timer{
		timerData: &timerData{
			Dir: dir,
		},
	}

	err := t.init()
	if err != nil {
		return nil, err
	}

	t.Start()
	return t, nil
}

func (t *Timer) init() error {
	t.read = make(chan chan time.Duration)
	t.marshal = make(chan chan []byte)
	t.unpause = make(chan struct{})
	t.pause = make(chan struct{})
	t.stop = make(chan chan struct{})
	t.reset = make(chan struct{})

	sysdir, err := SystemTimeglassPathCreateIfNotExist()
	if err != nil {
		return errwrap.Wrapf(fmt.Sprintf("Failed to read system config: {{err}}"), err)
	}

	conf, err := config.ReadConfig(t.Dir(), sysdir)
	if err != nil {
		return errwrap.Wrapf(fmt.Sprintf("Failed to read configuration for '%s': {{err}}, using default", t.Dir()), err)
	}

	if t.timerData.MBU == 0 {
		t.timerData.MBU = time.Duration(conf.MBU)
	}

	if t.timerData.Timeout == 0 {
		t.timerData.Timeout = t.timerData.MBU * 4
	}

	if t.timerData.Latency == 0 {
		t.timerData.Latency = time.Millisecond * 50
	}

	t.monitor, err = monitor.New(t.Dir(), monitor.Recursive, t.timerData.Latency)
	if err != nil {
		return errwrap.Wrapf(fmt.Sprintf("Failed to create monitor for '%s': {{err}}", t.Dir()), err)
	}

	t.index = index.NewLazy()
	if t.timerData.Distributor == nil {
		t.timerData.Distributor = NewDistributor()
	}

	return nil
}

func (t *Timer) emitSave() {
	if t.save != nil {
		t.save <- struct{}{}
	}
}

func (t *Timer) Start() {
	fevs, ierrs := t.index.Pipe(t.monitor.Events())
	_, err := t.monitor.Start()
	if err != nil {
		log.Printf("Failed to start monitor for '%s', this prevents automatic unpausing: %s", t.Dir(), err)
	}

	//timeout routine
	extend := make(chan struct{})
	tostop := make(chan struct{})
	go func(to time.Duration) {
		for {
			select {
			case <-extend:
				log.Print("Extend")
			case <-time.After(to):
				t.pause <- struct{}{}
			case <-tostop:
				return
			}
		}
	}(t.timerData.Timeout)

	//data manipulation routine
	go func(d *timerData) {
		ticker := time.NewTicker(d.MBU)
		d.Time += d.MBU
		for {
			select {
			case ch := <-t.read:
				ch <- d.Time
			case <-t.unpause:
				extend <- struct{}{}
				if d.Paused == true {
					log.Printf("[%s] Unpaused", d.Dir)
				}

				d.Paused = false
			case <-t.pause:
				if d.Paused == false {
					log.Printf("[%s] Pause", d.Dir)
				}

				d.Distributor.Break()
				d.Paused = true
			case <-t.reset:
				d.Time = 0
				d.Distributor.Reset()
			case <-ticker.C:
				if !d.Paused {
					d.Time += d.MBU
					log.Printf("[%s] Tick: %s", d.Dir, d.Time)
					d.Distributor.Distribute(d.MBU, time.Now())

					t.emitSave()
				}

			case fev := <-fevs:
				log.Printf("[%s] File system activity in '%s'", d.Dir, fev.Dir())
				extend <- struct{}{}
				d.Distributor.Register(fev.Path())
				d.Paused = false
			case ierr := <-ierrs:
				log.Printf("[%s] Index error: %s", d.Dir, ierr)
			case merr := <-t.monitor.Errors():
				log.Printf("[%s] Monitor error: %s", d.Dir, merr)
			case ch := <-t.marshal:
				bytes, err := t.marshalJSON()
				if err != nil {
					log.Printf("[%s] Failed to marshal JSON: %s", d.Dir, err)
				}

				ch <- bytes
			case ch := <-t.stop:
				tostop <- struct{}{}
				ticker.Stop()

				//stop index
				t.index.Stop()

				//@todo Remove this at some point, it normalizes
				//time after rapid stop start usage on OSX
				//for unkown reasons, long term solution should probably
				//involve some mechanism that prevents the darwin monitor
				//form stopping to quickly after being started
				<-time.After(time.Millisecond)
				err := t.monitor.Stop()
				if err != nil {
					log.Printf("Failed to stop monitor for '%s': %s", t.Dir(), err)
				}

				ch <- struct{}{}
				return
			}
		}
	}(t.timerData)

	t.running = true
}

func (t *Timer) Distributor() *Distributor {
	return t.timerData.Distributor
}

func (t *Timer) Unpause() {
	if !t.running {
		t.timerData.Paused = false
		return
	}

	t.unpause <- struct{}{}
}

func (t *Timer) Pause() {
	if !t.running {
		t.timerData.Paused = true
		return
	}

	t.pause <- struct{}{}
}

func (t *Timer) Reset() {
	if !t.running {
		t.timerData.Time = 0
		return
	}

	t.reset <- struct{}{}
}

func (t *Timer) Stop() {
	if !t.running {
		return
	}

	ch := make(chan struct{})
	t.stop <- ch
	<-ch
	t.running = false
}

func (t *Timer) Error() error { return t.err }

func (t *Timer) IsPaused() bool { return t.timerData.Paused }

func (t *Timer) Dir() string {
	return t.timerData.Dir
}

func (t *Timer) Time() time.Duration {
	if !t.running {
		return t.timerData.Time
	}

	ch := make(chan time.Duration)
	t.read <- ch
	return <-ch
}

func (t *Timer) SetSave(ch chan struct{}) {
	t.save = ch
}

func (t *Timer) UnmarshalJSON(b []byte) error {
	err := json.Unmarshal(b, &t.timerData)
	if err != nil {
		return err
	}

	t.init()
	return nil
}

func (t *Timer) marshalJSON() ([]byte, error) {
	return json.Marshal(t.timerData)
}

func (t *Timer) MarshalJSON() ([]byte, error) {
	if !t.running {
		return t.marshalJSON()
	}

	ch := make(chan []byte)
	t.marshal <- ch
	return <-ch, nil
}
