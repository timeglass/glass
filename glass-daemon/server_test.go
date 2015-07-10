package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/timeglass/glass/timer"

	"github.com/stretchr/testify/assert"
)

func TestApiRoot(t *testing.T) {
	dir, err := ioutil.TempDir("", fmt.Sprintf("glass_keeper"))
	assert.NoError(t, err)

	k, err := timer.NewKeeper(dir)
	assert.NoError(t, err)

	defer k.Stop()

	svr, err := NewServer(":0", k)
	assert.NoError(t, err)

	r, err := http.NewRequest("GET", "/api/", nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	svr.api(w, r)

	assert.Contains(t, w.Body.String(), "version")
	assert.Contains(t, w.Body.String(), "timers")
}

func TestCreateInfoRemoveTimer(t *testing.T) {
	dir, err := ioutil.TempDir("", fmt.Sprintf("glass_keeper"))
	assert.NoError(t, err)

	k, err := timer.NewKeeper(dir)
	assert.NoError(t, err)

	defer k.Stop()

	svr, err := NewServer(":0", k)
	assert.NoError(t, err)

	params := &url.Values{
		"dir": []string{dir},
	}

	//Create
	r, err := http.NewRequest("GET", "/api/timers.create?"+params.Encode(), nil)
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	svr.timersCreate(w, r)

	assert.Equal(t, http.StatusCreated, w.Code)

	// Info
	r, err = http.NewRequest("GET", "/api/timers.info?"+params.Encode(), nil)
	assert.NoError(t, err)
	w = httptest.NewRecorder()
	svr.timersInfo(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "latency")

	//Delete
	r, err = http.NewRequest("GET", "/api/timers.delete?"+params.Encode(), nil)

	assert.NoError(t, err)
	w = httptest.NewRecorder()
	svr.timersDelete(w, r)

	assert.Equal(t, http.StatusNoContent, w.Code)

	//Info
	r, err = http.NewRequest("GET", "/api/timers.info?"+params.Encode(), nil)
	assert.NoError(t, err)
	w = httptest.NewRecorder()
	svr.timersInfo(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
