package vcs

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/timeglass/glass/_vendor/github.com/hashicorp/errwrap"
)

var TimeSpentNotesRef = "time-spent"

const (
	TOTAL_PREFIX = "total="
)

var PrepCommitTmpl = template.Must(template.New("name").Parse(`#!/bin/sh
# only add time to template and message sources
# @see http://git-scm.com/docs/githooks#_prepare_commit_msg
case "$2" in
message|template) 
	printf "$(cat $1)$(glass status)" > "$1" ;;
esac
`))

var PostCommitTmpl = template.Must(template.New("name").Parse(`#!/bin/sh
#persist (punch) to newly created commit and reset the timer
glass status -t "{{"{{"}}.{{"}}"}}" | glass punch
glass reset
`))

var PrePushTmpl = template.Must(template.New("name").Parse(`#!/bin/sh
#push time data
glass push $1 --from-hook
`))

type gitTimeData struct {
	total time.Duration
}

func (g *gitTimeData) Total() time.Duration { return g.total }

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
func (g *Git) IsAvailable() bool {
	fi, err := os.Stat(g.dir)
	if err != nil || !fi.IsDir() {
		return false
	}

	return true
}

func (g *Git) Show(commit string) (TimeData, error) {
	data := &gitTimeData{}
	args := []string{"notes", "--ref=" + TimeSpentNotesRef, "show", commit}
	outbuff := bytes.NewBuffer(nil)
	errbuff := bytes.NewBuffer(nil)
	cmd := exec.Command("git", args...)
	cmd.Stdout = outbuff
	cmd.Stderr = errbuff

	err := cmd.Run()
	if err != nil && strings.Contains(errbuff.String(), "No note found for object") {
		return data, ErrNoCommitTimeData
	}

	//in other cases present user with git output
	_, err2 := io.Copy(os.Stderr, errbuff)
	if err2 != nil {
		return data, err
	}

	if err != nil {
		return data, errwrap.Wrapf(fmt.Sprintf("Failed to show time for commit '%s' using git args %s: {{err}}", commit, args), err)
	}

	//scan lines in note
	scanner := bufio.NewScanner(outbuff)
	for scanner.Scan() {
		line := scanner.Text()

		//@todo for now only read total line
		if strings.HasPrefix(line, TOTAL_PREFIX) {
			t, err := time.ParseDuration(line[len(TOTAL_PREFIX):])
			if err != nil {
				return data, errwrap.Wrapf(fmt.Sprintf("Failed to parse time from line '%s': {{err}}"), err)
			}

			data.total = t
		}
	}
	if err := scanner.Err(); err != nil {
		return data, errwrap.Wrapf(fmt.Sprintf("Failed to scan note for commit '%s': {{err}}"), err)
	}

	return data, nil
}

func (g *Git) Persist(t time.Duration) error {
	args := []string{"notes", "--ref=" + TimeSpentNotesRef, "add", "-f", "-m", fmt.Sprintf("%s%s", TOTAL_PREFIX, t)}
	cmd := exec.Command("git", args...)
	err := cmd.Run()
	if err != nil {
		return errwrap.Wrapf(fmt.Sprintf("Failed to persist time '%s' using git command %s: {{err}}", t, args), err)
	}

	return nil
}

func (g *Git) Pull(remote string) error {
	args := []string{"fetch", remote, fmt.Sprintf("refs/notes/%s:refs/notes/%s", TimeSpentNotesRef, TimeSpentNotesRef)}
	cmd := exec.Command("git", args...)
	buff := bytes.NewBuffer(nil)

	cmd.Stdout = os.Stdout
	cmd.Stderr = buff

	err := cmd.Run()
	if err != nil && strings.Contains(buff.String(), "Couldn't find remote ref") {
		return ErrNoRemoteTimeData
	}

	//in other cases present user with git output
	_, err2 := io.Copy(os.Stderr, buff)
	if err2 != nil {
		return err
	}

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
	buff := bytes.NewBuffer(nil)

	cmd.Stdout = os.Stdout
	cmd.Stderr = buff

	err := cmd.Run()
	if err != nil && strings.Contains(buff.String(), "src refspec refs/notes/"+TimeSpentNotesRef+" does not match any") {
		return ErrNoLocalTimeData
	}

	//in other cases present user with git output
	_, err2 := io.Copy(os.Stderr, buff)
	if err2 != nil {
		return err
	}

	if err != nil {
		return errwrap.Wrapf(fmt.Sprintf("Failed to push to remote '%s' using git command %s: {{err}}", remote, args), err)
	}

	return nil
}

func (g *Git) Hook() error {
	hpath := filepath.Join(g.dir, "hooks")

	//prepare commit msg: status()
	prepcopath := filepath.Join(hpath, "prepare-commit-msg")
	prepcof, err := os.Create(prepcopath)
	if err != nil {
		return errwrap.Wrapf(fmt.Sprintf("Failed to create prepare-commit-msg  '%s': {{err}}", prepcof.Name()), err)
	}

	err = os.Chmod(prepcopath, 0766)
	if err != nil {
		return errwrap.Wrapf(fmt.Sprintf("Failed to make prepare-commit-msg file '%s' executable: {{err}}", hpath), err)
	}

	err = PrepCommitTmpl.Execute(prepcof, struct{}{})
	if err != nil {
		return errwrap.Wrapf("Failed to run post-commit template: {{err}}", err)
	}

	//post commit: lap()
	postcopath := filepath.Join(hpath, "post-commit")
	postcof, err := os.Create(postcopath)
	if err != nil {
		return errwrap.Wrapf(fmt.Sprintf("Failed to create post-commit '%s': {{err}}", postcof.Name()), err)
	}

	err = os.Chmod(postcopath, 0766)
	if err != nil {
		return errwrap.Wrapf(fmt.Sprintf("Failed to make post-commit file '%s' executable: {{err}}", hpath), err)
	}

	err = PostCommitTmpl.Execute(postcof, struct{}{})
	if err != nil {
		return errwrap.Wrapf("Failed to run post-commit template: {{err}}", err)
	}

	//post receive: push()
	prepushpath := filepath.Join(hpath, "pre-push")
	prepushf, err := os.Create(prepushpath)
	if err != nil {
		return errwrap.Wrapf(fmt.Sprintf("Failed to create pre-push  '%s': {{err}}", prepushf.Name()), err)
	}

	err = os.Chmod(prepushpath, 0766)
	if err != nil {
		return errwrap.Wrapf(fmt.Sprintf("Failed to make pre-push file '%s' executable: {{err}}", hpath), err)
	}

	err = PrePushTmpl.Execute(prepushf, struct{}{})
	if err != nil {
		return errwrap.Wrapf("Failed to run pre-push template: {{err}}", err)
	}

	return nil
}
