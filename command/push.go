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
	return []cli.Flag{
		cli.BoolFlag{
			Name:  "from-hook",
			Usage: "Indicate it is called from a git, now expects refs on stdin",
		},
	}
}

func (c *Push) Action() func(ctx *cli.Context) {
	return c.command.Action(c.Run)
}

func (c *Push) Run(ctx *cli.Context) error {
	dir, err := os.Getwd()
	if err != nil {
		return errwrap.Wrapf("Failed to fetch current working dir: {{err}}", err)
	}

	//hooks require us require us to check the refs that are pushed over stdin
	//to prevent inifinte push loop
	refs := ""
	if ctx.Bool("from-hook") {
		bytes, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return errwrap.Wrapf("Failed to read from stdin: {{err}}", err)
		}

		refs = string(bytes)
		//when `glass push` triggers the pre-push hook it will not
		//provide any refs on stdin
		//this probalby means means there is nothing left to push and
		//we return here to prevent recursive push
		if refs == "" {
			return nil
		}
	}

	vc, err := vcs.GetVCS(dir)
	if err != nil {
		return errwrap.Wrapf("Failed to setup VCS: {{err}}", err)
	}

	remote := ctx.Args().First()
	if remote == "" {
		remote = vc.DefaultRemote()
	}

	err = vc.Push(remote, refs)
	if err != nil {
		return errwrap.Wrapf("Failed to push time data: {{err}}", err)
	}

	return nil
}
