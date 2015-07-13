package vcs

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrNoRemote = errors.New("Version control has no remote")
var ErrNoRemoteTimeData = errors.New("Remote doesn't have any time data")
var ErrNoLocalTimeData = errors.New("Local clone doesn't have any time data")
var ErrNoCommitTimeData = errors.New("Commit doesn't have any time data")

type VCS interface {
	Name() string
	Root() string
	IsAvailable() bool
	Hook() error
	Push(string, string) error
	Pull(string) error
	DefaultRemote() (string, error)
	Persist(time.Duration) error
	Staging() (map[string]*StagedFile, error)
	Show(string) (TimeData, error)
}

type StagedFile struct {
	date time.Time
	hash string
	path string
}

func NewStagedFile(date time.Time, hash, path string) *StagedFile {
	return &StagedFile{date, hash, path}
}

func (g *StagedFile) Date() time.Time { return g.date }
func (g *StagedFile) Hash() string    { return g.hash }
func (g *StagedFile) Path() string    { return g.path }

type TimeData interface {
	Total() time.Duration
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

	return nil, fmt.Errorf("No supported Version Control System found in '%s', checked for: %s", dir, strings.Join(checked, ","))
}
