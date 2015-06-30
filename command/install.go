package command

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/timeglass/glass/_vendor/github.com/codegangsta/cli"
	"github.com/timeglass/glass/_vendor/github.com/hashicorp/errwrap"
)

type Install struct {
	*command
}

func NewInstall() *Install {
	return &Install{newCommand()}
}

func (c *Install) Name() string {
	return "install"
}

func (c *Install) Description() string {
	return fmt.Sprintf("...")
}

func (c *Install) Usage() string {
	return "..."
}

func (c *Install) Flags() []cli.Flag {
	return []cli.Flag{}
}

func (c *Install) Action() func(ctx *cli.Context) {
	return c.command.Action(c.Run)
}

func (c *Install) Run(ctx *cli.Context) error {
	c.Println("Installing the Timeglass background service...")

	//attempt to install
	cmd := exec.Command("glass-daemon", "install")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	err := cmd.Run()
	if err != nil {
		return errwrap.Wrapf(fmt.Sprintf("Failed to install Daemon: {{err}}"), err)
	}

	//attempt to start
	cmd = exec.Command("glass-daemon", "start")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err = cmd.Run()
	if err != nil {
		return errwrap.Wrapf(fmt.Sprintf("Failed to start Daemon: {{err}}"), err)
	}

	c.Println("Done!")
	return nil
}
