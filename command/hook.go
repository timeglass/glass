package command

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
	"github.com/hashicorp/errwrap"

	"github.com/advanderveer/timer/vcs"
)

type Hook struct {
	*command
}

func NewHook() *Hook {
	return &Hook{newCommand()}
}

func (c *Hook) Name() string {
	return "hook"
}

func (c *Hook) Description() string {
	return fmt.Sprintf("<description>")
}

func (c *Hook) Usage() string {
	return "<usage>"
}

func (c *Hook) Flags() []cli.Flag {
	return []cli.Flag{}
}

func (c *Hook) Action() func(ctx *cli.Context) {
	return c.command.Action(c.Run)
}

func (c *Hook) Run(ctx *cli.Context) error {
	dir, err := os.Getwd()
	if err != nil {
		return errwrap.Wrapf("Failed to fetch current working dir: {{err}}", err)
	}

	vcs, err := vcs.GetVCS(dir)
	if err != nil {
		return errwrap.Wrapf("Failed to setup VCS: {{err}}", err)
	}

	err = vcs.Hook()
	if err != nil {
		return errwrap.Wrapf("Failed to write hooks: {{err}}", err)
	}

	return nil
}
