package main

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

	k, err := NewKeeper(dir)
	assert.NoError(t, err)

	_, err = ioutil.ReadFile(filepath.Join(dir, "ledger.json"))
	assert.Error(t, err)
	assert.True(t, os.IsNotExist(err))

	err = k.Load()
	assert.NoError(t, err)

	k.Save()
	data, err := ioutil.ReadFile(filepath.Join(dir, "ledger.json"))
	assert.NoError(t, err)
	assert.Contains(t, string(data), "timers")
}

func TestStartSave(t *testing.T) {
	dir, err := ioutil.TempDir("", "glass_keeper")
	assert.NoError(t, err)

	k, err := NewKeeper(dir)
	assert.NoError(t, err)

	go k.Start()
	defer k.Stop()

	<-time.After(time.Millisecond * 10)

	data, err := ioutil.ReadFile(filepath.Join(dir, "ledger.json"))
	assert.NoError(t, err)
	assert.Contains(t, string(data), "timers")
}

func TestAddRemoveTimer(t *testing.T) {
	dir, err := ioutil.TempDir("", "glass_keeper")
	assert.NoError(t, err)

	pdir := filepath.Join(dir, "project_x")
	err = os.Mkdir(pdir, 0755)
	assert.NoError(t, err)

	k, err := NewKeeper(dir)
	assert.NoError(t, err)

	go k.Start()
	defer k.Stop()

	timer, err := NewTimer(pdir)
	assert.NoError(t, err)

	//add new timer
	err = k.Add(timer)
	assert.NoError(t, err)
	assert.False(t, timer.IsPaused())

	<-time.After(time.Millisecond * 40)
	data, err := ioutil.ReadFile(filepath.Join(dir, "ledger.json"))
	assert.NoError(t, err)
	assert.Contains(t, string(data), "latency")

	//remove the timer
	err = k.Remove(timer.Dir())
	assert.NoError(t, err)

	<-time.After(time.Millisecond * 40)
	data, err = ioutil.ReadFile(filepath.Join(dir, "ledger.json"))
	assert.NoError(t, err)
	assert.NotContains(t, string(data), "latency")
	assert.True(t, timer.IsPaused())
}
