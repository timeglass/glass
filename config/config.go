package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/imdario/mergo"
	"github.com/timeglass/glass/_vendor/github.com/hashicorp/errwrap"
)

var confFilename = "timeglass.json"

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

func ReadConfig(dir, sysdir string) (*Config, error) {

	//get system wide config
	sysconf := &Config{}
	sysconfp := filepath.Join(sysdir, confFilename)
	sysconfdata, err := ioutil.ReadFile(sysconfp)
	if err != nil && !os.IsNotExist(err) {
		return nil, errwrap.Wrapf(fmt.Sprintf("Failed to read system configuration file '%s' even though it exist: {{err}}", sysconfp), err)
	} else if os.IsNotExist(err) {
		sysconf = DefaultConfig
	} else {
		err := json.Unmarshal(sysconfdata, &sysconf)
		if err != nil {
			return nil, errwrap.Wrapf(fmt.Sprintf("Failed to parse system configuration '%s': {{err}}", sysconfp), err)
		}

		err = mergo.Merge(sysconf, DefaultConfig)
		if err != nil {
			return nil, errwrap.Wrapf("Failed to merge system config with default config: {{err}} ", err)
		}
	}

	//get project wide config
	conf := &Config{}
	confp := filepath.Join(dir, confFilename)
	confdata, err := ioutil.ReadFile(confp)
	if err != nil && !os.IsNotExist(err) {
		return nil, errwrap.Wrapf(fmt.Sprintf("Failed to read project configuration  file '%s' even though it exist: {{err}}", confp), err)
	} else if os.IsNotExist(err) {
		conf = sysconf
	} else {
		err := json.Unmarshal(confdata, &conf)
		if err != nil {
			return nil, errwrap.Wrapf(fmt.Sprintf("Failed to parse project configuration '%s': {{err}}", confp), err)
		}

		err = mergo.Merge(conf, sysconf)
		if err != nil {
			return nil, errwrap.Wrapf("Failed to merge project config with system config: {{err}} ", err)
		}
	}

	return conf, nil
}
