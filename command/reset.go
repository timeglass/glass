package command

import (
	"fmt"
	"os"

	"github.com/timeglass/glass/_vendor/github.com/codegangsta/cli"
	"github.com/timeglass/glass/_vendor/github.com/hashicorp/errwrap"

	"github.com/timeglass/glass/vcs"
)

type Reset struct {
	*command
}

func NewReset() *Reset {
	return &Reset{newCommand()}
}

func (c *Reset) Name() string {
	return "reset"
}

func (c *Reset) Description() string {
	return fmt.Sprintf("Allows for setting the timer of the current repository to 0, this will discard the current measurement without saving")
}

func (c *Reset) Usage() string {
	return "Manually reset the current timer to 0s"
}

func (c *Reset) Flags() []cli.Flag {
	return []cli.Flag{}
}

func (c *Reset) Action() func(ctx *cli.Context) {
	return c.command.Action(c.Run)
}

func (c *Reset) Run(ctx *cli.Context) error {
	dir, err := os.Getwd()
	if err != nil {
		return errwrap.Wrapf("Failed to fetch current working dir: {{err}}", err)
	}

	vc, err := vcs.GetVCS(dir)
	if err != nil {
		return errwrap.Wrapf("Failed to setup VCS: {{err}}", err)
	}

	c.Printf("Resetting timer to 0s...")

	client := NewClient()
	err = client.ResetTimer(vc.Root())
	if err != nil {
		return errwrap.Wrapf(fmt.Sprintf("Failed to reset timer: {{err}}"), err)
	}

	c.Printf("Timer is reset!")
	return nil
}
