package timer

import (
	"log"
	"time"

	"github.com/timeglass/snow/index"
)

//maps file paths to time spent on each
type Distribution map[string]time.Duration

// A Distributor takes a channel of
// file events and uses the timestamp
// from the series to distribute the
// total time measured across files
type Distributor struct{}

func NewDistributer() *Distributor {
	return &Distributor{}
}

func (d *Distributor) Register(ev index.FileEvent) {
	log.Printf("Registered file event: %s", ev)

}

func (d *Distributor) Distribute(total time.Duration) (Distribution, error) {
	dist := make(Distribution)

	log.Println("Distributing...")
	return dist, nil
}
