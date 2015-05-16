package event

import (
	"fmt"
	"testing"
	"time"
)

type EventSequence []Event

func GatherErrors(t *testing.T, errs chan Error) {
	for err := range errs {
		t.Error(err.Describe())
	}
}

func WaitForNEvents(t *testing.T, evs chan Event, nr int, to time.Duration) EventSequence {
	seq := []Event{}
L:
	for {
		select {
		case ev := <-evs:
			seq = append(seq, ev)
			if len(seq) == nr {
				break L
			}
		case <-time.After(to):
			t.Error(fmt.Sprintf("Waiting for %dth event, but timed out after the %dth (%s)", nr, len(seq), to))
			break L
		}
	}

	return seq
}

func AssertNthEventName(t *testing.T, seq EventSequence, idx int, name string) {
	AssertNthEvent(t, seq, idx, func(ev Event) (bool, string) {
		msg := fmt.Sprintf("Event %d name was %s expected %s", idx, ev.Name(), name)
		return (ev.Name() == name), msg
	})
}

func AssertNthEvent(t *testing.T, seq EventSequence, idx int, ass func(ev Event) (bool, string)) {
	if len(seq) < idx+1 {
		t.Errorf("Asserting event at idx %d while only %d are in sequence", idx, seq)
		return
	}

	ok, msg := ass(seq[idx])
	if !ok {
		t.Errorf("Asserting Event at %d Failed: %s", idx, msg)
	}
}

func NoMoreEvents(t *testing.T, evs chan Event, to time.Duration) {
L:
	for {
		select {
		case ev := <-evs:
			fmt.Println(ev.Describe(), " <- !")
			t.Errorf("No more evs, received: %s", ev.Describe())
			break L
		case <-time.After(to):
			break L
		}
	}
}
