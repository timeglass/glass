package event

import (
	"time"
)

type Moment interface {
	Time() time.Time
}

type Describer interface {
	Describe() string
}

//
// A Generic interface that represents
// a significant named error in time with a
// human readable description
//
type Error interface {
	Describer
	Moment
	Error() error
}

//
// A Generic interface that represents
// a significant named event in time with a
// human readable description
//
type Event interface {
	Describer
	Moment
	Name() string
}

// a abstract event implementation
// that holds a time and a name
type TimedNamed struct {
	name string
	time time.Time
}

func NewTimedNamed(name string, t time.Time) *TimedNamed {
	return &TimedNamed{
		name: name,
		time: t,
	}
}

func (tn *TimedNamed) Name() string    { return tn.name }
func (tn *TimedNamed) Time() time.Time { return tn.time }

// allows for sorting a slice of events by time
type ByTime []Event

func (a ByTime) Len() int           { return len(a) }
func (a ByTime) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByTime) Less(i, j int) bool { return a[i].Time().Before(a[j].Time()) }

//
// Stores event and errors and
// allow for listing them
//
type Store interface {
	ErrorList() []Error
	EventList() []Event

	//returns a single list that contains
	//errors as events
	ErrorsAsEvents() []Event

	Serialize() (string, error)
}

//
// An emitter exposes two
// channels that contain events
// and errors
type Emitter interface {
	Errors() chan Error
	Events() chan Event
	Emit(Event)
	Throw(Error)
}

//
// Outlet a very simple emitter
// implementation that can easily be extended by
// other components
type Outlet struct {
	events chan Event
	errors chan Error
}

//factory that needs the channes to be provided
func NewOutlet(evs chan Event, errs chan Error) *Outlet {
	return &Outlet{
		events: evs,
		errors: errs,
	}
}

//creates new channels for outlet itself
func NewEOutlet() *Outlet {
	return NewOutlet(make(chan Event), make(chan Error))
}

func (o *Outlet) Throw(err Error)    { o.errors <- err }
func (o *Outlet) Emit(ev Event)      { o.events <- ev }
func (o *Outlet) Events() chan Event { return o.events }
func (o *Outlet) Errors() chan Error { return o.errors }

//
// A listener takes (<-) events,errors
// from N emitters
//
type Listener interface {
	Sources() []Emitter
	Source(emitters ...Emitter)
}

// abstract *no-op* handler that takes
// N Emitters, and has a configuarble handle
type Handler struct {
	sources []Emitter
	handle  func(evs chan Event, errs chan Error)
}

func NewHandler() *Handler {
	return &Handler{
		sources: []Emitter{},
		handle:  func(evs chan Event, errs chan Error) {},
	}
}

func (h *Handler) Handler(hn func(evs chan Event, errs chan Error)) { h.handle = hn }
func (h *Handler) Sources() []Emitter                               { return h.sources }

func (h *Handler) Source(emitters ...Emitter) {
	for _, e := range emitters {
		h.sources = append(h.sources, e)
		h.handle(e.Events(), e.Errors())
	}
}

//
// A dispatcher takes events (<-) but
// dispatches (multiplies) them across all
// outlets. outlets are emitters themselves
//
type Dispatcher interface {
	Listener
	Outlet() Emitter
}

// abstract outlet handling that
// can be used by oher modules to satisfy
// the dispatcher interface
type Outletter struct {
	outlets []Emitter
}

func NewOutletter() *Outletter {
	return &Outletter{
		outlets: []Emitter{},
	}
}

func (o *Outletter) Throw(err Error) {
	for _, ot := range o.outlets {
		ot.Throw(err)
	}
}

func (o *Outletter) Emit(ev Event) {
	for _, ot := range o.outlets {
		ot.Emit(ev)
	}
}

func (o *Outletter) Outlet() Emitter {
	ot := NewEOutlet()
	o.outlets = append(o.outlets, ot)
	return ot
}

//
// A modifier both takes in events
// from one ore multipe sources and
// emits new or modified events
//
type Modifier interface {
	Listener
	Emitter
}

//
// Whenever we want to treat
// errors as events we can use this
// concrete implementation
//
type ErrorEvent struct {
	Error
	time time.Time
}

func NewErrorEvent(err Error) *ErrorEvent {
	return &ErrorEvent{
		Error: err,
		time:  err.Time(),
	}
}

func (ev *ErrorEvent) Name() string {
	return "error"
}

func (ev *ErrorEvent) Time() time.Time {
	return ev.time
}
