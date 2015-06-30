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

func TestStartStopTimer(t *testing.T) {
	dir, err := ioutil.TempDir("", fmt.Sprintf("glass_keeper"))
	assert.NoError(t, err)

	pdir := filepath.Join(dir, "project_x")
	pconfp := filepath.Join(pdir, "timeglass.json")
	err = os.Mkdir(pdir, 0755)
	assert.NoError(t, err)

	timer, err := NewTimer(pconfp)
	assert.NoError(t, err)

	timer.timerData.latency = time.Millisecond
	timer.timerData.mbu = time.Millisecond * 5
	timer.timerData.timeout = time.Millisecond * 20

	//after initial start, immediately add 5 milliseconds
	timer.Start()
	<-time.After(time.Millisecond)
	assert.Equal(t, time.Millisecond*5, timer.Time())

	//should have timed out after 20 millisecond
	<-time.After(time.Millisecond * 40)
	assert.Equal(t, time.Millisecond*20, timer.Time())
	assert.True(t, timer.IsPaused())

	//adding a file should recontinue, add add another 10 milliseconds immediately
	err = ioutil.WriteFile(pconfp, []byte("{}"), 0777)
	assert.NoError(t, err)

	<-time.After(20 * time.Millisecond)
	assert.NotEqual(t, time.Millisecond*20, timer.Time())
	assert.False(t, timer.IsPaused())
}
