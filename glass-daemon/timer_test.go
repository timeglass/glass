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

	timer, err := NewTimer(pdir)
	assert.NoError(t, err)

	timer.timerData.Latency = time.Millisecond
	timer.timerData.MBU = time.Millisecond * 5
	timer.timerData.Timeout = time.Millisecond * 20

	//after initial start, immediately add 5 milliseconds
	timer.Start()
	<-time.After(time.Millisecond)
	assert.Equal(t, time.Millisecond*5, timer.Time())

	//should have timed out after 20 millisecond
	<-time.After(time.Millisecond * 40)
	assert.Equal(t, time.Millisecond*20, timer.Time())
	assert.True(t, timer.IsPaused())

	//adding a file should recontinue, add add another 10 milliseconds immediately
	err = ioutil.WriteFile(filepath.Join(pdir, "timeglass.json"), []byte("{}"), 0777)
	assert.NoError(t, err)

	<-time.After(20 * time.Millisecond)
	assert.NotEqual(t, time.Millisecond*20, timer.Time())
	assert.False(t, timer.IsPaused())

	//resetting should not ifluence paused state but start at 0
	//and add first tick immediately
	timer.Reset()
	<-time.After(time.Millisecond)
	assert.False(t, timer.IsPaused())
	assert.Equal(t, time.Millisecond*5, timer.Time())

}
