package timer

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoadSave(t *testing.T) {
	dir, err := ioutil.TempDir("", fmt.Sprintf("glass_keeper"))
	assert.NoError(t, err)

	_, err = ioutil.ReadFile(filepath.Join(dir, "ledger.json"))
	assert.Error(t, err)
	assert.True(t, os.IsNotExist(err))

	k, err := NewKeeper(dir)
	assert.NoError(t, err)
	err = k.Measure(dir)
	assert.NoError(t, err)

	<-time.After(time.Millisecond * 40)

	data, err := ioutil.ReadFile(filepath.Join(dir, "ledger.json"))
	assert.NoError(t, err)
	assert.Contains(t, string(data), dir)

	//load should cause the timer to start
	err = k.Load()
	assert.NoError(t, err)
	assert.True(t, k.keeperData.Timers[dir].running)

	err = k.Save()
	assert.NoError(t, err)
}

func TestStartSave(t *testing.T) {
	dir, err := ioutil.TempDir("", "glass_keeper")
	assert.NoError(t, err)

	k, err := NewKeeper(dir)
	assert.NoError(t, err)

	defer k.Stop()

	<-time.After(time.Millisecond * 40)

	data, err := ioutil.ReadFile(filepath.Join(dir, "ledger.json"))
	assert.NoError(t, err)
	assert.Contains(t, string(data), "timers")
}

func TestRapidAddRemove(t *testing.T) {
	dir, err := ioutil.TempDir("", "glass_keeper")
	assert.NoError(t, err)

	pdir := filepath.Join(dir, "project_x")
	err = os.Mkdir(pdir, 0755)
	assert.NoError(t, err)

	k, err := NewKeeper(dir)
	assert.NoError(t, err)

	defer k.Stop()

	err = k.Measure(pdir)
	assert.NoError(t, err)

	err = k.Discard(pdir)
	assert.NoError(t, err)

}

func TestAddRemoveTimer(t *testing.T) {
	dir, err := ioutil.TempDir("", "glass_keeper")
	assert.NoError(t, err)

	pdir := filepath.Join(dir, "project_x")
	err = os.Mkdir(pdir, 0755)
	assert.NoError(t, err)

	k, err := NewKeeper(dir)
	assert.NoError(t, err)

	defer k.Stop()

	//add new timer
	err = k.Measure(pdir)
	assert.NoError(t, err)

	<-time.After(time.Millisecond * 40)
	data, err := ioutil.ReadFile(filepath.Join(dir, "ledger.json"))
	assert.NoError(t, err)
	assert.Contains(t, string(data), "latency")

	//remove the timer
	err = k.Discard(pdir)
	assert.NoError(t, err)

	<-time.After(time.Millisecond * 40)
	data, err = ioutil.ReadFile(filepath.Join(dir, "ledger.json"))
	assert.NoError(t, err)
	assert.NotContains(t, string(data), "latency")
}
