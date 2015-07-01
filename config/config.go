package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/timeglass/glass/_vendor/github.com/hashicorp/errwrap"
)

type MBU time.Duration

func (m MBU) String() string { return time.Duration(m).String() }

func (t *MBU) UnmarshalJSON(data []byte) error {
	raw, err := strconv.Unquote(string(data))
	if err != nil {
		return err
	}

	parsed, err := time.ParseDuration(raw)
	if err != nil {
		return errwrap.Wrapf("Failed to parse duration: {{err}}", err)
	}

	*t = MBU(parsed)
	return nil
}

var DefaultConfig = &Config{
	MBU:           MBU(time.Minute),
	CommitMessage: " [{{.}}]",
	AutoPush:      true,
}

type Config struct {
	MBU           MBU    `json:"mbu"`
	CommitMessage string `json:"commit_message"`
	AutoPush      bool   `json:"auto_push"`
}

func ReadConfig(dir string) (*Config, error) {
	conf := &Config{}
	confp := filepath.Join(dir, "timeglass.json")
	confdata, err := ioutil.ReadFile(confp)
	if err != nil && !os.IsNotExist(err) {
		return nil, errwrap.Wrapf(fmt.Sprintf("Failed to read project configuration even though it exist: {{err}}"), err)
	} else if os.IsNotExist(err) {
		conf = DefaultConfig
	} else {
		err := json.Unmarshal(confdata, &conf)
		if err != nil {
			return nil, errwrap.Wrapf(fmt.Sprintf("Failed to parse configuration JSON: {{err}}"), err)
		}
	}

	return conf, nil
}
