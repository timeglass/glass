package command

import (
	"fmt"
	"os"
	"time"

	"github.com/codegangsta/cli"
	"github.com/hashicorp/errwrap"

	"github.com/timeglass/glass/vcs"
)

type Punch struct {
	*command
}

func NewPunch() *Punch {
	return &Punch{newCommand()}
}

func (c *Punch) Name() string {
	return "punch"
}

func (c *Punch) Description() string {
	return fmt.Sprintf("Writes time to the metadata of the last commit, should be provided in the following format: 6h20m12s")
}

func (c *Punch) Usage() string {
	return "Manually register time spent on the last commit"
}

func (c *Punch) Flags() []cli.Flag {
	return []cli.Flag{}
}

func (c *Punch) Action() func(ctx *cli.Context) {
	return c.command.Action(c.Run)
}

func (c *Punch) Run(ctx *cli.Context) error {
	dir, err := os.Getwd()
	if err != nil {
		return errwrap.Wrapf("Failed to fetch current working dir: {{err}}", err)
	}

	if ctx.Args().First() == "" {
		return fmt.Errorf("Please provide the time you spent as the first argument")
	}

	t, err := time.ParseDuration(ctx.Args().First())
	if err != nil {
		return errwrap.Wrapf(fmt.Sprintf("Failed to parse provided argument '%s' as a valid duration (e.g 1h2m10s): {{err}}", ctx.Args().First()), err)
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

	return nil
}
