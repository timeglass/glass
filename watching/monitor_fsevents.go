// +build darwin

package watching

import (
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/timeglass/glass/_vendor/github.com/ronbu/fsevents"
)

type Monitor struct {
	*monitor
	stream *fsevents.Stream
}

func NewMonitor(dir string) (*Monitor, error) {
	mon := newMonitor(dir)

	return &Monitor{
		monitor: mon,
		stream:  fsevents.New(0, fsevents.NOW, time.Millisecond*50, fsevents.CF_NODEFER|fsevents.CF_FILEEVENTS, dir),
	}, nil
}

func transFsEventOp(ev fsevents.Event) ([]int, *MonitorError) {
	res := []int{}
	if ev.Flags&fsevents.EF_MODIFIED == fsevents.EF_MODIFIED {
		res = append(res, Modify)
	}

	if ev.Flags&fsevents.EF_CREATED == fsevents.EF_CREATED {
		res = append(res, Create)
	}

	if ev.Flags&fsevents.EF_REMOVED == fsevents.EF_REMOVED {
		res = append(res, Remove)
	}

	if ev.Flags&fsevents.EF_RENAMED == fsevents.EF_RENAMED {
		res = append(res, Rename)
	}

	if len(res) == 0 {
		return res, NewMonitorError(errors.New("No known OP for event"), fmt.Sprintf("for: %s", ev))
	}

	return res, nil
}

func (m *Monitor) Start() error {

	//tunnel events to our Domain Event Stream
	go func() {
		for {
			select {
			case evs := <-m.stream.Chan:
				for _, ev := range evs {

					//get domain specific operation
					op, merr := transFsEventOp(ev)
					if merr != nil {
						m.Throw(merr)
					}

					//emit the directory event, normalize
					m.Emit(NewMonitorEvent(
						m.dir,
						filepath.Dir(ev.Path),
						ev.Path,
						op))
				}
			}
		}
	}()

	res := m.stream.Start()
	if res != true {
		return errors.New("Failed to start fsevent runloop, no information available")
	}

	return nil
}

func (m *Monitor) Stop() error {
	m.stream.Stop()
	//@todo close channels?
	return nil
}
