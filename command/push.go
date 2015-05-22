package command

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/codegangsta/cli"
	"github.com/hashicorp/errwrap"

	"github.com/timeglass/glass/vcs"
)

type Push struct {
	*command
}

func NewPush() *Push {
	return &Push{newCommand()}
}

func (c *Push) Name() string {
	return "push"
}

func (c *Push) Description() string {
	return fmt.Sprintf("Pushes the Timeglass notes branch to the remote repository. Provide the remote's name as the first argument, if no argument is provided it tries to push to the VCS default remote")
}

func (c *Push) Usage() string {
	return "Push time data to the remote repository"
}

func (c *Push) Flags() []cli.Flag {
	return []cli.Flag{}
}

func (c *Push) Action() func(ctx *cli.Context) {
	return c.command.Action(c.Run)
}

func (c *Push) Run(ctx *cli.Context) error {
	dir, err := os.Getwd()
	if err != nil {
		return errwrap.Wrapf("Failed to fetch current working dir: {{err}}", err)
	}

	bytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return errwrap.Wrapf("Failed to read from stdin: {{err}}", err)
	}

	fmt.Println(string(bytes))

	vc, err := vcs.GetVCS(dir)
	if err != nil {
		return errwrap.Wrapf("Failed to setup VCS: {{err}}", err)
	}

	remote := ctx.Args().First()
	if remote == "" {
		remote = vc.DefaultRemote()
	}

	fmt.Printf("Pushing time-data to remote '%s'...\n", remote)
	err = vc.Push(remote)
	if err != nil {
		return errwrap.Wrapf("Failed to push time data: {{err}}", err)
	}

	return nil
}
