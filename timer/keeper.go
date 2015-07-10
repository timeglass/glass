package timer

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/timeglass/glass/_vendor/github.com/hashicorp/errwrap"
)

type Keeper struct {
	save       chan struct{}
	stop       chan struct{}
	ledger     string
	keeperData *keeperData
	m          *sync.Mutex
}

type keeperData struct {
	Timers map[string]*Timer `json:"timers"`
}

func NewKeeper(path string) (*Keeper, error) {
	k := &Keeper{
		m:      &sync.Mutex{},
		ledger: filepath.Join(path, "ledger.json"),
		keeperData: &keeperData{
			Timers: map[string]*Timer{},
		},
		stop: make(chan struct{}),
		save: make(chan struct{}, 1),
	}

	err := k.Load()
	if err != nil {
		return nil, err
	}

	return k, k.Save()
}

func (k *Keeper) run() {
	log.Printf("Started time keeper on %s", time.Now())
	defer func() {
		log.Printf("Stopped time keeper on %s", time.Now())
	}()

	for {
		select {
		case <-k.save:
			err := k.Save()
			if err != nil {
				log.Printf("Failsed to save time keeping: %s", err)
			}

		case <-k.stop:
			return
		}
	}
}

func (k *Keeper) Measure(dir string) error {
	k.m.Lock()
	defer k.m.Unlock()

	if t, ok := k.keeperData.Timers[dir]; ok {
		log.Printf("Timer '%s' exists for keeper, unpausing...", t.Dir())
		t.Unpause()
		return nil
	}

	t, err := NewTimer(dir)
	if err != nil {
		return errwrap.Wrapf(fmt.Sprintf("Failed to create new timer for '%s': {{err}}", dir), err)
	}

	log.Printf("New timer '%s' for keeper, adding to collection...", t.Dir())
	k.keeperData.Timers[dir] = t
	time.AfterFunc(time.Second, func() {
		panic("")
	})

	k.save <- struct{}{}
	t.SetSave(k.save)
	return nil
}

func (k *Keeper) Discard(dir string) error {
	k.m.Lock()
	defer k.m.Unlock()

	if _, ok := k.keeperData.Timers[dir]; !ok {
		return fmt.Errorf("No known timer for: '%s'")
	} else {
		delete(k.keeperData.Timers, dir)
		k.save <- struct{}{}
	}

	return nil
}

func (k *Keeper) Inspect(dir string) (*Timer, error) {
	k.m.Lock()
	defer k.m.Unlock()

	if t, ok := k.keeperData.Timers[dir]; ok {
		return t, nil
	}

	return nil, fmt.Errorf("No known timer for: '%s'")
}

func (k *Keeper) Stop() {
	k.stop <- struct{}{}
}

func (k *Keeper) Save() error {
	k.m.Lock()
	defer k.m.Unlock()

	f, err := os.OpenFile(k.ledger, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	err = enc.Encode(k)
	if err != nil {
		return errwrap.Wrapf(fmt.Sprintf("Error saving ledger to '%s': {{err}}", k.ledger), err)
	}

	return nil
}

func (k *Keeper) Load() error {
	f, err := os.Open(k.ledger)
	if err != nil {
		if !os.IsNotExist(err) {
			return errwrap.Wrapf(fmt.Sprintf("Failed to open '%s': {{err}}", k.ledger), err)
		}
	} else {
		defer f.Close()
		defer k.m.Unlock()

		k.m.Lock()
		dec := json.NewDecoder(f)
		err := dec.Decode(k)
		if err != nil {
			return errwrap.Wrapf(fmt.Sprintf("Failed to decode JSON in '%s': {{err}}", k.ledger), err)
		}
	}

	go k.run()
	return nil
}

func (k *Keeper) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &k.keeperData)
}

func (k *Keeper) MarshalJSON() ([]byte, error) {
	return json.Marshal(k.keeperData)
}
