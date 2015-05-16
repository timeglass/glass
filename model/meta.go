package model

import (
	"encoding/json"
)

var MetaBucketName = "meta"
var DeamonKeyName = "_daemon"

type Daemon struct {
	Repo string `json:"repo"`
	Addr string `json:"addr"`
}

func NewDeamon(repopath string, addr string) *Daemon {
	return &Daemon{
		Repo: repopath,
		Addr: addr,
	}
}

func NewDaemonFromSerialized(data []byte) (*Daemon, error) {
	var d *Daemon
	err := json.Unmarshal(data, &d)

	return d, err
}

func (d *Daemon) Serialize() ([]byte, error) {
	return json.Marshal(d)
}
