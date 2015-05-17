package command

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
	"github.com/hashicorp/errwrap"

	"github.com/timeglass/glass/model"
)

type Status struct {
	*command
}

func NewStatus() *Status {
	return &Status{newCommand()}
}

func (c *Status) Name() string {
	return "status"
}

func (c *Status) Description() string {
	return ""
}

func (c *Status) Usage() string {
	return "Show info on the running timer"
}

func (c *Status) Flags() []cli.Flag {
	return []cli.Flag{}
}

func (c *Status) Action() func(ctx *cli.Context) {
	return c.command.Action(c.Run)
}

func (c *Status) Run(ctx *cli.Context) error {
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
	t, err := client.GetStatus()
	if err != nil {
		if err == ErrDaemonDown {
			return errwrap.Wrapf(fmt.Sprintf("No timer appears to be running for '%s': {{err}}", dir), err)
		} else {
			return err
		}
	}

	fmt.Println(t)
	return nil
}
