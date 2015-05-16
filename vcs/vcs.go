package vcs

import (
	"fmt"
	"strings"
)

type VCS interface {
	Name() string
	Supported() bool
	Hook() error
}

func GetVCS(dir string) (VCS, error) {
	var supported = []VCS{
		NewGit(dir),
	}

	var checked = []string{}
	for _, vcs := range supported {
		if vcs.Supported() {
			return vcs, nil
		}
		checked = append(checked, vcs.Name())
	}

	return nil, fmt.Errorf("No supported version system found in '%s', checked for: %s", dir, strings.Join(checked, ","))
}
