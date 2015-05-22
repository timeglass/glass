package vcs

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/hashicorp/errwrap"
)

var TimeSpentNotesRef = "time-spent"

var PostCheckoutTmpl = template.Must(template.New("name").Parse(`#!/bin/sh
# when checkout is a branch, start timer
if [ $3 -eq 1 ]; then
   glass start;
fi
`))

var PrepCommitTmpl = template.Must(template.New("name").Parse(`#!/bin/sh
# only add time to template and message sources
# @see http://git-scm.com/docs/githooks#_prepare_commit_msg
case "$2" in
message|template) 
	printf "$(cat $1)$(glass status --time-only)" > "$1" ;;
esac
`))

var PostCommitTmpl = template.Must(template.New("name").Parse(`#!/bin/sh
#always reset after commit
glass lap
`))

var PrePushTmpl = template.Must(template.New("name").Parse(`#!/bin/sh
#push time data
echo Hook $1 $2
glass push $1 --refs-on-stdin
`))

type Git struct {
	dir string
}

func NewGit(dir string) *Git {
	return &Git{
		dir: filepath.Join(dir, ".git"),
	}
}

func (g *Git) DefaultRemote() string { return "origin" }
func (g *Git) Name() string          { return "git" }
func (g *Git) Supported() bool {
	fi, err := os.Stat(g.dir)
	if err != nil || !fi.IsDir() {
		return false
	}

	return true
}

func (g *Git) Log(t time.Duration) error {
	args := []string{"notes", "--ref=" + TimeSpentNotesRef, "add", "-f", "-m", fmt.Sprintf("total=%s", t)}
	cmd := exec.Command("git", args...)
	err := cmd.Run()
	if err != nil {
		return errwrap.Wrapf(fmt.Sprintf("Failed to log time '%s' using git command %s: {{err}}", t, args), err)
	}

	return nil
}

func (g *Git) Fetch(remote string) error {
	args := []string{"fetch", remote, fmt.Sprintf("refs/notes/%s:refs/notes/%s", TimeSpentNotesRef, TimeSpentNotesRef)}
	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return errwrap.Wrapf(fmt.Sprintf("Failed to fetch from remote '%s' using git command %s: {{err}}", remote, args), err)
	}

	return nil
}

func (g *Git) Push(remote string, refs string) error {

	//if time ref is already pushed, dont do it again
	if strings.Contains(refs, TimeSpentNotesRef) {
		return nil
	}

	args := []string{"push", remote, fmt.Sprintf("refs/notes/%s", TimeSpentNotesRef)}
	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return errwrap.Wrapf(fmt.Sprintf("Failed to push to remote '%s' using git command %s: {{err}}", remote, args), err)
	}

	return nil
}

func (g *Git) Hook() error {
	hpath := filepath.Join(g.dir, "hooks")

	//post checkout: start()
	postchf, err := os.Create(filepath.Join(hpath, "post-checkout"))
	if err != nil {
		return errwrap.Wrapf(fmt.Sprintf("Failed to create post-checkout '%s': {{err}}", postchf.Name()), err)
	}

	err = postchf.Chmod(0766)
	if err != nil {
		return errwrap.Wrapf(fmt.Sprintf("Failed to make post-checkout file '%s' executable: {{err}}", hpath), err)
	}

	err = PostCheckoutTmpl.Execute(postchf, struct{}{})
	if err != nil {
		return errwrap.Wrapf("Failed to run post-checkout template: {{err}}", err)
	}

	//prepare commit msg: status()
	prepcof, err := os.Create(filepath.Join(hpath, "prepare-commit-msg"))
	if err != nil {
		return errwrap.Wrapf(fmt.Sprintf("Failed to create prepare-commit-msg  '%s': {{err}}", postchf.Name()), err)
	}

	err = prepcof.Chmod(0766)
	if err != nil {
		return errwrap.Wrapf(fmt.Sprintf("Failed to make prepare-commit-msg file '%s' executable: {{err}}", hpath), err)
	}

	err = PrepCommitTmpl.Execute(prepcof, struct{}{})
	if err != nil {
		return errwrap.Wrapf("Failed to run post-commit template: {{err}}", err)
	}

	//post commit: lap()
	postcof, err := os.Create(filepath.Join(hpath, "post-commit"))
	if err != nil {
		return errwrap.Wrapf(fmt.Sprintf("Failed to create post-commit '%s': {{err}}", postchf.Name()), err)
	}

	err = postcof.Chmod(0766)
	if err != nil {
		return errwrap.Wrapf(fmt.Sprintf("Failed to make post-commit file '%s' executable: {{err}}", hpath), err)
	}

	err = PostCommitTmpl.Execute(postcof, struct{}{})
	if err != nil {
		return errwrap.Wrapf("Failed to run post-commit template: {{err}}", err)
	}

	//post receive: push()
	prepushf, err := os.Create(filepath.Join(hpath, "pre-push"))
	if err != nil {
		return errwrap.Wrapf(fmt.Sprintf("Failed to create pre-push  '%s': {{err}}", postchf.Name()), err)
	}

	err = prepushf.Chmod(0766)
	if err != nil {
		return errwrap.Wrapf(fmt.Sprintf("Failed to make pre-push file '%s' executable: {{err}}", hpath), err)
	}

	err = PrePushTmpl.Execute(prepushf, struct{}{})
	if err != nil {
		return errwrap.Wrapf("Failed to run pre-push template: {{err}}", err)
	}

	return nil
}
