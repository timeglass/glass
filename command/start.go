package command

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/codegangsta/cli"
	"github.com/hashicorp/errwrap"

	"github.com/advanderveer/timer/model"
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
	return fmt.Sprintf("<description>")
}

func (c *Start) Usage() string {
	return "<usage>"
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

	m := model.New(filepath.Join(dir, "timer.db"))
	daddr, err := m.ReadDaemonAddr()
	if err != nil {
		return errwrap.Wrapf(fmt.Sprintf("Failed to get Daemon address: {{err}}"), err)
	}

	fmt.Println(daddr)

	return nil
}
