package vcs

import (
	"os"
	"path/filepath"
)

type Git struct {
	dir string
}

func NewGit(dir string) *Git {
	return &Git{
		dir: dir,
	}
}

func (g *Git) Name() string { return "git" }
func (g *Git) Supported() bool {
	fi, err := os.Stat(filepath.Join(g.dir, ".git"))
	if err != nil || !fi.IsDir() {
		return false
	}

	return true
}

//@todo implement
func (g *Git) Install() error {

	//parse templates

	//write files

	return nil
}

//@todo implement
func (g *Git) Uninstall() error {

	//remove files

	return nil
}
