package command

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
	"github.com/hashicorp/errwrap"

	"github.com/advanderveer/timer/vcs"
)

type Init struct {
	*command
}

func NewInit() *Init {
	return &Init{newCommand()}
}

func (c *Init) Name() string {
	return "init"
}

func (c *Init) Description() string {
	return fmt.Sprintf("<description>")
}

func (c *Init) Usage() string {
	return "<usage>"
}

func (c *Init) Flags() []cli.Flag {
	return []cli.Flag{}
}

func (c *Init) Action() func(ctx *cli.Context) {
	return c.command.Action(c.Run)
}

func (c *Init) Run(ctx *cli.Context) error {
	dir, err := os.Getwd()
	if err != nil {
		return errwrap.Wrapf("Failed to fetch current working dir: {{err}}", err)
	}

	vcs, err := vcs.GetVCS(dir)
	if err != nil {
		return errwrap.Wrapf("Failed to setup VCS: {{err}}", err)
	}

	err = vcs.WriteHooks()
	if err != nil {
		return errwrap.Wrapf("Failed to write hooks: {{err}}", err)
	}

	_ = vcs
	//@todo init hooks

	return nil
}
