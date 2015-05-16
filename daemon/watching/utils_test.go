package watching

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TempDir(t *testing.T) (string, func()) {
	path, err := ioutil.TempDir("", fmt.Sprint("timeglass"))
	path, _ = filepath.EvalSymlinks(path)
	if err != nil {
		t.Error(err)
	}
	return path, func() {
		err := os.RemoveAll(path)
		if err != nil {
			t.Error(err)
		}
	}
}

func DTempDir(t *testing.T, d time.Duration) (string, func()) {
	defer time.Sleep(d)
	return TempDir(t)
}

type EventSequence []DirEvent

func GatherErrors(t *testing.T, errs chan Error) {
	for err := range errs {
		t.Error(err.Describe())
	}
}

func WaitForNEvents(t *testing.T, evs chan DirEvent, nr int, to time.Duration) EventSequence {
	seq := []DirEvent{}
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
	AssertNthEvent(t, seq, idx, func(ev DirEvent) (bool, string) {
		msg := fmt.Sprintf("Event %d name was %s expected %s", idx, ev.Name(), name)
		return (ev.Name() == name), msg
	})
}

func AssertNthEvent(t *testing.T, seq EventSequence, idx int, ass func(ev DirEvent) (bool, string)) {
	if len(seq) < idx+1 {
		t.Errorf("Asserting event at idx %d while only %d are in sequence", idx, seq)
		return
	}

	ok, msg := ass(seq[idx])
	if !ok {
		t.Errorf("Asserting Event at %d Failed: %s", idx, msg)
	}
}

func NoMoreEvents(t *testing.T, evs chan DirEvent, to time.Duration) {
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
