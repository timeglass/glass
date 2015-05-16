package command

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/codegangsta/cli"
	"github.com/hashicorp/errwrap"

	"github.com/timeglass/glass/model"
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
	return fmt.Sprintf("Starts the daemon that is reposible for tracking time for the current repository, if the daemon is already running this operations is a no-op.")
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

	m := model.New(dir)
	info, err := m.ReadDaemonInfo()
	if err != nil {
		return errwrap.Wrapf(fmt.Sprintf("Failed to get Daemon address: {{err}}"), err)
	}

	client := NewClient(info)
	err = client.Call("timer.start")
	if err != nil {
		if err != ErrDaemonDown {
			return err
		}

		cmd := exec.Command("glass-daemon")
		err := cmd.Start()
		if err != nil {
			return errwrap.Wrapf(fmt.Sprintf("Failed to start Daemon: {{err}}"), err)
		}
	}

	fmt.Println("Timeglass: timer started")
	return nil
}
