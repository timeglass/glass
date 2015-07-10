package timer

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func init() {
	log.SetOutput(os.Stderr)
}

func setupTestProject(t *testing.T) string {
	dir, err := ioutil.TempDir("", fmt.Sprintf("glass_keeper"))
	assert.NoError(t, err)

	pdir := filepath.Join(dir, "project_x")
	err = os.Mkdir(pdir, 0755)
	assert.NoError(t, err)

	err = ioutil.WriteFile(filepath.Join(pdir, "timeglass.json"), []byte(`{"mbu": "5ms"}`), 0755)
	assert.NoError(t, err)

	<-time.After(time.Millisecond * 20)
	return pdir
}

func writeProjectFile(t *testing.T, dir, path, content string) {
	err := ioutil.WriteFile(filepath.Join(dir, path), []byte(content), 0755)
	assert.NoError(t, err)
}

func moveProjectFile(t *testing.T, from, to string) {
	err := os.Rename(from, to)
	assert.NoError(t, err)
}

func assertTime(t *testing.T, timer *Timer, expected time.Duration) {
	assert.InDelta(t, float64(expected), float64(timer.Time()), float64(timer.timerData.MBU+time.Millisecond), fmt.Sprintf("Max difference between %s and %s allowed is %s", expected, timer.Time(), timer.timerData.MBU+time.Millisecond))
}
