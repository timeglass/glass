package timer

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func assertDuration(t *testing.T, expected, actual time.Duration) {
	delta := time.Millisecond
	assert.InDelta(t, float64(expected), float64(actual), float64(delta), fmt.Sprintf("Max difference between %s and %s allowed is %s", expected, actual, delta))
}

func point(d string) time.Time {
	if d == "" {
		return time.Now()
	}

	parsed, err := time.ParseDuration(d)
	if err != nil {
		panic(err)
	}

	return time.Now().Add(parsed)
}

func TestSingleLine(t *testing.T) {
	d := NewDistributor()
	d.Register("")
	d.Distribute(time.Second, point("5s"))
	d.Distribute(time.Second, point("10s"))
	d.Distribute(time.Second, point("15s"))

	res, err := d.Extract("", point("15s"))
	assert.NoError(t, err)
	assert.Equal(t, time.Second*3, res)
}

func TestSingleLineStage(t *testing.T) {
	d := NewDistributor()
	d.Register("")
	d.Distribute(time.Second, point("5s"))
	d.Distribute(time.Second, point("10s"))
	d.Distribute(time.Second, point("15s"))
	d.Distribute(time.Second, point("20s"))

	res, err := d.Extract("", point("10s"))
	assert.NoError(t, err)
	assert.Equal(t, time.Second*2, res)

	d.Stage("", point("15s"))

	res = d.Timelines()[OverheadTimeline].Staged()
	assert.Equal(t, time.Second*3, res)

	res = d.Timelines()[OverheadTimeline].Unstaged()
	assert.Equal(t, time.Second*1, res)

	d.Stage("", point("5s"))

	res = d.Timelines()[OverheadTimeline].Staged()
	assert.Equal(t, time.Second*1, res)

	res = d.Timelines()[OverheadTimeline].Unstaged()
	assert.Equal(t, time.Second*3, res)

	d.Stage("", point("15s"))
	d.Reset(ResetOpts{Staged: true})

	res = d.Timelines()[OverheadTimeline].Staged()
	assert.Equal(t, time.Second*0, res)

	res = d.Timelines()[OverheadTimeline].Unstaged()
	assert.Equal(t, time.Second*1, res)

	res, err = d.Extract("", point("10s"))
	assert.NoError(t, err)
	assert.Equal(t, time.Second*0, res)

}

func TestSingleLineCutoff(t *testing.T) {
	d := NewDistributor()
	d.Register("")
	d.Distribute(time.Second, point("5s"))
	d.Distribute(time.Second, point("10s"))
	d.Distribute(time.Second, point("15s"))

	res, err := d.Extract("", point("10s"))
	assert.NoError(t, err)
	assert.Equal(t, time.Second*2, res)
}

func TestMultilineLineCutoff(t *testing.T) {
	d := NewDistributor()
	d.Register("x.go")
	d.Register("y.go")
	d.Distribute(time.Second, point("5s"))
	d.Distribute(time.Second, point("10s"))
	d.Distribute(time.Second, point("15s"))

	res, err := d.Extract("x.go", point("10s"))
	assert.NoError(t, err)
	assert.Equal(t, time.Second, res)

	res, err = d.Extract("y.go", point("10s"))
	assert.NoError(t, err)
	assert.Equal(t, time.Second, res)
}

func TestMultilineLineBreaked(t *testing.T) {
	d := NewDistributor()
	d.Register("x.go")
	d.Register("y.go")
	d.Distribute(time.Second, point("5s"))
	d.Distribute(time.Second, point("10s"))
	d.Break()
	d.Distribute(time.Second, point("15s"))

	res, err := d.Extract("x.go", point("30s"))
	assert.NoError(t, err)
	assert.Equal(t, time.Second, res)

	res, err = d.Extract("y.go", point("30s"))
	assert.NoError(t, err)
	assert.Equal(t, time.Second, res)
}
