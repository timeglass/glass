package vcs

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrNoRemoteTimeData = errors.New("Remote doesn't have any time data")
var ErrNoLocalTimeData = errors.New("Local clone doesn't have any time data")

type VCS interface {
	Name() string
	IsAvailable() bool
	Hook() error
	Push(string, string) error
	Fetch(string) error
	DefaultRemote() string
	Persist(time.Duration) error
	ParseHistory() error
}

func GetVCS(dir string) (VCS, error) {
	var supported = []VCS{
		NewGit(dir),
	}

	var checked = []string{}
	for _, vcs := range supported {
		if vcs.IsAvailable() {
			return vcs, nil
		}
		checked = append(checked, vcs.Name())
	}

	return nil, fmt.Errorf("No supported version system found in '%s', checked for: %s", dir, strings.Join(checked, ","))
}
