package main

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStartStop(t *testing.T) {
	dir := setupTestProject(t)

	timer, err := NewTimer(dir)
	assert.NoError(t, err)

	timer.Start()
	timer.Stop()

	assert.False(t, timer.running)
	assert.True(t, timer.IsPaused())
	assert.Nil(t, timer.monitor)
}

func TestDoubleStop(t *testing.T) {
	dir := setupTestProject(t)

	timer, err := NewTimer(dir)
	assert.NoError(t, err)

	timer.Start()
	timer.Stop()
	timer.Stop()

	assert.False(t, timer.running)
}

func TestNoStartStop(t *testing.T) {
	dir := setupTestProject(t)

	timer, err := NewTimer(dir)
	assert.NoError(t, err)

	timer.Stop()
	timer.Stop()

	assert.False(t, timer.running)
}

func TestStartTimer(t *testing.T) {
	dir := setupTestProject(t)

	timer, err := NewTimer(dir)
	assert.NoError(t, err)

	timer.Start()
	defer timer.Stop()

	assert.False(t, timer.IsPaused())
	assert.True(t, timer.running)
	assert.NotNil(t, timer.monitor)
	assert.Equal(t, "", timer.timerData.Failed)
	assert.Equal(t, time.Millisecond*5, timer.timerData.MBU)
}

func TestStartTimerDouble(t *testing.T) {
	dir := setupTestProject(t)

	timer, err := NewTimer(dir)
	assert.NoError(t, err)

	timer.Start()
	timer.Start()
	defer timer.Stop()

	assert.False(t, timer.IsPaused())
	assert.True(t, timer.running)
	assert.NotNil(t, timer.monitor)
	assert.Equal(t, "", timer.timerData.Failed)
	assert.Equal(t, time.Millisecond*5, timer.timerData.MBU)
}

func TestStartTimerFailedConfig(t *testing.T) {
	dir := setupTestProject(t)
	writeProjectFile(t, dir, "timeglass.json", fmt.Sprint(`{faulty: json`))

	timer, err := NewTimer(dir)
	assert.NoError(t, err)

	timer.Start()
	defer timer.Stop()

	assert.False(t, timer.IsPaused())
	assert.True(t, timer.running)
	assert.NotNil(t, timer.monitor)
	assert.Contains(t, timer.HasFailed(), "Failed to read configuration")
	assert.NotEqual(t, time.Millisecond*5, timer.timerData.MBU)
}

func TestStartTimerFailedMonitor(t *testing.T) {
	dir := setupTestProject(t)

	timer, err := NewTimer(dir)
	assert.NoError(t, err)

	moveProjectFile(t, dir, filepath.Join(dir, "..", "project_y"))

	timer.Start()
	defer timer.Stop()

	assert.False(t, timer.IsPaused())
	assert.True(t, timer.running)
	assert.Nil(t, timer.monitor)
	assert.Contains(t, timer.HasFailed(), "Failed to create monitor") //dir no longer exists on start
	assert.NotEqual(t, time.Millisecond*5, timer.timerData.MBU)       //config file was move
}

func TestStartTimerFailedMonitorStop(t *testing.T) {
	dir := setupTestProject(t)

	timer, err := NewTimer(dir)
	assert.NoError(t, err)

	moveProjectFile(t, dir, filepath.Join(dir, "..", "project_y"))

	timer.Start()
	assert.True(t, timer.running)
	assert.Nil(t, timer.monitor)

	timer.Stop()
	assert.False(t, timer.running)
	assert.Nil(t, timer.monitor)
}

func TestStartTimerFailedMonitorRecover(t *testing.T) {
	dir := setupTestProject(t)

	timer, err := NewTimer(dir)
	assert.NoError(t, err)

	moveProjectFile(t, dir, filepath.Join(dir, "..", "project_y"))

	timer.Start()
	assert.True(t, timer.running)
	assert.Nil(t, timer.monitor)

	moveProjectFile(t, filepath.Join(dir, "..", "project_y"), dir)

	timer.Start()
	defer timer.Stop()

	assert.False(t, timer.IsPaused())
	assert.True(t, timer.running)
	assert.NotNil(t, timer.monitor)
	assert.Equal(t, "", timer.timerData.Failed)
	assert.Equal(t, time.Millisecond*5, timer.timerData.MBU)
}

func TestResetPauseUnpauseWhileStopped(t *testing.T) {
	dir := setupTestProject(t)
	timer, err := NewTimer(dir)
	assert.NoError(t, err)

	timer.Start()
	defer timer.Stop()
	<-time.After(time.Millisecond * 10)

	timer.Stop()
	<-time.After(time.Millisecond)

	timer.Pause()
	timer.Unpause()

	assertTime(t, timer, time.Millisecond*10)
	timer.Reset()

	assert.Equal(t, 0, timer.Time())
}

func TestResetWhileRunning(t *testing.T) {
	dir := setupTestProject(t)
	timer, err := NewTimer(dir)
	assert.NoError(t, err)

	timer.Start()
	defer timer.Stop()
	<-time.After(time.Millisecond * 10)
	assertTime(t, timer, time.Millisecond*10)

	//reset but ticking
	timer.Reset()
	<-time.After(time.Millisecond * 2)
	assert.Equal(t, time.Millisecond*5, timer.Time())
}

func TestResetWhilePaused(t *testing.T) {
	dir := setupTestProject(t)
	timer, err := NewTimer(dir)
	assert.NoError(t, err)

	timer.Start()
	// defer timer.Stop()
	<-time.After(time.Millisecond * 10)
	assertTime(t, timer, time.Millisecond*10)

	timer.Pause()
	timer.Reset()
	<-time.After(time.Millisecond)
	assert.Equal(t, 0, timer.Time())
}

func TestUnpause(t *testing.T) {
	dir := setupTestProject(t)
	timer, err := NewTimer(dir)
	assert.NoError(t, err)

	timer.Start()
	defer timer.Stop()
	<-time.After(time.Millisecond * 10)
	assertTime(t, timer, time.Millisecond*10)

	timer.Pause()
	<-time.After(time.Millisecond * 30)
	assertTime(t, timer, time.Millisecond*10)

	timer.Unpause()
	<-time.After(time.Millisecond * 5)
	assertTime(t, timer, time.Millisecond*15)
}
