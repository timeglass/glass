package timer

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewStopTimer(t *testing.T) {
	dir := setupTestProject(t)

	//start
	timer, err := NewTimer(dir)
	assert.NoError(t, err)

	//stop
	timer.Stop()
	assert.NoError(t, err)

}

func TestDecodeStopTimer(t *testing.T) {
	var timer *Timer
	err := json.Unmarshal([]byte(`{}`), &timer)
	timer.Stop()
	assert.NoError(t, err)
}

func TestEncodeTimer(t *testing.T) {
	dir := setupTestProject(t)

	timer, err := NewTimer(dir)
	assert.NoError(t, err)

	bytes, err := json.Marshal(timer)
	assert.True(t, timer.running)
	assert.NoError(t, err)
	assert.Contains(t, string(bytes), `"paused":false`)
	assert.Contains(t, string(bytes), `"distributor":`)

	timer.Stop()
	timer.Pause()
	assert.NoError(t, err)

	bytes, err = json.Marshal(timer)
	assert.False(t, timer.running)
	assert.NoError(t, err)
	assert.Contains(t, string(bytes), `"paused":true`)
}

func TestDoubleStop(t *testing.T) {
	dir := setupTestProject(t)

	timer, err := NewTimer(dir)
	assert.NoError(t, err)

	timer.Stop()
	timer.Stop()
}

func TestStartTimerFailedConfig(t *testing.T) {
	dir := setupTestProject(t)
	writeProjectFile(t, dir, "timeglass.json", fmt.Sprint(`{faulty: json`))

	timer, err := NewTimer(dir)
	assert.Error(t, err)
	assert.Nil(t, timer)
}

func TestBasicTimeReadWhenStarted(t *testing.T) {
	dir := setupTestProject(t)
	timer, err := NewTimer(dir)
	assert.NoError(t, err)

	assertTime(t, timer, time.Millisecond*10)
	timer.Reset(false)
	assertTime(t, timer, 0)
}

func TestBasicTimeReadResetPauseUnpauseWhenStopped(t *testing.T) {
	dir := setupTestProject(t)
	timer, err := NewTimer(dir)
	assert.NoError(t, err)

	timer.Stop()

	assertTime(t, timer, time.Millisecond*10)
	timer.Reset(false)
	assertTime(t, timer, 0)

	timer.Pause()
	timer.Unpause()
}

func TestTimerTimeout(t *testing.T) {
	dir := setupTestProject(t)
	timer, err := NewTimer(dir)
	assert.NoError(t, err)

	<-time.After(time.Millisecond * 100)
	assertTime(t, timer, time.Millisecond*20)
}

func TestPauseUnpause(t *testing.T) {
	dir := setupTestProject(t)
	timer, err := NewTimer(dir)
	assert.NoError(t, err)

	<-time.After(time.Millisecond * 10)
	assertTime(t, timer, time.Millisecond*10)

	timer.Pause()
	<-time.After(time.Millisecond * 30)
	assertTime(t, timer, time.Millisecond*10)

	timer.Unpause()
	<-time.After(time.Millisecond * 10)
	assertTime(t, timer, time.Millisecond*20)
}

func TestPauseUnpauseWriteFile(t *testing.T) {
	dir := setupTestProject(t)
	timer, err := NewTimer(dir)
	assert.NoError(t, err)

	<-time.After(time.Millisecond * 10)
	assertTime(t, timer, time.Millisecond*10)

	timer.Pause()
	<-time.After(time.Millisecond * 30)
	assertTime(t, timer, time.Millisecond*10)

	writeProjectFile(t, dir, "test.go", `{a}`)
	<-time.After(time.Millisecond * 100)
	writeProjectFile(t, dir, "test.go", `{bc}`)
	<-time.After(time.Millisecond * 100)

	assertTime(t, timer, time.Millisecond*30)
}
