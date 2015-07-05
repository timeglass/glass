package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/timeglass/glass/_vendor/github.com/hashicorp/errwrap"
)

type Keeper struct {
	ledgerPath string
	stop       chan struct{}
	save       chan struct{}

	keeperData *keeperData
}

type keeperData struct {
	Timers map[string]*Timer `json:"timers"`
}

func NewKeeper(path string) (*Keeper, error) {
	k := &Keeper{
		stop: make(chan struct{}),
		save: make(chan struct{}),
		keeperData: &keeperData{
			Timers: map[string]*Timer{},
		},
	}

	//attempt to open json file, if it exsts
	k.ledgerPath = filepath.Join(path, "ledger.json")
	return k, k.Load()
}

func (k *Keeper) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &k.keeperData)
}

func (k *Keeper) MarshalJSON() ([]byte, error) {
	return json.Marshal(k.keeperData)
}

func (k *Keeper) Add(t *Timer) error {
	if tt, ok := k.keeperData.Timers[t.Dir()]; !ok {
		log.Printf("New timer '%s' for keeper, adding to collection...", t.Dir())
		k.keeperData.Timers[t.Dir()] = t
		t.SetSave(k.save)
	} else {
		log.Printf("Timer '%s' exists for keeper, unpausing...", t.Dir())
		tt.Unpause()
		t = tt
	}

	t.Start()
	return nil
}

func (k *Keeper) Get(dir string) (*Timer, error) {
	if t, ok := k.keeperData.Timers[dir]; ok {
		return t, nil
	}

	return nil, fmt.Errorf("No known timer for '%s'", dir)
}

func (k *Keeper) Remove(dir string) error {
	if t, ok := k.keeperData.Timers[dir]; ok {
		delete(k.keeperData.Timers, dir)

		k.save <- struct{}{}
		t.Stop()
		return nil
	}

	return fmt.Errorf("No known timer for '%s'", dir)
}

func (k *Keeper) Stop() {
	k.stop <- struct{}{}
}

func (k *Keeper) Start() {
	log.Printf("Started time keeper on %s", time.Now())
	defer func() {
		log.Printf("Stopped time keeper on %s", time.Now())
	}()

	for {

		//save state
		err := k.Save()
		if err != nil {
			log.Printf("Error while saving to ledger: %s", err)
		}

		//stop or wait for next tick
		select {
		case <-k.stop:
			return
		case <-k.save:
		}
	}
}

func (k *Keeper) Load() error {
	f, err := os.Open(k.ledgerPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return errwrap.Wrapf(fmt.Sprintf("Failed to open '%s': {{err}}", k.ledgerPath), err)
		}
	} else {
		defer f.Close()
		dec := json.NewDecoder(f)
		err := dec.Decode(k)
		if err != nil {
			return errwrap.Wrapf(fmt.Sprintf("Failed to decode JSON in '%s': {{err}}", k.ledgerPath), err)
		}

		//immediately restart and link save channel if not paused
		for _, t := range k.keeperData.Timers {
			t.SetSave(k.save)
			if !t.IsPaused() {
				t.Start()
			}
		}

	}

	return nil
}

func (k *Keeper) Save() error {
	f, err := os.OpenFile(k.ledgerPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	err = enc.Encode(k)
	if err != nil {
		return errwrap.Wrapf(fmt.Sprintf("Error saving ledger to '%s': {{err}}", k.ledgerPath), err)
	}

	return nil
}
