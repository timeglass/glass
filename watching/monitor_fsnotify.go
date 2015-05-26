// +build !darwin

package watching

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/timeglass/glass/_vendor/github.com/go-fsnotify/fsnotify"
)

type Monitor struct {
	*monitor
	watcher *fsnotify.Watcher
	latency time.Duration
	closed  bool
}

func NewMonitor(dir string) (*Monitor, error) {
	mon := newMonitor(dir)
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &Monitor{
		monitor: mon,
		closed:  false,
		watcher: watcher,
		latency: time.Millisecond * 50,
	}, nil
}

func transFsNotifyOp(ev fsnotify.Event) (int, *MonitorError) {
	res := 0
	if ev.Op&fsnotify.Write == fsnotify.Write {
		res = Modify
	}

	//fsnotify does not exibit double operations so
	//clearly indicate events that have two operations
	checkDouble := func() *MonitorError {
		if res != 0 {
			return NewMonitorError(errors.New("Double operation"), fmt.Sprintf("for: %s", ev))
		}
		return nil
	}

	if ev.Op&fsnotify.Create == fsnotify.Create {
		err := checkDouble()
		if err != nil {
			return 0, err
		}
		res = Create
	}

	if ev.Op&fsnotify.Remove == fsnotify.Remove {
		err := checkDouble()
		if err != nil {
			return 0, err
		}
		res = Remove
	}

	if ev.Op&fsnotify.Rename == fsnotify.Rename {
		err := checkDouble()
		if err != nil {
			return 0, err
		}
		res = Rename
	}

	if res == 0 {
		return 0, NewMonitorError(errors.New("Uknown event OP"), fmt.Sprintf("for: %s", ev))
	}

	return res, nil
}

func (m *Monitor) Start() error {
	//lazily create new fsnotify watcher if the previous one was
	//closed in the meantime
	var err error
	if m.closed {
		m.watcher, err = fsnotify.NewWatcher()
		if err != nil {
			return err
		}
	}

	//events scheduled for emission
	var scheduled []*MonitorEvent

	//handle watcher events
	go func() {
	F:
		for ev := range m.watcher.Events {

			//transform to domain specific operation
			op, merr := transFsNotifyOp(ev)
			if merr != nil {
				m.Throw(merr)
			}

			//creat resulting event
			res := NewMonitorEvent(m.dir, filepath.Dir(ev.Name), ev.Name, []int{op})

			//some filtering based on fileinfo
			fi, err := os.Stat(ev.Name)
			if err != nil {
				// dont bother with writes on nonexsisting files
				if ev.Op&fsnotify.Write == fsnotify.Write {
					continue //skip
				}

			} else {

				//dont bother with dir modifies
				if ev.Op&fsnotify.Write == fsnotify.Write && fi.IsDir() {
					continue //skip
				}

				//when a dir is added, start watching
				if ev.Op&fsnotify.Create == fsnotify.Create && fi.IsDir() {
					err := m.watcher.Add(ev.Name)
					if err != nil {
						m.Throw(NewMonitorError(err, fmt.Sprintf("tried to start watch on dir creation of: %s", ev.Name)))
						continue //skip
					}
				}
			}

			//discard events based on latency
			for _, past := range scheduled {
				diff := res.Time().Sub(past.Time())
				if diff < m.latency && past.file() == res.file() {
					past.ops = append(past.ops, res.ops[0:]...)
					continue F //skip
				}
			}

			//instead of emitting it immediately schedule it for
			//removal so ops can be merged from later events
			scheduled = append(scheduled, res)
			go func(r *MonitorEvent) {
				time.Sleep(m.latency)
				m.Emit(r)
			}(res)

		}
	}()

	//handle watcher errors
	go func() {
		for err := range m.watcher.Errors {
			m.Throw(NewMonitorError(err, fmt.Sprintf("root directory: %s", m.dir)))
		}
	}()

	//recursive watch
	err = filepath.Walk(m.dir, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			//@todo handle channel close?
			return err
		}

		if fi.IsDir() {
			err := m.watcher.Add(path)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	//watch root directory itself
	err = m.watcher.Add(m.dir)
	if err != nil {
		return err
	}

	return nil
}

func (m *Monitor) Stop() error {
	err := m.watcher.Close()
	if err != nil {
		return err
	}

	m.closed = true
	return nil
}
