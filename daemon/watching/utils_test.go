package watching

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TempDir(t *testing.T) (string, func()) {
	path, err := ioutil.TempDir("", fmt.Sprint("sourceclock"))
	path, _ = filepath.EvalSymlinks(path)
	if err != nil {
		t.Error(err)
	}
	return path, func() {
		err := os.RemoveAll(path)
		if err != nil {
			t.Error(err)
		}
	}
}

func DTempDir(t *testing.T, d time.Duration) (string, func()) {
	defer time.Sleep(d)
	return TempDir(t)
}
