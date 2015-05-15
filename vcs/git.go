package vcs

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/hashicorp/errwrap"
)

var PostCheckoutTmpl = template.Must(template.New("name").Parse(`#!/bin/sh

echo checkout!
`))

var PostCommitTmpl = template.Must(template.New("name").Parse(`#!/bin/sh

echo commit!
`))

type Git struct {
	dir string
}

func NewGit(dir string) *Git {
	return &Git{
		dir: filepath.Join(dir, ".git"),
	}
}

func (g *Git) Name() string { return "git" }
func (g *Git) Supported() bool {
	fi, err := os.Stat(g.dir)
	if err != nil || !fi.IsDir() {
		return false
	}

	return true
}

func (g *Git) WriteHooks() error {
	hpath := filepath.Join(g.dir, "hooks")

	postchf, err := os.Create(filepath.Join(hpath, "post-checkout"))
	if err != nil {
		return errwrap.Wrapf(fmt.Sprintf("Failed to create post-checkout in '%s': {{err}}", hpath), err)
	}

	err = PostCheckoutTmpl.Execute(postchf, struct{}{})
	if err != nil {
		return errwrap.Wrapf("Failed to run post-checkout template: {{err}}", err)
	}

	postcof, err := os.Create(filepath.Join(hpath, "post-commit"))
	if err != nil {
		return errwrap.Wrapf(fmt.Sprintf("Failed to create post-commit in '%s': {{err}}", hpath), err)
	}

	err = PostCheckoutTmpl.Execute(postcof, struct{}{})
	if err != nil {
		return errwrap.Wrapf("Failed to run post-commit template: {{err}}", err)
	}

	return nil
}
