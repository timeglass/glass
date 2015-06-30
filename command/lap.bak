package command

import (
	"fmt"
	"os"

	"github.com/timeglass/glass/_vendor/github.com/codegangsta/cli"
	"github.com/timeglass/glass/_vendor/github.com/hashicorp/errwrap"

	"github.com/timeglass/glass/model"
	"github.com/timeglass/glass/vcs"
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
	return fmt.Sprintf("Resets the running timer, report spent time and punch as time spent on last commit")
}

func (c *Lap) Usage() string {
	return "Register time spent on the last commit and reset the timer"
}

func (c *Lap) Flags() []cli.Flag {
	return []cli.Flag{
		cli.BoolFlag{
			Name:  "from-hook",
			Usage: "Indicate it is called from a git, now expects refs on stdin",
		},
	}
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

	//get time and reset
	client := NewClient(info)
	t, err := client.Lap()
	if err != nil {
		if err == ErrDaemonDown {
			//if were calling this from a hook, supress errors
			if ctx.Bool("from-hook") {
				return nil
			}

			return errwrap.Wrapf(fmt.Sprintf("No timer appears to be running for '%s': {{err}}", dir), err)
		} else {
			return err
		}
	}

	//write the vcs
	vc, err := vcs.GetVCS(dir)
	if err != nil {
		return errwrap.Wrapf("Failed to setup VCS: {{err}}", err)
	}

	err = vc.Persist(t)
	if err != nil {
		return errwrap.Wrapf("Failed to log time into VCS: {{err}}", err)
	}

	fmt.Println(t)
	return nil
}
