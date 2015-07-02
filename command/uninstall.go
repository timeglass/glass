package command

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/timeglass/glass/_vendor/github.com/codegangsta/cli"
	"github.com/timeglass/glass/_vendor/github.com/hashicorp/errwrap"
)

type Uninstall struct {
	*command
}

func NewUninstall() *Uninstall {
	return &Uninstall{newCommand()}
}

func (c *Uninstall) Name() string {
	return "uninstall"
}

func (c *Uninstall) Description() string {
	return fmt.Sprintf("Runs the glass-daemon executable with both stop and uninstall. It requires admin privileges on windows and linux.")
}

func (c *Uninstall) Usage() string {
	return "Stop and uninstall the background service"
}

func (c *Uninstall) Flags() []cli.Flag {
	return []cli.Flag{}
}

func (c *Uninstall) Action() func(ctx *cli.Context) {
	return c.command.Action(c.Run)
}

func (c *Uninstall) Run(ctx *cli.Context) error {
	c.Println("Stopping the Timeglass background service...")

	//attempt to stop
	cmd := exec.Command("glass-daemon", "stop")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		return errwrap.Wrapf(fmt.Sprintf("Failed to stop Daemon: {{err}}"), err)
	}

	c.Println("Uninstalling the Timeglass background service...")

	//attempt to Uninstall
	cmd = exec.Command("glass-daemon", "uninstall")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	err = cmd.Run()
	if err != nil {
		return errwrap.Wrapf(fmt.Sprintf("Failed to Uninstall Daemon: {{err}}"), err)
	}

	c.Println("Done!")
	return nil
}
