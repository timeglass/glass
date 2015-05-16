package watching

import (
	"fmt"
	"github.com/advanderveer/timer/daemon/watching/event"
	"path/filepath"
	"time"
)

const (
	Create = 1
	Modify = 2
	Remove = 3
	Rename = 4
)

//
// Errors emitted by the monitor
// concrete implementation
//
type MonitorError struct {
	err     error
	context string
	time    time.Time
}

func NewMonitorError(err error, ctx string) *MonitorError {
	return &MonitorError{
		err:     err,
		context: ctx,
		time:    time.Now(),
	}
}

func (e *MonitorError) Describe() string {
	return fmt.Sprintf("An error occured while handling monitoring events: %s, %s", e.err.Error(), e.context)
}

func (e *MonitorError) Error() error {
	return e.err
}

func (e *MonitorError) Time() time.Time {
	return e.time
}

//
// Events emitted by the monitor
// concrete implementation
//
type MonitorEvent struct {
	*event.TimedNamed
	r    string
	f    string
	dir  string
	ops  []int
	time time.Time
}

func NewMonitorEvent(root, dir, file string, ops []int) *MonitorEvent {
	ev := &MonitorEvent{
		TimedNamed: event.NewTimedNamed("watching.directory", time.Now()),
		r:          root,
		f:          file,
		dir:        dir,
		ops:        ops,
	}

	//setting this on construction is not really accurate but
	//good enough
	ev.happened()
	return ev
}

func (e *MonitorEvent) Operations() []int {
	return e.ops
}

func (e *MonitorEvent) Directory() string {
	return e.dir
}

func (e *MonitorEvent) Describe() string {
	ops := []string{}
	for _, op := range e.ops {
		switch op {
		case 1:
			ops = append(ops, "Create")
		case 2:
			ops = append(ops, "Modify")
		case 3:
			ops = append(ops, "Remove")
		case 4:
			ops = append(ops, "Rename")
		}
	}

	return fmt.Sprintf("%s happened on %s", ops, e.rel())
}

//
// Abstract for a monitor
//
type monitor struct {
	dir string
	*event.Outlet
}

func newMonitor(dir string) *monitor {
	return &monitor{
		dir:    dir,
		Outlet: event.NewEOutlet(),
	}
}

func (m *Monitor) Directory() string {
	return m.dir
}

//
// UNEXPORTED
// Methods for more indepth manipulation and inspection
//
func (e *MonitorEvent) file() string {
	return e.f
}

func (e *MonitorEvent) root() string {
	return e.r
}

func (e *MonitorEvent) relDir() string {
	path, _ := filepath.Rel(e.r, e.Directory())
	return path
}

func (e *MonitorEvent) rel() string {
	path, _ := filepath.Rel(e.r, e.file())
	return path
}

func (e *MonitorEvent) happened() {
	e.TimedNamed = event.NewTimedNamed(e.Name(), time.Now())
}
