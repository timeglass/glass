package command

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
	"github.com/hashicorp/errwrap"

	"github.com/timeglass/glass/vcs"
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
	return fmt.Sprintf("Install hooks for the current repository, if hooks already exists they are truncated and rewritten.")
}

func (c *Init) Usage() string {
	return "Install Timeglass for the current repository"
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

	vc, err := vcs.GetVCS(dir)
	if err != nil {
		return errwrap.Wrapf("Failed to setup VCS: {{err}}", err)
	}

	err = vc.Hook()
	if err != nil {
		return errwrap.Wrapf("Failed to write hooks: {{err}}", err)
	}

	fmt.Println("Timeglass: hooks written")
	return nil
}
