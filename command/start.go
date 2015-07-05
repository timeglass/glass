package command

import (
	"fmt"
	"os"

	"github.com/timeglass/glass/_vendor/github.com/codegangsta/cli"
	"github.com/timeglass/glass/_vendor/github.com/hashicorp/errwrap"

	"github.com/timeglass/glass/vcs"
)

type Start struct {
	*command
}

func NewStart() *Start {
	return &Start{newCommand()}
}

func (c *Start) Name() string {
	return "start"
}

func (c *Start) Description() string {
	return fmt.Sprintf("Creates a new timer for the current repository, if it is currently paused the timer continues.")
}

func (c *Start) Usage() string {
	return "Manually start timer for the current repository"
}

func (c *Start) Flags() []cli.Flag {
	return []cli.Flag{}
}

func (c *Start) Action() func(ctx *cli.Context) {
	return c.command.Action(c.Run)
}

func (c *Start) Run(ctx *cli.Context) error {
	dir, err := os.Getwd()
	if err != nil {
		return errwrap.Wrapf("Failed to fetch current working dir: {{err}}", err)
	}

	vc, err := vcs.GetVCS(dir)
	if err != nil {
		return errwrap.Wrapf("Failed to setup VCS: {{err}}", err)
	}

	c.Println("Starting timer...")

	client := NewClient()
	err = client.CreateTimer(vc.Root())
	if err != nil {
		return errwrap.Wrapf("Failed to create timer: {{err}}", err)
	}

	c.Println("Timer started!")
	return nil
}
