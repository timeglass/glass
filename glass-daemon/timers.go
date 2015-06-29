package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/timeglass/glass/_vendor/github.com/hashicorp/errwrap"

	"github.com/timeglass/glass/model"
)

type Timer struct{}

type Keeper struct {
	ledgerPath string
	stop       chan struct{}

	keeperData *keeperData
}

type keeperData struct {
	TickRate time.Duration    `json:"tick_rate"`
	Timers   map[string]Timer `json:"timers,omitempty"`
}

func NewKeeper() (*Keeper, error) {
	k := &Keeper{
		stop: make(chan struct{}),
		keeperData: &keeperData{
			TickRate: time.Second * 10,
		},
	}

	p, err := model.SystemTimeglassPathCreateIfNotExist()
	if err != nil {
		return nil, errwrap.Wrapf("Failed to find Timeglass system path: {{err}}", err)
	}

	//attempt to open json file, if it exsts
	k.ledgerPath = filepath.Join(p, "ledger.json")
	return k, k.Load()
}

func (k *Keeper) Data() *keeperData {
	return k.keeperData
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
		select {
		case <-k.stop:
			return

		case <-time.After(k.TickRate()):
			log.Printf("Tick on %s", time.Now())
			err := k.Load()
			if err != nil {
				log.Printf("Error while loading from ledger: %s", err)
			}

		}

		err := k.Save()
		if err != nil {
			log.Printf("Error while saving to ledger: %s", err)
		}
	}
}

func (k *Keeper) TickRate() time.Duration {
	return k.keeperData.TickRate
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
		err := dec.Decode(k.keeperData)
		if err != nil {
			return errwrap.Wrapf(fmt.Sprintf("Failed to decode JSOn in '%s': {{err}}", k.ledgerPath), err)
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
	err = enc.Encode(k.keeperData)
	if err != nil {
		return errwrap.Wrapf(fmt.Sprintf("Error saving ledger to '%s': {{err}}", k.ledgerPath), err)
	}

	return nil
}
