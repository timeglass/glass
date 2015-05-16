package command

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
	"github.com/hashicorp/errwrap"

	"github.com/advanderveer/timer/model"
)

type Lap struct {
	*command
}

func NewLap() *Lap {
	return &Lap{newCommand()}
}

func (c *Lap) Name() string {
	return "lap"
}

func (c *Lap) Description() string {
	return fmt.Sprintf("<description>")
}

func (c *Lap) Usage() string {
	return "<usage>"
}

func (c *Lap) Flags() []cli.Flag {
	return []cli.Flag{}
}

func (c *Lap) Action() func(ctx *cli.Context) {
	return c.command.Action(c.Run)
}

func (c *Lap) Run(ctx *cli.Context) error {
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
	t, err := client.Lap()
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
