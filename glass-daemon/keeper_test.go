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
	assert.Contains(t, string(data), "tick_rate")
}

func TestStartSave(t *testing.T) {
	dir, err := ioutil.TempDir("", "glass_keeper")
	assert.NoError(t, err)

	k, err := NewKeeper(dir)
	assert.NoError(t, err)

	go k.Start()
	defer k.Stop()

	<-time.After(time.Millisecond)

	data, err := ioutil.ReadFile(filepath.Join(dir, "ledger.json"))
	assert.NoError(t, err)
	assert.Contains(t, string(data), "tick_rate")
}

func TestAddRemoveTimer(t *testing.T) {
	dir, err := ioutil.TempDir("", "glass_keeper")
	assert.NoError(t, err)

	pdir := filepath.Join(dir, "project_x")
	pconfp := filepath.Join(pdir, "timeglass.json")
	err = os.Mkdir(pdir, 0755)
	assert.NoError(t, err)

	k, err := NewKeeper(dir)
	assert.NoError(t, err)

	p, err := NewTimer(pconfp)
	assert.NoError(t, err)

	err = k.Add(p)
	assert.NoError(t, err)
	assert.False(t, p.IsPaused())

	k.Save()
	data, err := ioutil.ReadFile(filepath.Join(dir, "ledger.json"))
	assert.NoError(t, err)
	assert.Contains(t, string(data), "/timeglass.json")

	err = k.Remove(p.ConfPath())
	assert.NoError(t, err)

	k.Save()
	data, err = ioutil.ReadFile(filepath.Join(dir, "ledger.json"))
	assert.NoError(t, err)
	assert.NotContains(t, string(data), "/timeglass.json")
	assert.True(t, p.IsPaused())

	err = k.Add(p)
	assert.NoError(t, err)
	assert.False(t, p.IsPaused())
}
