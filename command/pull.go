package command

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
	"github.com/hashicorp/errwrap"

	"github.com/timeglass/glass/vcs"
)

type Pull struct {
	*command
}

func NewPull() *Pull {
	return &Pull{newCommand()}
}

func (c *Pull) Name() string {
	return "pull"
}

func (c *Pull) Description() string {
	return fmt.Sprintf("PUll the Timeglass notes branch from the remote repository. Provide the remote's name as the first argument, if no argument is provided it tries to pull from to the VCS default remote")
}

func (c *Pull) Usage() string {
	return "Pull time data from the remote repository"
}

func (c *Pull) Flags() []cli.Flag {
	return []cli.Flag{}
}

func (c *Pull) Action() func(ctx *cli.Context) {
	return c.command.Action(c.Run)
}

func (c *Pull) Run(ctx *cli.Context) error {
	dir, err := os.Getwd()
	if err != nil {
		return errwrap.Wrapf("Failed to fetch current working dir: {{err}}", err)
	}

	vc, err := vcs.GetVCS(dir)
	if err != nil {
		return errwrap.Wrapf("Failed to setup VCS: {{err}}", err)
	}

	remote := ctx.Args().First()
	if remote == "" {
		remote = vc.DefaultRemote()
	}

	err = vc.Fetch(remote)
	if err != nil {
		if err == vcs.ErrNoRemoteTimeData {
			fmt.Printf("Timeglass: remote '%s' has no time data (yet), nothing to pull\n", remote)
			return nil
		}

		return errwrap.Wrapf("Failed to pull time data: {{err}}", err)
	}

	fmt.Println("Timeglass: time data fetched successfully")
	return nil
}
