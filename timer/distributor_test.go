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

func TestRegisterOverHeadOpenTimeline(t *testing.T) {
	d := NewDistributor()

	d.Register("", point(""))
	d.Register("", point("5s"))
	d.Register("", point("10s"))
	res, err := d.Extract("", point("15s"))

	//expect only 10s since the open timeline is
	//ver closed
	assert.NoError(t, err)
	assertDuration(t, time.Second*10, res)
}

func TestRegisterOverHeadClosedTimeline(t *testing.T) {
	d := NewDistributor()

	d.Register("", point(""))
	d.Register("", point("5s"))
	d.Register("", point("10s"))
	d.Break(point("15s"))
	res, err := d.Extract("", point("15s"))

	//expect only 10s since the open timeline is
	//ver closed
	assert.NoError(t, err)
	assertDuration(t, time.Second*15, res)
}
