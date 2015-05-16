package watching

import (
	"github.com/advanderveer/timer/daemon/watching/event"
)

//
// Dir event leaves the consumer
// with rescanning given directories
// for the actual changes, cross platform
// changes on the file level cannot be guaranteed
//
type DirEvent interface {
	event.Event
	Directory() string
	Operations() []int
}

//
// Watcher interface specifies
// an watcher that returns a channel of
// watching events when started
//
type Watcher interface {
	//watcher emits events
	event.Emitter

	//
	// Start watching the provided directory
	//
	Start() error

	//
	// return the (root) directory we are watching
	//
	Directory() string

	//
	// Stop watching the directory
	//
	Stop() error
}

//
// Unspecified generic error
//
type GenericError struct {
	err         error
	description string
}

func (e *GenericError) Error() error {
	return e.err
}

func (e *GenericError) Describe() string {
	return e.description
}
