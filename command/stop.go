package command

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
	"github.com/hashicorp/errwrap"

	"github.com/timeglass/glass/model"
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
	return fmt.Sprintf("Terminates the timer process gracefully, it no timer is running it returns an error.")
}

func (c *Stop) Usage() string {
	return "Stop the timer completely"
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

	m := model.New(dir)
	info, err := m.ReadDaemonInfo()
	if err != nil {
		return errwrap.Wrapf(fmt.Sprintf("Failed to get Daemon address: {{err}}"), err)
	}

	client := NewClient(info)
	err = client.Call("timer.stop")
	if err != nil {
		if err == ErrDaemonDown {
			return errwrap.Wrapf(fmt.Sprintf("No timer appears to be running for '%s': {{err}}", dir), err)
		} else {
			return err
		}
	}

	fmt.Println("Timeglass: timer started")
	return nil
}
