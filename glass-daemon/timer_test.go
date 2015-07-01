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

func TestStartStopResetTimer(t *testing.T) {
	dir, err := ioutil.TempDir("", fmt.Sprintf("glass_keeper"))
	assert.NoError(t, err)

	pdir := filepath.Join(dir, "project_x")
	err = os.Mkdir(pdir, 0755)
	assert.NoError(t, err)
	err = ioutil.WriteFile(filepath.Join(pdir, "timeglass.json"), []byte(`{"mbu": "5ms"}`), 0755)
	assert.NoError(t, err)

	timer, err := NewTimer(pdir)
	assert.NoError(t, err)

	timer.timerData.Latency = time.Millisecond

	//after initial start, immediately add 5 milliseconds
	timer.Start()
	<-time.After(time.Millisecond)
	assert.Equal(t, time.Millisecond*5, timer.Time())

	//should have timed out around 25ms
	<-time.After(time.Millisecond * 60)
	assert.True(t, timer.Time() < time.Millisecond*35)
	assert.True(t, timer.IsPaused())

	//adding a file should recontinue
	err = ioutil.WriteFile(filepath.Join(pdir, "file.go"), []byte("content"), 0777)
	assert.NoError(t, err)

	<-time.After(25 * time.Millisecond)
	assert.True(t, timer.Time() > time.Millisecond*20)
	assert.False(t, timer.IsPaused())

	//resetting should not ifluence paused state but start at 0
	//and add first tick immediately
	timer.Reset()
	<-time.After(time.Millisecond)
	assert.False(t, timer.IsPaused())
	assert.Equal(t, time.Millisecond*5, timer.Time())

}

// @todo test failing state
// @todo test pause/stop when timer is stopped
// @todo test start in multitude of situations
