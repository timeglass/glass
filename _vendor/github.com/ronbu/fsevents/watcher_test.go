package fsevents

import (
	"io/ioutil"
	//"log"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"testing"
)
import "github.com/couchbaselabs/go.assert"

import (
	"os"
	"time"
)

func TestCurrent(t *testing.T) {
	t.Parallel()
	id1 := Current()
	id2 := Current()
	assert.True(t, id1 == id2)
}

func TestLastEventBefore(t *testing.T) {
	t.Parallel()
	base, rm := TempDir()
	defer rm()

	fi, _ := os.Stat(base)
	dev := Device(fi.Sys().(*syscall.Stat_t).Dev)
	id := LastEventBefore(dev, time.Now())
	assert.True(t, id != 0)
}

func TestNew(t *testing.T) {
	t.Parallel()
	base, rm := TempDir()
	defer rm()
	stream := New(
		0,
		NOW,
		time.Millisecond*50,
		CF_NODEFER|CF_FILEEVENTS,
		base)
	assert.True(t, stream.Chan != nil)
}

func TestStreamPaths(t *testing.T) {
	t.Parallel()
	base, rm := TempDir()
	defer rm()
	stream := New(
		0,
		NOW,
		time.Millisecond*50,
		CF_NODEFER|CF_FILEEVENTS,
		base)
	path := stream.Paths()[0]
	assert.True(t, path == base)
}

func TestCreateRelativeToDevice(t *testing.T) {
	t.Parallel()
	base, rm := TempDir()
	defer rm()

	fi, _ := os.Stat(base)
	dev := Device(fi.Sys().(*syscall.Stat_t).Dev)

	stream := New(
		dev,
		NOW,
		time.Millisecond*50,
		CF_NODEFER|CF_FILEEVENTS,
		base)
	assert.True(t, stream.Chan != nil)
}

func TestFlushAsync(t *testing.T) {
	t.Parallel()
	base, rm := TempDir()
	defer rm()
	stream := New(
		0,
		NOW,
		time.Millisecond*50,
		CF_NODEFER|CF_FILEEVENTS,
		base)
	stream.Start()
	ioutil.WriteFile(base+"/holla", []byte{}, 777)
	time.Sleep(time.Millisecond * 50)
	event := stream.FlushAsync()
	assert.True(t, event != 0)
}

func TestFlush(t *testing.T) {
	t.Parallel()
	base, rm := TempDir()
	defer rm()
	stream := New(
		0,
		NOW,
		time.Millisecond*50,
		CF_NODEFER|CF_FILEEVENTS,
		base)
	stream.Start()
	stream.Flush()
}

func TestStreamDevice(t *testing.T) {
	t.Parallel()
	base, rm := TempDir()
	defer rm()

	fi, _ := os.Stat(base)
	dev := Device(fi.Sys().(*syscall.Stat_t).Dev)

	stream := New(
		dev,
		NOW,
		time.Millisecond*50,
		CF_NODEFER|CF_FILEEVENTS,
		base)

	adev := stream.Device()
	assert.True(t, dev == adev)
}

func TestStart(t *testing.T) {
	t.Parallel()
	base, rm := TempDir()
	defer rm()

	stream := New(0, NOW, time.Millisecond*50, CF_NODEFER|CF_FILEEVENTS, base)
	ok := stream.Start()
	if ok != true {
		t.Fatal("failed to start the stream")
	}
}

func withNew(base string, action func(string)) {
	dummyfile := "dummyfile.txt"
	os.Create(filepath.Join(base, dummyfile))

	action(filepath.Join(base, dummyfile))

	os.Remove(filepath.Join(base, dummyfile))
}

func TestFileChanges(t *testing.T) {
	t.Parallel()
	base, rm := TempDir()
	defer rm()

	s := New(0, NOW, time.Second/10, CF_FILEEVENTS, base)
	s.Start()
	defer s.Close()

	withNew(base, func(dummyfile string) {
		select {
		case <-s.Chan:
		case <-time.After(time.Minute):
			t.Errorf("should have got some file event, but timed out")
		}
	})
}

func TestEventFlags(t *testing.T) {
	t.Parallel()
	base, rm := TempDir()
	defer rm()

	s := New(0, NOW, time.Second/10, CF_FILEEVENTS, base)
	s.Start()
	defer s.Close()

	withNew(base, func(dummyfile string) {
		select {
		case events := <-s.Chan:
			events = getEvents(base, events)
			assert.Equals(t, len(events), 1)

			assert.True(t, events[0].Flags&EF_CREATED != 0)
		case <-time.After(time.Minute):
			t.Errorf("should have got some file event, but timed out")
		}
	})
}

func TestCanGetPath(t *testing.T) {
	t.Parallel()
	base, rm := TempDir()
	defer rm()

	s := New(0, NOW, time.Second/10, CF_FILEEVENTS, base)
	s.Start()
	defer s.Close()

	withNew(base, func(dummyfile string) {
		select {
		case events := <-s.Chan:
			events = getEvents(base, events)
			assert.Equals(t, len(events), 1)

			fullpath, _ := filepath.Abs(dummyfile)
			fullpath, _ = filepath.EvalSymlinks(fullpath)
			evPath, _ := filepath.EvalSymlinks(events[0].Path)
			assert.Equals(t, evPath, fullpath)
		case <-time.After(time.Minute):
			t.Errorf("timed out")
		}
	})
}

func TestOnlyWatchesSpecifiedPaths(t *testing.T) {
	t.Parallel()
	base, rm := TempDir()
	defer rm()

	s := New(0, NOW, time.Second/10, CF_FILEEVENTS,
		filepath.Join(base, "imaginaryfile"))
	s.Start()
	defer s.Close()

	withNew(base, func(dummyfile string) {
		select {
		case evs := <-s.Chan:
			t.Errorf("should have timed out, but received:%v", evs)
		case <-time.After(time.Millisecond * 200):
		}
	})
}

func TestCanUnwatch(t *testing.T) {
	t.Parallel()
	base, rm := TempDir()
	defer rm()

	s := New(0, NOW, time.Second/10, CF_FILEEVENTS, base)
	s.Start()
	s.Close()

	withNew(base, func(dummyfile string) {
		select {
		case evs, ok := <-s.Chan:
			evs = getEvents(base, evs)
			if ok && len(evs) > 0 {
				t.Errorf("should have timed out, but received: %#v", evs)
			}
		case <-time.After(time.Millisecond * 200):
		}
	})
}

func TestMultipleFile(t *testing.T) {
	t.Parallel()
	base, rm := TempDir()
	defer rm()

	s := New(
		0, NOW, time.Second/10, CF_FILEEVENTS, base)
	s.Start()
	defer s.Close()

	files := []string{"holla", "huhu", "heeeho", "haha"}
	for _, f := range files {
		ioutil.WriteFile(filepath.Join(base, f), []byte{12, 32}, 0777)
	}

	events := []string{}
LOOP:
	for {
		select {
		case e := <-s.Chan:
			e = getEvents(base, e)
			for _, item := range e {
				p, _ := filepath.Rel(base, item.Path)
				events = append(events, p)
			}
			if len(events) == len(files) {
				break LOOP
			}
		case <-time.After(time.Minute):
			break LOOP
		}
	}

	assert.Equals(t, strings.Join(events, " "), strings.Join(files, " "))
}

func getEvents(base string, in []Event) (out []Event) {
	for _, e := range in {
		if e.Path != base {
			out = append(out, e)
		}
	}
	return
}

func TempDir() (string, func()) {
	path, _ := ioutil.TempDir("", "fsevents")
	path, _ = filepath.EvalSymlinks(path)
	return path, func() {
		os.RemoveAll(path)
	}
}

// Create 10 folders with 10 files each, all under one top-level folder,
// for a total of 111 events.
func with100Files(base string, action func(base string)) {
	for i := 0; i < 10; i++ {
		dir := filepath.Join(base, strconv.Itoa(i)+".dir")
		os.Mkdir(dir, 0755)
		for j := 0; j < 10; j++ {
			os.Create(filepath.Join(dir, "dummy"+strconv.Itoa(j)+".txt"))
		}
	}
	action(base)
}

func Test100Files(t *testing.T) {
	t.Parallel()
	base, rm := TempDir()
	defer rm()

	s := New(0, NOW, time.Second/10, CF_FILEEVENTS, base)
	s.Start()
	defer s.Close()

	count := 0
	with100Files(base, func(base string) {
		for {
			select {
			case events := <-s.Chan:
				for _, e := range events {
					count++
					_ = e
					//log.Println("a)", count, e)
					if count >= 111 {
						return
					}
				}
			case <-time.After(time.Second * 10):
				t.Errorf("should have got received 111 file events, but timed out")
				return
			}
		}
	})
}

func Test100OldFiles(t *testing.T) {
	t.Parallel()
	base, rm := TempDir()
	defer rm()

	with100Files(base, func(base string) {})

	s := New(0, ALL, time.Second/10, CF_FILEEVENTS, base)
	s.Start()
	defer s.Close()

	count := 0
	for {
		select {
		case events := <-s.Chan:
			for _, e := range events {
				count++
				_ = e
				//log.Println("a)", count, e)
				if count >= 111 {
					return
				}
			}
		case <-time.After(time.Second * 10):
			t.Errorf("should have got received 111 file events, but timed out")
			return
		}
	}
}
