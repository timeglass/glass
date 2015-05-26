package watching

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var fsto = time.Millisecond * 30 //how long we wait for the file system to stabalize after setup
var to = time.Millisecond * 200  //how long it takes for the actual captured event to appear
var noto = time.Millisecond * 80 //how long after the last event we check wether we received no more

//
// Domain specific assertion
//

func assertNthMonitorEvent(t *testing.T, seq EventSequence, idx int, assertDir string, op int) {
	AssertNthEvent(t, seq, idx, func(ev DirEvent) (bool, string) {

		dev, ok := seq[idx].(*MonitorEvent)
		if !ok {
			return false, fmt.Sprintf("Could not type assert %s to MonitorEvent", seq[idx])
		}

		//see if the required op is in the event
		found := false
		if op == 0 {
			found = true
		}
		available := []int{}
		for _, o := range dev.Operations() {
			available = append(available, o)
			if o == op {
				found = true
				break
			}
		}

		if !found {
			return false, fmt.Sprintf("Didn't receive op: %d, event has: %s", op, available)
		}

		//make sure there is a sensible time
		if dev.Time().UnixNano() < 0 {
			return false, fmt.Sprintf("Event time was not set, time: %s", dev.Time())
		}

		asserts := strings.Split(assertDir, ",")
		relPaths := []string{}
		found = false
		for _, assertDir := range asserts {
			path, _ := filepath.Rel(dev.root(), assertDir)
			relPaths = append(relPaths, path)
			if dev.Directory() == assertDir {
				found = true
				break
			}
		}

		if !found {
			return false, fmt.Sprintf("Expected event in dir(s) %s, instead happend in dir %s", relPaths, dev.relDir())
		}

		return true, ""
	})
}

//
// Actual tests
//

func TestMonitorInterfaceCompliance(t *testing.T) {
	dir, rm := DTempDir(t, fsto)
	defer rm()

	//is a watcher
	var w Watcher
	w, err := NewMonitor(dir)
	if err != nil {
		t.Error(err)
	}

	//is a normal event
	var e DirEvent
	e = NewMonitorEvent(dir, dir, dir, []int{1})

	//is a watcher specific dir event
	var de DirEvent
	de = NewMonitorEvent(dir, dir, dir, []int{1})

	_ = w
	_ = e
	_ = de
}

//
// Creation
//

func Test_File_Create(t *testing.T) {
	//diablelogging
	log.SetOutput(ioutil.Discard)

	log.Println("[1]")
	dir, rm := DTempDir(t, fsto)
	defer rm()

	w, err := NewMonitor(dir)
	if err != nil {
		t.Error(err)
	}

	err = w.Start()
	if err != nil {
		t.Error(err)
	}

	defer w.Stop()

	//gather errors
	go GatherErrors(t, w.Errors())

	//create file
	ioutil.WriteFile(filepath.Join(dir, "test"), []byte{38, 38}, 0777)
	seq := WaitForNEvents(t, w.Events(), 1, to)
	AssertNthEventName(t, seq, 0, "watching.directory")
	assertNthMonitorEvent(t, seq, 0, dir, Create)

	NoMoreEvents(t, w.Events(), noto)
	log.Println("[1]")
}

func Test_File_Create_InExistingSubFolder(t *testing.T) {
	log.Println("[2]")
	dir, rm := DTempDir(t, fsto)
	defer rm()

	sub, err := ioutil.TempDir(dir, "subdir")
	if err != nil {
		t.Error(err)
	}

	//wait for the new directory to have been created
	time.Sleep(fsto)

	w, err := NewMonitor(dir)
	if err != nil {
		t.Error(err)
	}

	err = w.Start()
	if err != nil {
		t.Error(err)
	}
	defer w.Stop()

	//gather errors
	go GatherErrors(t, w.Errors())

	//create file in subdir
	ioutil.WriteFile(filepath.Join(sub, "test"), []byte{38, 38}, 0777)
	seq := WaitForNEvents(t, w.Events(), 1, to)
	AssertNthEventName(t, seq, 0, "watching.directory")
	assertNthMonitorEvent(t, seq, 0, sub, Create)
	NoMoreEvents(t, w.Events(), noto)
	log.Println("[2]")
}

func Test_File_Create_InSubFolder(t *testing.T) {
	log.Println("[3]")
	dir, rm := DTempDir(t, fsto)
	defer rm()

	w, err := NewMonitor(dir)
	if err != nil {
		t.Error(err)
	}

	err = w.Start()
	if err != nil {
		t.Error(err)
	}

	defer w.Stop()

	//gather errors
	go GatherErrors(t, w.Errors())

	//create sub dir
	sub, err := ioutil.TempDir(dir, "subdir")
	if err != nil {
		t.Error(err)
	}

	//
	// windows doesn't seem to be capable of keeping up with fast
	// changes
	//
	time.Sleep(fsto)

	ioutil.WriteFile(filepath.Join(sub, "test"), []byte{38, 38}, 0777)

	seq := WaitForNEvents(t, w.Events(), 2, to)
	AssertNthEventName(t, seq, 0, "watching.directory")
	assertNthMonitorEvent(t, seq, 0, dir, Create)

	//create file in subdir
	AssertNthEventName(t, seq, 1, "watching.directory")
	assertNthMonitorEvent(t, seq, 1, sub, Create)

	NoMoreEvents(t, w.Events(), noto)
	log.Println("[3]")
}

//
// Modification
//

func Test_File_Modify_InExistingSubFolder(t *testing.T) {
	log.Println("[4]")
	dir, rm := DTempDir(t, fsto)
	defer rm()

	//create subdir and subfile
	sub, err := ioutil.TempDir(dir, "subdir")
	ioutil.WriteFile(filepath.Join(sub, "test"), []byte{38, 38}, 0777)
	if err != nil {
		t.Error(err)
	}

	//wait for the new directory to have been created
	time.Sleep(fsto)

	w, err := NewMonitor(dir)
	if err != nil {
		t.Error(err)
	}

	err = w.Start()
	if err != nil {
		t.Error(err)
	}

	defer w.Stop()

	//gather errors
	go GatherErrors(t, w.Errors())

	//modify file in subdir
	ioutil.WriteFile(filepath.Join(sub, "test"), []byte{38, 38}, 0777)
	seq := WaitForNEvents(t, w.Events(), 1, to)
	AssertNthEventName(t, seq, 0, "watching.directory")
	assertNthMonitorEvent(t, seq, 0, sub, Modify)
	NoMoreEvents(t, w.Events(), noto)
	log.Println("[4]")
}

//
// Deletion
//

func Test_File_Delete_InExistingSubFolder(t *testing.T) {
	log.Println("[5]")
	dir, rm := DTempDir(t, fsto)
	defer rm()

	//create subdir and subfile
	sub, err := ioutil.TempDir(dir, "subdir")
	ioutil.WriteFile(filepath.Join(sub, "test"), []byte{38, 38}, 0777)
	if err != nil {
		t.Error(err)
	}

	//wait for the new directory to have been created
	time.Sleep(fsto)

	w, err := NewMonitor(dir)
	if err != nil {
		t.Error(err)
	}

	err = w.Start()
	if err != nil {
		t.Error(err)
	}

	//gather errors
	go GatherErrors(t, w.Errors())

	//remove file in sub folder
	err = os.Remove(filepath.Join(sub, "test"))
	if err != nil {
		t.Error(err)
	}

	seq := WaitForNEvents(t, w.Events(), 1, to)
	AssertNthEventName(t, seq, 0, "watching.directory")
	assertNthMonitorEvent(t, seq, 0, sub, Remove)
	NoMoreEvents(t, w.Events(), noto)
	log.Println("[5]")
}

func Test_Directory_Delete_WithSubFile(t *testing.T) {
	log.Println("[6]")
	dir, rm := DTempDir(t, fsto)
	defer rm()

	//create subdir and subfile
	sub, err := ioutil.TempDir(dir, "subdir")
	ioutil.WriteFile(filepath.Join(sub, "test"), []byte{38, 38}, 0777)
	if err != nil {
		t.Error(err)
	}

	//wait for the new directory to have been created
	time.Sleep(fsto)

	w, err := NewMonitor(dir)
	if err != nil {
		t.Error(err)
	}

	err = w.Start()
	if err != nil {
		t.Error(err)
	}

	defer w.Stop()

	//gather errors
	go GatherErrors(t, w.Errors())

	//remove folder and containering file, emits an event
	//for the file as well as the directory
	err = os.RemoveAll(sub)
	if err != nil {
		t.Error(err)
	}

	//it is undetermined which of the events will happen first
	seq := WaitForNEvents(t, w.Events(), 2, to)
	AssertNthEventName(t, seq, 0, "watching.directory")
	assertNthMonitorEvent(t, seq, 0, fmt.Sprintf("%s,%s", dir, sub), Remove)

	AssertNthEventName(t, seq, 1, "watching.directory")
	assertNthMonitorEvent(t, seq, 1, fmt.Sprintf("%s,%s", dir, sub), Remove)
	NoMoreEvents(t, w.Events(), noto)
	log.Println("[6]")
}

//
// Rename
//

func Test_File_RenameToSameDirectory(t *testing.T) {
	log.Println("[7]")
	dir, rm := DTempDir(t, fsto)
	defer rm()

	//create subdir and subfile
	sub, err := ioutil.TempDir(dir, "subdir")
	ioutil.WriteFile(filepath.Join(sub, "test"), []byte{38, 38}, 0777)
	if err != nil {
		t.Error(err)
	}

	//wait for the new directory to have been created
	time.Sleep(fsto)

	w, err := NewMonitor(dir)
	if err != nil {
		t.Error(err)
	}

	err = w.Start()
	if err != nil {
		t.Error(err)
	}

	defer w.Stop()

	//gather errors
	go GatherErrors(t, w.Errors())

	//rename a single file to something else in the same
	//directory
	err = os.Rename(filepath.Join(sub, "test"), filepath.Join(sub, "test2"))
	if err != nil {
		t.Error(err)
	}

	//first a rename event, then a modify
	seq := WaitForNEvents(t, w.Events(), 2, to)
	AssertNthEventName(t, seq, 0, "watching.directory")
	assertNthMonitorEvent(t, seq, 0, sub, Rename) //"rename" from

	AssertNthEventName(t, seq, 1, "watching.directory")
	assertNthMonitorEvent(t, seq, 1, sub, Rename) //"modify" to
	NoMoreEvents(t, w.Events(), noto)
	log.Println("[7]")
}

func Test_File_RenameToOtherDirectory(t *testing.T) {
	log.Println("[8]")
	dir, rm := DTempDir(t, fsto)
	defer rm()

	//create subdir and subfile
	sub, err := ioutil.TempDir(dir, "subdir")
	ioutil.WriteFile(filepath.Join(sub, "test"), []byte{38, 38}, 0777)
	if err != nil {
		t.Error(err)
	}

	sub2, err := ioutil.TempDir(dir, "subdir2")
	if err != nil {
		t.Error(err)
	}

	//wait for the new directory to have been created
	time.Sleep(fsto)

	w, err := NewMonitor(dir)
	if err != nil {
		t.Error(err)
	}

	err = w.Start()
	if err != nil {
		t.Error(err)
	}

	defer w.Stop()

	//gather errors
	go GatherErrors(t, w.Errors())

	//rename a single file to something else to another
	//directory
	err = os.Rename(filepath.Join(sub, "test"), filepath.Join(sub2, "test2"))
	if err != nil {
		t.Error(err)
	}

	// @todo
	// in this test windows gives Remove and then Create for
	// moment betweewn two different directories while osx gives
	// two rename events

	//it is undeterminend which will happend first
	seq := WaitForNEvents(t, w.Events(), 2, to)
	AssertNthEventName(t, seq, 0, "watching.directory")
	assertNthMonitorEvent(t, seq, 0, fmt.Sprintf("%s,%s", sub2, sub), 0) //"rename" from

	AssertNthEventName(t, seq, 1, "watching.directory")
	assertNthMonitorEvent(t, seq, 1, fmt.Sprintf("%s,%s", sub2, sub), 0) //"modify" to
	NoMoreEvents(t, w.Events(), noto)
	log.Println("[8]")
}

func Test_Directory_RenameToSameDirectory(t *testing.T) {
	log.Println("[9]")
	dir, rm := DTempDir(t, fsto)
	defer rm()

	//create subdir and subfile
	sub, err := ioutil.TempDir(dir, "subdir")
	ioutil.WriteFile(filepath.Join(sub, "test"), []byte{38, 38}, 0777)
	if err != nil {
		t.Error(err)
	}

	//wait for the new directory to have been created
	time.Sleep(fsto)

	w, err := NewMonitor(dir)
	if err != nil {
		t.Error(err)
	}

	err = w.Start()
	if err != nil {
		t.Error(err)
	}

	defer w.Stop()

	//gather errors
	go GatherErrors(t, w.Errors())

	//rename a single file to something else in the same
	//directory
	err = os.Rename(sub, sub+"extra")
	if err != nil {
		t.Error(err)
	}

	seq := WaitForNEvents(t, w.Events(), 2, to)
	AssertNthEventName(t, seq, 0, "watching.directory")
	assertNthMonitorEvent(t, seq, 0, dir, Rename) //"rename" event to new new name

	//dir rename causes a second event with fsevent
	AssertNthEventName(t, seq, 1, "watching.directory")
	assertNthMonitorEvent(t, seq, 1, dir, Rename)

	NoMoreEvents(t, w.Events(), noto)
	log.Println("[9]")
}

func Test_StopMonitoring(t *testing.T) {

	log.Println("[10]")
	dir, rm := DTempDir(t, fsto)
	defer rm()

	w, err := NewMonitor(dir)
	if err != nil {
		t.Error(err)
	}

	err = w.Start()
	if err != nil {
		t.Error(err)
	}

	defer w.Stop()

	//gather errors
	go GatherErrors(t, w.Errors())

	//create file
	ioutil.WriteFile(filepath.Join(dir, "test"), []byte{38, 38}, 0777)
	WaitForNEvents(t, w.Events(), 1, to)

	err = w.Stop()
	if err != nil {
		t.Error(err)
	}

	//cause another event
	ioutil.WriteFile(filepath.Join(dir, "test"), []byte{38, 38}, 0777)

	NoMoreEvents(t, w.Events(), noto)
	log.Println("[10]")

}

//
// This test might fail if the filesystem takes
// to long to react (i.e. my external drive is powering on)
//
func Test_StopStartMonitoring(t *testing.T) {

	log.Println("[11]")
	dir, rm := DTempDir(t, fsto)
	defer rm()

	w, err := NewMonitor(dir)
	if err != nil {
		t.Error(err)
	}

	err = w.Start()
	if err != nil {
		t.Error(err)
	}

	defer w.Stop()

	//gather errors
	go GatherErrors(t, w.Errors())

	//create file
	ioutil.WriteFile(filepath.Join(dir, "test"), []byte{38, 38}, 0777)
	_ = WaitForNEvents(t, w.Events(), 1, to)

	//stop
	err = w.Stop()
	if err != nil {
		t.Error(err)
	}

	//and start again
	w.Start()

	//give monitor time to actually
	time.Sleep(fsto)

	//cause another event
	ioutil.WriteFile(filepath.Join(dir, "test"), []byte{38, 38}, 0777)

	_ = WaitForNEvents(t, w.Events(), 1, to)
	log.Println("[11]")

}
