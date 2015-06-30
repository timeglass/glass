package command

import (
	"fmt"
	"os"

	"github.com/timeglass/glass/_vendor/github.com/codegangsta/cli"
	"github.com/timeglass/glass/_vendor/github.com/hashicorp/errwrap"
)

type Stop struct {
	*command
}

func NewStop() *Stop {
	return &Stop{newCommand()}
}

func (c *Stop) Name() string {
	return "stop"
}

func (c *Stop) Description() string {
	return fmt.Sprintf("...")
}

func (c *Stop) Usage() string {
	return "..."
}

func (c *Stop) Flags() []cli.Flag {
	return []cli.Flag{}
}

func (c *Stop) Action() func(ctx *cli.Context) {
	return c.command.Action(c.Run)
}

func (c *Stop) Run(ctx *cli.Context) error {
	dir, err := os.Getwd()
	if err != nil {
		return errwrap.Wrapf("Failed to fetch current working dir: {{err}}", err)
	}

	c.Println("Deleting timer...")

	client := NewClient()
	err = client.DeleteTimer(dir)
	if err != nil {
		return errwrap.Wrapf(fmt.Sprintf("Failed to delete timer: {{err}}"), err)
	}

	c.Println("Timer deleted!")
	return nil
}
