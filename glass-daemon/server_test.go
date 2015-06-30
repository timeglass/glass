package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApiRoot(t *testing.T) {
	dir, err := ioutil.TempDir("", fmt.Sprintf("glass_keeper"))
	assert.NoError(t, err)

	k, err := NewKeeper(dir)
	assert.NoError(t, err)

	svr, err := NewServer(":0", k)
	assert.NoError(t, err)

	r, err := http.NewRequest("GET", "/api/", nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	svr.api(w, r)

	assert.Contains(t, w.Body.String(), "version")
	assert.Contains(t, w.Body.String(), "timers")
}

func TestCreateRemoveTimer(t *testing.T) {
	dir, err := ioutil.TempDir("", fmt.Sprintf("glass_keeper"))
	assert.NoError(t, err)

	k, err := NewKeeper(dir)
	assert.NoError(t, err)

	svr, err := NewServer(":0", k)
	assert.NoError(t, err)

	params := &url.Values{
		"conf": []string{filepath.Join(dir, "timeglass.json")},
	}

	r, err := http.NewRequest("GET", "/api/timers.create?"+params.Encode(), nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	svr.timersCreate(w, r)

	assert.Equal(t, http.StatusCreated, w.Code)

	r, err = http.NewRequest("GET", "/api/timers.delete?"+params.Encode(), nil)
	assert.NoError(t, err)

	w = httptest.NewRecorder()
	svr.timersDelete(w, r)

	assert.Equal(t, http.StatusNoContent, w.Code)
}
