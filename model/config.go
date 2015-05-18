package model

import (
	"strconv"
	"time"

	"github.com/hashicorp/errwrap"
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
}

type Config struct {
	MBU           MBU    `json:"mbu"`
	CommitMessage string `json:"commit_message"`
}
