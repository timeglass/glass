package command

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
	"github.com/hashicorp/errwrap"

	"github.com/timeglass/glass/vcs"
)

type Log struct {
	*command
}

func NewLog() *Log {
	return &Log{newCommand()}
}

func (c *Log) Name() string {
	return "log"
}

func (c *Log) Description() string {
	return fmt.Sprintf("<description>")
}

func (c *Log) Usage() string {
	return "<usage>"
}

func (c *Log) Flags() []cli.Flag {
	return []cli.Flag{}
}

func (c *Log) Action() func(ctx *cli.Context) {
	return c.command.Action(c.Run)
}

func (c *Log) Run(ctx *cli.Context) error {
	dir, err := os.Getwd()
	if err != nil {
		return errwrap.Wrapf("Failed to fetch current working dir: {{err}}", err)
	}

	//write the vcs
	vc, err := vcs.GetVCS(dir)
	if err != nil {
		return errwrap.Wrapf("Failed to setup VCS: {{err}}", err)
	}

	err = vc.ParseHistory()
	if err != nil {
		return errwrap.Wrapf("Failed to parse VCS history: {{err}}", err)
	}

	return nil
}
